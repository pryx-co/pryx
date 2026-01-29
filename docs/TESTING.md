# Pryx Testing Strategy

This document outlines the comprehensive testing strategy for Pryx, covering unit, integration, and E2E tests across all components.

## Test Levels

### 1. Unit Tests
Fast, isolated tests that verify individual functions and components.

**Go Runtime:**
- Location: Colocated with source files (`*_test.go`)
- Focus: Pure functions, business logic, data transformations
- Tools: Standard `testing` package + `testify/assert`

**Rust Host:**
- Location: Colocated with source files (`#[cfg(test)]` modules)
- Focus: Sidecar state management, port extraction, RPC handling
- Tools: Standard Rust test framework

**TypeScript TUI:**
- Location: Colocated with source files (`*.test.ts`)
- Focus: Component logic, service methods, utilities
- Tools: `bun:test`

### 2. Integration Tests
Tests that verify component interactions and API contracts.

**Go Runtime:**
- Location: `apps/runtime/tests/integration/`
- Focus: HTTP handlers, WebSocket lifecycle, MCP calls
- Tools: `httptest`, test servers

**Rust Host:**
- Location: `apps/host/tests/integration/`
- Focus: Sidecar spawning, RPC communication
- Tools: Process spawning, temp directories

**TypeScript TUI:**
- Location: `apps/tui/tests/integration/`
- Focus: Service integration with runtime
- Tools: Mock WebSocket servers

### 3. E2E Tests
Full system tests that verify end-to-end workflows.

**CLI E2E:**
- Location: `apps/runtime/tests/e2e/`
- Focus: Complete CLI commands with real runtime
- Tools: Process spawning, temp directories

**TUI E2E:**
- Location: `apps/tui/tests/e2e/`
- Focus: Full TUI interactions
- Tools: Playwright

## Test Organization

```
apps/
├── runtime/
│   ├── internal/
│   │   ├── bus/
│   │   │   └── bus_test.go              # Unit tests
│   │   ├── server/
│   │   │   └── server_test.go           # Unit tests
│   │   └── ...
│   ├── tests/
│   │   ├── integration/                 # Integration tests
│   │   │   └── runtime_test.go
│   │   └── e2e/                         # E2E tests
│   │       └── cli_test.go
│   └── testutils/                       # Shared test utilities
│       └── helpers.go
├── host/
│   ├── src/
│   │   └── sidecar/
│   │       └── tests.rs                 # Unit tests
│   └── tests/
│       └── integration/
├── tui/
│   ├── src/
│   │   ├── services/
│   │   │   └── ws.test.ts               # Unit tests
│   │   └── components/
│   │       └── App.test.tsx
│   └── tests/
│       ├── integration/
│       └── e2e/
│           └── tui.e2e.ts
```

## Running Tests

### Run All Tests
```bash
make test
```

### Run Unit Tests Only
```bash
make test-unit
```

### Run Integration Tests
```bash
make test-integration
```

### Run E2E Tests
```bash
make test-e2e
```

### Run with Coverage
```bash
make test-coverage
```

### Run Specific Component
```bash
make test-runtime      # Go runtime tests
make test-host         # Rust host tests
make test-tui          # TUI tests
```

## Test Coverage Goals

- **Unit Tests**: 80%+ coverage of business logic
- **Integration Tests**: All major API endpoints and flows
- **E2E Tests**: Critical user journeys

## Testing Best Practices

1. **Table-Driven Tests**: Use for testing multiple scenarios in Go
2. **Deterministic Tests**: Tests should produce the same results every run
3. **Fast Tests**: Unit tests should complete in milliseconds
4. **Isolated Tests**: No shared state between tests
5. **Clear Names**: Test names should describe the behavior being tested
6. **Assert Over Verify**: Prefer assertions over mock verification
7. **Test Helpers**: Extract common setup/teardown into helpers

## Mocking Strategy

### Go
- Use interfaces for dependencies
- Create mock implementations in test files
- Use `httptest` for HTTP mocking

### Rust
- Use trait bounds for dependencies
- Mock implementations in `#[cfg(test)]` modules

### TypeScript
- Mock external services (WebSocket, filesystem)
- Use dependency injection for testability

## Continuous Integration

All tests run on:
- Every PR (unit + integration)
- Main branch (unit + integration + E2E)
- Release tags (full test suite)

## Test Data

- Use test fixtures for complex data
- Generate test data programmatically when possible
- Clean up test data after tests

## Debugging Tests

### Go
```bash
cd apps/runtime
go test -v ./internal/bus -run TestBus
go test -v ./... 2>&1 | head -50
```

### Rust
```bash
cd apps/host
cargo test --lib -- --nocapture
cargo test sidecar -- --nocapture
```

### TypeScript
```bash
cd apps/tui
bun test --verbose
bun test src/services/ws.test.ts
```
