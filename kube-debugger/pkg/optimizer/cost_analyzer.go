package optimizer

import (
	"fmt"

	"kube-debugger/pkg/metrics"
)

// CostSuggestion represents a cost optimization suggestion
type CostSuggestion struct {
	Title          string
	Description    string
	CurrentCost    float64
	OptimizedCost  float64
	Savings        float64 // percentage
	AnnualSavings  float64 // estimated annual savings in dollars
	Action         string
	Priority       string // high, medium, low
	Implementation string
}

// CostAnalyzer analyzes costs
type CostAnalyzer struct {
	metrics *metrics.PodMetrics
	hourlyRate float64 // cost per container hour in cents
}

// New creates a new cost analyzer
func New(m *metrics.PodMetrics, hourlyRate float64) *CostAnalyzer {
	if hourlyRate == 0 {
		hourlyRate = 0.05 // default $0.05 per hour
	}
	return &CostAnalyzer{
		metrics:    m,
		hourlyRate: hourlyRate,
	}
}

// AnalyzeOverProvisioning detects over-provisioned resources
func (ca *CostAnalyzer) AnalyzeOverProvisioning() []CostSuggestion {
	suggestions := make([]CostSuggestion, 0)

	if ca.metrics == nil || len(ca.metrics.DataPoints) == 0 {
		return suggestions
	}

	analyzer := metrics.NewTrendAnalyzer(ca.metrics)
	avg := analyzer.GetAverageMetrics()

	// Check if memory usage is low (indicating over-provisioning)
	// Assuming typical container is requested 512MB but uses 100MB
	if avg.Memory < 200 { // uses less than 200MB
		suggestion := CostSuggestion{
			Title:       "Right-size Memory Request",
			Description: fmt.Sprintf("Pod uses avg %.1f MB but likely has 512MB requested", avg.Memory),
			CurrentCost: 100,
			OptimizedCost: 50,
			Savings: 50,
			AnnualSavings: 438, // (512-256) * 0.05 * 24 * 365 / 1000
			Action: "Reduce memory request to 256Mi",
			Priority: "medium",
			Implementation: "kubectl set resources pod $POD --requests=memory=256Mi",
		}
		suggestions = append(suggestions, suggestion)
	}

	// Check if CPU usage is low
	if avg.CPU < 100 { // uses less than 100 millicores
		suggestion := CostSuggestion{
			Title:       "Right-size CPU Request",
			Description: fmt.Sprintf("Pod uses avg %.0f millicores but likely has 500m requested", avg.CPU),
			CurrentCost: 100,
			OptimizedCost: 40,
			Savings: 60,
			AnnualSavings: 876, // (500-200) * 0.05 * 24 * 365 / 1000
			Action: "Reduce CPU request to 200m",
			Priority: "medium",
			Implementation: "kubectl set resources pod $POD --requests=cpu=200m",
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

// AnalyzeCostTrends analyzes cost trends
func (ca *CostAnalyzer) AnalyzeCostTrends() []CostSuggestion {
	suggestions := make([]CostSuggestion, 0)

	if ca.metrics == nil || len(ca.metrics.DataPoints) < 2 {
		return suggestions
	}

	analyzer := metrics.NewTrendAnalyzer(ca.metrics)
	memoryTrend := analyzer.GetMemoryTrend()

	// Check if memory is growing (potential leak)
	if memoryTrend > 50 { // growing > 50MB per hour
		suggestion := CostSuggestion{
			Title:       "Potential Memory Leak",
			Description: fmt.Sprintf("Memory growing at %.1f MB/hour - indicates potential leak", memoryTrend),
			Savings: 0,
			Priority: "high",
			Implementation: "Check pod logs for growing memory usage, consider restarting pod weekly",
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

// EstimateMonthlySpend estimates monthly cost
func (ca *CostAnalyzer) EstimateMonthlySpend() float64 {
	if ca.metrics == nil || len(ca.metrics.DataPoints) == 0 {
		return 0
	}

	analyzer := metrics.NewTrendAnalyzer(ca.metrics)
	avg := analyzer.GetAverageMetrics()

	// Simplified: assume 1 container running 24/7
	// Cost = (CPU_millicores/1000) * hourly_rate * 24 * 30
	monthlyCost := (avg.CPU / 1000) * ca.hourlyRate * 24 * 30
	return monthlyCost
}

// GetRecommendations returns all optimization recommendations
func (ca *CostAnalyzer) GetRecommendations() []CostSuggestion {
	recommendations := make([]CostSuggestion, 0)
	recommendations = append(recommendations, ca.AnalyzeOverProvisioning()...)
	recommendations = append(recommendations, ca.AnalyzeCostTrends()...)
	return recommendations
}
