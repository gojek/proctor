package repository

import (
	"github.com/stretchr/testify/mock"
	"proctor/internal/app/service/schedule/model"
)

type MockScheduleRepository struct {
	mock.Mock
}

func (repository *MockScheduleRepository) Insert(context model.Schedule) (uint64, error) {
	args := repository.Called(context)
	return uint64(args.Int(0)), args.Error(1)
}

func (repository *MockScheduleRepository) Delete(id uint64) error {
	args := repository.Called(id)
	return args.Error(0)
}

func (repository *MockScheduleRepository) GetByID(id uint64) (*model.Schedule, error) {
	args := repository.Called(id)
	return args.Get(0).(*model.Schedule), args.Error(1)
}

func (repository *MockScheduleRepository) Disable(id uint64) error {
	args := repository.Called(id)
	return args.Error(0)
}

func (repository *MockScheduleRepository) Enable(id uint64) error {
	args := repository.Called(id)
	return args.Error(0)
}

func (repository *MockScheduleRepository) GetByUserEmail(userEmail string) ([]model.Schedule, error) {
	args := repository.Called(userEmail)
	return args.Get(0).([]model.Schedule), args.Error(1)
}

func (repository *MockScheduleRepository) GetByJobName(jobName string) ([]model.Schedule, error) {
	args := repository.Called(jobName)
	return args.Get(0).([]model.Schedule), args.Error(1)
}

func (repository *MockScheduleRepository) GetAllEnabled() ([]model.Schedule, error) {
	args := repository.Called()
	return args.Get(0).([]model.Schedule), args.Error(1)
}

func (repository *MockScheduleRepository) GetAll() ([]model.Schedule, error) {
	args := repository.Called()
	return args.Get(0).([]model.Schedule), args.Error(1)
}

func (repository *MockScheduleRepository) GetEnabledByID(id uint64) (*model.Schedule, error) {
	args := repository.Called(id)
	return args.Get(0).(*model.Schedule), args.Error(1)
}

func (repository *MockScheduleRepository) deleteAll() error {
	args := repository.Called()
	return args.Error(0)
}
