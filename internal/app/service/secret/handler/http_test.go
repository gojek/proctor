package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"proctor/internal/app/service/secret/model"
	"proctor/internal/app/service/secret/repository"
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"proctor/internal/pkg/constant"
)

type SecretsHandlerTestSuite struct {
	suite.Suite
	mockSecretRepository *repository.MockSecretRepository
	secretHandler        SecretHttpHandler
}

func (suite *SecretsHandlerTestSuite) SetupTest() {
	suite.mockSecretRepository = &repository.MockSecretRepository{}

	suite.secretHandler = NewSecretHttpHandler(suite.mockSecretRepository)
}

func (suite *SecretsHandlerTestSuite) TestPostSecretSuccess() {
	t := suite.T()

	secret := model.Secret{
		JobName: "job1",
		Secrets: map[string]string{"k1": "v1", "k2": "v2"},
	}

	requestBody, err := json.Marshal(secret)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/job-secrets", bytes.NewReader(requestBody))
	responseRecorder := httptest.NewRecorder()

	suite.mockSecretRepository.On("Save", secret).Return(nil).Once()

	suite.secretHandler.Post()(responseRecorder, req)

	suite.mockSecretRepository.AssertExpectations(t)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
}

func (suite *SecretsHandlerTestSuite) TestPostSecretsMalformedData() {
	t := suite.T()

	requestBody, err := json.Marshal("any-malformed-requ")
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/job-secrets", bytes.NewReader(requestBody))
	responseRecorder := httptest.NewRecorder()

	suite.secretHandler.Post()(responseRecorder, req)

	suite.mockSecretRepository.AssertNotCalled(t, "Save", mock.Anything)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, constant.ClientError, responseRecorder.Body.String())
}

func (suite *SecretsHandlerTestSuite) TestPostSecretsStoreFailure() {
	t := suite.T()

	secret := model.Secret{
		JobName: "job1",
		Secrets: map[string]string{"k1": "v1", "k2": "v2"},
	}

	requestBody, err := json.Marshal(secret)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/job-secrets", bytes.NewReader(requestBody))
	responseRecorder := httptest.NewRecorder()

	suite.mockSecretRepository.On("Save", secret).Return(errors.New("error")).Once()

	suite.secretHandler.Post()(responseRecorder, req)

	suite.mockSecretRepository.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, constant.ServerError, responseRecorder.Body.String())
}

func TestSecretsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SecretsHandlerTestSuite))
}
