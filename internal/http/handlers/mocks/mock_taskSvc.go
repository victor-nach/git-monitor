// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/victor-nach/git-monitor/internal/http/handlers (interfaces: taskSvc)
//
// Generated by this command:
//
//	mockgen -destination=./internal/http/handlers/mocks/mock_taskSvc.go -package=mocks github.com/victor-nach/git-monitor/internal/http/handlers taskSvc
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	models "github.com/victor-nach/git-monitor/internal/domain/models"
	gomock "go.uber.org/mock/gomock"
)

// MocktaskSvc is a mock of taskSvc interface.
type MocktaskSvc struct {
	ctrl     *gomock.Controller
	recorder *MocktaskSvcMockRecorder
	isgomock struct{}
}

// MocktaskSvcMockRecorder is the mock recorder for MocktaskSvc.
type MocktaskSvcMockRecorder struct {
	mock *MocktaskSvc
}

// NewMocktaskSvc creates a new mock instance.
func NewMocktaskSvc(ctrl *gomock.Controller) *MocktaskSvc {
	mock := &MocktaskSvc{ctrl: ctrl}
	mock.recorder = &MocktaskSvcMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocktaskSvc) EXPECT() *MocktaskSvcMockRecorder {
	return m.recorder
}

// GetTask mocks base method.
func (m *MocktaskSvc) GetTask(ctx context.Context, id string) (models.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTask", ctx, id)
	ret0, _ := ret[0].(models.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTask indicates an expected call of GetTask.
func (mr *MocktaskSvcMockRecorder) GetTask(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTask", reflect.TypeOf((*MocktaskSvc)(nil).GetTask), ctx, id)
}

// List mocks base method.
func (m *MocktaskSvc) List(ctx context.Context) ([]models.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx)
	ret0, _ := ret[0].([]models.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MocktaskSvcMockRecorder) List(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MocktaskSvc)(nil).List), ctx)
}

// TriggerTask mocks base method.
func (m *MocktaskSvc) TriggerTask(ctx context.Context, RepoInfo models.RepoInfo, since *time.Time) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TriggerTask", ctx, RepoInfo, since)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TriggerTask indicates an expected call of TriggerTask.
func (mr *MocktaskSvcMockRecorder) TriggerTask(ctx, RepoInfo, since any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TriggerTask", reflect.TypeOf((*MocktaskSvc)(nil).TriggerTask), ctx, RepoInfo, since)
}
