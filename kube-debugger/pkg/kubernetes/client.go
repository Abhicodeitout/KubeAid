package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"

	"kube-debugger/pkg/security"
)

func kubeConfigPath() string {
	if fromEnv := strings.TrimSpace(os.Getenv("KUBECONFIG")); fromEnv != "" {
		// KUBECONFIG may contain multiple paths; use the first as the primary source.
		return strings.Split(fromEnv, string(os.PathListSeparator))[0]
	}
	return filepath.Join(os.Getenv("HOME"), ".kube", "config")
}

// GetKubeConfig returns the Kubernetes REST config from kubeconfig or in-cluster config.
func GetKubeConfig() (*rest.Config, error) {
	kubeconfig := kubeConfigPath()

	// Validate and secure kubeconfig file
	configHandler := security.NewConfigHandler(os.TempDir())
	if err := configHandler.ValidateKubeconfig(); err != nil {
		// Log warning but allow fallback to in-cluster config
		fmt.Fprintf(os.Stderr, "⚠️  Kubeconfig validation warning: %v\n", err)
	}

	if _, err := os.Stat(kubeconfig); err == nil {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

		// Preserve kubeconfig CA/cert settings and only control insecure verify flag.
		if security.IsInsecureSkipVerifyEnabled() {
			// client-go rejects configs that set both Insecure=true and root CA data.
			config.TLSClientConfig.Insecure = true
			config.TLSClientConfig.CAFile = ""
			config.TLSClientConfig.CAData = nil
			security.WarnIfInsecureSkipVerify()
		} else {
			config.TLSClientConfig.Insecure = false
		}

		return config, nil
	}
	return rest.InClusterConfig()
}

func GetKubeClient() (*kubernetes.Clientset, error) {
	// Apply rate limiting
	secMgr := security.GetSecurityManager()
	secMgr.EnforceRateLimit("kubernetes_client_creation")

	config, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// CheckAccess verifies user has necessary permissions for an app
func CheckAccess(ctx context.Context, clientset *kubernetes.Clientset, appName, namespace string) error {
	rbacChecker := security.NewRBACChecker(clientset)
	return rbacChecker.CheckAppAccess(ctx, appName, namespace)
}
