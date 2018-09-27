package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gojektech/proctor/proctord/audit"
	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/jobs/secrets"
	"github.com/gojektech/proctor/proctord/kubernetes"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/utility"

	"github.com/gorilla/mux"
)

type executioner struct {
	kubeClient    kubernetes.Client
	metadataStore metadata.Store
	secretsStore  secrets.Store
	auditor       audit.Auditor
	store         storage.Store
}

type Executioner interface {
	Handle() http.HandlerFunc
	Status() http.HandlerFunc
}

func NewExecutioner(kubeClient kubernetes.Client, metadataStore metadata.Store, secretsStore secrets.Store, auditor audit.Auditor, store storage.Store) Executioner {
	return &executioner{
		kubeClient:    kubeClient,
		metadataStore: metadataStore,
		secretsStore:  secretsStore,
		auditor:       auditor,
		store:         store,
	}
}

func (executioner *executioner) Status() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		JobNameSubmittedForExecution := mux.Vars(req)["name"]
		jobExecutionStatus, err := executioner.store.GetJobExecutionStatus(JobNameSubmittedForExecution)
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
func (executioner *executioner) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		var job Job
		err := json.NewDecoder(req.Body).Decode(&job)
		userEmail := req.Header.Get(utility.UserEmailHeaderKey)
		defer req.Body.Close()
		if err != nil {
			logger.Error("Error parsing request body", err.Error())
			ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionClientError)

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(utility.ClientError))

			executioner.auditor.AuditJobsExecution(ctx)
			return
		}
		ctx = context.WithValue(ctx, utility.JobNameContextKey, job.Name)
		ctx = context.WithValue(ctx, utility.UserEmailContextKey, userEmail)
		ctx = context.WithValue(ctx, utility.JobArgsContextKey, job.Args)

		jobMetadata, err := executioner.metadataStore.GetJobMetadata(job.Name)
		if err != nil {
			logger.Error("Error finding job to image", job.Name, err.Error())
			ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))

			executioner.auditor.AuditJobsExecution(ctx)
			return
		}
		imageName := jobMetadata.ImageName
		ctx = context.WithValue(ctx, utility.ImageNameContextKey, imageName)

		jobSecrets, err := executioner.secretsStore.GetJobSecrets(job.Name)
		if err != nil {
			logger.Error("Error retrieving secrets for job", job.Name, err.Error())
			ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(utility.ServerError))

			executioner.auditor.AuditJobsExecution(ctx)
			return
		}

		envVars := utility.MergeMaps(job.Args, jobSecrets)

		JobNameSubmittedForExecution, err := executioner.kubeClient.ExecuteJob(imageName, envVars)
		if err != nil {
			logger.Error("Error executing job: %v", job, imageName, err.Error())
			ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))

			executioner.auditor.AuditJobsExecution(ctx)
			return
		}
		ctx = context.WithValue(ctx, utility.JobNameSubmittedForExecutionContextKey, JobNameSubmittedForExecution)
		ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionSuccess)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("{ \"name\":\"%s\" }", JobNameSubmittedForExecution)))

		executioner.auditor.AuditJobsExecution(ctx)
		return
	}
}
