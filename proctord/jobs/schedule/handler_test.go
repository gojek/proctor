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
	"testing"

	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SchedulerTestSuite struct {
	suite.Suite
	mockStore         *storage.MockStore
	mockMetadataStore *metadata.MockStore

	testScheduler Scheduler

	Client     *http.Client
	TestServer *httptest.Server
}

func (suite *SchedulerTestSuite) SetupTest() {
	suite.mockMetadataStore = &metadata.MockStore{}
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
	scheduledJob := ScheduledJob{
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
	req.Header.Set(utility.UserEmailHeaderKey, userEmail)

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, nil)
	insertedScheduledJobID := "123"
	suite.mockStore.On("InsertScheduledJob", scheduledJob.Name, scheduledJob.Tags, "0 * 2 * * *", scheduledJob.NotificationEmails, userEmail,scheduledJob.Group, scheduledJob.Args).Return(insertedScheduledJobID, nil)

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

func (suite *SchedulerTestSuite) TestInvalidCronExpression() {
	t := suite.T()

	scheduledJob := ScheduledJob{
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
	assert.Equal(t, utility.InvalidCronExpressionClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestInvalidEmailAddress() {
	t := suite.T()

	scheduledJob := ScheduledJob{
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
	assert.Equal(t, utility.InvalidEmailIdClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestInvalidTag() {
	t := suite.T()

	scheduledJob := ScheduledJob{
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
	assert.Equal(t, utility.InvalidTagError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestInvalidGroupName() {
	t := suite.T()

	scheduledJob := ScheduledJob{
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
	assert.Equal(t, utility.GroupNameMissingError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestNonExistentJobScheduling() {
	t := suite.T()

	scheduledJob := ScheduledJob{
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

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, errors.New("redigo: nil returned"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, utility.NonExistentProcClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestErrorFetchingJobMetadata() {
	t := suite.T()

	scheduledJob := ScheduledJob{
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

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, errors.New("any error"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, utility.ServerError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestUniqnessConstrainOnJobNameAndArg() {
	t := suite.T()

	scheduledJob := ScheduledJob{
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

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, nil)
	suite.mockStore.On("InsertScheduledJob", scheduledJob.Name, scheduledJob.Tags, "0 * 2 * * *", scheduledJob.NotificationEmails, "",scheduledJob.Group, scheduledJob.Args).Return("", errors.New("pq: duplicate key value violates unique constraint \"unique_jobs_schedule_name_args\""))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusConflict, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, utility.DuplicateJobNameArgsClientError, string(responseBody))
}

func (suite *SchedulerTestSuite) TestErrorPersistingScheduledJob() {
	t := suite.T()

	scheduledJob := ScheduledJob{
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

	suite.mockMetadataStore.On("GetJobMetadata", scheduledJob.Name).Return(&metadata.Metadata{}, nil)
	suite.mockStore.On("InsertScheduledJob", scheduledJob.Name, scheduledJob.Tags, "0 * 2 * * *", scheduledJob.NotificationEmails, "",scheduledJob.Group, scheduledJob.Args).Return("", errors.New("any-error"))

	suite.testScheduler.Schedule()(responseRecorder, req)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	responseBody, _ := ioutil.ReadAll(responseRecorder.Body)
	assert.Equal(t, utility.ServerError, string(responseBody))
}

func (s *SchedulerTestSuite) TestGetScheduledJobs() {
	t := s.T()

	req := httptest.NewRequest("GET", "/jobs/schedule", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	scheduledJobsStoreFormat := []postgres.JobsSchedule{
		postgres.JobsSchedule{
			ID: "some-id",
		},
	}
	s.mockStore.On("GetEnabledScheduledJobs").Return(scheduledJobsStoreFormat, nil).Once()

	s.testScheduler.GetScheduledJobs()(responseRecorder, req)

	s.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var scheduledJobs []ScheduledJob
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &scheduledJobs)
	assert.NoError(t, err)
	assert.Equal(t, scheduledJobsStoreFormat[0].ID, scheduledJobs[0].ID)
}

func (s *SchedulerTestSuite) TestGetScheduledJobsWhenNoJobsFound() {
	t := s.T()

	req := httptest.NewRequest("GET", "/jobs/schedule", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	scheduledJobsStoreFormat := []postgres.JobsSchedule{}
	s.mockStore.On("GetEnabledScheduledJobs").Return(scheduledJobsStoreFormat, nil).Once()

	s.testScheduler.GetScheduledJobs()(responseRecorder, req)

	s.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusNoContent, responseRecorder.Code)
}

func (s *SchedulerTestSuite) TestGetScheduledJobsFailure() {
	t := s.T()

	req := httptest.NewRequest("GET", "/jobs/schedule", bytes.NewReader([]byte{}))
	responseRecorder := httptest.NewRecorder()

	scheduledJobs := []postgres.JobsSchedule{}
	s.mockStore.On("GetEnabledScheduledJobs").Return(scheduledJobs, errors.New("error")).Once()

	s.testScheduler.GetScheduledJobs()(responseRecorder, req)

	s.mockStore.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, utility.ServerError, responseRecorder.Body.String())
}

func (s *SchedulerTestSuite) TestGetScheduledJobByID() {
	t := s.T()
	jobID := "some-id"

	scheduledJobsStoreFormat := []postgres.JobsSchedule{
		postgres.JobsSchedule{
			ID: jobID,
		},
	}
	s.mockStore.On("GetScheduledJob", jobID).Return(scheduledJobsStoreFormat, nil).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", s.TestServer.URL, jobID)
	req, _ := http.NewRequest("GET", url, nil)

	response, err := s.Client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	var scheduledJob ScheduledJob
	err = json.NewDecoder(response.Body).Decode(&scheduledJob)
	assert.NoError(t, err)
	assert.Equal(t, jobID, scheduledJob.ID)

	s.mockStore.AssertExpectations(t)
}

func (s *SchedulerTestSuite) TestGetScheduledJobByIDOnInvalidJobID() {
	t := s.T()
	jobID := "invalid-job-id"

	scheduledJobsStoreFormat := []postgres.JobsSchedule{}

	s.mockStore.On("GetScheduledJob", jobID).Return(scheduledJobsStoreFormat, errors.New(fmt.Sprintf("pq: invalid input syntax for type uuid: \"%s\"", jobID))).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", s.TestServer.URL, jobID)
	req, _ := http.NewRequest("GET", url, nil)

	response, err := s.Client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Invalid Job ID", responseBody)

	s.mockStore.AssertExpectations(t)
}

func (s *SchedulerTestSuite) TestGetScheduledJobByIDOnInternalServerError() {
	t := s.T()
	jobID := "job-id"

	scheduledJobsStoreFormat := []postgres.JobsSchedule{}

	s.mockStore.On("GetScheduledJob", jobID).Return(scheduledJobsStoreFormat, errors.New("some-error")).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", s.TestServer.URL, jobID)
	req, _ := http.NewRequest("GET", url, nil)

	response, err := s.Client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Something went wrong", responseBody)

	s.mockStore.AssertExpectations(t)
}

func (s *SchedulerTestSuite) TestGetScheduledJobByIDOnJobIDNotFound() {
	t := s.T()
	jobID := "absent-job-id"

	scheduledJobsStoreFormat := []postgres.JobsSchedule{}

	s.mockStore.On("GetScheduledJob", jobID).Return(scheduledJobsStoreFormat, nil).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", s.TestServer.URL, jobID)
	req, _ := http.NewRequest("GET", url, nil)

	response, err := s.Client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Job not found", responseBody)

	s.mockStore.AssertExpectations(t)
}

func (s *SchedulerTestSuite) TestRemoveScheduledJobByID() {
	t := s.T()
	jobID := "some-id"

	s.mockStore.On("RemoveScheduledJob", jobID).Return(int64(1), nil).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", s.TestServer.URL, jobID)
	req, _ := http.NewRequest("DELETE", url, nil)

	response, err := s.Client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Successfully unscheduled Job ID: some-id", responseBody)

	s.mockStore.AssertExpectations(t)
}

func (s *SchedulerTestSuite) TestRemoveScheduledJobByIDOnInvalidJobID() {
	t := s.T()
	jobID := "invalid-job-id"

	s.mockStore.On("RemoveScheduledJob", jobID).Return(int64(0), errors.New(fmt.Sprintf("pq: invalid input syntax for type uuid: \"%s\"", jobID))).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", s.TestServer.URL, jobID)
	req, _ := http.NewRequest("DELETE", url, nil)

	response, err := s.Client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Invalid Job ID", responseBody)

	s.mockStore.AssertExpectations(t)
}

func (s *SchedulerTestSuite) TestRemoveScheduledJobByIDOnInternalServerError() {
	t := s.T()
	jobID := "job-id"

	s.mockStore.On("RemoveScheduledJob", jobID).Return(int64(0), errors.New("some-error")).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", s.TestServer.URL, jobID)
	req, _ := http.NewRequest("DELETE", url, nil)

	response, err := s.Client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Something went wrong", responseBody)

	s.mockStore.AssertExpectations(t)
}

func (s *SchedulerTestSuite) TestRemoveScheduledJobByIDOnJobIDNotFound() {
	t := s.T()
	jobID := "absent-job-id"

	s.mockStore.On("RemoveScheduledJob", jobID).Return(int64(0), nil).Once()

	url := fmt.Sprintf("%s/jobs/schedule/%s", s.TestServer.URL, jobID)
	req, _ := http.NewRequest("DELETE", url, nil)

	response, err := s.Client.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	responseBody := buf.String()
	assert.Equal(t, "Job not found", responseBody)

	s.mockStore.AssertExpectations(t)
}

func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
