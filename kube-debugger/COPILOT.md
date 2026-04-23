# Copilot-Style Suggestions

KubeAid includes a built-in "Copilot-style" advisor that prints actionable troubleshooting guidance after CLI commands.

This feature works in two modes:
- AI provider mode (Ollama or Groq) for model-generated suggestions
- Pattern fallback mode when AI is not configured or unavailable

## Where Suggestions Appear

After most commands complete, KubeAid prints an `AI Suggestions` section to stderr.

Examples:
- `kube-debugger analyze my-app -n default`
- `kube-debugger report my-app -f json -o report.json`
- `kube-debugger crashloops -n default`
- `kube-debugger context list`

Notes:
- Completion generation commands skip advisor output.
- Suggestion output can influence process exit code for threshold workflows.

## How It Works

1. Command context is collected during execution (for example: app name, namespace, health score, status).
2. If an AI provider is configured, KubeAid sends a concise prompt and uses the model output.
3. If no provider is configured, or the provider call fails, KubeAid falls back to deterministic rule-based suggestions.
4. Suggestions are printed as next-step actions, usually including `kubectl` or `kube-debugger` commands.

## Configure AI Providers

Use environment variables (see `env/kube-debugger.env` for defaults).

### Ollama (local)

```bash
export KUBEAID_AI_PROVIDER=ollama
export KUBEAID_AI_MODEL=llama3.2
export KUBEAID_OLLAMA_URL=http://localhost:11434
export KUBEAID_AI_TIMEOUT_SECONDS=180
```

### Groq (cloud)

```bash
export KUBEAID_AI_PROVIDER=groq
export KUBEAID_AI_MODEL=llama3-8b-8192
export GROQ_API_KEY=<your_api_key>
export KUBEAID_AI_TIMEOUT_SECONDS=60
```

### Disable Advisor Output

```bash
export KUBEAID_AI_ADVISOR=off
```

When this is set, KubeAid skips post-command suggestion printing.

## Pattern Fallback Coverage

Fallback heuristics cover common Kubernetes failures such as:
- CrashLoopBackOff
- OOMKilled and memory pressure
- Image pull failures
- Probe failures (liveness/readiness)
- Permission or RBAC issues
- Network and timeout errors
- TLS/certificate errors

## Practical Workflow

1. Run analysis:

```bash
kube-debugger analyze my-app -n default
```

2. Review `AI Suggestions` and execute the recommended commands.

3. Export a report for sharing:

```bash
kube-debugger report my-app -f html -o report.html --open
```

4. Gate CI with thresholds:

```bash
kube-debugger analyze my-app --exit-code --threshold 80
```

## Troubleshooting

If suggestions are missing or generic:
- Confirm `KUBEAID_AI_PROVIDER` is set correctly.
- Verify provider connectivity (Ollama URL or Groq API key).
- Check that logs/events are available from the target pod.
- Unset `KUBEAID_AI_ADVISOR` or ensure it is not `off`.

If AI calls fail, KubeAid automatically falls back to pattern-based guidance so you still get next-step recommendations.
