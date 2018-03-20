package kubernetes

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/gojektech/proctor/proctord/config"
	"github.com/gojektech/proctor/proctord/logger"
)

func KubeConfig() string {
	if config.KubeConfig() == "out-of-cluster" {
		logger.Info("service is running outside kube cluster")
		home := os.Getenv("HOME")

		kubeConfig := flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()

		return *kubeConfig
	}
	logger.Info("Assuming service is running inside kube cluster")
	logger.Info("Kube config provided is:", config.KubeConfig())
	return ""
}
