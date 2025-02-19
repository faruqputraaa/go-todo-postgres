// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/cache/cache.go

// Package mock_cache is a generated GoMock package.
package mock_cache

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockCacheable is a mock of Cacheable interface.
type MockCacheable struct {
	ctrl     *gomock.Controller
	recorder *MockCacheableMockRecorder
}

// MockCacheableMockRecorder is the mock recorder for MockCacheable.
type MockCacheableMockRecorder struct {
	mock *MockCacheable
}

// NewMockCacheable creates a new mock instance.
func NewMockCacheable(ctrl *gomock.Controller) *MockCacheable {
	mock := &MockCacheable{ctrl: ctrl}
	mock.recorder = &MockCacheableMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheable) EXPECT() *MockCacheableMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockCacheable) Delete(key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", key)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockCacheableMockRecorder) Delete(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockCacheable)(nil).Delete), key)
}

// Get mocks base method.
func (m *MockCacheable) Get(key string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", key)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockCacheableMockRecorder) Get(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCacheable)(nil).Get), key)
}

// Set mocks base method.
func (m *MockCacheable) Set(key string, value interface{}, duration time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", key, value, duration)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockCacheableMockRecorder) Set(key, value, duration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCacheable)(nil).Set), key, value, duration)
}
