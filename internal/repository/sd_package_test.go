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

func TestSDPackageRepository_Update(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDPackageRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	templateID := uuid.New()
	packageID := uuid.New()
	now := time.Now().UTC()

	p := &model.SpeechDelayPackage{
		ID:         packageID,
		TemplateID: templateID,
		Name:       "ok",
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
				mock.ExpectExec(`^UPDATE "test_packages" SET`).
					WithArgs(p.TemplateID, p.Name, p.CreatedBy, sqlmock.AnyArg(), p.IsActive, p.IsLocked, p.CreatedAt, sqlmock.AnyArg(), p.DeletedAt, p.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Update(ctx, p, nil)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "test_packages" SET`).
					WithArgs(p.TemplateID, p.Name, p.CreatedBy, sqlmock.AnyArg(), p.IsActive, p.IsLocked, p.CreatedAt, sqlmock.AnyArg(), p.DeletedAt, p.ID).
					WillReturnError(errors.New("err db"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Update(ctx, p, nil)
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

func TestSDPackageRepository_Delete(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDPackageRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	p := &model.SpeechDelayPackage{
		ID:         uuid.New(),
		CreatedBy:  uuid.New(),
		Name:       "name",
		IsActive:   false,
		IsLocked:   false,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		TemplateID: uuid.New(),
		Package:    &model.SDPackage{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE "test_packages" SET`).WithArgs(sqlmock.AnyArg(), p.ID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(p.ID))
				mock.ExpectCommit()
			},
			Run: func() {
				_, err := repo.Delete(ctx, p.ID)
				assert.NoError(t, err)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE "test_packages" SET`).WithArgs(sqlmock.AnyArg(), p.ID).WillReturnError(errors.New("err db"))
				mock.ExpectCommit()
			},
			Run: func() {
				_, err := repo.Delete(ctx, p.ID)
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

func TestSDPackageRepository_UndoDelete(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDPackageRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	p := &model.SpeechDelayPackage{
		ID:         uuid.New(),
		CreatedBy:  uuid.New(),
		Name:       "name",
		IsActive:   false,
		IsLocked:   false,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		TemplateID: uuid.New(),
		Package:    &model.SDPackage{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE "test_packages" SET`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), p.ID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(p.ID))
				mock.ExpectCommit()
			},
			Run: func() {
				_, err := repo.UndoDelete(ctx, p.ID)
				assert.NoError(t, err)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE "test_packages" SET`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), p.ID).WillReturnError(errors.New("err db"))
				mock.ExpectCommit()
			},
			Run: func() {
				_, err := repo.UndoDelete(ctx, p.ID)
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

func TestSDPackageRepository_FindRandomActivePackage(t *testing.T) {
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
				mock.ExpectQuery(`^SELECT .+ FROM "test_packages" WHERE`).
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pack.ID))
			},
			Run: func() {
				res, err := repo.FindRandomActivePackage(ctx)
				assert.NoError(t, err)
				assert.Equal(t, res.ID, pack.ID)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_packages" WHERE`).
					WithArgs(true).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindRandomActivePackage(ctx)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_packages" WHERE`).
					WithArgs(true).
					WillReturnError(errors.New("err db"))
			},
			Run: func() {
				_, err := repo.FindRandomActivePackage(ctx)
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

func TestSDPackage_FindLeastUsedPackageIDByUserID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDPackageRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	userID := uuid.New()
	unusedPackageID := uuid.New()
	leastUsedPackageID := uuid.New()
	usedPackageID1 := uuid.New()
	usedPackageID2 := uuid.New()
	usedPackageID3 := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "failed when trying to find active package",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT "id" FROM "test_packages" WHERE is_active`).
					WillReturnError(errors.New("err"))
			},
			Run: func() {
				_, err := repo.FindLeastUsedPackageIDByUserID(ctx, userID)
				assert.Error(t, err)
			},
		},
		{
			Name: "returning error not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT "id" FROM "test_packages" WHERE is_active`).
					WithArgs(true).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindLeastUsedPackageIDByUserID(ctx, userID)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "if no rows found, must return err not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT "id" FROM "test_packages" WHERE is_active`).
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			Run: func() {
				_, err := repo.FindLeastUsedPackageIDByUserID(ctx, userID)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "failed to find history test result",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT "id" FROM "test_packages" WHERE is_active`).
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
				mock.ExpectQuery(`^SELECT .+ FROM "test_results"`).
					WithArgs(userID).
					WillReturnError(errors.New("err db"))
			},
			Run: func() {
				_, err := repo.FindLeastUsedPackageIDByUserID(ctx, userID)
				assert.Error(t, err)
			},
		},
		{
			Name: "error not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT "id" FROM "test_packages" WHERE is_active`).
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
				mock.ExpectQuery(`^SELECT .+ FROM "test_results"`).
					WithArgs(userID).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindLeastUsedPackageIDByUserID(ctx, userID)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "active package id was found and never used by user",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT "id" FROM "test_packages" WHERE is_active`).
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(unusedPackageID))
				mock.ExpectQuery(`^SELECT .+ FROM "test_results"`).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"package_id", "count"}).AddRow(uuid.New(), 1))
			},
			Run: func() {
				res, err := repo.FindLeastUsedPackageIDByUserID(ctx, userID)
				assert.NoError(t, err)
				assert.Equal(t, res, unusedPackageID)
			},
		},
		{
			Name: "returning the least used package count",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT "id" FROM "test_packages" WHERE is_active`).
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(usedPackageID1).AddRow(usedPackageID2).AddRow(usedPackageID3))
				mock.ExpectQuery(`^SELECT .+ FROM "test_results"`).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows(
						[]string{"package_id", "count"}).
						AddRow(usedPackageID1, 10).
						AddRow(usedPackageID2, 1029).
						AddRow(usedPackageID3, 1229).
						AddRow(leastUsedPackageID, 1),
					)
			},
			Run: func() {
				res, err := repo.FindLeastUsedPackageIDByUserID(ctx, userID)
				assert.NoError(t, err)
				assert.Equal(t, res, leastUsedPackageID)
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

func TestSDPackage_GetTemplateByPackageID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDPackageRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	pid := uuid.New()
	tid := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_templates"`).
					WithArgs(pid).WillReturnError(errors.New("err db"))
			},
			Run: func() {
				_, err := repo.GetTemplateByPackageID(ctx, pid)
				assert.Error(t, err)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_templates"`).
					WithArgs(pid).WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.GetTemplateByPackageID(ctx, pid)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_templates"`).
					WithArgs(pid).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.GetTemplateByPackageID(ctx, pid)
				assert.NoError(t, err)
				assert.Equal(t, res.ID, tid)
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
