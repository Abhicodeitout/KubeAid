package analyzer


import (
	"fmt"
	"context"
	"os"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-debugger/pkg/kubernetes"
	"kube-debugger/pkg/diagnostics"
)

func AnalyzeApp(appName string) string {
	namespace := os.Getenv("KUBE_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	clientset, err := kubernetes.GetKubeClient()
	if err != nil {
		return fmt.Sprintf("❌ Error connecting to cluster: %v", err)
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", appName),
	})
	if err != nil || len(pods.Items) == 0 {
		return fmt.Sprintf("❌ No pods found for app '%s' in namespace '%s'", appName, namespace)
	}
	pod := pods.Items[0]
	status := string(pod.Status.Phase)
	if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].State.Waiting != nil {
		status = pod.Status.ContainerStatuses[0].State.Waiting.Reason
	}
	logs, _ := kubernetes.GetPodLogs(clientset, namespace, pod.Name)
	events, _ := kubernetes.GetPodEvents(clientset, namespace, pod.Name)
	resources, _ := kubernetes.GetPodResourceUsage(clientset, namespace, pod.Name)
	lastError := ""
	if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].LastTerminationState.Terminated != nil {
		lastError = pod.Status.ContainerStatuses[0].LastTerminationState.Terminated.Message
	}
	suggestions := diagnostics.SuggestFix(status, lastError)
	aiHint := diagnostics.AnalyzeLogsAI(logs)

	return fmt.Sprintf(
		"❌ Pod: %s\nStatus: %s\n\n📜 Last Logs:\n%s\n\n🤖 %s\n\n📅 Events:\n%s\n\n📊 Resources:\n%s\n\n📌 Suggestions:\n- %s\n",
		pod.Name, status, logs, aiHint, events, resources, 
		fmt.Sprintf("%s", suggestions),
	)
}
