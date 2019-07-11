package repository

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"proctor/internal/app/service/infra/db/redis"
	modelMetadata "proctor/internal/pkg/model/metadata"
	"testing"
)

type MetadataRepositoryTestSuite struct {
	suite.Suite
	mockRedisClient   *redis.MockClient
	testMetadataStore MetadataRepository
}

func (s *MetadataRepositoryTestSuite) SetupTest() {
	s.mockRedisClient = &redis.MockClient{}

	s.testMetadataStore = NewMetadataRepository(s.mockRedisClient)
}

func (s *MetadataRepositoryTestSuite) TestSave() {
	t := s.T()

	metadata := modelMetadata.Metadata{
		Name:         "any-name",
		ImageName:    "any-image-name",
		Description:  "any-description",
		Author:       "Test User<testuser@example.com>",
		Contributors: "Test User<testuser@example.com>, Test Admin<testadmin@example.com>",
		Organization: "Test Org",
	}

	jsonData, err := json.Marshal(metadata)
	assert.NoError(t, err)

	s.mockRedisClient.On("SET", "any-name-metadata", jsonData).Return(nil).Once()

	err = s.testMetadataStore.Save(metadata)
	assert.NoError(t, err)
	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataRepositoryTestSuite) TestSaveFailure() {
	t := s.T()

	metadata := modelMetadata.Metadata{}

	expectedError := errors.New("any-error")
	s.mockRedisClient.On("SET", mock.Anything, mock.Anything).Return(expectedError).Once()

	err := s.testMetadataStore.Save(metadata)
	assert.EqualError(t, err, "any-error")
	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataRepositoryTestSuite) TestGetAll() {
	t := s.T()

	metadata1 := modelMetadata.Metadata{
		Name:         "job1",
		ImageName:    "job1-image-name",
		Description:  "desc1",
		Author:       "Test User<testuser@example.com",
		Contributors: "Test User<testuser@example.com",
		Organization: "Test Org",
	}

	metadata2 := modelMetadata.Metadata{
		Name:         "job2",
		ImageName:    "job2-image-name",
		Description:  "desc2",
		Author:       "Test User 2<testuser2@example.com",
		Contributors: "Test User 2<testuser2@example.com",
		Organization: "Test Org2",
	}

	s.mockRedisClient.On("KEYS", "*-metadata").Return(
		[]string{"job1-metadata", "job2-metadata"}, nil).Once()

	jsonMetadata1, err := json.Marshal(metadata1)
	assert.NoError(t, err)
	jsonMetadata2, err := json.Marshal(metadata2)
	assert.NoError(t, err)
	values := [][]byte{jsonMetadata1, jsonMetadata2}

	keys := []string{"job1-metadata", "job2-metadata"}
	jobKeys := make([]interface{}, len(keys))
	for i := range keys {
		jobKeys[i] = keys[i]
	}
	s.mockRedisClient.On("MGET", jobKeys...).Return(values, nil).Once()

	metadataSlice, err := s.testMetadataStore.GetAll()
	assert.NoError(t, err)

	assert.EqualValues(t, []modelMetadata.Metadata{metadata1, metadata2}, metadataSlice)
	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataRepositoryTestSuite) TestGetAllFailure() {
	t := s.T()

	s.mockRedisClient.On("KEYS", "*-metadata").Return([]string{}, errors.New("error")).Once()

	_, err := s.testMetadataStore.GetAll()
	assert.Error(t, err)

	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataRepositoryTestSuite) TestGetAllMgetFailure() {
	t := s.T()

	s.mockRedisClient.On("KEYS", "*-metadata").Return(
		[]string{"job1-metadata", "job2-metadata"}, nil).Once()

	keys := []string{"job1-metadata", "job2-metadata"}
	jobKeys := make([]interface{}, len(keys))
	for i := range keys {
		jobKeys[i] = keys[i]
	}
	s.mockRedisClient.On("MGET", jobKeys...).Return([][]byte{}, errors.New("error")).Once()

	_, err := s.testMetadataStore.GetAll()
	assert.Error(t, err)

	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataRepositoryTestSuite) TestGetByName() {
	t := s.T()

	metadata := modelMetadata.Metadata{
		Name:         "job1",
		ImageName:    "job1-image-name",
		Description:  "desc1",
		Author:       "Test User<testuser@example.com",
		Contributors: "Test User<testuser@example.com",
		Organization: "Test Org",
	}
	binaryJobMetadata, err := json.Marshal(metadata)
	assert.NoError(t, err)
	s.mockRedisClient.On("GET", "job1-metadata").Return(binaryJobMetadata, nil).Once()

	jobMetadata, err := s.testMetadataStore.GetByName("job1")
	assert.NoError(t, err)

	assert.EqualValues(t, metadata, *jobMetadata)
	s.mockRedisClient.AssertExpectations(t)
}

func TestMetadataStoreTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataRepositoryTestSuite))
}
