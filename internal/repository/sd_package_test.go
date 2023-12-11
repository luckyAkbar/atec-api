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

func TestSDPackageRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDPackageRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	now := time.Now().UTC()
	pack := &model.SpeechDelayPackage{
		ID:         uuid.NameSpaceDNS,
		TemplateID: uuid.New(),
		Name:       "name",
		CreatedBy:  uuid.New(),
		Package:    &model.SDPackage{},
		IsActive:   false,
		IsLocked:   false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "test_packages"`).
					WithArgs(pack.ID, pack.TemplateID, pack.Name, pack.CreatedBy, sqlmock.AnyArg(), pack.IsActive, pack.IsLocked, pack.CreatedAt, sqlmock.AnyArg(), pack.DeletedAt).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Create(ctx, pack)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "test_packages"`).
					WithArgs(pack.ID, pack.TemplateID, pack.Name, pack.CreatedBy, sqlmock.AnyArg(), pack.IsActive, pack.IsLocked, pack.CreatedAt, sqlmock.AnyArg(), pack.DeletedAt).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Create(ctx, pack)
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
