// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/luckyAkbar/atec-api/internal/model (interfaces: SDTemplateRepository)

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

// MockSDTemplateRepository is a mock of SDTemplateRepository interface.
type MockSDTemplateRepository struct {
	ctrl     *gomock.Controller
	recorder *MockSDTemplateRepositoryMockRecorder
}

// MockSDTemplateRepositoryMockRecorder is the mock recorder for MockSDTemplateRepository.
type MockSDTemplateRepositoryMockRecorder struct {
	mock *MockSDTemplateRepository
}

// NewMockSDTemplateRepository creates a new mock instance.
func NewMockSDTemplateRepository(ctrl *gomock.Controller) *MockSDTemplateRepository {
	mock := &MockSDTemplateRepository{ctrl: ctrl}
	mock.recorder = &MockSDTemplateRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSDTemplateRepository) EXPECT() *MockSDTemplateRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockSDTemplateRepository) Create(arg0 context.Context, arg1 *model.SpeechDelayTemplate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockSDTemplateRepositoryMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSDTemplateRepository)(nil).Create), arg0, arg1)
}

// FindByID mocks base method.
func (m *MockSDTemplateRepository) FindByID(arg0 context.Context, arg1 uuid.UUID) (*model.SpeechDelayTemplate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", arg0, arg1)
	ret0, _ := ret[0].(*model.SpeechDelayTemplate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID.
func (mr *MockSDTemplateRepositoryMockRecorder) FindByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockSDTemplateRepository)(nil).FindByID), arg0, arg1)
}

// Search mocks base method.
func (m *MockSDTemplateRepository) Search(arg0 context.Context, arg1 *model.SearchSDTemplateInput) ([]*model.SpeechDelayTemplate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search", arg0, arg1)
	ret0, _ := ret[0].([]*model.SpeechDelayTemplate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Search indicates an expected call of Search.
func (mr *MockSDTemplateRepositoryMockRecorder) Search(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Search", reflect.TypeOf((*MockSDTemplateRepository)(nil).Search), arg0, arg1)
}

// Update mocks base method.
func (m *MockSDTemplateRepository) Update(arg0 context.Context, arg1 *model.SpeechDelayTemplate, arg2 *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockSDTemplateRepositoryMockRecorder) Update(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSDTemplateRepository)(nil).Update), arg0, arg1, arg2)
}
