package execution

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockExecutioner struct {
	mock.Mock
}

func (m *MockExecutioner) Execute(ctx context.Context, jobName, userEmail string, jobArgs map[string]string) (string, error) {
	args := m.Called(ctx, jobName, userEmail, jobArgs)
	return args.String(0), args.Error(1)
}
