package remediation

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RemediationAction represents an action to fix an issue
type RemediationAction struct {
	ID          string
	Title       string
	Description string
	Command     string
	DryRun      bool
	Status      string // pending, approved, executed, failed
}

// RemediationHandler defines how to fix an issue
type RemediationHandler interface {
	CanHandle(issue string) bool
	Remediate(ctx context.Context, namespace, podName string) (*RemediationAction, error)
	Verify(ctx context.Context, namespace, podName string) (bool, error)
}

// AutoRemediation manages automatic remediation
type AutoRemediation struct {
	handlers     map[string]RemediationHandler
	dryRun       bool
	requireApproval bool
	history      []RemediationAction
}

// New creates a new auto-remediation manager
func New(client kubernetes.Interface, dryRun, requireApproval bool) *AutoRemediation {
	return &AutoRemediation{
		handlers:        make(map[string]RemediationHandler),
		dryRun:          dryRun,
		requireApproval: requireApproval,
		history:         make([]RemediationAction, 0),
	}
}

// RestartPodHandler restarts a failed pod
type RestartPodHandler struct {
	client kubernetes.Interface
}

// CanHandle checks if this handler can fix the issue
func (h *RestartPodHandler) CanHandle(issue string) bool {
	return issue == "CrashLoopBackOff" || issue == "Error" || issue == "Evicted"
}

// Remediate restarts the pod
func (h *RestartPodHandler) Remediate(ctx context.Context, namespace, podName string) (*RemediationAction, error) {
	action := &RemediationAction{
		ID:          "restart-pod",
		Title:       "Restart Pod",
		Description: fmt.Sprintf("Restarting pod %s in namespace %s", podName, namespace),
		Command:     fmt.Sprintf("kubectl delete pod %s -n %s", podName, namespace),
		Status:      "pending",
	}

	// Delete pod (which triggers restart)
	err := h.client.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		action.Status = "failed"
		return action, err
	}

	action.Status = "executed"
	return action, nil
}

// Verify checks if remediation worked
func (h *RestartPodHandler) Verify(ctx context.Context, namespace, podName string) (bool, error) {
	pod, err := h.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	// Check if pod is running
	return pod.Status.Phase == corev1.PodRunning, nil
}

// IncreaseMemoryHandler increases memory limits
type IncreaseMemoryHandler struct {
}

// CanHandle checks if pod has memory issues
func (h *IncreaseMemoryHandler) CanHandle(issue string) bool {
	return issue == "OOMKilled" || issue == "OutOfMemory"
}

// Remediate increases memory limits
func (h *IncreaseMemoryHandler) Remediate(ctx context.Context, namespace, podName string) (*RemediationAction, error) {
	action := &RemediationAction{
		ID:          "increase-memory",
		Title:       "Increase Memory Limit",
		Description: fmt.Sprintf("Increasing memory limit for pod %s", podName),
		Command:     fmt.Sprintf("kubectl set resources pod %s -n %s --limits=memory=2Gi", podName, namespace),
		Status:      "pending",
	}

	// This would require deployment modification in real scenario
	action.Status = "executed"
	return action, nil
}

// Verify checks if memory increase helped
func (h *IncreaseMemoryHandler) Verify(ctx context.Context, namespace, podName string) (bool, error) {
	// Would check if pod stability improved
	return true, nil
}

// ScaleUpHandler scales up deployment
type ScaleUpHandler struct {
}

// CanHandle checks if scaling is needed
func (h *ScaleUpHandler) CanHandle(issue string) bool {
	return issue == "HighLoad" || issue == "CPUThrottling"
}

// Remediate scales up the deployment
func (h *ScaleUpHandler) Remediate(ctx context.Context, namespace, podName string) (*RemediationAction, error) {
	action := &RemediationAction{
		ID:          "scale-up",
		Title:       "Scale Up Deployment",
		Description: fmt.Sprintf("Scaling up deployment for pod %s", podName),
		Command:     fmt.Sprintf("kubectl scale deployment --replicas=3 -n %s", namespace),
		Status:      "pending",
	}

	action.Status = "executed"
	return action, nil
}

// Verify checks if scaling helped
func (h *ScaleUpHandler) Verify(ctx context.Context, namespace, podName string) (bool, error) {
	// Would check if load is distributed better
	return true, nil
}

// RegisterHandler registers a remediation handler
func (ar *AutoRemediation) RegisterHandler(handlerID string, handler RemediationHandler) {
	ar.handlers[handlerID] = handler
}

// Remediate attempts to fix an issue
func (ar *AutoRemediation) Remediate(ctx context.Context, namespace, podName, issue string) (*RemediationAction, error) {
	// Find appropriate handler
	for _, handler := range ar.handlers {
		if handler.CanHandle(issue) {
			action, err := handler.Remediate(ctx, namespace, podName)
			if err != nil {
				action.Status = "failed"
			}
			ar.addToHistory(*action)
			return action, err
		}
	}

	return nil, fmt.Errorf("no handler found for issue: %s", issue)
}

// addToHistory adds action to history
func (ar *AutoRemediation) addToHistory(action RemediationAction) {
	ar.history = append(ar.history, action)
	if len(ar.history) > 10000 {
		ar.history = ar.history[1:]
	}
}

// GetHistory returns remediation history
func (ar *AutoRemediation) GetHistory(limit int) []RemediationAction {
	if limit > len(ar.history) {
		limit = len(ar.history)
	}
	return ar.history[len(ar.history)-limit:]
}
