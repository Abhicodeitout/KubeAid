package kubernetes

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPodEvents(clientset *kubernetes.Clientset, namespace, podName string) (string, error) {
	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	var result string
	for _, event := range events.Items {
		if event.InvolvedObject.Name == podName {
			result += fmt.Sprintf("%s: %s\n", event.Reason, event.Message)
		}
	}
	return result, nil
}
