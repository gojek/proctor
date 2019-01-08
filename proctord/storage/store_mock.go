package storage

import "github.com/stretchr/testify/mock"

type MockStore struct {
	mock.Mock
}

func (m *MockStore) JobsExecutionAuditLog(jobSubmissionStatus, jobExecutionStatus, jobName, userEmail, jobExecutedName, imageName string, jobArgs map[string]string) error {
	args := m.Called(jobSubmissionStatus, jobExecutionStatus, jobName, userEmail, jobExecutedName, imageName, jobArgs)
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
