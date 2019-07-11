package service

import (
	"errors"
	"fmt"
	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/kubernetes"
	"proctor/internal/app/service/infra/logger"
	svcMetadataRepository "proctor/internal/app/service/metadata/repository"
	svcSecretRepository "proctor/internal/app/service/secret/repository"
)

type ExecutionService interface {
	Execute(jobName string, userEmail string, args map[string]string) (*model.ExecutionContext, string, error)
	Save(executionContext *model.ExecutionContext)
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

func (service *executionService) Save(executionContext *model.ExecutionContext) {
	if executionContext.ExecutionID == 0 {
		_, err := service.repository.Insert(executionContext)
		logger.LogErrors(err, "save execution context to db ", executionContext)
	} else {
		context, err := service.repository.GetById(executionContext.ExecutionID)
		if err != nil || context == nil {
			service.repository.Insert(executionContext)
			logger.LogErrors(err, "save execution context to db ", executionContext)
		} else {
			err = service.repository.UpdateStatus(executionContext.ExecutionID, executionContext.Status)
			logger.LogErrors(err, "update execution context status", executionContext)
			if len(executionContext.Output) > 0 {
				err = service.repository.UpdateJobOutput(executionContext.ExecutionID, executionContext.Output)
				logger.LogErrors(err, "update execution context output", executionContext)
			}
		}
	}
}

func (service *executionService) Execute(name string, userEmail string, args map[string]string) (*model.ExecutionContext, string, error) {
	context := &model.ExecutionContext{
		UserEmail: userEmail,
		JobName:   name,
		Args:      args,
		Status:    status.Created,
	}

	defer service.Save(context)

	metadata, err := service.metadataRepository.GetByName(name)
	if err != nil {
		context.Status = status.RequirementNotMet
		return context, "", errors.New(fmt.Sprintf("metadata not found for %v, throws error %v", name, err.Error()))
	}

	secret, err := service.secretRepository.GetByJobName(name)
	if err != nil {
		context.Status = status.RequirementNotMet
		return context, "", errors.New(fmt.Sprintf("secret not found for %v, throws error %v", name, err.Error()))
	}

	executionArgs := mergeArgs(args, secret)

	context.Status = status.Created
	executionName, err := service.kubernetesClient.ExecuteJob(metadata.ImageName, executionArgs)
	logger.Info("Executed Job on Kubernetes got ", executionName, " execution name and ", err, "errors")
	if err != nil {
		context.Status = status.CreationFailed
		return context, "", errors.New(fmt.Sprintf("error when executing image %v with args %v, throws error %v", name, args, err.Error()))
	}

	return context, executionName, nil
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
