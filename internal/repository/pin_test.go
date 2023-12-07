package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/config"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPinRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewPinRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	userID := uuid.New()
	now := time.Now().UTC()
	pin := &model.Pin{
		ID:                uuid.New(),
		Pin:               "pin",
		UserID:            userID,
		ExpiredAt:         time.Now().Add(time.Minute * time.Duration(config.PinExpiryMinutes())).UTC(),
		RemainingAttempts: config.PinMaxRetry(),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "pins"`).WithArgs(pin.ID, pin.Pin, pin.UserID, pin.ExpiredAt, pin.RemainingAttempts, pin.CreatedAt, pin.UpdatedAt, pin.DeletedAt).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Create(ctx, pin, nil)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "pins"`).WithArgs(pin.ID, pin.Pin, pin.UserID, pin.ExpiredAt, pin.RemainingAttempts, pin.CreatedAt, pin.UpdatedAt, pin.DeletedAt).WillReturnError(errors.New("db err"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Create(ctx, pin, nil)
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

func TestPinRepository_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewPinRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	id := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "pins" WHERE`).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			},
			Run: func() {
				pin, err := repo.FindByID(ctx, id)
				assert.NoError(t, err)
				assert.Equal(t, pin.ID, id)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "pins" WHERE`).WithArgs(id).WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindByID(ctx, id)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "db error",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "pins" WHERE`).WithArgs(id).WillReturnError(errors.New("db err"))
			},
			Run: func() {
				_, err := repo.FindByID(ctx, id)
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "db err")
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

func TestPinRepository_DecrementRemainingAttempts(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewPinRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	id := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "pins" SET`).
					WithArgs(sqlmock.AnyArg(), id).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DecrementRemainingAttempts(ctx, id)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "pins" SET`).
					WithArgs(sqlmock.AnyArg(), id).WillReturnError(errors.New("err db"))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.DecrementRemainingAttempts(ctx, id)
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
