package kubernetes

import (
	"context"
	"io"
	"fmt"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/api/core/v1"
)

// LogLine is a timestamped pod log line.
type LogLine struct {
	Time    time.Time
	Message string
}

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

// GetTimestampedPodLogs fetches recent logs with Kubernetes-provided timestamps.
func GetTimestampedPodLogs(clientset *kubernetes.Clientset, namespace, podName string, tailLines int64) ([]LogLine, error) {
	podLogOpts := &corev1.PodLogOptions{TailLines: &tailLines, Timestamps: true}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}
	defer func() { _ = podLogs.Close() }()

	buf := new(strings.Builder)
	if _, err := io.Copy(buf, podLogs); err != nil {
		return nil, err
	}

	var lines []LogLine
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, parts[0])
		if err != nil {
			parsed, fallbackErr := time.Parse(time.RFC3339, parts[0])
			if fallbackErr != nil {
				continue
			}
			ts = parsed
		}
		lines = append(lines, LogLine{Time: ts, Message: parts[1]})
	}

	if len(lines) == 0 && strings.TrimSpace(buf.String()) != "" {
		return nil, fmt.Errorf("timestamped pod logs were unavailable for pod %s", podName)
	}

	return lines, nil
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
