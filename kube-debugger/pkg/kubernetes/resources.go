package kubernetes

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
)

func GetPodResourceUsage(clientset *kubernetes.Clientset, namespace, podName string) (string, error) {
	// Placeholder: In real use, integrate with metrics-server or custom metrics API
	return fmt.Sprintf("CPU: 90%% (High)\nMemory: 85%%"), nil
}
