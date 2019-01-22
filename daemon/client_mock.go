package daemon

import (
	proc_metadata "github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/jobs/schedule"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) ListProcs() ([]proc_metadata.Metadata, error) {
	args := m.Called()
	return args.Get(0).([]proc_metadata.Metadata), args.Error(1)
}

func (m *MockClient) ListScheduledProcs() ([]schedule.ScheduledJob, error) {
	args := m.Called()
	return args.Get(0).([]schedule.ScheduledJob), args.Error(1)
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

func (m *MockClient) ScheduleJob(name, tags, time, notificationEmails string,jobArgs map[string]string) (string, error) {
	args := m.Called(name, tags, time, notificationEmails, jobArgs)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockClient) DescribeScheduledProc(jobID string) (schedule.ScheduledJob, error) {
	args := m.Called(jobID)
	return args.Get(0).(schedule.ScheduledJob), args.Error(1)
}

func (m *MockClient) RemoveScheduledProc(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}
