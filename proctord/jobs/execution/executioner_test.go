package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/jobs/secrets"
	"github.com/gojektech/proctor/proctord/kubernetes"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ExecutionerTestSuite struct {
	suite.Suite
	mockKubeClient    kubernetes.MockClient
	mockMetadataStore *metadata.MockStore
	mockSecretsStore  *secrets.MockStore
	testExecutioner   Executioner
}

func (suite *ExecutionerTestSuite) SetupTest() {
	suite.mockKubeClient = kubernetes.MockClient{}
	suite.mockMetadataStore = &metadata.MockStore{}
	suite.mockSecretsStore = &secrets.MockStore{}
	suite.testExecutioner = NewExecutioner(&suite.mockKubeClient, suite.mockMetadataStore, suite.mockSecretsStore)
}

func (suite *ExecutionerTestSuite) TestSuccessfulJobExecution() {
	t := suite.T()

	jobName := "sample-job-name"
	userEmail := "mrproctor@example.com"
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

	jobNameSubmittedForExecution := "proctor-ipsum-lorem"
	envVarsForJob := utility.MergeMaps(jobArgs, jobSecrets)
	suite.mockKubeClient.On("ExecuteJob", jobMetadata.ImageName, envVarsForJob).Return(jobNameSubmittedForExecution, nil).Once()

	executedJobName, err := suite.testExecutioner.Execute(context.Background(), jobName, userEmail, jobArgs)
	assert.NoError(t, err)

	suite.mockMetadataStore.AssertExpectations(t)
	suite.mockSecretsStore.AssertExpectations(t)
	suite.mockKubeClient.AssertExpectations(t)

	assert.Equal(t, jobNameSubmittedForExecution, executedJobName)
}

func (suite *ExecutionerTestSuite) TestJobExecutionOnImageLookupFailure() {
	t := suite.T()

	suite.mockMetadataStore.On("GetJobMetadata", mock.Anything).Return(&metadata.Metadata{}, errors.New("No image found for job name")).Once()

	_, err := suite.testExecutioner.Execute(context.Background(), "any-job", "foo@bar.com", map[string]string{})
	assert.EqualError(t, err, "No image found for job name")
}

func (suite *ExecutionerTestSuite) TestJobExecutionOnSecretsFetchFailure() {
	t := suite.T()

	jobMetadata := metadata.Metadata{ImageName: "img"}
	suite.mockMetadataStore.On("GetJobMetadata", mock.Anything).Return(&jobMetadata, nil).Once()

	suite.mockSecretsStore.On("GetJobSecrets", mock.Anything).Return(map[string]string{}, errors.New("secrets fetch error")).Once()

	_, err := suite.testExecutioner.Execute(context.Background(), "any-job", "foo@bar.com", map[string]string{})
	assert.EqualError(t, err, "secrets fetch error")
}

func (suite *ExecutionerTestSuite) TestJobExecutionOnKubernetesJobExecutionFailure() {
	t := suite.T()

	jobMetadata := metadata.Metadata{ImageName: "img"}
	suite.mockMetadataStore.On("GetJobMetadata", mock.Anything).Return(&jobMetadata, nil).Once()

	suite.mockSecretsStore.On("GetJobSecrets", mock.Anything).Return(map[string]string{}, nil).Once()
	suite.mockKubeClient.On("ExecuteJob", mock.Anything, mock.Anything).Return("", errors.New("Kube client job execution error")).Once()

	_, err := suite.testExecutioner.Execute(context.Background(), "any-job", "foo@bar.com", map[string]string{})

	assert.EqualError(t, err, "Kube client job execution error")
}

func TestExecutionerTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutionerTestSuite))
}
