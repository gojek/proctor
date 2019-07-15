package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx/types"
	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/kubernetes"
	"proctor/internal/app/service/infra/logger"
	svcMetadataRepository "proctor/internal/app/service/metadata/repository"
	svcSecretRepository "proctor/internal/app/service/secret/repository"
	"time"
)

type ExecutionService interface {
	Execute(jobName string, userEmail string, args map[string]string) (*model.ExecutionContext, string, error)
	save(executionContext *model.ExecutionContext) error
}

type executionService struct {
	kubernetesClient   kubernetes.KubernetesClient
	repository         repository.ExecutionContextRepository
	metadataRepository svcMetadataRepository.MetadataRepository
	secretRepository   svcSecretRepository.SecretRepository
}

func NewExecutionService(
	kubernetesClient kubernetes.KubernetesClient,
	repository repository.ExecutionContextRepository,
	metadataRepository svcMetadataRepository.MetadataRepository,
	secretRepository svcSecretRepository.SecretRepository,
) ExecutionService {
	return &executionService{
		kubernetesClient:   kubernetesClient,
		repository:         repository,
		metadataRepository: metadataRepository,
		secretRepository:   secretRepository,
	}
}

func (service *executionService) save(executionContext *model.ExecutionContext) error {
	var err error
	if executionContext.ExecutionID == 0 {
		_, err = service.repository.Insert(executionContext)
		logger.LogErrors(err, "save execution context to db", *executionContext)
	} else {
		context, _err := service.repository.GetById(executionContext.ExecutionID)
		logger.LogErrors(_err, "get context from db by execution id", *executionContext)
		if _err != nil || context == nil {
			_, err = service.repository.Insert(executionContext)
			logger.LogErrors(err, "save execution context to db", *executionContext)
		} else {
			err = service.repository.UpdateStatus(executionContext.ExecutionID, executionContext.Status)
			logger.LogErrors(err, "update execution context status", *executionContext)
			if len(executionContext.Output) > 0 {
				err = service.repository.UpdateJobOutput(executionContext.ExecutionID, executionContext.Output)
				logger.LogErrors(err, "update execution context output", *executionContext)
			}
		}
	}
	return err
}

func (service *executionService) Execute(jobName string, userEmail string, args map[string]string) (*model.ExecutionContext, string, error) {
	context := &model.ExecutionContext{
		UserEmail: userEmail,
		JobName:   jobName,
		Args:      args,
		Status:    status.Created,
	}

	defer service.save(context)

	metadata, err := service.metadataRepository.GetByName(jobName)
	if err != nil {
		context.Status = status.RequirementNotMet
		return context, "", errors.New(fmt.Sprintf("metadata not found for %v, throws error %v", jobName, err.Error()))
	}

	secret, err := service.secretRepository.GetByJobName(jobName)
	if err != nil {
		context.Status = status.RequirementNotMet
		return context, "", errors.New(fmt.Sprintf("secret not found for %v, throws error %v", jobName, err.Error()))
	}

	executionArgs := mergeArgs(args, secret)

	context.Status = status.Created
	executionName, err := service.kubernetesClient.ExecuteJob(metadata.ImageName, executionArgs)
	logger.Info("Executed Job on Kubernetes got ", executionName, " execution jobName and ", err, "errors")
	if err != nil {
		context.Status = status.CreationFailed
		return context, "", errors.New(fmt.Sprintf("error when executing image %v with args %v, throws error %v", jobName, args, err.Error()))
	}

	service.watchProcess(executionName, context)

	return context, executionName, nil
}

func (service *executionService) watchProcess(executionName string, context *model.ExecutionContext) {
	waitTime := 15 * time.Minute
	err := service.kubernetesClient.WaitForReadyJob(executionName, waitTime)

	if err != nil {
		context.Status = status.JobCreationFailed
		return
	}

	context.Status = status.JobReady
	logger.Info("Job Ready for ", context.ExecutionID)

	pod, err := service.kubernetesClient.WaitForReadyPod(executionName, waitTime)
	if err != nil {
		context.Status = status.PodCreationFailed
		return
	}

	context.Status = status.PodReady
	logger.Info("Job Ready for ", context.ExecutionID)

	podLog, err := service.kubernetesClient.GetPodLogs(pod)
	if err != nil {
		context.Status = status.FetchPodLogFailed
		return
	}

	scanner := bufio.NewScanner(podLog)
	scanner.Split(bufio.ScanLines)

	var buffer bytes.Buffer
	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
	}

	output := types.GzippedText(buffer.Bytes())

	context.Output = output
	logger.Info("Execution Output Produced ", context.ExecutionID, " with length ", len(output))

	context.Status = status.Finished

	return
}

func mergeArgs(argsOne, argsTwo map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range argsOne {
		result[k] = v
	}
	for k, v := range argsTwo {
		result[k] = v
	}
	return result
}
