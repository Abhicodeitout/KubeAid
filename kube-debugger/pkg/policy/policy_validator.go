package policy

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// PolicyViolation represents a policy violation
type PolicyViolation struct {
	Policy      string
	Severity    string // Must, Should, May
	Issue       string
	Fix         string
	Reference   string
}

// Policy represents a validation policy
type Policy struct {
	ID          string
	Name        string
	Type        string // resource, security, naming, registry
	Enforcement string // Must, Should, May
	Rules       []string
	Description string
}

// PolicyValidator validates pod compliance
type PolicyValidator struct {
	policies map[string]Policy
}

// New creates a new policy validator
func New() *PolicyValidator {
	return &PolicyValidator{
		policies: initDefaultPolicies(),
	}
}

// initDefaultPolicies initializes default policies
func initDefaultPolicies() map[string]Policy {
	return map[string]Policy{
		"memory-requests": {
			ID:          "memory-requests",
			Name:        "Memory Requests Required",
			Type:        "resource",
			Enforcement: "Must",
			Description: "All containers must have memory requests defined",
		},
		"memory-limits": {
			ID:          "memory-limits",
			Name:        "Memory Limits Required",
			Type:        "resource",
			Enforcement: "Must",
			Description: "All containers must have memory limits defined",
		},
		"cpu-requests": {
			ID:          "cpu-requests",
			Name:        "CPU Requests Required",
			Type:        "resource",
			Enforcement: "Must",
			Description: "All containers must have CPU requests defined",
		},
		"cpu-limits": {
			ID:          "cpu-limits",
			Name:        "CPU Limits Required",
			Type:        "resource",
			Enforcement: "Should",
			Description: "All containers should have CPU limits defined",
		},
		"security-context": {
			ID:          "security-context",
			Name:        "Security Context Required",
			Type:        "security",
			Enforcement: "Must",
			Description: "Pods must define security context with runAsNonRoot",
		},
		"readonly-fs": {
			ID:          "readonly-fs",
			Name:        "Read-only Filesystem",
			Type:        "security",
			Enforcement: "Should",
			Description: "Containers should have read-only root filesystem",
		},
		"liveness-probe": {
			ID:          "liveness-probe",
			Name:        "Liveness Probe Required",
			Type:        "health",
			Enforcement: "Should",
			Description: "Containers should define liveness probes",
		},
		"readiness-probe": {
			ID:          "readiness-probe",
			Name:        "Readiness Probe Required",
			Type:        "health",
			Enforcement: "Should",
			Description: "Containers should define readiness probes",
		},
	}
}

// ValidatePod validates a pod against policies
func (pv *PolicyValidator) ValidatePod(pod *corev1.Pod) []PolicyViolation {
	violations := make([]PolicyViolation, 0)

	// Check resource policies
	violations = append(violations, pv.validateResourcePolicies(pod)...)

	// Check security policies
	violations = append(violations, pv.validateSecurityPolicies(pod)...)

	// Check health check policies
	violations = append(violations, pv.validateHealthPolicies(pod)...)

	// Check naming policies
	violations = append(violations, pv.validateNamingPolicies(pod)...)

	return violations
}

// validateResourcePolicies checks resource allocation
func (pv *PolicyValidator) validateResourcePolicies(pod *corev1.Pod) []PolicyViolation {
	violations := make([]PolicyViolation, 0)

	for _, container := range pod.Spec.Containers {
		resources := container.Resources

		// Check memory requests
		if resources.Requests == nil || resources.Requests.Memory() == nil || resources.Requests.Memory().IsZero() {
			violations = append(violations, PolicyViolation{
				Policy:    "memory-requests",
				Severity:  "Must",
				Issue:     fmt.Sprintf("Container '%s' has no memory request", container.Name),
				Fix:       "Add resources.requests.memory: 256Mi",
				Reference: "Kubernetes Resource Requests docs",
			})
		}

		// Check memory limits
		if resources.Limits == nil || resources.Limits.Memory() == nil || resources.Limits.Memory().IsZero() {
			violations = append(violations, PolicyViolation{
				Policy:    "memory-limits",
				Severity:  "Must",
				Issue:     fmt.Sprintf("Container '%s' has no memory limit", container.Name),
				Fix:       "Add resources.limits.memory: 512Mi",
				Reference: "Kubernetes Resource Limits docs",
			})
		}

		// Check CPU requests
		if resources.Requests == nil || resources.Requests.Cpu() == nil || resources.Requests.Cpu().IsZero() {
			violations = append(violations, PolicyViolation{
				Policy:    "cpu-requests",
				Severity:  "Must",
				Issue:     fmt.Sprintf("Container '%s' has no CPU request", container.Name),
				Fix:       "Add resources.requests.cpu: 100m",
				Reference: "Kubernetes Resource Requests docs",
			})
		}

		// Check CPU limits
		if resources.Limits == nil || resources.Limits.Cpu() == nil || resources.Limits.Cpu().IsZero() {
			violations = append(violations, PolicyViolation{
				Policy:    "cpu-limits",
				Severity:  "Should",
				Issue:     fmt.Sprintf("Container '%s' has no CPU limit", container.Name),
				Fix:       "Add resources.limits.cpu: 200m",
				Reference: "Kubernetes Resource Limits docs",
			})
		}
	}

	return violations
}

// validateSecurityPolicies checks security configuration
func (pv *PolicyValidator) validateSecurityPolicies(pod *corev1.Pod) []PolicyViolation {
	violations := make([]PolicyViolation, 0)

	// Check pod security context
	if pod.Spec.SecurityContext == nil || pod.Spec.SecurityContext.RunAsNonRoot == nil || !*pod.Spec.SecurityContext.RunAsNonRoot {
		violations = append(violations, PolicyViolation{
			Policy:    "security-context",
			Severity:  "Must",
			Issue:     "Pod does not enforce non-root user",
			Fix:       "Add securityContext: runAsNonRoot: true, runAsUser: 1000",
			Reference: "Pod Security Standards docs",
		})
	}

	// Check container security context
	for _, container := range pod.Spec.Containers {
		if container.SecurityContext == nil || container.SecurityContext.ReadOnlyRootFilesystem == nil || !*container.SecurityContext.ReadOnlyRootFilesystem {
			violations = append(violations, PolicyViolation{
				Policy:    "readonly-fs",
				Severity:  "Should",
				Issue:     fmt.Sprintf("Container '%s' has writable root filesystem", container.Name),
				Fix:       "Add readOnlyRootFilesystem: true in container securityContext",
				Reference: "Container Security docs",
			})
		}
	}

	return violations
}

// validateHealthPolicies checks health check configuration
func (pv *PolicyValidator) validateHealthPolicies(pod *corev1.Pod) []PolicyViolation {
	violations := make([]PolicyViolation, 0)

	for _, container := range pod.Spec.Containers {
		// Check liveness probe
		if container.LivenessProbe == nil {
			violations = append(violations, PolicyViolation{
				Policy:    "liveness-probe",
				Severity:  "Should",
				Issue:     fmt.Sprintf("Container '%s' has no liveness probe", container.Name),
				Fix:       "Add livenessProbe with httpGet or exec",
				Reference: "Health Check docs",
			})
		}

		// Check readiness probe
		if container.ReadinessProbe == nil {
			violations = append(violations, PolicyViolation{
				Policy:    "readiness-probe",
				Severity:  "Should",
				Issue:     fmt.Sprintf("Container '%s' has no readiness probe", container.Name),
				Fix:       "Add readinessProbe with httpGet or exec",
				Reference: "Health Check docs",
			})
		}
	}

	return violations
}

// validateNamingPolicies checks naming conventions
func (pv *PolicyValidator) validateNamingPolicies(pod *corev1.Pod) []PolicyViolation {
	violations := make([]PolicyViolation, 0)

	// Check pod name format (lowercase alphanumeric and hyphens)
	if !isValidK8sName(pod.Name) {
		violations = append(violations, PolicyViolation{
			Policy:    "naming-convention",
			Severity:  "Should",
			Issue:     fmt.Sprintf("Pod name '%s' doesn't follow conventions", pod.Name),
			Fix:       "Use lowercase alphanumeric and hyphens only",
			Reference: "Kubernetes Naming Conventions",
		})
	}

	return violations
}

// isValidK8sName checks if name follows Kubernetes naming conventions
func isValidK8sName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	return true
}

// GetPolicySummary returns a summary of compliance
func (pv *PolicyValidator) GetPolicySummary(violations []PolicyViolation) map[string]int {
	summary := make(map[string]int)
	summary["total"] = len(violations)
	summary["must"] = 0
	summary["should"] = 0
	summary["may"] = 0

	for _, v := range violations {
		switch v.Severity {
		case "Must":
			summary["must"]++
		case "Should":
			summary["should"]++
		case "May":
			summary["may"]++
		}
	}

	return summary
}

// AddPolicy adds a custom policy
func (pv *PolicyValidator) AddPolicy(policy Policy) {
	pv.policies[policy.ID] = policy
}

// RemovePolicy removes a policy
func (pv *PolicyValidator) RemovePolicy(policyID string) {
	delete(pv.policies, policyID)
}
