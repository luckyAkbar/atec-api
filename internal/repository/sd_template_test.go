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

func TestSDTemplateRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTemplateRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	now := time.Now().UTC()

	tem := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "name",
		IsActive:  false,
		IsLocked:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Template: &model.SDTemplate{
			Name:                   "name",
			IndicationThreshold:    1,
			PositiveIndiationText:  "positive",
			NegativeIndicationText: "negative",
			SubGroupDetails: []model.SDTemplateSubGroupDetail{
				{
					Name:              "name",
					QuestionCount:     1,
					AnswerOptionCount: 1,
				},
			},
		},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "test_templates"`).
					WithArgs(tem.ID, tem.CreatedBy, tem.Name, tem.IsActive, tem.IsLocked, tem.CreatedAt, sqlmock.AnyArg(), tem.DeletedAt, sqlmock.AnyArg()).
					//WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tem.ID))
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Create(ctx, tem)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "test_templates"`).
					WithArgs(tem.ID, tem.CreatedBy, tem.Name, tem.IsActive, tem.IsLocked, tem.CreatedAt, sqlmock.AnyArg(), tem.DeletedAt, sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
					//WillReturnError(errors.New("err db"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Create(ctx, tem)
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

func TestSDTemplateRepository_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTemplateRepository(kit.DB)
	mock := kit.DBmock
	ctx := context.Background()
	id := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "test_templates" WHERE`).
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
				mock.ExpectQuery(`^SELECT .+ FROM "test_templates" WHERE`).
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
				mock.ExpectQuery(`^SELECT .+ FROM "test_templates" WHERE`).
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

func TestSDTemplateRepository_Search(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTemplateRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock
	id := uuid.New()
	createdAfter := time.Now().Add(time.Hour * -9).UTC()

	tests := []common.TestStructure{
		{
			Name: "ok - include deleted",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_templates"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.SearchSDTemplateInput{
					IncludeDeleted: true,
				})

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok - exclude deleted",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_templates" .+ "test_templates"."deleted_at" IS NULL .+`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.SearchSDTemplateInput{
					IncludeDeleted: false,
				})

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok1",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "test_templates" .+ "test_templates"."deleted_at" IS NULL .+`).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchSDTemplateInput{
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
				mock.ExpectQuery(`SELECT .+ FROM "test_templates"`).
					WithArgs(id, createdAfter).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchSDTemplateInput{
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
				mock.ExpectQuery(`SELECT .+ FROM "test_templates"`).
					WithArgs(id, createdAfter).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchSDTemplateInput{
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
				mock.ExpectQuery(`SELECT .+ FROM "test_templates"`).
					WithArgs(id, createdAfter).
					WillReturnError(errors.New("err db"))
			},
			Run: func() {
				input := &model.SearchSDTemplateInput{
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

func TestSDTemplateRepository_Update(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewSDTemplateRepository(kit.DB)
	ctx := context.Background()
	mock := kit.DBmock

	te := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "name",
		IsActive:  false,
		IsLocked:  false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Template:  &model.SDTemplate{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "test_templates" SET`).
					WithArgs(te.CreatedBy, te.Name, te.IsActive, te.IsLocked, te.CreatedAt, sqlmock.AnyArg(), te.DeletedAt, sqlmock.AnyArg(), te.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Update(ctx, te, nil)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "test_templates" SET`).
					WithArgs(te.CreatedBy, te.Name, te.IsActive, te.IsLocked, te.CreatedAt, sqlmock.AnyArg(), te.DeletedAt, sqlmock.AnyArg(), te.ID).
					WillReturnError(errors.New("err db"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Update(ctx, te, nil)
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
