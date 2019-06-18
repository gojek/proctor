package execution

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	jobsMetadata "proctor/proctord/jobs/metadata"
	"proctor/proctord/jobs/secrets"
	"proctor/proctord/kubernetes"
	"proctor/proctord/storage/postgres"
	"proctor/shared/model/metadata"
	"proctor/shared/utility"
)

type ExecutionerTestSuite struct {
	suite.Suite
	mockKubeClient    kubernetes.MockClient
	mockMetadataStore *jobsMetadata.MockStore
	mockSecretsStore  *secrets.MockStore
	testExecutioner   Executioner
}

func (suite *ExecutionerTestSuite) SetupTest() {
	suite.mockKubeClient = kubernetes.MockClient{}
	suite.mockMetadataStore = &jobsMetadata.MockStore{}
	suite.mockSecretsStore = &secrets.MockStore{}
	suite.testExecutioner = NewExecutioner(&suite.mockKubeClient, suite.mockMetadataStore, suite.mockSecretsStore)
}

func (suite *ExecutionerTestSuite) TestSuccessfulJobExecution() {
	t := suite.T()

	jobsExecutionAuditLog := &postgres.JobsExecutionAuditLog{}
	jobName := "sample-job-name"
	jobArgs := map[string]string{
		"argOne": "sample-arg",
	}

	jobMetadata := metadata.Metadata{
		ImageName: "img",
	}
	suite.mockMetadataStore.On("GetJobMetadata", jobName).Return(&jobMetadata, nil).Once()

	jobSecrets := map[string]string{
		"secretOne": "sample-secrets",
	}
	suite.mockSecretsStore.On("GetJobSecrets", jobName).Return(jobSecrets, nil).Once()

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

	suite.mockMetadataStore.On("GetJobMetadata", mock.Anything).Return(&metadata.Metadata{}, errors.New("image-fetch-error")).Once()

	_, err := suite.testExecutioner.Execute(&postgres.JobsExecutionAuditLog{}, "any-job", map[string]string{})
	assert.EqualError(t, err, "Error finding image for job: any-job. Error: image-fetch-error")
}

func (suite *ExecutionerTestSuite) TestJobExecutionOnSecretsFetchFailure() {
	t := suite.T()

	jobMetadata := metadata.Metadata{ImageName: "img"}
	suite.mockMetadataStore.On("GetJobMetadata", mock.Anything).Return(&jobMetadata, nil).Once()

	suite.mockSecretsStore.On("GetJobSecrets", mock.Anything).Return(map[string]string{}, errors.New("secret-store-error")).Once()

	_, err := suite.testExecutioner.Execute(&postgres.JobsExecutionAuditLog{}, "any-job", map[string]string{})
	assert.EqualError(t, err, "Error retrieving secrets for job: any-job. Error: secret-store-error")
}

func (suite *ExecutionerTestSuite) TestJobExecutionOnKubernetesJobExecutionFailure() {
	t := suite.T()

	jobMetadata := metadata.Metadata{ImageName: "img"}
	suite.mockMetadataStore.On("GetJobMetadata", mock.Anything).Return(&jobMetadata, nil).Once()

	suite.mockSecretsStore.On("GetJobSecrets", mock.Anything).Return(map[string]string{}, nil).Once()
	suite.mockKubeClient.On("ExecuteJob", mock.Anything, mock.Anything).Return("", errors.New("kube-client-error")).Once()

	_, err := suite.testExecutioner.Execute(&postgres.JobsExecutionAuditLog{}, "any-job", map[string]string{})

	assert.EqualError(t, err, "Error submitting job to kube: any-job. Error: kube-client-error")
}

func TestExecutionerTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutionerTestSuite))
}
