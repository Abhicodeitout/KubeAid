APP_NAME=kube-debugger
VERSION?=v1.0.0
OLLAMA_MODEL?=llama3.2
SKIP_KUBECONFIG_CHECK?=false
ENV_FILE?=env/kube-debugger.env
LDFLAGS=-ldflags "-X kube-debugger/cmd.version=$(VERSION)"
NO_AI_ENV=KUBEAID_AI_ADVISOR=off KUBEAID_AI_PROVIDER=

.PHONY: build test unit-test smoke-test run deps-check bootstrap build-bootstrap install lint
build:
	go build $(LDFLAGS) -o $(APP_NAME) ./main.go

test: unit-test smoke-test

unit-test:
	go test ./...

smoke-test: build
	@echo "Running local CLI smoke tests..."
	@$(NO_AI_ENV) ./$(APP_NAME) --help >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) help >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) version >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) history --help >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) crashloops --help >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) tui --help >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) context --help >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) report --help >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) analyze --help >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) completion bash >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) completion zsh >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) completion fish >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) completion powershell >/dev/null
	@$(NO_AI_ENV) ./$(APP_NAME) bootstrap --skip-kubeconfig-check >/dev/null
	@set +e; \
	$(NO_AI_ENV) ./$(APP_NAME) analyze demo-app -n default >/tmp/kubeaid-smoke-analyze.out 2>/tmp/kubeaid-smoke-analyze.err; \
	rc=$$?; \
	[ $$rc -eq 1 ] || (echo "Expected analyze missing app to exit 1, got $$rc" && exit 1)
	@set +e; \
	$(NO_AI_ENV) ./$(APP_NAME) analyze demo-app --exit-code --threshold 1 >/tmp/kubeaid-smoke-analyze2.out 2>/tmp/kubeaid-smoke-analyze2.err; \
	rc=$$?; \
	[ $$rc -eq 1 ] || (echo "Expected analyze missing app with --exit-code to exit 1, got $$rc" && exit 1)
	@set +e; \
	$(NO_AI_ENV) ./$(APP_NAME) report demo-app -n default -f json >/tmp/kubeaid-smoke-report.out 2>/tmp/kubeaid-smoke-report.err; \
	rc=$$?; \
	[ $$rc -eq 1 ] || (echo "Expected report missing app to exit 1, got $$rc" && exit 1)
	@echo "Local smoke tests passed."

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || go vet ./...

install: build
	@mkdir -p $(HOME)/.local/bin
	cp $(APP_NAME) $(HOME)/.local/bin/$(APP_NAME)
	cp $(APP_NAME) $(HOME)/.local/bin/kubectl-debug
	@echo "✅ Installed to ~/.local/bin ($(APP_NAME) and kubectl-debug)"
	@echo "   Make sure ~/.local/bin is in your PATH"

run:
	@set -a; \
	if [ -f "$(ENV_FILE)" ]; then . "$(ENV_FILE)"; fi; \
	set +a; \
	go run ./main.go

deps-check:
	@echo "Checking required tooling..."
	@command -v go >/dev/null || (echo "Missing dependency: go" && exit 1)
	@command -v curl >/dev/null || (echo "Missing dependency: curl (required for Ollama install on Linux/macOS)" && exit 1)
	@echo "Required tooling looks good."
	@echo "Optional tooling checks (warnings only):"
	@command -v kubectl >/dev/null || echo "  - kubectl not found (needed for cluster analysis commands)"
	@command -v minikube >/dev/null || echo "  - minikube not found (optional local cluster)"
	@command -v docker >/dev/null || echo "  - docker not found (optional, used by minikube/kind setups)"

bootstrap: build deps-check
	@echo "Running kube-debugger bootstrap with Ollama setup..."
	@set -a; \
	if [ -f "$(ENV_FILE)" ]; then \
		. "$(ENV_FILE)"; \
		echo "Loaded env from $(ENV_FILE)"; \
	else \
		echo "Env file not found at $(ENV_FILE); using shell environment only"; \
	fi; \
	set +a; \
	flags="--install-ollama --start-ollama --pull-model $${KUBEAID_AI_MODEL:-$(OLLAMA_MODEL)}"; \
	if [ "$(SKIP_KUBECONFIG_CHECK)" = "true" ]; then \
		flags="$$flags --skip-kubeconfig-check"; \
	fi; \
	./$(APP_NAME) bootstrap $$flags

build-bootstrap: bootstrap
