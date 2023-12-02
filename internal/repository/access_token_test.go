package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEAccessTokenRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewAccessTokenRepository(kit.DB)
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

	repo := NewAccessTokenRepository(kit.DB)
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

	repo := NewAccessTokenRepository(kit.DB)
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
