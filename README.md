# KubeAid

[![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org)
[![Build](https://img.shields.io/badge/build-passing-brightgreen)]()
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

KubeAid is a Kubernetes debugging CLI that analyzes workloads, scores pod health, and gives actionable fix steps.

## What You Get
- App-level diagnosis from pod status, events, logs, and resource hints
- Health score (0-100) with AI or built-in fallback suggestions
- Watch mode, threshold-based exit codes, and webhook alerts
- Report export in text, JSON, and HTML
- Crash loop inspector with previous container logs
- TUI with pod selector, logs panel, and context switching

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
| `kube-debugger report [app-name]` | Export report in text/json/html |
| `kube-debugger crashloops` | Show CrashLoopBackOff pods and previous logs |
| `kube-debugger history [app-name]` | Show recorded health score history |
| `kube-debugger tui` | Launch terminal UI |
| `kube-debugger context list/switch` | Manage kubeconfig context |
| `kube-debugger bootstrap` | Environment checks and optional Ollama setup |

Full command examples: [kube-debugger/COMMANDS.md](kube-debugger/COMMANDS.md)

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

View score history:

```sh
./kube-debugger history my-app -n default
```

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
