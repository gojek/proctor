package kubernetes

import (
	"fmt"
	"io"
	"time"

	"net/http"

	"github.com/gojektech/proctor/proctord/config"
	"github.com/gojektech/proctor/proctord/logger"
	uuid "github.com/satori/go.uuid"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	batch_v1 "k8s.io/client-go/pkg/apis/batch/v1"
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
	clientSet kubernetes.Interface
}

type Client interface {
	ExecuteJob(string, map[string]string) (string, error)
	StreamJobLogs(string) (io.ReadCloser, error)
}

func NewClient(kubeconfig string) Client {
	var newClient client

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	newClient.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return &newClient
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
		RestartPolicy: v1.RestartPolicyOnFailure,
	}

	objectMeta := meta_v1.ObjectMeta{
		Name:   uniqueJobName,
		Labels: label,
	}

	template := v1.PodTemplateSpec{
		ObjectMeta: objectMeta,
		Spec:       podSpec,
	}

	jobSpec := batch_v1.JobSpec{
		Template:              template,
		ActiveDeadlineSeconds: config.KubeJobActiveDeadlineSeconds(),
	}

	jobToRun := batch_v1.Job{
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
	listOptions := meta_v1.ListOptions{
		TypeMeta:      typeMeta,
		LabelSelector: jobLabelSelector(jobName),
	}

	coreV1 := client.clientSet.CoreV1()
	kubernetesPods := coreV1.Pods(namespace)

	logger.Debug("list of pods")

	for {
		listOfPods, err := kubernetesPods.List(listOptions)
		if err != nil {
			return nil, fmt.Errorf("Error fetching kubernetes Pods list %v", err)
		}

		if len(listOfPods.Items) > 0 {
			podJob := listOfPods.Items[0]
			if podJob.Status.Phase == v1.PodRunning || podJob.Status.Phase == v1.PodSucceeded || podJob.Status.Phase == v1.PodFailed {
				return getLogsStreamReaderFor(podJob.ObjectMeta.Name)
			}
			watchPod, err := kubernetesPods.Watch(listOptions)
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

			watchJob, err := kubernetesJobs.Watch(listOptions)
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

func getLogsStreamReaderFor(podName string) (io.ReadCloser, error) {
	logger.Debug("reading pod logs for: ", podName)
	resp, err := http.Get("http://" + config.KubeClusterHostName() + "/api/v1/namespaces/" + namespace + "/pods/" + podName + "/log?follow=true")
	return resp.Body, err
}
