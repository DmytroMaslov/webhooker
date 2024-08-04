// Code generated by MockGen. DO NOT EDIT.
// Source: api.go
//
// Generated by this command:
//
//	mockgen -source=api.go -destination=mocks/api_mock.go
//

// Package mock_api is a generated GoMock package.
package mock_api

import (
	reflect "reflect"
	models "webhooker/internal/services/models"

	gomock "go.uber.org/mock/gomock"
)

// MockOrderStorage is a mock of OrderStorage interface.
type MockOrderStorage struct {
	ctrl     *gomock.Controller
	recorder *MockOrderStorageMockRecorder
}

// MockOrderStorageMockRecorder is the mock recorder for MockOrderStorage.
type MockOrderStorageMockRecorder struct {
	mock *MockOrderStorage
}

// NewMockOrderStorage creates a new mock instance.
func NewMockOrderStorage(ctrl *gomock.Controller) *MockOrderStorage {
	mock := &MockOrderStorage{ctrl: ctrl}
	mock.recorder = &MockOrderStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrderStorage) EXPECT() *MockOrderStorageMockRecorder {
	return m.recorder
}

// GetOrder mocks base method.
func (m *MockOrderStorage) GetOrder(arg0 string) (*models.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrder", arg0)
	ret0, _ := ret[0].(*models.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrder indicates an expected call of GetOrder.
func (mr *MockOrderStorageMockRecorder) GetOrder(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrder", reflect.TypeOf((*MockOrderStorage)(nil).GetOrder), arg0)
}

// GetOrders mocks base method.
func (m *MockOrderStorage) GetOrders(arg0 *models.OrderFilter) ([]*models.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", arg0)
	ret0, _ := ret[0].([]*models.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockOrderStorageMockRecorder) GetOrders(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockOrderStorage)(nil).GetOrders), arg0)
}

// SaveOrder mocks base method.
func (m *MockOrderStorage) SaveOrder(arg0 *models.Order) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveOrder", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveOrder indicates an expected call of SaveOrder.
func (mr *MockOrderStorageMockRecorder) SaveOrder(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveOrder", reflect.TypeOf((*MockOrderStorage)(nil).SaveOrder), arg0)
}

// UpdateOrder mocks base method.
func (m *MockOrderStorage) UpdateOrder(arg0 *models.Order) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrder", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrder indicates an expected call of UpdateOrder.
func (mr *MockOrderStorageMockRecorder) UpdateOrder(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrder", reflect.TypeOf((*MockOrderStorage)(nil).UpdateOrder), arg0)
}