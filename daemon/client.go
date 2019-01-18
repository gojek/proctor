package daemon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gojektech/proctor/cmd/version"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/gojektech/proctor/config"
	"github.com/gojektech/proctor/io"
	proc_metadata "github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/jobs/schedule"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/gorilla/websocket"
)

type Client interface {
	ListProcs() ([]proc_metadata.Metadata, error)
	ExecuteProc(string, map[string]string) (string, error)
	StreamProcLogs(string) error
	GetDefinitiveProcExecutionStatus(string) (string, error)
	ScheduleJob(string, string, string, string, map[string]string) (string, error)
	ListScheduledProcs() ([]schedule.ScheduledJob, error)
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

type ScheduleJobPayload struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Tags               string            `json:"tags"`
	Time               string            `json:"time"`
	NotificationEmails string            `json:"notification_emails"`
	Args               map[string]string `json:"args"`
}

func NewClient(printer io.Printer, proctorConfigLoader config.Loader) Client {
	return &client{
		clientVersion:       version.ClientVersion,
		printer:             printer,
		proctorConfigLoader: proctorConfigLoader,
	}
}

func (c *client) ScheduleJob(name, tags, time, notificationEmails string, jobArgs map[string]string) (string, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return "", err
	}
	jobPayload := ScheduleJobPayload{
		Name:               name,
		Tags:               tags,
		Time:               time,
		NotificationEmails: notificationEmails,
		Args:               jobArgs,
	}

	requestBody, err := json.Marshal(jobPayload)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://"+c.proctordHost+"/jobs/schedule", bytes.NewReader(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(utility.UserEmailHeaderKey, c.emailId)
	req.Header.Add(utility.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(utility.ClientVersionHeaderKey, c.clientVersion)
	resp, err := client.Do(req)

	if err != nil {
		return "", buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(body)
		return "", errors.New(bodyString)
	}

	var scheduledJob ScheduleJobPayload
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

func (c *client) ListProcs() ([]proc_metadata.Metadata, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return []proc_metadata.Metadata{}, err
	}

	client := &http.Client{
		Timeout: c.connectionTimeoutSecs,
	}
	req, err := http.NewRequest("GET", "http://"+c.proctordHost+"/jobs/metadata", nil)
	req.Header.Add(utility.UserEmailHeaderKey, c.emailId)
	req.Header.Add(utility.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(utility.ClientVersionHeaderKey, c.clientVersion)

	resp, err := client.Do(req)
	if err != nil {
		return []proc_metadata.Metadata{}, buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []proc_metadata.Metadata{}, buildHTTPError(c, resp)
	}

	var procList []proc_metadata.Metadata
	err = json.NewDecoder(resp.Body).Decode(&procList)
	return procList, err
}

func (c *client) ListScheduledProcs() ([]schedule.ScheduledJob, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return []schedule.ScheduledJob{}, err
	}

	client := &http.Client{
		Timeout: c.connectionTimeoutSecs,
	}
	req, err := http.NewRequest("GET", "http://"+c.proctordHost+"/jobs/schedule", nil)
	req.Header.Add(utility.UserEmailHeaderKey, c.emailId)
	req.Header.Add(utility.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(utility.ClientVersionHeaderKey, c.clientVersion)

	resp, err := client.Do(req)
	if err != nil {
		return []schedule.ScheduledJob{}, buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []schedule.ScheduledJob{}, buildHTTPError(c, resp)
	}

	var scheduledProcsList []schedule.ScheduledJob
	err = json.NewDecoder(resp.Body).Decode(&scheduledProcsList)
	return scheduledProcsList, err
}

func (c *client) ExecuteProc(name string, args map[string]string) (string, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return "", err
	}

	procToExecute := ProcToExecute{
		Name: name,
		Args: args,
	}

	requestBody, err := json.Marshal(procToExecute)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://"+c.proctordHost+"/jobs/execute", bytes.NewReader(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(utility.UserEmailHeaderKey, c.emailId)
	req.Header.Add(utility.AccessTokenHeaderKey, c.accessToken)
	req.Header.Add(utility.ClientVersionHeaderKey, c.clientVersion)
	resp, err := client.Do(req)

	if err != nil {
		return "", buildNetworkError(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", buildHTTPError(c, resp)
	}

	var executedProc ProcToExecute
	err = json.NewDecoder(resp.Body).Decode(&executedProc)

	return executedProc.Name, err
}

func (c *client) StreamProcLogs(name string) error {
	err := c.loadProctorConfig()
	if err != nil {
		return err
	}

	animation := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	animation.Color("green")
	animation.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	proctodWebsocketURL := url.URL{Scheme: "ws", Host: c.proctordHost, Path: "/jobs/logs"}
	proctodWebsocketURLWithProcName := proctodWebsocketURL.String() + "?" + "job_name=" + name

	headers := make(map[string][]string)
	token := []string{c.accessToken}
	emailId := []string{c.emailId}
	clientVersion := []string{c.clientVersion}
	headers[utility.AccessTokenHeaderKey] = token
	headers[utility.UserEmailHeaderKey] = emailId
	headers[utility.ClientVersionHeaderKey] = clientVersion

	wsConn, response, err := websocket.DefaultDialer.Dial(proctodWebsocketURLWithProcName, headers)
	if err != nil {
		animation.Stop()
		if response.StatusCode == http.StatusUnauthorized {
			if c.emailId == "" || c.accessToken == "" {
				return fmt.Errorf("%s\n%s", utility.UnauthorizedErrorHeader, utility.UnauthorizedErrorMissingConfig)
			}
			return fmt.Errorf("%s\n%s", utility.UnauthorizedErrorHeader, utility.UnauthorizedErrorInvalidConfig)
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

func (c *client) GetDefinitiveProcExecutionStatus(procName string) (string, error) {
	err := c.loadProctorConfig()
	if err != nil {
		return "", err
	}

	for count := 0; count < c.procExecutionStatusPollCount; count += 1 {
		httpClient := &http.Client{
			Timeout: c.connectionTimeoutSecs,
		}

		req, err := http.NewRequest("GET", "http://"+c.proctordHost+"/jobs/execute/"+procName+"/status", nil)
		req.Header.Add(utility.UserEmailHeaderKey, c.emailId)
		req.Header.Add(utility.AccessTokenHeaderKey, c.accessToken)
		req.Header.Add(utility.ClientVersionHeaderKey, c.clientVersion)

		resp, err := httpClient.Do(req)
		if err != nil {
			return "", buildNetworkError(err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", buildHTTPError(c, resp)
		}

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return "", err
		}

		procExecutionStatus := string(body)
		if procExecutionStatus == utility.JobSucceeded || procExecutionStatus == utility.JobFailed {
			return procExecutionStatus, nil
		}

		time.Sleep(time.Duration(count) * 100 * time.Millisecond)
	}
	return "", errors.New(fmt.Sprintf("No definitive status received for proc name %s from proctord", procName))
}

func buildNetworkError(err error) error {
	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return fmt.Errorf("%s\n%s\n%s", utility.GenericTimeoutErrorHeader, netError.Error(), utility.GenericTimeoutErrorBody)
	}
	return fmt.Errorf("%s\n%s", utility.GenericNetworkErrorHeader, err.Error())
}

func buildHTTPError(c *client, resp *http.Response) error {
	if resp.StatusCode == http.StatusUnauthorized {
		if c.emailId == "" || c.accessToken == "" {
			return fmt.Errorf("%s\n%s", utility.UnauthorizedErrorHeader, utility.UnauthorizedErrorMissingConfig)
		}
		return fmt.Errorf("%s\n%s", utility.UnauthorizedErrorHeader, utility.UnauthorizedErrorInvalidConfig)
	} else if resp.StatusCode == http.StatusBadRequest {
		body, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(body)
		return fmt.Errorf(bodyString)
	} else {
		return fmt.Errorf("%s\nStatus Code: %d, %s", utility.GenericResponseErrorHeader, resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}
