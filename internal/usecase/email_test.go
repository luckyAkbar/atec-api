package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/stretchr/testify/assert"
	custerr "github.com/sweet-go/stdlib/error"
)

func TestEmailUsecase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWorkerClient := mock.NewMockWorkerClient(ctrl)
	mockEmailRepo := mock.NewMockEmailRepository(ctrl)

	ctx := context.Background()
	uc := NewEmailUsecase(mockEmailRepo, mockWorkerClient)
	validInput := &model.RegisterEmailInput{
		Subject: "subject",
		Body:    "<html>body</html>",
		To:      []string{"valid.to2@email.com"},
		Cc:      []string{"valid.cc1@email.com", "okemail.22@gmail.com"},
		Bcc:     []string{"valid.cc1@email.com", "okemail.22@gmail.com"},
	}

	tests := []common.TestStructure{
		{
			Name:   "email subject is empty",
			MockFn: func() {},
			Run: func() {
				input := &model.RegisterEmailInput{
					Subject: "",
				}
				_, err := uc.Register(ctx, input)
				assert.Error(t, err)

				custErr, ok := err.(custerr.ErrChain)
				assert.True(t, ok)

				assert.Equal(t, custErr.Code, http.StatusBadRequest)
				assert.Equal(t, custErr.Type, ErrEmailInputInvalid)
			},
		},
		{
			Name:   "body is empty",
			MockFn: func() {},
			Run: func() {
				input := &model.RegisterEmailInput{
					Subject: "ada isiniya",
					Body:    "",
				}
				_, err := uc.Register(ctx, input)
				assert.Error(t, err)

				custErr, ok := err.(custerr.ErrChain)
				assert.True(t, ok)

				assert.Equal(t, custErr.Code, http.StatusBadRequest)
				assert.Equal(t, custErr.Type, ErrEmailInputInvalid)
			},
		},
		{
			Name:   "no receipient",
			MockFn: func() {},
			Run: func() {
				input := &model.RegisterEmailInput{
					Subject: "ada isiniya",
					Body:    "ada juga",
					To:      []string{},
				}
				_, err := uc.Register(ctx, input)
				assert.Error(t, err)

				custErr, ok := err.(custerr.ErrChain)
				assert.True(t, ok)

				assert.Equal(t, custErr.Code, http.StatusBadRequest)
				assert.Equal(t, custErr.Type, ErrEmailInputInvalid)
			},
		},
		{
			Name: "failed to write data to database",
			MockFn: func() {
				mockEmailRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(errors.New("db err"))
			},
			Run: func() {
				_, err := uc.Register(ctx, validInput)
				assert.Error(t, err)

				custErr, ok := err.(custerr.ErrChain)
				assert.True(t, ok)

				assert.Equal(t, custErr.Code, http.StatusInternalServerError)
				assert.Equal(t, custErr.Type, ErrInternal)
			},
		},
		{
			Name: "failed to enqueue email sending to worker",
			MockFn: func() {
				mockEmailRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(nil)
				mockWorkerClient.EXPECT().EnqueueSendEmailTask(ctx, gomock.Any()).Times(1).Return(nil, errors.New("failed to enqueue"))
			},
			Run: func() {
				_, err := uc.Register(ctx, validInput)
				assert.Error(t, err)

				custErr, ok := err.(custerr.ErrChain)
				assert.True(t, ok)

				assert.Equal(t, custErr.Code, http.StatusInternalServerError)
				assert.Equal(t, custErr.Type, ErrInternal)
			},
		},
		{
			Name: "ok",
			MockFn: func() {
				mockEmailRepo.EXPECT().Create(ctx, gomock.Any()).Times(1).Return(nil)
				mockWorkerClient.EXPECT().EnqueueSendEmailTask(ctx, gomock.Any()).Times(1).Return(&asynq.TaskInfo{
					ID: "id",
				}, nil)
			},
			Run: func() {
				_, err := uc.Register(ctx, validInput)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		tt.MockFn()
		tt.Run()
	}
}
