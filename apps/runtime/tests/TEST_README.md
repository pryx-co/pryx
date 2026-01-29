# Pryx Runtime Test Configuration

## Test Structure

```
apps/runtime/
├── tests/
│   ├── test_main.go              # Test setup and teardown
│   ├── cli/                      # CLI unit tests
│   │   ├── skills/
│   │   │   └── list_test.go     # Skills CLI tests (13 tests)
│   │   ├── mcp/
│   │   │   └── list_test.go     # MCP CLI tests (13 tests)
│   │   ├── audit/
│   │   │   └── query_test.go    # Audit CLI tests (15 tests)
│   │   ├── cost/
│   │   │   └── summary_test.go  # Cost CLI tests (16 tests)
│   │   └── memory/
│   │       └── usage_test.go    # Memory CLI tests (18 tests)
│   └── integration/
│       └── workflow_test.go     # Integration tests (9 tests)
├── e2e/
│   └── cli_test.go              # E2E CLI tests (11 tests)
├── integration/
│   └── memory_test.go           # Existing integration tests (7 tests)
└── Makefile.test                # Test Makefile
```

## Test Coverage Goals

- **Unit tests:** 90%+ coverage per module
- **Integration tests:** 80%+ coverage for module interactions
- **E2E tests:** 100% CLI command coverage

## Running Tests

```bash
# Build the binary first
make -f Makefile.test build

# Run all tests
make -f Makefile.test test

# Run specific test types
make -f Makefile.test test:unit      # Unit tests only
make -f Makefile.test test:integration # Integration tests only
make -f Makefile.test test:e2e       # E2E tests only

# Run tests with coverage
make -f Makefile.test test:coverage

# Run tests for specific module
make -f Makefile.test test:skills    # Skills tests
make -f Makefile.test test:mcp       # MCP tests
make -f Makefile.test test:audit     # Audit tests
make -f Makefile.test test:cost      # Cost tests
make -f Makefile.test test:memory    # Memory tests

# Generate test report
make -f Makefile.test test:report

# Check coverage thresholds
make -f Makefile.test test:check-coverage
```

## Test Categories

### CLI Unit Tests (75 tests)

Each CLI module has comprehensive tests covering:
- Command execution
- Flag handling (--json, --eligible, --verbose, etc.)
- Error handling
- Edge cases
- Help output validation

### Integration Tests (9 tests)

Integration tests verify:
- Module interactions (Memory + Session, CLI + Runtime)
- Full workflows (message creation → memory usage → summarization)
- Multiple session handling
- Archive/unarchive workflows
- Warning threshold behavior

### E2E Tests (11 tests)

End-to-end tests validate:
- Full CLI command execution against built binary
- Exit code handling
- Output format validation
- Error message consistency

## Test Environment

Tests use:
- In-memory SQLite database (`:memory:`)
- Built binary at `/tmp/pryx-core`
- Temporary directories for test data
- Environment variables for test mode

## Adding New Tests

### CLI Unit Test Template

```go
package cli_module

import (
	"os/exec"
	"strings"
	"testing"
)

func TestModuleCommand(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "module", "command")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("module command failed: %v", err)
	}

	if !strings.Contains(string(output), "Expected") {
		t.Errorf("Expected content in output, got: %s", output)
	}
}
```

### Integration Test Template

```go
package integration

import (
	"context"
	"testing"

	"pryx-core/internal/bus"
	"pryx-core/internal/store"
)

func TestFeatureIntegration(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	b := bus.New()
	// Create managers and test interactions
	ctx := context.Background()

	// Test implementation
}
```

## Coverage Requirements

All new features must include:
- Unit tests for all public functions
- Integration tests for module interactions
- E2E tests for CLI commands
- Edge case and error handling tests

Coverage should not drop below 80% for any module.
