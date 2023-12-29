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
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

func TestSDTestResultRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTestResultRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	now := time.Now().UTC()
	tt := &model.SDTest{
		ID:        uuid.New(),
		PackageID: uuid.New(),
		OpenUntil: now,
		SubmitKey: "key",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "test_results"`).
					WithArgs(tt.ID, tt.PackageID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), tt.OpenUntil, tt.SubmitKey, tt.CreatedAt, tt.UpdatedAt, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Create(ctx, tt, nil)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "test_results"`).
					WithArgs(tt.ID, tt.PackageID, tt.UserID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), tt.OpenUntil, tt.SubmitKey, tt.CreatedAt, tt.UpdatedAt, sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Create(ctx, tt, nil)
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

func TestSDTestResultRepository_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTestResultRepository(kit.DB)
	mock := kit.DBmock
	ctx := context.Background()
	id := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
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
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
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
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
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
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestSDTestResultRepository_Update(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTestResultRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	now := time.Now().UTC()

	p := &model.SDTest{
		ID:        uuid.New(),
		PackageID: uuid.New(),
		UserID:    uuid.NullUUID{UUID: uuid.New(), Valid: true},
		Answer: model.SDTestAnswer{
			TestAnswers: []*model.TestAnswer{},
		},
		Result:     model.SDTestResult{},
		FinishedAt: null.NewTime(now, true),
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "test_results" SET`).
					WithArgs(p.PackageID, p.UserID, sqlmock.AnyArg(), sqlmock.AnyArg(), p.FinishedAt, p.OpenUntil, p.SubmitKey, p.CreatedAt, sqlmock.AnyArg(), sqlmock.AnyArg(), p.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
					//WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(p.ID))
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
				mock.ExpectExec(`^UPDATE "test_results" SET`).
					WithArgs(p.PackageID, p.UserID, sqlmock.AnyArg(), sqlmock.AnyArg(), p.FinishedAt, p.OpenUntil, p.SubmitKey, p.CreatedAt, sqlmock.AnyArg(), sqlmock.AnyArg(), p.ID).
					WillReturnError(errors.New("err db"))
					//WillReturnError(errors.New("err db"))
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

func TestSDTestResultRepository_Search(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTestResultRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	tid := uuid.New()
	userID := uuid.New()
	pid := uuid.New()
	ca := time.Now().Add(time.Hour * -1).UTC()

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.ViewHistoriesInput{
					UserID: uuid.NullUUID{UUID: userID, Valid: true},
				})
				assert.NoError(t, err)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE finished_at IS NOT NULL AND "test_results"."deleted_at" IS NULL ORDER BY created_at DESC`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.ViewHistoriesInput{})
				assert.NoError(t, err)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
					WithArgs(userID, pid).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.ViewHistoriesInput{
					UserID:    uuid.NullUUID{UUID: userID, Valid: true},
					PackageID: uuid.NullUUID{UUID: pid, Valid: true},
				})
				assert.NoError(t, err)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
					WithArgs(userID, pid, ca).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.ViewHistoriesInput{
					UserID:       uuid.NullUUID{UUID: userID, Valid: true},
					PackageID:    uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter: null.NewTime(ca, true),
				})
				assert.NoError(t, err)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
					WithArgs(userID, pid, ca).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: userID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(ca, true),
					IncludeUnfinished: true,
				})
				assert.NoError(t, err)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE .+ finished_at IS NOT NULL`).
					WithArgs(userID, pid, ca).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: userID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(ca, true),
					IncludeUnfinished: false,
				})
				assert.NoError(t, err)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
					WithArgs(userID, pid, ca).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: userID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(ca, true),
					IncludeUnfinished: false,
					IncludeDeleted:    true,
				})
				assert.NoError(t, err)
				assert.Equal(t, res[0].ID, tid)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_results" WHERE`).
					WithArgs(userID, pid, ca).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tid))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.ViewHistoriesInput{
					UserID:            uuid.NullUUID{UUID: userID, Valid: true},
					PackageID:         uuid.NullUUID{UUID: pid, Valid: true},
					CreatedAfter:      null.NewTime(ca, true),
					IncludeUnfinished: false,
					IncludeDeleted:    true,
					Limit:             10,
					Offset:            11,
				})
				assert.NoError(t, err)
				assert.Equal(t, res[0].ID, tid)
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

func TestSDTestResultRepository_Statistic(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTestResultRepository(kit.DB)
	ctx := context.Background()
	uid := uuid.New()
	mock := kit.DBmock

	tests := []common.TestStructure{
		{
			Name: "db err",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM test_results`).WithArgs(uid).WillReturnError(errors.New("err db"))
			},
			Run: func() {
				_, err := repo.Statistic(ctx, uid)
				assert.Error(t, err)
			},
		},
		{
			Name: "0 data returned must return errnotfound",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM test_results`).WithArgs(uid).WillReturnRows(sqlmock.NewRows([]string{"template_id"}))
			},
			Run: func() {
				_, err := repo.Statistic(ctx, uid)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM test_results`).WithArgs(uid).WillReturnRows(sqlmock.NewRows([]string{"template_id"}).AddRow(uuid.New()))
			},
			Run: func() {
				_, err := repo.Statistic(ctx, uid)
				assert.NoError(t, err)
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
