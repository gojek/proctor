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
