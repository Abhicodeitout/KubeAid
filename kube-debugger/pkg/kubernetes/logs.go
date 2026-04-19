package kubernetes

import (
	"context"
	"io"
	"strings"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/api/core/v1"
)

func GetPodLogs(clientset *kubernetes.Clientset, namespace, podName string) (string, error) {
	podLogOpts := &corev1.PodLogOptions{TailLines: int64Ptr(20)}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer func() { _ = podLogs.Close() }()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetPodPreviousLogs fetches logs from the previous (terminated) container instance.
func GetPodPreviousLogs(clientset *kubernetes.Clientset, namespace, podName, containerName string) (string, error) {
	prev := true
	podLogOpts := &corev1.PodLogOptions{
		Previous:  prev,
		Container: containerName,
		TailLines: int64Ptr(50),
	}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer func() { _ = podLogs.Close() }()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	return buf.String(), err
}

func int64Ptr(i int64) *int64 { return &i }
