package execution

import (
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/stretchr/testify/mock"
)

type MockExecutioner struct {
	mock.Mock
}

func (m *MockExecutioner) Execute(jobExecutionAuditLog *postgres.JobsExecutionAuditLog, jobName string, jobArgs map[string]string) (string, error) {
	args := m.Called(jobExecutionAuditLog, jobName, jobArgs)
	return args.String(0), args.Error(1)
}
