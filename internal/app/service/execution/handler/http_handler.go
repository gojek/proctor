package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"proctor/internal/app/service/execution/handler/parameter"
	"proctor/internal/app/service/execution/handler/status"
	"proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/execution/service"
	executionStatus "proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"strconv"
	"strings"
	"time"
)

type ExecutionHttpHandler interface {
	Post() http.HandlerFunc
	Status() http.HandlerFunc
	Logs() http.HandlerFunc
}

type executionHttpHandler struct {
	service    service.ExecutionService
	repository repository.ExecutionContextRepository
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  config.LogsStreamReadBufferSize(),
	WriteBufferSize: config.LogsStreamWriteBufferSize(),
}

func NewExecutionHttpHandler(
	executionService service.ExecutionService,
	repository repository.ExecutionContextRepository,
) ExecutionHttpHandler {
	return &executionHttpHandler{
		service:    executionService,
		repository: repository,
	}
}

func (httpHandler *executionHttpHandler) closeWebSocket(message string, conn *websocket.Conn) {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, message))
	logger.LogErrors(err, "close WebSocket")
	return
}

func (httpHandler *executionHttpHandler) Logs() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(response, request, nil)
		logger.LogErrors(err, "upgrade http server connection to WebSocket")
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.WebSocketInitError))
			return
		}
		defer conn.Close()

		contextIdParam := strings.TrimLeft(request.URL.RawQuery, "job_name=")
		executionContextId, err := strconv.ParseUint(contextIdParam, 10, 64)
		if contextIdParam == "" || err != nil {
			logger.Error("No valid execution context id provided as part of URL ", request.URL.RawQuery)
			httpHandler.closeWebSocket("No valid execution context id provided when fetching logs", conn)
			return
		}

		context, err := httpHandler.repository.GetById(executionContextId)

		if err != nil {
			logger.Error("No execution context id found ", executionContextId)
			httpHandler.closeWebSocket(fmt.Sprintf("No execution context found with id %v", executionContextId), conn)
			return
		}

		if context.Status == executionStatus.Finished {
			err = conn.WriteMessage(websocket.TextMessage, context.Output)
			logger.LogErrors(err, "write output to socket:", string(context.Output))
			httpHandler.closeWebSocket("Finished Streaming log", conn)
			return
		}

		if context.Status == executionStatus.PodReady {
			waitTime := config.KubeLogProcessWaitTime() * time.Second
			podLog, _err := httpHandler.service.StreamJobLogs(context.Name, waitTime)

			logger.LogErrors(_err, "stream job log by execution name", context.Name)
			if _err != nil {
				logger.Error("failed to stream job log by execution name ", context.Name)
				httpHandler.closeWebSocket(fmt.Sprintf("failed to stream job log by execution name %s:", context.Name), conn)
				return
			}

			defer podLog.Close()

			scanner := bufio.NewScanner(podLog)
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {
				_ = conn.WriteMessage(websocket.TextMessage, scanner.Bytes())
			}
			httpHandler.closeWebSocket("Finished Streaming log", conn)
			return
		}

		httpHandler.closeWebSocket("No Logs Found", conn)
		return
	}
}

func (httpHandler *executionHttpHandler) Status() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		contextId := mux.Vars(request)["name"]
		executionContextId, err := strconv.ParseUint(contextId, 10, 64)
		logger.LogErrors(err, "parse execution context id from path parameter:", contextId)
		if err != nil {
			response.WriteHeader(http.StatusNotFound)
			_, _ = response.Write([]byte(status.PathParameterError))
			return
		}

		context, err := httpHandler.repository.GetById(executionContextId)
		logger.LogErrors(err, "fetch execution context by id:", contextId)

		if err != nil {
			response.WriteHeader(http.StatusNotFound)
			_, _ = response.Write([]byte(status.ExecutionContextNotFound))
			return
		}

		responseMap := map[string]string{
			"ExecutionId": string(context.ExecutionID),
			"JobName":     context.JobName,
			"ImageTag":    context.ImageTag,
			"CreatedAt":   context.CreatedAt.String(),
			"Status":      string(context.Status),
		}

		response.WriteHeader(http.StatusOK)

		responseJson, err := json.Marshal(responseMap)
		logger.LogErrors(err, "marshal json from: ", responseMap)

		_, _ = response.Write(responseJson)

	}
}

func (httpHandler *executionHttpHandler) Post() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		userEmail := request.Header.Get(parameter.UserEmailHeader)

		var job parameter.Job

		err := json.NewDecoder(request.Body).Decode(&job)
		defer request.Body.Close()

		logger.LogErrors(err, "parsing request body to Job object")

		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.MalformedRequest))
			return
		}

		context, executionName, err := httpHandler.service.Execute(job.Name, userEmail, job.Args)

		logger.LogErrors(err, "execute job: ", job)

		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(fmt.Sprintf("%s , Errors Detail %s", status.JobExecutionError, err.Error())))
			return
		}

		responseMap := map[string]string{
			"ExecutionId":   string(context.ExecutionID),
			"JobName":       context.JobName,
			"ExecutionName": executionName,
			"ImageTag":      context.ImageTag,
			"CreatedAt":     context.CreatedAt.String(),
			"Status":        string(context.Status),
		}

		response.WriteHeader(http.StatusCreated)

		responseJson, err := json.Marshal(responseMap)
		logger.LogErrors(err, "marshal json from: ", responseMap)

		_, _ = response.Write(responseJson)
		return
	}
}
