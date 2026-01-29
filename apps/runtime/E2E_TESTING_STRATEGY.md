# Pryx E2E & Integration Testing Strategy

## Phase 1: Feature Inventory & Test Strategy

### 1.1 Implemented Features Catalog

| Feature | Interface | Type | Dependencies | Status |
|---------|-----------|------|--------------|---------|
| Performance Testing | CLI | Command | None | ✅ Complete |
| Audit Log | CLI + Internal | Service | Store, Bus | ✅ Complete |
| Token Cost Awareness | CLI + Internal | Service | LLM Providers, Store | ✅ Complete |
| Skills CLI | CLI | Command | Skills System | ✅ Complete |
| MCP CLI | CLI | Command | MCP System | ✅ Complete |
| Memory Management | Internal | Service | Store, Bus | ✅ Complete |
| Security Audit | Internal | Review | All systems | ✅ Complete |

### 1.2 Interface Mapping

**CLI Commands Available:**
```
pryx-core skills list [--eligible] [--json]
pryx-core skills info <name>
pryx-core skills check
pryx-core skills enable <name>
pryx-core skills disable <name>
pryx-core skills install <name>
pryx-core skills uninstall <name>

pryx-core mcp list
pryx-core mcp add <name> --url <url> [--auth <type> --token-ref <ref>]
pryx-core mcp add <name> --cmd <command>
pryx-core mcp remove <name>
pryx-core mcp test <name>
pryx-core mcp auth <name>

pryx-core doctor
pryx-core config <set|get|list>
```

**Internal APIs Available:**
- Cost tracking: `/api/cost/*`
- Audit logging: `/api/audit/*`
- Session management: `/api/sessions/*`
- Memory management: Internal service

**Bundled MCP Servers:**
- filesystem (read/write files)
- shell (execute commands)
- browser (web browsing)
- clipboard (clipboard access)

**Bundled Skills:**
- cloud-deploy (deployment automation)
- docker-manager (container management)
- git-tool (version control)

### 1.3 Cross-Interface Workflows

| Workflow | Interfaces | Description |
|----------|------------|-------------|
| Authentication | CLI → Web | OAuth device flow |
| Settings Management | Web → CLI | Configure via web, use via CLI |
| Session Continuity | CLI ↔ Web | Continue sessions across interfaces |
| Cost Analytics | CLI → Web | Usage in CLI, analytics in web |
| Skills Management | Web → CLI | Install via web, use via CLI |
| MCP Configuration | Web → CLI | Configure servers, use in sessions |

## Phase 2: CLI/TUI E2E Test Plan

### 2.1 Skills CLI E2E Tests

```go
// Test file: e2e/skills_cli_test.go
func TestSkillsCLI_List(t *testing.T) {
    // Setup: Ensure bundled skills available
    // Execute: pryx-core skills list
    // Verify: Output shows bundled skills
}

func TestSkillsCLI_ListWithEligible(t *testing.T) {
    // Execute: pryx-core skills list --eligible
    // Verify: Only eligible skills shown
}

func TestSkillsCLI_ListJSON(t *testing.T) {
    // Execute: pryx-core skills list --json
    // Verify: Valid JSON output
}

func TestSkillsCLI_Info(t *testing.T) {
    // Execute: pryx-core skills info docker-manager
    // Verify: Shows skill metadata
}

func TestSkillsCLI_Check(t *testing.T) {
    // Execute: pryx-core skills check
    // Verify: All checks pass for bundled skills
}

func TestSkillsCLI_EnableDisable(t *testing.T) {
    // Execute: pryx-core skills enable docker-manager
    // Verify: Skill enabled
    // Execute: pryx-core skills disable docker-manager
    // Verify: Skill disabled
}

func TestSkillsCLI_InstallUninstall(t *testing.T) {
    // Execute: pryx-core skills install <name>
    // Verify: Installation prepared
}
```

### 2.2 MCP CLI E2E Tests

```go
// Test file: e2e/mcp_cli_test.go
func TestMCPCLI_List(t *testing.T) {
    // Setup: Add test MCP servers
    // Execute: pryx-core mcp list
    // Verify: Shows configured servers
}

func TestMCPCLI_AddHTTP(t *testing.T) {
    // Execute: pryx-core mcp add test-server --url "https://example.com"
    // Verify: Server added, config saved
}

func TestMCPCLI_AddStdio(t *testing.T) {
    // Execute: pryx-core mcp add test-server --cmd "/path/to/server"
    // Verify: Server added with command
}

func TestMCPCLI_AddWithAuth(t *testing.T) {
    // Execute: pryx-core mcp add secure --url "https://api" --auth bearer --token-ref "key"
    // Verify: Auth config saved
}

func TestMCPCLI_Remove(t *testing.T) {
    // Execute: pryx-core mcp remove test-server
    // Verify: Server removed from config
}

func TestMCPCLI_Test(t *testing.T) {
    // Execute: pryx-core mcp test test-server
    // Verify: Configuration validated
}

func TestMCPCLI_Auth(t *testing.T) {
    // Execute: pryx-core mcp auth test-server
    // Verify: Auth configuration shown
}

func TestMCPCLI_JSONOutput(t *testing.T) {
    // Execute: pryx-core mcp list --json
    // Verify: Valid JSON with all servers
}

func TestMCPCLI_ErrorHandling(t *testing.T) {
    // Test: Unknown server, missing args, invalid flags
    // Verify: Proper error messages and exit codes
}
```

### 2.3 Cost CLI E2E Tests

```go
// Test file: e2e/cost_cli_test.go
func TestCostCLI_Summary(t *testing.T) {
    // Setup: Generate some LLM usage
    // Execute: pryx-core cost summary
    // Verify: Shows total costs
}

func TestCostCLI_Daily(t *testing.T) {
    // Execute: pryx-core cost daily
    // Verify: Shows daily breakdown
}

func TestCostCLI_Monthly(t *testing.T) {
    // Execute: pryx-core cost monthly
    // Verify: Shows monthly breakdown
}

func TestCostCLI_Budget(t *testing.T) {
    // Execute: pryx-core cost budget set --daily 10.00 --monthly 100.00
    // Verify: Budget saved
    // Execute: pryx-core cost budget status
    // Verify: Shows current usage
}

func TestCostCLI_Pricing(t *testing.T) {
    // Execute: pryx-core cost pricing
    // Verify: Shows model pricing table
}

func TestCostCLI_Optimizations(t *testing.T) {
    // Execute: pryx-core cost optimize
    // Verify: Shows cost optimization suggestions
}
```

### 2.4 Audit CLI E2E Tests

```go
// Test file: e2e/audit_cli_test.go
func TestAuditCLI_List(t *testing.T) {
    // Setup: Generate some audit entries
    // Execute: pryx-core audit list
    // Verify: Shows audit entries
}

func TestAuditCLI_Export(t *testing.T) {
    // Execute: pryx-core audit export --json
    // Verify: Valid JSON export
    // Execute: pryx-core audit export --csv
    // Verify: Valid CSV export
}

func TestAuditCLI_Query(t *testing.T) {
    // Execute: pryx-core audit query --session <id>
    // Verify: Filtered results
    // Execute: pryx-core audit query --tool <name>
    // Verify: Tool-specific entries
}

func TestAuditCLI_Cost(t *testing.T) {
    // Execute: pryx-core audit cost
    // Verify: Cost summary from audit
}
```

## Phase 3: Integration Test Plan

### 3.1 LLM + Audit Integration

```go
// Test file: integration/llm_audit_test.go
func TestLLMProvider_AuditLogging(t *testing.T) {
    // Setup: Initialize LLM provider and audit
    // Execute: Make LLM request
    // Verify: Audit entry created with cost data
}

func TestLLMProvider_TokenCounting(t *testing.T) {
    // Execute: Make LLM request
    // Verify: Tokens counted and cost calculated
    // Verify: Cost stored in audit log
}

func TestLLMProvider_MultipleProviders(t *testing.T)) {
    // Setup: Configure multiple providers
    // Execute: Requests to different providers
    // Verify: Each request logged with correct cost
}
```

### 3.2 MCP + Policy Integration

```go
// Test file: integration/mcp_policy_test.go
func TestMCPServer_PolicyEnforcement(t *testing.T) {
    // Setup: Configure policy engine with rules
    // Execute: Tool request through MCP
    // Verify: Policy evaluated, approval flow triggered
}

func TestMCPServer_BundledServers(t *testing.T) {
    // Execute: Use bundled filesystem server
    // Verify: Policy applied correctly
}
```

### 3.3 Skills + Dependencies Integration

```go
// Test file: integration/skills_dependencies_test.go
func TestSkills_InstallWithDependencies(t *testing.T) {
    // Setup: Skill with required bins/env
    // Execute: pryx-core skills install <skill>
    // Verify: Dependencies checked and installed
}

func TestSkills_LoadAndExecute(t *testing.T) {
    // Execute: Load skill
    // Verify: Skill loaded, prompts available
}
```

### 3.4 Session + Memory + Cost Integration

```go
// Test file: integration/session_memory_cost_test.go
func TestSession_LifecycleWithMemoryAndCost(t *testing.T) {
    // Setup: Create session
    // Execute: Add messages, make LLM requests
    // Verify: 
    //   - Messages stored
    //   - Token count updated
    //   - Cost tracked
    //   - Memory usage calculated
    //   - Auto-summarization triggered at threshold
}

func TestSession_ArchiveWithCostData(t *testing.T) {
    // Setup: Old session with cost data
    // Execute: Archive session
    // Verify: Session archived, cost data preserved
}
```

### 3.5 Channels + Sessions Integration

```go
// Test file: integration/channels_sessions_test.go
func TestChannels_SessionMessaging(t *testing.T) {
    // Setup: Configure Telegram channel
    // Execute: Send message through channel
    // Verify: Message stored in session
}

func TestChannels_MultipleInterfaces(t *testing.T) {
    // Setup: CLI and Telegram connected to same session
    // Execute: Message from CLI
    // Verify: Message appears in Telegram
}
```

## Phase 4: Test Execution Strategy

### 4.1 Test Runner Selection

**CLI E2E Tests:**
- Use Go's `os/exec` to run commands
- Capture output and exit codes
- Parse JSON output for verification
- Use temporary directories for config

**Integration Tests:**
- Use real SQLite database (in-memory for tests)
- Mock external services where needed
- Test actual component integration
- Use context for timeout handling

**Web E2E Tests:**
- Use Playwright for browser automation
- Test HTTP API endpoints
- Verify WebSocket communication

### 4.2 Test Data Management

```go
// testdata/ structure:
testdata/
├── skills/
│   ├── bundled/
│   │   ├── docker-manager/
│   │   └── git-tool/
│   └── eligible/
├── mcp/
│   └── servers.json
├── sessions/
│   └── test-sessions/
└── audit/
    └── test-entries/
```

### 4.3 Test Coverage Goals

| Category | Coverage Target | Priority |
|----------|-----------------|----------|
| CLI Commands | 100% of commands | P0 |
| CLI Flags | 100% of flags | P0 |
| Error Cases | 90% coverage | P1 |
| Integration Paths | 80% coverage | P1 |
| Edge Cases | 70% coverage | P2 |

### 4.4 Test Execution Matrix

| Test Type | Execution Time | Frequency | Environment |
|-----------|----------------|-----------|-------------|
| Unit Tests | < 1s | Every commit | Local |
| CLI E2E | 1-5s | Every PR | CI |
| Integration | 5-30s | Every PR | CI |
| Full E2E | 30-120s | Nightly | CI/CD |
| Performance | Variable | Weekly | Performance CI |

## Phase 5: Implementation Priority

### Priority 1: Critical Path (Must Have)
1. Skills CLI basic commands (list, info)
2. MCP CLI basic commands (list, add, remove)
3. Cost tracking integration
4. Audit logging integration
5. Session lifecycle

### Priority 2: Important (Should Have)
1. All CLI flags and options
2. Error handling verification
3. Configuration management
4. Channel integration

### Priority 3: Nice to Have
1. Web UI E2E tests
2. Performance regression tests
3. Security testing
4. Load testing

## Deliverables

1. **Test Inventory Document** (this file)
2. **Test Matrix** (feature → test type → coverage)
3. **E2E Test Suite** (implementation in `e2e/` directory)
4. **Integration Test Suite** (implementation in `integration/` directory)
5. **Test Documentation** (README with execution instructions)

## Execution

Start with Priority 1 tests:
```bash
# Run CLI E2E tests
go test ./e2e/... -v

# Run Integration tests
go test ./integration/... -v

# Run all tests
go test ./... -v
```
