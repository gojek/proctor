package schedule

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SchedulerTestSuite struct {
	suite.Suite
	mockStore         *storage.MockStore
	mockMetadataStore *metadata.MockStore

	testScheduler Scheduler
}

func (suite *SchedulerTestSuite) SetupTest() {
	suite.mockMetadataStore = &metadata.MockStore{}
	suite.mockStore = &storage.MockStore{}

	suite.testScheduler = NewScheduler(suite.mockStore, suite.mockMetadataStore)
}

func (suite *SchedulerTestSuite) TestSuccessfulJobScheduling() {
	t := suite.T()

	userEmail := "mrproctor@example.com"
	scheduledJob := ScheduledJob{
		Name:               "any-job",
		Args:               map[string]string{},
		Time:               "* 2 * * *",
		NotificationEmails: "foo@bar.com,bar@foo.com",
		Tags:               "tag-one,tag-two",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))
	req.Header.Set(utility.UserEmailHeaderKey, userEmail)

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, nil)
	insertedScheduledJobID := "123"
	suite.mockStore.On("InsertScheduledJob", scheduledJob.Name, scheduledJob.Tags, scheduledJob.Time, scheduledJob.NotificationEmails, userEmail, scheduledJob.Args).Return(insertedScheduledJobID, nil)

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	expectedResponse := ScheduledJob{}
	err = json.NewDecoder(responseRecorder.Body).Decode(&expectedResponse)
	assert.NoError(t, err)
	assert.Equal(t, insertedScheduledJobID, expectedResponse.ID)
}

func (suite *SchedulerTestSuite) TestBadRequestWhenRequestBodyIsIncorrectForJobScheduling() {
	t := suite.T()

	req := httptest.NewRequest("POST", "/schedule", bytes.NewBuffer([]byte("invalid json")))
	responseRecorder := httptest.NewRecorder()

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, utility.ClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestNonExistentJobScheduling() {
	t := suite.T()

	scheduledJob := ScheduledJob{
		Name: "non-existent",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, errors.New("redigo: nil returned"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, utility.NonExistentProcClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestErrorFetchingJobMetadata() {
	t := suite.T()

	scheduledJob := ScheduledJob{
		Name: "non-existent",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, errors.New("any error"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, utility.ServerError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestErrorPersistingScheduledJob() {
	t := suite.T()

	scheduledJob := ScheduledJob{
		Name: "non-existent",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, nil)
	suite.mockStore.On("InsertScheduledJob", scheduledJob.Name, scheduledJob.Tags, scheduledJob.Time, scheduledJob.NotificationEmails, "", scheduledJob.Args).Return("", errors.New("any-error"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, utility.ServerError, string(responseBody))
}

func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
