package metrics

import (
	"context"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DataPoint represents a single metric measurement
type DataPoint struct {
	Timestamp time.Time
	CPU       float64 // in millicores
	Memory    float64 // in MB
	Disk      float64 // in MB
	Network   float64 // in Mbps
}

// PodMetrics represents metrics for a pod
type PodMetrics struct {
	Pod         string
	Namespace   string
	DataPoints  []DataPoint
	LastUpdated time.Time
}

// MetricsCollector collects and stores metrics
type MetricsCollector struct {
	client     kubernetes.Interface
	timeseries map[string]*PodMetrics
	retention  time.Duration
}

// New creates a new metrics collector
func New(client kubernetes.Interface, retention time.Duration) *MetricsCollector {
	return &MetricsCollector{
		client:     client,
		timeseries: make(map[string]*PodMetrics),
		retention:  retention,
	}
}

// CollectMetrics collects metrics for a pod
func (mc *MetricsCollector) CollectMetrics(ctx context.Context, namespace, podName string) (*DataPoint, error) {
	key := namespace + "/" + podName

	// Get pod to access status
	pod, err := mc.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	dp := &DataPoint{
		Timestamp: time.Now(),
		CPU:       0,
		Memory:    0,
	}

	// Extract resource requests/limits (fallback to estimates)
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			// This is a simplified collection - real implementation would use metrics-server
			if cpu, ok := container.Resources.Requests["cpu"]; ok {
				cpuVal, _ := strconv.ParseFloat(cpu.String(), 64)
				dp.CPU += cpuVal
			}
			if mem, ok := container.Resources.Requests["memory"]; ok {
				memVal, _ := strconv.ParseFloat(mem.String(), 64)
				dp.Memory += memVal / 1024 / 1024 // convert to MB
			}
		}
	}

	// Store in timeseries
	if _, exists := mc.timeseries[key]; !exists {
		mc.timeseries[key] = &PodMetrics{
			Pod:        podName,
			Namespace:  namespace,
			DataPoints: make([]DataPoint, 0),
		}
	}

	mc.timeseries[key].DataPoints = append(mc.timeseries[key].DataPoints, *dp)
	mc.timeseries[key].LastUpdated = time.Now()

	// Clean old data points
	mc.cleanOldDataPoints(key)

	return dp, nil
}

// cleanOldDataPoints removes data points older than retention period
func (mc *MetricsCollector) cleanOldDataPoints(key string) {
	if metrics, exists := mc.timeseries[key]; exists {
		cutoff := time.Now().Add(-mc.retention)
		newPoints := make([]DataPoint, 0)
		for _, dp := range metrics.DataPoints {
			if dp.Timestamp.After(cutoff) {
				newPoints = append(newPoints, dp)
			}
		}
		metrics.DataPoints = newPoints
	}
}

// GetMetrics returns metrics for a pod
func (mc *MetricsCollector) GetMetrics(namespace, podName string) *PodMetrics {
	key := namespace + "/" + podName
	return mc.timeseries[key]
}

// GetAllMetrics returns all collected metrics
func (mc *MetricsCollector) GetAllMetrics() map[string]*PodMetrics {
	return mc.timeseries
}

// AnalyzeTrends analyzes trends in metrics
type TrendAnalyzer struct {
	metrics *PodMetrics
}

// NewTrendAnalyzer creates a trend analyzer
func NewTrendAnalyzer(metrics *PodMetrics) *TrendAnalyzer {
	return &TrendAnalyzer{metrics: metrics}
}

// GetCPUTrend returns CPU trend (bytes per second)
func (ta *TrendAnalyzer) GetCPUTrend() float64 {
	if len(ta.metrics.DataPoints) < 2 {
		return 0
	}

	first := ta.metrics.DataPoints[0]
	last := ta.metrics.DataPoints[len(ta.metrics.DataPoints)-1]

	duration := last.Timestamp.Sub(first.Timestamp).Seconds()
	if duration == 0 {
		return 0
	}

	return (last.CPU - first.CPU) / duration
}

// GetMemoryTrend returns memory trend (MB per hour)
func (ta *TrendAnalyzer) GetMemoryTrend() float64 {
	if len(ta.metrics.DataPoints) < 2 {
		return 0
	}

	first := ta.metrics.DataPoints[0]
	last := ta.metrics.DataPoints[len(ta.metrics.DataPoints)-1]

	duration := last.Timestamp.Sub(first.Timestamp).Hours()
	if duration == 0 {
		return 0
	}

	return (last.Memory - first.Memory) / duration
}

// GetAverageMetrics returns average metrics
func (ta *TrendAnalyzer) GetAverageMetrics() DataPoint {
	avg := DataPoint{Timestamp: time.Now()}

	if len(ta.metrics.DataPoints) == 0 {
		return avg
	}

	for _, dp := range ta.metrics.DataPoints {
		avg.CPU += dp.CPU
		avg.Memory += dp.Memory
		avg.Disk += dp.Disk
		avg.Network += dp.Network
	}

	count := float64(len(ta.metrics.DataPoints))
	avg.CPU /= count
	avg.Memory /= count
	avg.Disk /= count
	avg.Network /= count

	return avg
}

// IsAnomalous detects anomalies in metrics
func (ta *TrendAnalyzer) IsAnomalous() bool {
	if len(ta.metrics.DataPoints) < 5 {
		return false
	}

	// Simple anomaly: if last CPU is 3x average
	avg := ta.GetAverageMetrics()
	last := ta.metrics.DataPoints[len(ta.metrics.DataPoints)-1]

	return last.CPU > avg.CPU*3 || last.Memory > avg.Memory*3
}
