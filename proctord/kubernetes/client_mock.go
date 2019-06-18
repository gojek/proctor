package kubernetes

import (
	"io"

	"github.com/stretchr/testify/mock"
	"proctor/shared/utility"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) ExecuteJob(jobName string, envMap map[string]string) (string, error) {
	args := m.Called(jobName, envMap)
	return args.String(0), args.Error(1)
}

func (m *MockClient) StreamJobLogs(jobName string) (io.ReadCloser, error) {
	args := m.Called(jobName)
	return args.Get(0).(*utility.Buffer), args.Error(1)
}

func (m *MockClient) JobExecutionStatus(jobExecutionID string) (string, error) {
	args := m.Called(jobExecutionID)
	return args.String(0), args.Error(1)
}
