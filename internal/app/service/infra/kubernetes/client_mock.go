package kubernetes

import (
	"io"

	"github.com/stretchr/testify/mock"
	"proctor/internal/pkg/utility"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) ExecuteJob(jobName string, envMap map[string]string) (string, error) {
	args := m.Called(jobName, envMap)
	return args.String(0), args.Error(1)
}

func (m *MockClient) ExecuteJobWithCommand(jobName string, envMap map[string]string, command []string) (string, error) {
	args := m.Called(jobName, envMap)
	return args.String(0), args.Error(1)
}

func (m *MockClient) StreamJobLogs(executionName string) (io.ReadCloser, error) {
	args := m.Called(executionName)
	return args.Get(0).(*utility.Buffer), args.Error(1)
}

func (m *MockClient) JobExecutionStatus(executionName string) (string, error) {
	args := m.Called(executionName)
	return args.String(0), args.Error(1)
}
