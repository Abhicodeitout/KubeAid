# KubeAid Command Reference

This document shows practical examples for all `kube-debugger` commands.

## Full Command Tree

```text
kube-debugger
├── analyze [app-name]
├── crashloops
├── history [app-name]
│   └── clear [app-name]
├── report [app-name]
├── bootstrap
├── context
│   ├── list
│   └── switch [context-name]
├── completion [bash|zsh|fish|powershell]
├── tui
├── version
└── help [command]
```

## Command Surface Status

As of this version, the top-level CLI command tree above is current and complete.

- No additional root commands have been added yet for advanced modules.
- New enterprise packages in `pkg/` are integrated as internal building blocks and are surfaced through existing flows (primarily `analyze` and `report`).
- When dedicated CLI commands are added later, this file will be updated with full examples and flags.

## Global Flags

| Flag | Description |
|---|---|
| `-h`, `--help` | Show help for root command or any subcommand |

## Quick Start

```sh
cd kube-debugger
make build
make bootstrap
```

If you want bootstrap without kubeconfig validation:

```sh
make bootstrap SKIP_KUBECONFIG_CHECK=true
```

## OS Compatibility

`kube-debugger` is supported on Linux, macOS, and Windows.

| OS | Status | Notes |
|---|---|---|
| Linux | Supported | `report --open` uses `xdg-open` or `$BROWSER`. |
| macOS | Supported | `report --open` uses `open` or `$BROWSER`. |
| Windows | Supported | `report --open` uses `cmd /c start` or `$BROWSER`. |

Required on all OS:
- `kubectl` installed and configured
- Valid kubeconfig (`~/.kube/config` or `KUBECONFIG`)

Optional by feature:
- `gh` CLI for `report --create-issue`
- Ollama for local AI provider
- `GROQ_API_KEY` for Groq provider

Notes:
- `make` targets are easiest on Linux/macOS.
- On Windows, use WSL or Git Bash for `make` workflows.
- If browser auto-open is not available, set `$BROWSER`.

## Root and Help

```sh
./kube-debugger --help
./kube-debugger help
./kube-debugger help analyze
```

## Analyze

Analyze app using default namespace (`default` unless `KUBE_NAMESPACE` is set):

```sh
./kube-debugger analyze my-app
```

Analyze app in specific namespace:

```sh
./kube-debugger analyze my-app --namespace my-namespace
./kube-debugger analyze my-app -n my-namespace
```

Use Ollama AI analysis with env file values:

```sh
source env/kube-debugger.env
./kube-debugger analyze my-app
```

Flags:

| Flag | Short | Default | Description |
|---|---|---|---|
| `--namespace` | `-n` | `default` | Kubernetes namespace for pod lookup |
| `--all-namespaces` | `-A` | `false` | Scan the app across all namespaces |
| `--watch` | - | `false` | Continuously rerun analysis |
| `--interval` | - | `10` | Watch interval in seconds |
| `--exit-code` | - | `false` | Exit with code 2 when score is below threshold |
| `--threshold` | - | `80` | Health score threshold for `--exit-code` |
| `--alert-webhook` | - | empty | Send JSON alert to webhook on low score |
| `--alert-threshold` | - | `80` | Health score threshold for webhook alerts |
| `--help` | `-h` | - | Show help |

Examples:

```sh
./kube-debugger analyze my-app --watch --interval 5
./kube-debugger analyze my-app --exit-code --threshold 85
./kube-debugger analyze my-app --alert-webhook https://example.com/hook --alert-threshold 75
./kube-debugger analyze my-app -A
```

Notes:
- `analyze` is the primary entrypoint for Copilot-style troubleshooting output.
- AI provider output is used when configured; otherwise KubeAid falls back to built-in pattern analysis.
- Security controls (input validation, redaction, RBAC checks, optional audit logging) apply automatically during command execution.

## Crashloops

Detect crash loops and print previous container logs:

```sh
./kube-debugger crashloops
./kube-debugger crashloops -n default
```

Flags:

| Flag | Short | Default | Description |
|---|---|---|---|
| `--namespace` | `-n` | `default` | Namespace to scan |
| `--help` | `-h` | - | Show help |

## History

Show recorded health score history for an app:

```sh
./kube-debugger history my-app
./kube-debugger history my-app -n default
```

Clear stored history:

```sh
./kube-debugger history clear
./kube-debugger history clear my-app
```

## Report

Text report (default):

```sh
./kube-debugger report my-app
```

JSON report:

```sh
./kube-debugger report my-app --format json
./kube-debugger report my-app -f json
```

HTML report to file:

```sh
./kube-debugger report my-app --format html --output report.html
./kube-debugger report my-app -f html -o report.html
```

Report from a specific namespace:

```sh
./kube-debugger report my-app -n my-namespace
```

Flags:

| Flag | Short | Default | Description |
|---|---|---|---|
| `--format` | `-f` | `text` | Output format: `text`, `json`, `html` |
| `--output` | `-o` | stdout | Write report to a file path |
| `--namespace` | `-n` | `default` | Kubernetes namespace for pod lookup |
| `--diff` | - | empty | Compare with a previous JSON report file |
| `--open` | - | `false` | Open HTML report automatically after write |
| `--create-issue` | - | `false` | Create GitHub issue from report via gh CLI |
| `--help` | `-h` | - | Show help |

## Context

List available kubeconfig contexts:

```sh
./kube-debugger context list
```

Switch context:

```sh
./kube-debugger context switch minikube
```

Subcommands:

| Subcommand | Description |
|---|---|
| `context list` | List available kubeconfig contexts |
| `context switch [context-name]` | Switch active kubeconfig context |

## Completion

Generate shell completion scripts:

```sh
./kube-debugger completion bash
./kube-debugger completion zsh
./kube-debugger completion fish
./kube-debugger completion powershell
```

Args:

| Arg | Allowed values |
|---|---|
| `[shell]` | `bash`, `zsh`, `fish`, `powershell` |

## Version

```sh
./kube-debugger version
```

## Bootstrap

Basic environment pre-check:

```sh
./kube-debugger bootstrap
```

Install and start Ollama, then pull free local model:

```sh
./kube-debugger bootstrap --install-ollama --start-ollama --pull-model llama3.2
```

Skip kubeconfig check (for AI-only local setup):

```sh
./kube-debugger bootstrap --install-ollama --start-ollama --pull-model llama3.2 --skip-kubeconfig-check
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `--install-ollama` | `false` | Install Ollama if missing |
| `--start-ollama` | `false` | Start Ollama if installed but not running |
| `--pull-model` | empty | Pull model after setup (example: `llama3.2`) |
| `--skip-kubeconfig-check` | `false` | Skip kubeconfig validation |
| `--help` | - | Show help |

## TUI

Launch interactive terminal UI:

```sh
./kube-debugger tui
```

## Make Targets

Build binary:

```sh
make build
make build VERSION=v1.2.0
```

Run tests:

```sh
make test
```

Run app with `go run`:

```sh
make run
```

Check local dependencies used by bootstrap:

```sh
make deps-check
```

Build + bootstrap in one command:

```sh
make bootstrap
```

Alias target:

```sh
make build-bootstrap
```

Install binary and kubectl-style alias into local bin:

```sh
make install
```

Configurable make variables:

| Variable | Default | Used by |
|---|---|---|
| `ENV_FILE` | `env/kube-debugger.env` | `make run`, `make bootstrap` |
| `OLLAMA_MODEL` | `llama3.2` | `make bootstrap` fallback if env model missing |
| `SKIP_KUBECONFIG_CHECK` | `false` | `make bootstrap` |

Examples:

```sh
# Use custom env file
make bootstrap ENV_FILE=env/my.env

# Skip kubeconfig check for AI-only setup
make bootstrap SKIP_KUBECONFIG_CHECK=true

# Override model without editing env file
make bootstrap OLLAMA_MODEL=qwen2.5:3b
```

## Environment File

Repository-managed env file:

```sh
kube-debugger/env/kube-debugger.env
```

Loaded automatically by `make bootstrap` and `make run`.

Additional runtime variable used by command execution:

| Variable | Default | Purpose |
|---|---|---|
| `KUBE_DEBUGGER_AUDIT` | `false` | Enable security audit logs during command runs |

## One Scenario Per Command

Use these as quick "what should I run now?" examples.

### 1) Root Help

Scenario: You just cloned the repo and want to see all available commands.

```sh
./kube-debugger --help
```

### 2) Analyze

Scenario: Your app is unstable in `payments` namespace and you want an immediate diagnosis.

```sh
./kube-debugger analyze payments-api -n payments
```

### 3) Report

Scenario: You need to share a structured incident report with your team.

```sh
./kube-debugger report payments-api -n payments -f html -o payments-incident-report.html
```

### 4) Bootstrap

Scenario: New machine setup. Install/start Ollama and pull a free model in one shot.

```sh
./kube-debugger bootstrap --install-ollama --start-ollama --pull-model llama3.2 --skip-kubeconfig-check
```

### 5) Context List

Scenario: You work with multiple clusters and need to see available contexts first.

```sh
./kube-debugger context list
```

### 6) Context Switch

Scenario: You are currently on the wrong cluster and want to switch to `prod-eu`.

```sh
./kube-debugger context switch prod-eu
```

### 7) Completion

Scenario: You want CLI tab-completion in your current shell.

```sh
./kube-debugger completion bash
```

### 8) TUI

Scenario: You prefer an interactive terminal view instead of long command output.

```sh
./kube-debugger tui
```

### 9) Version

Scenario: A teammate asks which build you are running before troubleshooting.

```sh
./kube-debugger version
```

### 10) Make Build

Scenario: Build local binary after pulling latest changes.

```sh
make build
```

### 11) Make Bootstrap

Scenario: Run dependency checks + environment setup + Ollama bootstrap from one command.

```sh
make bootstrap
```

### 12) Make Run

Scenario: Run the CLI from source while auto-loading env defaults.

```sh
make run
```

### 13) Make Test

Scenario: Validate project changes before opening a PR.

```sh
make test
```
