package engine

import (
	"github.com/gojektech/proctor/proc"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) ListProcs() ([]proc.Metadata, error) {
	args := m.Called()
	return args.Get(0).([]proc.Metadata), args.Error(1)
}

func (m *MockClient) ExecuteProc(name string, procArgs map[string]string) (string, error) {
	args := m.Called(name, procArgs)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockClient) StreamProcLogs(name string) error {
	args := m.Called(name)
	return args.Error(0)
}
