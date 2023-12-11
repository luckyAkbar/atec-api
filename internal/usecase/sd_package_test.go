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
			{
				Name: "another valid group name",
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
					{
						Question: "valid question?",
						AnswersAndValue: []model.SDAnswerAndValue{
							{
								Text:  "pilihan pertama",
								Value: 1001,
							},
							{
								Text:  "pilihan kedua, tapi value nya sama",
								Value: 100,
							},
							{
								Text:  "pilihan ketiga, tapi ya begitulah",
								Value: 11,
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
