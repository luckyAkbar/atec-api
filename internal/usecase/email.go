// Package usecase holds all the bussiness rules related function
package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	custerr "github.com/sweet-go/stdlib/error"
	"github.com/sweet-go/stdlib/helper"
	"gopkg.in/guregu/null.v4"
)

type emailUc struct {
	emailRepo    model.EmailRepository
	workerClient model.WorkerClient
}

// NewEmailUsecase satisfy model.EmailUsecase interface
func NewEmailUsecase(emailRepo model.EmailRepository, workerClient model.WorkerClient) model.EmailUsecase {
	return &emailUc{
		emailRepo:    emailRepo,
		workerClient: workerClient,
	}
}

func (uc *emailUc) Register(ctx context.Context, input *model.RegisterEmailInput) (*model.Email, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "emailUc.Register",
		"input": helper.Dump(input),
	})

	if err := input.Validate(); err != nil {
		return nil, custerr.ErrChain{
			Message: "invalid values on input",
			Cause:   err,
			Code:    http.StatusBadRequest,
			Type:    ErrEmailInputInvalid,
		}
	}

	deadline := null.Int{}
	if input.DeadlineSecond != 0 {
		deadline = null.NewInt(input.DeadlineSecond, true)
	}

	email := &model.Email{
		ID:        uuid.New(),
		Subject:   input.Subject,
		Body:      input.Body,
		To:        input.To,
		Cc:        input.Cc,
		Bcc:       input.Bcc,
		Deadline:  deadline,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := uc.emailRepo.Create(ctx, email); err != nil {
		logger.WithError(err).Error("repository layer return error when create emails")
		return nil, custerr.ErrChain{
			Message: "failed to create email",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	taskInfo, err := uc.workerClient.EnqueueSendEmailTask(ctx, email.ID)
	if err != nil {
		logger.WithError(err).Error("failed to enqueue send email task")
		return nil, custerr.ErrChain{
			Message: "failed to enqueue send email task",
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
	}

	logger.Debug("received task info: ", helper.Dump(taskInfo))
	return email, nil
}
