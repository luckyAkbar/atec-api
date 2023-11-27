package model

import (
	"context"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Task is the datatype for all background processor task
type Task string

// list of all available task
const (
	TaskSendEmail Task = "ATEC-API:sendEmail"
)

// WorkerClient is the interface for all worker client mainly to enqueue task
type WorkerClient interface {
	EnqueueSendEmailTask(ctx context.Context, id uuid.UUID) (*asynq.TaskInfo, error)
}
