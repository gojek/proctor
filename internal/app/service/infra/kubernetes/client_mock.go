package kubernetes

import (
	"io"
	v1 "k8s.io/api/core/v1"
	"time"

	"github.com/stretchr/testify/mock"
	"proctor/internal/pkg/utility"
)

type MockKubernetesClient struct {
	mock.Mock
}

func (m *MockKubernetesClient) ExecuteJob(jobName string, envMap map[string]string) (string, error) {
	args := m.Called(jobName, envMap)
	return args.String(0), args.Error(1)
}

func (m *MockKubernetesClient) ExecuteJobWithCommand(jobName string, envMap map[string]string, command []string) (string, error) {
	args := m.Called(jobName, envMap, command)
	return args.String(0), args.Error(1)
}

func (m *MockKubernetesClient) StreamJobLogs(executionName string, waitTime time.Duration) (io.ReadCloser, error) {
	args := m.Called(executionName, waitTime)
	return args.Get(0).(*utility.Buffer), args.Error(1)
}

func (m *MockKubernetesClient) JobExecutionStatus(executionName string) (string, error) {
	args := m.Called(executionName)
	return args.String(0), args.Error(1)
}

func (m *MockKubernetesClient) WaitForReadyJob(executionName string, waitTime time.Duration) error {
	args := m.Called(executionName, waitTime)
	return args.Error(0)
}
func (m *MockKubernetesClient) WaitForReadyPod(executionName string, waitTime time.Duration) (*v1.Pod, error) {
	args := m.Called(executionName)
	return args.Get(0).(*v1.Pod), args.Error(1)
}
func (m *MockKubernetesClient) GetPodLogs(pod *v1.Pod) (io.ReadCloser, error) {
	args := m.Called(pod)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
