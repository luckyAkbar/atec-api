// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/luckyAkbar/atec-api/internal/model (interfaces: WorkerClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	asynq "github.com/hibiken/asynq"
)

// MockWorkerClient is a mock of WorkerClient interface.
type MockWorkerClient struct {
	ctrl     *gomock.Controller
	recorder *MockWorkerClientMockRecorder
}

// MockWorkerClientMockRecorder is the mock recorder for MockWorkerClient.
type MockWorkerClientMockRecorder struct {
	mock *MockWorkerClient
}

// NewMockWorkerClient creates a new mock instance.
func NewMockWorkerClient(ctrl *gomock.Controller) *MockWorkerClient {
	mock := &MockWorkerClient{ctrl: ctrl}
	mock.recorder = &MockWorkerClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkerClient) EXPECT() *MockWorkerClientMockRecorder {
	return m.recorder
}

// EnqueueEnforceActiveTokenLimitterTask mocks base method.
func (m *MockWorkerClient) EnqueueEnforceActiveTokenLimitterTask(arg0 context.Context, arg1 uuid.UUID) (*asynq.TaskInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnqueueEnforceActiveTokenLimitterTask", arg0, arg1)
	ret0, _ := ret[0].(*asynq.TaskInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnqueueEnforceActiveTokenLimitterTask indicates an expected call of EnqueueEnforceActiveTokenLimitterTask.
func (mr *MockWorkerClientMockRecorder) EnqueueEnforceActiveTokenLimitterTask(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnqueueEnforceActiveTokenLimitterTask", reflect.TypeOf((*MockWorkerClient)(nil).EnqueueEnforceActiveTokenLimitterTask), arg0, arg1)
}

// EnqueueSendEmailTask mocks base method.
func (m *MockWorkerClient) EnqueueSendEmailTask(arg0 context.Context, arg1 uuid.UUID) (*asynq.TaskInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnqueueSendEmailTask", arg0, arg1)
	ret0, _ := ret[0].(*asynq.TaskInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnqueueSendEmailTask indicates an expected call of EnqueueSendEmailTask.
func (mr *MockWorkerClientMockRecorder) EnqueueSendEmailTask(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnqueueSendEmailTask", reflect.TypeOf((*MockWorkerClient)(nil).EnqueueSendEmailTask), arg0, arg1)
}
