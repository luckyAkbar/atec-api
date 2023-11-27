package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEmailRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewEmailRepository(kit.DB)
	mock := kit.DBmock
	ctx := context.Background()
	email := &model.Email{
		ID:        uuid.New(),
		Subject:   "test subject",
		Body:      "test body",
		To:        pq.StringArray{"test1@gmail.com", "test1@test.com"},
		Cc:        pq.StringArray{"test2@gmail.com", "test2@test.com"},
		Bcc:       pq.StringArray{"test3@gmail.com", "test3@test.com"},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	tests := []common.TestStructure{
		{
			Name: "success",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "emails"`).
					WithArgs(email.ID, email.Subject, email.Body, email.To, email.Cc, email.Bcc, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), email.CreatedAt, email.UpdatedAt, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Create(ctx, email)
				assert.NoError(t, err)
			},
		},
		{
			Name: "failed on database returning error",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "emails"`).
					WithArgs(email.ID, email.Subject, email.Body, email.To, email.Cc, email.Bcc, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), email.CreatedAt, email.UpdatedAt, sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Create(ctx, email)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		tt.MockFn()
		tt.Run()
	}
}

func TestEmailRepository_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewEmailRepository(kit.DB)
	mock := kit.DBmock
	ctx := context.Background()
	id := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "emails" WHERE`).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			},
			Run: func() {
				email, err := repo.FindByID(ctx, id)
				assert.NoError(t, err)

				assert.Equal(t, email.ID, id)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "emails" WHERE`).
					WithArgs(id).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindByID(ctx, id)
				assert.Error(t, err)

				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "db return error",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "emails" WHERE`).
					WithArgs(id).
					WillReturnError(errors.New("err db"))
			},
			Run: func() {
				_, err := repo.FindByID(ctx, id)
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "err db")
			},
		},
	}

	for _, tt := range tests {
		tt.MockFn()
		tt.Run()
	}
}
