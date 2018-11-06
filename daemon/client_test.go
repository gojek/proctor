package daemon

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gojektech/proctor/config"
	"github.com/gorilla/websocket"
	"github.com/thingful/httpmock"

	"github.com/gojektech/proctor/proc/env"

	"github.com/gojektech/proctor/proc"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/stretchr/testify/assert"
)

func TestListProcs(t *testing.T) {
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	body := `[ { "name": "job-1", "description": "job description", "image_name": "hub.docker.com/job-1:latest", "env_vars": { "secrets": [ { "name": "SECRET1", "description": "Base64 encoded secret for authentication." } ], "args": [ { "name": "ARG1", "description": "Argument name" } ] } } ]`
	var args = []env.VarMetadata{env.VarMetadata{Name: "ARG1", Description: "Argument name"}}
	var secrets = []env.VarMetadata{env.VarMetadata{Name: "SECRET1", Description: "Base64 encoded secret for authentication."}}
	envVars := env.Vars{Secrets: secrets, Args: args}
	var procListExpected = []proc.Metadata{proc.Metadata{Name: "job-1", Description: "job description", EnvVars: envVars}}

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, body), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:   []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey: []string{"access-token"},
			},
		),
	)

	procList, err := proctorClient.ListProcs()

	assert.NoError(t, err)
	assert.Equal(t, procListExpected, procList)
}

func TestListProcsReturnInternalServerError(t *testing.T) {
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var procListExpected = []proc.Metadata{}

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, `{}`), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:   []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey: []string{"access-token"},
			},
		),
	)

	procList, err := proctorClient.ListProcs()

	assert.Equal(t, procListExpected, procList)
	assert.Error(t, err)
}

func TestListProcsReturnClientSideConnectionError(t *testing.T) {
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)
	connectionTimeOut := "Connection TimeOut"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var procListExpected = []proc.Metadata{}

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return nil, errors.New(connectionTimeOut)
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:   []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey: []string{"access-token"},
			},
		),
	)

	procList, err := proctorClient.ListProcs()

	assert.Equal(t, errors.New("Get http://proctor.example.com/jobs/metadata: Connection TimeOut"), err)
	assert.Equal(t, procListExpected, procList)
}

func TestListProcsForUnauthorizedUser(t *testing.T) {
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var procListExpected = []proc.Metadata{}

	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"http://"+proctorConfig.Host+"/jobs/metadata",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(401, `{}`), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:   []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey: []string{"access-token"},
			},
		),
	)

	procList, err := proctorClient.ListProcs()

	assert.Equal(t, procListExpected, procList)
	assert.Equal(t, err.Error(), http.StatusText(http.StatusUnauthorized))
}

func TestExecuteProc(t *testing.T) {
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)
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
				utility.UserEmailHeaderKey:   []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey: []string{"access-token"},
			},
		),
	)

	executeProcResponse, err := proctorClient.ExecuteProc(procName, procArgs)

	assert.NoError(t, err)
	assert.Equal(t, expectedProcResponse, executeProcResponse)
}

func TestExecuteProcInternalServerError(t *testing.T) {
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)
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
				utility.UserEmailHeaderKey:   []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey: []string{"access-token"},
			},
		),
	)

	executeProcResponse, err := proctorClient.ExecuteProc(procName, procArgs)

	assert.Error(t, err)
	assert.Equal(t, expectedProcResponse, executeProcResponse)
}

func TestExecuteProcUnAuthorized(t *testing.T) {
	proctorConfig := config.ProctorConfig{Host: "proctor.example.com", Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)
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
				return httpmock.NewStringResponse(401, ""), nil
			},
		).WithHeader(
			&http.Header{
				utility.UserEmailHeaderKey:   []string{"proctor@example.com"},
				utility.AccessTokenHeaderKey: []string{"access-token"},
			},
		),
	)

	executeProcResponse, err := proctorClient.ExecuteProc(procName, procArgs)

	assert.Equal(t, expectedProcResponse, executeProcResponse)
	assert.Error(t, errors.New(http.StatusText(http.StatusUnauthorized)), err)
}

func TestLogStreamForAuthorizedUser(t *testing.T) {
	logStreamAuthorizer := func(t *testing.T) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			assert.Equal(t, "proctor@example.com", r.Header.Get(utility.UserEmailHeaderKey))
			assert.Equal(t, "access-token", r.Header.Get(utility.AccessTokenHeaderKey))
			conn, _ := upgrader.Upgrade(w, r, nil)
			defer conn.Close()
		}
	}
	testServer := httptest.NewServer(logStreamAuthorizer(t))
	defer testServer.Close()
	proctorConfig := config.ProctorConfig{Host: makeHostname(testServer.URL), Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)

	err := proctorClient.StreamProcLogs("test-job-id")
	assert.NoError(t, err)
}

func TestLogStreamForBadWebSocketHandshake(t *testing.T) {
	badWebSocketHandshakeHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {}
	}
	testServer := httptest.NewServer(badWebSocketHandshakeHandler())
	defer testServer.Close()
	proctorConfig := config.ProctorConfig{Host: makeHostname(testServer.URL), Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)

	errStreamLogs := proctorClient.StreamProcLogs("test-job-id")
	assert.Equal(t, errors.New("websocket: bad handshake"), errStreamLogs)
}

func TestLogStreamForUnauthorizedUser(t *testing.T) {
	unauthorizedUserHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
	testServer := httptest.NewServer(unauthorizedUserHandler())
	defer testServer.Close()
	proctorConfig := config.ProctorConfig{Host: makeHostname(testServer.URL), Email: "proctor@example.com", AccessToken: "access-token"}
	proctorClient := NewClient(proctorConfig)

	errStreamLogs := proctorClient.StreamProcLogs("test-job-id")
	assert.Error(t, errors.New(http.StatusText(http.StatusUnauthorized)), errStreamLogs)
}

func makeHostname(s string) string {
	return strings.TrimPrefix(s, "http://")
}
