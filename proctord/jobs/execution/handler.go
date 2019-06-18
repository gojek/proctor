package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	"net/http"
	"proctor/proctord/audit"
	"proctor/proctord/logger"
	"proctor/proctord/storage"
	"proctor/proctord/storage/postgres"
	"proctor/proctord/utility"
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
			w.Write([]byte(utility.ServerError))
			logger.Error(fmt.Sprintf("Error getting job status for job_id: %s", jobExecutionID), err.Error())
			raven.CaptureError(err, map[string]string{"job_id": jobExecutionID})

			return
		}

		if jobExecutionStatus == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, jobExecutionStatus)
	}
}

func (handler *executionHandler) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		jobsExecutionAuditLog := &postgres.JobsExecutionAuditLog{
			JobExecutionStatus: "WAITING",
		}

		userEmail := req.Header.Get(utility.UserEmailHeaderKey)
		jobsExecutionAuditLog.UserEmail = userEmail

		var job Job
		err := json.NewDecoder(req.Body).Decode(&job)
		defer req.Body.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("User: %s: Error parsing request body", userEmail), err.Error())
			raven.CaptureError(err, map[string]string{"user_email": userEmail})

			jobsExecutionAuditLog.Errors = fmt.Sprintf("Error parsing request body: %s", err.Error())
			jobsExecutionAuditLog.JobSubmissionStatus = utility.JobSubmissionClientError

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(utility.ClientError))
			go handler.auditor.JobsExecutionAndStatus(jobsExecutionAuditLog)

			return
		}
		jobExecutionID, err := handler.executioner.Execute(jobsExecutionAuditLog, job.Name, job.Args)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: User %s: Error executing job: ", job.Name, userEmail), err.Error())
			raven.CaptureError(err, map[string]string{"user_email": userEmail, "job_name": job.Name})

			jobsExecutionAuditLog.Errors = fmt.Sprintf("Error executing job: %s", err.Error())
			jobsExecutionAuditLog.JobSubmissionStatus = utility.JobSubmissionServerError
			go handler.auditor.JobsExecutionAndStatus(jobsExecutionAuditLog)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))

			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("{ \"name\":\"%s\" }", jobExecutionID)))

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
	status := utility.JobWaiting

	for {
		jobExecutionStatus, _ := handler.store.GetJobExecutionStatus(jobExecutionID)
		if jobExecutionStatus == "" {
			status = utility.JobNotFound
			break
		}

		if jobExecutionStatus == utility.JobSucceeded || jobExecutionStatus == utility.JobFailed {
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
