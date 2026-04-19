package kubernetes

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

func GetPodResourceUsage(clientset *kubernetes.Clientset, namespace, podName string) (string, error) {
	_ = clientset

	config, err := GetKubeConfig()
	if err != nil {
		return "Resource metrics unavailable: kubeconfig not accessible", nil
	}

	mClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return "Resource metrics unavailable: metrics client initialization failed", nil
	}

	metrics, err := mClient.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "Resource metrics unavailable: metrics-server is missing or no samples yet", nil
		}
		return "Resource metrics unavailable: unable to query metrics API", nil
	}

	if len(metrics.Containers) == 0 {
		return "Resource metrics unavailable: no container metrics returned", nil
	}

	totalMilliCPU := int64(0)
	totalMemoryBytes := int64(0)
	perContainer := make([]string, 0, len(metrics.Containers))

	for _, c := range metrics.Containers {
		cpu := c.Usage.Cpu().MilliValue()
		mem := c.Usage.Memory().Value()
		totalMilliCPU += cpu
		totalMemoryBytes += mem
		perContainer = append(perContainer, fmt.Sprintf("%s: CPU %dm, Memory %s", c.Name, cpu, c.Usage.Memory().String()))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("CPU: %dm\n", totalMilliCPU))
	b.WriteString(fmt.Sprintf("Memory: %s\n", formatBytesBinary(totalMemoryBytes)))
	b.WriteString("Container Breakdown:\n")
	for _, line := range perContainer {
		b.WriteString("- " + line + "\n")
	}

	return strings.TrimSuffix(b.String(), "\n"), nil
}

func formatBytesBinary(bytes int64) string {
	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
	)

	switch {
	case bytes >= GiB:
		return fmt.Sprintf("%.1fGi", float64(bytes)/float64(GiB))
	case bytes >= MiB:
		return fmt.Sprintf("%.1fMi", float64(bytes)/float64(MiB))
	case bytes >= KiB:
		return fmt.Sprintf("%.1fKi", float64(bytes)/float64(KiB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
