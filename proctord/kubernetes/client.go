package kubernetes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"proctor/proctord/config"
	"proctor/proctord/logger"
	"proctor/proctord/utility"

	uuid "github.com/satori/go.uuid"
	batch_v1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	//Package needed for kubernetes cluster in google cloud
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var typeMeta meta_v1.TypeMeta
var namespace string

func init() {
	typeMeta = meta_v1.TypeMeta{
		Kind:       "Job",
		APIVersion: "batch/v1",
	}
	namespace = config.DefaultNamespace()
}

type client struct {
	clientSet  kubernetes.Interface
	httpClient *http.Client
}

type Client interface {
	ExecuteJob(string, map[string]string) (string, error)
	StreamJobLogs(string) (io.ReadCloser, error)
	JobExecutionStatus(string) (string, error)
}

func NewClient(kubeconfig string, httpClient *http.Client) Client {
	newClient := &client{
		httpClient: httpClient,
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	newClient.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return newClient
}

func getEnvVars(envMap map[string]string) []v1.EnvVar {
	var envVars []v1.EnvVar
	for k, v := range envMap {
		envVar := v1.EnvVar{
			Name:  k,
			Value: v,
		}
		envVars = append(envVars, envVar)
	}
	return envVars
}

func uniqueName() string {
	return "proctor" + "-" + uuid.NewV4().String()
}

func jobLabel(jobName string) map[string]string {
	return map[string]string{
		"job": jobName,
	}
}

func jobLabelSelector(jobName string) string {
	return fmt.Sprintf("job=%s", jobName)
}

func (client *client) ExecuteJob(imageName string, envMap map[string]string) (string, error) {
	uniqueJobName := uniqueName()
	label := jobLabel(uniqueJobName)

	batchV1 := client.clientSet.BatchV1()
	kubernetesJobs := batchV1.Jobs(namespace)

	container := v1.Container{
		Name:  uniqueJobName,
		Image: imageName,
		Env:   getEnvVars(envMap),
	}

	podSpec := v1.PodSpec{
		Containers:    []v1.Container{container},
		RestartPolicy: v1.RestartPolicyNever,
	}

	objectMeta := meta_v1.ObjectMeta{
		Name:        uniqueJobName,
		Labels:      label,
		Annotations: config.JobPodAnnotations(),
	}

	template := v1.PodTemplateSpec{
		ObjectMeta: objectMeta,
		Spec:       podSpec,
	}

	jobSpec := batch_v1.JobSpec{
		Template:              template,
		ActiveDeadlineSeconds: config.KubeJobActiveDeadlineSeconds(),
		BackoffLimit:          config.KubeJobRetries(),
	}

	jobToRun := batch_v1.Job{
		TypeMeta:   typeMeta,
		ObjectMeta: objectMeta,
		Spec:       jobSpec,
	}

	_, err := kubernetesJobs.Create(context.Background(), &jobToRun, meta_v1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return uniqueJobName, nil
}

func (client *client) StreamJobLogs(jobName string) (io.ReadCloser, error) {
	listOptions := meta_v1.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(jobName),
	}

	coreV1 := client.clientSet.CoreV1()
	kubernetesPods := coreV1.Pods(namespace)

	logger.Debug("list of pods")

	for {
		listOfPods, err := kubernetesPods.List(context.Background(), listOptions)
		if err != nil {
			return nil, fmt.Errorf("Error fetching kubernetes Pods list %v", err)
		}

		if len(listOfPods.Items) > 0 {
			podJob := listOfPods.Items[0]
			if podJob.Status.Phase == v1.PodRunning || podJob.Status.Phase == v1.PodSucceeded || podJob.Status.Phase == v1.PodFailed {
				return client.getLogsStreamReaderFor(podJob.ObjectMeta.Name)
			}
			watchPod, err := kubernetesPods.Watch(context.Background(), listOptions)
			if err != nil {
				return nil, fmt.Errorf("Error watching kubernetes Pods %v", err)
			}

			resultChan := watchPod.ResultChan()
			defer watchPod.Stop()

			waitingForKubePods := make(chan bool)
			go func() {
				defer close(waitingForKubePods)
				time.Sleep(time.Duration(config.KubePodsListWaitTime()) * time.Second)
				waitingForKubePods <- true
			}()

			select {
			case <-resultChan:
				continue
			case <-waitingForKubePods:
				return nil, fmt.Errorf("Pod didn't reach active state after waiting for %d minutes", config.KubePodsListWaitTime())
			}

		} else {
			batchV1 := client.clientSet.BatchV1()
			kubernetesJobs := batchV1.Jobs(namespace)

			watchJob, err := kubernetesJobs.Watch(context.Background(), listOptions)
			if err != nil {
				return nil, fmt.Errorf("Error watching kubernetes Jobs %v", err)
			}

			resultChan := watchJob.ResultChan()
			defer watchJob.Stop()

			waitingForKubeJobs := make(chan bool)
			go func() {
				defer close(waitingForKubeJobs)
				time.Sleep(time.Duration(config.KubePodsListWaitTime()) * time.Second)
				waitingForKubeJobs <- true
			}()

			select {
			case <-resultChan:
				continue
			case <-waitingForKubeJobs:
				return nil, fmt.Errorf("Couldn't find a pod for job's given list options %v after waiting for %d minutes", listOptions, config.KubePodsListWaitTime())
			}
		}
	}
}

func (client *client) JobExecutionStatus(jobExecutionID string) (string, error) {
	batchV1 := client.clientSet.BatchV1()
	kubernetesJobs := batchV1.Jobs(namespace)
	listOptions := meta_v1.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(jobExecutionID),
	}

	watchJob, err := kubernetesJobs.Watch(context.Background(), listOptions)
	if err != nil {
		return utility.JobFailed, err
	}

	resultChan := watchJob.ResultChan()
	defer watchJob.Stop()
	var event watch.Event
	var jobEvent *batch_v1.Job

	for event = range resultChan {
		if event.Type == watch.Error {
			return utility.JobExecutionStatusFetchError, nil
		}

		jobEvent = event.Object.(*batch_v1.Job)
		if jobEvent.Status.Succeeded >= int32(1) {
			return utility.JobSucceeded, nil
		} else if jobEvent.Status.Failed >= int32(1) {
			return utility.JobFailed, nil
		}
	}

	return utility.NoDefinitiveJobExecutionStatusFound, nil
}

func (client *client) getLogsStreamReaderFor(podName string) (io.ReadCloser, error) {
	logger.Debug("reading pod logs for: ", podName)

	// Use the authenticated client instead of manually requesting the control plane
	clt := client.clientSet.CoreV1()
	req := clt.Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Follow: true,
	})
	logs, err := req.Stream(context.Background())
	if err != nil {
		return nil, err
	}
	return logs, err
}
