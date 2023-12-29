// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/luckyAkbar/atec-api/internal/model (interfaces: SDTestRepository)

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

// MockSDTestRepository is a mock of SDTestRepository interface.
type MockSDTestRepository struct {
	ctrl     *gomock.Controller
	recorder *MockSDTestRepositoryMockRecorder
}

// MockSDTestRepositoryMockRecorder is the mock recorder for MockSDTestRepository.
type MockSDTestRepositoryMockRecorder struct {
	mock *MockSDTestRepository
}

// NewMockSDTestRepository creates a new mock instance.
func NewMockSDTestRepository(ctrl *gomock.Controller) *MockSDTestRepository {
	mock := &MockSDTestRepository{ctrl: ctrl}
	mock.recorder = &MockSDTestRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSDTestRepository) EXPECT() *MockSDTestRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockSDTestRepository) Create(arg0 context.Context, arg1 *model.SDTest, arg2 *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockSDTestRepositoryMockRecorder) Create(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSDTestRepository)(nil).Create), arg0, arg1, arg2)
}

// FindByID mocks base method.
func (m *MockSDTestRepository) FindByID(arg0 context.Context, arg1 uuid.UUID) (*model.SDTest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", arg0, arg1)
	ret0, _ := ret[0].(*model.SDTest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID.
func (mr *MockSDTestRepositoryMockRecorder) FindByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockSDTestRepository)(nil).FindByID), arg0, arg1)
}

// Search mocks base method.
func (m *MockSDTestRepository) Search(arg0 context.Context, arg1 *model.ViewHistoriesInput) ([]*model.SDTest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search", arg0, arg1)
	ret0, _ := ret[0].([]*model.SDTest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Search indicates an expected call of Search.
func (mr *MockSDTestRepositoryMockRecorder) Search(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Search", reflect.TypeOf((*MockSDTestRepository)(nil).Search), arg0, arg1)
}

// Statistic mocks base method.
func (m *MockSDTestRepository) Statistic(arg0 context.Context, arg1 uuid.UUID) ([]model.SDTestStatistic, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Statistic", arg0, arg1)
	ret0, _ := ret[0].([]model.SDTestStatistic)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Statistic indicates an expected call of Statistic.
func (mr *MockSDTestRepositoryMockRecorder) Statistic(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Statistic", reflect.TypeOf((*MockSDTestRepository)(nil).Statistic), arg0, arg1)
}

// Update mocks base method.
func (m *MockSDTestRepository) Update(arg0 context.Context, arg1 *model.SDTest, arg2 *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockSDTestRepositoryMockRecorder) Update(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSDTestRepository)(nil).Update), arg0, arg1, arg2)
}
