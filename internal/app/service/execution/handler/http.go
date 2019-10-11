package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"proctor/internal/app/service/execution/handler/parameter"
	"proctor/internal/app/service/execution/handler/status"
	"proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/execution/service"
	executionStatus "proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	serviceNotification "proctor/internal/app/service/notification/service"
	"proctor/internal/pkg/constant"
	"proctor/internal/pkg/model/execution"
	"proctor/pkg/notification/event"
)

type ExecutionHTTPHandler interface {
	Post() http.HandlerFunc
	GetStatus() http.HandlerFunc
	GetLogs() http.HandlerFunc
}

type executionHTTPHandler struct {
	service             service.ExecutionService
	repository          repository.ExecutionContextRepository
	notificationService serviceNotification.NotificationService
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  config.Config().LogsStreamReadBufferSize,
	WriteBufferSize: config.Config().LogsStreamWriteBufferSize,
}

func NewExecutionHTTPHandler(
	executionService service.ExecutionService,
	repository repository.ExecutionContextRepository,
	notificationService serviceNotification.NotificationService,
) ExecutionHTTPHandler {
	return &executionHTTPHandler{
		service:             executionService,
		repository:          repository,
		notificationService: notificationService,
	}
}

func (httpHandler *executionHTTPHandler) closeWebSocket(message string, conn *websocket.Conn) {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, message))
	logger.LogErrors(err, "close WebSocket ", message)
	return
}

func (httpHandler *executionHTTPHandler) GetLogs() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(response, request, nil)
		logger.LogErrors(err, "upgrade http server connection to WebSocket")
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(status.WebSocketInitError))
			return
		}
		defer conn.Close()

		contextIdParam := strings.TrimLeft(request.URL.RawQuery, "context_id=")
		executionContextId, err := strconv.ParseUint(contextIdParam, 10, 64)
		if contextIdParam == "" || err != nil {
			logger.Error("No valid execution context id provided as part of URL ", request.URL.RawQuery)
			httpHandler.closeWebSocket("No valid execution context id provided when fetching logs", conn)
			return
		}

		context, err := httpHandler.repository.GetById(executionContextId)
		logger.LogErrors(err, "fetch context from repository ", *context)
		if err != nil {
			logger.Error("No execution context id found ", executionContextId)
			httpHandler.closeWebSocket(fmt.Sprintf("No execution context found with id %v", executionContextId), conn)
			return
		}

		if context.Status == executionStatus.Finished {
			logger.Debug("Execution is Finished, return output from repository")
			err = conn.WriteMessage(websocket.TextMessage, context.Output)
			logger.LogErrors(err, "write output to socket:", string(context.Output))
			httpHandler.closeWebSocket("Finished Streaming log from repository", conn)
			return
		}

		if context.Status == executionStatus.Created {
			logger.Debug("Execution is Created, Stream output from pod")
			waitTime := config.Config().KubeLogProcessWaitTime * time.Second
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
			httpHandler.closeWebSocket("Finished Streaming log from pod", conn)
			return
		}

		httpHandler.closeWebSocket("No Logs Found", conn)
		return
	}
}

func (httpHandler *executionHTTPHandler) GetStatus() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		contextId := mux.Vars(request)["contextId"]
		executionContextId, err := strconv.ParseUint(contextId, 10, 64)
		logger.LogErrors(err, "parse execution context id from path parameter:", contextId)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
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

		responseBody := &execution.ExecutionResult{
			ExecutionId:   context.ExecutionID,
			JobName:       context.JobName,
			ExecutionName: context.Name,
			ImageTag:      context.ImageTag,
			CreatedAt:     context.CreatedAt.String(),
			UpdatedAt:     context.UpdatedAt.String(),
			Status:        string(context.Status),
		}

		response.WriteHeader(http.StatusOK)

		responseJSON, err := json.Marshal(responseBody)
		logger.LogErrors(err, "marshal json from: ", responseBody)

		_, _ = response.Write(responseJSON)

	}
}

func (httpHandler *executionHTTPHandler) Post() http.HandlerFunc {
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
		job.Args[constant.AuthorEmailKey] = userEmail

		context, executionName, err := httpHandler.service.Execute(job.Name, userEmail, job.Args)

		logger.LogErrors(err, "execute job: ", job)

		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(fmt.Sprintf("%s, Errors Detail %s", status.JobExecutionError, err.Error())))
			return
		}

		evt := event.NewExecutionEvent(userEmail, *context)
		httpHandler.notificationService.Notify(evt)

		responseBody := &execution.ExecutionResult{
			ExecutionId:   context.ExecutionID,
			JobName:       context.JobName,
			ExecutionName: executionName,
			ImageTag:      context.ImageTag,
			CreatedAt:     context.CreatedAt.String(),
			UpdatedAt:     context.UpdatedAt.String(),
			Status:        string(context.Status),
		}

		response.WriteHeader(http.StatusCreated)

		responseJSON, err := json.Marshal(responseBody)
		logger.LogErrors(err, "marshal json from: ", responseBody)

		_, _ = response.Write(responseJSON)
		return
	}
}
