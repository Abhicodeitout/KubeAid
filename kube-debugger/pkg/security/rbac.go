package security

import (
	"context"
	"fmt"

	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RBACChecker provides methods to check Kubernetes RBAC permissions
type RBACChecker struct {
	clientset *kubernetes.Clientset
}

// NewRBACChecker creates a new RBAC checker
func NewRBACChecker(clientset *kubernetes.Clientset) *RBACChecker {
	return &RBACChecker{clientset: clientset}
}

// CanRead checks if user can read a resource
func (rc *RBACChecker) CanRead(ctx context.Context, namespace, resource, name string) (bool, error) {
	return rc.checkPermission(ctx, namespace, resource, name, "get")
}

// CanList checks if user can list resources
func (rc *RBACChecker) CanList(ctx context.Context, namespace, resource string) (bool, error) {
	return rc.checkPermission(ctx, namespace, resource, "", "list")
}

// CanWatch checks if user can watch resources
func (rc *RBACChecker) CanWatch(ctx context.Context, namespace, resource string) (bool, error) {
	return rc.checkPermission(ctx, namespace, resource, "", "watch")
}

// CanGetLogs checks if user can read pod logs
func (rc *RBACChecker) CanGetLogs(ctx context.Context, namespace, podName string) (bool, error) {
	return rc.checkPermission(ctx, namespace, "pods/log", podName, "get")
}

// CanDescribe checks if user can describe a resource
func (rc *RBACChecker) CanDescribe(ctx context.Context, namespace, resource, name string) (bool, error) {
	return rc.checkPermission(ctx, namespace, resource, name, "get")
}

// checkPermission performs a SelfSubjectAccessReview to check permission
func (rc *RBACChecker) checkPermission(ctx context.Context, namespace, resource, name, verb string) (bool, error) {
	if rc.clientset == nil {
		return false, fmt.Errorf("kubernetes client not initialized")
	}

	// If namespace is not specified, use default
	if namespace == "" {
		namespace = "default"
	}

	review := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Resource:  resource,
				Name:      name,
			},
		},
	}

	result, err := rc.clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to check permission for %s %s: %w", verb, resource, err)
	}

	return result.Status.Allowed, nil
}

// RequirePermissions checks multiple permissions and returns error if any are missing
func (rc *RBACChecker) RequirePermissions(ctx context.Context, permissions map[string]map[string][]string) error {
	for resource, namespaces := range permissions {
		for namespace, verbs := range namespaces {
			for _, verb := range verbs {
				allowed, err := rc.checkPermission(ctx, namespace, resource, "", verb)
				if err != nil {
					return fmt.Errorf("permission check failed for %s/%s in %s: %w", verb, resource, namespace, err)
				}
				if !allowed {
					return fmt.Errorf("insufficient permissions: %s %s in namespace %s", verb, resource, namespace)
				}
			}
		}
	}
	return nil
}

// CheckAppAccess verifies user can access an app (pod) in a namespace
func (rc *RBACChecker) CheckAppAccess(ctx context.Context, appName, namespace string) error {
	// Check read permission on pods
	can, err := rc.CanRead(ctx, namespace, "pods", appName)
	if err != nil {
		return fmt.Errorf("failed to check pod read permission: %w", err)
	}
	if !can {
		return fmt.Errorf("insufficient permissions to read pod '%s' in namespace '%s'", appName, namespace)
	}

	// Check read permission on pod logs
	can, err = rc.CanGetLogs(ctx, namespace, appName)
	if err != nil {
		// Ignore error for logs permission check as it may not exist
		// Some clusters may not have pod/logs resource
	} else if !can {
		return fmt.Errorf("insufficient permissions to read logs for pod '%s' in namespace '%s'", appName, namespace)
	}

	return nil
}

// GetAllowedNamespaces returns list of namespaces user can access
func (rc *RBACChecker) GetAllowedNamespaces(ctx context.Context) ([]string, error) {
	if rc.clientset == nil {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	// Try to list all namespaces
	namespaces, err := rc.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var allowed []string
	for _, ns := range namespaces.Items {
		accessible, _ := rc.CanList(ctx, ns.Name, "pods")
		if accessible {
			allowed = append(allowed, ns.Name)
		}
	}

	return allowed, nil
}
