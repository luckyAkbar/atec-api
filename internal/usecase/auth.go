package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/config"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
)

type authUc struct {
	accessTokenRepo model.AccessTokenRepository
	userRepo        model.UserRepository
	sharedCryptor   common.SharedCryptor
}

// NewAuthUsecase returns a new AuthUsecase
func NewAuthUsecase(accessTokenRepo model.AccessTokenRepository, userRepo model.UserRepository, sharedCryptor common.SharedCryptor) model.AuthUsecase {
	return &authUc{
		accessTokenRepo: accessTokenRepo,
		userRepo:        userRepo,
		sharedCryptor:   sharedCryptor,
	}
}

func (u *authUc) LogIn(ctx context.Context, input *model.LogInInput) (*model.LogInOutput, *common.Error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "authUc.LogIn",
		"input": helper.Dump(input),
	})

	if err := input.Validate(); err != nil {
		return nil, &common.Error{
			Message: "invalid login input",
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrInvalidLoginInput,
		}
	}

	emailEnc, err := u.sharedCryptor.Encrypt(input.Email)
	if err != nil {
		logger.WithError(err).Error("failed to encrypt user mail")
		return nil, &common.Error{
			Message: "failed to encrypt email",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	user, err := u.userRepo.FindByEmail(ctx, emailEnc)
	switch err {
	default:
		return nil, &common.Error{
			Message: "failed to query user",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	case repository.ErrNotFound:
		return nil, &common.Error{
			Message: "not found",
			Cause:   repository.ErrNotFound,
			Code:    http.StatusNotFound,
			Type:    ErrResourceNotFound,
		}
	case nil:
		break
	}

	if user.IsBlocked() {
		return nil, &common.Error{
			Message: "user's account is blocked",
			Cause:   errors.New("user's account is blocked"),
			Code:    http.StatusForbidden,
			Type:    ErrUserIsBlocked,
		}
	}

	pwDecoded, err := base64.StdEncoding.DecodeString(user.Password)
	if err != nil {
		return nil, &common.Error{
			Message: "failed to decode base64 text",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	if err := u.sharedCryptor.CompareHash(pwDecoded, []byte(input.Password)); err != nil {
		return nil, &common.Error{
			Message: "invalid password",
			Cause:   err,
			Code:    http.StatusUnauthorized,
			Type:    ErrInvalidPassword,
		}
	}

	plain, crypted, err := u.sharedCryptor.CreateSecureToken()
	if err != nil {
		return nil, &common.Error{
			Message: "failed to create access token",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	now := time.Now().UTC()
	at := &model.AccessToken{
		ID:         uuid.New(),
		Token:      crypted,
		UserID:     user.ID,
		ValidUntil: time.Now().Add(config.AccessTokenActiveDuration()).UTC(),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := u.accessTokenRepo.Create(ctx, at); err != nil {
		return nil, &common.Error{
			Message: "failed to save access token data",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	return at.ToLogInOutput(plain), nilErr
}
