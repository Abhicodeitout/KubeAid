package diagnostics

import (
	"fmt"
	"strings"
)

// CopilotSuggestion represents an AI suggestion with context
type CopilotSuggestion struct {
	Title       string   // Main suggestion title
	Severity    string   // "critical", "warning", "info"
	Description string   // Detailed explanation
	Steps       []string // Step-by-step fix instructions
	YAMLFix     string   // Suggested YAML fix (if applicable)
	Commands    []string // kubectl commands to try
	Resources   []string // Links or references
}

// BuildCopilotPrompt creates a comprehensive prompt for better AI suggestions
func BuildCopilotPrompt(appName, namespace, podName, status string, restarts int32, logs, events string) string {
	prompt := fmt.Sprintf(`You are a Kubernetes troubleshooting expert assistant (like GitHub Copilot for K8s).
Analyze this pod issue and provide comprehensive, actionable suggestions.

**Pod Details:**
- App/Pod Name: %s
- Namespace: %s
- Pod Name: %s
- Status: %s
- Restart Count: %d

**Recent Logs:**
%s

**Recent Events:**
%s

Please provide:
1. Root cause analysis (what's likely wrong)
2. Severity level (critical/warning/info)
3. Step-by-step fix instructions
4. Suggested kubectl commands to run
5. YAML configuration fixes (if applicable)
6. Prevention tips for the future

Format your response as actionable suggestions that can be copy-pasted.`, appName, namespace, podName, status, restarts, truncateText(logs, 1000), truncateText(events, 500))

	return prompt
}

// EnhancedAnalyzeWithContext provides copilot-like suggestions
func EnhancedAnalyzeWithContext(appName, namespace, podName, status string, restarts int32, logs, events string) *CopilotSuggestion {
	cfg := ResolveLLMConfig()
	
	// Try AI provider first
	if cfg.Provider != "" {
		prompt := BuildCopilotPrompt(appName, namespace, podName, status, restarts, logs, events)
		result, err := CallLLM(cfg, prompt)
		if err == nil && result != "" {
			return parseLLMSuggestion(appName, cfg.Provider, cfg.Model, result)
		}
	}
	
	// Fall back to pattern-based suggestions
	return analyzePatternsCopilot(appName, namespace, podName, status, restarts, logs, events)
}

// parseLLMSuggestion converts LLM response to structured suggestion
func parseLLMSuggestion(appName, provider, model, response string) *CopilotSuggestion {
	suggestion := &CopilotSuggestion{
		Title:       fmt.Sprintf("🤖 AI Analysis (%s/%s)", provider, model),
		Severity:    determineSeverity(response),
		Description: response,
		Steps:       extractSteps(response),
		Commands:    extractCommands(response),
		YAMLFix:     extractYAML(response),
	}
	return suggestion
}

// analyzePatternsCopilot provides structured pattern-based suggestions
func analyzePatternsCopilot(appName, namespace, podName, status string, restarts int32, logs, events string) *CopilotSuggestion {
	l := strings.ToLower(logs)
	e := strings.ToLower(events)

	// Network issues
	if strings.Contains(l, "connection refused") || strings.Contains(l, "dial tcp") {
		return &CopilotSuggestion{
			Title:       "🔗 Network Connectivity Issue",
			Severity:    "critical",
			Description: "Pod cannot establish connection to a service or external resource",
			Steps: []string{
				"1. Verify the target service is running: kubectl get svc -n " + namespace,
				"2. Test DNS resolution: kubectl exec -it " + podName + " -n " + namespace + " -- nslookup [service-name]",
				"3. Check network policies: kubectl get networkpolicies -n " + namespace,
				"4. Verify service endpoints: kubectl get endpoints -n " + namespace,
			},
			Commands: []string{
				"kubectl describe svc [service-name] -n " + namespace,
				"kubectl logs " + podName + " -n " + namespace + " | grep -i connection",
				"kubectl get networkpolicies -n " + namespace + " -o yaml",
			},
			Resources: []string{
				"https://kubernetes.io/docs/tasks/debug-application-cluster/debug-service/",
			},
		}
	}

	// Memory issues
	if strings.Contains(l, "oomkilled") || strings.Contains(e, "out of memory") {
		return &CopilotSuggestion{
			Title:       "💾 Out of Memory (OOMKilled)",
			Severity:    "critical",
			Description: "Container killed due to exceeding memory limit",
			Steps: []string{
				"1. Check current memory limit: kubectl get pod " + podName + " -n " + namespace + " -o yaml | grep -A2 resources",
				"2. Monitor memory usage: kubectl top pod " + podName + " -n " + namespace,
				"3. Identify memory leak patterns in logs",
				"4. Increase memory limit or fix memory leak",
			},
			Commands: []string{
				"kubectl set resources pod " + podName + " --limits=memory=1Gi -n " + namespace,
				"kubectl logs " + podName + " -n " + namespace + " --previous | tail -50",
			},
			YAMLFix: `resources:
  limits:
    memory: "1Gi"  # Increase this value
  requests:
    memory: "512Mi"`,
			Resources: []string{
				"https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource/",
			},
		}
	}

	// Image pull issues
	if strings.Contains(e, "imagepullbackoff") || strings.Contains(l, "pull access denied") {
		return &CopilotSuggestion{
			Title:       "📦 Image Pull Failed",
			Severity:    "critical",
			Description: "Container cannot pull the Docker image from registry",
			Steps: []string{
				"1. Verify image name and tag: kubectl describe pod " + podName + " -n " + namespace,
				"2. Check registry credentials exist: kubectl get secrets -n " + namespace,
				"3. Test manual pull: docker pull [image-name:tag]",
				"4. Create docker secret if needed: kubectl create secret docker-registry",
			},
			Commands: []string{
				"kubectl describe pod " + podName + " -n " + namespace + " | grep -A5 Image",
				"docker pull [image-name]",
				"kubectl create secret docker-registry regcred --docker-server=docker.io --docker-username=USER --docker-password=PASS -n " + namespace,
			},
			YAMLFix: `imagePullSecrets:
- name: regcred
containers:
- name: container-name
  image: yourregistry.com/image:tag  # Full image path`,
			Resources: []string{
				"https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/",
			},
		}
	}

	// Crash loop
	if strings.Contains(e, "crashloopbackoff") || restarts > 5 {
		return &CopilotSuggestion{
			Title:       "🔄 CrashLoopBackOff",
			Severity:    "critical",
			Description: "Pod is continuously crashing and restarting",
			Steps: []string{
				"1. Check current logs: kubectl logs " + podName + " -n " + namespace,
				"2. Check previous logs: kubectl logs " + podName + " -n " + namespace + " --previous",
				"3. Check pod events: kubectl describe pod " + podName + " -n " + namespace,
				"4. Fix the root cause (typically: app error, missing config, bad probe settings)",
				"5. Monitor restarts: watch kubectl get pods -n " + namespace,
			},
			Commands: []string{
				"kubectl logs " + podName + " -n " + namespace + " --previous",
				"kubectl describe pod " + podName + " -n " + namespace,
				"kubectl logs " + podName + " -n " + namespace + " -f",
			},
			Resources: []string{
				"https://kubernetes.io/docs/tasks/debug-application-cluster/debug-running-pod/",
			},
		}
	}

	// Probe failures
	if strings.Contains(l, "probe") || strings.Contains(e, "probe failed") {
		return &CopilotSuggestion{
			Title:       "❌ Health Probe Failed",
			Severity:    "warning",
			Description: "Liveness or readiness probe is failing",
			Steps: []string{
				"1. Check probe configuration: kubectl get pod " + podName + " -n " + namespace + " -o yaml | grep -A10 livenessProbe",
				"2. Verify probe endpoint is responding: kubectl exec -it " + podName + " -n " + namespace + " -- curl localhost:[port]/[path]",
				"3. Check initial delay: is it enough for app startup?",
				"4. Verify port is correct",
				"5. Try manual health check",
			},
			Commands: []string{
				"kubectl exec -it " + podName + " -n " + namespace + " -- curl localhost:8080/health",
				"kubectl exec -it " + podName + " -n " + namespace + " -- wget localhost:8080/healthz",
			},
			YAMLFix: `livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30  # Increase for slow-starting apps
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5`,
			Resources: []string{
				"https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
			},
		}
	}

	// Permission issues
	if strings.Contains(l, "permission denied") || strings.Contains(e, "forbidden") {
		return &CopilotSuggestion{
			Title:       "🔐 Permission Denied",
			Severity:    "warning",
			Description: "Pod lacks required permissions or RBAC access",
			Steps: []string{
				"1. Check service account: kubectl get sa -n " + namespace,
				"2. Check RBAC roles: kubectl get roles -n " + namespace,
				"3. Check role bindings: kubectl get rolebindings -n " + namespace,
				"4. Create needed role/rolebinding or update service account",
			},
			Commands: []string{
				"kubectl get sa -n " + namespace + " -o yaml",
				"kubectl auth can-i list pods --as=system:serviceaccount:" + namespace + ":default",
			},
			Resources: []string{
				"https://kubernetes.io/docs/reference/access-authn-authz/rbac/",
			},
		}
	}

	// Default suggestion
	return &CopilotSuggestion{
		Title:       "🔍 Analyze Logs Manually",
		Severity:    "info",
		Description: "No obvious pattern detected. Review the logs below carefully.",
		Steps: []string{
			"1. Check recent logs: kubectl logs " + podName + " -n " + namespace,
			"2. Check previous logs: kubectl logs " + podName + " -n " + namespace + " --previous",
			"3. Check pod description: kubectl describe pod " + podName + " -n " + namespace,
			"4. Search for error patterns: 'Error', 'Exception', 'panic', 'fatal'",
		},
		Commands: []string{
			"kubectl logs " + podName + " -n " + namespace,
			"kubectl describe pod " + podName + " -n " + namespace,
			"kubectl get events -n " + namespace + " --sort-by='.lastTimestamp'",
		},
		Resources: []string{
			"https://kubernetes.io/docs/tasks/debug-application-cluster/",
		},
	}
}

// Format suggestion for display
func (s *CopilotSuggestion) Format() string {
	if s == nil {
		return ""
	}

	output := fmt.Sprintf("## %s [%s]\n\n", s.Title, s.Severity)
	output += fmt.Sprintf("**Description:** %s\n\n", s.Description)

	if len(s.Steps) > 0 {
		output += "**Fix Steps:**\n"
		for _, step := range s.Steps {
			output += fmt.Sprintf("- %s\n", step)
		}
		output += "\n"
	}

	if len(s.Commands) > 0 {
		output += "**Quick Commands:**\n"
		for _, cmd := range s.Commands {
			output += fmt.Sprintf("```bash\n%s\n```\n", cmd)
		}
		output += "\n"
	}

	if s.YAMLFix != "" {
		output += "**YAML Configuration Fix:**\n"
		output += fmt.Sprintf("```yaml\n%s\n```\n\n", s.YAMLFix)
	}

	if len(s.Resources) > 0 {
		output += "**Learn More:**\n"
		for _, resource := range s.Resources {
			output += fmt.Sprintf("- %s\n", resource)
		}
	}

	return output
}

// Helper functions

func truncateText(text string, maxLen int) string {
	if len(text) > maxLen {
		return text[:maxLen] + "..."
	}
	return text
}

func determineSeverity(text string) string {
	text = strings.ToLower(text)
	if strings.Contains(text, "critical") || strings.Contains(text, "fatal") || strings.Contains(text, "crash") {
		return "critical"
	}
	if strings.Contains(text, "warning") || strings.Contains(text, "error") {
		return "warning"
	}
	return "info"
}

func extractSteps(text string) []string {
	var steps []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") || strings.Contains(line, "Step") || strings.HasPrefix(line, "•") {
			steps = append(steps, strings.TrimLeft(line, "-•"))
		}
	}
	return steps
}

func extractCommands(text string) []string {
	var commands []string
	lines := strings.Split(text, "\n")
	inCodeBlock := false
	for _, line := range lines {
		if strings.Contains(line, "```bash") || strings.Contains(line, "`kubectl") {
			inCodeBlock = true
			continue
		}
		if strings.Contains(line, "```") && inCodeBlock {
			inCodeBlock = false
			continue
		}
		if inCodeBlock && strings.HasPrefix(strings.TrimSpace(line), "kubectl") {
			commands = append(commands, strings.TrimSpace(line))
		}
	}
	return commands
}

func extractYAML(text string) string {
	start := strings.Index(text, "```yaml")
	if start == -1 {
		return ""
	}
	start += 7
	end := strings.Index(text[start:], "```")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(text[start : start+end])
}
