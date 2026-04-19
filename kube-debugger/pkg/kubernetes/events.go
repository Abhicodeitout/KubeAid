package kubernetes

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

func GetPodEvents(clientset *kubernetes.Clientset, namespace, podName string) (string, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.kind", "Pod"),
		fields.OneTermEqualSelector("involvedObject.name", podName),
	).String()
	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: selector})
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
