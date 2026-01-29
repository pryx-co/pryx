# Comprehensive Testing Implementation Plan

> **Status**: In Progress  
> **Created**: 2026-01-29  
> **Last Updated**: 2026-01-29  

---

## Background and Motivation

The Pryx project requires comprehensive testing coverage across all interfaces and components. Based on the audit of existing tests and analysis of reference implementations (moltbot, opencode), we have good coverage but need to fill gaps in:

1. **Unit Tests**: Missing for Agent, Mesh, Telemetry
2. **Integration Tests**: Missing for service interactions (LLM+Audit, MCP+Policy, etc.)
3. **E2E Tests**: Missing for some CLI commands (Audit, Skills install/uninstall)
4. **Test Infrastructure**: Need comprehensive helpers and fixtures

---

## Current State Assessment

### ✅ Existing Test Coverage (All Passing)

| Category | Count | Coverage |
|----------|-------|----------|
| **E2E Tests** | 28 tests | CLI commands (Skills, MCP, Config, Cost, Doctor) |
| **Unit Tests** | 29 packages | Internal services (audit, auth, bus, channels, config, cost, etc.) |
| **Integration Tests** | 3 files | Memory, Runtime, Workflow |

### 📋 Test Gaps Identified

| Component | Test Type | Priority | Status |
|-----------|-----------|----------|--------|
| Agent | Unit | P1 | Missing |
| Mesh | Unit | P2 | Missing |
| Telemetry | Unit | P2 | Missing |
| LLM + Audit | Integration | P0 | Missing |
| MCP + Policy | Integration | P0 | Missing |
| Session + Memory + Cost | Integration | P0 | Missing |
| Audit CLI | E2E | P0 | Missing |
| Skills Install/Uninstall | E2E | P1 | Missing |
| MCP Auth | E2E | P1 | Missing |

---

## High-Level Task Breakdown

### Phase 1: Missing Unit Tests
1. Create `internal/agent/agent_test.go`
2. Create `internal/mesh/mesh_test.go`
3. Create `internal/telemetry/telemetry_test.go`

### Phase 2: Core Integration Tests
1. Create `integration/llm_audit_test.go`
2. Create `integration/mcp_policy_test.go`
3. Create `integration/session_memory_cost_test.go`

### Phase 3: Remaining E2E Tests
1. Create `e2e/audit_cli_test.go`
2. Extend `e2e/skills_cli_test.go` with install/uninstall tests
3. Extend `e2e/mcp_cli_test.go` with auth tests

### Phase 4: Test Infrastructure
1. Create `testhelpers/` package with common utilities
2. Create `testdata/` structure with fixtures
3. Create mock servers for external dependencies

---

## Phase 1: Unit Tests Implementation

### 1.1 Agent Package Tests

**File**: `apps/runtime/internal/agent/agent_test.go`

```go
// Test cases needed:
- TestNewAgent - Agent initialization
- TestAgent_Spawn - Spawning agent processes
- TestAgent_ExecuteTool - Tool execution
- TestAgent_HandleMessage - Message handling
- TestAgent_ContextCancellation - Proper cleanup
```

### 1.2 Mesh Package Tests

**File**: `apps/runtime/internal/mesh/mesh_test.go`

```go
// Test cases needed:
- TestNewMesh - Mesh initialization
- TestMesh_DeviceDiscovery - Device discovery
- TestMesh_PeerConnection - Peer connection management
- TestMesh_MessageRouting - Message routing between peers
- TestMesh_Sync - State synchronization
```

### 1.3 Telemetry Package Tests

**File**: `apps/runtime/internal/telemetry/telemetry_test.go`

```go
// Test cases needed:
- TestNewTelemetry - Telemetry initialization
- TestTelemetry_RecordSpan - Span recording
- TestTelemetry_Export - Data export
- TestTelemetry_ContextPropagation - Context handling
```

---

## Phase 2: Integration Tests Implementation

### 2.1 LLM + Audit Integration

**File**: `apps/runtime/integration/llm_audit_test.go`

**Purpose**: Verify that LLM requests create proper audit entries with cost data.

```go
// Test cases:
- TestLLMRequest_CreatesAuditEntry
- TestLLMRequest_AuditContainsCost
- TestLLMRequest_MultipleProvidersCreateSeparateAudits
- TestLLMRequest_AuditOnFailure
```

### 2.2 MCP + Policy Integration

**File**: `apps/runtime/integration/mcp_policy_test.go`

**Purpose**: Verify policy enforcement on MCP tool requests.

```go
// Test cases:
- TestMCPTool_PolicyAllow - Allowed tool execution
- TestMCPTool_PolicyDeny - Denied tool execution
- TestMCPTool_PolicyApprovalFlow - Approval workflow
- TestMCPTool_BundledServersPolicy - Bundled server policy
```

### 2.3 Session + Memory + Cost Integration

**File**: `apps/runtime/integration/session_memory_cost_test.go`

**Purpose**: Verify session lifecycle with memory and cost tracking.

```go
// Test cases:
- TestSession_LifecycleWithMemoryAndCost
- TestSession_MemorySummarization
- TestSession_CostTrackingAcrossMessages
- TestSession_ArchivePreservesCostData
```

---

## Phase 3: E2E Tests Implementation

### 3.1 Audit CLI E2E Tests

**File**: `apps/runtime/e2e/audit_cli_test.go`

```go
// Test cases:
- TestAuditCLI_List - List audit entries
- TestAuditCLI_ExportJSON - Export as JSON
- TestAuditCLI_ExportCSV - Export as CSV
- TestAuditCLI_QuerySession - Query by session
- TestAuditCLI_QueryTool - Query by tool
- TestAuditCLI_Cost - Show cost from audit
```

### 3.2 Extended Skills CLI E2E Tests

**Extend**: `apps/runtime/e2e/skills_cli_test.go`

```go
// Additional test cases:
- TestSkillsCLI_EnableDisable - Enable/disable skills
- TestSkillsCLI_Install - Install skills
- TestSkillsCLI_Uninstall - Uninstall skills
- TestSkillsCLI_Info - Get skill info
```

### 3.3 Extended MCP CLI E2E Tests

**Extend**: `apps/runtime/e2e/mcp_cli_test.go`

```go
// Additional test cases:
- TestMCPCLI_Auth - Auth command
- TestMCPCLI_AddStdio - Add stdio server
- TestMCPCLI_AddWithAuth - Add with authentication
- TestMCPCLI_Test - Test server connection
```

---

## Phase 4: Test Infrastructure

### 4.1 Test Helpers Package

**File**: `apps/runtime/testhelpers/helpers.go`

```go
// Helpers to implement:
- SetupTestDB() - Create test database
- CleanupTestDB() - Cleanup database
- CreateTestConfig() - Create test configuration
- MockMCPServer() - Mock MCP server
- MockLLMProvider() - Mock LLM provider
- WaitForCondition() - Async condition waiter
- GenerateTestSession() - Generate test session data
```

### 4.2 Test Data Structure

```
apps/runtime/testdata/
├── sessions/
│   └── test-sessions.json
├── skills/
│   ├── bundled/
│   └── eligible/
├── mcp/
│   └── servers.json
├── audit/
│   └── test-entries.json
└── channels/
    ├── telegram/
    ├── discord/
    └── webhook/
```

### 4.3 Mock Servers

**File**: `apps/runtime/testhelpers/mocks/mcp_server.go`

```go
// Mock implementations:
- MockMCPServer - HTTP MCP server mock
- MockLLMProvider - LLM provider mock
- MockTelegramAPI - Telegram API mock
- MockDiscordAPI - Discord API mock
```

---

## Testing Patterns (Based on Moltbot)

### Unit Test Pattern

```go
func TestComponent_Method(t *testing.T) {
    // Arrange
    deps := setupDependencies(t)
    
    // Act
    result, err := component.Method(deps)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Integration Test Pattern

```go
func TestIntegration_ServiceInteraction(t *testing.T) {
    // Setup real database
    db := testhelpers.SetupTestDB(t)
    defer testhelpers.CleanupTestDB(t, db)
    
    // Initialize services with real dependencies
    serviceA := NewServiceA(db)
    serviceB := NewServiceB(db)
    
    // Execute interaction
    result := serviceA.DoSomething(serviceB)
    
    // Verify state in database
    // Verify side effects
}
```

### E2E Test Pattern

```go
func TestE2E_CLICommand(t *testing.T) {
    // Setup isolated environment
    home := t.TempDir()
    
    // Run CLI command
    out, code := runPryxCoreWithEnv(t, home, nil, "command", "subcommand")
    
    // Verify exit code and output
    if code != 0 {
        t.Fatalf("unexpected exit code: %d, output: %s", code, out)
    }
    
    if !strings.Contains(out, "expected") {
        t.Errorf("expected output to contain 'expected', got: %s", out)
    }
}
```

---

## Execution Plan

### Week 1: Unit Tests
- [ ] Day 1-2: Agent package tests
- [ ] Day 3: Mesh package tests
- [ ] Day 4-5: Telemetry package tests

### Week 2: Integration Tests
- [ ] Day 1-2: LLM + Audit integration
- [ ] Day 3-4: MCP + Policy integration
- [ ] Day 5: Session + Memory + Cost integration

### Week 3: E2E Tests
- [ ] Day 1-2: Audit CLI E2E tests
- [ ] Day 3: Skills CLI extended tests
- [ ] Day 4: MCP CLI extended tests
- [ ] Day 5: Test infrastructure helpers

### Week 4: Infrastructure & Documentation
- [ ] Day 1-2: Test helpers package
- [ ] Day 3: Mock servers
- [ ] Day 4: Test fixtures
- [ ] Day 5: Documentation

---

## Success Metrics

| Metric | Current | Target | Timeline |
|--------|---------|--------|----------|
| Unit Test Coverage | 85% | 95% | Week 1 |
| Integration Test Coverage | 30% | 80% | Week 2 |
| E2E CLI Coverage | 80% | 100% | Week 3 |
| Test Execution Time | ~60s | < 60s | Week 4 |
| Test Reliability | 100% | 100% | Ongoing |

---

## Progress Tracking

### Phase 1: Unit Tests
- [ ] Agent tests implemented
- [ ] Mesh tests implemented
- [ ] Telemetry tests implemented

### Phase 2: Integration Tests
- [ ] LLM + Audit integration tests
- [ ] MCP + Policy integration tests
- [ ] Session + Memory + Cost integration tests

### Phase 3: E2E Tests
- [ ] Audit CLI E2E tests
- [ ] Skills CLI extended tests
- [ ] MCP CLI extended tests

### Phase 4: Infrastructure
- [ ] Test helpers package
- [ ] Mock servers
- [ ] Test fixtures
- [ ] Documentation

---

## Lessons Learned

### [2026-01-29]
- Project already has excellent E2E test infrastructure following moltbot patterns
- All existing 28 E2E tests pass successfully
- 29 internal packages have unit tests
- Missing unit tests for Agent, Mesh, Telemetry packages
- Integration tests exist but need expansion for core service interactions
- Test execution is fast (~60s for all tests)
- Strong foundation to build upon

---

*This plan will be updated as implementation progresses.*
