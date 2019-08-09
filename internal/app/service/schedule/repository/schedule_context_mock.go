package repository

import (
	"github.com/stretchr/testify/mock"
	executionModel "proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/schedule/model"
)

type MockScheduleContextRepository struct {
	mock.Mock
}

func (repository *MockScheduleContextRepository) Insert(context model.ScheduleContext) (*model.ScheduleContext, error) {
	args := repository.Called(context)
	return args.Get(0).(*model.ScheduleContext), args.Error(1)
}

func (repository *MockScheduleContextRepository) Delete(id uint64) error {
	args := repository.Called(id)
	return args.Error(0)
}

func (repository *MockScheduleContextRepository) GetByID(id uint64) (*model.ScheduleContext, error) {
	args := repository.Called(id)
	return args.Get(0).(*model.ScheduleContext), args.Error(1)
}

func (repository *MockScheduleContextRepository) GetContextByScheduleId(scheduleId uint64) ([]executionModel.ExecutionContext, error) {
	args := repository.Called(scheduleId)
	return args.Get(0).([]executionModel.ExecutionContext), args.Error(1)
}

func (repository *MockScheduleContextRepository) GetScheduleByContextId(contextId uint64) (*model.Schedule, error) {
	args := repository.Called(contextId)
	return args.Get(0).(*model.Schedule), args.Error(1)
}

func (repository *MockScheduleContextRepository) deleteAll() error {
	args := repository.Called()
	return args.Error(0)
}
