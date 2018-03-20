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
	"github.com/gojektech/proctor/proctord/utility"
)

type executioner struct {
	kubeClient    kubernetes.Client
	metadataStore metadata.Store
	secretsStore  secrets.Store
	auditor       audit.Auditor
}

type Executioner interface {
	Handle() http.HandlerFunc
}

func NewExecutioner(kubeClient kubernetes.Client, metadataStore metadata.Store, secretsStore secrets.Store, auditor audit.Auditor) Executioner {
	return &executioner{
		kubeClient:    kubeClient,
		metadataStore: metadataStore,
		secretsStore:  secretsStore,
		auditor:       auditor,
	}
}

func (executioner *executioner) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		var job Job
		err := json.NewDecoder(req.Body).Decode(&job)
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

		jobSubmittedForExecution, err := executioner.kubeClient.ExecuteJob(imageName, envVars)
		if err != nil {
			logger.Error("Error executing job: %v", job, imageName, err.Error())
			ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(utility.ServerError))

			executioner.auditor.AuditJobsExecution(ctx)
			return
		}
		ctx = context.WithValue(ctx, utility.JobSubmittedForExecutionContextKey, jobSubmittedForExecution)
		ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionSuccess)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("{ \"name\":\"%s\" }", jobSubmittedForExecution)))

		executioner.auditor.AuditJobsExecution(ctx)
		return
	}
}
