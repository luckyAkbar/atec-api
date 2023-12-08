package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/stretchr/testify/assert"
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
