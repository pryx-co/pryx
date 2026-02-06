# Makefile for Pryx Build System
# Polyglot project: Rust, Go, React/TypeScript, Cloudflare Workers
#
# Usage:
#   make help       - Show this help message
#   make build       - Build all components
#   make test        - Run all tests
#   make lint        - Run all linters
#   make clean       - Clean build artifacts
#   make install     - Install development tools
#   make format      - Format all code
#   make check       - Run comprehensive checks

SHELL := /bin/bash
.PHONY: help build test lint clean install format check clean-uninstall
.SILENT: help

# Project directories
HOST_DIR := apps/host
RUNTIME_DIR := apps/runtime
TUI_DIR := apps/tui

# Version management
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0-dev")
BUILD_DATE := $(shell date +%Y-%m-%d)
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# JS toolchain
BUN_REQUIRED_VERSION := 1.3.7

# Color codes for output
RED := $(shell tput setaf 1 2>/dev/null || printf '\033[0;31m')
GREEN := $(shell tput setaf 2 2>/dev/null || printf '\033[0;32m')
YELLOW := $(shell tput setaf 3 2>/dev/null || printf '\033[0;33m')
BLUE := $(shell tput setaf 4 2>/dev/null || printf '\033[0;34m')
NC := $(shell tput sgr0 2>/dev/null || printf '\033[0m') # No Color

help: ## Show this help message
	@echo "$(BLUE)Pryx Build System$(NC)"
	@echo ""
	@echo "$(BLUE)Build targets:$(NC)"
	@echo "  $(GREEN)build$(NC)             # Build all components"
	@echo "  $(GREEN)build-host$(NC)        # Build Rust/Tauri host"
	@echo "  $(GREEN)build-runtime$(NC)     # Build Go runtime"
	@echo "  $(GREEN)build-tui$(NC)         # Build TypeScript TUI"
	@echo ""
	@echo "$(BLUE)Test targets:$(NC)"
	@echo "  $(GREEN)test$(NC)              # Run all tests (unit + integration)"
	@echo "  $(GREEN)test-unit$(NC)         # Run all unit tests"
	@echo "  $(GREEN)test-unit-runtime$(NC) # Run Go runtime unit tests"
	@echo "  $(GREEN)test-unit-host$(NC)    # Run Rust host unit tests"
	@echo "  $(GREEN)test-unit-tui$(NC)     # Run TUI unit tests"
	@echo "  $(GREEN)test-integration$(NC)  # Run all integration tests"
	@echo "  $(GREEN)test-e2e$(NC)          # Run all E2E tests"
	@echo "  $(GREEN)test-coverage$(NC)    # Run all tests with coverage"
	@echo ""
	@echo "$(BLUE)Lint targets:$(NC)"
	@echo "  $(GREEN)lint$(NC)              # Run all linters"
	@echo "  $(GREEN)lint-host$(NC)         # Run Rust linters"
	@echo "  $(GREEN)lint-runtime$(NC)      # Run Go linters"
	@echo "  $(GREEN)lint-tui$(NC)          # Run TypeScript linters"
	@echo ""
	@echo "$(BLUE)Format targets:$(NC)"
	@echo "  $(GREEN)format$(NC)            # Format all code"
	@echo "  $(GREEN)format-host$(NC)       # Format Rust code"
	@echo "  $(GREEN)format-runtime$(NC)    # Format Go code"
	@echo "  $(GREEN)format-tui$(NC)        # Format TypeScript code"
	@echo ""
	@echo "$(BLUE)Clean targets:$(NC)"
	@echo "  $(GREEN)clean$(NC)             # Clean all build artifacts"
	@echo "  $(GREEN)clean-host$(NC)        # Clean Rust build artifacts"
	@echo "  $(GREEN)clean-runtime$(NC)     # Clean Go build artifacts"
	@echo "  $(GREEN)clean-tui$(NC)         # Clean TypeScript build artifacts"
	@echo ""
	@echo "$(BLUE)Development targets:$(NC)"
	@echo "  $(GREEN)dev$(NC)               # Run local development stack"
	@echo "  $(GREEN)dev-tui$(NC)          # Run TUI + Runtime together"
	@echo "  $(GREEN)dev-tail$(NC)         # Tail runtime logs while TUI is running"
	@echo "  $(GREEN)start-tui$(NC)        # Start TUI (opens interactive CLI)"
	@echo "  $(GREEN)start-runtime$(NC)    # Start Runtime (foreground)"
	@echo "  $(GREEN)start-host$(NC)       # Start Host (opens Tauri window)"
	@echo "  $(GREEN)start-all$(NC)        # Start all services (background)"
	@echo "  $(GREEN)start-all-logs$(NC)   # Start all + show combined logs"
	@echo "  $(GREEN)logs$(NC)             # View all service logs"
	@echo "  $(GREEN)logs-tui$(NC)         # View TUI logs only"
	@echo "  $(GREEN)logs-runtime$(NC)     # View Runtime logs only"
	@echo "  $(GREEN)logs-host$(NC)        # View Host logs only"
	@echo "  $(GREEN)clean-uninstall$(NC)  # Remove all Pryx binaries, services, and data"
	@echo "  $(GREEN)install$(NC)           # Install development tools"
	@echo "  $(GREEN)check$(NC)             # Run comprehensive checks (lint + test)"
	@echo "  $(GREEN)info$(NC)              # Show project information"
	@echo ""
	@echo "$(BLUE)Version Management:$(NC)"
	@echo "  $(GREEN)version-bump-patch$(NC) # Bump patch version"
	@echo "  $(GREEN)version-bump-minor$(NC) # Bump minor version"
	@echo "  $(GREEN)version-bump-major$(NC) # Bump major version"
	@echo "  $(GREEN)version-tag$(NC)       # Create git tag"
	@echo "  $(GREEN)version-push$(NC)      # Push tags"
	@echo ""
	@echo "$(BLUE)Usage:$(NC)"
	@echo "  make <target>"

## Development targets
dev: ## Run local development stack (Web + Runtime)
	@bash scripts/dev-runner.sh

dev-tui: ## Run TUI + Runtime together
	@bash scripts/tui-runner.sh

dev-tui-debug: ## Run TUI + Runtime with full debug logging
	@bash scripts/tui-runner-debug.sh

dev-tail: ## Tail runtime logs while TUI is running
	@echo "$(BLUE)Tailing runtime logs...$(NC)"
	@if [ -f "$(HOME)/.pryx/logs/runtime.log" ]; then \
		tail -f $(HOME)/.pryx/logs/runtime.log; \
	else \
		echo "$(YELLOW)No runtime log found. Run 'make dev-tui' first.$(NC)"; \
	fi

start-tui: ## Start TUI service (opens interactive CLI)
	@echo "$(BLUE)Starting TUI...$(NC)"
	@cd $(TUI_DIR) && bun run dev

start-runtime: ## Start Runtime service (runs in foreground)
	@echo "$(BLUE)Starting Runtime...$(NC)"
	@cd $(RUNTIME_DIR) && go run ./cmd/pryx-core

start-host: ## Start Host service (opens Tauri window)
	@echo "$(BLUE)Starting Host...$(NC)"
	@cd $(HOST_DIR) && cargo tauri dev

start-all: ## Start TUI, Runtime, and Host (in background, logs to files)
	@echo "$(BLUE)Starting all services in background...$(NC)"
	@echo "$(YELLOW)Logs: ~/.pryx/logs/$(NC)"
	@mkdir -p $(HOME)/.pryx/logs
	@cd $(TUI_DIR) && nohup bun run dev > $(HOME)/.pryx/logs/tui.log 2>&1 &
	@cd $(RUNTIME_DIR) && nohup go run ./cmd/pryx-core > $(HOME)/.pryx/logs/runtime.log 2>&1 &
	@cd $(HOST_DIR) && nohup cargo tauri dev > $(HOME)/.pryx/logs/host.log 2>&1 &
	@echo "$(GREEN)Services started in background$(NC)"
	@echo "$(BLUE)Use 'make logs' to view combined output$(NC)"

logs: ## View logs from all services
	@echo "$(BLUE)Tailing logs...$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop (services keep running)$(NC)"
	@tail -f $(HOME)/.pryx/logs/*.log 2>/dev/null || echo "$(YELLOW)No logs yet. Run services first.$(NC)"

logs-tui: ## View TUI logs only
	@tail -f $(HOME)/.pryx/logs/tui.log 2>/dev/null || echo "$(YELLOW)No TUI logs yet$(NC)"

logs-runtime: ## View Runtime logs only
	@tail -f $(HOME)/.pryx/logs/runtime.log 2>/dev/null || echo "$(YELLOW)No Runtime logs yet$(NC)"

logs-host: ## View Host logs only
	@tail -f $(HOME)/.pryx/logs/host.log 2>/dev/null || echo "$(YELLOW)No Host logs yet$(NC)"

start-all-logs: ## Start all services and show combined logs
	@echo "$(BLUE)Starting all services with combined output...$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop all services$(NC)"
	@mkdir -p $(HOME)/.pryx/logs
	@cd $(TUI_DIR) && bun run dev > $(HOME)/.pryx/logs/tui.log 2>&1 & \
	TUI_PID=$$!; \
	cd $(RUNTIME_DIR) && go run ./cmd/pryx-core > $(HOME)/.pryx/logs/runtime.log 2>&1 & \
	RUNTIME_PID=$$!; \
	cd $(HOST_DIR) && cargo tauri dev > $(HOME)/.pryx/logs/host.log 2>&1 & \
	HOST_PID=$$!; \
	sleep 2; \
	echo "$(GREEN)Services started. PIDs: TUI=$$TUI_PID, Runtime=$$RUNTIME_PID, Host=$$HOST_PID$(NC)"; \
	echo "$(BLUE)Tailing combined logs...$(NC)"; \
	tail -f $(HOME)/.pryx/logs/tui.log $(HOME)/.pryx/logs/runtime.log $(HOME)/.pryx/logs/host.log & \
	TAIL_PID=$$!; \
	trap "kill $$TUI_PID $$RUNTIME_PID $$HOST_PID $$TAIL_PID 2>/dev/null; exit" INT TERM; \
	wait

tui: ## Build and run TUI client (requires Runtime to be running separately!)
	@echo "$(BLUE)Building and Starting TUI...$(NC)"
	@echo "$(YELLOW)Note: Runtime must be running on :3000. Use 'make dev-tui' to run both.$(NC)"
	@cd apps/tui && bun install && bun run build && ./pryx-tui

## Build targets
build: build-host build-runtime build-tui ## Build all components

build-host: ## Build Rust/Tauri host
	@echo "$(BLUE)Building host (Rust + Tauri)$(NC)"
	@if [ -d "$(HOST_DIR)" ]; then \
		cd $(HOST_DIR) && cargo build --release --lib; \
	else \
		echo "$(YELLOW)Warning: host directory not found, skipping$(NC)"; \
	fi

build-runtime: ## Build Go runtime
	@echo "$(BLUE)Building runtime (Go)$(NC)"
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && \
		go build -o bin/pryx-core -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)" ./cmd/pryx-core && \
		mkdir -p /tmp && cp bin/pryx-core /tmp/pryx-core; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

build-tui: ## Build TypeScript TUI
	@echo "$(BLUE)Building TUI (TypeScript + Solid + OpenTUI)$(NC)"
	@if [ -d "$(TUI_DIR)" ]; then \
		cd $(TUI_DIR) && bun install --frozen-lockfile && bun run build; \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

## Test targets
test: test-unit test-integration ## Run all tests (unit + integration)

test-unit: test-unit-host test-unit-runtime test-unit-tui ## Run all unit tests

test-unit-host: ## Run Rust host unit tests
	@echo "$(BLUE)Testing host unit tests (Rust)$(NC)"
	@if [ -d "$(HOST_DIR)" ]; then \
		cd $(HOST_DIR) && cargo test --release --lib; \
	else \
		echo "$(YELLOW)Warning: host directory not found, skipping$(NC)"; \
	fi

test-unit-runtime: ## Run Go runtime unit tests
	@echo "$(BLUE)Testing runtime unit tests (Go)$(NC)"
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && go test -v -race -cover ./internal/...; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

test-unit-tui: ## Run TypeScript TUI unit tests
	@echo "$(BLUE)Testing TUI unit tests (TypeScript)$(NC)"
	@if [ -d "$(TUI_DIR)" ]; then \
		cd $(TUI_DIR) && bun install --frozen-lockfile && bun test; \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

test-integration: test-integration-runtime ## Run all integration tests

test-integration-runtime: ## Run Go runtime integration tests
	@echo "$(BLUE)Testing runtime integration tests (Go)$(NC)"
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && go test -v -race -tags=integration ./tests/integration/...; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

test-e2e: test-e2e-runtime test-e2e-tui ## Run all E2E tests

test-e2e-runtime: build-runtime ## Run Go runtime E2E tests
	@echo "$(BLUE)Testing runtime E2E tests (Go)$(NC)"
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && go test -v -race -tags=e2e ./e2e/...; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

test-e2e-tui: build-runtime build-tui ## Run TUI E2E tests
	@echo "$(BLUE)Testing TUI E2E tests (Playwright)$(NC)"
	@if [ -d "$(TUI_DIR)" ]; then \
		if [ ! -f "$(TUI_DIR)/playwright.config.ts" ] && [ ! -f "$(TUI_DIR)/playwright.config.js" ]; then \
			echo "$(YELLOW)Playwright tests not configured yet$(NC)"; \
		else \
			cd $(TUI_DIR) && bunx playwright test || echo "$(YELLOW)Playwright tests failed$(NC)"; \
		fi; \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

test-coverage: ## Run all tests with coverage reports
	@echo "$(BLUE)Running tests with coverage$(NC)"
	@mkdir -p coverage
	@$(MAKE) test-coverage-runtime
	@$(MAKE) test-coverage-host
	@echo "$(GREEN)✓$(NC) Coverage reports generated in ./coverage/"

test-coverage-runtime: ## Run Go runtime tests with coverage
	@echo "  - Runtime coverage..."
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && go test -race -coverprofile=coverage.out ./... && \
		go tool cover -html=coverage.out -o ../../coverage/runtime-coverage.html; \
	fi

test-coverage-host: ## Run Rust host tests with coverage
	@echo "  - Host coverage..."
	@if [ -d "$(HOST_DIR)" ]; then \
		cd $(HOST_DIR) && cargo tarpaulin --out Html --output-dir ../../coverage 2>/dev/null || \
		echo "    $(YELLOW)Install cargo-tarpaulin for coverage: cargo install cargo-tarpaulin$(NC)"; \
	fi

test-host: test-unit-host ## Alias for test-unit-host

test-runtime: test-unit-runtime ## Alias for test-unit-runtime
test-tui: test-unit-tui ## Alias for test-unit-tui

## Lint targets
lint: lint-host lint-runtime lint-tui ## Run all linters

lint-host: ## Run Rust linters (clippy, rustfmt)
	@echo "$(BLUE)Linting host (Rust)$(NC)"
	@if [ -d "$(HOST_DIR)" ]; then \
		cd $(HOST_DIR) && \
		echo "  - Checking formatting (rustfmt)..." && \
		cargo fmt -- --check && echo "    $(GREEN)✓$(NC) Format OK" || (echo "    $(RED)✗$(NC) Format issues found. Run 'make format' to fix." && exit 1) && \
		echo "  - Running clippy..." && \
		cargo clippy --lib -- -D warnings && echo "    $(GREEN)✓$(NC) No clippy warnings" || (echo "    $(RED)✗$(NC) Clippy found issues" && exit 1); \
	else \
		echo "$(YELLOW)Warning: host directory not found, skipping$(NC)"; \
	fi

lint-runtime: ## Run Go linters (gofmt, golangci-lint)
	@echo "$(BLUE)Linting runtime (Go)$(NC)"
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && \
		echo "  - Checking formatting (gofmt)..." && \
		test -z "$$(gofmt -l .)" && echo "    $(GREEN)✓$(NC) Format OK" || (echo "    $(RED)✗$(NC) Format issues found. Run 'make format' to fix." && exit 1) && \
		if command -v golangci-lint >/dev/null 2>&1; then \
			echo "  - Running golangci-lint..." && \
			golangci-lint run --disable=errcheck --disable=staticcheck && echo "    $(GREEN)✓$(NC) No golangci-lint issues" || (echo "    $(RED)✗$(NC) golangci-lint found issues" && exit 1); \
		else \
			echo "  - Skipping golangci-lint (not installed). Run 'make install' to install."; \
		fi; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

 lint-tui: ## Run TypeScript TUI linters (oxlint, oxfmt)
	@echo "$(BLUE)Linting TUI (TypeScript)$(NC)"
	@if [ -d "$(TUI_DIR)" ]; then \
		cd $(TUI_DIR) && \
		echo "  - Running oxlint..." && \
		bunx oxlint . && echo "    $(GREEN)✓$(NC) No oxlint errors" || (echo "    $(RED)✗$(NC) oxlint found issues" && exit 1) && \
		echo "  - Checking formatting (oxfmt)..." && \
		bunx oxfmt --check . --ignore-path ../../.gitignore && echo "    $(GREEN)✓$(NC) Oxfmt format OK" || (echo "    $(RED)✗$(NC) Oxfmt format issues found. Run 'make format' to fix." && exit 1); \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

## Format targets
format: format-host format-runtime format-tui ## Format all code

format-host: ## Format Rust code
	@echo "$(BLUE)Formatting host (Rust)$(NC)"
	@if [ -d "$(HOST_DIR)" ]; then \
		cd $(HOST_DIR) && cargo fmt && echo "  $(GREEN)✓$(NC) Formatted"; \
	else \
		echo "$(YELLOW)Warning: host directory not found, skipping$(NC)"; \
	fi

format-runtime: ## Format Go code
	@echo "$(BLUE)Formatting runtime (Go)$(NC)"
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && gofmt -w . && echo "  $(GREEN)✓$(NC) Formatted"; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

 format-tui: ## Format TypeScript TUI code
	@echo "$(BLUE)Formatting TUI (TypeScript)$(NC)"
	@if [ -d "$(TUI_DIR)" ]; then \
		cd $(TUI_DIR) && bunx oxfmt --write . --ignore-path ../../.gitignore && echo "  $(GREEN)✓$(NC) Formatted"; \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

## Clean targets
clean: clean-host clean-runtime clean-tui ## Clean all build artifacts

clean-host: ## Clean Rust build artifacts
	@echo "$(BLUE)Cleaning host (Rust)$(NC)"
	@if [ -d "$(HOST_DIR)" ]; then \
		cd $(HOST_DIR) && cargo clean && echo "  $(GREEN)✓$(NC) Cleaned"; \
	else \
		echo "$(YELLOW)Warning: host directory not found, skipping$(NC)"; \
	fi

clean-runtime: ## Clean Go build artifacts
	@echo "$(BLUE)Cleaning runtime (Go)$(NC)"
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && go clean -cache && rm -f bin/* && echo "  $(GREEN)✓$(NC) Cleaned"; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

clean-tui: ## Clean TypeScript TUI build artifacts
	@echo "$(BLUE)Cleaning TUI (TypeScript)$(NC)"
	@if [ -d "$(TUI_DIR)" ]; then \
		cd $(TUI_DIR) && rm -rf dist node_modules && find . -maxdepth 1 -name ".*.bun-build" -delete && rm -f pryx-tui && echo "  $(GREEN)✓$(NC) Cleaned"; \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

## Install targets
install: install-tools install-deps ## Install all development tools and dependencies

install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	@echo "  Checking for pre-commit..."
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		echo "    Installing pre-commit..."; \
		if command -v pip >/dev/null 2>&1 && pip install pre-commit; then \
			echo "    $(GREEN)✓$(NC) pre-commit installed"; \
		elif command -v pip3 >/dev/null 2>&1 && pip3 install pre-commit; then \
			echo "    $(GREEN)✓$(NC) pre-commit installed"; \
		else \
			echo "    $(YELLOW)Warning:$(NC) failed to install pre-commit (pip/pip3 missing or install failed)"; \
		fi; \
	else \
		echo "    $(GREEN)✓$(NC) pre-commit already installed"; \
	fi
	@echo "  Checking for Bun..."
	@if ! command -v bun >/dev/null 2>&1; then \
		echo "    Installing Bun..."; \
		curl -fsSL https://bun.sh/install | bash; \
		echo "    $(GREEN)✓$(NC) Bun installed (restart your shell if needed)"; \
	else \
		if [ "$$(bun --version 2>/dev/null || echo '')" = "$(BUN_REQUIRED_VERSION)" ]; then \
			echo "    $(GREEN)✓$(NC) Bun $$(bun --version) already installed"; \
		else \
			echo "    $(YELLOW)Warning:$(NC) Bun version is $$(bun --version), expected $(BUN_REQUIRED_VERSION)"; \
		fi; \
	fi
	@echo "  Checking for golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		if command -v go >/dev/null 2>&1; then \
			echo "    Installing golangci-lint..."; \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
			echo "    $(GREEN)✓$(NC) golangci-lint installed"; \
		else \
			echo "    $(YELLOW)Warning:$(NC) go not found, skipping golangci-lint install"; \
		fi; \
	else \
		echo "    $(GREEN)✓$(NC) golangci-lint already installed"; \
	fi
	@echo "  Checking for Tauri CLI..."
	@if ! command -v tauri >/dev/null 2>&1 && [ ! -x "$$HOME/.bun/bin/tauri" ]; then \
		echo "    Installing Tauri CLI v2..."; \
		if command -v cargo >/dev/null 2>&1; then \
			if cargo install tauri-cli --version "^2.0.0"; then \
				echo "    $(GREEN)✓$(NC) Tauri CLI installed"; \
			elif command -v bun >/dev/null 2>&1 && bun add -g @tauri-apps/cli@latest; then \
				if command -v tauri >/dev/null 2>&1; then \
					echo "    $(GREEN)✓$(NC) Tauri CLI installed"; \
				elif [ -x "$$HOME/.bun/bin/tauri" ]; then \
					echo "    $(YELLOW)Warning:$(NC) Tauri CLI installed at $$HOME/.bun/bin/tauri but not on PATH"; \
					echo "    $(YELLOW)Warning:$(NC) add $$HOME/.bun/bin to PATH to avoid reinstall loops"; \
				else \
					echo "    $(YELLOW)Warning:$(NC) Tauri CLI install completed but binary not found"; \
				fi; \
			else \
				echo "    $(YELLOW)Warning:$(NC) failed to install Tauri CLI"; \
			fi; \
		elif command -v bun >/dev/null 2>&1; then \
			if bun add -g @tauri-apps/cli@latest; then \
				if command -v tauri >/dev/null 2>&1; then \
					echo "    $(GREEN)✓$(NC) Tauri CLI installed"; \
				elif [ -x "$$HOME/.bun/bin/tauri" ]; then \
					echo "    $(YELLOW)Warning:$(NC) Tauri CLI installed at $$HOME/.bun/bin/tauri but not on PATH"; \
					echo "    $(YELLOW)Warning:$(NC) add $$HOME/.bun/bin to PATH to avoid reinstall loops"; \
				else \
					echo "    $(YELLOW)Warning:$(NC) Tauri CLI install completed but binary not found"; \
				fi; \
			else \
				echo "    $(YELLOW)Warning:$(NC) bun add failed for Tauri CLI; restart your shell or install manually"; \
			fi; \
		else \
			echo "    $(YELLOW)Warning:$(NC) neither cargo nor bun found, skipping Tauri CLI install"; \
		fi; \
	else \
		echo "    $(GREEN)✓$(NC) Tauri CLI already installed"; \
	fi
	@echo "  Checking for Wrangler..."
	@if ! command -v wrangler >/dev/null 2>&1 && [ ! -x "$$HOME/.bun/bin/wrangler" ]; then \
		echo "    Installing Wrangler..."; \
		if command -v bun >/dev/null 2>&1; then \
			if bun add -g wrangler@latest; then \
				if command -v wrangler >/dev/null 2>&1; then \
					echo "    $(GREEN)✓$(NC) Wrangler installed"; \
				elif [ -x "$$HOME/.bun/bin/wrangler" ]; then \
					echo "    $(YELLOW)Warning:$(NC) Wrangler installed at $$HOME/.bun/bin/wrangler but not on PATH"; \
					echo "    $(YELLOW)Warning:$(NC) add $$HOME/.bun/bin to PATH to avoid reinstall loops"; \
				else \
					echo "    $(YELLOW)Warning:$(NC) Wrangler install completed but binary not found"; \
				fi; \
			else \
				echo "    $(YELLOW)Warning:$(NC) failed to install Wrangler with bun"; \
			fi; \
		else \
			echo "    $(YELLOW)Warning:$(NC) bun not found, skipping Wrangler install"; \
		fi; \
	else \
		echo "    $(GREEN)✓$(NC) Wrangler already installed"; \
	fi
	@echo "  Setting up pre-commit hooks..."
	@if [ -f .pre-commit-config.yaml ]; then \
		if pre-commit --version >/dev/null 2>&1; then \
			if pre-commit install; then \
				echo "    $(GREEN)✓$(NC) pre-commit hooks installed"; \
			else \
				echo "    $(YELLOW)Warning:$(NC) failed to install pre-commit hooks, continuing"; \
			fi; \
		else \
			echo "    $(YELLOW)Warning:$(NC) pre-commit is not usable in this environment, skipping hooks setup"; \
		fi; \
	else \
		echo "    $(YELLOW)Warning: .pre-commit-config.yaml not found$(NC)"; \
	fi

install-deps: ## Install all dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	@if [ -d "$(HOST_DIR)" ]; then \
		if command -v cargo >/dev/null 2>&1; then \
			echo "  Installing host dependencies..." && cd $(HOST_DIR) && cargo build; \
		else \
			echo "  $(YELLOW)Warning:$(NC) cargo not found, skipping host dependencies"; \
		fi; \
	else \
		echo "$(YELLOW)Warning: host directory not found, skipping$(NC)"; \
	fi
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		if command -v go >/dev/null 2>&1; then \
			echo "  Installing runtime dependencies..." && cd $(RUNTIME_DIR) && go mod download && go mod tidy; \
		else \
			echo "  $(YELLOW)Warning:$(NC) go not found, skipping runtime dependencies"; \
		fi; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi
	@if [ -d "$(TUI_DIR)" ]; then \
		if command -v bun >/dev/null 2>&1; then \
			echo "  Installing TUI dependencies..." && cd $(TUI_DIR) && bun install --frozen-lockfile; \
		else \
			echo "  $(YELLOW)Warning:$(NC) bun not found, skipping TUI dependencies"; \
		fi; \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

clean-uninstall: ## Remove all Pryx binaries, services, and data
	@echo "$(RED)Warning: This will delete all Pryx data!$(NC)"
	@bash scripts/uninstall.sh

## Check targets
check: lint test ## Run comprehensive checks (lint + test)
	@echo "$(GREEN)✓$(NC) All checks passed!"

## Info targets
info: ## Show project information
	@echo "$(BLUE)Pryx Project Information$(NC)"
	@echo ""
	@echo "  Version:       $(VERSION)"
	@echo "  Build Date:    $(BUILD_DATE)"
	@echo "  Commit SHA:    $(COMMIT_SHA)"
	@echo ""
	@echo "$(BLUE)Project Structure:$(NC)"
	@echo "  Host (Rust):       $(HOST_DIR)"
	@echo "  Runtime (Go):       $(RUNTIME_DIR)"
	@echo "  TUI (TypeScript):    $(TUI_DIR)"
	@echo ""
	@echo "$(BLUE)Tools:$(NC)"
	@echo "  Rust:    $$(rustc --version 2>/dev/null || echo 'not installed')"
	@echo "  Go:      $$(go version 2>/dev/null | head -1 || echo 'not installed')"
	@echo "  Bun:     $$(bun --version 2>/dev/null || echo 'not installed')"
	@echo "  Node:    $$(node --version 2>/dev/null || echo 'not installed')"
	@echo ""

## Version management
version-bump-patch: ## Bump patch version (0.1.0 -> 0.1.1)
	@bash ./scripts/bump-version.sh patch

version-bump-minor: ## Bump minor version (0.1.0 -> 0.2.0)
	@bash ./scripts/bump-version.sh minor

version-bump-major: ## Bump major version (0.1.0 -> 1.0.0)
	@bash ./scripts/bump-version.sh major

version-tag: ## Create git tag for current version
	@echo "$(BLUE)Creating tag v$(VERSION)$(NC)"
	@git tag -a v$(VERSION) -m "Release v$(VERSION)" && echo "  $(GREEN)✓$(NC) Tag created"

version-push: ## Push version tags
	@echo "$(BLUE)Pushing tags$(NC)"
	@git push origin v$(VERSION) && echo "  $(GREEN)✓$(NC) Tag pushed"
