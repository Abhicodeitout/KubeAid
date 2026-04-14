package kubernetes

import (
	"os"
	"path/filepath"
	"k8s.io/client-go/tools/clientcmd"
)

// ListKubeContexts returns all available kubeconfig contexts
func ListKubeContexts() ([]string, error) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	var contexts []string
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}
	return contexts, nil
}

// SwitchKubeContext switches the current context in kubeconfig
func SwitchKubeContext(contextName string) error {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return err
	}
	config.CurrentContext = contextName
	return clientcmd.WriteToFile(*config, kubeconfig)
}
