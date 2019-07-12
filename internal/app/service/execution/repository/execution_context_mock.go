package repository

import (
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/mock"
	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/execution/status"
)

type MockExecutionContextRepository struct {
	mock.Mock
}

func (mockRepository *MockExecutionContextRepository) Insert(context *model.ExecutionContext) (uint64, error) {
	args := mockRepository.Called(context)
	return uint64(args.Int(0)), args.Error(1)
}

func (mockRepository *MockExecutionContextRepository) UpdateJobOutput(executionId uint64, output types.GzippedText) error {
	args := mockRepository.Called(executionId, output)
	return args.Error(0)
}

func (mockRepository *MockExecutionContextRepository) UpdateStatus(executionId uint64, status status.ExecutionStatus) error {
	args := mockRepository.Called(executionId, status)
	return args.Error(0)
}

func (mockRepository *MockExecutionContextRepository) Delete(executionId uint64) error {
	args := mockRepository.Called(executionId)
	return args.Error(0)
}

func (mockRepository *MockExecutionContextRepository) GetById(executionId uint64) (*model.ExecutionContext, error) {
	args := mockRepository.Called(executionId)
	return args.Get(0).(*model.ExecutionContext), args.Error(1)
}

func (mockRepository *MockExecutionContextRepository) GetByEmail(userEmail string) ([]model.ExecutionContext, error) {
	args := mockRepository.Called(userEmail)
	return args.Get(0).([]model.ExecutionContext), args.Error(1)
}

func (mockRepository *MockExecutionContextRepository) GetByJobName(jobName string) ([]model.ExecutionContext, error) {
	args := mockRepository.Called(jobName)
	return args.Get(0).([]model.ExecutionContext), args.Error(1)
}

func (mockRepository *MockExecutionContextRepository) GetByStatus(status string) ([]model.ExecutionContext, error) {
	args := mockRepository.Called(status)
	return args.Get(0).([]model.ExecutionContext), args.Error(1)
}

func (mockRepository *MockExecutionContextRepository) deleteAll() error {
	args := mockRepository.Called()
	return args.Error(0)
}
