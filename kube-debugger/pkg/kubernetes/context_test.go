package kubernetes

import "testing"

func TestListKubeContexts(t *testing.T) {
	_, err := ListKubeContexts()
	if err != nil {
		t.Log("ListKubeContexts error (expected if no kubeconfig):", err)
	}
}
