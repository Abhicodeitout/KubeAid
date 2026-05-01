package kubernetes

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

// PodEvent is a timestamped event for a pod.
type PodEvent struct {
	Time    time.Time
	Type    string
	Reason  string
	Message string
	Count   int32
}

// ListPodEvents returns timestamped pod events sorted by time.
func ListPodEvents(clientset *kubernetes.Clientset, namespace, podName string) ([]PodEvent, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.kind", "Pod"),
		fields.OneTermEqualSelector("involvedObject.name", podName),
	).String()
	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: selector})
	if err != nil {
		return nil, err
	}

	result := make([]PodEvent, 0, len(events.Items))
	for _, event := range events.Items {
		if event.InvolvedObject.Name != podName {
			continue
		}

		timestamp := event.LastTimestamp.Time
		if timestamp.IsZero() {
			timestamp = event.EventTime.Time
		}
		if timestamp.IsZero() {
			timestamp = event.FirstTimestamp.Time
		}
		if timestamp.IsZero() {
			timestamp = event.CreationTimestamp.Time
		}

		result = append(result, PodEvent{
			Time:    timestamp,
			Type:    event.Type,
			Reason:  event.Reason,
			Message: event.Message,
			Count:   event.Count,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Time.Before(result[j].Time)
	})

	return result, nil
}

func GetPodEvents(clientset *kubernetes.Clientset, namespace, podName string) (string, error) {
	events, err := ListPodEvents(clientset, namespace, podName)
	if err != nil {
		return "", err
	}
	var lines []string
	for _, event := range events {
		prefix := event.Reason
		if !event.Time.IsZero() {
			prefix = fmt.Sprintf("%s %s", event.Time.UTC().Format(time.RFC3339), prefix)
		}
		if event.Type != "" {
			prefix = fmt.Sprintf("%s [%s]", prefix, event.Type)
		}
		lines = append(lines, fmt.Sprintf("%s: %s", prefix, event.Message))
	}
	if len(lines) == 0 {
		return "", nil
	}
	return strings.Join(lines, "\n") + "\n", nil
}
