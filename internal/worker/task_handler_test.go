package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/sweet-go/stdlib/mail"
	mailMock "github.com/sweet-go/stdlib/mail/mock"
	workerPkg "github.com/sweet-go/stdlib/worker"
	"golang.org/x/time/rate"
	"gopkg.in/guregu/null.v4"
)

func TestWorker_HandleSendEmail(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	mockMailUtility := mailMock.NewMockUtility(ctrl)
	mockMailRepo := mock.NewMockEmailRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAccessTokenRepo := mock.NewMockAccessTokenRepository(ctrl)

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

	beforeDeadlineEmail := &model.Email{
		ID:        id,
		Subject:   "test",
		Body:      "test",
		To:        pq.StringArray{"test"},
		Cc:        pq.StringArray{"test"},
		Bcc:       pq.StringArray{"test"},
		CreatedAt: time.Now().UTC(),
		Deadline:  null.IntFrom(1),
	}

	deadlineExceedEmail := &model.Email{
		ID:        id,
		Subject:   "test",
		Body:      "test",
		To:        pq.StringArray{"test"},
		Cc:        pq.StringArray{"test"},
		Bcc:       pq.StringArray{"test"},
		CreatedAt: time.Now().Add(time.Second * -2).UTC(),
		Deadline:  null.IntFrom(1),
	}

	sendEmailInput := &mail.Mail{
		ID:          email.ID.String(),
		To:          email.GenericReceipientsTo(),
		Cc:          email.GenericReceipientsCc(),
		Bcc:         email.GenericReceipientsBcc(),
		HTMLContent: email.Body,
		Subject:     email.Subject,
	}

	taskHandler := newTaskHandler(mockMailUtility, normalLimiter, mockMailRepo, mockUserRepo, mockAccessTokenRepo)

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
				rlTaskHandler := newTaskHandler(mockMailUtility, rateLimited, mockMailRepo, mockUserRepo, mockAccessTokenRepo)
				err := rlTaskHandler.HandleSendEmail(ctx, task)
				assert.Error(t, err)
			},
		},
		{
			Name: "email is already passed the deadline: success update",
			MockFn: func() {
				mockMailRepo.EXPECT().FindByID(ctx, id).Times(1).Return(deadlineExceedEmail, nil)
				mockMailRepo.EXPECT().Update(ctx, gomock.Any()).Times(1).Return(nil)
			},
			Run: func() {
				err := taskHandler.HandleSendEmail(ctx, task)
				assert.NoError(t, err)
			},
		},
		{
			Name: "email is already passed the deadline: failed update are just reported and ignored",
			MockFn: func() {
				mockMailRepo.EXPECT().FindByID(ctx, id).Times(1).Return(deadlineExceedEmail, nil)
				mockMailRepo.EXPECT().Update(ctx, gomock.Any()).Times(1).Return(errors.New("failed on db"))
			},
			Run: func() {
				err := taskHandler.HandleSendEmail(ctx, task)
				assert.NoError(t, err)
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
		{
			Name: "successfully handle task to sent email before the deadline",
			MockFn: func() {
				mockMailRepo.EXPECT().FindByID(ctx, id).Times(1).Return(beforeDeadlineEmail, nil)
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

func TestTaskHandler_HandleEnforceActiveTokenLimiter(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	mockMailUtility := mailMock.NewMockUtility(ctrl)
	mockMailRepo := mock.NewMockEmailRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAccessTokenRepo := mock.NewMockAccessTokenRepository(ctrl)

	normalLimiter := rate.NewLimiter(10, 20)
	id := uuid.New()

	payload, err := json.Marshal(id)
	assert.NoError(t, err)

	task := asynq.NewTask(string(model.TaskEnforceActiveTokenLimitter), payload)

	th := newTaskHandler(mockMailUtility, normalLimiter, mockMailRepo, mockUserRepo, mockAccessTokenRepo)

	activeTokenLimit := 5
	viper.Set("server.auth.active_token_limit", activeTokenLimit)

	belowLimitAccessToken := []model.AccessToken{}
	for i := 0; i < activeTokenLimit; i++ {
		belowLimitAccessToken = append(belowLimitAccessToken, model.AccessToken{
			ID:        uuid.New(),
			Token:     uuid.New().String(),
			CreatedAt: time.Now().Add(time.Hour * -14),
		})
	}

	tobeDeletedAccessToken := model.AccessToken{
		ID:        uuid.New(),
		Token:     uuid.New().String(),
		CreatedAt: time.Now().Add(time.Hour * -99),
	}

	aboveLimitAccessToken := belowLimitAccessToken
	aboveLimitAccessToken = append(aboveLimitAccessToken, tobeDeletedAccessToken)

	user := &model.User{
		ID: id,
	}

	tests := []common.TestStructure{
		{
			Name:   "invalid payload -> failed to unmarshal",
			MockFn: func() {},
			Run: func() {
				_task := asynq.NewTask(string(model.TaskEnforceActiveTokenLimitter), []byte("][=-098765321234567890]["))
				err := th.HandleEnforceActiveTokenLimitter(ctx, _task)

				assert.Error(t, err)
				assert.Equal(t, err.Error(), "invalid character ']' looking for beginning of value")
			},
		},
		{
			Name:   "got rate limited error",
			MockFn: func() {},
			Run: func() {
				rateLimited := rate.NewLimiter(0, 0)
				rlTaskHandler := newTaskHandler(mockMailUtility, rateLimited, mockMailRepo, mockUserRepo, mockAccessTokenRepo)
				err := rlTaskHandler.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.Error(t, err)

				assert.Equal(t, err, newWorkerRateLimitError())
			},
		},
		{
			Name: "failed to find user data",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctx, id).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				err := th.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.Error(t, err)
			},
		},
		{
			Name: "got error not found from db -> continue without retrying",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctx, id).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				err := th.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.NoError(t, err)
			},
		},
		{
			Name: "failed to find access tokens",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctx, id).Times(1).Return(user, nil)
				mockAccessTokenRepo.EXPECT().FindByUserID(ctx, user.ID, activeTokenLimit*2).Times(1).Return(nil, errors.New("err db"))
			},
			Run: func() {
				err := th.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.Error(t, err)
			},
		},
		{
			Name: "no access token found, will continue without retrying",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctx, id).Times(1).Return(user, nil)
				mockAccessTokenRepo.EXPECT().FindByUserID(ctx, user.ID, activeTokenLimit*2).Times(1).Return(nil, repository.ErrNotFound)
			},
			Run: func() {
				err := th.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.NoError(t, err)
			},
		},
		{
			Name: "active access token count still below upper limit",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctx, id).Times(1).Return(user, nil)
				mockAccessTokenRepo.EXPECT().FindByUserID(ctx, user.ID, activeTokenLimit*2).Times(1).Return(belowLimitAccessToken, nil)
			},
			Run: func() {
				err := th.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.NoError(t, err)
			},
		},
		{
			Name: "access token are passing maximum limit, but fails when deleting cached creds",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctx, id).Times(1).Return(user, nil)
				mockAccessTokenRepo.EXPECT().FindByUserID(ctx, user.ID, activeTokenLimit*2).Times(1).Return(aboveLimitAccessToken, nil)
				mockAccessTokenRepo.EXPECT().DeleteCredentialsFromCache(ctx, []string{tobeDeletedAccessToken.Token}).Times(1).Return(errors.New("err redis"))
			},
			Run: func() {
				err := th.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.Error(t, err)
			},
		},
		{
			Name: "access token are passing maximum limit, but fails when deleting db record",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctx, id).Times(1).Return(user, nil)
				mockAccessTokenRepo.EXPECT().FindByUserID(ctx, user.ID, activeTokenLimit*2).Times(1).Return(aboveLimitAccessToken, nil)
				mockAccessTokenRepo.EXPECT().DeleteCredentialsFromCache(ctx, []string{tobeDeletedAccessToken.Token}).Times(1).Return(nil)
				mockAccessTokenRepo.EXPECT().DeleteByIDs(ctx, []uuid.UUID{tobeDeletedAccessToken.ID}, true).Times(1).Return(errors.New("err db"))
			},
			Run: func() {
				err := th.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.Error(t, err)
			},
		},
		{
			Name: "access token are passing maximum limit, all process success",
			MockFn: func() {
				mockUserRepo.EXPECT().FindByID(ctx, id).Times(1).Return(user, nil)
				mockAccessTokenRepo.EXPECT().FindByUserID(ctx, user.ID, activeTokenLimit*2).Times(1).Return(aboveLimitAccessToken, nil)
				mockAccessTokenRepo.EXPECT().DeleteCredentialsFromCache(ctx, []string{tobeDeletedAccessToken.Token}).Times(1).Return(nil)
				mockAccessTokenRepo.EXPECT().DeleteByIDs(ctx, []uuid.UUID{tobeDeletedAccessToken.ID}, true).Times(1).Return(nil)
			},
			Run: func() {
				err := th.HandleEnforceActiveTokenLimitter(ctx, task)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		tt.MockFn()
		tt.Run()
	}
}
