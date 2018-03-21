package kubernetes

import (
	"bufio"
	"testing"
	"time"

	"github.com/gojektech/proctor/proctord/config"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	batch_v1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/pkg/api/v1"

	fakeclientset "k8s.io/client-go/kubernetes/fake"
	batchV1 "k8s.io/client-go/pkg/apis/batch/v1"
	testing_kubernetes "k8s.io/client-go/testing"
)

type ClientTestSuite struct {
	suite.Suite
	testClient             Client
	testKubernetesJobs     batch_v1.JobInterface
	fakeClientSet          *fakeclientset.Clientset
	jobName                string
	podName                string
	fakeClientSetStreaming *fakeclientset.Clientset
	testClientStreaming    Client
}

func (suite *ClientTestSuite) SetupTest() {
	suite.fakeClientSet = fakeclientset.NewSimpleClientset()
	suite.testClient = &client{
		clientSet: suite.fakeClientSet,
	}
	suite.jobName = "job1"
	suite.podName = "pod1"
	namespace := config.DefaultNamespace()
	suite.fakeClientSetStreaming = fakeclientset.NewSimpleClientset(&v1.Pod{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
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

	suite.testClientStreaming = &client{
		clientSet: suite.fakeClientSetStreaming,
	}
}

func (suite *ClientTestSuite) TestJobExecution() {
	t := suite.T()

	envVarsForContainer := map[string]string{"SAMPLE_ARG": "samle-value"}
	sampleImageName := "img1"

	executedJobname, err := suite.testClient.ExecuteJob(sampleImageName, envVarsForContainer)
	assert.NoError(t, err)

	typeMeta := meta_v1.TypeMeta{
		Kind:       "Job",
		APIVersion: "batch/v1",
	}

	listOptions := meta_v1.ListOptions{
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

	assert.Equal(t, config.KubeJobActiveDeadlineSeconds(), executedJob.Spec.ActiveDeadlineSeconds)

	assert.Equal(t, v1.RestartPolicyOnFailure, executedJob.Spec.Template.Spec.RestartPolicy)

	container := executedJob.Spec.Template.Spec.Containers[0]
	assert.Equal(t, executedJobname, container.Name)

	assert.Equal(t, sampleImageName, container.Image)

	expectedEnvVars := getEnvVars(envVarsForContainer)
	assert.Equal(t, expectedEnvVars, container.Env)
}

func (suite *ClientTestSuite) TestStreamLogsSuccess() {
	t := suite.T()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	namespace := config.DefaultNamespace()
	httpmock.RegisterResponder("GET", "http://"+config.KubeClusterHostName()+"/api/v1/namespaces/"+namespace+"/pods/"+suite.podName+"/log?follow=true",
		httpmock.NewStringResponder(200, "logs are streaming"))

	logStream, err := suite.testClientStreaming.StreamJobLogs(suite.jobName)
	assert.NoError(t, err)

	defer logStream.Close()

	bufioReader := bufio.NewReader(logStream)

	jobLogSingleLine, _, err := bufioReader.ReadLine()
	assert.NoError(t, err)

	assert.Equal(t, "logs are streaming", string(jobLogSingleLine[:]))

}

func (suite *ClientTestSuite) TestStreamLogsPodNotFoundFailure() {
	t := suite.T()

	_, err := suite.testClientStreaming.StreamJobLogs("unknown-job")
	assert.Error(t, err)
}

func (suite *ClientTestSuite) TestShouldReturnSuccessJobExecutionStatus() {
	t := suite.T()

	watcher := watch.NewFake()
	suite.fakeClientSet.PrependWatchReactor("jobs", testing_kubernetes.DefaultWatchReactor(watcher, nil))

	var testJob batchV1.Job
	uniqueJobName := "proctor-job-2"
	label := jobLabel(uniqueJobName)
	objectMeta := meta_v1.ObjectMeta{
		Name:   uniqueJobName,
		Labels: label,
	}
	testJob.ObjectMeta = objectMeta

	go func() {
		testJob.Status.Succeeded = 1
		watcher.Modify(&testJob)

		time.Sleep(time.Second * 1)
		watcher.Stop()
	}()

	jobExecutionStatus, err := suite.testClient.JobExecutionStatus(uniqueJobName)
	assert.NoError(t, err)

	assert.Equal(t, "SUCCEEDED", jobExecutionStatus, "Should return true for job success")
}

func (suite *ClientTestSuite) TestShouldReturnFailedJobExecutionStatus() {
	t := suite.T()

	watcher := watch.NewFake()
	suite.fakeClientSet.PrependWatchReactor("jobs", testing_kubernetes.DefaultWatchReactor(watcher, nil))

	var testJob batchV1.Job
	uniqueJobName := "proctor-job-1"
	label := jobLabel(uniqueJobName)
	objectMeta := meta_v1.ObjectMeta{
		Name:   uniqueJobName,
		Labels: label,
	}
	testJob.ObjectMeta = objectMeta

	go func() {
		testJob.Status.Active = 1
		watcher.Modify(&testJob)
		testJob.Status.Active = 0
		testJob.Status.Failed = 1
		watcher.Modify(&testJob)

		time.Sleep(time.Second * 1)
		watcher.Stop()
	}()

	jobExecutionStatus, err := suite.testClient.JobExecutionStatus(uniqueJobName)
	assert.NoError(t, err)

	assert.Equal(t, "FAILED", jobExecutionStatus, "Should return true for job success")
}

func (suite *ClientTestSuite) TestShouldReturnErrorJobExecutionStatus() {
	t := suite.T()

	watcher := watch.NewFake()
	suite.fakeClientSet.PrependWatchReactor("jobs", testing_kubernetes.DefaultWatchReactor(watcher, nil))

	var testJob batchV1.Job
	uniqueJobName := "proctor-job-3"
	label := jobLabel(uniqueJobName)
	objectMeta := meta_v1.ObjectMeta{
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

	assert.Equal(t, "FAILED", jobExecutionStatus, "Should return true for job success")
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
