package daemon

import (
	"github.com/stretchr/testify/mock"

	modelExecution "proctor/internal/pkg/model/execution"
	modelMetadata "proctor/internal/pkg/model/metadata"
	modelSchedule "proctor/internal/pkg/model/schedule"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) ListProcs() ([]modelMetadata.Metadata, error) {
	args := m.Called()
	return args.Get(0).([]modelMetadata.Metadata), args.Error(1)
}

func (m *MockClient) ListScheduledProcs() ([]modelSchedule.ScheduledJob, error) {
	args := m.Called()
	return args.Get(0).([]modelSchedule.ScheduledJob), args.Error(1)
}

func (m *MockClient) ExecuteProc(name string, procArgs map[string]string) (*modelExecution.ExecutionResult, error) {
	args := m.Called(name, procArgs)
	return args.Get(0).(*modelExecution.ExecutionResult), args.Error(1)
}

func (m *MockClient) StreamProcLogs(executionId uint64) error {
	args := m.Called(executionId)
	return args.Error(0)
}

func (m *MockClient) GetExecutionContextStatusWithPolling(executionId uint64) (*modelExecution.ExecutionResult, error) {
	args := m.Called(executionId)
	return args.Get(0).(*modelExecution.ExecutionResult), args.Error(1)
}

func (m *MockClient) GetExecutionContextStatus(executionId uint64) (*modelExecution.ExecutionResult, error) {
	args := m.Called(executionId)
	return args.Get(0).(*modelExecution.ExecutionResult), args.Error(1)
}

func (m *MockClient) ScheduleJob(name, tags, time, notificationEmails string, group string, jobArgs map[string]string) (string, error) {
	args := m.Called(name, tags, time, notificationEmails, group, jobArgs)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockClient) DescribeScheduledProc(jobID string) (modelSchedule.ScheduledJob, error) {
	args := m.Called(jobID)
	return args.Get(0).(modelSchedule.ScheduledJob), args.Error(1)
}

func (m *MockClient) RemoveScheduledProc(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}
