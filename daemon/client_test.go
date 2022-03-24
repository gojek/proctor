package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"proctor/cmd/version"

	"proctor/config"
	"proctor/io"
	"github.com/gorilla/websocket"
	"github.com/thingful/httpmock"

	proc_metadata "proctor/proctord/jobs/metadata"
	"proctor/proctord/jobs/metadata/env"
	"proctor/proctord/utility"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestConnectionError struct {
	message string
	timeout bool
}

func (e TestConnectionError) Error() string   { return e.message }
func (e TestConnectionError) Timeout() bool   { return e.timeout }
func (e TestConnectionError) Temporary() bool { return false }

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

type ClientTestSuite struct {
	suite.Suite
	testClient       Client
	mockConfigLoader *config.MockLoader
	mockPrinter      *io.MockPrinter
}

func (s *ClientTestSuite) SetupTest() {
	s.mockConfigLoader = &config.MockLoader{}
	s.mockPrinter = &io.MockPrinter{}

	s.testClient = NewClient(s.mockPrinter, s.mockConfigLoader)
}

func (s *ClientTestSuite) TestListProcsReturnsListOfProcsWithDetails() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	body := `[ { "name": "job-1", "description": "job description", "image_name": "hub.docker.com/job-1:latest", "env_vars": { "secrets": [ { "name": "SECRET1", "description": "Base64 encoded secret for authentication." } ], "args": [ { "name": "ARG1", "description": "Argument name" } ] } } ]`
	var args = []env.VarMetadata{env.VarMetadata{Name: "ARG1", Description: "Argument name"}}
	var secrets = []env.VarMetadata{env.VarMetadata{Name: "SECRET1", Description: "Base64 encoded secret for authentication."}}
	envVars := env.Vars{Secrets: secrets, Args: args}
	var expectedProcList = []proc_metadata.Metadata{
		proc_metadata.Metadata{
			Name:        "job-1",
			Description: "job description",
			ImageName:   "hub.docker.com/job-1:latest",
			EnvVars:     envVars,
		},
	}

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procList, err := s.testClient.ListProcs()

	assert.NoError(t, err)
	s.mockConfigLoader.AssertExpectations(t)
	assert.Equal(t, expectedProcList, procList)
}

func (s *ClientTestSuite) TestListProcsReturnErrorFromResponseBody() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, `{}`), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procList, err := s.testClient.ListProcs()

	assert.Equal(t, []proc_metadata.Metadata{}, procList)
	assert.Error(t, err)
	s.mockConfigLoader.AssertExpectations(t)
	assert.Equal(t, "Server Error!!!\nStatus Code: 500, Internal Server Error", err.Error())
}

func (s *ClientTestSuite) TestListProcsReturnClientSideTimeoutError() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return nil, TestConnectionError{message: "Unable to reach http://proctor.example.com/", timeout: true}
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procList, err := s.testClient.ListProcs()

	assert.Equal(t, errors.New("Connection Timeout!!!\nGet \"http://proctor.example.com/jobs/metadata\": Unable to reach http://proctor.example.com/\nPlease check your Internet/VPN connection for connectivity to ProctorD."), err)
	assert.Equal(t, []proc_metadata.Metadata{}, procList)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestListProcsReturnClientSideConnectionError() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return nil, TestConnectionError{message: "Unknown Error", timeout: false}
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procList, err := s.testClient.ListProcs()

	assert.Equal(t, errors.New("Network Error!!!\nGet \"http://proctor.example.com/jobs/metadata\": Unknown Error"), err)
	assert.Equal(t, []proc_metadata.Metadata{}, procList)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestListProcsForUnauthorizedUser() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(401, `{}`), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procList, err := s.testClient.ListProcs()

	assert.Equal(t, []proc_metadata.Metadata{}, procList)
	assert.Equal(t, "Unauthorized Access!!!\nPlease check the EMAIL_ID and ACCESS_TOKEN validity in proctor config file.", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestListProcsForUnauthorizedErrorWithConfigMissing() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: ""}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(401, `{}`), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{""},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()
	procList, err := s.testClient.ListProcs()

	assert.Equal(t, []proc_metadata.Metadata{}, procList)
	assert.Equal(t, "Unauthorized Access!!!\nEMAIL_ID or ACCESS_TOKEN is not present in proctor config file.", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestExecuteProc() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	expectedProcResponse := "proctor-777b1dfb-ea27-46d9-b02c-839b75a542e2"
	body := `{ "name": "proctor-777b1dfb-ea27-46d9-b02c-839b75a542e2"}`
	procName := "run-sample"
	procArgs := map[string]string{"SAMPLE_ARG1": "sample-value"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"http://"+proctorConfig.Host+"/jobs/execute",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(201, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	executeProcResponse, err := s.testClient.ExecuteProc(procName, procArgs)

	assert.NoError(t, err)
	assert.Equal(t, expectedProcResponse, executeProcResponse)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestSuccessScheduledJob() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	expectedProcResponse := "8965fce9-5025-43b3-b21c-920c5ff41cd9"
	procName := "run-sample"
	time := "*/1 * * * *"
	notificationEmails := "user@mail.com"
	tags := "db,backup"
	group := "test"
	procArgs := map[string]string{"ARG_ONE": "sample-value"}

	body := `{"id":"8965fce9-5025-43b3-b21c-920c5ff41cd9","name":"run-sample","args":{"ARG_ONE":"sample-value"},"notification_emails":"user@mail.com","time":"*/1 * * * *","tags":"db,backup", "group":"test"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"http://"+proctorConfig.Host+"/jobs/schedule",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(201, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	executeProcResponse, err := s.testClient.ScheduleJob(procName, tags, time, notificationEmails, group, procArgs)

	assert.NoError(t, err)
	assert.Equal(t, expectedProcResponse, executeProcResponse)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestSchedulingAlreadyExistedScheduledJob() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	procName := "run-sample"
	time := "*/1 * * * *"
	notificationEmails := "user@mail.com"
	tags := "db,backup"
	procArgs := map[string]string{"ARG_ONE": "sample-value"}
	group := "testgroup"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"http://"+proctorConfig.Host+"/jobs/schedule",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(409, "Server Error!!!\nStatus Code: 409, Conflict"), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	_, err := s.testClient.ScheduleJob(procName, tags, time, notificationEmails, group, procArgs)
	assert.Equal(t, "Server Error!!!\nStatus Code: 409, Conflict", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestExecuteProcInternalServerError() {
	t := s.T()
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	expectedProcResponse := ""
	procName := "run-sample"
	procArgs := map[string]string{"SAMPLE_ARG1": "sample-value"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"http://"+proctorConfig.Host+"/jobs/execute",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, ""), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()
	executeProcResponse, err := s.testClient.ExecuteProc(procName, procArgs)

	assert.Equal(t, "Server Error!!!\nStatus Code: 500, Internal Server Error", err.Error())
	assert.Equal(t, expectedProcResponse, executeProcResponse)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestExecuteProcUnAuthorized() {
	t := s.T()
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"http://"+proctorConfig.Host+"/jobs/execute",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(401, ""), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	executeProcResponse, err := s.testClient.ExecuteProc("run-sample", map[string]string{"SAMPLE_ARG1": "sample-value"})

	assert.Equal(t, "", executeProcResponse)
	assert.Equal(t, "Unauthorized Access!!!\nPlease check the EMAIL_ID and ACCESS_TOKEN validity in proctor config file.", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestExecuteProcUnAuthorizedWhenEmailAndAccessTokenNotSet() {
	t := s.T()
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"http://"+proctorConfig.Host+"/jobs/execute",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(401, ""), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{""},
				utility.AccessTokenHeaderKey:   []string{""},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	executeProcResponse, err := s.testClient.ExecuteProc("run-sample", map[string]string{"SAMPLE_ARG1": "sample-value"})

	assert.Equal(t, "", executeProcResponse)
	assert.Equal(t, "Unauthorized Access!!!\nEMAIL_ID or ACCESS_TOKEN is not present in proctor config file.", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestExecuteProcsReturnClientSideConnectionError() {
	t := s.T()
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"http://"+proctorConfig.Host+"/jobs/execute",
			func(req *http.Request) (*http.Response, error) {
				return nil, TestConnectionError{message: "Unknown Error", timeout: false}
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	response, err := s.testClient.ExecuteProc("run-sample", map[string]string{"SAMPLE_ARG1": "sample-value"})

	assert.Equal(t, "", response)
	assert.Equal(t, errors.New("Network Error!!!\nPost \"http://proctor.example.com/jobs/execute\": Unknown Error"), err)
	s.mockConfigLoader.AssertExpectations(t)
}

func makeHostname(s string) string {
	return strings.TrimPrefix(s, "http://")
}

func (s *ClientTestSuite) TestLogStreamForAuthorizedUser() {
	t := s.T()
	logStreamAuthorizer := func(t *testing.T) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			assert.Equal(t, "proctor@example.com", r.Header.Get(utility.UserEmailHeaderKey))
			assert.Equal(t, "access-token", r.Header.Get(utility.AccessTokenHeaderKey))
			assert.Equal(t, version.ClientVersion, r.Header.Get(utility.ClientVersionHeaderKey))
			conn, _ := upgrader.Upgrade(w, r, nil)
			defer conn.Close()
		}
	}
	testServer := httptest.NewServer(logStreamAuthorizer(t))
	defer testServer.Close()
	proctorConfig := config.ProctorConfig{Host: makeHostname(testServer.URL), Email: "proctor@example.com", AccessToken: "access-token"}

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	err := s.testClient.StreamProcLogs("test-job-id")
	assert.NoError(t, err)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestLogStreamForBadWebSocketHandshake() {
	t := s.T()
	badWebSocketHandshakeHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {}
	}
	testServer := httptest.NewServer(badWebSocketHandshakeHandler())
	defer testServer.Close()
	proctorConfig := config.ProctorConfig{Host: makeHostname(testServer.URL), Email: "proctor@example.com", AccessToken: "access-token"}

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	errStreamLogs := s.testClient.StreamProcLogs("test-job-id")
	assert.Equal(t, errors.New("websocket: bad handshake"), errStreamLogs)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestLogStreamForUnauthorizedUser() {
	t := s.T()
	unauthorizedUserHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
	testServer := httptest.NewServer(unauthorizedUserHandler())
	defer testServer.Close()
	proctorConfig := config.ProctorConfig{Host: makeHostname(testServer.URL), Email: "proctor@example.com", AccessToken: "access-token"}

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	errStreamLogs := s.testClient.StreamProcLogs("test-job-id")
	assert.Error(t, errors.New(http.StatusText(http.StatusUnauthorized)), errStreamLogs)
	s.mockConfigLoader.AssertExpectations(t)

}

func (s *ClientTestSuite) TestGetDefinitiveProcExecutionStatusForSucceededProcs() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token", ProcExecutionStatusPollCount: 1}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedProcExecutionStatus := utility.JobSucceeded
	responseBody := expectedProcExecutionStatus

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/execute/some-proc-name/status",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, responseBody), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procExecutionStatus, err := s.testClient.GetDefinitiveProcExecutionStatus("some-proc-name")

	assert.NoError(t, err)
	s.mockConfigLoader.AssertExpectations(t)
	assert.Equal(t, expectedProcExecutionStatus, procExecutionStatus)
}

func (s *ClientTestSuite) TestGetDefinitiveProcExecutionStatusForFailedProcs() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token", ProcExecutionStatusPollCount: 1}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedProcExecutionStatus := utility.JobFailed
	responseBody := expectedProcExecutionStatus

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/execute/some-proc-name/status",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, responseBody), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procExecutionStatus, err := s.testClient.GetDefinitiveProcExecutionStatus("some-proc-name")

	assert.NoError(t, err)
	s.mockConfigLoader.AssertExpectations(t)
	assert.Equal(t, expectedProcExecutionStatus, procExecutionStatus)
}

func (s *ClientTestSuite) TestGetDefinitiveProcExecutionStatusForHTTPRequestFailure() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token", ProcExecutionStatusPollCount: 1}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/execute/some-proc-name/status",
			func(req *http.Request) (*http.Response, error) {
				return nil, TestConnectionError{message: "Unable to reach http://proctor.example.com/", timeout: true}
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procExecutionStatus, err := s.testClient.GetDefinitiveProcExecutionStatus("some-proc-name")

	assert.Equal(t, errors.New("Connection Timeout!!!\nGet \"http://proctor.example.com/jobs/execute/some-proc-name/status\": Unable to reach http://proctor.example.com/\nPlease check your Internet/VPN connection for connectivity to ProctorD."), err)
	s.mockConfigLoader.AssertExpectations(t)
	assert.Equal(t, "", procExecutionStatus)
}

func (s *ClientTestSuite) TestGetDefinitiveProcExecutionStatusForNonOKResponse() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token", ProcExecutionStatusPollCount: 1}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/execute/some-proc-name/status",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, ""), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procExecutionStatus, err := s.testClient.GetDefinitiveProcExecutionStatus("some-proc-name")

	assert.Equal(t, errors.New("Server Error!!!\nStatus Code: 500, Internal Server Error"), err)
	s.mockConfigLoader.AssertExpectations(t)
	assert.Equal(t, "", procExecutionStatus)
}

func (s *ClientTestSuite) TestGetDefinitiveProcExecutionStatusWhenPollCountReached() {
	t := s.T()

	expectedRequestsToProctorDCount := 2
	requestsToProctorDCount := 0

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token", ProcExecutionStatusPollCount: expectedRequestsToProctorDCount}

	expectedProcExecutionStatus := utility.JobWaiting
	responseBody := expectedProcExecutionStatus

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/execute/some-proc-name/status",
			func(req *http.Request) (*http.Response, error) {
				requestsToProctorDCount += 1
				return httpmock.NewStringResponse(200, responseBody), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	procExecutionStatus, err := s.testClient.GetDefinitiveProcExecutionStatus("some-proc-name")

	assert.Equal(t, errors.New("No definitive status received for proc name some-proc-name from proctord"), err)
	s.mockConfigLoader.AssertExpectations(t)
	assert.Equal(t, expectedRequestsToProctorDCount, requestsToProctorDCount)
	assert.Equal(t, "", procExecutionStatus)
}

func (s *ClientTestSuite) TestSuccessDescribeScheduledJob() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "8965fce9-5025-43b3-b21c-920c5ff41cd9"
	body := `{"id":"8965fce9-5025-43b3-b21c-920c5ff41cd9","name":"run-sample","args":{"ARG_ONE":"sample-value"},"notification_emails":"user@mail.com","time":"*/1 * * * *","tags":"db,backup"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule/%s", jobID),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	describeScheduledJob, err := s.testClient.DescribeScheduledProc(jobID)

	assert.NoError(t, err)
	assert.Equal(t, jobID, describeScheduledJob.ID)
	assert.Equal(t, "run-sample", describeScheduledJob.Name)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestDescribeScheduledJobWithInvalidJobID() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "invalid-job-id"
	body := "Invalid Job ID"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule/%s", jobID),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(400, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	_, err := s.testClient.DescribeScheduledProc(jobID)

	assert.Equal(t, "Invalid Job ID", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestDescribeScheduledJobWhenJobIDNotFound() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "invalid-job-id"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule/%s", jobID),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(404, "Job not found"), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	_, err := s.testClient.DescribeScheduledProc(jobID)

	assert.Equal(t, "Job not found", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestDescribeScheduledJobWitInternalServerError() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "invalid-job-id"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule/%s", jobID),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, ""), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	_, err := s.testClient.DescribeScheduledProc(jobID)

	assert.Equal(t, "Server Error!!!\nStatus Code: 500, Internal Server Error", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestSuccessListOfScheduledJobs() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "c3e040b1-c2b8-4d23-bebd-246c8b7c6f87"
	body := `[{"id":"c3e040b1-c2b8-4d23-bebd-246c8b7c6f87","name":"run-sample","args":{"ARG2":"bar","ARG3":"test","ARG_ONE1":"foobar"},"notification_emails":"username@mail.com","time":"0 2 * * *","tags":"sample,proctor"}]`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule"),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	listScheduledJobs, err := s.testClient.ListScheduledProcs()

	assert.NoError(t, err)
	assert.Equal(t, jobID, listScheduledJobs[0].ID)
	assert.Equal(t, "run-sample", listScheduledJobs[0].Name)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestSuccessListOfScheduledJobsWhenNoJobsScheduled() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	body := "No scheduled jobs found"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule"),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(204, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	_, err := s.testClient.ListScheduledProcs()

	assert.Equal(t, "No scheduled jobs found", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestSuccessListOfScheduledJobsWhenServerReturnInternalServerError() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule"),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, ""), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	_, err := s.testClient.ListScheduledProcs()

	assert.Equal(t, "Server Error!!!\nStatus Code: 500, Internal Server Error", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestSuccessRemoveScheduledJob() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "8965fce9-5025-43b3-b21c-920c5ff41cd9"
	body := fmt.Sprintf("Sucessfully removed the scheduled job ID: %s", jobID)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"DELETE",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule/%s", jobID),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	err := s.testClient.RemoveScheduledProc(jobID)

	assert.NoError(t, err)
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestRemoveScheduledJobWithInvalidJobID() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "invalid-job-id"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"DELETE",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule/%s", jobID),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(400, "Invalid Job ID"), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	err := s.testClient.RemoveScheduledProc(jobID)

	assert.Equal(t, "Invalid Job ID", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestRemoveScheduledJobWhenJobIDNotFound() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "invalid-job-id"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"DELETE",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule/%s", jobID),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(404, "Job not found"), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	err := s.testClient.RemoveScheduledProc(jobID)

	assert.Equal(t, "Job not found", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}

func (s *ClientTestSuite) TestRemoveScheduledJobWitInternalServerError() {
	t := s.T()

	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	jobID := "invalid-job-id"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"DELETE",
			fmt.Sprintf("http://"+proctorConfig.Host+"/jobs/schedule/%s", jobID),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, ""), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:     []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey:   []string{"access-token"},
				utility.ClientVersionHeaderKey: []string{version.ClientVersion},
			},
		),
	)

	s.mockConfigLoader.On("Load").Return(proctorConfig, config.ConfigError{}).Once()

	err := s.testClient.RemoveScheduledProc(jobID)

	assert.Equal(t, "Server Error!!!\nStatus Code: 500, Internal Server Error", err.Error())
	s.mockConfigLoader.AssertExpectations(t)
}
