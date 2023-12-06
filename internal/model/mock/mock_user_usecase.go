// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/luckyAkbar/atec-api/internal/model (interfaces: UserUsecase)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	common "github.com/luckyAkbar/atec-api/internal/common"
	model "github.com/luckyAkbar/atec-api/internal/model"
)

// MockUserUsecase is a mock of UserUsecase interface.
type MockUserUsecase struct {
	ctrl     *gomock.Controller
	recorder *MockUserUsecaseMockRecorder
}

// MockUserUsecaseMockRecorder is the mock recorder for MockUserUsecase.
type MockUserUsecaseMockRecorder struct {
	mock *MockUserUsecase
}

// NewMockUserUsecase creates a new mock instance.
func NewMockUserUsecase(ctrl *gomock.Controller) *MockUserUsecase {
	mock := &MockUserUsecase{ctrl: ctrl}
	mock.recorder = &MockUserUsecaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserUsecase) EXPECT() *MockUserUsecaseMockRecorder {
	return m.recorder
}

// InitiateResetPassword mocks base method.
func (m *MockUserUsecase) InitiateResetPassword(arg0 context.Context, arg1 uuid.UUID) (*model.InitiateResetPasswordOutput, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitiateResetPassword", arg0, arg1)
	ret0, _ := ret[0].(*model.InitiateResetPasswordOutput)
	ret1, _ := ret[1].(*common.Error)
	return ret0, ret1
}

// InitiateResetPassword indicates an expected call of InitiateResetPassword.
func (mr *MockUserUsecaseMockRecorder) InitiateResetPassword(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitiateResetPassword", reflect.TypeOf((*MockUserUsecase)(nil).InitiateResetPassword), arg0, arg1)
}

// SignUp mocks base method.
func (m *MockUserUsecase) SignUp(arg0 context.Context, arg1 *model.SignUpInput) (*model.SignUpResponse, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignUp", arg0, arg1)
	ret0, _ := ret[0].(*model.SignUpResponse)
	ret1, _ := ret[1].(*common.Error)
	return ret0, ret1
}

// SignUp indicates an expected call of SignUp.
func (mr *MockUserUsecaseMockRecorder) SignUp(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignUp", reflect.TypeOf((*MockUserUsecase)(nil).SignUp), arg0, arg1)
}

// VerifyAccount mocks base method.
func (m *MockUserUsecase) VerifyAccount(arg0 context.Context, arg1 *model.AccountVerificationInput) (*model.SuccessAccountVerificationResponse, *model.FailedAccountVerificationResponse, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyAccount", arg0, arg1)
	ret0, _ := ret[0].(*model.SuccessAccountVerificationResponse)
	ret1, _ := ret[1].(*model.FailedAccountVerificationResponse)
	ret2, _ := ret[2].(*common.Error)
	return ret0, ret1, ret2
}

// VerifyAccount indicates an expected call of VerifyAccount.
func (mr *MockUserUsecaseMockRecorder) VerifyAccount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyAccount", reflect.TypeOf((*MockUserUsecase)(nil).VerifyAccount), arg0, arg1)
}
