package daemon

import (
	"github.com/stretchr/testify/mock"
	modelMetadata "proctor/shared/model/metadata"
	modelSchedule "proctor/shared/model/schedule"
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

func (m *MockClient) ExecuteProc(name string, procArgs map[string]string) (string, error) {
	args := m.Called(name, procArgs)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockClient) StreamProcLogs(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockClient) GetDefinitiveProcExecutionStatus(name string) (string, error) {
	args := m.Called(name)
	return args.Get(0).(string), args.Error(1)
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
