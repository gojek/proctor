package kubernetes

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"time"

	uuid "github.com/satori/go.uuid"
	batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"proctor/internal/pkg/constant"
	//Package needed for kubernetes cluster in google cloud
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	kubeRestClient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var typeMeta meta.TypeMeta
var namespace string

func init() {
	typeMeta = meta.TypeMeta{
		Kind:       "Job",
		APIVersion: "batch/v1",
	}
	namespace = config.DefaultNamespace()
}

type client struct {
	clientSet  kubernetes.Interface
	httpClient *http.Client
}

type KubernetesClient interface {
	ExecuteJobWithCommand(string, map[string]string, []string) (string, error)
	ExecuteJob(string, map[string]string) (string, error)
	StreamJobLogs(string) (io.ReadCloser, error)
	JobExecutionStatus(string) (string, error)
}

func NewClientSet() (*kubernetes.Clientset, error) {
	var kubeConfig *kubeRestClient.Config
	if config.KubeConfig() == "out-of-cluster" {
		logger.Info("service is running outside kube cluster")
		home := os.Getenv("HOME")

		kubeConfigPath := filepath.Join(home, ".kube", "config")

		configOverrides := &clientcmd.ConfigOverrides{}
		if config.KubeContext() != "default" {
			configOverrides.CurrentContext = config.KubeContext()
		}

		var err error
		kubeConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
			configOverrides).ClientConfig()
		if err != nil {
			return nil, err
		}

	} else {
		var err error
		kubeConfig, err = kubeRestClient.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func NewKubernetesClient(httpClient *http.Client) KubernetesClient {
	newClient := &client{
		httpClient: httpClient,
	}

	var err error
	newClient.clientSet, err = NewClientSet()
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
	return client.ExecuteJobWithCommand(imageName, envMap, []string{})
}

func (client *client) ExecuteJobWithCommand(imageName string, envMap map[string]string, command []string) (string, error) {
	uniqueJobName := uniqueName()
	label := jobLabel(uniqueJobName)

	batchV1 := client.clientSet.BatchV1()
	kubernetesJobs := batchV1.Jobs(namespace)

	container := v1.Container{
		Name:  uniqueJobName,
		Image: imageName,
		Env:   getEnvVars(envMap),
	}

	if len(command) != 0 {
		container.Command = command
	}

	podSpec := v1.PodSpec{
		Containers:    []v1.Container{container},
		RestartPolicy: v1.RestartPolicyNever,
	}

	objectMeta := meta.ObjectMeta{
		Name:        uniqueJobName,
		Labels:      label,
		Annotations: config.JobPodAnnotations(),
	}

	template := v1.PodTemplateSpec{
		ObjectMeta: objectMeta,
		Spec:       podSpec,
	}

	jobSpec := batch.JobSpec{
		Template:              template,
		ActiveDeadlineSeconds: config.KubeJobActiveDeadlineSeconds(),
		BackoffLimit:          config.KubeJobRetries(),
	}

	jobToRun := batch.Job{
		TypeMeta:   typeMeta,
		ObjectMeta: objectMeta,
		Spec:       jobSpec,
	}

	_, err := kubernetesJobs.Create(&jobToRun)
	if err != nil {
		return "", err
	}
	return uniqueJobName, nil
}

func (client *client) StreamJobLogs(jobName string) (io.ReadCloser, error) {
	err := client.waitForReadyJob(jobName)
	if err != nil {
		return nil, err
	}

	pod, err := client.waitForReadyPod(jobName)
	if err != nil {
		return nil, err
	}

	result, err := client.getPodLogs(pod)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (client *client) waitForReadyJob(jobName string) error {
	batchV1 := client.clientSet.BatchV1()
	jobs := batchV1.Jobs(namespace)
	listOptions := meta.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(jobName),
	}

	watchJob, err := jobs.Watch(listOptions)
	if err != nil {
		return err
	}

	timeoutChan := time.After(config.KubePodsListWaitTime() * time.Second)
	resultChan := watchJob.ResultChan()
	defer watchJob.Stop()

	select {
	case <-timeoutChan:
		return fmt.Errorf("timeout when waiting job to be available")
	case <-resultChan:
		for event := range resultChan {
			if event.Type == watch.Error {
				return fmt.Errorf("watch error when waiting for job with list option %v", listOptions)
			}
			job := event.Object.(*batch.Job)
			if job.Status.Active >= 1 || job.Status.Succeeded >= 1 || job.Status.Failed >= 1 {
				return nil
			}

			select {
			case <-timeoutChan:
				return fmt.Errorf("timeout when waiting job to be ready")
			case <-resultChan:
				continue
			}
		}
	}

	return fmt.Errorf("job never reach the active status")
}

func (client *client) waitForReadyPod(jobName string) (*v1.Pod, error) {
	coreV1 := client.clientSet.CoreV1()
	kubernetesPods := coreV1.Pods(namespace)
	listOptions := meta.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(jobName),
	}

	watchJob, err := kubernetesPods.Watch(listOptions)
	if err != nil {
		return nil, err
	}

	timeoutChan := time.After(config.KubePodsListWaitTime() * time.Second)
	resultChan := watchJob.ResultChan()
	defer watchJob.Stop()
	var pod *v1.Pod

	select {
	case <-timeoutChan:
		return nil,fmt.Errorf("timeout when waiting pod to be available")
	case <-resultChan:
		for event := range resultChan {
			if event.Type == watch.Error {
				return nil, fmt.Errorf("watch error when waiting for pod with list option %v", listOptions)
			}
			pod = event.Object.(*v1.Pod)
			if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
				return pod, nil
			}
			select {
			case <-timeoutChan:
				return nil, fmt.Errorf("timeout when waiting pod to be ready")
			case <-resultChan:
				continue
			}
		}
	}

	return nil, fmt.Errorf("pod never get the intended state")
}

func (client *client) JobExecutionStatus(jobName string) (string, error) {
	batchV1 := client.clientSet.BatchV1()
	kubernetesJobs := batchV1.Jobs(namespace)
	listOptions := meta.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(jobName),
	}

	watchJob, err := kubernetesJobs.Watch(listOptions)
	if err != nil {
		return constant.JobFailed, err
	}

	resultChan := watchJob.ResultChan()
	defer watchJob.Stop()
	var event watch.Event
	var jobEvent *batch.Job

	for event = range resultChan {
		if event.Type == watch.Error {
			return constant.JobExecutionStatusFetchError, nil
		}

		jobEvent = event.Object.(*batch.Job)
		if jobEvent.Status.Succeeded >= int32(1) {
			return constant.JobSucceeded, nil
		} else if jobEvent.Status.Failed >= int32(1) {
			return constant.JobFailed, nil
		}
	}

	return constant.NoDefinitiveJobExecutionStatusFound, nil
}

func (client *client) getPodLogs(pod *v1.Pod) (io.ReadCloser, error) {
	logger.Debug("reading pod logs for: ", pod.Name)
	podLogOpts := v1.PodLogOptions{
		Follow: true,
	}
	request := client.clientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	response, err := request.Stream()

	if err != nil {
		return nil, err
	}
	return response, nil
}
