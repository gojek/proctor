package audit

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAuditor struct {
	mock.Mock
}

func (m *MockAuditor) AuditJobsExecution(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockAuditor) AuditJobExecutionStatus(jobExecutionID string) (string, error) {
	args := m.Called(jobExecutionID)
	return args.String(0), args.Error(1)
}
