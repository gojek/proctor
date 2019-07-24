package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"proctor/internal/app/service/infra/db/types"
	metadataRepository "proctor/internal/app/service/metadata/repository"
	"proctor/internal/app/service/schedule/model"
	scheduleRepository "proctor/internal/app/service/schedule/repository"
	metadataModel "proctor/internal/pkg/model/metadata"
	"proctor/internal/pkg/status"
)

type ScheduleHttpHandlerTestSuite struct {
	suite.Suite
	mockScheduleRepository  *scheduleRepository.MockScheduleRepository
	mockMetadataRepository  *metadataRepository.MockMetadataRepository
	testScheduleHttpHandler ScheduleHttpHandler
}

func (suite *ScheduleHttpHandlerTestSuite) SetupTest() {
	suite.mockScheduleRepository = &scheduleRepository.MockScheduleRepository{}
	suite.mockMetadataRepository = &metadataRepository.MockMetadataRepository{}
	suite.testScheduleHttpHandler = NewScheduleHttpHandler(suite.mockScheduleRepository, suite.mockMetadataRepository)
}

func (suite *ScheduleHttpHandlerTestSuite) TestSuccessfulSchedulePostHttpHandler() {
	t := suite.T()

	scheduleID := uint64(0)
	argsMap := types.Base64Map{
		"COMMAND": "test",
	}
	requestSchedule := model.Schedule{
		JobName:            "test",
		Args:               argsMap,
		Tags:               "test",
		Cron:               "* * * * *",
		UserEmail:          "mrproctor@example.com",
		NotificationEmails: "mrproctor@example.com",
		Group:              "mrproctor",
	}
	responseSchedule := requestSchedule
	responseSchedule.ID = scheduleID
	responseSchedule.Cron = "0 * * * * *"

	requestBody, err := json.Marshal(requestSchedule)
	assert.NoError(t, err)

	responseBody, err := json.Marshal(responseSchedule)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))
	responseRecorder := httptest.NewRecorder()

	suite.mockMetadataRepository.On("GetByName", requestSchedule.JobName).Return(&metadataModel.Metadata{}, nil).Once()
	defer suite.mockMetadataRepository.AssertExpectations(t)

	requestSchedule.Cron = "0 * * * * *"
	suite.mockScheduleRepository.On("Insert", &requestSchedule).Return(0, nil).Once()
	defer suite.mockScheduleRepository.AssertExpectations(t)

	suite.testScheduleHttpHandler.Post()(responseRecorder, req)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
	assert.Equal(t, string(responseBody), responseRecorder.Body.String())
}

func (suite *ScheduleHttpHandlerTestSuite) TestErrorSchedulePostHttpHandler() {
	t := suite.T()

	argsMap := types.Base64Map{
		"COMMAND": "test",
	}
	requestSchedule := model.Schedule{
		JobName:            "test",
		Args:               argsMap,
		Tags:               "test",
		Cron:               "* * * * *",
		UserEmail:          "mrproctor@example.com",
		NotificationEmails: "mrproctor@example.com",
		Group:              "mrproctor",
	}
	tagMissingSchedule := requestSchedule
	tagMissingSchedule.Tags = ""
	cronFormatInvalidSchedule := requestSchedule
	cronFormatInvalidSchedule.Cron = "test"
	emailInvalidSchedule := requestSchedule
	emailInvalidSchedule.NotificationEmails = "test"
	groupMissingSchedule := requestSchedule
	groupMissingSchedule.Group = ""

	requestBody, err := json.Marshal(requestSchedule)
	assert.NoError(t, err)
	tagMissingRequestBody, err := json.Marshal(tagMissingSchedule)
	assert.NoError(t, err)
	cronFormatInvalidRequestBody, err := json.Marshal(cronFormatInvalidSchedule)
	assert.NoError(t, err)
	emailInvalidRequestBody, err := json.Marshal(emailInvalidSchedule)
	assert.NoError(t, err)
	groupMissingRequestBody, err := json.Marshal(groupMissingSchedule)
	assert.NoError(t, err)

	requestSchedule.Cron = fmt.Sprintf("0 %s", requestSchedule.Cron)

	schedulePostErrorTests := []struct {
		requestBody    []byte
		responseStatus int
		responseBody   status.HandlerStatus
	}{
		{tagMissingRequestBody, http.StatusBadRequest, status.ScheduleTagMissingError},
		{cronFormatInvalidRequestBody, http.StatusBadRequest, status.ScheduleCronFormatInvalidError},
		{emailInvalidRequestBody, http.StatusBadRequest, status.EmailInvalidError},
		{groupMissingRequestBody, http.StatusBadRequest, status.ScheduleGroupMissingError},

		// Metadata not found error
		{requestBody, http.StatusNotFound, status.MetadataNotFoundError},

		// Duplicate schedule job name and args error
		{requestBody, http.StatusConflict, status.ScheduleDuplicateJobNameArgsError},
	}

	// Metadata not found error
	suite.mockMetadataRepository.On("GetByName", requestSchedule.JobName).Return(&metadataModel.Metadata{}, errors.New("redigo: nil returned")).Once()
	// Metadata success
	suite.mockMetadataRepository.On("GetByName", requestSchedule.JobName).Return(&metadataModel.Metadata{}, nil).Once()
	defer suite.mockMetadataRepository.AssertExpectations(t)
	// Schedule duplicate job name and args error
	suite.mockScheduleRepository.On("Insert", &requestSchedule).Return(0, errors.New("duplicate key value violates unique constraint")).Once()
	defer suite.mockScheduleRepository.AssertExpectations(t)

	for _, errorTest := range schedulePostErrorTests {
		req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(errorTest.requestBody))
		responseRecorder := httptest.NewRecorder()

		suite.testScheduleHttpHandler.Post()(responseRecorder, req)

		assert.Equal(t, errorTest.responseStatus, responseRecorder.Code)
		assert.Equal(t, string(errorTest.responseBody), responseRecorder.Body.String())
	}
}

func (suite *ScheduleHttpHandlerTestSuite) TestSuccessfulScheduleGetAllHttpHandler() {
	t := suite.T()

	argsMap := types.Base64Map{
		"COMMAND": "test",
	}
	responseScheduleList := []model.Schedule{
		{
			ID:                 uint64(1),
			JobName:            "test1",
			Args:               argsMap,
			Tags:               "test",
			Cron:               "0 * * * * *",
			UserEmail:          "mrproctor@example.com",
			NotificationEmails: "mrproctor@example.com",
			Group:              "mrproctor",
		},
		{
			ID:                 uint64(2),
			JobName:            "test2",
			Args:               argsMap,
			Tags:               "test",
			Cron:               "0 * * * * *",
			UserEmail:          "mrproctor@example.com",
			NotificationEmails: "mrproctor@example.com",
			Group:              "mrproctor",
		},
	}
	responseBody, err := json.Marshal(responseScheduleList)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/schedule", bytes.NewReader([]byte("")))
	responseRecorder := httptest.NewRecorder()

	suite.mockScheduleRepository.On("GetAllEnabled").Return(responseScheduleList, nil).Once()
	defer suite.mockScheduleRepository.AssertExpectations(t)

	suite.testScheduleHttpHandler.GetAll()(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	assert.Equal(t, string(responseBody), responseRecorder.Body.String())
}

func (suite *ScheduleHttpHandlerTestSuite) TestErrorScheduleGetAllHttpHandler() {
	t := suite.T()

	scheduleGetAllErrorTests := []struct {
		responseStatus int
		responseBody   status.HandlerStatus
	}{
		// Error fetching
		{http.StatusInternalServerError, status.GenericServerError},
		// Empty schedule list
		{http.StatusNoContent, ""},
	}

	suite.mockScheduleRepository.On("GetAllEnabled").Return([]model.Schedule{}, errors.New("test")).Once()
	suite.mockScheduleRepository.On("GetAllEnabled").Return([]model.Schedule{}, nil).Once()
	defer suite.mockScheduleRepository.AssertExpectations(t)

	for _, errorTest := range scheduleGetAllErrorTests {
		req := httptest.NewRequest("GET", "/schedule", bytes.NewReader([]byte("")))
		responseRecorder := httptest.NewRecorder()

		suite.testScheduleHttpHandler.GetAll()(responseRecorder, req)

		assert.Equal(t, errorTest.responseStatus, responseRecorder.Code)
		assert.Equal(t, string(errorTest.responseBody), responseRecorder.Body.String())
	}
}

func (suite *ScheduleHttpHandlerTestSuite) TestSuccessfulScheduleGetHttpHandler() {
	t := suite.T()

	scheduleID := uint64(1)
	argsMap := types.Base64Map{
		"COMMAND": "test",
	}
	responseSchedule := model.Schedule{
		ID:                 scheduleID,
		JobName:            "test1",
		Args:               argsMap,
		Tags:               "test",
		Cron:               "0 * * * * *",
		UserEmail:          "mrproctor@example.com",
		NotificationEmails: "mrproctor@example.com",
		Group:              "mrproctor",
	}

	responseBody, err := json.Marshal(responseSchedule)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/schedule", bytes.NewReader([]byte("")))
	req = mux.SetURLVars(req, map[string]string{"scheduleID": fmt.Sprint(scheduleID)})
	responseRecorder := httptest.NewRecorder()

	suite.mockScheduleRepository.On("GetByID", scheduleID).Return(&responseSchedule, nil).Once()
	defer suite.mockScheduleRepository.AssertExpectations(t)

	suite.testScheduleHttpHandler.Get()(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	assert.Equal(t, string(responseBody), responseRecorder.Body.String())
}

func (suite *ScheduleHttpHandlerTestSuite) TestErrorScheduleGetHttpHandler() {
	t := suite.T()

	scheduleID := uint64(1)

	scheduleGetErrorTests := []struct {
		requestParams  string
		responseStatus int
		responseBody   status.HandlerStatus
	}{
		{"test", http.StatusBadRequest, status.PathParameterError},
		{"1", http.StatusBadRequest, status.ScheduleIDInvalidError},
		// Schedule fetch error
		{"1", http.StatusInternalServerError, status.GenericServerError},
	}

	suite.mockScheduleRepository.On("GetByID", scheduleID).Return(&model.Schedule{}, errors.New("invalid input syntax")).Once()
	suite.mockScheduleRepository.On("GetByID", scheduleID).Return(&model.Schedule{}, errors.New("")).Once()
	defer suite.mockScheduleRepository.AssertExpectations(t)

	for _, errorTest := range scheduleGetErrorTests {
		req := httptest.NewRequest("GET", "/schedule", bytes.NewReader([]byte("")))
		req = mux.SetURLVars(req, map[string]string{"scheduleID": errorTest.requestParams})
		responseRecorder := httptest.NewRecorder()

		suite.testScheduleHttpHandler.Get()(responseRecorder, req)

		assert.Equal(t, errorTest.responseStatus, responseRecorder.Code)
		assert.Equal(t, string(errorTest.responseBody), responseRecorder.Body.String())
	}
}

func (suite *ScheduleHttpHandlerTestSuite) TestSuccessfulScheduleDeleteHttpHandler() {
	t := suite.T()

	scheduleID := uint64(1)

	req := httptest.NewRequest("GET", "/schedule", bytes.NewReader([]byte("")))
	req = mux.SetURLVars(req, map[string]string{"scheduleID": fmt.Sprint(scheduleID)})
	responseRecorder := httptest.NewRecorder()

	suite.mockScheduleRepository.On("Delete", scheduleID).Return(nil).Once()
	defer suite.mockScheduleRepository.AssertExpectations(t)

	suite.testScheduleHttpHandler.Delete()(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	assert.Equal(t, string(status.ScheduleDeleteSuccess), responseRecorder.Body.String())
}

func (suite *ScheduleHttpHandlerTestSuite) TestErrorScheduleDeleteHttpHandler() {
	t := suite.T()

	scheduleID := uint64(1)

	scheduleDeleteErrorTests := []struct {
		requestParams  string
		responseStatus int
		responseBody   status.HandlerStatus
	}{
		{"test", http.StatusBadRequest, status.PathParameterError},
		{"1", http.StatusBadRequest, status.ScheduleIDInvalidError},
		// Schedule delete error
		{"1", http.StatusInternalServerError, status.GenericServerError},
	}

	suite.mockScheduleRepository.On("Delete", scheduleID).Return(errors.New("invalid input syntax")).Once()
	suite.mockScheduleRepository.On("Delete", scheduleID).Return(errors.New("")).Once()
	defer suite.mockScheduleRepository.AssertExpectations(t)

	for _, errorTest := range scheduleDeleteErrorTests {
		req := httptest.NewRequest("DELETE", "/schedule", bytes.NewReader([]byte("")))
		req = mux.SetURLVars(req, map[string]string{"scheduleID": errorTest.requestParams})
		responseRecorder := httptest.NewRecorder()

		suite.testScheduleHttpHandler.Delete()(responseRecorder, req)

		assert.Equal(t, errorTest.responseStatus, responseRecorder.Code)
		assert.Equal(t, string(errorTest.responseBody), responseRecorder.Body.String())
	}
}

func TestScheduleHttpHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleHttpHandlerTestSuite))
}
