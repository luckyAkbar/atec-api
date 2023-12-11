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

func TestSDPackageRepository_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDPackageRepository(kit.DB)
	mock := kit.DBmock
	ctx := context.Background()
	id := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_packages" WHERE`).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			},
			Run: func() {
				email, err := repo.FindByID(ctx, id, true)
				assert.NoError(t, err)

				assert.Equal(t, email.ID, id)
			},
		},
		{
			Name: "ok-exclude deleted",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_packages" WHERE .+ "test_packages"."deleted_at" IS NULL .+`).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			},
			Run: func() {
				email, err := repo.FindByID(ctx, id, false)
				assert.NoError(t, err)

				assert.Equal(t, email.ID, id)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_packages" WHERE`).
					WithArgs(id).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindByID(ctx, id, true)
				assert.Error(t, err)

				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "db return error",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_packages" WHERE`).
					WithArgs(id).
					WillReturnError(errors.New("err db"))
			},
			Run: func() {
				_, err := repo.FindByID(ctx, id, true)
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "err db")
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

func TestSDPackageRepository_Search(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDPackageRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	id := uuid.New()
	createdAfter := time.Now().Add(time.Hour * -9).UTC()

	tests := []common.TestStructure{
		{
			Name: "ok - include deleted",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_packages"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.SearchSDPackageInput{
					IncludeDeleted: true,
				})

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok - exclude deleted",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_packages" .+ "test_packages"."deleted_at" IS NULL .+`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.SearchSDPackageInput{
					IncludeDeleted: false,
				})

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok1",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_packages" .+ "test_packages"."deleted_at" IS NULL .+`).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchSDPackageInput{
					CreatedBy:      id,
					IncludeDeleted: false,
				}
				res, err := repo.Search(ctx, input)

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok2",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_packages"`).
					WithArgs(id, createdAfter).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchSDPackageInput{
					CreatedBy:      id,
					IncludeDeleted: true,
					CreatedAfter:   createdAfter,
				}
				res, err := repo.Search(ctx, input)

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok3",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_packages"`).
					WithArgs(id, createdAfter).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchSDPackageInput{
					CreatedBy:      id,
					IncludeDeleted: true,
					CreatedAfter:   createdAfter,
					Limit:          10,
				}
				res, err := repo.Search(ctx, input)

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "db err",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_packages"`).
					WithArgs(id, createdAfter).
					WillReturnError(errors.New("err db"))
			},
			Run: func() {
				input := &model.SearchSDPackageInput{
					CreatedBy:      id,
					IncludeDeleted: true,
					CreatedAfter:   createdAfter,
					Limit:          10,
				}
				_, err := repo.Search(ctx, input)

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
