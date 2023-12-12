// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/luckyAkbar/atec-api/internal/model (interfaces: SDPackageUsecase)

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

// MockSDPackageUsecase is a mock of SDPackageUsecase interface.
type MockSDPackageUsecase struct {
	ctrl     *gomock.Controller
	recorder *MockSDPackageUsecaseMockRecorder
}

// MockSDPackageUsecaseMockRecorder is the mock recorder for MockSDPackageUsecase.
type MockSDPackageUsecaseMockRecorder struct {
	mock *MockSDPackageUsecase
}

// NewMockSDPackageUsecase creates a new mock instance.
func NewMockSDPackageUsecase(ctrl *gomock.Controller) *MockSDPackageUsecase {
	mock := &MockSDPackageUsecase{ctrl: ctrl}
	mock.recorder = &MockSDPackageUsecaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSDPackageUsecase) EXPECT() *MockSDPackageUsecaseMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockSDPackageUsecase) Create(arg0 context.Context, arg1 *model.SDPackage) (*model.GeneratedSDPackage, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(*model.GeneratedSDPackage)
	ret1, _ := ret[1].(*common.Error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockSDPackageUsecaseMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSDPackageUsecase)(nil).Create), arg0, arg1)
}

// Delete mocks base method.
func (m *MockSDPackageUsecase) Delete(arg0 context.Context, arg1 uuid.UUID) (*model.GeneratedSDPackage, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(*model.GeneratedSDPackage)
	ret1, _ := ret[1].(*common.Error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockSDPackageUsecaseMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSDPackageUsecase)(nil).Delete), arg0, arg1)
}

// FindByID mocks base method.
func (m *MockSDPackageUsecase) FindByID(arg0 context.Context, arg1 uuid.UUID) (*model.GeneratedSDPackage, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", arg0, arg1)
	ret0, _ := ret[0].(*model.GeneratedSDPackage)
	ret1, _ := ret[1].(*common.Error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID.
func (mr *MockSDPackageUsecaseMockRecorder) FindByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockSDPackageUsecase)(nil).FindByID), arg0, arg1)
}

// Search mocks base method.
func (m *MockSDPackageUsecase) Search(arg0 context.Context, arg1 *model.SearchSDPackageInput) (*model.SearchPackageOutput, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search", arg0, arg1)
	ret0, _ := ret[0].(*model.SearchPackageOutput)
	ret1, _ := ret[1].(*common.Error)
	return ret0, ret1
}

// Search indicates an expected call of Search.
func (mr *MockSDPackageUsecaseMockRecorder) Search(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Search", reflect.TypeOf((*MockSDPackageUsecase)(nil).Search), arg0, arg1)
}

// UndoDelete mocks base method.
func (m *MockSDPackageUsecase) UndoDelete(arg0 context.Context, arg1 uuid.UUID) (*model.GeneratedSDPackage, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UndoDelete", arg0, arg1)
	ret0, _ := ret[0].(*model.GeneratedSDPackage)
	ret1, _ := ret[1].(*common.Error)
	return ret0, ret1
}

// UndoDelete indicates an expected call of UndoDelete.
func (mr *MockSDPackageUsecaseMockRecorder) UndoDelete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UndoDelete", reflect.TypeOf((*MockSDPackageUsecase)(nil).UndoDelete), arg0, arg1)
}

// Update mocks base method.
func (m *MockSDPackageUsecase) Update(arg0 context.Context, arg1 uuid.UUID, arg2 *model.SDPackage) (*model.GeneratedSDPackage, *common.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1, arg2)
	ret0, _ := ret[0].(*model.GeneratedSDPackage)
	ret1, _ := ret[1].(*common.Error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockSDPackageUsecaseMockRecorder) Update(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSDPackageUsecase)(nil).Update), arg0, arg1, arg2)
}
