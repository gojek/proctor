package repository

import (
	"github.com/stretchr/testify/mock"
	"proctor/internal/app/service/secret/model"
)

type MockSecretRepository struct {
	mock.Mock
}

func (m *MockSecretRepository) Save(secret model.Secret) error {
	args := m.Called(secret)
	return args.Error(0)
}

func (m *MockSecretRepository) GetByJobName(jobName string) (map[string]string, error) {
	args := m.Called(jobName)
	return args.Get(0).(map[string]string), args.Error(1)
}
