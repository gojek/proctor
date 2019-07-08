package schedule

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	jobMetadata "proctor/internal/app/proctord/jobs/metadata"
	"proctor/internal/app/proctord/storage"
	"proctor/internal/app/proctord/storage/postgres"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"proctor/internal/pkg/constant"
	modelMetadata "proctor/internal/pkg/model/metadata"
	modelSchedule "proctor/internal/pkg/model/schedule"
)

type SchedulerTestSuite struct {
	suite.Suite
	mockStore         *storage.MockStore
	mockMetadataStore *jobMetadata.MockStore

	testScheduler Scheduler

	Client     *http.Client
	TestServer *httptest.Server
}

func (suite *SchedulerTestSuite) SetupTest() {
	suite.mockMetadataStore = &jobMetadata.MockStore{}
	suite.mockStore = &storage.MockStore{}
	suite.testScheduler = NewScheduler(suite.mockStore, suite.mockMetadataStore)

	suite.Client = &http.Client{}
	router := mux.NewRouter()
	router.HandleFunc("/jobs/schedule/{id}", suite.testScheduler.GetScheduledJob()).Methods("GET")
	router.HandleFunc("/jobs/schedule/{id}", suite.testScheduler.RemoveScheduledJob()).Methods("DELETE")
	n := negroni.Classic()
	n.UseHandler(router)
	suite.TestServer = httptest.NewServer(n)
}

func (suite *SchedulerTestSuite) TestSuccessfulJobScheduling() {
	t := suite.T()

	userEmail := "mrproctor@example.com"
	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "any-job",
		Args:               map[string]string{},
		Time:               "* 2 * * *",
		NotificationEmails: "foo@bar.com,bar@foo.com",
		Tags:               "tag-one,tag-two",
		Group:              "some-group",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))
	req.Header.Set(constant.UserEmailHeaderKey, userEmail)

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&modelMetadata.Metadata{}, nil)
	insertedScheduledJobID := "123"
	suite.mockStore.On("InsertScheduledJob", scheduledJob.Name, scheduledJob.Tags, "0 * 2 * * *", scheduledJob.NotificationEmails, userEmail, scheduledJob.Group, scheduledJob.Args).Return(insertedScheduledJobID, nil)

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	expectedResponse := modelSchedule.ScheduledJob{}
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
	assert.Equal(t, constant.ClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestInvalidCronExpression() {
	t := suite.T()

	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "non-existent",
		Time:               "2 * invalid *",
		NotificationEmails: "foo@bar.com,bar@foo.com",
		Tags:               "tag-one,tag-two",
		Group:              "some-group",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, constant.InvalidCronExpressionClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestInvalidEmailAddress() {
	t := suite.T()

	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "non-existent",
		Time:               "* 2 * * *",
		NotificationEmails: "user-test.com",
		Group:              "some-group",
		Tags:               "tag-one,tag-two",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, constant.InvalidEmailIdClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestInvalidTag() {
	t := suite.T()

	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "non-existent",
		Time:               "* 2 * * *",
		Group:              "some-group",
		NotificationEmails: "user@proctor.com",
		Tags:               "",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, constant.InvalidTagError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestInvalidGroupName() {
	t := suite.T()

	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "non-existent",
		Time:               "* 2 * * *",
		NotificationEmails: "user@proctor.com",
		Tags:               "backup",
		Group:              "",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, constant.GroupNameMissingError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestNonExistentJobScheduling() {
	t := suite.T()

	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "non-existent",
		Time:               "* 2 * * *",
		NotificationEmails: "foo@bar.com,bar@foo.com",
		Tags:               "tag-one,tag-two",
		Group:              "some-group",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&modelMetadata.Metadata{}, errors.New("redigo: nil returned"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, constant.NonExistentProcClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestErrorFetchingJobMetadata() {
	t := suite.T()

	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "non-existent",
		Time:               "* 2 * * *",
		NotificationEmails: "foo@bar.com,bar@foo.com",
		Group:              "some-group",
		Tags:               "tag-one,tag-two",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&modelMetadata.Metadata{}, errors.New("any error"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, constant.ServerError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestUniqnessConstrainOnJobNameAndArg() {
	t := suite.T()

	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "non-existent",
		Time:               "* 2 * * *",
		NotificationEmails: "foo@bar.com,bar@foo.com",
		Tags:               "tag-one,tag-two",
		Group:              "group1",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&modelMetadata.Metadata{}, nil)
	suite.mockStore.On("InsertScheduledJob", scheduledJob.Name, scheduledJob.Tags, "0 * 2 * * *", scheduledJob.NotificationEmails, "", scheduledJob.Group, scheduledJob.Args).Return("", errors.New("pq: duplicate key value violates unique constraint \"unique_jobs_schedule_name_args\""))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusConflict, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, constant.DuplicateJobNameArgsClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestErrorPersistingScheduledJob() {
	t := suite.T()

	scheduledJob := modelSchedule.ScheduledJob{
		Name:               "non-existent",
		Time:               "* 2 * * *",
		NotificationEmails: "foo@bar.com,bar@foo.com",
		Tags:               "tag-one,tag-two",
		Group:              "group",
	}
	requestBody, err := json.Marshal(scheduledJob)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/schedule", bytes.NewReader(requestBody))

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&modelMetadata.Metadata{}, nil)
	suite.mockStore.On("InsertScheduledJob", scheduledJob.Name, scheduledJob.Tags, "0 * 2 * * *", scheduledJob.NotificationEmails, "", scheduledJob.Group, scheduledJob.Args).Return("", errors.New("any-error"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, constant.ServerError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestGetScheduledJobs() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/jobs/schedule", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	scheduledJobsStoreFormat := []postgres.JobsSchedule{
		{
			ID: "some-id",
		},
	}
	suite.mockStore.On("GetEnabledScheduledJobs").Return(scheduledJobsStoreFormat, nil).Once()

	suite.testScheduler.GetScheduledJobs()(responseRecorder, req)

	suite.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var scheduledJobs []modelSchedule.ScheduledJob
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &scheduledJobs)
	assert.NoError(t, err)
	assert.Equal(t, scheduledJobsStoreFormat[0].ID, scheduledJobs[0].ID)
}

func (suite *SchedulerTestSuite) TestGetScheduledJobsWhenNoJobsFound() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/jobs/schedule", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	var scheduledJobsStoreFormat []postgres.JobsSchedule
	suite.mockStore.On("GetEnabledScheduledJobs").Return(scheduledJobsStoreFormat, nil).Once()

	suite.testScheduler.GetScheduledJobs()(responseRecorder, req)

	suite.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusNoContent, responseRecorder.Code)
}

func (suite *SchedulerTestSuite) TestGetScheduledJobsFailure() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/jobs/schedule", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	var scheduledJobs []postgres.JobsSchedule
	suite.mockStore.On("GetEnabledScheduledJobs").Return(scheduledJobs, errors.New("error")).Once()

	suite.testScheduler.GetScheduledJobs()(responseRecorder, req)

	suite.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, constant.ServerError, responseRecorder.Body.String())
}

func (suite *SchedulerTestSuite) TestGetScheduledJobByID() {
	t := suite.T()
	jobID := "some-id"

	scheduledJobsStoreFormat := []postgres.JobsSchedule{
		{
			ID: jobID,
		},
	}
	suite.mockStore.On("GetScheduledJob", jobID).Return(scheduledJobsStoreFormat, nil).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", suite.TestServer.URL, jobID)
	req, _ := http.NewRequest("GET", url, nil)

	response, err := suite.Client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	var scheduledJob modelSchedule.ScheduledJob
	err = json.NewDecoder(response.Body).Decode(&scheduledJob)
	assert.NoError(t, err)
	assert.Equal(t, jobID, scheduledJob.ID)

	suite.mockStore.AssertExpectations(t)
}

func (suite *SchedulerTestSuite) TestGetScheduledJobByIDOnInvalidJobID() {
	t := suite.T()
	jobID := "invalid-job-id"

	var scheduledJobsStoreFormat []postgres.JobsSchedule

	suite.mockStore.On("GetScheduledJob", jobID).Return(scheduledJobsStoreFormat, errors.New(fmt.Sprintf("pq: invalid input syntax for type uuid: \"%s\"", jobID))).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", suite.TestServer.URL, jobID)
	req, _ := http.NewRequest("GET", url, nil)

	response, err := suite.Client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Invalid Job ID", responseBody)

	suite.mockStore.AssertExpectations(t)
}

func (suite *SchedulerTestSuite) TestGetScheduledJobByIDOnInternalServerError() {
	t := suite.T()
	jobID := "job-id"

	var scheduledJobsStoreFormat []postgres.JobsSchedule

	suite.mockStore.On("GetScheduledJob", jobID).Return(scheduledJobsStoreFormat, errors.New("some-error")).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", suite.TestServer.URL, jobID)
	req, _ := http.NewRequest("GET", url, nil)

	response, err := suite.Client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Something went wrong", responseBody)

	suite.mockStore.AssertExpectations(t)
}

func (suite *SchedulerTestSuite) TestGetScheduledJobByIDOnJobIDNotFound() {
	t := suite.T()
	jobID := "absent-job-id"

	var scheduledJobsStoreFormat []postgres.JobsSchedule

	suite.mockStore.On("GetScheduledJob", jobID).Return(scheduledJobsStoreFormat, nil).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", suite.TestServer.URL, jobID)
	req, _ := http.NewRequest("GET", url, nil)

	response, err := suite.Client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Job not found", responseBody)

	suite.mockStore.AssertExpectations(t)
}

func (suite *SchedulerTestSuite) TestRemoveScheduledJobByID() {
	t := suite.T()
	jobID := "some-id"

	suite.mockStore.On("RemoveScheduledJob", jobID).Return(int64(1), nil).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", suite.TestServer.URL, jobID)
	req, _ := http.NewRequest("DELETE", url, nil)

	response, err := suite.Client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Successfully unscheduled Job ID: some-id", responseBody)

	suite.mockStore.AssertExpectations(t)
}

func (suite *SchedulerTestSuite) TestRemoveScheduledJobByIDOnInvalidJobID() {
	t := suite.T()
	jobID := "invalid-job-id"

	suite.mockStore.On("RemoveScheduledJob", jobID).Return(int64(0), errors.New(fmt.Sprintf("pq: invalid input syntax for type uuid: \"%s\"", jobID))).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", suite.TestServer.URL, jobID)
	req, _ := http.NewRequest("DELETE", url, nil)

	response, err := suite.Client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Invalid Job ID", responseBody)

	suite.mockStore.AssertExpectations(t)
}

func (suite *SchedulerTestSuite) TestRemoveScheduledJobByIDOnInternalServerError() {
	t := suite.T()
	jobID := "job-id"

	suite.mockStore.On("RemoveScheduledJob", jobID).Return(int64(0), errors.New("some-error")).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", suite.TestServer.URL, jobID)
	req, _ := http.NewRequest("DELETE", url, nil)

	response, err := suite.Client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Something went wrong", responseBody)

	suite.mockStore.AssertExpectations(t)
}

func (suite *SchedulerTestSuite) TestRemoveScheduledJobByIDOnJobIDNotFound() {
	t := suite.T()
	jobID := "absent-job-id"

	suite.mockStore.On("RemoveScheduledJob", jobID).Return(int64(0), nil).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", suite.TestServer.URL, jobID)
	req, _ := http.NewRequest("DELETE", url, nil)

	response, err := suite.Client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Job not found", responseBody)

	suite.mockStore.AssertExpectations(t)
}

func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
