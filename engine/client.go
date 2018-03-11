package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/gojekfarm/proctor/config"
	"github.com/gojekfarm/proctor/jobs"
	"github.com/gorilla/websocket"
)

type Client interface {
	ListJobs() ([]jobs.Metadata, error)
	ExecuteJob(string, map[string]string) (string, error)
	StreamJobLogs(string) error
}

type client struct {
	proctorEngineURL string
}

type job struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}

func NewClient() Client {
	return &client{
		proctorEngineURL: config.ProctorURL(),
	}
}

func (c *client) ListJobs() ([]jobs.Metadata, error) {
	resp, err := http.Get("http://" + c.proctorEngineURL + "/jobs/metadata")
	if err != nil || resp.StatusCode != http.StatusOK {
		return []jobs.Metadata{}, err
	}
	defer resp.Body.Close()

	var jobList []jobs.Metadata
	err = json.NewDecoder(resp.Body).Decode(&jobList)
	return jobList, err
}

func (c *client) ExecuteJob(name string, args map[string]string) (string, error) {
	jobToExecute := job{
		Name: name,
		Args: args,
	}

	requestBody, err := json.Marshal(jobToExecute)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://"+c.proctorEngineURL+"/jobs/execute", "application/json", bytes.NewReader(requestBody))
	if err != nil || resp.StatusCode != http.StatusCreated {
		return "", err
	}

	defer resp.Body.Close()
	var executedJob job
	err = json.NewDecoder(resp.Body).Decode(&executedJob)

	return executedJob.Name, err
}

func (c *client) StreamJobLogs(name string) error {
	animation := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	animation.Color("green")
	animation.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	proctorEngineWebsocketURL := url.URL{Scheme: "ws", Host: c.proctorEngineURL, Path: "/jobs/logs"}
	proctorEngineWebsocketURLWithJobName := proctorEngineWebsocketURL.String() + "?" + "job_name=" + name

	wsConn, _, err := websocket.DefaultDialer.Dial(proctorEngineWebsocketURLWithJobName, nil)
	if err != nil {
		animation.Stop()
		return err
	}
	defer wsConn.Close()

	logStreaming := make(chan int)
	go func() {
		for {
			_, message, err := wsConn.ReadMessage()
			animation.Stop()
			if err != nil {
				logStreaming <- 0
				return
			}
			fmt.Println(string(message))
		}
	}()

	for {
		select {
		case <-interrupt:
			color.New(color.FgRed).Println("User interrupt while streaming job logs")
			err := wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return err
		case <-logStreaming:
			return nil
		}
	}
}
