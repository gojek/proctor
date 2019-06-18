package execution

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/urfave/negroni"
	"net/http"
	"net/http/httptest"
	"proctor/proctord/audit"
	"proctor/proctord/storage"
	utility "proctor/shared/constant"
	"testing"
)

type ExecutionHandlerTestSuite struct {
	suite.Suite
	mockAuditor          *audit.MockAuditor
	mockStore            *storage.MockStore
	mockExecutioner      *MockExecutioner
	testExecutionHandler ExecutionHandler

	Client     *http.Client
	TestServer *httptest.Server
}

func (suite *ExecutionHandlerTestSuite) SetupTest() {
	suite.mockAuditor = &audit.MockAuditor{}
	suite.mockStore = &storage.MockStore{}
	suite.mockExecutioner = &MockExecutioner{}
	suite.testExecutionHandler = NewExecutionHandler(suite.mockAuditor, suite.mockStore, suite.mockExecutioner)

	suite.Client = &http.Client{}
	router := mux.NewRouter()
	router.HandleFunc("/jobs/execute/{name}/status", suite.testExecutionHandler.Status()).Methods("GET")
	n := negroni.Classic()
	n.UseHandler(router)
	suite.TestServer = httptest.NewServer(n)
}

func (suite *ExecutionHandlerTestSuite) TestSuccessfulJobExecutionHandler() {
	t := suite.T()

	jobExecutionID := "proctor-ipsum-lorem"
	statusChan := make(chan bool)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var value map[string]string
		err := json.NewDecoder(req.Body).Decode(&value)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)

		assert.Equal(t, jobExecutionID, value["name"])
		assert.Equal(t, utility.JobSucceeded, value["status"])

		statusChan <- true
	}))
	defer ts.Close()

	remoteCallerURL := fmt.Sprintf("%s/status", ts.URL)

	userEmail := "mrproctor@example.com"
	job := Job{
		Name:        "sample-job-name",
		Args:        map[string]string{"argOne": "sample-arg"},
		CallbackURL: remoteCallerURL,
	}

	requestBody, err := json.Marshal(job)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/execute", bytes.NewReader(requestBody))
	req.Header.Set(utility.UserEmailHeaderKey, userEmail)
	responseRecorder := httptest.NewRecorder()

	suite.mockExecutioner.On("Execute", mock.Anything, job.Name, job.Args).Return(jobExecutionID, nil).Once()

	auditingChan := make(chan bool)

	suite.mockAuditor.On("JobsExecutionAndStatus", mock.Anything).Return("", nil).Run(
		func(args mock.Arguments) { auditingChan <- true },
	)
	suite.mockStore.On("GetJobExecutionStatus", jobExecutionID).Return(utility.JobSucceeded, nil).Once()

	suite.testExecutionHandler.Handle()(responseRecorder, req)

	<-auditingChan
	<-statusChan
	suite.mockAuditor.AssertExpectations(t)
	suite.mockExecutioner.AssertExpectations(t)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
	assert.Equal(t, fmt.Sprintf("{ \"name\":\"%s\" }", jobExecutionID), responseRecorder.Body.String())
}

func (suite *ExecutionHandlerTestSuite) TestSuccessfulJobExecutionHandlerWithoutCallbackURL() {
	t := suite.T()

	jobExecutionID := "proctor-ipsum-lorem"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "should not call status callback", "called status callback")
	}))
	defer ts.Close()

	userEmail := "mrproctor@example.com"
	job := Job{
		Name: "sample-job-name",
		Args: map[string]string{"argOne": "sample-arg"},
	}

	requestBody, err := json.Marshal(job)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/execute", bytes.NewReader(requestBody))
	req.Header.Set(utility.UserEmailHeaderKey, userEmail)
	responseRecorder := httptest.NewRecorder()

	suite.mockExecutioner.On("Execute", mock.Anything, job.Name, job.Args).Return(jobExecutionID, nil).Once()

	auditingChan := make(chan bool)

	suite.mockAuditor.On("JobsExecutionAndStatus", mock.Anything).Return("", nil).Run(
		func(args mock.Arguments) { auditingChan <- true },
	)
	suite.mockStore.On("GetJobExecutionStatus", jobExecutionID).Return(utility.JobSucceeded, nil).Once()

	suite.testExecutionHandler.Handle()(responseRecorder, req)

	<-auditingChan
	suite.mockAuditor.AssertExpectations(t)
	suite.mockExecutioner.AssertExpectations(t)

	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
	assert.Equal(t, fmt.Sprintf("{ \"name\":\"%s\" }", jobExecutionID), responseRecorder.Body.String())
}

func (suite *ExecutionHandlerTestSuite) TestJobExecutionOnMalformedRequest() {
	t := suite.T()

	jobExecutionRequest := fmt.Sprintf("{ some-malformed-request }")
	req := httptest.NewRequest("POST", "/execute", bytes.NewReader([]byte(jobExecutionRequest)))
	responseRecorder := httptest.NewRecorder()

	auditingChan := make(chan bool)
	suite.mockAuditor.On("JobsExecutionAndStatus", mock.Anything).Return("", nil).Run(
		func(args mock.Arguments) { auditingChan <- true },
	)

	suite.testExecutionHandler.Handle()(responseRecorder, req)

	<-auditingChan
	suite.mockAuditor.AssertExpectations(t)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, utility.ClientError, responseRecorder.Body.String())
}

func (suite *ExecutionHandlerTestSuite) TestJobExecutionServerFailure() {
	t := suite.T()

	job := Job{
		Name: "sample-job-name",
		Args: map[string]string{"argOne": "sample-arg"},
	}

	requestBody, err := json.Marshal(job)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/execute", bytes.NewReader(requestBody))
	responseRecorder := httptest.NewRecorder()

	suite.mockExecutioner.On("Execute", mock.Anything, job.Name, job.Args).Return("", errors.New("error executing job")).Once()

	auditingChan := make(chan bool)
	suite.mockAuditor.On("JobsExecutionAndStatus", mock.Anything).Return("", nil).Run(
		func(args mock.Arguments) { auditingChan <- true },
	)
	suite.mockAuditor.On("JobsExecutionAndStatus", mock.Anything).Return("", nil).Run(
		func(args mock.Arguments) { auditingChan <- true },
	)

	suite.testExecutionHandler.Handle()(responseRecorder, req)

	<-auditingChan
	suite.mockAuditor.AssertExpectations(t)
	suite.mockExecutioner.AssertExpectations(t)

	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, utility.ServerError, responseRecorder.Body.String())
}

func (suite *ExecutionHandlerTestSuite) TestJobStatusShouldReturn200OnSuccess() {
	t := suite.T()

	jobName := "sample-job-name"

	url := fmt.Sprintf("%s/jobs/execute/%s/status", suite.TestServer.URL, jobName)

	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobSucceeded, nil).Once()

	req, _ := http.NewRequest("GET", url, nil)

	response, _ := suite.Client.Do(req)
	suite.mockStore.AssertExpectations(t)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	jobStatus := buf.String()
	assert.Equal(suite.T(), utility.JobSucceeded, jobStatus)
}

func (suite *ExecutionHandlerTestSuite) TestJobStatusShouldReturn404IfJobStatusIsNotFound() {
	t := suite.T()

	jobName := "sample-job-name"

	url := fmt.Sprintf("%s/jobs/execute/%s/status", suite.TestServer.URL, jobName)

	suite.mockStore.On("GetJobExecutionStatus", jobName).Return("", nil).Once()

	req, _ := http.NewRequest("GET", url, nil)

	response, _ := suite.Client.Do(req)
	suite.mockStore.AssertExpectations(t)
	assert.Equal(suite.T(), http.StatusNotFound, response.StatusCode)
}

func (suite *ExecutionHandlerTestSuite) TestJobStatusShouldReturn200IfJobStatusIsWaiting() {
	t := suite.T()

	jobName := "sample-job-name"

	url := fmt.Sprintf("%s/jobs/execute/%s/status", suite.TestServer.URL, jobName)

	suite.mockStore.On("GetJobExecutionStatus", jobName).Return("WAITING", nil).Once()

	req, _ := http.NewRequest("GET", url, nil)

	response, _ := suite.Client.Do(req)
	suite.mockStore.AssertExpectations(t)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	jobStatus := buf.String()
	assert.Equal(suite.T(), utility.JobWaiting, jobStatus)
}

func (suite *ExecutionHandlerTestSuite) TestJobStatusShouldReturn500OnError() {
	t := suite.T()

	jobName := "sample-job-name"

	url := fmt.Sprintf("%s/jobs/execute/%s/status", suite.TestServer.URL, jobName)

	suite.mockStore.On("GetJobExecutionStatus", jobName).Return("", errors.New("error")).Once()

	req, _ := http.NewRequest("GET", url, nil)

	response, _ := suite.Client.Do(req)
	suite.mockStore.AssertExpectations(t)
	assert.Equal(suite.T(), http.StatusInternalServerError, response.StatusCode)
}

func (suite *ExecutionHandlerTestSuite) TestSendStatusToCallerOnSuccess() {
	t := suite.T()

	jobName := "sample-job-name"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var value map[string]string
		err := json.NewDecoder(req.Body).Decode(&value)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)

		assert.Equal(t, jobName, value["name"])
		assert.Equal(t, utility.JobSucceeded, value["status"])
	}))
	defer ts.Close()

	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobWaiting, nil).Once()
	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobWaiting, nil).Once()
	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobWaiting, nil).Once()
	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobSucceeded, nil).Once()

	remoteCallerURL := fmt.Sprintf("%s/status", ts.URL)

	suite.testExecutionHandler.sendStatusToCaller(remoteCallerURL, jobName)
	suite.mockStore.AssertExpectations(t)
}

func (suite *ExecutionHandlerTestSuite) TestSendStatusToCallerOnFailure() {
	t := suite.T()

	jobName := "sample-job-name"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var value map[string]string
		err := json.NewDecoder(req.Body).Decode(&value)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)

		assert.Equal(t, jobName, value["name"])
		assert.Equal(t, utility.JobFailed, value["status"])
	}))
	defer ts.Close()

	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobWaiting, nil).Once()
	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobWaiting, nil).Once()
	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobWaiting, nil).Once()
	suite.mockStore.On("GetJobExecutionStatus", jobName).Return(utility.JobFailed, nil).Once()

	remoteCallerURL := fmt.Sprintf("%s/status", ts.URL)

	suite.testExecutionHandler.sendStatusToCaller(remoteCallerURL, jobName)
	suite.mockStore.AssertExpectations(t)
}

func (suite *ExecutionHandlerTestSuite) TestSendStatusToCallerOnJobNotFound() {
	t := suite.T()

	jobName := "sample-job-name"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var value map[string]string
		err := json.NewDecoder(req.Body).Decode(&value)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)

		assert.Equal(t, jobName, value["name"])
		assert.Equal(t, utility.JobNotFound, value["status"])
	}))
	defer ts.Close()

	suite.mockStore.On("GetJobExecutionStatus", jobName).Return("", nil).Once()

	remoteCallerURL := fmt.Sprintf("%s/status", ts.URL)

	suite.testExecutionHandler.sendStatusToCaller(remoteCallerURL, jobName)
	suite.mockStore.AssertExpectations(t)
}

func TestExecutionHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutionHandlerTestSuite))
}
