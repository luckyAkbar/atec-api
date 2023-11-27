// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/luckyAkbar/atec-api/internal/model (interfaces: EmailUsecase)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/luckyAkbar/atec-api/internal/model"
)

// MockEmailUsecase is a mock of EmailUsecase interface.
type MockEmailUsecase struct {
	ctrl     *gomock.Controller
	recorder *MockEmailUsecaseMockRecorder
}

// MockEmailUsecaseMockRecorder is the mock recorder for MockEmailUsecase.
type MockEmailUsecaseMockRecorder struct {
	mock *MockEmailUsecase
}

// NewMockEmailUsecase creates a new mock instance.
func NewMockEmailUsecase(ctrl *gomock.Controller) *MockEmailUsecase {
	mock := &MockEmailUsecase{ctrl: ctrl}
	mock.recorder = &MockEmailUsecaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEmailUsecase) EXPECT() *MockEmailUsecaseMockRecorder {
	return m.recorder
}

// Register mocks base method.
func (m *MockEmailUsecase) Register(arg0 context.Context, arg1 *model.RegisterEmailInput) (*model.Email, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0, arg1)
	ret0, _ := ret[0].(*model.Email)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register.
func (mr *MockEmailUsecaseMockRecorder) Register(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockEmailUsecase)(nil).Register), arg0, arg1)
}
