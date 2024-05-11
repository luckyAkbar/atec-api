// Package usecase holds all the bussiness rules related function
package usecase

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	custerr "github.com/sweet-go/stdlib/error"
	"github.com/sweet-go/stdlib/helper"
)

type emailUc struct {
	emailRepo     model.EmailRepository
	workerClient  model.WorkerClient
	sharedCryptor common.SharedCryptor
}

// NewEmailUsecase satisfy model.EmailUsecase interface
func NewEmailUsecase(emailRepo model.EmailRepository, workerClient model.WorkerClient, sharedCryptor common.SharedCryptor) model.EmailUsecase {
	return &emailUc{
		emailRepo:     emailRepo,
		workerClient:  workerClient,
		sharedCryptor: sharedCryptor,
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

	email := &model.Email{
		ID:        uuid.New(),
		Subject:   input.Subject,
		Body:      input.Body,
		To:        input.To,
		Cc:        input.Cc,
		Bcc:       input.Bcc,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := email.Encrypt(uc.sharedCryptor); err != nil {
		logger.WithError(err).Error("failed to encrypt email on registration")
		return nil, custerr.ErrChain{
			Message: fmt.Sprintf("failed to encrypt email data %s", input.Body),
			Cause:   err,
			Code:    http.StatusInternalServerError,
			Type:    ErrInternal,
		}
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

	if err := email.Decrypt(uc.sharedCryptor); err != nil {
		logger.WithError(err).Error("failed to decrypt email on registration, may caused receiver to get encrypted data. Skipping...")
	}

	logger.Debug("received task info: ", helper.Dump(taskInfo))
	return email, nil
}
