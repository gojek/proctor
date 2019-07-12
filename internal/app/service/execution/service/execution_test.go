package service

import (
	fake "github.com/brianvoe/gofakeit"
	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/execution/repository"
	"proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/id"
	"proctor/internal/app/service/infra/kubernetes"
	svcMetadataRepository "proctor/internal/app/service/metadata/repository"
	svcSecretRepository "proctor/internal/app/service/secret/repository"
	"proctor/internal/pkg/model/metadata"
	"testing"
)

type TestExecutionServiceSuite struct {
	suite.Suite
	service                ExecutionService
	mockKubernetesClient   *kubernetes.MockKubernetesClient
	mockRepository         *repository.MockExecutionContextRepository
	mockMetadataRepository *svcMetadataRepository.MockMetadataRepository
	mockSecretRepository   *svcSecretRepository.MockSecretRepository
}

func (suite *TestExecutionServiceSuite) SetupTest() {
	suite.mockKubernetesClient = &kubernetes.MockKubernetesClient{}
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

func (suite *TestExecutionServiceSuite) TestSaveNoExecutionId() {
	t := suite.T()
	context := &model.ExecutionContext{}

	suite.mockRepository.On("Insert", context).Return(0, errors.New("Insert Failed")).Once()
	err := suite.service.save(context)
	assert.Error(t, err, "Insert Failed")

	suite.mockRepository.On("Insert", context).Return(0, nil).Once()
	err = suite.service.save(context)
	assert.NilError(t, err)
}

func (suite *TestExecutionServiceSuite) TestSaveWithExecutionId() {
	t := suite.T()
	id, _ := id.NextId()
	context := &model.ExecutionContext{
		ExecutionID: id,
		Status:      status.Created,
	}

	suite.mockRepository.On("GetById", id).Return(context, errors.New("Get By Id Error")).Once()
	suite.mockRepository.On("Insert", context).Return(0, errors.New("Insert Failed")).Once()
	err := suite.service.save(context)
	assert.Error(t, err, "Insert Failed")

	suite.mockRepository.On("GetById", id).Return(context, nil).Once()
	suite.mockRepository.On("UpdateStatus", context.ExecutionID, context.Status).Return(errors.New("Update Status Failed")).Once()
	err = suite.service.save(context)
	assert.Error(t, err, "Update Status Failed")

	context.Output = types.GzippedText("This is some output")
	suite.mockRepository.On("GetById", id).Return(context, nil).Once()
	suite.mockRepository.On("UpdateStatus", context.ExecutionID, context.Status).Return(nil).Once()
	suite.mockRepository.On("UpdateJobOutput", context.ExecutionID, context.Output).Return(errors.New("Update Output Failed")).Once()
	err = suite.service.save(context)
	assert.Error(t, err, "Update Output Failed")
}

func (suite *TestExecutionServiceSuite) TestExecuteMetadataNotFound() {
	t := suite.T()
	jobName := fake.Username()
	userEmail := fake.Email()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()

	jobArgs := map[string]string{
		mapKey: mapValue,
	}

	suite.mockMetadataRepository.On("GetByName", jobName).Return(&metadata.Metadata{}, errors.New("metadataNotFound")).Once()
	suite.mockRepository.On("Insert", mock.Anything).Return(0, nil).Once()

	context, _, err := suite.service.Execute(jobName, userEmail, jobArgs)

	assert.Error(t, err, "metadata not found")
	assert.NotNil(t, context)
	assert.Equal(t, context.Status, status.RequirementNotMet)
}

func (suite *TestExecutionServiceSuite) TestExecuteSecretNotFound() {
	t := suite.T()
	jobName := fake.Username()
	userEmail := fake.Email()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()

	jobArgs := map[string]string{
		mapKey: mapValue,
	}

	suite.mockMetadataRepository.On("GetByName", jobName).Return(&metadata.Metadata{}, nil).Once()
	suite.mockSecretRepository.On("GetByJobName", jobName).Return(map[string]string{}, errors.New("secret not found")).Once()
	suite.mockRepository.On("Insert", mock.Anything).Return(0, nil).Once()

	context, _, err := suite.service.Execute(jobName, userEmail, jobArgs)
	assert.Error(t, err, "secret not found")
	assert.NotNil(t, context)
	assert.Equal(t, context.Status, status.RequirementNotMet)
}

func (suite *TestExecutionServiceSuite) TestExecuteJobFailed() {
	t := suite.T()
	jobName := fake.Username()
	userEmail := fake.Email()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()

	jobArgs := map[string]string{
		mapKey: mapValue,
	}

	imageName := fake.BeerYeast()
	fakeMetadata := &metadata.Metadata{
		ImageName: imageName,
	}

	suite.mockMetadataRepository.On("GetByName", jobName).Return(fakeMetadata, nil).Once()
	suite.mockSecretRepository.On("GetByJobName", jobName).Return(map[string]string{}, nil).Once()
	suite.mockRepository.On("Insert", mock.Anything).Return(0, nil).Once()
	suite.mockKubernetesClient.On("ExecuteJob", imageName, jobArgs).Return("", errors.New("Execution Failed"))

	context, _, err := suite.service.Execute(jobName, userEmail, jobArgs)
	assert.Error(t, err, "error when executing image")
	assert.NotNil(t, context)
	assert.Equal(t, context.Status, status.CreationFailed)
}

func (suite *TestExecutionServiceSuite) TestExecuteJobSuccess() {
	t := suite.T()
	jobName := fake.Username()
	userEmail := fake.Email()
	mapKey := fake.FirstName()
	mapValue := fake.LastName()

	jobArgs := map[string]string{
		mapKey: mapValue,
	}

	imageName := fake.BeerYeast()
	fakeMetadata := &metadata.Metadata{
		ImageName: imageName,
	}

	suite.mockMetadataRepository.On("GetByName", jobName).Return(fakeMetadata, nil).Once()
	suite.mockSecretRepository.On("GetByJobName", jobName).Return(map[string]string{}, nil).Once()
	suite.mockRepository.On("Insert", mock.Anything).Return(0, nil).Once()
	suite.mockKubernetesClient.On("ExecuteJob", imageName, jobArgs).Return("", nil)

	context, _, err := suite.service.Execute(jobName, userEmail, jobArgs)
	assert.NilError(t, err)
	assert.NotNil(t, context)
	assert.Equal(t, context.Status, status.Created)
}

func TestExecutionServiceSuiteTest(t *testing.T) {
	suite.Run(t, new(TestExecutionServiceSuite))
}
