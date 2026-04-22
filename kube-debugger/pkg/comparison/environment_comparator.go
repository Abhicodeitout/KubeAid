package comparison

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Environment represents an environment
type Environment struct {
	Name      string
	Namespace string
	Client    kubernetes.Interface
	PodCount  int
	HealthScore int
}

// EnvironmentDifference represents a difference between environments
type EnvironmentDifference struct {
	Type     string // resource, config, performance
	Issue    string
	Env1     string
	Value1   string
	Env2     string
	Value2   string
	Severity string
}

// EnvironmentComparator compares environments
type EnvironmentComparator struct {
	environments map[string]*Environment
}

// New creates a new environment comparator
func New() *EnvironmentComparator {
	return &EnvironmentComparator{
		environments: make(map[string]*Environment),
	}
}

// AddEnvironment adds an environment
func (ec *EnvironmentComparator) AddEnvironment(name, namespace string, client kubernetes.Interface) {
	ec.environments[name] = &Environment{
		Name:      name,
		Namespace: namespace,
		Client:    client,
	}
}

// CompareResourceRequests compares resource requests across environments
func (ec *EnvironmentComparator) CompareResourceRequests(ctx context.Context, podName string) []EnvironmentDifference {
	differences := make([]EnvironmentDifference, 0)

	podConfigs := make(map[string]*corev1.Pod)

	// Get pods from all environments
	for envName, env := range ec.environments {
		pod, err := env.Client.CoreV1().Pods(env.Namespace).Get(ctx, podName, metav1.GetOptions{})
		if err == nil {
			podConfigs[envName] = pod
		}
	}

	if len(podConfigs) < 2 {
		return differences
	}

	// Compare resource requests between pairs of environments
	envNames := make([]string, 0, len(podConfigs))
	for name := range podConfigs {
		envNames = append(envNames, name)
	}

	for i := 0; i < len(envNames)-1; i++ {
		for j := i + 1; j < len(envNames); j++ {
			env1Name := envNames[i]
			env2Name := envNames[j]
			pod1 := podConfigs[env1Name]
			pod2 := podConfigs[env2Name]

			// Compare containers
			if len(pod1.Spec.Containers) > 0 && len(pod2.Spec.Containers) > 0 {
				container1 := pod1.Spec.Containers[0]
				container2 := pod2.Spec.Containers[0]

				// Compare memory requests
				mem1 := "none"
				if container1.Resources.Requests != nil && container1.Resources.Requests.Memory() != nil {
					mem1 = container1.Resources.Requests.Memory().String()
				}

				mem2 := "none"
				if container2.Resources.Requests != nil && container2.Resources.Requests.Memory() != nil {
					mem2 = container2.Resources.Requests.Memory().String()
				}

				if mem1 != mem2 {
					differences = append(differences, EnvironmentDifference{
						Type:     "resource",
						Issue:    "Memory request mismatch",
						Env1:     env1Name,
						Value1:   mem1,
						Env2:     env2Name,
						Value2:   mem2,
						Severity: "medium",
					})
				}

				// Compare CPU requests
				cpu1 := "none"
				if container1.Resources.Requests != nil && container1.Resources.Requests.Cpu() != nil {
					cpu1 = container1.Resources.Requests.Cpu().String()
				}

				cpu2 := "none"
				if container2.Resources.Requests != nil && container2.Resources.Requests.Cpu() != nil {
					cpu2 = container2.Resources.Requests.Cpu().String()
				}

				if cpu1 != cpu2 {
					differences = append(differences, EnvironmentDifference{
						Type:     "resource",
						Issue:    "CPU request mismatch",
						Env1:     env1Name,
						Value1:   cpu1,
						Env2:     env2Name,
						Value2:   cpu2,
						Severity: "medium",
					})
				}
			}
		}
	}

	return differences
}

// CompareConfig compares pod configs across environments
func (ec *EnvironmentComparator) CompareConfig(ctx context.Context, podName string) []EnvironmentDifference {
	differences := make([]EnvironmentDifference, 0)

	podConfigs := make(map[string]*corev1.Pod)

	for envName, env := range ec.environments {
		pod, err := env.Client.CoreV1().Pods(env.Namespace).Get(ctx, podName, metav1.GetOptions{})
		if err == nil {
			podConfigs[envName] = pod
		}
	}

	// Compare image versions
	if len(podConfigs) >= 2 {
		envNames := make([]string, 0, len(podConfigs))
		for name := range podConfigs {
			envNames = append(envNames, name)
		}

		for i := 0; i < len(envNames)-1; i++ {
			for j := i + 1; j < len(envNames); j++ {
				env1Name := envNames[i]
				env2Name := envNames[j]
				pod1 := podConfigs[env1Name]
				pod2 := podConfigs[env2Name]

				if len(pod1.Spec.Containers) > 0 && len(pod2.Spec.Containers) > 0 {
					img1 := pod1.Spec.Containers[0].Image
					img2 := pod2.Spec.Containers[0].Image

					if img1 != img2 {
						differences = append(differences, EnvironmentDifference{
							Type:     "config",
							Issue:    "Container image mismatch",
							Env1:     env1Name,
							Value1:   img1,
							Env2:     env2Name,
							Value2:   img2,
							Severity: "high",
						})
					}
				}
			}
		}
	}

	return differences
}

// ComparePodCounts compares pod counts across environments
func (ec *EnvironmentComparator) ComparePodCounts(ctx context.Context, appLabel string) map[string]int {
	counts := make(map[string]int)

	for envName, env := range ec.environments {
		pods, err := env.Client.CoreV1().Pods(env.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", appLabel),
		})
		if err == nil {
			counts[envName] = len(pods.Items)
		}
	}

	return counts
}

// DetectDrift detects configuration drift between environments
func (ec *EnvironmentComparator) DetectDrift(ctx context.Context, podName string) []string {
	driftIssues := make([]string, 0)

	resourceDiffs := ec.CompareResourceRequests(ctx, podName)
	if len(resourceDiffs) > 0 {
		driftIssues = append(driftIssues, fmt.Sprintf("%d resource request differences detected", len(resourceDiffs)))
	}

	configDiffs := ec.CompareConfig(ctx, podName)
	if len(configDiffs) > 0 {
		driftIssues = append(driftIssues, fmt.Sprintf("%d config differences detected", len(configDiffs)))
	}

	return driftIssues
}

// GetComparisonSummary returns a summary of environment comparison
func (ec *EnvironmentComparator) GetComparisonSummary(ctx context.Context, podName string) map[string]interface{} {
	summary := make(map[string]interface{})

	resourceDiffs := ec.CompareResourceRequests(ctx, podName)
	configDiffs := ec.CompareConfig(ctx, podName)

	summary["pod_name"] = podName
	summary["resource_differences"] = len(resourceDiffs)
	summary["config_differences"] = len(configDiffs)
	summary["total_differences"] = len(resourceDiffs) + len(configDiffs)
	summary["environments_compared"] = len(ec.environments)
	summary["has_drift"] = (len(resourceDiffs) + len(configDiffs)) > 0

	return summary
}
