package metadata

import (
	"github.com/stretchr/testify/mock"
	modelMetadata "proctor/internal/pkg/model/metadata"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateOrUpdateJobMetadata(metadata modelMetadata.Metadata) error {
	args := m.Called(metadata)
	return args.Error(0)
}

func (m *MockStore) GetAllJobsMetadata() ([]modelMetadata.Metadata, error) {
	args := m.Called()
	return args.Get(0).([]modelMetadata.Metadata), args.Error(1)
}

func (m *MockStore) GetJobMetadata(jobName string) (*modelMetadata.Metadata, error) {
	args := m.Called(jobName)
	return args.Get(0).(*modelMetadata.Metadata), args.Error(1)
}
