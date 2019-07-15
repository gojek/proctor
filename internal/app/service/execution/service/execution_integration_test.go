package service

import (
	fake "github.com/brianvoe/gofakeit"
	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/kubernetes"
	"proctor/internal/app/service/infra/kubernetes/http"
	svcMetadataRepository "proctor/internal/app/service/metadata/repository"
	svcSecretRepository "proctor/internal/app/service/secret/repository"
	"proctor/internal/pkg/model/metadata"
	"proctor/internal/pkg/model/metadata/env"
	"testing"
	"time"
)

type TestExecutionIntegrationSuite struct {
	suite.Suite
	service                ExecutionService
	mockKubernetesClient   kubernetes.KubernetesClient
	repository             repository.ExecutionContextRepository
	mockMetadataRepository *svcMetadataRepository.MockMetadataRepository
	mockSecretRepository   *svcSecretRepository.MockSecretRepository
}

func (suite *TestExecutionIntegrationSuite) SetupTest() {
	httpClient, _ := http.NewClient()
	suite.mockKubernetesClient = kubernetes.NewKubernetesClient(httpClient)
	pgClient := postgresql.NewClient()
	suite.repository = repository.NewExecutionContextRepository(pgClient)
	suite.mockMetadataRepository = &svcMetadataRepository.MockMetadataRepository{}
	suite.mockSecretRepository = &svcSecretRepository.MockSecretRepository{}
	suite.service = NewExecutionService(
		suite.mockKubernetesClient,
		suite.repository,
		suite.mockMetadataRepository,
		suite.mockSecretRepository,
	)
}

func (suite *TestExecutionIntegrationSuite) TestExecuteJobSuccess() {
	t := suite.T()
	jobName := fake.Username()
	userEmail := fake.Email()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()

	jobArgs := map[string]string{
		mapKey: mapValue,
	}

	imageName := "ubuntu"
	fakeMetadata := &metadata.Metadata{
		ImageName:        imageName,
		Author:           "bimo.horizon",
		Description:      fake.HackerIngverb(),
		Organization:     fake.BuzzWord(),
		AuthorizedGroups: []string{},
		EnvVars: env.Vars{
			Args: []env.VarMetadata{
				{
					Name:        fake.BeerYeast(),
					Description: fake.JobDescriptor(),
				},
			},
			Secrets: []env.VarMetadata{},
		},
	}

	suite.mockMetadataRepository.On("GetByName", jobName).Return(fakeMetadata, nil).Once()
	suite.mockSecretRepository.On("GetByJobName", jobName).Return(map[string]string{}, nil).Once()

	context, _, err := suite.service.ExecuteWithCommand(jobName, userEmail, jobArgs, []string{"bash", "-c", "for run in {1..10}; do sleep 1 && echo bimo; done"})
	assert.NilError(t, err)
	assert.NotNil(t, context)

	time.Sleep(30 * time.Second)
	expectedContext, err := suite.repository.GetById(context.ExecutionID)
	assert.NilError(t, err)
	assert.NotNil(t, expectedContext)
	assert.Equal(t, expectedContext.Status, status.Finished)
	assert.NotNil(t, expectedContext.Output)
}

func TestExecutionIntegrationSuiteTest(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_INTEGRATION_TEST")
	if available == true && value == "true" {
		suite.Run(t, new(TestExecutionIntegrationSuite))
	}
}
