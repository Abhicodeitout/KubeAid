package kubernetes

import (
	"os"
	"path/filepath"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
)

func GetKubeClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if _, err = os.Stat(kubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
