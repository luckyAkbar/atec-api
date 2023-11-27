package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/sweet-go/stdlib/mail"
	mailMock "github.com/sweet-go/stdlib/mail/mock"
	workerPkg "github.com/sweet-go/stdlib/worker"
	"golang.org/x/time/rate"
)

func TestWorker_HandleSendEmail(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	mockMailUtility := mailMock.NewMockUtility(ctrl)
	mockMailRepo := mock.NewMockEmailRepository(ctrl)

	normalLimiter := rate.NewLimiter(10, 20)
	id := uuid.New()

	payload, err := json.Marshal(id)
	assert.NoError(t, err)

	task := asynq.NewTask(string(model.TaskSendEmail), payload, asynq.Queue(string(workerPkg.PriorityHigh)))
	email := &model.Email{
		ID:      id,
		Subject: "test",
		Body:    "test",
		To:      pq.StringArray{"test"},
		Cc:      pq.StringArray{"test"},
		Bcc:     pq.StringArray{"test"},
	}

	sendEmailInput := &mail.Mail{
		ID:          email.ID.String(),
		To:          email.GenericReceipientsTo(),
		Cc:          email.GenericReceipientsCc(),
		Bcc:         email.GenericReceipientsBcc(),
		HTMLContent: email.Body,
		Subject:     email.Subject,
	}

	taskHandler := newTaskHandler(mockMailUtility, normalLimiter, mockMailRepo)

	tests := []common.TestStructure{
		{
			Name: "failed to find the email data from database, yet don't consider it an error",
			MockFn: func() {
				mockMailRepo.EXPECT().FindByID(ctx, id).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				err := taskHandler.HandleSendEmail(ctx, task)
				assert.NoError(t, err)
			},
		},
		{
			Name: "should return error when database also returning unexpected error",
			MockFn: func() {
				mockMailRepo.EXPECT().FindByID(ctx, id).Times(1).Return(nil, errors.New("unexpected"))
			},
			Run: func() {
				err := taskHandler.HandleSendEmail(ctx, task)
				assert.Error(t, err)
			},
		},
		{
			Name:   "got rate limited error",
			MockFn: func() {},
			Run: func() {
				rateLimited := rate.NewLimiter(0, 0)
				rlTaskHandler := newTaskHandler(mockMailUtility, rateLimited, mockMailRepo)
				err := rlTaskHandler.HandleSendEmail(ctx, task)
				assert.Error(t, err)
			},
		},
		{
			Name: "when email utility returns non nil error, should return error to be retried later",
			MockFn: func() {
				mockMailRepo.EXPECT().FindByID(ctx, id).Times(1).Return(email, nil)
				mockMailUtility.EXPECT().SendEmail(ctx, sendEmailInput).Times(1).Return("", mail.ClientSignature(""), errors.New("any random error here"))
			},
			Run: func() {
				err := taskHandler.HandleSendEmail(ctx, task)
				assert.Error(t, err)
			},
		},
		{
			Name: "avoid retry if failed to update data after success on send email",
			MockFn: func() {
				mockMailRepo.EXPECT().FindByID(ctx, id).Times(1).Return(email, nil)
				mockMailUtility.EXPECT().SendEmail(ctx, sendEmailInput).Times(1).Return("metadata", mail.MailgunSignature, nil)
				mockMailRepo.EXPECT().Update(ctx, gomock.Any()).Times(1).Return(errors.New("failure on update to db"))
			},
			Run: func() {
				err := taskHandler.HandleSendEmail(ctx, task)
				assert.NoError(t, err)
			},
		},
		{
			Name: "successfully handle task to sent email",
			MockFn: func() {
				mockMailRepo.EXPECT().FindByID(ctx, id).Times(1).Return(email, nil)
				mockMailUtility.EXPECT().SendEmail(ctx, sendEmailInput).Times(1).Return("metadata", mail.MailgunSignature, nil)
				mockMailRepo.EXPECT().Update(ctx, gomock.Any()).Times(1).Return(nil)
			},
			Run: func() {
				err := taskHandler.HandleSendEmail(ctx, task)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		tt.MockFn()
		tt.Run()
	}
}
