package usecase

import (
	"context"
	"encoding/base64"
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
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAuthUsecase_LogIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockAccessTokenRepo := mock.NewMockAccessTokenRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)

	uc := NewAuthUsecase(mockAccessTokenRepo, mockUserRepo, mockSharedCryptor)

	input := &model.LogInInput{
		Email:    "valid.email@format.com",
		Password: "not2short",
	}

	encEmail := "encrypted email"
	encPw := base64.StdEncoding.EncodeToString([]byte(input.Password))
	now := time.Now().UTC()
	user := &model.User{
		ID:        uuid.New(),
		Email:     encEmail,
		Password:  encPw,
		Username:  "input.Username",
		IsActive:  true,
		Role:      model.RoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	}
	pwDecoded, err := base64.StdEncoding.DecodeString(user.Password)
	assert.NoError(t, err)

	tests := []common.TestStructure{
		{
			Name:   "invalid input: email invalid",
			MockFn: func() {},
			Run: func() {
				_, err := uc.LogIn(ctx, &model.LogInInput{
					Email:    "invalid format",
					Password: "passwordnya ok",
				})

				assert.Error(t, err)
				assert.Equal(t, err.Type, ErrInvalidLoginInput)
			},
		},
		{
			Name:   "invalid input: password too short",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.LogIn(ctx, &model.LogInInput{
					Email:    "valid.email@format.com",
					Password: "2short",
				})

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInvalidLoginInput)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
			},
		},
		{
			Name: "failed to encrypt email",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Return("", errors.New("failed to encrypt"))
			},
			Run: func() {
				_, cerr := uc.LogIn(ctx, input)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "failed to fetch user data from db",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Return(nil, errors.New("failed to fetch data"))
			},
			Run: func() {
				_, cerr := uc.LogIn(ctx, input)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "user not found",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.LogIn(ctx, input)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "user blocked by active status",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Return(&model.User{
					IsActive: false,
				}, nil)
			},
			Run: func() {
				_, cerr := uc.LogIn(ctx, input)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrUserIsBlocked)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "user blocked by deleted at",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Return(&model.User{
					DeletedAt: gorm.DeletedAt{
						Time:  time.Now().UTC(),
						Valid: true,
					},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.LogIn(ctx, input)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrUserIsBlocked)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "password mismatch",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Times(1).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().CompareHash(pwDecoded, []byte(input.Password)).Times(1).Return(errors.New("failed"))
			},
			Run: func() {
				_, cerr := uc.LogIn(ctx, input)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInvalidPassword)
				assert.Equal(t, cerr.Code, http.StatusUnauthorized)
			},
		},
		{
			Name: "failed to hash access token",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Times(1).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().CompareHash(pwDecoded, []byte(input.Password)).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Hash(gomock.Any()).Times(1).Return("", errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.LogIn(ctx, input)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "failed to save access token to db",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Times(1).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().CompareHash(pwDecoded, []byte(input.Password)).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Hash(gomock.Any()).Times(1).Times(1).Return("anythinglaaa", nil)
				mockAccessTokenRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.LogIn(ctx, input)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Times(1).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().CompareHash(pwDecoded, []byte(input.Password)).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Hash(gomock.Any()).Times(1).Times(1).Return("anythinglaaa", nil)
				mockAccessTokenRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(nil)
			},
			Run: func() {
				resp, cerr := uc.LogIn(ctx, input)

				assert.NoError(t, cerr.Type)
				assert.Equal(t, resp.UserID, user.ID)
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
