package execution

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gojektech/proctor/proctord/audit"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/gojektech/proctor/proctord/utility"

	"github.com/gorilla/mux"
)

type executionHandler struct {
	auditor     audit.Auditor
	store       storage.Store
	executioner Executioner
}

type ExecutionHandler interface {
	Handle() http.HandlerFunc
	Status() http.HandlerFunc
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
		defer func() { go handler.auditor.JobsExecutionAndStatus(jobsExecutionAuditLog) }()

		userEmail := req.Header.Get(utility.UserEmailHeaderKey)
		jobsExecutionAuditLog.UserEmail = userEmail

		var job Job
		err := json.NewDecoder(req.Body).Decode(&job)
		defer req.Body.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("User: %s: Error parsing request body", userEmail), err.Error())
			jobsExecutionAuditLog.Errors = fmt.Sprintf("Error parsing request body: %s", err.Error())
			jobsExecutionAuditLog.JobSubmissionStatus = utility.JobSubmissionClientError

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(utility.ClientError))
			return
		}

		jobExecutionID, err := handler.executioner.Execute(jobsExecutionAuditLog, job.Name, job.Args)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: User %s: Error executing job: ", job.Name, userEmail), err.Error())
			jobsExecutionAuditLog.Errors = fmt.Sprintf("Error executing job: %s", err.Error())
			jobsExecutionAuditLog.JobSubmissionStatus = utility.JobSubmissionServerError

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("{ \"name\":\"%s\" }", jobExecutionID)))
		return
	}
}
