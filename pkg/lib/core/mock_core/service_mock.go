// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/anyproto/anytype-heart/pkg/lib/core (interfaces: Service)

// Package mock_core is a generated GoMock package.
package mock_core

import (
	context "context"
	reflect "reflect"

	app "github.com/anyproto/any-sync/app"
	core "github.com/anyproto/anytype-heart/pkg/lib/core"
	threads "github.com/anyproto/anytype-heart/pkg/lib/threads"
	gomock "go.uber.org/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockService) Close(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockServiceMockRecorder) Close(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockService)(nil).Close), arg0)
}

// EnsurePredefinedBlocks mocks base method.
func (m *MockService) EnsurePredefinedBlocks(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsurePredefinedBlocks", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnsurePredefinedBlocks indicates an expected call of EnsurePredefinedBlocks.
func (mr *MockServiceMockRecorder) EnsurePredefinedBlocks(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsurePredefinedBlocks", reflect.TypeOf((*MockService)(nil).EnsurePredefinedBlocks), arg0)
}

// GetAllWorkspaces mocks base method.
func (m *MockService) GetAllWorkspaces() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllWorkspaces")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllWorkspaces indicates an expected call of GetAllWorkspaces.
func (mr *MockServiceMockRecorder) GetAllWorkspaces() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllWorkspaces", reflect.TypeOf((*MockService)(nil).GetAllWorkspaces))
}

// GetWorkspaceIdForObject mocks base method.
func (m *MockService) GetWorkspaceIdForObject(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWorkspaceIdForObject", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWorkspaceIdForObject indicates an expected call of GetWorkspaceIdForObject.
func (mr *MockServiceMockRecorder) GetWorkspaceIdForObject(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkspaceIdForObject", reflect.TypeOf((*MockService)(nil).GetWorkspaceIdForObject), arg0)
}

// Init mocks base method.
func (m *MockService) Init(arg0 *app.App) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockServiceMockRecorder) Init(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockService)(nil).Init), arg0)
}

// IsStarted mocks base method.
func (m *MockService) IsStarted() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsStarted")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsStarted indicates an expected call of IsStarted.
func (mr *MockServiceMockRecorder) IsStarted() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsStarted", reflect.TypeOf((*MockService)(nil).IsStarted))
}

// LocalProfile mocks base method.
func (m *MockService) LocalProfile() (core.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LocalProfile")
	ret0, _ := ret[0].(core.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LocalProfile indicates an expected call of LocalProfile.
func (mr *MockServiceMockRecorder) LocalProfile() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LocalProfile", reflect.TypeOf((*MockService)(nil).LocalProfile))
}

// Name mocks base method.
func (m *MockService) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockServiceMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockService)(nil).Name))
}

// PredefinedBlocks mocks base method.
func (m *MockService) PredefinedBlocks() threads.DerivedSmartblockIds {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PredefinedBlocks")
	ret0, _ := ret[0].(threads.DerivedSmartblockIds)
	return ret0
}

// PredefinedBlocks indicates an expected call of PredefinedBlocks.
func (mr *MockServiceMockRecorder) PredefinedBlocks() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PredefinedBlocks", reflect.TypeOf((*MockService)(nil).PredefinedBlocks))
}

// ProfileID mocks base method.
func (m *MockService) ProfileID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ProfileID indicates an expected call of ProfileID.
func (mr *MockServiceMockRecorder) ProfileID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileID", reflect.TypeOf((*MockService)(nil).ProfileID))
}

// Run mocks base method.
func (m *MockService) Run(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockServiceMockRecorder) Run(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockService)(nil).Run), arg0)
}

// Stop mocks base method.
func (m *MockService) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockServiceMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockService)(nil).Stop))
}
