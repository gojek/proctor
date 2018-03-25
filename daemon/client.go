package daemon

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
	"github.com/gojektech/proctor/config"
	"github.com/gojektech/proctor/proc"
	"github.com/gorilla/websocket"
)

type Client interface {
	ListProcs() ([]proc.Metadata, error)
	ExecuteProc(string, map[string]string) (string, error)
	StreamProcLogs(string) error
}

type client struct {
	proctorEngineURL string
}

type ProcToExecute struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}

func NewClient() Client {
	return &client{
		proctorEngineURL: config.ProctorURL(),
	}
}

func (c *client) ListProcs() ([]proc.Metadata, error) {
	resp, err := http.Get("http://" + c.proctorEngineURL + "/jobs/metadata")
	if err != nil || resp.StatusCode != http.StatusOK {
		return []proc.Metadata{}, err
	}
	defer resp.Body.Close()

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

	resp, err := http.Post("http://"+c.proctorEngineURL+"/jobs/execute", "application/json", bytes.NewReader(requestBody))
	if err != nil || resp.StatusCode != http.StatusCreated {
		return "", err
	}

	defer resp.Body.Close()
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

	proctorEngineWebsocketURL := url.URL{Scheme: "ws", Host: c.proctorEngineURL, Path: "/jobs/logs"}
	proctorEngineWebsocketURLWithProcName := proctorEngineWebsocketURL.String() + "?" + "job_name=" + name

	wsConn, _, err := websocket.DefaultDialer.Dial(proctorEngineWebsocketURLWithProcName, nil)
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
			color.New(color.FgRed).Println("User interrupt while streaming proc logs")
			err := wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return err
		case <-logStreaming:
			return nil
		}
	}
}
