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
.PHONY: help build test lint clean install format check
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
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Pryx Build System$(NC)"
	@echo ""
	@echo "$(BLUE)Available targets:$(NC)"
	@echo "  $(GREEN)build$(NC)             # Build all components"
	@echo "  $(GREEN)build-host$(NC)        # Build Rust/Tauri host"
	@echo "  $(GREEN)build-runtime$(NC)     # Build Go runtime"
	@echo "  $(GREEN)build-tui$(NC)         # Build TypeScript TUI"
	@echo "  $(GREEN)test$(NC)              # Run all tests"
	@echo "  $(GREEN)lint$(NC)              # Run all linters"
	@echo "  $(GREEN)format$(NC)            # Format all code"
	@echo "  $(GREEN)clean$(NC)             # Clean build artifacts"
	@echo "  $(GREEN)install$(NC)           # Install development tools"
	@echo "  $(GREEN)info$(NC)              # Show project information"
	@echo ""
	@echo "$(BLUE)Version Management:$(NC)"
	@echo "  $(GREEN)version-bump-patch$(NC)  # Bump patch version"
	@echo "  $(GREEN)version-bump-minor$(NC)  # Bump minor version"
	@echo "  $(GREEN)version-bump-major$(NC)  # Bump major version"
	@echo "  $(GREEN)version-tag$(NC)       # Create git tag"
	@echo "  $(GREEN)version-push$(NC)      # Push tags"
	@echo ""
	@echo "$(BLUE)Usage:$(NC)"
	@echo "  make <target>"

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
		go build -o bin/pryx-core -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)" ./cmd/pryx-core; \
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
test: test-host test-runtime test-tui ## Run all tests

test-host: ## Run Rust host tests
	@echo "$(BLUE)Testing host (Rust)$(NC)"
	@if [ -d "$(HOST_DIR)" ]; then \
		cd $(HOST_DIR) && cargo test --release --lib; \
	else \
		echo "$(YELLOW)Warning: host directory not found, skipping$(NC)"; \
	fi

test-runtime: ## Run Go runtime tests
	@echo "$(BLUE)Testing runtime (Go)$(NC)"
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		cd $(RUNTIME_DIR) && go test -v -race -cover ./...; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

test-tui: ## Run TypeScript TUI checks
	@echo "$(BLUE)Testing TUI (TypeScript)$(NC)"
	@if [ -d "$(TUI_DIR)" ]; then \
		cd $(TUI_DIR) && bun install --frozen-lockfile && bun run build:ci; \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

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
			golangci-lint run && echo "    $(GREEN)✓$(NC) No golangci-lint issues" || (echo "    $(RED)✗$(NC) golangci-lint found issues" && exit 1); \
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
		bunx oxfmt --check . && echo "    $(GREEN)✓$(NC) Oxfmt format OK" || (echo "    $(RED)✗$(NC) Oxfmt format issues found. Run 'make format' to fix." && exit 1); \
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
		cd $(TUI_DIR) && bunx oxfmt --write . && echo "  $(GREEN)✓$(NC) Formatted"; \
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
		cd $(RUNTIME_DIR) && go clean -cache -modcache -i && rm -f bin/* && echo "  $(GREEN)✓$(NC) Cleaned"; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi

clean-tui: ## Clean TypeScript TUI build artifacts
	@echo "$(BLUE)Cleaning TUI (TypeScript)$(NC)"
	@if [ -d "$(TUI_DIR)" ]; then \
		cd $(TUI_DIR) && rm -rf dist node_modules && rm -f pryx-tui && echo "  $(GREEN)✓$(NC) Cleaned"; \
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
		pip install pre-commit || pip3 install pre-commit; \
		echo "    $(GREEN)✓$(NC) pre-commit installed"; \
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
		echo "    Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
		echo "    $(GREEN)✓$(NC) golangci-lint installed"; \
	else \
		echo "    $(GREEN)✓$(NC) golangci-lint already installed"; \
	fi
	@echo "  Checking for Tauri CLI..."
	@if ! command -v tauri >/dev/null 2>&1; then \
		echo "    Installing Tauri CLI v2..."; \
		cargo install tauri-cli --version "^2.0.0" || bun add -g @tauri-apps/cli@latest; \
		echo "    $(GREEN)✓$(NC) Tauri CLI installed"; \
	else \
		echo "    $(GREEN)✓$(NC) Tauri CLI already installed"; \
	fi
	@echo "  Checking for Wrangler..."
	@if ! command -v wrangler >/dev/null 2>&1; then \
		echo "    Installing Wrangler..."; \
		bun add -g wrangler@latest; \
		echo "    $(GREEN)✓$(NC) Wrangler installed"; \
	else \
		echo "    $(GREEN)✓$(NC) Wrangler already installed"; \
	fi
	@echo "  Setting up pre-commit hooks..."
	@if [ -f .pre-commit-config.yaml ]; then \
		pre-commit install; \
		echo "    $(GREEN)✓$(NC) pre-commit hooks installed"; \
	else \
		echo "    $(YELLOW)Warning: .pre-commit-config.yaml not found$(NC)"; \
	fi

install-deps: ## Install all dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	@if [ -d "$(HOST_DIR)" ]; then \
		echo "  Installing host dependencies..." && cd $(HOST_DIR) && cargo build; \
	else \
		echo "$(YELLOW)Warning: host directory not found, skipping$(NC)"; \
	fi
	@if [ -d "$(RUNTIME_DIR)" ]; then \
		echo "  Installing runtime dependencies..." && cd $(RUNTIME_DIR) && go mod download && go mod tidy; \
	else \
		echo "$(YELLOW)Warning: runtime directory not found, skipping$(NC)"; \
	fi
	@if [ -d "$(TUI_DIR)" ]; then \
		echo "  Installing TUI dependencies..." && cd $(TUI_DIR) && bun install --frozen-lockfile; \
	else \
		echo "$(YELLOW)Warning: tui directory not found, skipping$(NC)"; \
	fi

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
