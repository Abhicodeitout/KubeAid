package kubernetes

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPodsByApp returns pods matching the app label
import "k8s.io/client-go/kubernetes"
func GetPodsByApp(clientset *kubernetes.Clientset, namespace, appName string) ([]string, error) {
       pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
	       LabelSelector: fmt.Sprintf("app=%s", appName),
       })
       if err != nil {
	       return nil, err
       }
       var podNames []string
       for _, pod := range pods.Items {
	       podNames = append(podNames, pod.Name)
       }
       return podNames, nil
}
