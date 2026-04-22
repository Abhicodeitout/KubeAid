package multicluster

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Cluster represents a Kubernetes cluster
type Cluster struct {
	Name       string
	Context    string
	Kubeconfig string
	Region     string
	Provider   string
	Client     kubernetes.Interface
}

// ClusterAnalysis represents analysis of a cluster
type ClusterAnalysis struct {
	ClusterName   string
	HealthScore   int
	TotalPods     int
	FailingPods   int
	WarningPods   int
	Issues        []string
	LastScanned   string
}

// ClusterManager manages multiple clusters
type ClusterManager struct {
	clusters map[string]*Cluster
}

// New creates a new cluster manager
func New() *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*Cluster),
	}
}

// AddCluster adds a cluster to the manager
func (cm *ClusterManager) AddCluster(name, context, kubeconfig, region, provider string) error {
	// Load kubeconfig and create client
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Override context if provided
	if context != "" {
		configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{CurrentContext: context},
		)
		config, err = configLoader.ClientConfig()
		if err != nil {
			return err
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	cm.clusters[name] = &Cluster{
		Name:       name,
		Context:    context,
		Kubeconfig: kubeconfig,
		Region:     region,
		Provider:   provider,
		Client:     client,
	}

	return nil
}

// ListClusters returns all clusters
func (cm *ClusterManager) ListClusters() []*Cluster {
	clusters := make([]*Cluster, 0, len(cm.clusters))
	for _, cluster := range cm.clusters {
		clusters = append(clusters, cluster)
	}
	return clusters
}

// GetCluster returns a specific cluster
func (cm *ClusterManager) GetCluster(name string) *Cluster {
	return cm.clusters[name]
}

// AnalyzeCluster analyzes a single cluster
func (cm *ClusterManager) AnalyzeCluster(ctx context.Context, clusterName string) (*ClusterAnalysis, error) {
	cluster := cm.clusters[clusterName]
	if cluster == nil {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}

	pods, err := cluster.Client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	analysis := &ClusterAnalysis{
		ClusterName: clusterName,
		TotalPods:   len(pods.Items),
		Issues:      make([]string, 0),
	}

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case corev1.PodFailed, corev1.PodUnknown:
			analysis.FailingPods++
		case corev1.PodPending:
			analysis.WarningPods++
		}
	}

	// Calculate health score: (Total - Failing) / Total * 100
	if analysis.TotalPods > 0 {
		analysis.HealthScore = ((analysis.TotalPods - analysis.FailingPods) * 100) / analysis.TotalPods
	} else {
		analysis.HealthScore = 100
	}

	return analysis, nil
}

// AnalyzeAcrossClusters analyzes app across all clusters
func (cm *ClusterManager) AnalyzeAcrossClusters(ctx context.Context, appLabel string) []ClusterAnalysis {
	analyses := make([]ClusterAnalysis, 0, len(cm.clusters))

	for clusterName := range cm.clusters {
		analysis, err := cm.AnalyzeCluster(ctx, clusterName)
		if err == nil {
			analyses = append(analyses, *analysis)
		}
	}

	return analyses
}

// FindDeploymentInconsistencies finds configs that differ across clusters
func (cm *ClusterManager) FindDeploymentInconsistencies(ctx context.Context, namespace, deploymentName string) map[string]interface{} {
	inconsistencies := make(map[string]interface{})
	deployments := make(map[string]interface{})

	for clusterName, cluster := range cm.clusters {
		dep, err := cluster.Client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err == nil {
			deployments[clusterName] = dep.Spec.Replicas
		}
	}

	// Compare replicas across clusters
	if len(deployments) > 1 {
		var firstReplicas *int32
		for _, replicas := range deployments {
			if firstReplicas == nil {
				firstReplicas = replicas.(*int32)
			} else if *replicas.(*int32) != *firstReplicas {
				inconsistencies["replica_inconsistency"] = deployments
			}
		}
	}

	return inconsistencies
}

// GetClusterHealth returns overall health across clusters
func (cm *ClusterManager) GetClusterHealth(ctx context.Context) map[string]int {
	health := make(map[string]int)
	totalScore := 0
	count := 0

	for clusterName := range cm.clusters {
		analysis, err := cm.AnalyzeCluster(ctx, clusterName)
		if err == nil {
			health[clusterName] = analysis.HealthScore
			totalScore += analysis.HealthScore
			count++
		}
	}

	if count > 0 {
		health["average"] = totalScore / count
	}

	return health
}
