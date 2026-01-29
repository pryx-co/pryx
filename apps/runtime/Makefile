# Pryx Runtime Makefile
# Unified build and test commands for Pryx agent runtime

.PHONY: all build test test-go test-node test-unit test-integration test-e2e test-coverage test-all test-verbose test-short clean help run dev install
.PHONY: test-config test-bus test-agent test-llm test-server test-mesh test-telemetry
.PHONY: test-skills test-mcp test-audit test-cost test-memory
.PHONY: test-report test-check-coverage test-benchmark test-parallel test-stats test-coverage-summary

# Default target
all: build

# Build the pryx-core binary
build:
	@echo "Building pryx-core..."
	go build -o bin/pryx-core ./cmd/pryx-core/
	@echo "Build complete: bin/pryx-core"

# Install binary to ~/.pryx/bin
install: build
	@echo "Installing to ~/.pryx/bin..."
	mkdir -p ~/.pryx/bin
	cp bin/pryx-core ~/.pryx/bin/
	@echo "Installation complete"

# Run all tests (Go unit + Node E2E)
test: test-go test-node
	@echo "All tests complete!"

# Run Go tests only (unit + integration)
test-go:
	@echo "Running Go tests..."
	go test -v -race ./internal/... ./tests/... ./e2e/... 2>&1 | tee test_output.txt
	@echo "Go tests complete"

# Run Go unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -race ./internal/... 2>&1 | tee test_output.txt

# Run Go integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/... 2>&1 | tee -a test_output.txt

# Run Go E2E tests only
test-e2e:
	@echo "Running Go E2E tests..."
	go test -v ./e2e/... 2>&1 | tee -a test_output.txt

# Run Node.js E2E test suite
test-node:
	@echo "Running Node.js E2E tests..."
	@if [ -f scripts/e2e-test-suite.js ]; then \
		npm list ws >/dev/null 2>&1 || npm install ws --save-dev; \
		node scripts/e2e-test-suite.js; \
	else \
		echo "E2E test suite not found at scripts/e2e-test-suite.js"; \
		exit 1; \
	fi

# Run all tests with coverage
test-coverage: build
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./internal/... ./tests/... ./e2e/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run all tests without caching
test-all: build
	@echo "Running all tests (no cache)..."
	go clean -testcache
	go test -v -race ./internal/... ./tests/... ./e2e/... 2>&1 | tee test_output.txt
	node scripts/e2e-test-suite.js

# Run tests with verbose output
test-verbose: build
	go test -v -race ./internal/... ./tests/... ./e2e/...

# Run tests with short flag (skip slow tests)
test-short: build
	go test -short -v ./internal/... ./tests/... ./e2e/...

# Run tests for specific packages
test-config:
	go test -v ./internal/config/...

test-bus:
	go test -v ./internal/bus/...

test-agent:
	go test -v ./internal/agent/...

test-llm:
	go test -v ./internal/llm/...

test-server:
	go test -v ./internal/server/...

test-mesh:
	go test -v ./internal/mesh/...

test-telemetry:
	go test -v ./internal/telemetry/...

# Run specific CLI tests
test-skills:
	go test -v ./tests/cli/skills/...

test-mcp:
	go test -v ./tests/cli/mcp/...

test-audit:
	go test -v ./tests/cli/audit/...

test-cost:
	go test -v ./tests/cli/cost/...

test-memory:
	go test -v ./tests/cli/memory/...

# Generate test report
test-report: test-all
	@echo "## Test Results" > TEST_REPORT.md
	@echo "" >> TEST_REPORT.md
	@echo "Generated: $$(date)" >> TEST_REPORT.md
	@echo "" >> TEST_REPORT.md
	@echo "### Summary" >> TEST_REPORT.md
	@tail -20 test_output.txt >> TEST_REPORT.md 2>/dev/null || echo "No test output" >> TEST_REPORT.md
	@echo "" >> TEST_REPORT.md
	@echo "### Full Output" >> TEST_REPORT.md
	@echo '```' >> TEST_REPORT.md
	@cat test_output.txt >> TEST_REPORT.md 2>/dev/null
	@echo '```' >> TEST_REPORT.md
	@echo "Test report generated: TEST_REPORT.md"

# Run tests and check coverage thresholds
test-check-coverage: build
	go test -coverprofile=coverage.out ./internal/... ./tests/... ./e2e/...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ "$${COVERAGE%.*}" -lt "50" ]; then \
		echo "WARNING: Coverage below 50% threshold"; \
	fi

# Benchmark tests
test-benchmark:
	go test -bench=. -benchmem ./internal/... ./tests/... ./e2e/...

# Run tests in parallel
test-parallel:
	go test -parallel=4 -v ./internal/... ./tests/... ./e2e/...

# Development commands
dev:
	go run ./cmd/pryx-core

# Run with race detector
run:
	go run -race ./cmd/pryx-core

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html test_output.txt TEST_REPORT.md
	go clean -testcache

# Show test statistics
test-stats:
	@echo "Test Statistics:"
	@echo "==============="
	@echo ""
	@echo "Unit tests (internal):"
	@find ./internal -name "*_test.go" | wc -l | xargs echo "  - Test files:"
	@echo ""
	@echo "Integration tests:"
	@find ./tests -name "*_test.go" | wc -l | xargs echo "  - Test files:"
	@echo ""
	@echo "E2E tests:"
	@find ./e2e -name "*_test.go" | wc -l | xargs echo "  - Go E2E test files:"
	@echo "  - Node E2E test files: 1 (scripts/e2e-test-suite.js)"

# Show coverage summary
test-coverage-summary: test-coverage
	@echo ""
	@echo "Coverage Summary:"
	@echo "================="
	@go tool cover -func=coverage.out | grep -E "(coverage:|total)" | tail -1

# Help target
help:
	@echo "Pryx Runtime - Available Commands:"
	@echo ""
	@echo "BUILD:"
	@echo "  make build             - Build pryx-core binary"
	@echo "  make install           - Install to ~/.pryx/bin"
	@echo "  make clean             - Clean build artifacts"
	@echo ""
	@echo "TEST (All):"
	@echo "  make test              - Run all tests (Go + Node E2E)"
	@echo "  make test-go           - Run Go tests only"
	@echo "  make test-node         - Run Node.js E2E tests only"
	@echo "  make test-all          - Run all tests without cache"
	@echo ""
	@echo "TEST (Categories):"
	@echo "  make test-unit         - Run unit tests"
	@echo "  make test-integration  - Run integration tests"
	@echo "  make test-e2e          - Run Go E2E tests"
	@echo ""
	@echo "TEST (Specific):"
	@echo "  make test-config       - Test config package"
	@echo "  make test-bus          - Test bus package"
	@echo "  make test-agent        - Test agent package"
	@echo "  make test-llm          - Test LLM packages"
	@echo "  make test-server       - Test server package"
	@echo "  make test-mesh         - Test mesh package"
	@echo "  make test-telemetry    - Test telemetry package"
	@echo "  make test-skills       - Test skills CLI"
	@echo "  make test-mcp          - Test MCP CLI"
	@echo "  make test-cost         - Test cost CLI"
	@echo ""
	@echo "TEST (Coverage):"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-coverage-summary - Show coverage summary"
	@echo "  make test-check-coverage - Check coverage threshold"
	@echo ""
	@echo "TEST (Other):"
	@echo "  make test-verbose      - Verbose test output"
	@echo "  make test-short        - Skip slow tests"
	@echo "  make test-benchmark    - Run benchmarks"
	@echo "  make test-parallel     - Run in parallel"
	@echo "  make test-stats        - Show test statistics"
	@echo "  make test-report       - Generate test report"
	@echo ""
	@echo "DEVELOPMENT:"
	@echo "  make dev               - Run in development mode"
	@echo "  make run               - Run with race detector"
	@echo "  make help              - Show this help"
