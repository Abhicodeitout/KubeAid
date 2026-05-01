# KubeAid

[![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org)
[![Build](https://img.shields.io/badge/build-passing-brightgreen)]()
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

KubeAid is a Kubernetes debugging CLI that analyzes workloads, scores pod health, and gives actionable fix steps.

## Demo

> **KubeAid v0.0.1** — deploying a crash-looping app and running every command end-to-end.

[![Download Demo MP4](https://img.shields.io/badge/Watch%20Demo-Download%20MP4-blue?style=for-the-badge&logo=github)](./demo/kubeaid-v001-demo.mp4?raw=1)

> 📥 Click the button above to download and play the MP4 video (33s · 1074×872 · H.264)

**What's shown:**
1. `version` — tool version banner
2. `context` — active Kubernetes cluster
3. `kubectl get pods` — crash-looping `kubeaid-demo` app
4. `crashloops` — auto-detection of all crashing pods
5. `analyze` — health score 0/100, events, AI hints, Copilot fix steps
6. `report -f json` — full JSON debug report export

## What You Get
- App-level diagnosis from pod status, events, logs, and resource hints
- Health score (0-100) with AI or built-in fallback suggestions
- Watch mode, threshold-based exit codes, and webhook alerts
- Report export in text, JSON, and HTML
- Crash loop inspector with previous container logs
- TUI with pod selector, logs panel, and context switching
- Enterprise security controls (validation, redaction, RBAC checks, audit logging)
- Copilot-style troubleshooting suggestions with AI and pattern fallback

## Quick Start

```sh
cd kube-debugger
make build VERSION=v1.2.0
./kube-debugger --help
```

Analyze one app:

```sh
./kube-debugger analyze my-app -n default
```

Analyze continuously:

```sh
./kube-debugger analyze my-app --watch --interval 10
```

Fail CI when health is too low:

```sh
./kube-debugger analyze my-app --exit-code --threshold 80
```

## Testing

From `kube-debugger/`:

```sh
make test
```

`make test` now runs both:
- `unit-test`: `go test ./...`
- `smoke-test`: local CLI command and exit-code checks

During `make test`, AI execution is intentionally disabled for smoke checks using:
- `KUBEAID_AI_ADVISOR=off`
- `KUBEAID_AI_PROVIDER=`

CI/CD pipelines also run with AI disabled. AI suggestions are intended for developer-invoked commands during manual troubleshooting.

Run targets individually:

```sh
make unit-test
make smoke-test
```

Run integration tests (tagged, optional for local dev):

```sh
go test -tags=integration ./pkg/analyzer -v
```

Recommended pre-release local check:

```sh
make build VERSION=v1.2.0
make test
```

## Install

Homebrew:

```sh
brew tap abhicodeitout/tap
brew install kube-debugger
```

Go install:

```sh
GO111MODULE=on go install github.com/Abhicodeitout/KubeAid/kube-debugger@latest
```

Local plugin-style install:

```sh
cd kube-debugger
make install
```

This installs both:
- `~/.local/bin/kube-debugger`
- `~/.local/bin/kubectl-debug`

## OS Compatibility

KubeAid works on Linux, macOS, and Windows.

| OS | Status | Notes |
|---|---|---|
| Linux | Supported | Full CLI support. `report --open` uses `xdg-open` or `$BROWSER`. |
| macOS | Supported | Full CLI support. `report --open` uses `open` or `$BROWSER`. |
| Windows | Supported | Full CLI support. `report --open` uses `cmd /c start` or `$BROWSER`. |

Required on all OS:
- A working Kubernetes context (`~/.kube/config` or `KUBECONFIG`)
- `kubectl` installed and configured

Optional by feature:
- `gh` CLI for `report --create-issue`
- Ollama for local AI provider (`KUBEAID_AI_PROVIDER=ollama`)
- Groq API key for cloud AI provider (`KUBEAID_AI_PROVIDER=groq`)

Notes:
- `make` targets are easiest on Linux/macOS; on Windows, running via WSL or Git Bash is recommended.
- Browser auto-open depends on an available launcher; set `$BROWSER` to override.

## Core Commands

| Command | Purpose |
|---|---|
| `kube-debugger analyze [app-name]` | Analyze an app and print diagnosis |
| `kube-debugger timeline [app-name]` | Reconstruct the ordered failure timeline |
| `kube-debugger report [app-name]` | Export report in text/json/html |
| `kube-debugger crashloops` | Show CrashLoopBackOff pods and previous logs |
| `kube-debugger history [app-name]` | Show recorded health score history |
| `kube-debugger tui` | Launch terminal UI |
| `kube-debugger context list/switch` | Manage kubeconfig context |
| `kube-debugger bootstrap` | Environment checks and optional Ollama setup |

Full command examples: [kube-debugger/COMMANDS.md](kube-debugger/COMMANDS.md)

Roadmap tracker: [ROADMAP_V2.md](ROADMAP_V2.md)

## Copilot-Style Suggestions

KubeAid includes Copilot-style troubleshooting guidance that produces structured, actionable fix suggestions.

- Detects common failure patterns (CrashLoopBackOff, OOMKilled, ImagePull, probe failures, RBAC, network timeouts, and TLS issues)
- Provides step-by-step remediation guidance
- Suggests copy-runnable kubectl commands
- Includes YAML snippets when configuration fixes are needed
- Uses AI provider output when available, with automatic pattern-based fallback

See full details in [kube-debugger/COPILOT.md](kube-debugger/COPILOT.md).

## Advanced Modules

The repository now includes advanced modules under [kube-debugger/pkg](kube-debugger/pkg) for enterprise expansion:

- [kube-debugger/pkg/alerts](kube-debugger/pkg/alerts): alert manager, channel abstraction, deduplication, throttling
- [kube-debugger/pkg/metrics](kube-debugger/pkg/metrics): metrics collection, retention, trend and anomaly helpers
- [kube-debugger/pkg/optimizer](kube-debugger/pkg/optimizer): cost optimization suggestions
- [kube-debugger/pkg/prediction](kube-debugger/pkg/prediction): predictive failure analysis helpers
- [kube-debugger/pkg/remediation](kube-debugger/pkg/remediation): auto-remediation handler framework
- [kube-debugger/pkg/policy](kube-debugger/pkg/policy): policy and compliance validation primitives
- [kube-debugger/pkg/integrations](kube-debugger/pkg/integrations): integration hub abstractions (Slack, PagerDuty, Email, Datadog)
- [kube-debugger/pkg/multicluster](kube-debugger/pkg/multicluster): multi-cluster management helpers
- [kube-debugger/pkg/comparison](kube-debugger/pkg/comparison): cross-environment comparison utilities
- [kube-debugger/pkg/rules](kube-debugger/pkg/rules): custom rule engine primitives
- [kube-debugger/pkg/reporting](kube-debugger/pkg/reporting): markdown/html/json report generation helpers

Current status:
- These modules compile and are ready for integration.
- They are package-level building blocks and are not exposed as top-level CLI commands yet.

## Common Workflows

Alert when health drops:

```sh
./kube-debugger analyze my-app \
	--alert-webhook https://example.com/hook \
	--alert-threshold 75
```

Scan across all namespaces:

```sh
./kube-debugger analyze my-app -A
```

Export and diff reports:

```sh
./kube-debugger report my-app -f json -o current.json
./kube-debugger report my-app --diff previous.json
```

Open HTML report in browser:

```sh
./kube-debugger report my-app -f html -o report.html --open
```

Create a GitHub issue from report data:

```sh
./kube-debugger report my-app --create-issue
```

Inspect crash loops:

```sh
./kube-debugger crashloops -n default
```

Reconstruct incident flow:

```sh
./kube-debugger timeline my-app -n default
```

View score history:

```sh
./kube-debugger history my-app -n default
```

## Security

KubeAid includes enterprise-grade security features to protect your cluster data and operations.

### Features

- **Input Validation**: Validates all user inputs against Kubernetes naming conventions (RFC 1123)
- **Secret Redaction**: Automatically redacts API keys, tokens, passwords, and credentials from logs and output
- **Audit Logging**: Logs all operations with timestamps, users, commands, and results for compliance and troubleshooting
- **RBAC Checks**: Verifies user permissions before executing operations
- **TLS/Certificate Validation**: Enforces secure cluster connections with strong cipher suites
- **Output Filtering**: Removes sensitive data from command output (file paths, emails, IPs, secrets)
- **Rate Limiting**: Prevents abuse with configurable operation limits (100 req/sec default)
- **Configuration Security**: Handles kubeconfig and environment variables securely

### Enable Audit Logging

Audit logging is **disabled by default**. Enable it to capture all operations:

```bash
export KUBE_DEBUGGER_AUDIT=true
./kube-debugger analyze my-app
```

Audit logs are stored in:
```
~/.kube-debugger/audit/kube-debugger-audit.log
```

### Audit Log Example

Each log entry is structured JSON with complete operation details:

```json
{
  "timestamp": "2026-04-22T14:21:28Z",
  "event_type": "command_execution",
  "username": "alice",
  "command": "analyze",
  "arguments": ["my-app"],
  "app_name": "my-app",
  "namespace": "production",
  "status": "success"
}
```

### View Audit Logs

```bash
# View last 10 audit entries
tail -10 ~/.kube-debugger/audit/kube-debugger-audit.log

# Search for failures
grep "failure" ~/.kube-debugger/audit/kube-debugger-audit.log

# Monitor in real-time
tail -f ~/.kube-debugger/audit/kube-debugger-audit.log
```

### Input Validation Examples

KubeAid validates all inputs to prevent injection attacks:

```bash
# Invalid app name - REJECTED
./kube-debugger analyze "app;rm -rf"
# Error: app name contains invalid characters; must match DNS subdomain rules

# Invalid namespace - REJECTED
./kube-debugger analyze my-app -n "invalid_ns"
# Error: namespace contains invalid characters

# Invalid threshold - REJECTED
./kube-debugger analyze my-app --threshold 150
# Error: threshold must be between 0 and 100

# Valid inputs - ACCEPTED
./kube-debugger analyze my-app -n default --threshold 80
# Success: analysis completed
```

### Secret Protection

Sensitive data is automatically redacted:

```
Input:  connection using api_key=sk-12345abcdef and password=super-secret
Output: connection using api_key=[REDACTED] and password=[REDACTED]

Input:  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Output: Authorization: [REDACTED_JWT]
```

### Security Best Practices

1. **Enable Audit Logging**: Set `KUBE_DEBUGGER_AUDIT=true` for compliance and troubleshooting
2. **Secure Kubeconfig**: Ensure `~/.kube/config` has permissions `0600` (owner only)
3. **Review Audit Logs**: Monitor regularly for suspicious activity
4. **Use RBAC**: Grant kube-debugger minimal required permissions
5. **Don't Disable TLS**: Never set `KUBECONFIG_INSECURE_SKIP_VERIFY=true` in production
6. **Keep Updated**: Run latest version for security patches

### Configuration

Security settings can be customized via environment variables:

| Variable | Description | Default |
|---|---|---|
| `KUBE_DEBUGGER_AUDIT` | Enable audit logging | `false` |
| `KUBECONFIG_INSECURE_SKIP_VERIFY` | Disable cert verification (not recommended) | `false` |
| `HOME/.kube/config` | Kubeconfig file location | `~/.kube/config` |

### Audit Log Rotation

Audit logs automatically rotate when they exceed 10MB:

```
kube-debugger-audit.log      (current)
kube-debugger-audit.log.1    (previous)
kube-debugger-audit.log.2    (older)
...
kube-debugger-audit.log.10   (oldest)
```

### Security Documentation

For detailed security architecture, implementation, and advanced configuration, see:
- [kube-debugger/pkg/security](kube-debugger/pkg/security) - Security implementation package

## AI Configuration

KubeAid supports two providers:
- Ollama (local)
- Groq (cloud)

If provider settings are missing or provider call fails, KubeAid falls back automatically to built-in pattern analysis.

Use the repository env file:

```sh
cd kube-debugger
source env/kube-debugger.env
```

Key variables:

| Variable | Description |
|---|---|
| `KUBEAID_AI_PROVIDER` | `ollama` or `groq` |
| `KUBEAID_AI_MODEL` | Model override |
| `KUBEAID_OLLAMA_URL` | Ollama endpoint (default `http://localhost:11434`) |
| `KUBEAID_AI_TIMEOUT_SECONDS` | Request timeout override |
| `GROQ_API_KEY` | Required for Groq |

Bootstrap helpers:

```sh
cd kube-debugger
make bootstrap
make bootstrap SKIP_KUBECONFIG_CHECK=true
```

## Why KubeAid
- Faster first diagnosis than manually chaining `kubectl` commands
- Suggestions are operational and copy-runnable
- Works in local dev, CI, and multi-context environments

## Contributing
PRs and issues are welcome.

## License
Apache-2.0
