package service

import (
	"github.com/stretchr/testify/mock"
	"io"
	"proctor/internal/app/service/execution/model"
	"time"
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

func (mockService *MockExecutionService) update(executionContext model.ExecutionContext) {
	mockService.Called(executionContext)
}

func (mockService *MockExecutionService) insertContext(executionContext model.ExecutionContext) {
	mockService.Called(executionContext)
}

func (mockService *MockExecutionService) StreamJobLogs(executionName string, waitTime time.Duration) (io.ReadCloser, error) {
	args := mockService.Called(executionName, waitTime)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
