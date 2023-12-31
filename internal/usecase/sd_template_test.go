package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSDTemplateUsecase_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)

	au := model.AuthUser{
		UserID:      uuid.New(),
		AccessToken: "token",
		Role:        model.RoleAdmin,
	}

	ctx := model.SetUserToCtx(context.Background(), au)

	uc := NewSDTemplateUsecase(mockSDTemplateRepo)

	input := &model.SDTemplate{
		Name:                   "name",
		IndicationThreshold:    10,
		PositiveIndiationText:  "pos",
		NegativeIndicationText: "neg",
		SubGroupDetails: []model.SDTemplateSubGroupDetail{
			{
				Name:              "ok",
				QuestionCount:     99,
				AnswerOptionCount: 12,
			},
		},
	}

	tests := []common.TestStructure{
		{
			// invalid input has many edge cases, and will be tested on the struct Validate function on model repository
			Name:   "invalid input",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.Create(ctx, &model.SDTemplate{})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateInputInvalid)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
			},
		},
		{
			Name: "db err when insert data",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Create(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(nil)
			},
			Run: func() {
				res, cerr := uc.Create(ctx, input)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.Template.SubGroupDetails, input.SubGroupDetails)
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

func TestSDTemplateUsecase_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDTemplateUsecase(mockSDTemplateRepo)

	id := uuid.New()
	now := time.Now()

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
			Name: "ok found",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(tem, nil)
			},
			Run: func() {
				res, cerr := uc.FindByID(ctx, id)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, tem.ToRESTResponse().ID)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.FindByID(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "db err",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.FindByID(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
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

func TestSDTemplateUsecase_Search(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDTemplateUsecase(mockSDTemplateRepo)

	trueVal := true
	falseVal := false

	input := &model.SearchSDTemplateInput{
		CreatedBy:      uuid.New(),
		CreatedAfter:   time.Now().Add(time.Hour * -10).UTC(),
		IsActive:       &trueVal,
		IsLocked:       &falseVal,
		IncludeDeleted: false,
		Limit:          10,
		Offset:         0,
	}

	tests := []common.TestStructure{
		{
			Name: "repo err",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().Search(ctx, input).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Search(ctx, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "repo return 0 array",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().Search(ctx, input).Times(1).Return([]*model.SpeechDelayTemplate{}, nil)
			},
			Run: func() {
				res, cerr := uc.Search(ctx, input)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.Count, 0)
				assert.Equal(t, len(res.Templates), 0)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().Search(ctx, input).Times(1).Return([]*model.SpeechDelayTemplate{
					{
						ID: uuid.New(),
					},
					{
						ID: uuid.New(),
					},
				}, nil)
			},
			Run: func() {
				res, cerr := uc.Search(ctx, input)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.Count, 2)
				assert.Equal(t, len(res.Templates), 2)
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

func TestSDTemplateUsecase_Update(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDTemplateUsecase(mockSDTemplateRepo)

	id := uuid.New()

	input := &model.SDTemplate{
		Name:                   "name",
		IndicationThreshold:    10,
		PositiveIndiationText:  "pos",
		NegativeIndicationText: "neg",
		SubGroupDetails: []model.SDTemplateSubGroupDetail{
			{
				Name:              "ok",
				QuestionCount:     99,
				AnswerOptionCount: 12,
			},
		},
	}

	now := time.Now().UTC()

	tem := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "name",
		IsActive:  false,
		IsLocked:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Template:  input,
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().Update(ctx, tem, nil).Times(1).Return(nil)
			},
			Run: func() {
				res, cerr := uc.Update(ctx, id, input)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, tem.ToRESTResponse())

			},
		},
		{
			Name: "invalid input",
			MockFn: func() {
			},
			Run: func() {
				_, cerr := uc.Update(ctx, id, &model.SDTemplate{})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateInputInvalid)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)
			},
		},
		{
			Name: "template not found",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Update(ctx, id, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "template is locked",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsActive:  false,
					IsLocked:  true,
					CreatedAt: now,
					UpdatedAt: now,
					Template:  input,
				}, nil)
			},
			Run: func() {
				_, cerr := uc.Update(ctx, id, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateAlreadyLocked)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "template is already active",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsActive:  true,
					IsLocked:  false,
					CreatedAt: now,
					UpdatedAt: now,
					Template:  input,
				}, nil)
			},
			Run: func() {
				_, cerr := uc.Update(ctx, id, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateIsAlreadyActive)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "failed to update",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().Update(ctx, tem, nil).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Update(ctx, id, input)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
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

func TestSDTemplateUsecase_Delete(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDTemplateUsecase(mockSDTemplateRepo)

	id := uuid.New()

	now := time.Now().UTC()
	tem := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "name",
		IsActive:  false,
		IsLocked:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Template:  &model.SDTemplate{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().Delete(ctx, tem.ID).Times(1).Return(tem, nil)
			},
			Run: func() {
				res, cerr := uc.Delete(ctx, id)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, tem.ToRESTResponse())
			},
		},
		{
			Name: "not found or maybe already deleted",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Delete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "failed to find the sd template",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.Delete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "template locked",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsActive:  true,
					IsLocked:  true,
					CreatedAt: now,
					UpdatedAt: now,
					Template:  &model.SDTemplate{},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.Delete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateAlreadyLocked)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "failed to delete",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().Delete(ctx, tem.ID).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.Delete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
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

func TestSDTemplateUsecase_UndoDelete(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDTemplateUsecase(mockSDTemplateRepo)

	id := uuid.New()

	now := time.Now().UTC()
	tem := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "name",
		IsActive:  false,
		IsLocked:  false,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{Time: time.Now().UTC(), Valid: true},
		Template:  &model.SDTemplate{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().UndoDelete(ctx, tem.ID).Times(1).Return(tem, nil)
			},
			Run: func() {
				res, cerr := uc.UndoDelete(ctx, id)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, tem.ToRESTResponse())
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.UndoDelete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "failed to find the sd template",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.UndoDelete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "template locked",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(&model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsActive:  true,
					IsLocked:  true,
					CreatedAt: now,
					UpdatedAt: now,
					Template:  &model.SDTemplate{},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.UndoDelete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateAlreadyLocked)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name:   "ok - template is not deleted",
			MockFn: func() {},
			Run: func() {
				notDeleted := &model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsActive:  true,
					IsLocked:  false,
					CreatedAt: now,
					UpdatedAt: now,
					DeletedAt: gorm.DeletedAt{Valid: false},
					Template:  &model.SDTemplate{},
				}
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(notDeleted, nil)

				res, cerr := uc.UndoDelete(ctx, id)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, notDeleted.ToRESTResponse())
			},
		},
		{
			Name: "failed to delete",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().UndoDelete(ctx, tem.ID).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.UndoDelete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
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

func TestSDTemplateUsecase_ChangeSDTemplateActiveStatus(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDTemplateUsecase(mockSDTemplateRepo)
	id := uuid.New()

	now := time.Now().UTC()
	activeTemplate := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "name",
		IsLocked:  true,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Template:  &model.SDTemplate{},
	}

	inactiveTemplate := &model.SpeechDelayTemplate{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "name",
		IsLocked:  true,
		IsActive:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Template:  &model.SDTemplate{},
	}

	tests := []common.TestStructure{
		{
			Name: "failed to find sd template",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.ChangeSDTemplateActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "template not found",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.ChangeSDTemplateActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "template status is the same with input (active = true)",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(activeTemplate, nil)
			},
			Run: func() {
				res, cerr := uc.ChangeSDTemplateActiveStatus(ctx, id, true)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, activeTemplate.ToRESTResponse())
			},
		},
		{
			Name: "template status is the same with input (active = false)",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(inactiveTemplate, nil)
			},
			Run: func() {
				res, cerr := uc.ChangeSDTemplateActiveStatus(ctx, id, false)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, inactiveTemplate.ToRESTResponse())
			},
		},
		{
			Name: "db failure on deactivating template",
			MockFn: func() {

			},
			Run: func() {
				tem := &model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsLocked:  true,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
					Template:  &model.SDTemplate{},
				}

				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().Update(ctx, gomock.Any(), nil).Return(errors.New("err db"))

				_, cerr := uc.ChangeSDTemplateActiveStatus(ctx, id, false)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "ok: deactivating template",
			MockFn: func() {
			},
			Run: func() {
				tem := &model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsLocked:  true,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
					Template:  &model.SDTemplate{},
				}

				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().Update(ctx, gomock.Any(), nil).Return(nil)

				res, cerr := uc.ChangeSDTemplateActiveStatus(ctx, id, false)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, tem.ToRESTResponse().ID)
			},
		},
		{
			Name: "failed activating: failure on full validation",
			MockFn: func() {
			},
			Run: func() {
				tem := &model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsLocked:  true,
					IsActive:  false,
					CreatedAt: now,
					UpdatedAt: now,
					Template:  &model.SDTemplate{},
				}

				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(tem, nil)

				_, cerr := uc.ChangeSDTemplateActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateCantBeActivated)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "ok: template activated",
			MockFn: func() {
			},
			Run: func() {
				tem := &model.SpeechDelayTemplate{
					ID:        uuid.New(),
					CreatedBy: uuid.New(),
					Name:      "name",
					IsLocked:  true,
					IsActive:  false,
					CreatedAt: now,
					UpdatedAt: now,
					Template: &model.SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    10,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []model.SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     10,
								AnswerOptionCount: 3,
							},
							{
								Name:              "okeh juga",
								QuestionCount:     10,
								AnswerOptionCount: 5,
							},
						},
					},
				}

				mockSDTemplateRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(tem, nil)
				mockSDTemplateRepo.EXPECT().Update(ctx, gomock.Any(), nil).Return(nil)

				res, cerr := uc.ChangeSDTemplateActiveStatus(ctx, id, true)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, tem.ToRESTResponse().ID)
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
