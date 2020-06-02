package repository

import (
	"encoding/json"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/app/service/secret/model"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SecretsStoreTestSuite struct {
	suite.Suite
	mockRedisClient *redis.MockClient
	repository      SecretRepository
}

func (s *SecretsStoreTestSuite) SetupTest() {
	s.mockRedisClient = &redis.MockClient{}

	s.repository = NewSecretRepository(s.mockRedisClient)
}

func (s *SecretsStoreTestSuite) TestCreateOrUpdateJobSecret() {
	t := s.T()

	secret := model.Secret{
		JobName: "job1",
		Secrets: map[string]string{"k1": "v1", "k2": "v2"},
	}

	binaryJobSecret, err := json.Marshal(secret.Secrets)
	assert.NoError(t, err)

	s.mockRedisClient.On("SET", "job1-secret", binaryJobSecret).Return(nil).Once()

	err = s.repository.Save(secret)
	assert.NoError(t, err)

	s.mockRedisClient.AssertExpectations(t)
}

func (s *SecretsStoreTestSuite) TestCreateOrUpdateJobSecretRedisFailure() {
	t := s.T()

	s.mockRedisClient.On("SET", mock.Anything, mock.Anything).Return(errors.New("error")).Once()

	err := s.repository.Save(model.Secret{})
	assert.Error(t, err)

	s.mockRedisClient.AssertExpectations(t)
}

func (s *SecretsStoreTestSuite) TestGetJobSecrets() {
	t := s.T()

	jobSecrets := map[string]string{"k1": "v1", "k2": "v2"}

	binaryJobSecrets, err := json.Marshal(jobSecrets)
	assert.NoError(t, err)
	s.mockRedisClient.On("GET", "job1-secret").Return(binaryJobSecrets, nil).Once()

	secrets, err := s.repository.GetByJobName("job1")
	assert.NoError(t, err)

	assert.EqualValues(t, jobSecrets, secrets)
}

func (s *SecretsStoreTestSuite) TestGetJobSecretsRedisFailure() {
	t := s.T()

	s.mockRedisClient.On("GET", "job1-secret").Return([]byte{}, errors.New("error")).Once()

	_, err := s.repository.GetByJobName("job1")
	assert.Error(t, err)
}

func (s *SecretsStoreTestSuite) TestGetJobSecretsCorruptData() {
	t := s.T()

	s.mockRedisClient.On("GET", "job1-secret").Return([]byte("a"), nil).Once()

	_, err := s.repository.GetByJobName("job1")
	assert.Error(t, err)
}

func TestSecretsStoreTestSuite(t *testing.T) {
	suite.Run(t, new(SecretsStoreTestSuite))
}
