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

func TestUserRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewUserRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	now := time.Now().UTC()
	user := &model.User{
		ID:        uuid.New(),
		Email:     "email",
		Password:  "password",
		Username:  "username",
		IsActive:  false,
		Role:      model.RoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "users"`).WithArgs(
					user.ID,
					user.Email,
					user.Password,
					user.Username,
					user.IsActive,
					user.Role,
					user.CreatedAt,
					user.UpdatedAt,
					user.DeletedAt,
				).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Create(ctx, user, nil)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "pins"`).WithArgs(
					user.ID,
					user.Email,
					user.Password,
					user.Username,
					user.IsActive,
					user.Role,
					user.CreatedAt,
					user.UpdatedAt,
					user.DeletedAt,
				).WillReturnError(errors.New("db err"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Create(ctx, user, nil)
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

func TestUserRepository_FindByEmail(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewUserRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	id := uuid.New()
	email := "email@gmail.com"

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(email).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			},
			Run: func() {
				user, err := repo.FindByEmail(ctx, email)
				assert.NoError(t, err)
				assert.Equal(t, user.ID, id)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(email).WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindByEmail(ctx, email)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "db error",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(email).WillReturnError(errors.New("db err"))
			},
			Run: func() {
				_, err := repo.FindByEmail(ctx, email)
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
