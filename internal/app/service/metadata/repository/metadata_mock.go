package repository

import (
	"github.com/stretchr/testify/mock"
	modelMetadata "proctor/internal/pkg/model/metadata"
)

type MockMetadataRepository struct {
	mock.Mock
}

func (m *MockMetadataRepository) Save(metadata modelMetadata.Metadata) error {
	args := m.Called(metadata)
	return args.Error(0)
}

func (m *MockMetadataRepository) GetAll() ([]modelMetadata.Metadata, error) {
	args := m.Called()
	return args.Get(0).([]modelMetadata.Metadata), args.Error(1)
}

func (m *MockMetadataRepository) GetAllByGroups(group []string) ([]modelMetadata.Metadata, error) {
	args := m.Called(group)
	return args.Get(0).([]modelMetadata.Metadata), args.Error(1)
}

func (m *MockMetadataRepository) GetByName(name string) (*modelMetadata.Metadata, error) {
	args := m.Called(name)
	return args.Get(0).(*modelMetadata.Metadata), args.Error(1)
}
