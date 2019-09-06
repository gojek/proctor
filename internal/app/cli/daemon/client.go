package daemon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	ioReader "io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"

	"proctor/internal/app/cli/command/version"
	"proctor/internal/app/cli/config"
	"proctor/internal/app/cli/utility/io"
	"proctor/internal/pkg/constant"
	modelExecution "proctor/internal/pkg/model/execution"
	modelMetadata "proctor/internal/pkg/model/metadata"
	modelSchedule "proctor/internal/pkg/model/schedule"
)

const (
	ExecutionRoute     string = "/execution"
	ExecutionLogsRoute string = "/execution/logs"
	MetadataRoute      string = "/metadata"
	ScheduleRoute      string = "/schedule"
)

type Client interface {
	ListProcs() ([]modelMetadata.Metadata, error)
	ExecuteProc(string, map[string]string) (*modelExecution.ExecutionResult, error)
	StreamProcLogs(executionId uint64) error
	GetExecutionContextStatusWithPolling(executionId uint64) (*modelExecution.ExecutionResult, error)
	GetExecutionContextStatus(executionId uint64) (*modelExecution.ExecutionResult, error)
	ScheduleJob(string, string, string, string, string, map[string]string) (uint64, error)
	ListScheduledProcs() ([]modelSchedule.ScheduledJob, error)
	DescribeScheduledProc(uint64) (modelSchedule.ScheduledJob, error)
	RemoveScheduledProc(string) error
}

type client struct {
	printer                      io.Printer
	proctorConfigLoader          config.Loader
	proctordHost                 string
	emailId                      string
	accessToken                  string
	clientVersion                string
	connectionTimeoutSecs        time.Duration
	procExecutionStatusPollCount int
}

type ProcToExecute struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}

func NewClient(printer io.Printer, proctorConfigLoader config.Loader) Client {
	return &client{
		clientVersion:       version.ClientVersion,
		printer:             printer,
		proctorConfigLoader: proctorConfigLoader,
	}
}

func (c *client) ScheduleJob(name, tags, time, notificationEmails, group string, jobArgs map[string]string) (uint64, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return 0, err
	}
	jobPayload := modelSchedule.ScheduledJob{
		Name:               name,
		Tags:               tags,
		Time:               time,
		NotificationEmails: notificationEmails,
		Args:               jobArgs,
		Group:              group,
	}

	requestBody, err := json.Marshal(jobPayload)
	if err != nil {
		return 0, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://"+c.proctordHost+ScheduleRoute, bytes.NewReader(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(constant.UserEmailHeaderKey, c.emailId)
	req.Header.Add(constant.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(constant.ClientVersionHeaderKey, c.clientVersion)
	resp, err := client.Do(req)

	if err != nil {
		return 0, buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return 0, buildHTTPError(c, resp)
	}

	var scheduledJob modelSchedule.ScheduledJob
	err = json.NewDecoder(resp.Body).Decode(&scheduledJob)

	return scheduledJob.ID, err
}

func (c *client) loadProctorConfig() error {
	proctorConfig, err := c.proctorConfigLoader.Load()
	if err != (config.ConfigError{}) {
		c.printer.Println(err.RootError().Error(), color.FgRed)
		c.printer.Println(err.Message, color.FgGreen)
		return errors.New("Encountered error while loading config, exiting.")
	}

	c.proctordHost = proctorConfig.Host
	c.emailId = proctorConfig.Email
	c.accessToken = proctorConfig.AccessToken
	c.connectionTimeoutSecs = proctorConfig.ConnectionTimeoutSecs
	c.procExecutionStatusPollCount = proctorConfig.ProcExecutionStatusPollCount

	return nil
}

func (c *client) ListProcs() ([]modelMetadata.Metadata, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return []modelMetadata.Metadata{}, err
	}

	client := &http.Client{
		Timeout: c.connectionTimeoutSecs,
	}
	req, err := http.NewRequest("GET", "http://"+c.proctordHost+MetadataRoute, nil)
	req.Header.Add(constant.UserEmailHeaderKey, c.emailId)
	req.Header.Add(constant.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(constant.ClientVersionHeaderKey, c.clientVersion)

	resp, err := client.Do(req)
	if err != nil {
		return []modelMetadata.Metadata{}, buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []modelMetadata.Metadata{}, buildHTTPError(c, resp)
	}

	var procList []modelMetadata.Metadata
	err = json.NewDecoder(resp.Body).Decode(&procList)
	return procList, err
}

func (c *client) ListScheduledProcs() ([]modelSchedule.ScheduledJob, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return []modelSchedule.ScheduledJob{}, err
	}

	client := &http.Client{
		Timeout: c.connectionTimeoutSecs,
	}
	req, err := http.NewRequest("GET", "http://"+c.proctordHost+ScheduleRoute, nil)
	req.Header.Add(constant.UserEmailHeaderKey, c.emailId)
	req.Header.Add(constant.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(constant.ClientVersionHeaderKey, c.clientVersion)

	resp, err := client.Do(req)
	if err != nil {
		return []modelSchedule.ScheduledJob{}, buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []modelSchedule.ScheduledJob{}, buildHTTPError(c, resp)
	}

	var scheduledProcsList []modelSchedule.ScheduledJob
	err = json.NewDecoder(resp.Body).Decode(&scheduledProcsList)
	return scheduledProcsList, err
}

func (c *client) DescribeScheduledProc(jobID uint64) (modelSchedule.ScheduledJob, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return modelSchedule.ScheduledJob{}, err
	}

	client := &http.Client{
		Timeout: c.connectionTimeoutSecs,
	}
	url := fmt.Sprintf("http://"+c.proctordHost+ScheduleRoute+"/%d", jobID)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add(constant.UserEmailHeaderKey, c.emailId)
	req.Header.Add(constant.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(constant.ClientVersionHeaderKey, c.clientVersion)

	resp, err := client.Do(req)
	if err != nil {
		return modelSchedule.ScheduledJob{}, buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return modelSchedule.ScheduledJob{}, buildHTTPError(c, resp)
	}

	var scheduledProc modelSchedule.ScheduledJob
	err = json.NewDecoder(resp.Body).Decode(&scheduledProc)
	return scheduledProc, err
}

func (c *client) RemoveScheduledProc(jobID string) error {
	err := c.loadProctorConfig()
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: c.connectionTimeoutSecs,
	}
	url := fmt.Sprintf("http://"+c.proctordHost+ScheduleRoute+"/%s", jobID)
	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Add(constant.UserEmailHeaderKey, c.emailId)
	req.Header.Add(constant.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(constant.ClientVersionHeaderKey, c.clientVersion)

	resp, err := client.Do(req)
	if err != nil {
		return buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return buildHTTPError(c, resp)
	}

	return nil
}

func (c *client) ExecuteProc(name string, args map[string]string) (*modelExecution.ExecutionResult, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return nil, err
	}

	procToExecute := ProcToExecute{
		Name: name,
		Args: args,
	}

	requestBody, err := json.Marshal(procToExecute)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://"+c.proctordHost+ExecutionRoute, bytes.NewReader(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(constant.UserEmailHeaderKey, c.emailId)
	req.Header.Add(constant.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(constant.ClientVersionHeaderKey, c.clientVersion)
	resp, err := client.Do(req)
	if err != nil {
		return nil, buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, buildHTTPError(c, resp)
	}

	var executionResult modelExecution.ExecutionResult
	err = json.NewDecoder(resp.Body).Decode(&executionResult)

	return &executionResult, err
}

func (c *client) StreamProcLogs(executionId uint64) error {
	err := c.loadProctorConfig()
	if err != nil {
		return err
	}

	animation := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	animation.Color("green")
	animation.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	proctodWebsocketURL := url.URL{Scheme: "ws", Host: c.proctordHost, Path: ExecutionLogsRoute}
	proctodWebsocketURLWithProcName := fmt.Sprintf("%s?context_id=%v", proctodWebsocketURL.String(), executionId)

	headers := make(map[string][]string)
	token := []string{c.accessToken}
	emailId := []string{c.emailId}
	clientVersion := []string{c.clientVersion}
	headers[constant.AccessTokenHeaderKey] = token
	headers[constant.UserEmailHeaderKey] = emailId
	headers[constant.ClientVersionHeaderKey] = clientVersion

	wsConn, response, err := websocket.DefaultDialer.Dial(proctodWebsocketURLWithProcName, headers)
	if err != nil {
		animation.Stop()
		if response.StatusCode == http.StatusUnauthorized {
			if c.emailId == "" || c.accessToken == "" {
				return fmt.Errorf("%s\n%s", constant.UnauthorizedErrorHeader, constant.UnauthorizedErrorMissingConfig)
			}
			return fmt.Errorf("%s\n%s", constant.UnauthorizedErrorHeader, constant.UnauthorizedErrorInvalidConfig)
		}
		return err
	}
	defer wsConn.Close()

	logStreaming := make(chan int)
	go func() {
		for {
			_, message, err := wsConn.ReadMessage()
			animation.Stop()
			if err != nil {
				fmt.Println()
				logStreaming <- 0
				return
			}
			fmt.Println(string(message))
		}
	}()

	for {
		select {
		case <-interrupt:
			color.New(color.FgRed).Println("User interrupt while streaming proc logs")
			err := wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return err
		case <-logStreaming:
			return nil
		}
	}
}

func (c *client) GetExecutionContextStatusWithPolling(executionId uint64) (*modelExecution.ExecutionResult, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return nil, err
	}

	for count := 0; count < c.procExecutionStatusPollCount; count += 1 {
		executionContextStatus, err := c.GetExecutionContextStatus(executionId)
		if err != nil {
			return nil, err
		}
		if executionContextStatus.Status == constant.JobSucceeded || executionContextStatus.Status == constant.JobFailed {
			return executionContextStatus, nil
		}

		time.Sleep(time.Duration(count) * 100 * time.Millisecond)
	}
	return nil, errors.New(fmt.Sprintf("No definitive status received for execution with id %v from proctord", executionId))
}

func (c *client) GetExecutionContextStatus(executionId uint64) (*modelExecution.ExecutionResult, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: c.connectionTimeoutSecs,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s%s/%v/status", c.proctordHost, ExecutionRoute, executionId), nil)
	req.Header.Add(constant.UserEmailHeaderKey, c.emailId)
	req.Header.Add(constant.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(constant.ClientVersionHeaderKey, c.clientVersion)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, buildNetworkError(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, buildHTTPError(c, resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var executionResult modelExecution.ExecutionResult
	err = json.Unmarshal(body, &executionResult)
	if err != nil {
		return nil, err
	}

	return &executionResult, nil
}

func buildNetworkError(err error) error {
	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return fmt.Errorf("%s\n%s\n%s", constant.GenericTimeoutErrorHeader, netError.Error(), constant.GenericTimeoutErrorBody)
	}
	return fmt.Errorf("%s\n%s", constant.GenericNetworkErrorHeader, err.Error())
}

func buildHTTPError(c *client, resp *http.Response) error {
	if resp.StatusCode == http.StatusUnauthorized {
		if c.emailId == "" || c.accessToken == "" {
			return fmt.Errorf("%s\n%s", constant.UnauthorizedErrorHeader, constant.UnauthorizedErrorMissingConfig)
		}
		return fmt.Errorf("%s\n%s", constant.UnauthorizedErrorHeader, constant.UnauthorizedErrorInvalidConfig)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return getHTTPResponseError(resp.Body)
	}

	if resp.StatusCode == http.StatusNoContent {
		return fmt.Errorf(constant.NoScheduledJobsError)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf(constant.JobNotFoundError)
	}

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf(constant.JobForbiddenErrorHeader)
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return getHTTPResponseError(resp.Body)
	}

	return fmt.Errorf("%s\nStatus Code: %d, %s", constant.GenericResponseErrorHeader, resp.StatusCode, http.StatusText(resp.StatusCode))
}

func getHTTPResponseError(response ioReader.ReadCloser) error {
	body, _ := ioutil.ReadAll(response)
	bodyString := string(body)
	return fmt.Errorf(bodyString)
}
