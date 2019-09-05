package kubernetes

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
)

var typeMeta meta.TypeMeta
var namespace string
var timeoutError = errors.New("timeout when waiting job to be available")

func init() {
	typeMeta = meta.TypeMeta{
		Kind:       "Job",
		APIVersion: "batch/v1",
	}
	namespace = config.DefaultNamespace()
}

type KubernetesClient interface {
	ExecuteJobWithCommand(imageName string, args map[string]string, commands []string) (string, error)
	ExecuteJob(imageName string, args map[string]string) (string, error)
	JobExecutionStatus(executionName string) (string, error)
	WaitForReadyJob(executionName string, waitTime time.Duration) error
	WaitForReadyPod(executionName string, waitTime time.Duration) (*v1.Pod, error)
	GetPodLogs(pod *v1.Pod) (io.ReadCloser, error)
}

type kubernetesClient struct {
	clientSet  kubernetes.Interface
	httpClient *http.Client
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
	newClient := &kubernetesClient{
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

func jobLabel(executionName string) map[string]string {
	return map[string]string{
		"job": executionName,
	}
}

func jobLabelSelector(executionName string) string {
	return fmt.Sprintf("job=%s", executionName)
}

func (client *kubernetesClient) ExecuteJob(imageName string, envMap map[string]string) (string, error) {
	return client.ExecuteJobWithCommand(imageName, envMap, []string{})
}

func (client *kubernetesClient) ExecuteJobWithCommand(imageName string, envMap map[string]string, command []string) (string, error) {
	executionName := uniqueName()
	label := jobLabel(executionName)

	batchV1 := client.clientSet.BatchV1()
	kubernetesJobs := batchV1.Jobs(namespace)

	container := v1.Container{
		Name:  executionName,
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
		Name:        executionName,
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
	return executionName, nil
}

func (client *kubernetesClient) WaitForReadyJob(executionName string, waitTime time.Duration) error {
	batchV1 := client.clientSet.BatchV1()
	jobs := batchV1.Jobs(namespace)
	listOptions := meta.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(executionName),
	}

	var err error
	for i := 0; i < config.KubeWaitForResourcePollCount(); i += 1 {
		watchJob, watchErr := jobs.Watch(listOptions)
		if watchErr != nil {
			err = watchErr
			continue
		}

		timeoutChan := time.After(waitTime)
		resultChan := watchJob.ResultChan()

		var job *batch.Job
		for {
			select {
			case event := <-resultChan:
				if event.Type == watch.Error {
					err = watcherError("job", listOptions)
					break
				}

				// Ignore empty events
				if event.Object == nil {
					continue
				}

				job = event.Object.(*batch.Job)
				if job.Status.Active >= 1 || job.Status.Succeeded >= 1 || job.Status.Failed >= 1 {
					watchJob.Stop()
					return nil
				}
			case <-timeoutChan:
				err = timeoutError
				break
			}
			if err != nil {
				watchJob.Stop()
				break
			}
		}
	}

	return err
}

func (client *kubernetesClient) WaitForReadyPod(executionName string, waitTime time.Duration) (*v1.Pod, error) {
	coreV1 := client.clientSet.CoreV1()
	kubernetesPods := coreV1.Pods(namespace)
	listOptions := meta.ListOptions{
		LabelSelector: jobLabelSelector(executionName),
	}

	var err error
	for i := 0; i < config.KubeWaitForResourcePollCount(); i += 1 {
		watchJob, watchErr := kubernetesPods.Watch(listOptions)
		if watchErr != nil {
			err = watchErr
			continue
		}

		timeoutChan := time.After(waitTime)
		resultChan := watchJob.ResultChan()

		var pod *v1.Pod
		for {
			select {
			case event := <-resultChan:
				if event.Type == watch.Error {
					err = watcherError("pod", listOptions)
					watchJob.Stop()
					break
				}

				// Ignore empty events
				if event.Object == nil {
					continue
				}

				pod = event.Object.(*v1.Pod)
				if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
					watchJob.Stop()
					return pod, nil
				}
			case <-timeoutChan:
				err = timeoutError
				watchJob.Stop()
				break
			}
			if err != nil {
				watchJob.Stop()
				break
			}
		}
	}

	return nil, err
}

func (client *kubernetesClient) JobExecutionStatus(executionName string) (string, error) {
	batchV1 := client.clientSet.BatchV1()
	kubernetesJobs := batchV1.Jobs(namespace)
	listOptions := meta.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(executionName),
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

func (client *kubernetesClient) GetPodLogs(pod *v1.Pod) (io.ReadCloser, error) {
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

func watcherError(resource string, listOptions meta.ListOptions) error {
	return fmt.Errorf("watch error when waiting for %s with list option %v", resource, listOptions)
}
