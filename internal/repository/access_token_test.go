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
