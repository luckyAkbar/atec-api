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
	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/atec-api/internal/common"
	commonMock "github.com/luckyAkbar/atec-api/internal/common/mock"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAuthUsecase_LogIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockAccessTokenRepo := mock.NewMockAccessTokenRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)
	mockWorkerClient := mock.NewMockWorkerClient(ctrl)

	uc := NewAuthUsecase(mockAccessTokenRepo, mockUserRepo, mockSharedCryptor, mockWorkerClient)

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
			Name: "failed to generate access token",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Times(1).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().CompareHash(pwDecoded, []byte(input.Password)).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("", "", errors.New("err access token"))
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
				mockSharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("plain", "crypted", nil)
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
				mockSharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("plain", "crypted", nil)
				mockAccessTokenRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(nil)
				mockWorkerClient.EXPECT().EnqueueEnforceActiveTokenLimitterTask(ctx, user.ID).Times(1).Return(&asynq.TaskInfo{}, nil)
			},
			Run: func() {
				viper.Set("server.auth.active_token_limit", 10)
				resp, cerr := uc.LogIn(ctx, input)

				assert.NoError(t, cerr.Type)
				assert.Equal(t, resp.UserID, user.ID)
			},
		},
		{
			Name: "failed to enqueue enforce active token limiter, but thats ok",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Times(1).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().CompareHash(pwDecoded, []byte(input.Password)).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("plain", "crypted", nil)
				mockAccessTokenRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(nil)
				mockWorkerClient.EXPECT().EnqueueEnforceActiveTokenLimitterTask(ctx, user.ID).Times(1).Return(nil, errors.New("err worker"))
			},
			Run: func() {
				viper.Set("server.auth.active_token_limit", 10)
				resp, cerr := uc.LogIn(ctx, input)

				assert.NoError(t, cerr.Type)
				assert.Equal(t, resp.UserID, user.ID)
			},
		},
		{
			Name: "ok-active token limiter is disabled",
			MockFn: func() {
				mockSharedCryptor.EXPECT().Encrypt(input.Email).Times(1).Return(encEmail, nil)
				mockUserRepo.EXPECT().FindByEmail(ctx, encEmail).Times(1).Return(user, nil)
				mockSharedCryptor.EXPECT().CompareHash(pwDecoded, []byte(input.Password)).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().CreateSecureToken().Times(1).Return("plain", "crypted", nil)
				mockAccessTokenRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(nil)
				//mockWorkerClient.EXPECT().EnqueueEnforceActiveTokenLimitterTask(ctx, user.ID).Times(1).Return(&asynq.TaskInfo{}, nil)
			},
			Run: func() {
				viper.Set("server.auth.active_token_limit", 0)
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

func TestAuthUsecase_LogOut(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockAccessTokenRepo := mock.NewMockAccessTokenRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)
	mockWorkerClient := mock.NewMockWorkerClient(ctrl)

	uc := NewAuthUsecase(mockAccessTokenRepo, mockUserRepo, mockSharedCryptor, mockWorkerClient)
	tokenEnc := "encrypted token"
	token := &model.AccessToken{
		ID:    uuid.New(),
		Token: tokenEnc,
	}
	au := model.AuthUser{
		UserID:      uuid.New(),
		AccessToken: token.Token,
		Role:        model.RoleUser,
	}

	ctx := model.SetUserToCtx(context.Background(), au)
	tests := []common.TestStructure{
		{
			Name: "token not found",
			MockFn: func() {
				mockAccessTokenRepo.EXPECT().FindByToken(ctx, tokenEnc).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				cerr := uc.LogOut(ctx)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
			},
		},
		{
			Name: "failed to find token",
			MockFn: func() {
				mockAccessTokenRepo.EXPECT().FindByToken(ctx, tokenEnc).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				cerr := uc.LogOut(ctx)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "failed to delete token",
			MockFn: func() {
				mockAccessTokenRepo.EXPECT().FindByToken(ctx, tokenEnc).Times(1).Return(token, nil)
				mockAccessTokenRepo.EXPECT().DeleteByID(ctx, token.ID).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				cerr := uc.LogOut(ctx)

				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockAccessTokenRepo.EXPECT().FindByToken(ctx, tokenEnc).Times(1).Return(token, nil)
				mockAccessTokenRepo.EXPECT().DeleteByID(ctx, token.ID).Times(1).Return(nil)
			},
			Run: func() {
				cerr := uc.LogOut(ctx)

				assert.Error(t, cerr)
				assert.Equal(t, cerr, nilErr)
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

func TestAuthUsecase_ValidateAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockAccessTokenRepo := mock.NewMockAccessTokenRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)
	mockWorkerClient := mock.NewMockWorkerClient(ctrl)

	uc := NewAuthUsecase(mockAccessTokenRepo, mockUserRepo, mockSharedCryptor, mockWorkerClient)
	token := "token"
	revToken := "rev token"
	user := &model.User{
		ID:   uuid.New(),
		Role: model.RoleUser,
	}
	at := &model.AccessToken{
		ID:         uuid.New(),
		Token:      revToken,
		ValidUntil: time.Now().Add(time.Hour * 7),
	}

	tests := []common.TestStructure{
		{
			Name: "failed to find token from db",
			MockFn: func() {
				mockSharedCryptor.EXPECT().ReverseSecureToken(token).Times(1).Return(revToken)
				mockAccessTokenRepo.EXPECT().FindCredentialByToken(ctx, revToken).Times(1).Return(nil, nil, errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.ValidateAccess(ctx, token)
				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "creds not found",
			MockFn: func() {
				mockSharedCryptor.EXPECT().ReverseSecureToken(token).Times(1).Return(revToken)
				mockAccessTokenRepo.EXPECT().FindCredentialByToken(ctx, revToken).Times(1).Return(nil, nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.ValidateAccess(ctx, token)
				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
			},
		},
		{
			Name: "token expired by valid until",
			MockFn: func() {
				mockSharedCryptor.EXPECT().ReverseSecureToken(token).Times(1).Return(revToken)
				mockAccessTokenRepo.EXPECT().FindCredentialByToken(ctx, revToken).Times(1).Return(&model.AccessToken{
					ValidUntil: time.Now().Add(time.Hour * -24).UTC(),
				}, user, nil)
			},
			Run: func() {
				_, cerr := uc.ValidateAccess(ctx, token)
				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrAccessTokenExpired)
			},
		},
		{
			Name: "token expired by deleted at",
			MockFn: func() {
				mockSharedCryptor.EXPECT().ReverseSecureToken(token).Times(1).Return(revToken)
				mockAccessTokenRepo.EXPECT().FindCredentialByToken(ctx, revToken).Times(1).Return(&model.AccessToken{
					ValidUntil: time.Now().Add(time.Hour * 24).UTC(),
					DeletedAt: gorm.DeletedAt{
						Time:  time.Now().UTC(),
						Valid: true,
					},
				}, user, nil)
			},
			Run: func() {
				_, cerr := uc.ValidateAccess(ctx, token)
				assert.Error(t, cerr)
				assert.Equal(t, cerr.Type, ErrAccessTokenExpired)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockSharedCryptor.EXPECT().ReverseSecureToken(token).Times(1).Return(revToken)
				mockAccessTokenRepo.EXPECT().FindCredentialByToken(ctx, revToken).Times(1).Return(at, user, nil)
			},
			Run: func() {
				res, cerr := uc.ValidateAccess(ctx, token)
				assert.Equal(t, cerr.Type, nil)
				assert.Equal(t, res.Role, model.RoleUser)
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

func TestAuthUsecase_ValidateResetPasswordSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockAccessTokenRepo := mock.NewMockAccessTokenRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)
	mockWorkerClient := mock.NewMockWorkerClient(ctrl)

	uc := NewAuthUsecase(mockAccessTokenRepo, mockUserRepo, mockSharedCryptor, mockWorkerClient)
	key := "key"
	session := &model.ChangePasswordSession{
		UserID:    uuid.New(),
		ExpiredAt: time.Now().Add(time.Minute * 15).UTC(),
		CreatedAt: time.Now().UTC(),
		CreatedBy: uuid.New(),
	}

	tests := []common.TestStructure{
		{
			Name:   "key is empty",
			MockFn: func() {},
			Run: func() {
				cerr := uc.ValidateResetPasswordSession(ctx, "")
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInvalidValidateChangePasswordSessionInput)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
			},
		},
		{
			Name: "unable to find reset password session",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, key).Times(1).Return(nil, errors.New("err"))
			},
			Run: func() {
				cerr := uc.ValidateResetPasswordSession(ctx, key)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "session not found",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, key).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				cerr := uc.ValidateResetPasswordSession(ctx, key)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "session is expired",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, key).Times(1).Return(&model.ChangePasswordSession{
					ExpiredAt: time.Now().Add(time.Minute * -1).UTC(),
				}, nil)
			},
			Run: func() {
				cerr := uc.ValidateResetPasswordSession(ctx, key)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResetPasswordSessionExpired)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, key).Times(1).Return(session, nil)
			},
			Run: func() {
				cerr := uc.ValidateResetPasswordSession(ctx, key)
				assert.NoError(t, cerr.Type)
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

func TestAuthUsecase_ResetPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockAccessTokenRepo := mock.NewMockAccessTokenRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)
	mockWorkerClient := mock.NewMockWorkerClient(ctrl)

	uc := NewAuthUsecase(mockAccessTokenRepo, mockUserRepo, mockSharedCryptor, mockWorkerClient)
	input := &model.ResetPasswordInput{
		Key:                 "valid key oke",
		Password:            "validpassword",
		PasswordConfimation: "validpassword",
	}

	changePwSess := &model.ChangePasswordSession{
		UserID:    uuid.New(),
		ExpiredAt: time.Now().Add(time.Minute * 15).UTC(),
		CreatedAt: time.Now().UTC(),
		CreatedBy: uuid.New(),
	}

	u := &model.User{
		ID:        uuid.New(),
		Email:     "email",
		Password:  "XXXXXXXX",
		Username:  "username",
		IsActive:  true,
		Role:      model.RoleAdmin,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		DeletedAt: gorm.DeletedAt{},
	}

	tests := []common.TestStructure{
		{
			Name:   "invalid input key empty",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, &model.ResetPasswordInput{})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInvalidResetPasswordInput)
			},
		},
		{
			Name:   "invalid input password empty",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, &model.ResetPasswordInput{
					Key: "somekey",
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInvalidResetPasswordInput)
			},
		},
		{
			Name:   "invalid input password confirmation mismatch",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, &model.ResetPasswordInput{
					Key:                 "somekey",
					Password:            "passwordhmhmhm",
					PasswordConfimation: "password??????",
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInvalidResetPasswordInput)
			},
		},
		{
			Name:   "invalid input password only 7 chars",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, &model.ResetPasswordInput{
					Key:                 "somekey",
					Password:            "7chars.",
					PasswordConfimation: "7chars.",
				})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInvalidResetPasswordInput)
			},
		},
		{
			Name: "db err: unable to find change pw session",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(nil, errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "db err: session not found",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
			},
		},
		{
			Name: "db err: session is expired",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(&model.ChangePasswordSession{
					ExpiredAt: time.Now().Add(time.Minute * -1).UTC(),
				}, nil)
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
				assert.Equal(t, cerr.Type, ErrResetPasswordSessionExpired)
			},
		},
		{
			Name: "db err: failed to find user data",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(changePwSess, nil)
				mockUserRepo.EXPECT().FindByID(ctx, changePwSess.UserID).Times(1).Return(nil, errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "db err: user data not found",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(changePwSess, nil)
				mockUserRepo.EXPECT().FindByID(ctx, changePwSess.UserID).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
			},
		},
		{
			Name: "user is blocked by active status",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(changePwSess, nil)
				mockUserRepo.EXPECT().FindByID(ctx, changePwSess.UserID).Times(1).Return(&model.User{
					IsActive: false,
				}, nil)
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusPreconditionFailed)
				assert.Equal(t, cerr.Type, ErrUserIsBlocked)
			},
		},
		{
			Name: "user is blocked by deleted at",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(changePwSess, nil)
				mockUserRepo.EXPECT().FindByID(ctx, changePwSess.UserID).Times(1).Return(&model.User{
					IsActive:  true,
					DeletedAt: gorm.DeletedAt{Time: time.Now().UTC(), Valid: true},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusPreconditionFailed)
				assert.Equal(t, cerr.Type, ErrUserIsBlocked)
			},
		},
		{
			Name: "failed to hash user password",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(changePwSess, nil)
				mockUserRepo.EXPECT().FindByID(ctx, changePwSess.UserID).Times(1).Return(u, nil)
				mockSharedCryptor.EXPECT().Hash([]byte(input.Password)).Times(1).Return("", errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "failed to update user password",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(changePwSess, nil)
				mockUserRepo.EXPECT().FindByID(ctx, changePwSess.UserID).Times(1).Return(u, nil)
				mockSharedCryptor.EXPECT().Hash([]byte(input.Password)).Times(1).Return("hashed", nil)
				mockUserRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(errors.New("err"))
			},
			Run: func() {
				_, cerr := uc.ResetPassword(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "ok - even if failed to decrypt user email",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(changePwSess, nil)
				mockUserRepo.EXPECT().FindByID(ctx, changePwSess.UserID).Times(1).Return(u, nil)
				mockSharedCryptor.EXPECT().Hash([]byte(input.Password)).Times(1).Return("hashed", nil)
				mockUserRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Decrypt(u.Email).Times(1).Return("", errors.New("err"))
			},
			Run: func() {
				res, cerr := uc.ResetPassword(ctx, input)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.Email, "")
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockUserRepo.EXPECT().FindChangePasswordSession(ctx, input.Key).Times(1).Return(changePwSess, nil)
				mockUserRepo.EXPECT().FindByID(ctx, changePwSess.UserID).Times(1).Return(u, nil)
				mockSharedCryptor.EXPECT().Hash([]byte(input.Password)).Times(1).Return("hashed", nil)
				mockUserRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(nil)
				mockSharedCryptor.EXPECT().Decrypt(u.Email).Times(1).Return("decrypted", nil)
			},
			Run: func() {
				res, cerr := uc.ResetPassword(ctx, input)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.Email, "decrypted")
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
