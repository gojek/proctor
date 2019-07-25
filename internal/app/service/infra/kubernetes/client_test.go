package kubernetes

import (
	"net/http"
	"os"
	"proctor/internal/app/service/infra/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	batchV1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	batch "k8s.io/client-go/kubernetes/typed/batch/v1"
	testing_kubernetes "k8s.io/client-go/testing"
	utility "proctor/internal/pkg/constant"
)

type ClientTestSuite struct {
	suite.Suite
	testClient             KubernetesClient
	testKubernetesJobs     batch.JobInterface
	fakeClientSet          *fakeclientset.Clientset
	jobName                string
	podName                string
	fakeClientSetStreaming *fakeclientset.Clientset
	fakeHTTPClient         *http.Client
	testClientStreaming    KubernetesClient
}

func (suite *ClientTestSuite) SetupTest() {
	suite.fakeClientSet = fakeclientset.NewSimpleClientset()
	suite.testClient = &kubernetesClient{
		clientSet: suite.fakeClientSet,
	}
	suite.jobName = "job1"
	suite.podName = "pod1"
	namespace := config.DefaultNamespace()
	suite.fakeClientSetStreaming = fakeclientset.NewSimpleClientset(&v1.Pod{
		TypeMeta: meta.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      suite.podName,
			Namespace: namespace,
			Labels: map[string]string{
				"tag": "",
				"job": suite.jobName,
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	})

	suite.fakeHTTPClient = &http.Client{}
	suite.testClientStreaming = &kubernetesClient{
		clientSet:  suite.fakeClientSetStreaming,
		httpClient: suite.fakeHTTPClient,
	}
}

func (suite *ClientTestSuite) TestJobExecution() {
	t := suite.T()
	_ = os.Setenv("PROCTOR_JOB_POD_ANNOTATIONS", "{\"key.one\":\"true\"}")
	envVarsForContainer := map[string]string{"SAMPLE_ARG": "samle-value"}
	sampleImageName := "img1"

	executedJobname, err := suite.testClient.ExecuteJob(sampleImageName, envVarsForContainer)
	assert.NoError(t, err)

	typeMeta := meta.TypeMeta{
		Kind:       "Job",
		APIVersion: "batch/v1",
	}

	listOptions := meta.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(executedJobname),
	}
	namespace := config.DefaultNamespace()
	listOfJobs, err := suite.fakeClientSet.BatchV1().Jobs(namespace).List(listOptions)
	assert.NoError(t, err)
	executedJob := listOfJobs.Items[0]

	assert.Equal(t, typeMeta, executedJob.TypeMeta)

	assert.Equal(t, executedJobname, executedJob.ObjectMeta.Name)
	assert.Equal(t, executedJobname, executedJob.Spec.Template.ObjectMeta.Name)

	expectedLabel := jobLabel(executedJobname)
	assert.Equal(t, expectedLabel, executedJob.ObjectMeta.Labels)
	assert.Equal(t, expectedLabel, executedJob.Spec.Template.ObjectMeta.Labels)
	assert.Equal(t, map[string]string{"key.one": "true"}, executedJob.Spec.Template.Annotations)

	assert.Equal(t, config.KubeJobActiveDeadlineSeconds(), executedJob.Spec.ActiveDeadlineSeconds)
	assert.Equal(t, config.KubeJobRetries(), executedJob.Spec.BackoffLimit)

	assert.Equal(t, v1.RestartPolicyNever, executedJob.Spec.Template.Spec.RestartPolicy)

	container := executedJob.Spec.Template.Spec.Containers[0]
	assert.Equal(t, executedJobname, container.Name)

	assert.Equal(t, sampleImageName, container.Image)

	expectedEnvVars := getEnvVars(envVarsForContainer)
	assert.Equal(t, expectedEnvVars, container.Env)
}

func (suite *ClientTestSuite) TestShouldReturnSuccessJobExecutionStatus() {
	t := suite.T()

	watcher := watch.NewFake()
	suite.fakeClientSet.PrependWatchReactor("jobs", testing_kubernetes.DefaultWatchReactor(watcher, nil))

	var activeJob batchV1.Job
	var succeededJob batchV1.Job
	uniqueJobName := "proctor-job-2"
	label := jobLabel(uniqueJobName)
	objectMeta := meta.ObjectMeta{
		Name:   uniqueJobName,
		Labels: label,
	}
	activeJob.ObjectMeta = objectMeta
	succeededJob.ObjectMeta = objectMeta

	go func() {
		activeJob.Status.Active = 1
		watcher.Modify(&activeJob)

		succeededJob.Status.Active = 0
		succeededJob.Status.Succeeded = 1
		watcher.Modify(&succeededJob)

		time.Sleep(time.Second * 1)
		watcher.Stop()
	}()

	jobExecutionStatus, err := suite.testClient.JobExecutionStatus(uniqueJobName)
	assert.NoError(t, err)

	assert.Equal(t, utility.JobSucceeded, jobExecutionStatus, "Should return SUCCEEDED")
}

func (suite *ClientTestSuite) TestShouldReturnFailedJobExecutionStatus() {
	t := suite.T()

	watcher := watch.NewFake()
	suite.fakeClientSet.PrependWatchReactor("jobs", testing_kubernetes.DefaultWatchReactor(watcher, nil))

	var activeJob batchV1.Job
	var failedJob batchV1.Job
	uniqueJobName := "proctor-job-1"
	label := jobLabel(uniqueJobName)
	objectMeta := meta.ObjectMeta{
		Name:   uniqueJobName,
		Labels: label,
	}
	activeJob.ObjectMeta = objectMeta
	failedJob.ObjectMeta = objectMeta

	go func() {
		activeJob.Status.Active = 1
		watcher.Modify(&activeJob)
		failedJob.Status.Active = 0
		failedJob.Status.Failed = 1
		watcher.Modify(&failedJob)

		time.Sleep(time.Second * 1)
		watcher.Stop()
	}()

	jobExecutionStatus, err := suite.testClient.JobExecutionStatus(uniqueJobName)
	assert.NoError(t, err)

	assert.Equal(t, utility.JobFailed, jobExecutionStatus, "Should return FAILED")
}

func (suite *ClientTestSuite) TestJobExecutionStatusForNonDefinitiveStatus() {
	t := suite.T()

	watcher := watch.NewFake()
	suite.fakeClientSet.PrependWatchReactor("jobs", testing_kubernetes.DefaultWatchReactor(watcher, nil))

	var testJob batchV1.Job
	uniqueJobName := "proctor-job-1"
	label := jobLabel(uniqueJobName)
	objectMeta := meta.ObjectMeta{
		Name:   uniqueJobName,
		Labels: label,
	}
	testJob.ObjectMeta = objectMeta

	go func() {
		testJob.Status.Active = 1
		watcher.Modify(&testJob)

		time.Sleep(time.Second * 1)
		watcher.Stop()
	}()

	jobExecutionStatus, err := suite.testClient.JobExecutionStatus(uniqueJobName)
	assert.NoError(t, err)

	assert.Equal(t, utility.NoDefinitiveJobExecutionStatusFound, jobExecutionStatus, "Should return NO_DEFINITIVE_JOB_EXECUTION_STATUS_FOUND")
}

func (suite *ClientTestSuite) TestShouldReturnJobExecutionStatusFetchError() {
	t := suite.T()

	watcher := watch.NewFake()
	suite.fakeClientSet.PrependWatchReactor("jobs", testing_kubernetes.DefaultWatchReactor(watcher, nil))

	var testJob batchV1.Job
	uniqueJobName := "proctor-job-3"
	label := jobLabel(uniqueJobName)
	objectMeta := meta.ObjectMeta{
		Name:   uniqueJobName,
		Labels: label,
	}
	testJob.ObjectMeta = objectMeta

	go func() {
		watcher.Error(&testJob)

		time.Sleep(time.Second * 1)
		watcher.Stop()
	}()

	jobExecutionStatus, err := suite.testClient.JobExecutionStatus(uniqueJobName)
	assert.NoError(t, err)

	assert.Equal(t, utility.JobExecutionStatusFetchError, jobExecutionStatus, "Should return JOB_EXECUTION_STATUS_FETCH_ERROR")
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
