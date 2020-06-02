package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	metadataRepository "proctor/internal/app/service/metadata/repository"
	"proctor/internal/app/service/security/middleware"
	"proctor/pkg/auth"
	"testing"

	"proctor/internal/pkg/model/metadata/env"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"proctor/internal/pkg/constant"
	modelMetadata "proctor/internal/pkg/model/metadata"
)

type MetadataHandlerTestSuite struct {
	suite.Suite
	mockRepository      *metadataRepository.MockMetadataRepository
	metadataHTTPHandler MetadataHTTPHandler
	serverError         string
}

func (s *MetadataHandlerTestSuite) SetupTest() {
	s.mockRepository = &metadataRepository.MockMetadataRepository{}

	s.metadataHTTPHandler = NewMetadataHTTPHandler(s.mockRepository)

	s.serverError = "Something went wrong"
}

func (s *MetadataHandlerTestSuite) TestSuccessfulMetadataSubmission() {
	t := s.T()

	secrets := []env.VarMetadata{
		{
			Name:        "SAMPLE_SECRET",
			Description: "description of secret",
		},
	}
	args := []env.VarMetadata{
		{
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
	req := httptest.NewRequest("PUT", "/metadata", bytes.NewReader(metadataSubmissionRequestBody))
	responseRecorder := httptest.NewRecorder()

	s.mockRepository.On("Save", metadata).Return(nil).Once()

	s.metadataHTTPHandler.Post()(responseRecorder, req)

	s.mockRepository.AssertExpectations(t)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
}

func (s *MetadataHandlerTestSuite) TestJobMetadataSubmissionMalformedRequest() {
	t := s.T()

	jobMetadataSubmissionRequest := fmt.Sprintf("{ some-malformed-reque")
	req := httptest.NewRequest("PUT", "/metadata", bytes.NewReader([]byte(jobMetadataSubmissionRequest)))
	responseRecorder := httptest.NewRecorder()

	s.metadataHTTPHandler.Post()(responseRecorder, req)

	s.mockRepository.AssertNotCalled(t, "Save", mock.Anything)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, constant.ClientError, responseRecorder.Body.String())
}

func (s *MetadataHandlerTestSuite) TestJobMetadataSubmissionForStoreFailure() {
	t := s.T()

	metadata := modelMetadata.Metadata{}

	jobMetadata := []modelMetadata.Metadata{metadata}

	metadataSubmissionRequestBody, err := json.Marshal(jobMetadata)
	assert.NoError(t, err)
	req := httptest.NewRequest("PUT", "/metadata", bytes.NewReader(metadataSubmissionRequestBody))
	responseRecorder := httptest.NewRecorder()

	s.mockRepository.On("Save", metadata).Return(errors.New("error")).Once()

	s.metadataHTTPHandler.Post()(responseRecorder, req)

	s.mockRepository.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, constant.ServerError, responseRecorder.Body.String())
}

func (s *MetadataHandlerTestSuite) TestHandleBulkDisplay() {
	t := s.T()

	req := httptest.NewRequest("GET", "/metadata", bytes.NewReader([]byte{}))
	groups := []string{"admin", "migratior"}
	userDetail := &auth.UserDetail{
		Name:   "jasoet",
		Email:  "jasoet@ambyar.com",
		Active: true,
		Groups: groups,
	}

	ctx := context.WithValue(req.Context(), middleware.ContextUserDetailKey, userDetail)
	ctx = context.WithValue(ctx, middleware.ContextAuthEnabled, true)
	req = req.WithContext(ctx)
	responseRecorder := httptest.NewRecorder()

	var jobsMetadata []modelMetadata.Metadata
	s.mockRepository.On("GetAllByGroups", groups).Return(jobsMetadata, nil).Once()

	s.metadataHTTPHandler.GetAll()(responseRecorder, req)

	s.mockRepository.AssertExpectations(t)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	expectedJobDetails, err := json.Marshal(jobsMetadata)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobDetails, responseRecorder.Body.Bytes())
}

func (s *MetadataHandlerTestSuite) TestHandleBulkDisplayWithoutAuth() {
	t := s.T()

	req := httptest.NewRequest("GET", "/metadata", bytes.NewReader([]byte{}))
	ctx := context.WithValue(req.Context(), middleware.ContextAuthEnabled, false)
	req = req.WithContext(ctx)
	responseRecorder := httptest.NewRecorder()

	var jobsMetadata []modelMetadata.Metadata
	s.mockRepository.On("GetAll").Return(jobsMetadata, nil).Once()

	s.metadataHTTPHandler.GetAll()(responseRecorder, req)

	s.mockRepository.AssertExpectations(t)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	expectedJobDetails, err := json.Marshal(jobsMetadata)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobDetails, responseRecorder.Body.Bytes())
}

func (s *MetadataHandlerTestSuite) TestHandleBulkDisplayStoreFailure() {
	t := s.T()

	req := httptest.NewRequest("GET", "/metadata", bytes.NewReader([]byte{}))
	groups := []string{"admin", "migratior"}
	userDetail := &auth.UserDetail{
		Name:   "jasoet",
		Email:  "jasoet@ambyar.com",
		Active: true,
		Groups: groups,
	}

	ctx := context.WithValue(req.Context(), middleware.ContextUserDetailKey, userDetail)
	ctx = context.WithValue(ctx, middleware.ContextAuthEnabled, true)
	req = req.WithContext(ctx)
	responseRecorder := httptest.NewRecorder()

	jobsMetadata := []modelMetadata.Metadata{}
	s.mockRepository.On("GetAllByGroups", groups).Return(jobsMetadata, errors.New("error")).Once()

	s.metadataHTTPHandler.GetAll()(responseRecorder, req)

	s.mockRepository.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, constant.ServerError, responseRecorder.Body.String())
}

func TestMetadataHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataHandlerTestSuite))
}
