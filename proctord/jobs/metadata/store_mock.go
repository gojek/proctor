package metadata

import (
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateOrUpdateJobMetadata(metadata Metadata) error {
	args := m.Called(metadata)
	return args.Error(0)
}

func (m *MockStore) GetAllJobsMetadata() ([]Metadata, error) {
	args := m.Called()
	return args.Get(0).([]Metadata), args.Error(1)
}

func (m *MockStore) GetJobMetadata(jobName string) (*Metadata, error) {
	args := m.Called(jobName)
	return args.Get(0).(*Metadata), args.Error(1)
}
