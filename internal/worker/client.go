// Package worker hold the implementation of all the background processor for this service
package worker

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/helper"
	workerPkg "github.com/sweet-go/stdlib/worker"
)

type client struct {
	workerClient workerPkg.Client
}

// NewClient returns a new worker client
func NewClient(workerClient workerPkg.Client) model.WorkerClient {
	return &client{
		workerClient: workerClient,
	}
}

func (c *client) EnqueueSendEmailTask(ctx context.Context, id uuid.UUID) (*asynq.TaskInfo, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "client.EnqueueSendEmailTask",
		"input": helper.Dump(id),
	})

	payload, err := json.Marshal(id)
	if err != nil {
		logger.WithError(err).Error("failed to marshal payload for enqueue send email task")
		return nil, err
	}

	info, err := c.workerClient.EnqueueTask(ctx, asynq.NewTask(string(model.TaskSendEmail), payload, asynq.Queue(string(workerPkg.PriorityHigh))))
	if err != nil {
		logger.WithError(err).Error("failed to enqueue send email task")
		return nil, err
	}

	return info, nil
}

func (c *client) EnqueueEnforceActiveTokenLimitterTask(ctx context.Context, userID uuid.UUID) (*asynq.TaskInfo, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":  "client.EnqueueEnforceActiveTokenLimitterTask",
		"input": helper.Dump(userID),
	})

	payload, err := json.Marshal(userID)
	if err != nil {
		logger.WithError(err).Error("failed to marshal payload for enqueue enforce active token limitter task")
		return nil, err
	}

	info, err := c.workerClient.EnqueueTask(ctx,
		asynq.NewTask(
			string(model.TaskEnforceActiveTokenLimitter),
			payload,
			asynq.Queue(string(workerPkg.PriorityHigh))))

	if err != nil {
		logger.WithError(err).Error("failed to enqueue task")
		return nil, err
	}

	return info, nil
}
