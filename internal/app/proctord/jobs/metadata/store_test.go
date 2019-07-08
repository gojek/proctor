package metadata

import (
	"encoding/json"
	"errors"
	"proctor/internal/app/proctord/redis"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	modelMetadata "proctor/internal/pkg/model/metadata"
)

type MetadataStoreTestSuite struct {
	suite.Suite
	mockRedisClient   *redis.MockClient
	testMetadataStore Store
}

func (s *MetadataStoreTestSuite) SetupTest() {
	s.mockRedisClient = &redis.MockClient{}

	s.testMetadataStore = NewStore(s.mockRedisClient)
}

func (s *MetadataStoreTestSuite) TestCreateOrUpdateJobMetadata() {
	t := s.T()

	metadata := modelMetadata.Metadata{
		Name:         "any-name",
		ImageName:    "any-image-name",
		Description:  "any-description",
		Author:       "Test User<testuser@example.com>",
		Contributors: "Test User<testuser@example.com>, Test Admin<testadmin@example.com>",
		Organization: "Test Org",
	}

	binaryJobMetadata, err := json.Marshal(metadata)
	assert.NoError(t, err)

	s.mockRedisClient.On("SET", "any-name-metadata", binaryJobMetadata).Return(nil).Once()

	err = s.testMetadataStore.CreateOrUpdateJobMetadata(metadata)
	assert.NoError(t, err)
	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataStoreTestSuite) TestCreateOrUpdateJobMetadataForRedisClientFailure() {
	t := s.T()

	metadata := modelMetadata.Metadata{}

	expectedError := errors.New("any-error")
	s.mockRedisClient.On("SET", mock.Anything, mock.Anything).Return(expectedError).Once()

	err := s.testMetadataStore.CreateOrUpdateJobMetadata(metadata)
	assert.EqualError(t, err, "any-error")
	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataStoreTestSuite) TestGetAllJobsMetadata() {
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

	binaryJobMetadata1, err := json.Marshal(metadata1)
	assert.NoError(t, err)
	binaryJobMetadata2, err := json.Marshal(metadata2)
	assert.NoError(t, err)
	values := [][]byte{binaryJobMetadata1, binaryJobMetadata2}

	keys := []string{"job1-metadata", "job2-metadata"}
	jobKeys := make([]interface{}, len(keys))
	for i := range keys {
		jobKeys[i] = keys[i]
	}
	s.mockRedisClient.On("MGET", jobKeys...).Return(values, nil).Once()

	jobMetadata, err := s.testMetadataStore.GetAllJobsMetadata()
	assert.NoError(t, err)

	assert.EqualValues(t, []modelMetadata.Metadata{metadata1, metadata2}, jobMetadata)
	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataStoreTestSuite) TestGetAllJobsMetadataRedisClientKeysFailure() {
	t := s.T()

	s.mockRedisClient.On("KEYS", "*-metadata").Return([]string{}, errors.New("error")).Once()

	_, err := s.testMetadataStore.GetAllJobsMetadata()
	assert.Error(t, err)

	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataStoreTestSuite) TestGetAllJobsMetadataRedisClientMgetFailure() {
	t := s.T()

	s.mockRedisClient.On("KEYS", "*-metadata").Return(
		[]string{"job1-metadata", "job2-metadata"}, nil).Once()

	keys := []string{"job1-metadata", "job2-metadata"}
	jobKeys := make([]interface{}, len(keys))
	for i := range keys {
		jobKeys[i] = keys[i]
	}
	s.mockRedisClient.On("MGET", jobKeys...).Return([][]byte{}, errors.New("error")).Once()

	_, err := s.testMetadataStore.GetAllJobsMetadata()
	assert.Error(t, err)

	s.mockRedisClient.AssertExpectations(t)
}

func (s *MetadataStoreTestSuite) TestGetJobMetadata() {
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

	jobMetadata, err := s.testMetadataStore.GetJobMetadata("job1")
	assert.NoError(t, err)

	assert.EqualValues(t, metadata, *jobMetadata)
	s.mockRedisClient.AssertExpectations(t)
}

func TestMetadataStoreTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataStoreTestSuite))
}
