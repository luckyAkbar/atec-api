package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/config"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

type userUc struct {
	userRepo      model.UserRepository
	pinRepo       model.PinRepository
	sharedCryptor common.SharedCryptor
	emailUsecase  model.EmailUsecase
	dbTrx         *gorm.DB
}

// NewUserUsecase create a new user usecase. Satisfy model.UserUsecase interface
func NewUserUsecase(userRepo model.UserRepository, pinRepo model.PinRepository, sharedCryptor common.SharedCryptor, emailUsecase model.EmailUsecase, dbTrx *gorm.DB) model.UserUsecase {
	return &userUc{
		userRepo:      userRepo,
		pinRepo:       pinRepo,
		sharedCryptor: sharedCryptor,
		emailUsecase:  emailUsecase,
		dbTrx:         dbTrx,
	}
}

func (u *userUc) SignUp(ctx context.Context, input *model.SignUpInput) (*model.SignUpResponse, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "userUc.SignUp",
		"input": helper.Dump(input),
	})

	if err := input.Validate(); err != nil {
		return nil, &common.Error{
			Message: "invalid values on input for signup",
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrInputSignUpInvalid,
		}
	}

	emailEnc, err := u.sharedCryptor.Encrypt(input.Email)
	if err != nil {
		return nil, &common.Error{
			Message: "failed to encrypt email",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	_, err = u.userRepo.FindByEmail(ctx, emailEnc)
	switch err {
	default:
		logger.WithError(err).Error("failed to find user by email")
		return nil, &common.Error{
			Message: "failed to find user by email",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case nil:
		return nil, &common.Error{
			Message: "unable to sign up using an already registered email. try to use another",
			Cause:   nil,
			Code:    http.StatusBadRequest,
			Type:    ErrEmailAlreadyRegistered,
		}
	case repository.ErrNotFound:
		break
	}

	hashedPassword, err := u.sharedCryptor.Hash([]byte(input.Password))
	if err != nil {
		return nil, &common.Error{
			Message: "failed to hash password",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	now := time.Now().UTC()
	user := &model.User{
		ID:        uuid.New(),
		Email:     emailEnc,
		Password:  hashedPassword,
		Username:  input.Username,
		IsActive:  false,
		Role:      model.RoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// start db transaction process
	tx := u.dbTrx.Begin()

	if err := u.userRepo.Create(ctx, user, tx); err != nil {
		logger.WithError(err).Error("failed to create user")
		tx.Rollback()
		return nil, &common.Error{
			Message: "failed to create user",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	otpPlain := generatePinForOTP()
	otpEnc, err := u.sharedCryptor.Hash([]byte(otpPlain))
	if err != nil {
		tx.Rollback()
		return nil, &common.Error{
			Message: "failed to encrypt otp",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	pin := &model.Pin{
		ID:                uuid.New(),
		Pin:               otpEnc,
		UserID:            user.ID,
		ExpiredAt:         time.Now().Add(time.Minute * time.Duration(config.PinExpiryMinutes())).UTC(),
		RemainingAttempts: config.PinMaxRetry(),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := u.pinRepo.Create(ctx, pin, tx); err != nil {
		logger.WithError(err).Error("failed to create pin")
		tx.Rollback()
		return nil, &common.Error{
			Message: "failed to create pin",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	_, err = u.emailUsecase.Register(ctx, generateEmailTemplateForPinVerification(user.Username, input.Email, otpPlain))
	if err != nil {
		logger.WithError(err).Error("failed to register PIN verification email")
		tx.Rollback()
		return nil, &common.Error{
			Message: "failed to register PIN verification email",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	// commit the transaction after all process succeed
	tx.Commit()

	// TODO: what if there is an error when commit / rollback? for now, it's considered a non problem
	// but it might be at the future

	return &model.SignUpResponse{
		PinValidationID:   pin.ID.String(),
		PinExpiredAt:      pin.ExpiredAt,
		RemainingAttempts: config.PinMaxRetry(),
	}, nilErr
}

func (u *userUc) VerifyAccount(ctx context.Context, input *model.AccountVerificationInput) (*model.SuccessAccountVerificationResponse, *model.FailedAccountVerificationResponse, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "userUc.VerifyAccount",
		"input": helper.Dump(input),
	})

	if err := input.Validate(); err != nil {
		return nil, nil, &common.Error{
			Message: "invalid values on input for account verification",
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrInputAccountVerificationInvalid,
		}
	}

	pin, err := u.pinRepo.FindByID(ctx, input.PinValidationID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find pin by id")
		return nil, nil, &common.Error{
			Message: "failed to find pin by id",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, nil, &common.Error{
			Message: "data not found",
			Cause:   err,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		break
	}

	if pin.IsExpired() {
		return nil, nil, &common.Error{
			Message: "pin is expired",
			Cause:   nil,
			Code:    http.StatusBadRequest,
			Type:    ErrPinExpired,
		}
	}

	pinDecoded, err := base64.StdEncoding.DecodeString(pin.Pin)
	if err != nil {
		return nil, nil, &common.Error{
			Message: "failed to decode pin",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	err = u.sharedCryptor.CompareHash(pinDecoded, []byte(input.Pin))
	switch err {
	default:
		logger.WithError(err).Warn("sharedCryptor.Compare returning non nil error")
		err = u.pinRepo.DecrementRemainingAttempts(ctx, pin.ID)
		if err != nil {
			logger.WithError(err).Error("failed to decrement remaining attempts for failed pin validation")
			return nil, nil, &common.Error{
				Message: "failed to decrement remaining attempts for failed pin validation",
				Cause:   err,
				Code:    http.StatusInternalServerError,
				Type:    ErrInternal,
			}
		}

		return nil, &model.FailedAccountVerificationResponse{
				RemainingAttempts: pin.RemainingAttempts - 1,
			}, &common.Error{
				Message: "invalid pin",
				Cause:   err,
				Code:    http.StatusBadRequest,
				Type:    ErrPinInvalid,
			}
	case nil:
		break
	}

	user, err := u.userRepo.UpdateActiveStatus(ctx, pin.UserID, true)
	if err != nil {
		logger.WithError(err).Error("failed to update user active status")
		return nil, nil, &common.Error{
			Message: "failed to update user active status",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	return &model.SuccessAccountVerificationResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		IsActive:  user.IsActive,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil, nilErr
}

func generatePinForOTP() string {
	// was made for easier testing on a non production environment
	if strings.ToLower(config.Env()) == "local" {
		return "123456"
	}

	max := 6
	b := make([]byte, max)
	chars := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

	for i := 0; i < 6; i++ {

		n, err := io.ReadAtLeast(rand.Reader, b, max)
		if n != max {
			panic(err)
		}
		for i := 0; i < len(b); i++ {
			b[i] = chars[int(b[i])%len(chars)]
		}
	}

	return string(b)
}

func generateEmailTemplateForPinVerification(username, email, pin string) *model.RegisterEmailInput {
	return &model.RegisterEmailInput{
		Subject: "Verifikasi Akun",
		Body: fmt.Sprintf(`
			<h2>Halo %s!</h2>
			<p>Terimakasih telah mendaftar pada layanan Autism Treatment Evaluation Checklist (ATEC)</p>
			</p> Untuk mengaktifkan akun Anda, silakan masukan kode PIN berikut</p> <br>
			<h3><strong>%s</strong></h3>
		`, username, pin),
		To: []string{email},
	}
}
