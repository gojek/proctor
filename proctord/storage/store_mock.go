package storage

import "github.com/stretchr/testify/mock"

type MockStore struct {
	mock.Mock
}

func (m *MockStore) JobsExecutionAuditLog(jobSubmissionStatus, jobName, jobExecutedName, imageName string, jobArgs map[string]string) error {
	args := m.Called(jobSubmissionStatus, jobName, jobExecutedName, imageName, jobArgs)
	return args.Error(0)
}
