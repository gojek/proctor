package mail

import (
	"github.com/stretchr/testify/mock"
)

type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) Send(jobName, jobExecutionID, jobExecutionStatus string, jobArgs map[string]string, recipients []string) error {
	args := m.Called(jobName, jobExecutionID, jobExecutionStatus, jobArgs, recipients)
	return args.Error(0)
}
