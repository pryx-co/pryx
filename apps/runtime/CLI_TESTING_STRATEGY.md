# Pryx CLI E2E & Integration Test Strategy

> Based on beads task implementation and Moltbot testing patterns
> Last Updated: 2026-01-29

## Executive Summary

This document provides a comprehensive E2E and Integration testing strategy for all Pryx CLI features implemented in the beads task, following patterns from reference projects (Moltbot, opencode).

## Scope

**Focus Areas:**
1. ✅ CLI Commands - Complete E2E coverage for all implemented CLI commands
2. ✅ Integration Tests - Core service integration validation
3. ✅ Test Infrastructure - Reusable helpers and fixtures

**CLI Commands to Test:**
- `config` - Configuration management (set/get/list)
- `cost` - Cost tracking (summary/daily/monthly/budget/pricing/optimize)
- `mcp` - MCP server management (list/add/remove/test/auth)
- `skills` - Skill management (list/info/check/enable/disable/install/uninstall)
- `doctor` - Health diagnostics (all checks)

**Excluded (Future Work):**
- `audit` - CLI command not yet implemented
- `login` - OAuth device flow (commented out)
- `auth` - Authentication manager

---

## Testing Philosophy

Based on Moltbot and opencode patterns:

### 1. E2E Tests (End-to-End)
- **Purpose**: Validate entire command flows from entry point to exit
- **Pattern**: Spawn real processes, execute CLI, validate output/exit code
- **Cleanup**: Always cleanup (afterAll hooks to kill processes, delete temp files)
- **Timeouts**: Generous timeouts (E2E: 120s, operations: 10s)
- **Isolation**: Each test gets its own config/database directory

### 2. Integration Tests (In-Process)
- **Purpose**: Validate component integration without external dependencies
- **Pattern**: Direct Go testing with in-memory database and mock services
- **Speed**: Fast (<1s per test)
- **Coverage**: Unit test all code paths, integration test service boundaries

### 3. Test Organization

```
apps/runtime/e2e/                    # E2E tests (spawn processes)
├── cli_test.go                      # Existing CLI tests
├── config_cli_test.go               # Config command tests
├── cost_cli_test.go                 # Cost command tests
├── mcp_cli_test.go                  # MCP command tests
├── skills_cli_test.go                # Skills command tests
├── doctor_cli_test.go                # Doctor command tests
├── helpers.go                       # Test helpers
└── setup.go                         # Test setup fixtures

apps/runtime/integration/              # Integration tests
├── llm_audit_test.go              # LLM + Audit integration
├── mcp_policy_test.go              # MCP + Policy integration
├── session_memory_cost_test.go    # Session + Memory + Cost integration
├── channels_session_test.go         # Channels + Sessions integration
└── performance_test.go             # Performance benchmarks
```

---

## Test Patterns from Reference Projects

### Pattern 1: Process Spawning for E2E (from Moltbot)

```go
// E2E test with process spawning
func TestCommand_E2E(t *testing.T) {
    // Get unique port for this test
    port := getFreePort()

    // Create temporary directory for this test
    tmpDir, _ := ioutil.TempDir("pryx-e2e-" + t.Name())
    defer os.RemoveAll(tmpDir)

    // Spawn the process with specific config
    cmd := exec.Command(binaryPath,
        "--port", port,
        "--config", filepath.Join(tmpDir, "config.json"),
    )

    // Collect output for validation
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    // Execute with timeout
    err := cmd.Start()
    if err != nil {
        t.Fatalf("Failed to start process: %v", err)
    }

    done := make(chan error, 1)
    go func() {
        done <- cmd.Wait()
    }()

    select {
    case <-done:
        return
    case <-time.After(time.Second * 10):
        cmd.Process.Kill()
        t.Fatalf("Process timeout after 10s")
    }

    // Validate exit code and output
    if cmd.ProcessState.ExitCode() != 0 {
        t.Errorf("Process exited with code %d", cmd.ProcessState.ExitCode())
    }

    output := stdout.String()
    if !strings.Contains(output, "Expected text") {
        t.Errorf("Output missing expected text. Got: %s", output)
    }

    // Clean up: kill process, remove temp dir
    cmd.Process.Kill()
}
```

**Key learnings from Moltbot:**
- Use `time.After` with a channel for timeout
- Always `defer` cleanup (even if test fails)
- Parse output as JSON when appropriate
- Use `select` for race condition (process done vs timeout)

### Pattern 2: Direct Command Execution for Speed (from Moltbot)

```go
// Integration test - direct function calls
func TestService_Integration(t *testing.T) {
    // Create in-memory database
    db := setupInMemoryDB(t)
    defer db.Close()

    // Create event bus
    bus := bus.New()

    // Initialize service with mock dependencies
    mockAuditRepo := audit.NewAuditRepository(db.DB)
    mockStore := store.New(db)
    mockPricingMgr := cost.NewPricingManager()

    // Create service
    service := cost.NewService(
        cost.NewCostTracker(mockAuditRepo, mockPricingMgr),
        cost.NewCostCalculator(mockPricingMgr),
        mockPricingMgr,
        mockStore,
    )

    // Test service method
    result, err := service.GetCurrentSessionCost()
    if err != nil {
        t.Fatalf("Service failed: %v", err)
    }

    // Validate result
    if result.TotalCost < 0 {
        t.Error("Expected positive cost")
    }
}
```

**Key learnings:**
- Use in-memory SQLite for fast tests
- Mock external dependencies (HTTP, file system)
- Test happy paths AND error paths
- Use table-driven tests for multiple scenarios

### Pattern 3: Test Helpers (from Moltbot)

```go
// Reusable test helpers
type TestEnv struct {
    BinaryPath  string
    ConfigDir   string
    DBPath      string
}

func SetupTestEnv(t *testing.T) TestEnv {
    // Create temp directory
    tmpDir := t.TempDir()
    os.MkdirAll(filepath.Join(tmpDir, "config"), 0755)

    // Set up environment variables
    env := map[string]string{
        "CLAWDBOT_CONFIG_PATH": filepath.Join(tmpDir, "config.json"),
        "CLAWDBOT_STATE_DIR": tmpDir,
    }

    return TestEnv{
        BinaryPath:  filepath.Join("apps", "runtime", "pryx-core"),
        ConfigDir:   tmpDir,
        DBPath:      filepath.Join(tmpDir, "test.db"),
        Env:         env,
    }
}

func CleanupTestEnv(env TestEnv) {
    // Kill any running processes
    if env.Process != nil {
        env.Process.Kill()
    }

    // Remove temp directory
    os.RemoveAll(env.ConfigDir)
}
```

**Key learnings:**
- Create reusable helpers for common test setup
- Isolate each test's environment
- Ensure cleanup happens even on test failure
- Use table-driven tests for multiple scenarios

---

## Implementation Plan

### Phase 1: Test Infrastructure (Priority: HIGH)

| Component | Files | Tests | Status |
|-----------|-------|-------|--------|
| Test Helpers | `e2e/helpers.go` | setup, cleanup, temp management | TODO |
| E2E Base | `e2e/setup.go` | port management, spawn wrappers | TODO |

**Goal**: Create reusable test infrastructure that all E2E tests can use.

### Phase 2: CLI E2E Tests (Priority: HIGH)

| Command | File | Tests | Status |
|---------|------|-------|--------|
| Config | `e2e/config_cli_test.go` | get, set, list, error handling | PARTIAL |
| Cost | `e2e/cost_cli_test.go` | summary, daily, monthly, budget, pricing, optimize | COMPLETE |
| MCP | `e2e/mcp_cli_test.go` | list, add (HTTP/stdio), remove, test, auth | PARTIAL |
| Skills | `e2e/skills_cli_test.go` | list, info, check, enable, disable, install, uninstall | PARTIAL |
| Doctor | `e2e/doctor_cli_test.go` | all health checks | COMPLETE |

**Goal**: 100% coverage of all CLI commands with E2E tests.

### Phase 3: Integration Tests (Priority: MEDIUM)

| Integration | File | Tests | Status |
|------------|------|-------|--------|
| LLM + Audit | `integration/llm_audit_test.go` | audit logging on LLM calls | TODO |
| MCP + Policy | `integration/mcp_policy_test.go` | tool requests with policy enforcement | TODO |
| Session + Memory + Cost | `integration/session_memory_cost_test.go` | lifecycle with all three services | TODO |
| Channels + Sessions | `integration/channels_session_test.go` | channel messages stored in sessions | TODO |

**Goal**: Validate cross-component interactions at the service level.

---

## Test Coverage Requirements

### Config Command Tests

| Test Case | Description | Priority |
|-----------|-------------|----------|
| `config get <key>` | Retrieve single config value | P0 |
| `config set <key> <value>` | Set config value | P0 |
| `config list` | List all config values | P0 |
| Error: unknown key | Handle non-existent config key gracefully | P1 |
| Error: invalid value type | Validate value types | P2 |
| JSON output | Verify `--json` flag produces valid JSON | P2 |

### Cost Command Tests (IMPLEMENTED)

| Test Case | Description | Status |
|-----------|-------------|--------|
| `cost summary` | Show total cost breakdown | ✅ |
| `cost daily` | Show daily costs | ✅ |
| `cost monthly` | Show monthly costs | ✅ |
| `cost budget status` | Show current budget status | ✅ |
| `cost budget set --daily X --monthly Y` | Set budget limits | ✅ |
| `cost pricing` | Show model pricing table | ✅ |
| `cost optimize` | Show optimization suggestions | ✅ |
| Error: no data | Handle empty cost data gracefully | ✅ |

### MCP Command Tests

| Test Case | Description | Priority |
|-----------|-------------|----------|
| `mcp list` | List configured MCP servers | P0 |
| `mcp list --json` | JSON output of servers | P0 |
| `mcp add <name> --url <url>` | Add HTTP MCP server | P0 |
| `mcp add <name> --cmd <command>` | Add stdio MCP server | P0 |
| `mcp add <name> --auth bearer --token-ref <ref>` | Add server with auth | P1 |
| `mcp remove <name>` | Remove MCP server | P0 |
| `mcp test <name>` | Test MCP server configuration | P0 |
| `mcp auth <name>` | Show auth configuration | P1 |
| Error: duplicate name | Handle duplicate server name | P1 |
| Error: invalid URL | Validate URL format | P2 |
| Error: invalid command | Validate stdio command | P2 |

### Skills Command Tests

| Test Case | Description | Priority |
|-----------|-------------|----------|
| `skills list` | List available skills | P0 |
| `skills list --eligible` | List only eligible skills | P0 |
| `skills list --json` | JSON output of skills | P0 |
| `skills info <name>` | Show skill details | P0 |
| `skills check` | Validate all skills for issues | P0 |
| `skills enable <name>` | Enable a skill | P0 |
| `skills disable <name>` | Disable a skill | P0 |
| `skills install <name>` | Prepare skill for use | P1 |
| `skills uninstall <name>` | Remove skill | P1 |
| Error: unknown skill | Handle non-existent skill gracefully | P1 |
| Error: install failure | Handle skill install errors | P2 |

### Doctor Command Tests (IMPLEMENTED)

| Test Case | Description | Status |
|-----------|-------------|--------|
| `doctor` | Run all health checks | ✅ |
| Database check | Verify SQLite database integrity | ✅ |
| Config check | Validate configuration files | ✅ |
| Dependency check | Verify all dependencies installed | ✅ |
| MCP servers check | Verify MCP server configuration | ✅ |
| Channels check | Verify channel setup | ✅ |
| Error handling | Show proper error messages | ✅ |

---

## Success Criteria

### Completion Criteria

The comprehensive testing plan is complete when:

1. ✅ **Test Infrastructure**: All helpers and setup code implemented
2. ✅ **CLI E2E Tests**: 100% coverage of all CLI commands
   - Config: get/set/list
   - Cost: summary/daily/monthly/budget/pricing/optimize
   - MCP: list/add/remove/test/auth (HTTP & stdio)
   - Skills: list/info/check/enable/disable/install/uninstall
   - Doctor: all health checks
3. ✅ **Integration Tests**: Core service integration validated
   - LLM + Audit
   - MCP + Policy
   - Session + Memory + Cost
   - Channels + Sessions
4. ✅ **Test Documentation**: All test files have clear documentation
5. ✅ **CI/CD Ready**: Tests can run in CI environment
6. ✅ **Performance**: All tests run in acceptable time (<5s total for E2E)

### Quality Gates

| Metric | Target | Rationale |
|---------|--------|-----------|
| E2E Test Execution Time | < 5s | Fast feedback in development |
| Integration Test Time | < 10s | Comprehensive validation |
| Test Failure Rate | < 5% | High reliability |
| Code Coverage (CLI Commands) | 90%+ | All critical paths tested |
| Cleanup Success | 100% | No zombie processes or temp files |

---

## Testing Tools Reference

### Recommended Tools
- **Go**: `go test` with `-v` flag for verbose output
- **Go**: `go test -race` for race condition detection
- **Go**: `go test -cover` for coverage reports
- **Go**: `go test -timeout 30s` for runaway test protection
- **Make**: `make test` (uses Go test with proper flags)

### Test Organization Best Practices
1. **Atomic Tests**: Each test should be independent and reproducible
2. **Clear Names**: Test names should clearly describe what is being tested
3. **Cleanup**: Always cleanup resources (processes, temp files) in `defer` blocks
4. **Table-Driven Tests**: Use subtests for multiple scenarios instead of separate test functions
5. **Assertion Quality**: Use specific error messages that help diagnose failures

### Anti-Patterns to Avoid

❌ **NO** global state in tests
❌ **NO** inter-test dependencies (Test A should not depend on Test B's output)
❌ **NO** sleeping in tests (use timeouts instead)
❌ **NO** hardcoded paths (use `os.TempDir()`)
❌ **NO** `t.Fatal` without cleanup (cleanup in `defer`)

---

## Next Steps

1. Implement test helpers (`e2e/helpers.go`)
2. Implement E2E setup (`e2e/setup.go`)
3. Complete remaining CLI E2E tests:
   - `config_cli_test.go` - expand with JSON output tests
   - `mcp_cli_test.go` - add auth and stdio server tests
   - `skills_cli_test.go` - add install/uninstall tests
4. Implement integration tests:
   - Start with `integration/llm_audit_test.go`
   - Add `integration/mcp_policy_test.go`
5. Add Makefile target for running all E2E tests
6. Add CI workflow configuration

---

*This document will be updated as tests are implemented.*
