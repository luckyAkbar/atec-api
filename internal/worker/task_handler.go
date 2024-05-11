package worker

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
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
	mailUtil        mail.Utility
	limiter         *rate.Limiter
	mailRepo        model.EmailRepository
	userRepo        model.UserRepository
	accessTokenRepo model.AccessTokenRepository
}

func newTaskHandler(mailUtil mail.Utility, limiter *rate.Limiter, mailRepo model.EmailRepository, userRepo model.UserRepository, accessTokenRepo model.AccessTokenRepository) *th {
	return &th{
		mailUtil:        mailUtil,
		limiter:         limiter,
		mailRepo:        mailRepo,
		accessTokenRepo: accessTokenRepo,
		userRepo:        userRepo,
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

	if email.IsAlreadyPastDeadline() {
		logger.WithError(errors.New("detected a stale email on mailing list")).
			WithField("email", helper.Dump(email)).
			Error("unable to send stale email on mailing list, will not be retried")
		email.UpdatedAt = time.Now().UTC()

		if err := th.mailRepo.Update(ctx, email); err != nil {
			logger.WithField("id", email.ID).WithError(err).Error("failed to update email data for updated_at field")
		}

		return nil
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
		logger.WithError(err).Error("mailUtil.SendEmail returns error, failing to send email")
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

func (th *th) HandleEnforceActiveTokenLimiter(ctx context.Context, task *asynq.Task) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":    "taskHandler.HandleEnforceActiveTokenLimiter",
		"payload": string(task.Payload()),
	})

	var userID uuid.UUID
	if err := json.Unmarshal(task.Payload(), &userID); err != nil {
		logger.WithError(err).Error("failed to unmarshal payload for send email")
		return err
	}

	if !th.limiter.Allow() {
		logger.WithField("userID", userID).Warn("rate limit exceeded for task: ", task.Type())
		return newWorkerRateLimitError()
	}

	user, err := th.userRepo.FindByID(ctx, userID)
	switch err {
	default:
		logger.WithError(err).Error("failed to fetch user data")
		return err
	case repository.ErrNotFound:
		logger.WithError(err).Error("user data not found on database thus unable to perform task and will not be retried")
		return nil
	case nil:
		break
	}

	accessTokens, err := th.accessTokenRepo.FindByUserID(ctx, user.ID, config.ActiveTokenLimit()*2)
	switch err {
	default:
		logger.WithError(err).Error("failed to fetch access token data")
		return err
	case repository.ErrNotFound:
		logger.WithField("userID", userID).Info("no access token found for this userID, thus marked as success")
		return nil
	case nil:
		break
	}

	if len(accessTokens) <= config.ActiveTokenLimit() {
		logger.WithField("userID", user.ID).Info("this user still doesn't have exceed active token limit")
		return nil
	}

	// sort ascending, where the result will be sorted by created_at asc
	sort.SliceStable(accessTokens, func(i, j int) bool {
		return accessTokens[i].CreatedAt.Before(accessTokens[j].CreatedAt)
	})

	idsToDelete := []uuid.UUID{}
	tokensToDelete := []string{}
	for i := 0; i < len(accessTokens)-config.ActiveTokenLimit(); i++ {
		idsToDelete = append(idsToDelete, accessTokens[i].ID)
		tokensToDelete = append(tokensToDelete, accessTokens[i].Token)
	}

	if err := th.accessTokenRepo.DeleteCredentialsFromCache(ctx, tokensToDelete); err != nil {
		logger.WithError(err).Error("failed to delete credentials from cache")
		return err
	}

	if err := th.accessTokenRepo.DeleteByIDs(ctx, idsToDelete, true); err != nil {
		logger.WithError(err).Error("failed to batch delete exceeding access token")
		return err
	}

	return nil
}

func newWorkerRateLimitError() error {
	return workerPkg.NewRateLimitError(config.WorkerLimiterRetryInterval())
}
