package execution

import (
	"errors"
	"fmt"

	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/jobs/secrets"
	"github.com/gojektech/proctor/proctord/kubernetes"
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/gojektech/proctor/proctord/utility"
)

type executioner struct {
	kubeClient    kubernetes.Client
	metadataStore metadata.Store
	secretsStore  secrets.Store
}

type Executioner interface {
	Execute(*postgres.JobsExecutionAuditLog, string, map[string]string) (string, error)
}

func NewExecutioner(kubeClient kubernetes.Client, metadataStore metadata.Store, secretsStore secrets.Store) Executioner {
	return &executioner{
		kubeClient:    kubeClient,
		metadataStore: metadataStore,
		secretsStore:  secretsStore,
	}
}

func (executioner *executioner) Execute(jobsExecutionAuditLog *postgres.JobsExecutionAuditLog, jobName string, jobArgs map[string]string) (string, error) {
	jobsExecutionAuditLog.JobName = jobName

	jobMetadata, err := executioner.metadataStore.GetJobMetadata(jobName)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error finding image for job: %s. Error: %s", jobName, err.Error()))
	}
	imageName := jobMetadata.ImageName
	jobsExecutionAuditLog.ImageName = imageName

	jobSecrets, err := executioner.secretsStore.GetJobSecrets(jobName)
	if err != nil && err.Error() != "redigo: nil returned" {
		return "", errors.New(fmt.Sprintf("Error retrieving secrets for job: %s. Error: %s", jobName, err.Error()))
	}

	envVars := utility.MergeMaps(jobArgs, jobSecrets)
	jobsExecutionAuditLog.AddJobArgs(envVars)

	jobExecutionID, err := executioner.kubeClient.ExecuteJob(imageName, envVars)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error submitting job to kube: %s. Error: %s", jobName, err.Error()))
	}
	jobsExecutionAuditLog.AddExecutionID(jobExecutionID)
	jobsExecutionAuditLog.JobSubmissionStatus = utility.JobSubmissionSuccess

	return jobExecutionID, nil
}
