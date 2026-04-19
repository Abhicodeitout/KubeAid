package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"kube-debugger/pkg/diagnostics"
)

var (
	advisorMu       sync.Mutex
	advisorContext  []string
	advisorExitCode int
)

func resetAdvisorState() {
	advisorMu.Lock()
	defer advisorMu.Unlock()
	advisorContext = advisorContext[:0]
	advisorExitCode = 0
}

func setAdvisorContextLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	advisorMu.Lock()
	advisorContext = append(advisorContext, line)
	advisorMu.Unlock()
}

func requestExitCode(code int) {
	if code <= 0 {
		return
	}
	advisorMu.Lock()
	if code > advisorExitCode {
		advisorExitCode = code
	}
	advisorMu.Unlock()
}

func consumeAdvisorState() (string, int) {
	advisorMu.Lock()
	defer advisorMu.Unlock()
	ctx := strings.Join(advisorContext, "; ")
	code := advisorExitCode
	advisorContext = advisorContext[:0]
	advisorExitCode = 0
	return ctx, code
}

func printAICommandAdvisor(cmd *cobra.Command, args []string) int {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("KUBEAID_AI_ADVISOR")), "off") {
		_, code := consumeAdvisorState()
		return code
	}
	ctx, exitCode := consumeAdvisorState()

	suggestion := suggestForCommandFallback(cmd, args, ctx)
	cfg := diagnostics.ResolveLLMConfig()
	if cfg.Provider != "" {
		prompt := buildCommandAdvisorPrompt(cmd, args, ctx)
		if out, err := diagnostics.CallLLM(cfg, prompt); err == nil && strings.TrimSpace(out) != "" {
			suggestion = strings.TrimSpace(out)
		}
	}

	fmt.Fprintln(os.Stderr, "\nAI Suggestions")
	fmt.Fprintln(os.Stderr, "--------------")
	fmt.Fprintln(os.Stderr, suggestion)
	return exitCode
}

func buildCommandAdvisorPrompt(cmd *cobra.Command, args []string, contextSummary string) string {
	path := cmd.CommandPath()
	argText := "(none)"
	if len(args) > 0 {
		argText = strings.Join(args, " ")
	}
	if strings.TrimSpace(contextSummary) == "" {
		contextSummary = "(none)"
	}
	return fmt.Sprintf(
		"You are advising after a kube-debugger command completed. Command: %s. Args: %s. Execution context: %s. Return 2-4 concise, actionable next steps for a Kubernetes operator. Prefer concrete kube-debugger and kubectl commands. Do not use markdown code fences.",
		path,
		argText,
		contextSummary,
	)
}

func parseIntContext(ctx, key string) (int, bool) {
	needle := key + "="
	idx := strings.Index(ctx, needle)
	if idx == -1 {
		return 0, false
	}
	start := idx + len(needle)
	end := start
	for end < len(ctx) && ctx[end] >= '0' && ctx[end] <= '9' {
		end++
	}
	if end == start {
		return 0, false
	}
	v, err := strconv.Atoi(ctx[start:end])
	if err != nil {
		return 0, false
	}
	return v, true
}

func suggestForCommandFallback(cmd *cobra.Command, args []string, contextSummary string) string {
	name := cmd.Name()
	switch name {
	case "analyze":
		app := "<app-name>"
		if len(args) > 0 {
			app = args[0]
		}
		if score, ok := parseIntContext(contextSummary, "health_score"); ok {
			switch {
			case score < 50:
				return strings.Join([]string{
					fmt.Sprintf("- Health is critical (%d/100). Capture a report now: kube-debugger report %s -f json -o %s-critical.json", score, app, app),
					fmt.Sprintf("- Trigger fast triage watch: kube-debugger analyze %s --watch --interval 10", app),
					fmt.Sprintf("- Enforce pipeline guard: kube-debugger analyze %s --exit-code --threshold 80", app),
				}, "\n")
			case score < 80:
				return strings.Join([]string{
					fmt.Sprintf("- Health is degraded (%d/100). Compare snapshots: kube-debugger report %s --diff previous.json", score, app),
					fmt.Sprintf("- Keep monitoring: kube-debugger analyze %s --watch --interval 15", app),
					"- Inspect pod events and restarts to isolate the regression window",
				}, "\n")
			}
		}
		return strings.Join([]string{
			"- Save structured output for traceability: kube-debugger report " + app + " -f json -o " + app + "-report.json",
			"- Gate CI on health score: kube-debugger analyze " + app + " --exit-code --threshold 80",
			"- Watch for state changes: kube-debugger analyze " + app + " --watch --interval 10",
		}, "\n")
	case "report":
		if strings.Contains(strings.ToLower(contextSummary), "format=json") {
			return strings.Join([]string{
				"- JSON report is ready for automation; archive it with your incident artifacts",
				"- Compare this output with prior run: kube-debugger report <app-name> --diff previous.json",
				"- Generate a human-readable HTML version for sharing: kube-debugger report <app-name> -f html -o report.html",
			}, "\n")
		}
		return strings.Join([]string{
			"- Compare with previous run: kube-debugger report <app-name> --diff previous.json",
			"- Export HTML for incident sharing: kube-debugger report <app-name> -f html -o report.html --open",
			"- Open a GitHub incident issue: kube-debugger report <app-name> --create-issue",
		}, "\n")
	case "crashloops":
		return strings.Join([]string{
			"- Inspect events around the failing pod: kubectl describe pod <pod> -n <namespace>",
			"- Verify previous container logs manually: kubectl logs <pod> -n <namespace> --previous",
			"- Re-run targeted analysis: kube-debugger analyze <app-name> -n <namespace>",
		}, "\n")
	case "history":
		return strings.Join([]string{
			"- Track trend while debugging: kube-debugger analyze <app-name> --watch --interval 10",
			"- Compare with a saved report baseline to spot regressions quickly",
			"- Use --namespace to narrow history to one environment",
		}, "\n")
	case "context":
		return strings.Join([]string{
			"- Verify active context: kubectl config current-context",
			"- Re-run analysis after switching: kube-debugger analyze <app-name> -n <namespace>",
			"- Keep context explicit in scripts to avoid cross-cluster mistakes",
		}, "\n")
	case "bootstrap":
		return strings.Join([]string{
			"- Validate cluster access: kubectl get pods -A",
			"- Run first diagnosis: kube-debugger analyze <app-name> -n default",
			"- If using local AI, verify provider vars in env/kube-debugger.env",
		}, "\n")
	case "tui":
		return strings.Join([]string{
			"- Use tab to switch views (Pods, Logs, Contexts)",
			"- Press enter on a pod to load logs",
			"- Use r to refresh and q to quit",
		}, "\n")
	case "version":
		return strings.Join([]string{
			"- Confirm expected binary: which kube-debugger",
			"- Run smoke check: kube-debugger --help",
			"- Build with explicit version for release: make build VERSION=vX.Y.Z",
		}, "\n")
	default:
		if strings.TrimSpace(contextSummary) != "" {
			return strings.Join([]string{
				"- Based on this run: " + contextSummary,
				"- Re-run with --help on this command to discover related options",
				"- Use report/analyze pair to move from diagnosis to shareable output",
			}, "\n")
		}
		return strings.Join([]string{
			"- Run kube-debugger --help to discover command-specific options",
			"- Use analyze for live diagnosis and report for shareable outputs",
			"- Use --namespace explicitly to avoid ambiguity across clusters",
		}, "\n")
	}
}
