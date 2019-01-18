package storage

import (
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) AuditJobsExecution(jobsExecutionAuditLog *postgres.JobsExecutionAuditLog) error {
	args := m.Called(jobsExecutionAuditLog)
	return args.Error(0)
}

func (m *MockStore) UpdateJobsExecutionAuditLog(JobNameSubmittedForExecution, status string) error {
	args := m.Called(JobNameSubmittedForExecution, status)
	return args.Error(0)
}

func (m *MockStore) GetJobExecutionStatus(jobName string) (string, error) {
	args := m.Called(jobName)
	return args.String(0), args.Error(1)
}

func (m *MockStore) InsertScheduledJob(jobName, tags, time, notificationEmails, userEmail string, jobArgs map[string]string) (string, error) {
	args := m.Called(jobName, tags, time, notificationEmails, userEmail, jobArgs)
	return args.String(0), args.Error(1)
}

func (m *MockStore) GetScheduledJobs() ([]postgres.JobsSchedule, error) {
	args := m.Called()
	return args.Get(0).([]postgres.JobsSchedule), args.Error(1)
}

func (m *MockStore) GetEnabledScheduledJobs() ([]postgres.JobsSchedule, error) {
	args := m.Called()
	return args.Get(0).([]postgres.JobsSchedule), args.Error(1)
}
