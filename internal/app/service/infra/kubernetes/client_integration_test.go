package kubernetes

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"proctor/internal/app/service/infra/config"
	kubeHTTPClient "proctor/internal/app/service/infra/kubernetes/http"
	"proctor/internal/pkg/constant"
	"testing"
)

type IntegrationTestSuite struct {
	suite.Suite
	testClient KubernetesClient
	clientSet  kubernetes.Interface
}

func (suite *IntegrationTestSuite) SetupTest() {
	t := suite.T()
	kubeHTTPClient, err := kubeHTTPClient.NewClient()
	assert.NoError(t, err)
	suite.testClient = NewKubernetesClient(kubeHTTPClient)
	suite.clientSet, err = NewClientSet()
	assert.NoError(t, err)
}

func (suite *IntegrationTestSuite) TestJobExecution() {
	t := suite.T()
	_ = os.Setenv("PROCTOR_JOB_POD_ANNOTATIONS", "{\"key.one\":\"true\"}")
	config.Reset()
	envVarsForContainer := map[string]string{"SAMPLE_ARG": "samle-value"}
	sampleImageName := "busybox"

	executedJobname, err := suite.testClient.ExecuteJobWithCommand(sampleImageName, envVarsForContainer, []string{"echo", "Bimo Horizon"})
	assert.NoError(t, err)

	typeMeta := meta_v1.TypeMeta{
		Kind:       "Job",
		APIVersion: "batch/v1",
	}

	listOptions := meta_v1.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(executedJobname),
	}

	namespace := config.Config().DefaultNamespace
	listOfJobs, err := suite.clientSet.BatchV1().Jobs(namespace).List(listOptions)
	assert.NoError(t, err)
	executedJob := listOfJobs.Items[0]

	assert.Equal(t, executedJobname, executedJob.ObjectMeta.Name)
	assert.Equal(t, executedJobname, executedJob.Spec.Template.ObjectMeta.Name)

	expectedLabel := jobLabel(executedJobname)
	assert.Equal(t, expectedLabel, executedJob.ObjectMeta.Labels)
	assert.Equal(t, map[string]string{"key.one": "true"}, executedJob.Spec.Template.Annotations)

	assert.Equal(t, config.Config().KubeJobActiveDeadlineSeconds, executedJob.Spec.ActiveDeadlineSeconds)
	assert.Equal(t, config.Config().KubeJobRetries, executedJob.Spec.BackoffLimit)

	assert.Equal(t, v1.RestartPolicyNever, executedJob.Spec.Template.Spec.RestartPolicy)

	container := executedJob.Spec.Template.Spec.Containers[0]
	assert.Equal(t, executedJobname, container.Name)

	assert.Equal(t, sampleImageName, container.Image)

	expectedEnvVars := getEnvVars(envVarsForContainer)
	assert.Equal(t, expectedEnvVars, container.Env)
}

func (suite *IntegrationTestSuite) TestJobExecutionStatus() {
	t := suite.T()
	_ = os.Setenv("PROCTOR_JOB_POD_ANNOTATIONS", "{\"key.one\":\"true\"}")
	envVarsForContainer := map[string]string{"SAMPLE_ARG": "samle-value"}
	sampleImageName := "busybox"

	executedJobname, err := suite.testClient.ExecuteJobWithCommand(sampleImageName, envVarsForContainer, []string{"echo", "Bimo Horizon"})
	assert.NoError(t, err)

	status, err := suite.testClient.JobExecutionStatus(executedJobname)
	assert.Equal(t, status, constant.JobSucceeded)
}

func TestIntegrationTestSuite(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_INTEGRATION_TEST")
	if available == true && value == "true" {
		suite.Run(t, new(IntegrationTestSuite))
	}
}
