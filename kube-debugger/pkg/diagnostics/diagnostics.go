package diagnostics

import (
	"fmt"
	"strings"
)

// SuggestFix returns actionable kubectl suggestions based on pod status and last error message.
// podName is substituted into the commands; pass "" to use the placeholder "<pod>".
func SuggestFix(status, lastError string) []string {
	return SuggestFixForPod(status, lastError, "<pod>", "<namespace>")
}

// SuggestFixForPod is like SuggestFix but substitutes real pod/namespace values.
func SuggestFixForPod(status, lastError, podName, namespace string) []string {
	s := strings.ToLower(status)
	e := strings.ToLower(lastError)
	ns := fmt.Sprintf("-n %s", namespace)

	switch {
	case s == "crashloopbackoff":
		suggestions := []string{
			fmt.Sprintf("kubectl logs %s %s --previous", podName, ns),
			fmt.Sprintf("kubectl describe pod %s %s", podName, ns),
			"Verify all required environment variables are set in the deployment spec",
			"Ensure ConfigMaps and Secrets referenced in the spec exist in the namespace",
		}
		if strings.Contains(e, "oom") || strings.Contains(e, "memory") {
			suggestions = append(suggestions, "Increase memory limits: kubectl edit deployment <deployment>")
		}
		if strings.Contains(e, "connection refused") || strings.Contains(e, "dial") {
			suggestions = append(suggestions, "Pod cannot reach a dependency — check service DNS and network policies")
		}
		return suggestions

	case s == "oomkilled":
		return []string{
			fmt.Sprintf("kubectl describe pod %s %s  # check memory usage and limits", podName, ns),
			"Increase memory limit: kubectl set resources deployment <name> --limits=memory=512Mi",
			"Profile application memory usage and fix leaks",
			"Consider enabling a Vertical Pod Autoscaler (VPA)",
		}

	case s == "imagepullbackoff" || s == "errimagepull":
		return []string{
			fmt.Sprintf("kubectl describe pod %s %s  # see exact pull error", podName, ns),
			"Verify the image name and tag are correct in the deployment spec",
			"Check that imagePullSecret exists: kubectl get secret <secret> " + ns,
			"Test pull manually: docker pull <image>",
		}

	case s == "pending":
		return []string{
			fmt.Sprintf("kubectl describe pod %s %s  # check scheduling events", podName, ns),
			"kubectl describe nodes  # check resource availability and taints",
			"kubectl get pvc " + ns + "  # verify PersistentVolumeClaims are Bound",
			"Check affinity/anti-affinity rules and node selectors in the deployment spec",
		}

	case s == "evicted":
		return []string{
			fmt.Sprintf("kubectl describe pod %s %s", podName, ns),
			"kubectl describe node <node>  # check disk, memory, or PID pressure",
			"Set resource requests to avoid low-priority eviction",
			"Consider adding a PodDisruptionBudget for critical workloads",
		}

	case s == "terminating":
		return []string{
			fmt.Sprintf("kubectl get pod %s %s -o jsonpath='{.metadata.finalizers}'", podName, ns),
			fmt.Sprintf("kubectl delete pod %s %s --grace-period=0 --force  # use only if stuck", podName, ns),
		}

	case strings.Contains(s, "containercreating"):
		return []string{
			fmt.Sprintf("kubectl describe pod %s %s  # check volume mount and secret events", podName, ns),
			"kubectl get pvc " + ns,
			"kubectl get configmap " + ns + "  # verify ConfigMap references exist",
		}

	case s == "runcontainererror":
		return []string{
			fmt.Sprintf("kubectl describe pod %s %s  # see exact runtime error", podName, ns),
			"Check that the container entrypoint/command is correct in the deployment spec",
		}

	case strings.Contains(s, "probe") || strings.Contains(e, "liveness") || strings.Contains(e, "readiness"):
		return []string{
			fmt.Sprintf("kubectl describe pod %s %s  # see probe failure details", podName, ns),
			"Check probe endpoint, port, and initialDelaySeconds in the deployment spec",
			"Ensure the app is healthy and responding within the probe timeout",
		}

	default:
		return []string{
			fmt.Sprintf("kubectl describe pod %s %s", podName, ns),
			fmt.Sprintf("kubectl logs %s %s --tail=100", podName, ns),
			"kubectl get events " + ns + " --sort-by='.lastTimestamp'",
			"Check resource limits and requests in the deployment spec",
		}
	}
}

