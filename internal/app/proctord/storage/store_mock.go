package storage

import (
	"github.com/stretchr/testify/mock"
	"proctor/internal/app/proctord/storage/postgres"
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

func (m *MockStore) InsertScheduledJob(jobName, tags, time, notificationEmails, userEmail, groupName string, jobArgs map[string]string) (string, error) {
	args := m.Called(jobName, tags, time, notificationEmails, userEmail, groupName, jobArgs)
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

func (m *MockStore) GetScheduledJob(jobID string) ([]postgres.JobsSchedule, error) {
	args := m.Called(jobID)
	return args.Get(0).([]postgres.JobsSchedule), args.Error(1)
}

func (m *MockStore) RemoveScheduledJob(jobID string) (int64, error) {
	args := m.Called(jobID)
	return args.Get(0).(int64), args.Error(1)
}
