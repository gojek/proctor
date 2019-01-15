package execution

import (
	"context"

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
}

type Executioner interface {
	Execute(context.Context, string, string, map[string]string) (string, error)
}

func NewExecutioner(kubeClient kubernetes.Client, metadataStore metadata.Store, secretsStore secrets.Store) Executioner {
	return &executioner{
		kubeClient:    kubeClient,
		metadataStore: metadataStore,
		secretsStore:  secretsStore,
	}
}

func (executioner *executioner) Execute(ctx context.Context, jobName, userEmail string, jobArgs map[string]string) (string, error) {
	ctx = context.WithValue(ctx, utility.JobNameContextKey, jobName)
	ctx = context.WithValue(ctx, utility.UserEmailContextKey, userEmail)
	ctx = context.WithValue(ctx, utility.JobArgsContextKey, jobArgs)

	jobMetadata, err := executioner.metadataStore.GetJobMetadata(jobName)
	if err != nil {
		logger.Error("Error finding job to image", jobName, err.Error())

		ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

		return "", err
	}

	imageName := jobMetadata.ImageName
	jobSecrets, err := executioner.secretsStore.GetJobSecrets(jobName)
	if err != nil {
		//TODO: add check for nil, which means no job secrets configured
		logger.Error("Error retrieving secrets for job", jobName, err.Error())
		ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

		return "", err
	}

	envVars := utility.MergeMaps(jobArgs, jobSecrets)

	jobNameSubmittedForExecution, err := executioner.kubeClient.ExecuteJob(imageName, envVars)
	if err != nil {
		logger.Error("Error executing job:", jobName, imageName, err.Error())
		ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

		return "", err
	}
	ctx = context.WithValue(ctx, utility.JobNameSubmittedForExecutionContextKey, jobNameSubmittedForExecution)
	ctx = context.WithValue(ctx, utility.JobSubmissionStatusContextKey, utility.JobSubmissionSuccess)

	return jobNameSubmittedForExecution, nil
}
