package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx/types"
	"io"
	v1 "k8s.io/api/core/v1"
	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/id"
	"proctor/internal/app/service/infra/kubernetes"
	"proctor/internal/app/service/infra/logger"
	svcMetadataRepository "proctor/internal/app/service/metadata/repository"
	svcSecretRepository "proctor/internal/app/service/secret/repository"
	"time"
)

type ExecutionService interface {
	Execute(jobName string, userEmail string, args map[string]string) (*model.ExecutionContext, string, error)
	ExecuteWithCommand(jobName string, userEmail string, args map[string]string, commands []string) (*model.ExecutionContext, string, error)
	StreamJobLogs(executionName string, waitTime time.Duration) (io.ReadCloser, error)
	update(executionContext model.ExecutionContext)
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

func (service *executionService) StreamJobLogs(executionName string, waitTime time.Duration) (io.ReadCloser, error) {
	err := service.kubernetesClient.WaitForReadyJob(executionName, waitTime)
	if err != nil {
		return nil, err
	}

	pod, err := service.kubernetesClient.WaitForReadyPod(executionName, waitTime)
	if err != nil {
		return nil, err
	}

	result, err := service.kubernetesClient.GetPodLogs(pod)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (service *executionService) Execute(jobName string, userEmail string, args map[string]string) (*model.ExecutionContext, string, error) {
	return service.ExecuteWithCommand(jobName, userEmail, args, []string{})
}

func (service *executionService) ExecuteWithCommand(jobName string, userEmail string, args map[string]string, commands []string) (*model.ExecutionContext, string, error) {
	executionID, _ := id.NextID()
	context := model.ExecutionContext{
		ExecutionID: executionID,
		UserEmail:   userEmail,
		JobName:     jobName,
		Args:        args,
		Status:      status.Created,
	}

	metadata, err := service.metadataRepository.GetByName(jobName)
	if err != nil {
		context.Status = status.RequirementNotMet
		service.insertContext(context)
		return &context, "", errors.New(fmt.Sprintf("metadata not found for %v, throws error %v", jobName, err.Error()))
	}

	context.ImageTag = metadata.ImageName
	secret, err := service.secretRepository.GetByJobName(jobName)
	if err != nil {
		context.Status = status.RequirementNotMet
		service.insertContext(context)
		return &context, "", errors.New(fmt.Sprintf("secret not found for %v, throws error %v", jobName, err.Error()))
	}

	executionArgs := mergeArgs(args, secret)

	context.Status = status.Created
	executionName, err := service.kubernetesClient.ExecuteJobWithCommand(metadata.ImageName, executionArgs, commands)
	logger.Info("Executed Job on Kubernetes got ", executionName, " execution jobName and ", err, "errors")
	if err != nil {
		context.Status = status.CreationFailed
		service.insertContext(context)
		return &context, "", errors.New(fmt.Sprintf("error when executing image %v with args %v, throws error %v", jobName, args, err.Error()))
	}

	context.Name = executionName

	service.insertContext(context)
	go service.watchProcess(context)

	return &context, executionName, nil
}

func (service *executionService) insertContext(context model.ExecutionContext) {
	_, err := service.repository.Insert(context)
	logger.LogErrors(err, "save execution context to db", context)
}

func (service *executionService) update(executionContext model.ExecutionContext) {
	err := service.repository.UpdateStatus(executionContext.ExecutionID, executionContext.Status)
	logger.LogErrors(err, "update execution context status", executionContext)
	if len(executionContext.Output) > 0 {
		err = service.repository.UpdateJobOutput(executionContext.ExecutionID, executionContext.Output)
		logger.LogErrors(err, "update execution context output", executionContext)
	}
}

func (service *executionService) watchProcess(context model.ExecutionContext) {

	waitTime := config.KubeLogProcessWaitTime() * time.Second
	err := service.kubernetesClient.WaitForReadyJob(context.Name, waitTime)

	if err != nil {
		context.Status = status.JobCreationFailed
		return
	}

	context.Status = status.JobReady
	logger.Info("Job Ready for ", context.ExecutionID)

	pod, err := service.kubernetesClient.WaitForReadyPod(context.Name, waitTime)
	if err != nil {
		context.Status = status.PodCreationFailed
		return
	}

	if pod.Status.Phase == v1.PodFailed {
		context.Status = status.PodFailed
		logger.Info("Pod Failed for ", context.ExecutionID, " reason: ", pod.Status.Reason, " message: ", pod.Status.Message)
	} else {
		context.Status = status.PodReady
		logger.Info("Pod Ready for ", context.ExecutionID)
	}

	podLog, err := service.kubernetesClient.GetPodLogs(pod)
	if err != nil {
		context.Status = status.FetchPodLogFailed
		return
	}

	scanner := bufio.NewScanner(podLog)
	scanner.Split(bufio.ScanLines)

	var buffer bytes.Buffer
	for scanner.Scan() {
		buffer.WriteString(scanner.Text() + "\n")
	}

	output := types.GzippedText(buffer.Bytes())

	context.Output = output
	logger.Info("Execution Output Produced ", context.ExecutionID, " with length ", len(output))

	if context.Status == status.PodReady {
		context.Status = status.Finished
	}

	defer service.update(context)

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
