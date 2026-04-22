package rules

import (
	"fmt"
	"strings"
)

// RuleAction defines what to do when a rule matches
type RuleAction string

const (
	ActionAlert    RuleAction = "alert"
	ActionRemediate RuleAction = "remediate"
	ActionScale    RuleAction = "scale"
	ActionRestart  RuleAction = "restart"
	ActionLog      RuleAction = "log"
)

// RuleCondition defines when a rule matches
type RuleCondition struct {
	Field    string      // pod.restarts, memory.usage, etc.
	Operator string      // >, <, ==, contains, etc.
	Value    interface{} // comparison value
}

// Rule represents a custom rule
type Rule struct {
	ID          string
	Name        string
	Description string
	Enabled     bool
	Conditions  []RuleCondition
	Actions     []RuleAction
	Priority    string // high, medium, low
	Tags        []string
}

// RuleEngine evaluates rules
type RuleEngine struct {
	rules map[string]*Rule
}

// New creates a new rule engine
func New() *RuleEngine {
	return &RuleEngine{
		rules: make(map[string]*Rule),
	}
}

// AddRule adds a rule
func (re *RuleEngine) AddRule(rule *Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}
	re.rules[rule.ID] = rule
	return nil
}

// RemoveRule removes a rule
func (re *RuleEngine) RemoveRule(ruleID string) {
	delete(re.rules, ruleID)
}

// ListRules returns all rules
func (re *RuleEngine) ListRules() []*Rule {
	rules := make([]*Rule, 0, len(re.rules))
	for _, rule := range re.rules {
		rules = append(rules, rule)
	}
	return rules
}

// EvaluateContext evaluates rules against a context
type EvaluationContext struct {
	PodRestarts       int
	MemoryUsage       float64
	CPUUsage          float64
	ReplicaCount      int
	ImageVersion      string
	LastHealthCheck   string
	Namespace         string
	PodName           string
	Status            string
	CustomFields      map[string]interface{}
}

// EvaluationResult represents rule evaluation result
type EvaluationResult struct {
	Rule       *Rule
	Matched    bool
	Actions    []RuleAction
	Message    string
}

// Evaluate evaluates all rules against context
func (re *RuleEngine) Evaluate(ctx EvaluationContext) []EvaluationResult {
	results := make([]EvaluationResult, 0)

	for _, rule := range re.rules {
		if !rule.Enabled {
			continue
		}

		result := EvaluationResult{
			Rule:    rule,
			Matched: re.evaluateConditions(ctx, rule.Conditions),
			Actions: rule.Actions,
		}

		if result.Matched {
			result.Message = fmt.Sprintf("Rule '%s' matched for pod %s", rule.Name, ctx.PodName)
		}

		results = append(results, result)
	}

	return results
}

// evaluateConditions evaluates all conditions in a rule (AND logic)
func (re *RuleEngine) evaluateConditions(ctx EvaluationContext, conditions []RuleCondition) bool {
	for _, cond := range conditions {
		if !re.evaluateCondition(ctx, cond) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (re *RuleEngine) evaluateCondition(ctx EvaluationContext, cond RuleCondition) bool {
	value := re.getContextValue(ctx, cond.Field)
	if value == nil {
		return false
	}

	switch cond.Operator {
	case ">":
		if fVal, ok := value.(float64); ok {
			if fCond, ok := cond.Value.(float64); ok {
				return fVal > fCond
			}
		}
	case "<":
		if fVal, ok := value.(float64); ok {
			if fCond, ok := cond.Value.(float64); ok {
				return fVal < fCond
			}
		}
	case "==":
		return value == cond.Value
	case "!=":
		return value != cond.Value
	case "contains":
		if sVal, ok := value.(string); ok {
			if sCond, ok := cond.Value.(string); ok {
				return strings.Contains(sVal, sCond)
			}
		}
	case ">=":
		if fVal, ok := value.(float64); ok {
			if fCond, ok := cond.Value.(float64); ok {
				return fVal >= fCond
			}
		}
	case "<=":
		if fVal, ok := value.(float64); ok {
			if fCond, ok := cond.Value.(float64); ok {
				return fVal <= fCond
			}
		}
	}

	return false
}

// getContextValue extracts a value from evaluation context
func (re *RuleEngine) getContextValue(ctx EvaluationContext, field string) interface{} {
	parts := strings.Split(field, ".")

	switch parts[0] {
	case "pod":
		if len(parts) > 1 {
			switch parts[1] {
			case "restarts":
				return float64(ctx.PodRestarts)
			case "name":
				return ctx.PodName
			case "status":
				return ctx.Status
			}
		}
	case "memory":
		if len(parts) > 1 && parts[1] == "usage" {
			return ctx.MemoryUsage
		}
	case "cpu":
		if len(parts) > 1 && parts[1] == "usage" {
			return ctx.CPUUsage
		}
	case "replica":
		if len(parts) > 1 && parts[1] == "count" {
			return float64(ctx.ReplicaCount)
		}
	case "image":
		if len(parts) > 1 && parts[1] == "version" {
			return ctx.ImageVersion
		}
	case "namespace":
		return ctx.Namespace
	default:
		// Check custom fields
		if val, ok := ctx.CustomFields[field]; ok {
			return val
		}
	}

	return nil
}

// CreateDefaultRules creates commonly used rules
func (re *RuleEngine) CreateDefaultRules() {
	// High restart rate
	re.AddRule(&Rule{
		ID:          "high-restart-rate",
		Name:        "High Pod Restart Rate",
		Description: "Alert when pod restarts > 5",
		Enabled:     true,
		Conditions: []RuleCondition{
			{Field: "pod.restarts", Operator: ">", Value: 5.0},
		},
		Actions:  []RuleAction{ActionAlert},
		Priority: "high",
	})

	// Memory pressure
	re.AddRule(&Rule{
		ID:          "memory-pressure",
		Name:        "High Memory Usage",
		Description: "Alert when memory usage > 80%",
		Enabled:     true,
		Conditions: []RuleCondition{
			{Field: "memory.usage", Operator: ">", Value: 80.0},
		},
		Actions:  []RuleAction{ActionAlert},
		Priority: "high",
	})

	// CPU throttling
	re.AddRule(&Rule{
		ID:          "cpu-throttling",
		Name:        "High CPU Usage",
		Description: "Alert when CPU usage > 90%",
		Enabled:     true,
		Conditions: []RuleCondition{
			{Field: "cpu.usage", Operator: ">", Value: 90.0},
		},
		Actions:  []RuleAction{ActionAlert},
		Priority: "medium",
	})

	// Replica mismatch
	re.AddRule(&Rule{
		ID:          "low-replicas",
		Name:        "Low Replica Count",
		Description: "Scale up when replicas < 2",
		Enabled:     true,
		Conditions: []RuleCondition{
			{Field: "replica.count", Operator: "<", Value: 2.0},
		},
		Actions:  []RuleAction{ActionScale},
		Priority: "high",
	})
}

// GetMatchedRules returns rules that matched
func (re *RuleEngine) GetMatchedRules(results []EvaluationResult) []*Rule {
	matched := make([]*Rule, 0)
	for _, result := range results {
		if result.Matched {
			matched = append(matched, result.Rule)
		}
	}
	return matched
}

// GetRulesByPriority returns rules filtered by priority
func (re *RuleEngine) GetRulesByPriority(priority string) []*Rule {
	filtered := make([]*Rule, 0)
	for _, rule := range re.rules {
		if rule.Priority == priority {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

// ExportRulesYAML exports rules as YAML (simplified)
func (re *RuleEngine) ExportRulesYAML() string {
	yaml := "rules:\n"
	for _, rule := range re.rules {
		yaml += fmt.Sprintf("  - id: %s\n", rule.ID)
		yaml += fmt.Sprintf("    name: %s\n", rule.Name)
		yaml += fmt.Sprintf("    enabled: %v\n", rule.Enabled)
	}
	return yaml
}
