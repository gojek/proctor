package daemon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/gojektech/proctor/config"
	"github.com/gojektech/proctor/proc"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/gorilla/websocket"
)

type Client interface {
	ListProcs() ([]proc.Metadata, error)
	ExecuteProc(string, map[string]string) (string, error)
	StreamProcLogs(string) error
}

type client struct {
	proctordHost string
	emailId      string
	accessToken  string
}

type ProcToExecute struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}

func NewClient() Client {
	return &client{
		proctordHost: config.ProctorHost(),
		emailId:      config.EmailId(),
		accessToken:  config.AccessToken(),
	}
}

func (c *client) ListProcs() ([]proc.Metadata, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://"+c.proctordHost+"/jobs/metadata", nil)
	req.Header.Add(utility.UserEmailHeaderKey, c.emailId)
	req.Header.Add(utility.AccessTokenHeaderKey, c.accessToken)
	resp, err := client.Do(req)

	if err != nil {
		return []proc.Metadata{}, errors.New(err.Error())
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []proc.Metadata{}, errors.New(http.StatusText(resp.StatusCode))
	}

	var procList []proc.Metadata
	err = json.NewDecoder(resp.Body).Decode(&procList)
	return procList, err
}

func (c *client) ExecuteProc(name string, args map[string]string) (string, error) {
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
	resp, err := client.Do(req)

	if err != nil {
		return "", errors.New(err.Error())
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", errors.New(http.StatusText(resp.StatusCode))
	}

	var executedProc ProcToExecute
	err = json.NewDecoder(resp.Body).Decode(&executedProc)

	return executedProc.Name, err
}

func (c *client) StreamProcLogs(name string) error {
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
	headers[utility.AccessTokenHeaderKey] = token
	headers[utility.UserEmailHeaderKey] = emailId

	wsConn, response, err := websocket.DefaultDialer.Dial(proctodWebsocketURLWithProcName, headers)
	if err != nil {
		animation.Stop()
		if response.StatusCode == http.StatusUnauthorized {
			return errors.New(http.StatusText(http.StatusUnauthorized))
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
