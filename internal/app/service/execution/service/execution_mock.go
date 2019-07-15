package service

import (
	"github.com/stretchr/testify/mock"
	"proctor/internal/app/service/execution/model"
)

type MockExecutionService struct {
	mock.Mock
}

func (mockService *MockExecutionService) Execute(jobName string, userEmail string, args map[string]string) (*model.ExecutionContext, string, error) {
	arguments := mockService.Called(jobName, userEmail, args)
	return arguments.Get(0).(*model.ExecutionContext), arguments.String(1), arguments.Error(2)
}

func (mockService *MockExecutionService) ExecuteWithCommand(jobName string, userEmail string, args map[string]string, commands []string) (*model.ExecutionContext, string, error) {
	arguments := mockService.Called(jobName, userEmail, args, commands)
	return arguments.Get(0).(*model.ExecutionContext), arguments.String(1), arguments.Error(2)
}

func (mockService *MockExecutionService) save(executionContext *model.ExecutionContext) {
	mockService.Called(executionContext)
}
