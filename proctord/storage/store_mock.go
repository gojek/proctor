package storage

import "github.com/stretchr/testify/mock"

type MockStore struct {
	mock.Mock
}

func (m *MockStore) JobsExecutionAuditLog(jobSubmissionStatus, jobExecutionStatus, jobName, jobExecutedName, imageName string, jobArgs map[string]string) error {
	args := m.Called(jobSubmissionStatus, jobExecutionStatus, jobName, jobExecutedName, imageName, jobArgs)
	return args.Error(0)
}

func (m *MockStore) UpdateJobsExecutionAuditLog(jobSubmittedForExecution, status string) error {
	args := m.Called(jobSubmittedForExecution, status)
	return args.Error(0)
}
