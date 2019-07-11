package execution

import (
	"errors"
	"fmt"
	"proctor/internal/app/proctord/storage/postgres"
	"proctor/internal/app/service/infra/kubernetes"
	repository2 "proctor/internal/app/service/metadata/repository"
	"proctor/internal/app/service/secret/repository"

	"proctor/internal/pkg/constant"
	"proctor/internal/pkg/utility"
)

type executioner struct {
	kubeClient    kubernetes.KubernetesClient
	metadataStore repository2.MetadataRepository
	secretsStore  repository.SecretRepository
}

type Executioner interface {
	Execute(*postgres.JobsExecutionAuditLog, string, map[string]string) (string, error)
}

func NewExecutioner(kubeClient kubernetes.KubernetesClient, metadataStore repository2.MetadataRepository, secretsStore repository.SecretRepository) Executioner {
	return &executioner{
		kubeClient:    kubeClient,
		metadataStore: metadataStore,
		secretsStore:  secretsStore,
	}
}

func (executioner *executioner) Execute(jobsExecutionAuditLog *postgres.JobsExecutionAuditLog, jobName string, jobArgs map[string]string) (string, error) {
	jobsExecutionAuditLog.JobName = jobName

	jobMetadata, err := executioner.metadataStore.GetByName(jobName)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error finding image for job: %s. Error: %s", jobName, err.Error()))
	}
	imageName := jobMetadata.ImageName
	jobsExecutionAuditLog.ImageName = imageName

	jobSecrets, err := executioner.secretsStore.GetByJobName(jobName)
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
	jobsExecutionAuditLog.JobSubmissionStatus = constant.JobSubmissionSuccess

	return jobExecutionID, nil
}
