package audit

import (
	"proctor/proctord/storage/postgres"
	"github.com/stretchr/testify/mock"
)

type MockAuditor struct {
	mock.Mock
}

func (m *MockAuditor) JobsExecution(JobsExecutionAuditLog *postgres.JobsExecutionAuditLog) {
	m.Called(JobsExecutionAuditLog)
}

func (m *MockAuditor) JobsExecutionAndStatus(JobsExecutionAuditLog *postgres.JobsExecutionAuditLog) {
	m.Called(JobsExecutionAuditLog)
}

func (m *MockAuditor) JobsExecutionStatus(jobExecutionID string) (string, error) {
	args := m.Called(jobExecutionID)
	return args.String(0), args.Error(1)
}
