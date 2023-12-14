package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	commonMock "github.com/luckyAkbar/atec-api/internal/common/mock"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserUsecase_SignUp(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	dbmock := kit.DBmock
	ctrl := gomock.NewController(t)

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockPinRepo := mock.NewMockPinRepository(ctrl)
	mockEmailUsecase := mock.NewMockEmailUsecase(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)

	ctx := context.Background()

	uc := NewUserUsecase(mockUserRepo, mockPinRepo, mockSharedCryptor, mockEmailUsecase, kit.DB)

	validInput := &model.SignUpInput{
		Username:            "okelah",
		Email:               "ok.valid@email.test",
		Password:            "8charsss",
		PasswordConfimation: "8charsss",
	}

	emailEncrypted := "encrypted-email"
	hashedPassword := "hashed-password"
	hashedPin := "hashed-pin"
	alreadyRegisteredUser := &model.User{
		Email: emailEncrypted,
	}

	tests := []common.TestStructure{
		{
			Name:   "input signup using malformed email",
			MockFn: func() {},
			Run: func() {
				invalidInput := &model.SignUpInput{
					Username:            "okelah",
					Email:               "invalidmalformed #%@gmail.com",
					Password:            "iniPasswoordvalid",
					PasswordConfimation: "iniPasswoordvalid",
				}
				_, err := uc.SignUp(ctx, invalidInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusBadRequest)
				assert.Equal(t, err.Type, ErrInputSignUpInvalid)
			},
		},
		{
			Name:   "input signup password no longer than 8",
			MockFn: func() {},
			Run: func() {
				invalidInput := &model.SignUpInput{
					Username:            "okelah",
					Email:               "invalidmalformed #%@gmail.com",
					Password:            "7charss",
					PasswordConfimation: "7charss",
				}
				_, err := uc.SignUp(ctx, invalidInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusBadRequest)
				assert.Equal(t, err.Type, ErrInputSignUpInvalid)
			},
		},
		{
			Name:   "input signup password not same with confirmation",
			MockFn: func() {},
			Run: func() {
				invalidInput := &model.SignUpInput{
					Username:            "okelah",
					Email:               "invalidmalformed #%@gmail.com",
					Password:            "7charss",
					PasswordConfimation: "7charss-not matching",
				}
				_, err := uc.SignUp(ctx, invalidInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusBadRequest)
				assert.Equal(t, err.Type, ErrInputSignUpInvalid)
			},
		},
		{
			Name: "failed to encrypt email",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return("", errors.New("encryption failed"))
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)
			},
		},
		{
			Name: "failed to fetch user data from db",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return(emailEncrypted, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, emailEncrypted).Times(1).Return(nil, errors.New("failed to fetch data"))
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)
			},
		},
		{
			Name: "user email already registered",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return(emailEncrypted, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, emailEncrypted).Times(1).Return(alreadyRegisteredUser, nil)
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusBadRequest)
				assert.Equal(t, err.Type, ErrEmailAlreadyRegistered)
			},
		},
		{
			Name: "failed to hash new user password",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return(emailEncrypted, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, emailEncrypted).Times(1).Return(nil, repository.ErrNotFound)
				mockSharedCryptor.EXPECT().Hash([]byte(validInput.Password)).Times(1).Return("", errors.New("failed to hash"))
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)
			},
		},
		{
			Name: "failed when creating new user data",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return(emailEncrypted, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, emailEncrypted).Times(1).Return(nil, repository.ErrNotFound)
				mockSharedCryptor.EXPECT().Hash([]byte(validInput.Password)).Times(1).Return(hashedPassword, nil)
				dbmock.ExpectBegin()
				mockUserRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(errors.New("db error"))
				dbmock.ExpectRollback()
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)
			},
		},
		{
			Name: "failed encrypting OTP after success saving user's data",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return(emailEncrypted, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, emailEncrypted).Times(1).Return(nil, repository.ErrNotFound)
				mockSharedCryptor.EXPECT().Hash([]byte(validInput.Password)).Times(1).Return(hashedPassword, nil)
				dbmock.ExpectBegin()
				mockUserRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Hash(gomock.Any()).Times(1).Return("", errors.New("failed to hash"))
				dbmock.ExpectRollback()
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)
			},
		},
		{
			Name: "failed to save pin after success encrypting OTP & success saving user's data",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return(emailEncrypted, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, emailEncrypted).Times(1).Return(nil, repository.ErrNotFound)
				mockSharedCryptor.EXPECT().Hash([]byte(validInput.Password)).Times(1).Return(hashedPassword, nil)
				dbmock.ExpectBegin()
				mockUserRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Hash(gomock.Any()).Times(1).Return(hashedPin, nil)
				mockPinRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(errors.New("db error"))
				dbmock.ExpectRollback()
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)
			},
		},
		{
			Name: "failed when register pin verification email",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return(emailEncrypted, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, emailEncrypted).Times(1).Return(nil, repository.ErrNotFound)
				mockSharedCryptor.EXPECT().Hash([]byte(validInput.Password)).Times(1).Return(hashedPassword, nil)
				dbmock.ExpectBegin()
				mockUserRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Hash(gomock.Any()).Times(1).Return(hashedPin, nil)
				mockPinRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockEmailUsecase.EXPECT().Register(ctx, gomock.Any()).Times(1).Return(nil, errors.New("anything error"))
				dbmock.ExpectRollback()
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)
			},
		},
		{
			Name: "all success",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(validInput.Email).Times(1).Return(emailEncrypted, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, emailEncrypted).Times(1).Return(nil, repository.ErrNotFound)
				mockSharedCryptor.EXPECT().Hash([]byte(validInput.Password)).Times(1).Return(hashedPassword, nil)
				dbmock.ExpectBegin()
				mockUserRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Hash(gomock.Any()).Times(1).Return(hashedPin, nil)
				mockPinRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Times(1).Return(nil)
				mockEmailUsecase.EXPECT().Register(ctx, gomock.Any()).Times(1).Return(&model.Email{}, nil)
				dbmock.ExpectCommit()
			},
			Run: func() {
				_, err := uc.SignUp(ctx, validInput)
				assert.Equal(t, err.Type, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestUserUsecase_VerifyAccount(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	//dbmock := kit.DBmock
	ctrl := gomock.NewController(t)

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockPinRepo := mock.NewMockPinRepository(ctrl)
	mockEmailUsecase := mock.NewMockEmailUsecase(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)

	ctx := context.Background()

	uc := NewUserUsecase(mockUserRepo, mockPinRepo, mockSharedCryptor, mockEmailUsecase, kit.DB)

	input := &model.AccountVerificationInput{
		PinValidationID: uuid.New(),
		Pin:             "123456",
	}

	pin := &model.Pin{
		ID:                uuid.New(),
		Pin:               "hmmz",
		UserID:            uuid.New(),
		ExpiredAt:         time.Now().Add(time.Hour * 24).UTC(),
		RemainingAttempts: 3,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	user := &model.User{
		ID:        uuid.New(),
		Email:     "test@email.com",
		Password:  "hmmz",
		Username:  "okelah",
		IsActive:  true,
		Role:      model.RoleUser,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	tests := []common.TestStructure{
		{
			Name:   "missing input",
			MockFn: func() {},
			Run: func() {
				_, _, err := uc.VerifyAccount(ctx, &model.AccountVerificationInput{})
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusBadRequest)
				assert.Equal(t, err.Type, ErrInputAccountVerificationInvalid)

			},
		},
		{
			Name: "pin was not found on db",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, _, err := uc.VerifyAccount(ctx, input)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusNotFound)
				assert.Equal(t, err.Type, ErrResourceNotFound)

			},
		},
		{
			Name: "failed to query pin",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, _, err := uc.VerifyAccount(ctx, input)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)

			},
		},
		{
			Name: "pin expired by time",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(&model.Pin{
					ExpiredAt: time.Now().Add(-time.Hour * 24),
				}, nil)
			},
			Run: func() {
				_, _, err := uc.VerifyAccount(ctx, input)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusBadRequest)
				assert.Equal(t, err.Type, ErrPinExpired)

			},
		},
		{
			Name: "pin has 0 remaining attempts",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(&model.Pin{
					RemainingAttempts: 0,
				}, nil)
			},
			Run: func() {
				_, _, err := uc.VerifyAccount(ctx, input)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusBadRequest)
				assert.Equal(t, err.Type, ErrPinExpired)

			},
		},
		{
			Name: "hash verification failed also failed to decrement the remaining attempts",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(pin, nil)
				mockSharedCryptor.EXPECT().CompareHash(gomock.Any(), []byte(input.Pin)).Times(1).Return(errors.New("verification failed"))
				mockPinRepo.EXPECT().DecrementRemainingAttempts(ctx, pin.ID).Times(1).Return(errors.New("db err"))
			},
			Run: func() {
				_, _, err := uc.VerifyAccount(ctx, input)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)

			},
		},
		{
			Name: "hash verification failed yet success to decrement the remaining attempts",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(pin, nil)
				mockSharedCryptor.EXPECT().CompareHash(gomock.Any(), []byte(input.Pin)).Times(1).Return(errors.New("verification failed"))
				mockPinRepo.EXPECT().DecrementRemainingAttempts(ctx, pin.ID).Times(1).Return(nil)
			},
			Run: func() {
				_, failedResp, err := uc.VerifyAccount(ctx, input)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusBadRequest)
				assert.Equal(t, err.Type, ErrPinInvalid)
				assert.Equal(t, failedResp.RemainingAttempts, pin.RemainingAttempts-1)
			},
		},
		{
			Name: "failed when updating the user's active status",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(pin, nil)
				mockSharedCryptor.EXPECT().CompareHash(gomock.Any(), []byte(input.Pin)).Times(1).Return(nil)
				mockUserRepo.EXPECT().UpdateActiveStatus(ctx, pin.UserID, true).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, _, err := uc.VerifyAccount(ctx, input)
				assert.Error(t, err)
				assert.Equal(t, err.Code, http.StatusInternalServerError)
				assert.Equal(t, err.Type, ErrInternal)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(pin, nil)
				mockSharedCryptor.EXPECT().CompareHash(gomock.Any(), []byte(input.Pin)).Times(1).Return(nil)
				mockUserRepo.EXPECT().UpdateActiveStatus(ctx, pin.UserID, true).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().Decrypt(gomock.Any()).Return("decrypted", nil)
			},
			Run: func() {
				okresp, _, err := uc.VerifyAccount(ctx, input)
				assert.NoError(t, err.Type)
				assert.Equal(t, okresp.IsActive, true)
			},
		},
		{
			Name: "ok - even if failed to decrypt",
			MockFn: func() {
				mockPinRepo.EXPECT().FindByID(ctx, input.PinValidationID).Times(1).Return(pin, nil)
				mockSharedCryptor.EXPECT().CompareHash(gomock.Any(), []byte(input.Pin)).Times(1).Return(nil)
				mockUserRepo.EXPECT().UpdateActiveStatus(ctx, pin.UserID, true).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().Decrypt(gomock.Any()).Return("", errors.New("err"))
			},
			Run: func() {
				okresp, _, err := uc.VerifyAccount(ctx, input)
				assert.NoError(t, err.Type)
				assert.Equal(t, okresp.IsActive, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}
func TestUserUsecase_generatePinForOTP(t *testing.T) {
	viper.Set("env", "definitly-production")
	for i := 0; i < 1000000; i++ {
		res := generatePinForOTP()
		assert.Len(t, res, 6)
	}

	viper.Reset()

	viper.Set("env", "LOcAL")
	for i := 0; i < 1000000; i++ {
		res := generatePinForOTP()
		assert.Len(t, res, 6)
		assert.Equal(t, res, "123456")
	}
	viper.Reset()
}

func TestUserUsecase_InitiateResetPassword(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctrl := gomock.NewController(t)

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockPinRepo := mock.NewMockPinRepository(ctrl)
	mockEmailUsecase := mock.NewMockEmailUsecase(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)

	user := model.AuthUser{
		UserID:      uuid.New(),
		Role:        model.RoleUser,
		AccessToken: "token1",
	}

	admin := model.AuthUser{
		UserID:      uuid.New(),
		Role:        model.RoleAdmin,
		AccessToken: "token2",
	}

	plainEmail := "email@mail.com"
	emailEnc := "encEmail"

	targetUser := &model.User{
		ID:       user.UserID,
		Email:    emailEnc,
		Username: "username",
		IsActive: true,
	}

	ctx := context.Background()
	ctxAdmin := model.SetUserToCtx(ctx, admin)
	ctxUser := model.SetUserToCtx(ctx, user)

	uc := NewUserUsecase(mockUserRepo, mockPinRepo, mockSharedCryptor, mockEmailUsecase, kit.DB)

	tests := []common.TestStructure{
		{
			Name:   "uuid is nil",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctx, uuid.Nil)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
				assert.Equal(t, cerr.Type, ErrInputResetPasswordInvalid)
			},
		},
		{
			Name:   "safety check: self change password",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxAdmin, admin.UserID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
				assert.Equal(t, cerr.Type, ErrInputResetPasswordInvalid)
			},
		},
		{
			Name:   "safety check: requester role is user",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxUser, uuid.New())
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
				assert.Equal(t, cerr.Type, ErrInputResetPasswordInvalid)
			},
		},
		{
			Name: "failed to query user from db",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctxAdmin, user.UserID).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxAdmin, user.UserID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "user is not found",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctxAdmin, user.UserID).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxAdmin, user.UserID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
			},
		},
		{
			Name: "user is blocked by is active status",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctxAdmin, user.UserID).Times(1).Return(&model.User{
					IsActive: false,
				}, nil)
			},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxAdmin, user.UserID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusPreconditionFailed)
				assert.Equal(t, cerr.Type, ErrUserIsBlocked)
			},
		},
		{
			Name: "user is blocked by deleted at",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctxAdmin, user.UserID).Times(1).Return(&model.User{
					IsActive: true,
					DeletedAt: gorm.DeletedAt{
						Time:  time.Now(),
						Valid: true,
					},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxAdmin, user.UserID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusPreconditionFailed)
				assert.Equal(t, cerr.Type, ErrUserIsBlocked)
			},
		},
		{
			Name: "failed to write reset pw session",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctxAdmin, user.UserID).Times(1).Return(targetUser, nil)
				mockUserRepo.EXPECT().CreateChangePasswordSession(ctxAdmin, gomock.Any(), time.Minute*15, gomock.Any()).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxAdmin, user.UserID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "failed to decrypt user email",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctxAdmin, user.UserID).Times(1).Return(targetUser, nil)
				mockUserRepo.EXPECT().CreateChangePasswordSession(ctxAdmin, gomock.Any(), time.Minute*15, gomock.Any()).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Decrypt(targetUser.Email).Times(1).Return("", errors.New("err dec"))
			},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxAdmin, user.UserID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "failed to register email",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctxAdmin, user.UserID).Times(1).Return(targetUser, nil)
				mockUserRepo.EXPECT().CreateChangePasswordSession(ctxAdmin, gomock.Any(), time.Minute*15, gomock.Any()).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Decrypt(targetUser.Email).Times(1).Return(plainEmail, nil)
				mockEmailUsecase.EXPECT().Register(ctxAdmin, gomock.Any()).Times(1).Return(nil, errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.InitiateResetPassword(ctxAdmin, user.UserID)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctxAdmin, user.UserID).Times(1).Return(targetUser, nil)
				mockUserRepo.EXPECT().CreateChangePasswordSession(ctxAdmin, gomock.Any(), time.Minute*15, gomock.Any()).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Decrypt(targetUser.Email).Times(1).Return(plainEmail, nil)
				mockEmailUsecase.EXPECT().Register(ctxAdmin, gomock.Any()).Times(1).Return(&model.Email{ID: uuid.New()}, nil)
			},
			Run: func() {
				res, cerr := uc.InitiateResetPassword(ctxAdmin, user.UserID)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, targetUser.ID)
				assert.Equal(t, res.Email, plainEmail)
				assert.Equal(t, res.Username, targetUser.Username)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}
