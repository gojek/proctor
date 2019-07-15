package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	"net/http"
	"proctor/internal/app/proctord/audit"
	"proctor/internal/app/proctord/storage"
	"proctor/internal/app/proctord/storage/postgres"
	"proctor/internal/app/service/execution/handler/parameter"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/pkg/constant"
	"time"
)

type executionHandler struct {
	auditor     audit.Auditor
	store       storage.Store
	executioner Executioner
}

type ExecutionHandler interface {
	Handle() http.HandlerFunc
	Status() http.HandlerFunc
	sendStatusToCaller(remoteCallerURL, jobExecutionID string)
}

func NewExecutionHandler(auditor audit.Auditor, store storage.Store, executioner Executioner) ExecutionHandler {
	return &executionHandler{
		auditor:     auditor,
		store:       store,
		executioner: executioner,
	}
}

func (handler *executionHandler) Status() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		jobExecutionID := mux.Vars(req)["name"]
		jobExecutionStatus, err := handler.store.GetJobExecutionStatus(jobExecutionID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))
			logger.Error(fmt.Sprintf("Error getting job status for job_id: %s", jobExecutionID), err.Error())
			raven.CaptureError(err, map[string]string{"job_id": jobExecutionID})

			return
		}

		if jobExecutionStatus == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, _ = fmt.Fprintf(w, jobExecutionStatus)
	}
}

func (handler *executionHandler) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		jobsExecutionAuditLog := &postgres.JobsExecutionAuditLog{
			JobExecutionStatus: "WAITING",
		}

		userEmail := req.Header.Get(constant.UserEmailHeaderKey)
		jobsExecutionAuditLog.UserEmail = userEmail

		var job parameter.Job
		err := json.NewDecoder(req.Body).Decode(&job)
		defer req.Body.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("User: %s: Error parsing request body", userEmail), err.Error())
			raven.CaptureError(err, map[string]string{"user_email": userEmail})

			jobsExecutionAuditLog.Errors = fmt.Sprintf("Error parsing request body: %s", err.Error())
			jobsExecutionAuditLog.JobSubmissionStatus = constant.JobSubmissionClientError

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(constant.ClientError))
			go handler.auditor.JobsExecutionAndStatus(jobsExecutionAuditLog)

			return
		}
		jobExecutionID, err := handler.executioner.Execute(jobsExecutionAuditLog, job.Name, job.Args)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: User %s: Error executing job: ", job.Name, userEmail), err.Error())
			raven.CaptureError(err, map[string]string{"user_email": userEmail, "job_name": job.Name})

			jobsExecutionAuditLog.Errors = fmt.Sprintf("Error executing job: %s", err.Error())
			jobsExecutionAuditLog.JobSubmissionStatus = constant.JobSubmissionServerError
			go handler.auditor.JobsExecutionAndStatus(jobsExecutionAuditLog)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(constant.ServerError))

			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(fmt.Sprintf("{ \"name\":\"%s\" }", jobExecutionID)))

		remoteCallerURL := job.CallbackURL
		go handler.postJobExecute(jobsExecutionAuditLog, remoteCallerURL, jobExecutionID)
		return
	}
}

func (handler *executionHandler) postJobExecute(jobsExecutionAuditLog *postgres.JobsExecutionAuditLog, remoteCallerURL, jobExecutionID string) {
	handler.auditor.JobsExecutionAndStatus(jobsExecutionAuditLog)
	if remoteCallerURL != "" {
		handler.sendStatusToCaller(remoteCallerURL, jobExecutionID)
	}
}

func (handler *executionHandler) sendStatusToCaller(remoteCallerURL, jobExecutionID string) {
	status := constant.JobWaiting

	for {
		jobExecutionStatus, _ := handler.store.GetJobExecutionStatus(jobExecutionID)
		if jobExecutionStatus == "" {
			status = constant.JobNotFound
			break
		}

		if jobExecutionStatus == constant.JobSucceeded || jobExecutionStatus == constant.JobFailed {
			status = jobExecutionStatus
			break
		}

		time.Sleep(1 * time.Second)
	}

	value := map[string]string{"name": jobExecutionID, "status": status}
	jsonValue, err := json.Marshal(value)
	if err != nil {
		logger.Error(fmt.Sprintf("StatusCallback: Error parsing %#v", value), err.Error())
		raven.CaptureError(err, nil)

		return
	}

	_, err = http.Post(remoteCallerURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		logger.Error("StatusCallback: Error sending request to callback url", err.Error())
		raven.CaptureError(err, nil)

		return
	}

	return
}
