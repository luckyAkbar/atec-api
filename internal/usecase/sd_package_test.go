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

func TestSDPackageUsecase_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	sdpRepo := mock.NewMockSDPackageRepository(ctrl)
	sdtRepo := mock.NewMockSDTemplateRepository(ctrl)
	uc := NewSDPackageUsecase(sdpRepo, sdtRepo)

	ctx := model.SetUserToCtx(context.Background(), model.AuthUser{
		UserID: uuid.New(),
	})

	validInput := &model.SDPackage{
		PackageName: "valid package name",
		TemplateID:  uuid.New(),
		SubGroupDetails: []model.SDSubGroupDetail{
			{
				Name: "valid name",
				QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
					{
						Question: "valid question?",
						AnswersAndValue: []model.SDAnswerAndValue{
							{
								Text:  "pilihan pertama",
								Value: 99,
							},
							{
								Text:  "pilihan kedua, tapi value nya sama",
								Value: 100,
							},
						},
					},
				},
			},
		},
	}

	now := time.Now().UTC()
	inactiveTem := &model.SpeechDelayTemplate{
		ID:        validInput.TemplateID,
		CreatedBy: uuid.New(),
		Name:      "name",
		IsLocked:  false,
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

	activetem := &model.SpeechDelayTemplate{
		ID:        validInput.TemplateID,
		CreatedBy: uuid.New(),
		Name:      "name",
		IsLocked:  false,
		IsActive:  true,
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

	tests := []common.TestStructure{
		{
			// no need to cover all the cases to trigger the validation error
			// already done on model test
			Name:   "input invalid",
			MockFn: func() {},
			Run: func() {
				_, cerr := uc.Create(ctx, &model.SDPackage{})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDPackageInputInvalid)
			},
		},
		{
			Name: "failed to find template",
			MockFn: func() {
				sdtRepo.EXPECT().FindByID(ctx, validInput.TemplateID, false).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Create(ctx, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "template not found",
			MockFn: func() {
				sdtRepo.EXPECT().FindByID(ctx, validInput.TemplateID, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Create(ctx, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
			},
		},
		{
			Name: "template not activated",
			MockFn: func() {
				sdtRepo.EXPECT().FindByID(ctx, validInput.TemplateID, false).Times(1).Return(inactiveTem, nil)
			},
			Run: func() {
				_, cerr := uc.Create(ctx, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateIsDeactivated)
			},
		},
		{
			Name: "failed to create sd package",
			MockFn: func() {
				sdtRepo.EXPECT().FindByID(ctx, validInput.TemplateID, false).Times(1).Return(activetem, nil)
				sdpRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Create(ctx, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				sdtRepo.EXPECT().FindByID(ctx, validInput.TemplateID, false).Times(1).Return(activetem, nil)
				sdpRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(nil)
			},
			Run: func() {
				res, cerr := uc.Create(ctx, validInput)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.TemplateID, activetem.ID)
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

func TestSDPackageUsecase_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDPackageRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDPackageUsecase(mockSDPackageRepo, mockSDTemplateRepo)

	id := uuid.New()
	now := time.Now()

	pack := &model.SpeechDelayPackage{
		ID:         id,
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
			Name: "ok found",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(pack, nil)
			},
			Run: func() {
				res, cerr := uc.FindByID(ctx, id)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, pack.ToRESTResponse().ID)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(nil, repository.ErrNotFound)
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
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(nil, errors.New("err db"))
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

func TestSDPackageUsecase_Search(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDPackageRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDPackageUsecase(mockSDPackageRepo, mockSDTemplateRepo)

	trueVal := true
	falseVal := false

	input := &model.SearchSDPackageInput{
		CreatedBy:      uuid.New(),
		CreatedAfter:   time.Now().Add(time.Hour * -10).UTC(),
		IsActive:       &trueVal,
		IsLocked:       &falseVal,
		IncludeDeleted: false,
		Limit:          10,
		Offset:         0,
		TemplateID:     uuid.New(),
	}

	tests := []common.TestStructure{
		{
			Name: "repo err",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().Search(ctx, input).Times(1).Return(nil, errors.New("err db"))
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
				mockSDPackageRepo.EXPECT().Search(ctx, input).Times(1).Return([]*model.SpeechDelayPackage{}, nil)
			},
			Run: func() {
				res, cerr := uc.Search(ctx, input)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.Count, 0)
				assert.Equal(t, len(res.Packages), 0)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().Search(ctx, input).Times(1).Return([]*model.SpeechDelayPackage{
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
				assert.Equal(t, len(res.Packages), 2)
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

func TestSDPackageUsecase_Update(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	mockSDPackageRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	uc := NewSDPackageUsecase(mockSDPackageRepo, mockSDTemplateRepo)

	templateID := uuid.New()
	packageID := uuid.New()
	now := time.Now().UTC()

	validInput := &model.SDPackage{
		PackageName: "valid package name",
		TemplateID:  templateID,
		SubGroupDetails: []model.SDSubGroupDetail{
			{
				Name: "valid name",
				QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
					{
						Question: "valid question?",
						AnswersAndValue: []model.SDAnswerAndValue{
							{
								Text:  "pilihan pertama",
								Value: 99,
							},
							{
								Text:  "pilihan kedua, tapi value nya sama",
								Value: 100,
							},
						},
					},
				},
			},
		},
	}

	pack := &model.SpeechDelayPackage{
		ID:         packageID,
		TemplateID: templateID,
		Name:       "ok",
		CreatedBy:  uuid.New(),
		Package:    validInput,
		IsActive:   false,
		IsLocked:   false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	activeTem := &model.SpeechDelayTemplate{
		ID:        templateID,
		CreatedBy: uuid.New(),
		Name:      "name",
		IsActive:  true,
		IsLocked:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Template:  &model.SDTemplate{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(activeTem, nil)
				mockSDPackageRepo.EXPECT().FindByID(ctx, packageID, false).Times(1).Return(pack, nil)
				mockSDPackageRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(nil)
			},
			Run: func() {
				res, cerr := uc.Update(ctx, packageID, validInput)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.TemplateID, templateID)
			},
		},
		{
			// the full validation testing cases are in model test
			Name: "invalid input",
			MockFn: func() {
			},
			Run: func() {
				_, cerr := uc.Update(ctx, packageID, &model.SDPackage{})
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDPackageInputInvalid)
				assert.Equal(t, cerr.Code, http.StatusBadRequest)

			},
		},
		{
			Name: "template not found",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Update(ctx, packageID, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "template repo return error",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Update(ctx, packageID, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "inactive template",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(&model.SpeechDelayTemplate{IsActive: false}, nil)
			},
			Run: func() {
				_, cerr := uc.Update(ctx, packageID, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateIsDeactivated)
				assert.Equal(t, cerr.Code, http.StatusForbidden)

			},
		},
		{
			Name: "package not found",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(activeTem, nil)
				mockSDPackageRepo.EXPECT().FindByID(ctx, packageID, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Update(ctx, packageID, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "package not found",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(activeTem, nil)
				mockSDPackageRepo.EXPECT().FindByID(ctx, packageID, false).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Update(ctx, packageID, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "package already locked",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(activeTem, nil)
				mockSDPackageRepo.EXPECT().FindByID(ctx, packageID, false).Times(1).Return(&model.SpeechDelayPackage{IsLocked: true}, nil)
			},
			Run: func() {
				_, cerr := uc.Update(ctx, packageID, validInput)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDPackageAlreadyLocked)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "failed to update",
			MockFn: func() {
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(activeTem, nil)
				mockSDPackageRepo.EXPECT().FindByID(ctx, packageID, false).Times(1).Return(pack, nil)
				mockSDPackageRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.Update(ctx, packageID, validInput)
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

func TestSDPackageUsecase_Delete(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDPackageRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	uc := NewSDPackageUsecase(mockSDPackageRepo, nil)

	id := uuid.New()
	now := time.Now().UTC()

	p := &model.SpeechDelayPackage{
		ID:         uuid.New(),
		CreatedBy:  uuid.New(),
		Name:       "name",
		IsActive:   false,
		IsLocked:   false,
		CreatedAt:  now,
		UpdatedAt:  now,
		TemplateID: uuid.New(),
		Package:    &model.SDPackage{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(p, nil)
				mockSDPackageRepo.EXPECT().Delete(ctx, p.ID).Times(1).Return(p, nil)
			},
			Run: func() {
				res, cerr := uc.Delete(ctx, id)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, p.ToRESTResponse())
			},
		},
		{
			Name: "not found or maybe already deleted",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.Delete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "failed to find the sd package",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.Delete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "package locked",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:         uuid.New(),
					CreatedBy:  uuid.New(),
					Name:       "name",
					IsActive:   true,
					IsLocked:   true,
					CreatedAt:  now,
					UpdatedAt:  now,
					TemplateID: uuid.New(),
					Package:    &model.SDPackage{},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.Delete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDPackageAlreadyLocked)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "failed to delete",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(p, nil)
				mockSDPackageRepo.EXPECT().Delete(ctx, p.ID).Times(1).Return(nil, errors.New("db err"))
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

func TestSDPackageUsecase_UndoDelete(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDPackageRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	uc := NewSDPackageUsecase(mockSDPackageRepo, nil)

	id := uuid.New()

	now := time.Now().UTC()
	pack := &model.SpeechDelayPackage{
		ID:         id,
		CreatedBy:  uuid.New(),
		Name:       "name",
		IsActive:   false,
		IsLocked:   false,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  gorm.DeletedAt{Time: time.Now().UTC(), Valid: true},
		TemplateID: uuid.New(),
		Package:    &model.SDPackage{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(pack, nil)
				mockSDPackageRepo.EXPECT().UndoDelete(ctx, pack.ID).Times(1).Return(pack, nil)
			},
			Run: func() {
				res, cerr := uc.UndoDelete(ctx, id)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, pack.ToRESTResponse())
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.UndoDelete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "failed to find the sd package",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.UndoDelete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "package locked",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(&model.SpeechDelayPackage{
					ID:         uuid.New(),
					CreatedBy:  uuid.New(),
					Name:       "name",
					IsActive:   true,
					IsLocked:   true,
					CreatedAt:  now,
					UpdatedAt:  now,
					TemplateID: uuid.New(),
					Package:    &model.SDPackage{},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.UndoDelete(ctx, id)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDPackageAlreadyLocked)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name:   "ok - template is not deleted",
			MockFn: func() {},
			Run: func() {
				notDeleted := &model.SpeechDelayPackage{
					ID:         uuid.New(),
					CreatedBy:  uuid.New(),
					Name:       "name",
					IsActive:   true,
					IsLocked:   false,
					CreatedAt:  now,
					UpdatedAt:  now,
					DeletedAt:  gorm.DeletedAt{Valid: false},
					TemplateID: uuid.New(),
					Package:    &model.SDPackage{},
				}
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(notDeleted, nil)

				res, cerr := uc.UndoDelete(ctx, id)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res, notDeleted.ToRESTResponse())
			},
		},
		{
			Name: "failed to delete",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, true).Times(1).Return(pack, nil)
				mockSDPackageRepo.EXPECT().UndoDelete(ctx, pack.ID).Times(1).Return(nil, errors.New("db err"))
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

func TestSDPackageUsecase_ChangeSDPackageActiveStatus(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	ctx := context.Background()

	mockSDPackageRepo := mock.NewMockSDPackageRepository(kit.Ctrl)
	mockSDTemplateRepo := mock.NewMockSDTemplateRepository(kit.Ctrl)
	uc := NewSDPackageUsecase(mockSDPackageRepo, mockSDTemplateRepo)

	id := uuid.New()
	templateID := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "db err when find package by id",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "package was not found",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "package was already active",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       id,
					IsActive: true,
				}, nil)
			},
			Run: func() {
				res, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, id)
			},
		},
		{
			Name: "package was already deactivated",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       id,
					IsActive: false,
				}, nil)
			},
			Run: func() {
				res, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, false)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, id)
			},
		},
		{
			Name: "failure to update when just deactivating",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       id,
					IsActive: true,
				}, nil)
				mockSDPackageRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, false)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "success deactivating",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:       id,
					IsActive: true,
				}, nil)
				mockSDPackageRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(nil)
			},
			Run: func() {
				res, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, false)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, id)
			},
		},
		{
			Name: "failed when fetching template data",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:         id,
					TemplateID: templateID,
					IsActive:   false,
				}, nil)
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(nil, errors.New("db err"))
			},
			Run: func() {
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "somehow the template is not found",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:         id,
					TemplateID: templateID,
					IsActive:   false,
				}, nil)
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrResourceNotFound)
				assert.Equal(t, cerr.Code, http.StatusNotFound)
			},
		},
		{
			Name: "safety check: template is still inactive",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:         id,
					TemplateID: templateID,
					IsActive:   false,
				}, nil)
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(&model.SpeechDelayTemplate{
					IsActive: false,
				}, nil)
			},
			Run: func() {
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateIsDeactivated)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "safety check: template was deleted",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:         id,
					TemplateID: templateID,
					IsActive:   false,
				}, nil)
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(&model.SpeechDelayTemplate{
					IsActive:  true,
					DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
				}, nil)
			},
			Run: func() {
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDTemplateIsDeactivated)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			// no need to cover all the edge cases here. The testing on FullValidation are on the model package
			Name: "fail on package full validation",
			MockFn: func() {
				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(&model.SpeechDelayPackage{
					ID:         id,
					TemplateID: templateID,
					IsActive:   false,
					Package:    &model.SDPackage{},
				}, nil)
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(&model.SpeechDelayTemplate{
					IsActive: true,
				}, nil)
			},
			Run: func() {
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrSDPackageCantBeActivated)
				assert.Equal(t, cerr.Code, http.StatusForbidden)
			},
		},
		{
			Name: "failure on update",
			MockFn: func() {

			},
			Run: func() {
				validPackage := &model.SpeechDelayPackage{
					ID:         id,
					TemplateID: templateID,
					Name:       "basing",
					CreatedBy:  uuid.New(),
					Package: &model.SDPackage{
						PackageName: "valid package name",
						TemplateID:  uuid.New(),
						SubGroupDetails: []model.SDSubGroupDetail{
							{
								Name: "okelah",
								QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
									{
										Question: "valid question?",
										AnswersAndValue: []model.SDAnswerAndValue{
											{
												Text:  "pilihan pertama",
												Value: 1,
											},
											{
												Text:  "pilihan kedua, tapi value nya sama",
												Value: 2,
											},

											{
												Text:  "pilihan ketigax",
												Value: 3,
											},
										},
									},
								},
							},
						},
					},
				}

				validTemplate := &model.SpeechDelayTemplate{
					IsActive: true,
					Template: &model.SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []model.SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     1,
								AnswerOptionCount: 3,
							},
						},
					},
				}

				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(validPackage, nil)
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(validTemplate, nil)
				mockSDPackageRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(errors.New("err db"))
				_, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.Error(t, cerr.Type)
				assert.Equal(t, cerr.Type, ErrInternal)
				assert.Equal(t, cerr.Code, http.StatusInternalServerError)
			},
		},
		{
			Name: "ok",
			MockFn: func() {

			},
			Run: func() {
				validPackage := &model.SpeechDelayPackage{
					ID:         id,
					TemplateID: templateID,
					Name:       "basing",
					CreatedBy:  uuid.New(),
					Package: &model.SDPackage{
						PackageName: "valid package name",
						TemplateID:  uuid.New(),
						SubGroupDetails: []model.SDSubGroupDetail{
							{
								Name: "okelah",
								QuestionAndAnswerLists: []model.SDQuestionAndAnswers{
									{
										Question: "valid question?",
										AnswersAndValue: []model.SDAnswerAndValue{
											{
												Text:  "pilihan pertama",
												Value: 1,
											},
											{
												Text:  "pilihan kedua, tapi value nya sama",
												Value: 2,
											},

											{
												Text:  "pilihan ketigax",
												Value: 3,
											},
										},
									},
								},
							},
						},
					},
				}

				validTemplate := &model.SpeechDelayTemplate{
					IsActive: true,
					Template: &model.SDTemplate{
						Name:                   "ok",
						IndicationThreshold:    2,
						PositiveIndiationText:  "ok",
						NegativeIndicationText: "ok jg",
						SubGroupDetails: []model.SDTemplateSubGroupDetail{
							{
								Name:              "okelah",
								QuestionCount:     1,
								AnswerOptionCount: 3,
							},
						},
					},
				}

				mockSDPackageRepo.EXPECT().FindByID(ctx, id, false).Times(1).Return(validPackage, nil)
				mockSDTemplateRepo.EXPECT().FindByID(ctx, templateID, false).Times(1).Return(validTemplate, nil)
				mockSDPackageRepo.EXPECT().Update(ctx, gomock.Any(), nil).Times(1).Return(nil)

				res, cerr := uc.ChangeSDPackageActiveStatus(ctx, id, true)
				assert.NoError(t, cerr.Type)
				assert.Equal(t, res.ID, validPackage.ID)
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
