package diagnostics

import "strings"

// SuggestFix returns actionable suggestions based on pod status and last error message.
func SuggestFix(status, lastError string) []string {
	s := strings.ToLower(status)
	e := strings.ToLower(lastError)

	switch {
	case s == "crashloopbackoff":
		suggestions := []string{
			"Check logs for stack traces: kubectl logs <pod> --previous",
			"Verify all required environment variables are set",
			"Ensure config maps and secrets referenced in the spec exist",
		}
		if strings.Contains(e, "oom") || strings.Contains(e, "memory") {
			suggestions = append(suggestions, "Pod may be OOMKilled — increase memory limits")
		}
		if strings.Contains(e, "connection refused") || strings.Contains(e, "dial") {
			suggestions = append(suggestions, "Pod cannot reach a dependency — check service DNS and network policies")
		}
		return suggestions

	case s == "oomkilled":
		return []string{
			"Pod was killed due to out-of-memory — increase resources.limits.memory",
			"Profile application memory usage and fix leaks",
			"Consider enabling a Vertical Pod Autoscaler (VPA)",
		}

	case s == "imagepullbackoff" || s == "errimagepull":
		return []string{
			"Verify the image name and tag are correct",
			"Check that the imagePullSecret exists in the namespace",
			"Confirm the registry is accessible from the cluster network",
			"Run: kubectl describe pod <pod> for the exact pull error",
		}

	case s == "pending":
		return []string{
			"Check node resource availability: kubectl describe nodes",
			"Look for Unschedulable taints or affinity/anti-affinity mismatches",
			"Inspect scheduling events: kubectl describe pod <pod>",
			"Ensure PersistentVolumeClaims are bound if the pod uses storage",
		}

	case s == "evicted":
		return []string{
			"Node was under resource pressure — check disk, memory, or PID pressure",
			"Review eviction events: kubectl describe node <node>",
			"Set proper resource requests to avoid low-priority eviction",
			"Consider adding a PodDisruptionBudget for critical workloads",
		}

	case s == "terminating":
		return []string{
			"Pod is stuck terminating — check for finalizers: kubectl get pod <pod> -o json | jq .metadata.finalizers",
			"Forcefully delete if stuck: kubectl delete pod <pod> --grace-period=0 --force",
		}

	case strings.Contains(s, "containercreating"):
		return []string{
			"Check volume mounts and PVC binding status",
			"Verify ConfigMap and Secret references exist in this namespace",
			"Inspect events: kubectl describe pod <pod>",
		}

	case s == "runcontainererror":
		return []string{
			"Container failed to start — review the entrypoint/command in the spec",
			"Check that the container filesystem is writable if needed",
			"Run: kubectl describe pod <pod> for the exact runtime error",
		}

	case strings.Contains(s, "probe") || strings.Contains(e, "liveness") || strings.Contains(e, "readiness"):
		return []string{
			"Liveness or readiness probe is failing",
			"Check probe endpoint, port, and initial delay settings",
			"Ensure the app is healthy and responding within the timeout",
		}

	default:
		return []string{
			"Review pod events: kubectl describe pod <pod>",
			"Check resource limits and requests",
			"Verify network policies and service endpoints",
		}
	}
}
