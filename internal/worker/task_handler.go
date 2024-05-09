package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/config"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	"github.com/sweet-go/stdlib/mail"
	workerPkg "github.com/sweet-go/stdlib/worker"
	"golang.org/x/time/rate"
	"gopkg.in/guregu/null.v4"
)

type th struct {
	mailUtil      mail.Utility
	limiter       *rate.Limiter
	mailRepo      model.EmailRepository
	sharedCryptor common.SharedCryptor
}

func newTaskHandler(mailUtil mail.Utility, limiter *rate.Limiter, mailRepo model.EmailRepository, sharedCryptor common.SharedCryptor) *th {
	return &th{
		mailUtil:      mailUtil,
		limiter:       limiter,
		mailRepo:      mailRepo,
		sharedCryptor: sharedCryptor,
	}
}

func (th *th) HandleSendEmail(ctx context.Context, task *asynq.Task) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func": "taskHandler.HandleSendEmail",
	})

	var id uuid.UUID
	if err := json.Unmarshal(task.Payload(), &id); err != nil {
		logger.WithError(err).Error("failed to unmarshal payload for send email")
		return err
	}

	if !th.limiter.Allow() {
		logger.WithField("id", id).Warn("rate limit exceeded for task: ", task.Type())
		return newWorkerRateLimitError()
	}

	email, err := th.mailRepo.FindByID(ctx, id)
	switch err {
	default:
		logger.WithError(err).Error("failed to find email")
		return err
	case repository.ErrNotFound:
		logger.WithError(err).Warn("email doesn't found on db. skipping without marking error")
		return nil
	case nil:
		break
	}

	if err := email.Decrypt(th.sharedCryptor); err != nil {
		logger.WithError(err).Error("failed to decrypt email data")
		return err
	}

	md, sig, err := th.mailUtil.SendEmail(ctx, &mail.Mail{
		ID:          email.ID.String(),
		To:          email.GenericReceipientsTo(),
		Cc:          email.GenericReceipientsCc(),
		Bcc:         email.GenericReceipientsBcc(),
		HTMLContent: email.Body,
		Subject:     email.Subject,
	})

	if err != nil {
		logger.WithField("email", helper.Dump(email)).WithError(err).Error("mailUtil.SendEmail returns error, failing to send email")
		return err
	}

	email.ClientSignature = null.StringFrom(string(sig))
	email.UpdatedAt = time.Now().UTC()
	email.Metadata = null.StringFrom(md)
	email.SentAt = null.TimeFrom(time.Now().UTC())

	if err := th.mailRepo.Update(ctx, email); err != nil {
		logger.WithError(err).Error("failed to update email after successfully sent it, not marking as failure")
	}

	return nil
}

func newWorkerRateLimitError() error {
	return workerPkg.NewRateLimitError(config.WorkerLimiterRetryInterval())
}
