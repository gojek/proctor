package mail

import (
	"github.com/stretchr/testify/mock"

	executionContextModel "proctor/internal/app/service/execution/model"
	scheduleModel "proctor/internal/app/service/schedule/model"
)

type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) Send(executionContext executionContextModel.ExecutionContext, schedule scheduleModel.Schedule) error {
	args := m.Called(executionContext, schedule)
	return args.Error(0)
}
