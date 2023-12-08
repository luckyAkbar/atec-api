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
