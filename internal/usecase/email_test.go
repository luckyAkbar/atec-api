package usecase

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/luckyAkbar/atec-api/internal/common"
	commonMock "github.com/luckyAkbar/atec-api/internal/common/mock"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/stretchr/testify/assert"
	custerr "github.com/sweet-go/stdlib/error"
)

func TestEmailUsecase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWorkerClient := mock.NewMockWorkerClient(ctrl)
	mockEmailRepo := mock.NewMockEmailRepository(ctrl)
	mockSharedCryptor := commonMock.NewMockSharedCryptor(ctrl)

	ctx := context.Background()
	uc := NewEmailUsecase(mockEmailRepo, mockWorkerClient, mockSharedCryptor)

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
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}
