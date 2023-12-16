// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/luckyAkbar/atec-api/internal/model (interfaces: AccessTokenRepository)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	model "github.com/luckyAkbar/atec-api/internal/model"
	gorm "gorm.io/gorm"
)

// MockAccessTokenRepository is a mock of AccessTokenRepository interface.
type MockAccessTokenRepository struct {
	ctrl     *gomock.Controller
	recorder *MockAccessTokenRepositoryMockRecorder
}

// MockAccessTokenRepositoryMockRecorder is the mock recorder for MockAccessTokenRepository.
type MockAccessTokenRepositoryMockRecorder struct {
	mock *MockAccessTokenRepository
}

// NewMockAccessTokenRepository creates a new mock instance.
func NewMockAccessTokenRepository(ctrl *gomock.Controller) *MockAccessTokenRepository {
	mock := &MockAccessTokenRepository{ctrl: ctrl}
	mock.recorder = &MockAccessTokenRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccessTokenRepository) EXPECT() *MockAccessTokenRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockAccessTokenRepository) Create(arg0 context.Context, arg1 *model.AccessToken) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockAccessTokenRepositoryMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockAccessTokenRepository)(nil).Create), arg0, arg1)
}

// DeleteByID mocks base method.
func (m *MockAccessTokenRepository) DeleteByID(arg0 context.Context, arg1 uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByID indicates an expected call of DeleteByID.
func (mr *MockAccessTokenRepositoryMockRecorder) DeleteByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByID", reflect.TypeOf((*MockAccessTokenRepository)(nil).DeleteByID), arg0, arg1)
}

// DeleteByUserID mocks base method.
func (m *MockAccessTokenRepository) DeleteByUserID(arg0 context.Context, arg1 uuid.UUID, arg2 *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByUserID", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByUserID indicates an expected call of DeleteByUserID.
func (mr *MockAccessTokenRepositoryMockRecorder) DeleteByUserID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByUserID", reflect.TypeOf((*MockAccessTokenRepository)(nil).DeleteByUserID), arg0, arg1, arg2)
}

// FindByToken mocks base method.
func (m *MockAccessTokenRepository) FindByToken(arg0 context.Context, arg1 string) (*model.AccessToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByToken", arg0, arg1)
	ret0, _ := ret[0].(*model.AccessToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByToken indicates an expected call of FindByToken.
func (mr *MockAccessTokenRepositoryMockRecorder) FindByToken(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByToken", reflect.TypeOf((*MockAccessTokenRepository)(nil).FindByToken), arg0, arg1)
}

// FindCredentialByToken mocks base method.
func (m *MockAccessTokenRepository) FindCredentialByToken(arg0 context.Context, arg1 string) (*model.AccessToken, *model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindCredentialByToken", arg0, arg1)
	ret0, _ := ret[0].(*model.AccessToken)
	ret1, _ := ret[1].(*model.User)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// FindCredentialByToken indicates an expected call of FindCredentialByToken.
func (mr *MockAccessTokenRepositoryMockRecorder) FindCredentialByToken(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindCredentialByToken", reflect.TypeOf((*MockAccessTokenRepository)(nil).FindCredentialByToken), arg0, arg1)
}
