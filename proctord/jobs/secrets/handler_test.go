package secrets

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	utility "proctor/shared/constant"
)

type SecretsHandlerTestSuite struct {
	suite.Suite
	mockSecretsStore   *MockStore
	testSecretsHandler Handler
}

func (suite *SecretsHandlerTestSuite) SetupTest() {
	suite.mockSecretsStore = &MockStore{}

	suite.testSecretsHandler = NewHandler(suite.mockSecretsStore)
}

func (suite *SecretsHandlerTestSuite) TestSuccessfulSecretsUpdation() {
	t := suite.T()

	secret := Secret{
		JobName: "job1",
		Secrets: map[string]string{"k1": "v1", "k2": "v2"},
	}

	requestBody, err := json.Marshal(secret)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/job-secrets", bytes.NewReader(requestBody))
	responseRecorder := httptest.NewRecorder()

	suite.mockSecretsStore.On("CreateOrUpdateJobSecret", secret).Return(nil).Once()

	suite.testSecretsHandler.HandleSubmission()(responseRecorder, req)

	suite.mockSecretsStore.AssertExpectations(t)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
}

func (suite *SecretsHandlerTestSuite) TestSecretsUpdationSecretsMalformedData() {
	t := suite.T()

	requestBody, err := json.Marshal("any-malformed-requ")
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/job-secrets", bytes.NewReader(requestBody))
	responseRecorder := httptest.NewRecorder()

	suite.testSecretsHandler.HandleSubmission()(responseRecorder, req)

	suite.mockSecretsStore.AssertNotCalled(t, "CreateOrUpdateJobSecret", mock.Anything)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, utility.ClientError, responseRecorder.Body.String())
}

func (suite *SecretsHandlerTestSuite) TestSecretsUpdationSecretsStoreFailure() {
	t := suite.T()

	secret := Secret{
		JobName: "job1",
		Secrets: map[string]string{"k1": "v1", "k2": "v2"},
	}

	requestBody, err := json.Marshal(secret)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/job-secrets", bytes.NewReader(requestBody))
	responseRecorder := httptest.NewRecorder()

	suite.mockSecretsStore.On("CreateOrUpdateJobSecret", secret).Return(errors.New("error")).Once()

	suite.testSecretsHandler.HandleSubmission()(responseRecorder, req)

	suite.mockSecretsStore.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, utility.ServerError, responseRecorder.Body.String())
}

func TestSecretsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SecretsHandlerTestSuite))
}
