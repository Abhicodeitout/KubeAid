package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"
)

func kubeConfigPath() string {
	if fromEnv := strings.TrimSpace(os.Getenv("KUBECONFIG")); fromEnv != "" {
		// KUBECONFIG may contain multiple paths; use the first as the primary source.
		return strings.Split(fromEnv, string(os.PathListSeparator))[0]
	}
	return filepath.Join(os.Getenv("HOME"), ".kube", "config")
}

// GetKubeConfig returns the Kubernetes REST config from kubeconfig or in-cluster config.
func GetKubeConfig() (*rest.Config, error) {
	kubeconfig := kubeConfigPath()
	if _, err := os.Stat(kubeconfig); err == nil {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func GetKubeClient() (*kubernetes.Clientset, error) {
	config, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
