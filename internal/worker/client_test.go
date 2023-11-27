package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/stretchr/testify/assert"
	workerPkg "github.com/sweet-go/stdlib/worker"
	workerMock "github.com/sweet-go/stdlib/worker/mock"
)

func TestWorker_EnqueueSendEmailTask(t *testing.T) {
	ctx := context.Background()
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	ctrl := gomock.NewController(t)

	mockWorkerClient := workerMock.NewMockClient(ctrl)

	id := uuid.New()
	payload, err := json.Marshal(id)
	assert.NoError(t, err)

	client := NewClient(mockWorkerClient)

	successTaskInfo := &asynq.TaskInfo{
		ID: uuid.NewString(),
	}
	errFailedEnqueue := errors.New("failed to enqueue")

	tests := []common.TestStructure{
		{
			Name: "failed when enqueue task to worker client",
			MockFn: func() {
				mockWorkerClient.EXPECT().
					EnqueueTask(
						ctx,
						asynq.NewTask(string(model.TaskSendEmail), payload, asynq.Queue(string(workerPkg.PriorityHigh))),
					).
					Times(1).Return(nil, errFailedEnqueue)
			},
			Run: func() {
				_, err := client.EnqueueSendEmailTask(ctx, id)
				assert.Error(t, err)
				assert.EqualError(t, err, errFailedEnqueue.Error())
			},
		},
		{
			Name: "success, task email enqueued",
			MockFn: func() {
				mockWorkerClient.EXPECT().
					EnqueueTask(
						ctx,
						asynq.NewTask(string(model.TaskSendEmail), payload, asynq.Queue(string(workerPkg.PriorityHigh))),
					).
					Times(1).Return(successTaskInfo, nil)
			},
			Run: func() {
				ti, err := client.EnqueueSendEmailTask(ctx, id)
				assert.NoError(t, err)
				assert.Equal(t, ti, successTaskInfo)
			},
		},
	}

	for _, tt := range tests {
		tt.MockFn()
		tt.Run()
	}
}
