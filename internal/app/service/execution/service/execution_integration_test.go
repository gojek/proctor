package service

import (
	"github.com/stretchr/testify/suite"
	"os"
	"proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/infra/kubernetes"
	"proctor/internal/app/service/infra/kubernetes/http"
	svcMetadataRepository "proctor/internal/app/service/metadata/repository"
	svcSecretRepository "proctor/internal/app/service/secret/repository"
	"testing"
)

type TestExecutionIntegrationSuite struct {
	suite.Suite
	service                ExecutionService
	mockKubernetesClient   kubernetes.KubernetesClient
	mockRepository         *repository.MockExecutionContextRepository
	mockMetadataRepository *svcMetadataRepository.MockMetadataRepository
	mockSecretRepository   *svcSecretRepository.MockSecretRepository
}

func (suite *TestExecutionIntegrationSuite) SetupTest() {
	httpClient, _ := http.NewClient()
	suite.mockKubernetesClient = kubernetes.NewKubernetesClient(httpClient)
	suite.mockRepository = &repository.MockExecutionContextRepository{}
	suite.mockMetadataRepository = &svcMetadataRepository.MockMetadataRepository{}
	suite.mockSecretRepository = &svcSecretRepository.MockSecretRepository{}
	suite.service = NewExecutionService(
		suite.mockKubernetesClient,
		suite.mockRepository,
		suite.mockMetadataRepository,
		suite.mockSecretRepository,
	)
}

func TestExecutionIntegrationSuiteTest(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_INTEGRATION_TEST")
	if available == true && value == "true" {
		suite.Run(t, new(TestExecutionIntegrationSuite))
	}
}
