package engine

import (
	"github.com/gojektech/proctor/jobs"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) ListJobs() ([]jobs.Metadata, error) {
	args := m.Called()
	return args.Get(0).([]jobs.Metadata), args.Error(1)
}

func (m *MockClient) ExecuteJob(name string, jobArgs map[string]string) (string, error) {
	args := m.Called(name, jobArgs)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockClient) StreamJobLogs(name string) error {
	args := m.Called(name)
	return args.Error(0)
}
