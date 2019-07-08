package secrets

import (
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateOrUpdateJobSecret(secret Secret) error {
	args := m.Called(secret)
	return args.Error(0)
}

func (m *MockStore) GetJobSecrets(jobName string) (map[string]string, error) {
	args := m.Called(jobName)
	return args.Get(0).(map[string]string), args.Error(1)
}
