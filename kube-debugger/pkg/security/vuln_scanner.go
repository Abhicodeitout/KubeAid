package security

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID        string // CVE ID
	Severity  string // critical, high, medium, low
	Package   string
	FixedIn   string
	Description string
}

// SecurityIssue represents a security issue
type SecurityIssue struct {
	Type     string // psp, rbac, network, secret
	Issue    string
	Severity string
	Fix      string
}

// VulnerabilityScanner scans for vulnerabilities
type VulnerabilityScanner struct {
	client kubernetes.Interface
}

// New creates a new vulnerability scanner
func NewVulnerabilityScanner(client kubernetes.Interface) *VulnerabilityScanner {
	return &VulnerabilityScanner{client: client}
}

// ScanPodSecurityContext checks pod security context
func (vs *VulnerabilityScanner) ScanPodSecurityContext(pod *corev1.Pod) []SecurityIssue {
	issues := make([]SecurityIssue, 0)

	// Check if running as root
	if pod.Spec.SecurityContext == nil || pod.Spec.SecurityContext.RunAsNonRoot == nil || !*pod.Spec.SecurityContext.RunAsNonRoot {
		issues = append(issues, SecurityIssue{
			Type:     "security-context",
			Issue:    "Pod runs as root",
			Severity: "high",
			Fix:      "Set securityContext.runAsNonRoot: true, runAsUser: 1000",
		})
	}

	// Check for privileged containers
	for _, container := range pod.Spec.Containers {
		if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
			issues = append(issues, SecurityIssue{
				Type:     "security-context",
				Issue:    fmt.Sprintf("Container '%s' runs in privileged mode", container.Name),
				Severity: "critical",
				Fix:      "Remove privileged: true unless absolutely necessary",
			})
		}

		// Check for read-only filesystem
		if container.SecurityContext == nil || container.SecurityContext.ReadOnlyRootFilesystem == nil || !*container.SecurityContext.ReadOnlyRootFilesystem {
			issues = append(issues, SecurityIssue{
				Type:     "security-context",
				Issue:    fmt.Sprintf("Container '%s' has writable root filesystem", container.Name),
				Severity: "medium",
				Fix:      "Set readOnlyRootFilesystem: true in securityContext",
			})
		}
	}

	return issues
}

// ScanImageRegistry checks if container images are from approved registries
func (vs *VulnerabilityScanner) ScanImageRegistry(pod *corev1.Pod, approvedRegistries []string) []SecurityIssue {
	issues := make([]SecurityIssue, 0)

	for _, container := range pod.Spec.Containers {
		image := container.Image

		// Check if image is from approved registry
		isApproved := false
		for _, registry := range approvedRegistries {
			if len(image) >= len(registry) && image[:len(registry)] == registry {
				isApproved = true
				break
			}
		}

		if !isApproved {
			issues = append(issues, SecurityIssue{
				Type:     "image-registry",
				Issue:    fmt.Sprintf("Image '%s' is not from approved registry", image),
				Severity: "high",
				Fix:      fmt.Sprintf("Use approved registry: %v", approvedRegistries),
			})
		}

		// Check for 'latest' tag (bad practice)
		if len(image) >= 6 && image[len(image)-6:] == ":latest" {
			issues = append(issues, SecurityIssue{
				Type:     "image-tag",
				Issue:    fmt.Sprintf("Image '%s' uses 'latest' tag", image),
				Severity: "medium",
				Fix:      "Use specific image tags instead of 'latest'",
			})
		}
	}

	return issues
}

// ScanRBAC checks RBAC configuration
func (vs *VulnerabilityScanner) ScanRBAC(ctx context.Context, namespace, serviceAccount string) []SecurityIssue {
	issues := make([]SecurityIssue, 0)

	// Get service account
	sa, err := vs.client.CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccount, metav1.GetOptions{})
	if err != nil {
		return issues
	}

	// Check if using default service account
	if sa.Name == "default" {
		issues = append(issues, SecurityIssue{
			Type:     "rbac",
			Issue:    "Using default service account",
			Severity: "medium",
			Fix:      "Create a dedicated service account with minimal permissions",
		})
	}

	return issues
}

// ScanNetworkPolicy checks network policies
func (vs *VulnerabilityScanner) ScanNetworkPolicy(ctx context.Context, namespace string) []SecurityIssue {
	issues := make([]SecurityIssue, 0)

	// Check if any network policies exist
	npList, err := vs.client.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil || npList == nil || len(npList.Items) == 0 {
		issues = append(issues, SecurityIssue{
			Type:     "network-policy",
			Issue:    "No network policies defined in namespace",
			Severity: "high",
			Fix:      "Create network policies to restrict pod-to-pod communication",
		})
	}

	return issues
}

// ScanSecrets checks for exposed secrets
func (vs *VulnerabilityScanner) ScanSecrets(ctx context.Context, namespace string) []SecurityIssue {
	issues := make([]SecurityIssue, 0)

	secrets, err := vs.client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues
	}

	for _, secret := range secrets.Items {
		// Check for generic secrets (often used for passwords)
		if secret.Type == corev1.SecretTypeOpaque {
			// Check for common password-like keys
			for key := range secret.Data {
				if key == "password" || key == "apikey" || key == "token" {
					issues = append(issues, SecurityIssue{
						Type:     "secret",
						Issue:    fmt.Sprintf("Secret '%s' contains sensitive key '%s'", secret.Name, key),
						Severity: "high",
						Fix:      "Use external secret management (Vault, Sealed Secrets)",
					})
				}
			}
		}
	}

	return issues
}

// ComprehensiveScan runs all security scans
func (vs *VulnerabilityScanner) ComprehensiveScan(ctx context.Context, pod *corev1.Pod, approvedRegistries []string) []SecurityIssue {
	allIssues := make([]SecurityIssue, 0)

	allIssues = append(allIssues, vs.ScanPodSecurityContext(pod)...)
	allIssues = append(allIssues, vs.ScanImageRegistry(pod, approvedRegistries)...)
	allIssues = append(allIssues, vs.ScanNetworkPolicy(ctx, pod.Namespace)...)
	allIssues = append(allIssues, vs.ScanSecrets(ctx, pod.Namespace)...)

	return allIssues
}
