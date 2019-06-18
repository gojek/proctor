package metadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"proctor/shared/model/metadata/env"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"proctor/shared/constant"
	modelMetadata "proctor/shared/model/metadata"
)

type MetadataHandlerTestSuite struct {
	suite.Suite
	mockStore           *MockStore
	testMetadataHandler Handler
	serverError         string
}

func (s *MetadataHandlerTestSuite) SetupTest() {
	s.mockStore = &MockStore{}

	s.testMetadataHandler = NewHandler(s.mockStore)

	s.serverError = "Something went wrong"
}

func (s *MetadataHandlerTestSuite) TestSuccessfulMetadataSubmission() {
	t := s.T()

	secrets := []env.VarMetadata{
		env.VarMetadata{
			Name:        "SAMPLE_SECRET",
			Description: "description of secret",
		},
	}
	args := []env.VarMetadata{
		env.VarMetadata{
			Name:        "SAMPLE_ARG",
			Description: "description of arg",
		},
	}
	envVars := env.Vars{
		Secrets: secrets,
		Args:    args,
	}
	metadata := modelMetadata.Metadata{
		Name:             "run-sample",
		Description:      "This is a hello world script",
		ImageName:        "proctor-jobs-run-sample",
		EnvVars:          envVars,
		AuthorizedGroups: []string{"group_one", "group_two"},
		Author:           "Test User<testuser@example.com>",
		Contributors:     "Test User<testuser@example.com>",
		Organization:     "Test Org",
	}

	jobsMetadata := []modelMetadata.Metadata{metadata}

	metadataSubmissionRequestBody, err := json.Marshal(jobsMetadata)
	assert.NoError(t, err)
	req := httptest.NewRequest("PUT", "/jobs/metadata", bytes.NewReader(metadataSubmissionRequestBody))
	responseRecorder := httptest.NewRecorder()

	s.mockStore.On("CreateOrUpdateJobMetadata", metadata).Return(nil).Once()

	s.testMetadataHandler.HandleSubmission()(responseRecorder, req)

	s.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
}

func (s *MetadataHandlerTestSuite) TestJobMetadataSubmissionMalformedRequest() {
	t := s.T()

	jobMetadataSubmissionRequest := fmt.Sprintf("{ some-malformed-reque")
	req := httptest.NewRequest("PUT", "/jobs/metadata", bytes.NewReader([]byte(jobMetadataSubmissionRequest)))
	responseRecorder := httptest.NewRecorder()

	s.testMetadataHandler.HandleSubmission()(responseRecorder, req)

	s.mockStore.AssertNotCalled(t, "CreateOrUpdateJobMetadata", mock.Anything)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, constant.ClientError, responseRecorder.Body.String())
}

func (s *MetadataHandlerTestSuite) TestJobMetadataSubmissionForStoreFailure() {
	t := s.T()

	metadata := modelMetadata.Metadata{}

	jobMetadata := []modelMetadata.Metadata{metadata}

	metadataSubmissionRequestBody, err := json.Marshal(jobMetadata)
	assert.NoError(t, err)
	req := httptest.NewRequest("PUT", "/jobs/metadata", bytes.NewReader(metadataSubmissionRequestBody))
	responseRecorder := httptest.NewRecorder()

	s.mockStore.On("CreateOrUpdateJobMetadata", metadata).Return(errors.New("error")).Once()

	s.testMetadataHandler.HandleSubmission()(responseRecorder, req)

	s.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, constant.ServerError, responseRecorder.Body.String())
}

func (s *MetadataHandlerTestSuite) TestHandleBulkDisplay() {
	t := s.T()

	req := httptest.NewRequest("GET", "/jobs/metadata", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	jobsMetadata := []modelMetadata.Metadata{}
	s.mockStore.On("GetAllJobsMetadata").Return(jobsMetadata, nil).Once()

	s.testMetadataHandler.HandleBulkDisplay()(responseRecorder, req)

	s.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	expectedJobDetails, err := json.Marshal(jobsMetadata)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobDetails, responseRecorder.Body.Bytes())
}

func (s *MetadataHandlerTestSuite) TestHandleBulkDisplayStoreFailure() {
	t := s.T()

	req := httptest.NewRequest("GET", "/jobs/metadata", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	jobsMetadata := []modelMetadata.Metadata{}
	s.mockStore.On("GetAllJobsMetadata").Return(jobsMetadata, errors.New("error")).Once()

	s.testMetadataHandler.HandleBulkDisplay()(responseRecorder, req)

	s.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, constant.ServerError, responseRecorder.Body.String())
}

func TestMetadataHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataHandlerTestSuite))
}
