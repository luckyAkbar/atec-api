package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/sweet-go/stdlib/helper"
	"gorm.io/gorm"
)

func TestEAccessTokenRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)

	repo := NewAccessTokenRepository(kit.DB, mockCacher)
	mock := kit.DBmock
	ctx := context.Background()
	now := time.Now().UTC()
	at := &model.AccessToken{
		ID:         uuid.New(),
		Token:      "hashedToken",
		UserID:     uuid.New(),
		ValidUntil: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tests := []common.TestStructure{
		{
			Name: "success",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "access_tokens"`).
					WithArgs(at.ID, at.Token, at.UserID, at.ValidUntil, at.CreatedAt, at.UpdatedAt, at.DeletedAt).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Create(ctx, at)
				assert.NoError(t, err)
			},
		},
		{
			Name: "failed on database returning error",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "access_tokens"`).
					WithArgs(at.ID, at.Token, at.UserID, at.ValidUntil, at.CreatedAt, at.UpdatedAt, at.DeletedAt).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Create(ctx, at)
				assert.Error(t, err)
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

func TestEAccessTokenRepository_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)

	repo := NewAccessTokenRepository(kit.DB, mockCacher)
	mock := kit.DBmock
	ctx := context.Background()
	now := time.Now().UTC()
	at := &model.AccessToken{
		ID:         uuid.New(),
		Token:      "hashedToken",
		UserID:     uuid.New(),
		ValidUntil: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tests := []common.TestStructure{
		{
			Name: "found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens" WHERE`).
					WithArgs(at.Token).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(at.ID))
			},
			Run: func() {
				res, err := repo.FindByToken(ctx, at.Token)
				assert.NoError(t, err)

				assert.Equal(t, res.ID, at.ID)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens" WHERE`).
					WithArgs(at.Token).WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindByToken(ctx, at.Token)
				assert.Error(t, err)

				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "err",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens" WHERE`).
					WithArgs(at.Token).WillReturnError(errors.New("db error"))
			},
			Run: func() {
				_, err := repo.FindByToken(ctx, at.Token)
				assert.Error(t, err)
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

func TestEAccessTokenRepository_DeleteByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewAccessTokenRepository(kit.DB, mockCacher)
	mock := kit.DBmock
	ctx := context.Background()
	now := time.Now().UTC()
	at := &model.AccessToken{
		ID:         uuid.New(),
		Token:      "hashedToken",
		UserID:     uuid.New(),
		ValidUntil: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tests := []common.TestStructure{
		{
			Name: "success",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^DELETE FROM "access_tokens" WHERE "access_tokens"."id"`).
					WithArgs(at.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DeleteByID(ctx, at.ID)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^DELETE FROM "access_tokens" WHERE "access_tokens"."id"`).
					WithArgs(at.ID).WillReturnError(errors.New("err db"))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DeleteByID(ctx, at.ID)
				assert.Error(t, err)
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

func TestAccessTokenRepository_DeleteByUserID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewAccessTokenRepository(kit.DB, mockCacher)
	mock := kit.DBmock
	ctx := context.Background()

	at := &model.AccessToken{
		ID:         uuid.New(),
		Token:      "token",
		UserID:     uuid.New(),
		ValidUntil: time.Now().UTC().Add(time.Hour * 7),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "access_tokens" SET`).
					WithArgs(sqlmock.AnyArg(), at.UserID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DeleteByUserID(ctx, at.UserID, nil)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "access_tokens" SET`).
					WithArgs(sqlmock.AnyArg(), at.UserID).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.DeleteByUserID(ctx, at.UserID, nil)
				assert.Error(t, err)
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

func TestAccessTokenRepo_FindCredentialByToken(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewAccessTokenRepository(kit.DB, mockCacher)
	mock := kit.DBmock
	ctx := context.Background()
	token := "token"
	userEmail := "email@test.com"
	defaultNilCacheTTLMinute := 67
	id := helper.GenerateID()

	tests := []common.TestStructure{
		{
			Name: "got model.NilKey from cache must result to aborting the next process",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return(model.NilKey, nil)
			},
			Run: func() {
				_, _, err := repo.FindCredentialByToken(ctx, token)
				assert.Error(t, err)

				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "got unmarshall error from data in cache, fallback to db but fails to cache the data",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return("in<valid>UnmarshalError</iis?", nil)
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens"`).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"id", "token", "email"}).AddRow(id, token, userEmail))
				mockCacher.EXPECT().Set(ctx, token, gomock.Any(), gomock.Any()).Times(1).Return(errors.New("redis error")) // <- should not store nil on cache
			},
			Run: func() {
				userToken, user, err := repo.FindCredentialByToken(ctx, token)
				assert.NoError(t, err)

				assert.Equal(t, userToken.Token, token)
				assert.Equal(t, user.Email, userEmail)
			},
		},
		{
			Name: "got unmarshall error from data in cache when trying to fetch cache, fallback to db and all good",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return("in<valid>UnmarshalError</iis?", nil)
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens"`).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"id", "token", "email"}).AddRow(id, token, userEmail))
				mockCacher.EXPECT().Set(ctx, token, gomock.Any(), gomock.Any()).Times(1).Return(nil) // <- should not store nil on cache
			},
			Run: func() {
				userToken, user, err := repo.FindCredentialByToken(ctx, token)
				assert.NoError(t, err)

				assert.Equal(t, userToken.Token, token)
				assert.Equal(t, user.Email, userEmail)
			},
		},
		{
			Name: "failure on redis when trying to fetch cache, fallback to db and all good",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return("", errors.New("err redis"))
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens"`).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"id", "token", "email"}).AddRow(id, token, userEmail))
				mockCacher.EXPECT().Set(ctx, token, gomock.Any(), gomock.Any()).Times(1).Return(nil) // <- should not store nil on cache
			},
			Run: func() {
				userToken, user, err := repo.FindCredentialByToken(ctx, token)
				assert.NoError(t, err)

				assert.Equal(t, userToken.Token, token)
				assert.Equal(t, user.Email, userEmail)
			},
		},
		{
			Name: "failure on redis when trying to fetch cache, fallback to db but fails to cache the data",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return("", errors.New("err redis"))
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens"`).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"id", "token", "email"}).AddRow(id, token, userEmail))
				mockCacher.EXPECT().Set(ctx, token, gomock.Any(), gomock.Any()).Times(1).Return(errors.New("redis error")) // <- should not store nil on cache
			},
			Run: func() {
				userToken, user, err := repo.FindCredentialByToken(ctx, token)
				assert.NoError(t, err)

				assert.Equal(t, userToken.Token, token)
				assert.Equal(t, user.Email, userEmail)
			},
		},
		{
			Name: "failure on redis when trying to fetch cache, fallback to db and got not found then all good",
			MockFn: func() {
				viper.Set("server.auth.access_token_duration_minutes", defaultNilCacheTTLMinute)
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return("", errors.New("err redis"))
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens"`).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"id", "token", "email"}))
				mockCacher.EXPECT().Set(ctx, token, model.NilKey, time.Minute*time.Duration(defaultNilCacheTTLMinute)).Times(1).Return(nil) // <- should store nil on cache
			},
			Run: func() {
				_, _, err := repo.FindCredentialByToken(ctx, token)
				assert.Error(t, err)

				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "failure on redis when trying to fetch cache, fallback to db and got not found then fails to write cache",
			MockFn: func() {
				viper.Set("server.auth.access_token_duration_minutes", defaultNilCacheTTLMinute)
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return("", errors.New("err redis"))
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens"`).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"id", "token", "email"}))
				mockCacher.EXPECT().Set(ctx, token, model.NilKey, time.Minute*time.Duration(defaultNilCacheTTLMinute)).Times(1).Return(errors.New("err redis")) // <- should store nil on cache
			},
			Run: func() {
				_, _, err := repo.FindCredentialByToken(ctx, token)
				assert.Error(t, err)

				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "not found on redis when trying to fetch cache, fallback to db and all good",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return("", redis.Nil)
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens"`).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"id", "token", "email"}).AddRow(id, token, userEmail))
				mockCacher.EXPECT().Set(ctx, token, gomock.Any(), gomock.Any()).Times(1).Return(nil) // <- should not store nil on cache
			},
			Run: func() {
				userToken, user, err := repo.FindCredentialByToken(ctx, token)
				assert.NoError(t, err)

				assert.Equal(t, userToken.Token, token)
				assert.Equal(t, user.Email, userEmail)
			},
		},
		{
			Name: "not found on redis when trying to fetch cache, fallback to db but fails to cache the data",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, token).Times(1).Return("", redis.Nil)
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens"`).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"id", "token", "email"}).AddRow(id, token, userEmail))
				mockCacher.EXPECT().Set(ctx, token, gomock.Any(), gomock.Any()).Times(1).Return(errors.New("redis error")) // <- should not store nil on cache
			},
			Run: func() {
				userToken, user, err := repo.FindCredentialByToken(ctx, token)
				assert.NoError(t, err)

				assert.Equal(t, userToken.Token, token)
				assert.Equal(t, user.Email, userEmail)
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

func TestAccessTokenRepository_DeleteByIDs(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewAccessTokenRepository(kit.DB, mockCacher)
	mock := kit.DBmock
	ctx := context.Background()

	idsSingle := []uuid.UUID{uuid.New()}
	idsMulti := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []common.TestStructure{
		{
			Name: "ok - hard delete - single",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^DELETE FROM .+ WHERE`).
					WithArgs(idsSingle[0]).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DeleteByIDs(ctx, idsSingle, true)
				assert.NoError(t, err)
			},
		},
		{
			Name: "ok - hard delete - multiple",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^DELETE FROM .+ WHERE`).
					WithArgs(idsMulti[0], idsMulti[1]).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DeleteByIDs(ctx, idsMulti, true)
				assert.NoError(t, err)
			},
		},
		{
			Name: "ok - soft delete - single",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "access_tokens" .+ WHERE`).
					WithArgs(sqlmock.AnyArg(), idsSingle[0]).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DeleteByIDs(ctx, idsSingle, false)
				assert.NoError(t, err)
			},
		},
		{
			Name: "ok - soft delete - multiple",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "access_tokens" .+ WHERE`).
					WithArgs(sqlmock.AnyArg(), idsMulti[0], idsMulti[1]).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DeleteByIDs(ctx, idsMulti, false)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err - hard delete",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^DELETE FROM .+ WHERE`).
					WithArgs(idsSingle[0]).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("err db")))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.DeleteByIDs(ctx, idsSingle, true)
				assert.Error(t, err)
			},
		},
		{
			Name: "err - soft delete",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "access_tokens" .+ WHERE`).
					WithArgs(sqlmock.AnyArg(), idsMulti[0], idsMulti[1]).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("err db")))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.DeleteByIDs(ctx, idsMulti, false)
				assert.Error(t, err)
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

func TestAccessTokenRepository_FindByUserID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewAccessTokenRepository(kit.DB, mockCacher)
	mock := kit.DBmock
	ctx := context.Background()

	userID := uuid.New()
	limit := 69
	tid := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "db return error",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens" WHERE`).
					WithArgs(userID).
					WillReturnError(errors.New("err db"))
			},
			Run: func() {
				_, err := repo.FindByUserID(ctx, userID, limit)
				assert.Error(t, err)
			},
		},
		{
			Name: "ok found row > 0",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens" WHERE`).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.FindByUserID(ctx, userID, limit)
				assert.NoError(t, err)

				assert.Equal(t, len(res), 1)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "ok found row = 0",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens" WHERE`).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			Run: func() {
				_, err := repo.FindByUserID(ctx, userID, limit)
				assert.Error(t, err)

				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "gorm return error not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "access_tokens" WHERE`).
					WithArgs(userID).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindByUserID(ctx, userID, limit)
				assert.Error(t, err)

				assert.Equal(t, err, ErrNotFound)
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
