package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gojektech/proctor/proctord/audit"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage"
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
		JobNameSubmittedForExecution := mux.Vars(req)["name"]
		jobExecutionStatus, err := handler.store.GetJobExecutionStatus(JobNameSubmittedForExecution)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			logger.Error("Error getting job status", err.Error())
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
		ctx := req.Context()
		//TODO: maybe move auditing inside execution itself?
		defer handler.auditor.AuditJobsExecution(ctx)

		var job Job
		err := json.NewDecoder(req.Body).Decode(&job)
		userEmail := req.Header.Get(utility.UserEmailHeaderKey)
		defer req.Body.Close()
		if err != nil {
			logger.Error("Error parsing request body", err.Error())
			ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionClientError)

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(utility.ClientError))
			return
		}

		jobExecutionID, err := handler.executioner.Execute(ctx, job.Name, userEmail, job.Args)
		if err != nil {
			logger.Error("Error executing job: ", err.Error())
			ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))
			return
		}

		go handler.auditor.AuditJobExecutionStatus(jobExecutionID)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("{ \"name\":\"%s\" }", jobExecutionID)))
		return
	}
}
