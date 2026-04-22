package prediction

import (
	"time"

	"kube-debugger/pkg/metrics"
)

// PredictedFailure represents a predicted failure
type PredictedFailure struct {
	Type              string        // OOM, Timeout, Probe, etc.
	Probability       float64       // 0.0-1.0
	TimeToEvent       time.Duration // when failure might occur
	Severity          string        // critical, high, medium, low
	ConfidenceScore   float64       // based on data points
	RecommendedAction string
}

// FailurePredictor predicts failures
type FailurePredictor struct {
	metrics *metrics.PodMetrics
}

// New creates a new failure predictor
func New(m *metrics.PodMetrics) *FailurePredictor {
	return &FailurePredictor{metrics: m}
}

// PredictOOMKill predicts OOM kill likelihood
func (fp *FailurePredictor) PredictOOMKill() *PredictedFailure {
	if fp.metrics == nil || len(fp.metrics.DataPoints) < 5 {
		return nil
	}

	analyzer := metrics.NewTrendAnalyzer(fp.metrics)
	trend := analyzer.GetMemoryTrend()
	avg := analyzer.GetAverageMetrics()

	// If memory growing > 100MB/hour and average > 400MB
	if trend > 100 && avg.Memory > 400 {
		// Assume 512MB limit, calculate time to OOM
		memoryAvailable := 512 - avg.Memory
		hoursToOOM := memoryAvailable / trend

		return &PredictedFailure{
			Type:            "OOMKill",
			Probability:     0.75 + (trend / 500), // higher trend = higher probability
			TimeToEvent:     time.Duration(hoursToOOM) * time.Hour,
			Severity:        "critical",
			ConfidenceScore: 0.75,
			RecommendedAction: "Increase memory limit from 512Mi to 1Gi, check for memory leak in appcode",
		}
	}

	return nil
}

// PredictCrashLoop predicts crash loop likelihood
func (fp *FailurePredictor) PredictCrashLoop() *PredictedFailure {
	if fp.metrics == nil {
		return nil
	}

	// In a real implementation, would check pod restart count
	// For now, return nil (would check container restart status)
	return nil
}

// PredictHighCPUUsage predicts CPU spikes
func (fp *FailurePredictor) PredictHighCPUUsage() *PredictedFailure {
	if fp.metrics == nil || len(fp.metrics.DataPoints) < 5 {
		return nil
	}

	analyzer := metrics.NewTrendAnalyzer(fp.metrics)
	avg := analyzer.GetAverageMetrics()
	trend := analyzer.GetCPUTrend()

	// If CPU trend is increasing rapidly
	if trend > 2 && avg.CPU > 300 {
		return &PredictedFailure{
			Type:            "HighCPU",
			Probability:     0.6 + (trend / 100),
			TimeToEvent:     2 * time.Hour,
			Severity:        "high",
			ConfidenceScore: 0.65,
			RecommendedAction: "Monitor CPU usage trends, consider horizontal pod autoscaling",
		}
	}

	return nil
}

// PredictImagePullFailure predicts image pull issues
func (fp *FailurePredictor) PredictImagePullFailure() *PredictedFailure {
	// Would check image registry availability, network conditions
	return nil
}

// PredictTimeoutIssues predicts timeout problems
func (fp *FailurePredictor) PredictTimeoutIssues() *PredictedFailure {
	// Would check network latency, connection patterns
	return nil
}

// GetAllPredictions returns all failure predictions
func (fp *FailurePredictor) GetAllPredictions() []*PredictedFailure {
	predictions := make([]*PredictedFailure, 0)

	if oom := fp.PredictOOMKill(); oom != nil {
		predictions = append(predictions, oom)
	}
	if cpu := fp.PredictHighCPUUsage(); cpu != nil {
		predictions = append(predictions, cpu)
	}
	if crash := fp.PredictCrashLoop(); crash != nil {
		predictions = append(predictions, crash)
	}

	return predictions
}

// GetCriticalPredictions returns only critical predictions
func (fp *FailurePredictor) GetCriticalPredictions() []*PredictedFailure {
	all := fp.GetAllPredictions()
	critical := make([]*PredictedFailure, 0)

	for _, pred := range all {
		if pred.Probability > 0.7 && pred.Severity == "critical" {
			critical = append(critical, pred)
		}
	}

	return critical
}
