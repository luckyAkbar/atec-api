package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/luckyAkbar/atec-api/internal/common"
	commonMock "github.com/luckyAkbar/atec-api/internal/common/mock"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/stretchr/testify/assert"
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

func TestUserUsecase_generatePinForOTP(t *testing.T) {
	for i := 0; i < 1000000; i++ {
		res := generatePinForOTP()
		assert.Len(t, res, 6)
	}
}
