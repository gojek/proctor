package execution

import (
	"errors"
	"proctor/internal/app/proctord/storage/postgres"
	"proctor/internal/app/service/infra/kubernetes"
	metadataRepository "proctor/internal/app/service/metadata/repository"
	secretRepository "proctor/internal/app/service/secret/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	modelMetadata "proctor/internal/pkg/model/metadata"
	"proctor/internal/pkg/utility"
)

type ExecutionerTestSuite struct {
	suite.Suite
	mockKubeClient    kubernetes.MockKubernetesClient
	mockMetadataStore *metadataRepository.MockMetadataRepository
	mockSecretsStore  *secretRepository.MockSecretRepository
	testExecutioner   Executioner
}

func (suite *ExecutionerTestSuite) SetupTest() {
	suite.mockKubeClient = kubernetes.MockKubernetesClient{}
	suite.mockMetadataStore = &metadataRepository.MockMetadataRepository{}
	suite.mockSecretsStore = &secretRepository.MockSecretRepository{}
	suite.testExecutioner = NewExecutioner(&suite.mockKubeClient, suite.mockMetadataStore, suite.mockSecretsStore)
}

func (suite *ExecutionerTestSuite) TestSuccessfulJobExecution() {
	t := suite.T()

	jobsExecutionAuditLog := &postgres.JobsExecutionAuditLog{}
	jobName := "sample-job-name"
	jobArgs := map[string]string{
		"argOne": "sample-arg",
	}

	jobMetadata := modelMetadata.Metadata{
		ImageName: "img",
	}
	suite.mockMetadataStore.On("GetByName", jobName).Return(&jobMetadata, nil).Once()

	jobSecrets := map[string]string{
		"secretOne": "sample-secrets",
	}
	suite.mockSecretsStore.On("GetByJobName", jobName).Return(jobSecrets, nil).Once()

	jobExecutionID := "proctor-ipsum-lorem"
	envVarsForJob := utility.MergeMaps(jobArgs, jobSecrets)
	suite.mockKubeClient.On("ExecuteJob", jobMetadata.ImageName, envVarsForJob).Return(jobExecutionID, nil).Once()

	executedJobName, err := suite.testExecutioner.Execute(jobsExecutionAuditLog, jobName, jobArgs)
	assert.NoError(t, err)

	suite.mockMetadataStore.AssertExpectations(t)
	suite.mockSecretsStore.AssertExpectations(t)
	suite.mockKubeClient.AssertExpectations(t)

	assert.Equal(t, jobExecutionID, executedJobName)
	assert.Equal(t, jobsExecutionAuditLog.JobName, jobName)
}

func (suite *ExecutionerTestSuite) TestJobExecutionOnImageLookupFailure() {
	t := suite.T()

	suite.mockMetadataStore.On("GetByName", mock.Anything).Return(&modelMetadata.Metadata{}, errors.New("image-fetch-error")).Once()

	_, err := suite.testExecutioner.Execute(&postgres.JobsExecutionAuditLog{}, "any-job", map[string]string{})
	assert.EqualError(t, err, "Error finding image for job: any-job. Error: image-fetch-error")
}

func (suite *ExecutionerTestSuite) TestJobExecutionOnSecretsFetchFailure() {
	t := suite.T()

	jobMetadata := modelMetadata.Metadata{ImageName: "img"}
	suite.mockMetadataStore.On("GetByName", mock.Anything).Return(&jobMetadata, nil).Once()

	suite.mockSecretsStore.On("GetByJobName", mock.Anything).Return(map[string]string{}, errors.New("secret-store-error")).Once()

	_, err := suite.testExecutioner.Execute(&postgres.JobsExecutionAuditLog{}, "any-job", map[string]string{})
	assert.EqualError(t, err, "Error retrieving secrets for job: any-job. Error: secret-store-error")
}

func (suite *ExecutionerTestSuite) TestJobExecutionOnKubernetesJobExecutionFailure() {
	t := suite.T()

	jobMetadata := modelMetadata.Metadata{ImageName: "img"}
	suite.mockMetadataStore.On("GetByName", mock.Anything).Return(&jobMetadata, nil).Once()

	suite.mockSecretsStore.On("GetByJobName", mock.Anything).Return(map[string]string{}, nil).Once()
	suite.mockKubeClient.On("ExecuteJob", mock.Anything, mock.Anything).Return("", errors.New("kube-client-error")).Once()

	_, err := suite.testExecutioner.Execute(&postgres.JobsExecutionAuditLog{}, "any-job", map[string]string{})

	assert.EqualError(t, err, "Error submitting job to kube: any-job. Error: kube-client-error")
}

func TestExecutionerTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutionerTestSuite))
}
