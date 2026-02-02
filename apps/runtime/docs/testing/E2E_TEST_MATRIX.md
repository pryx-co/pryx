# Pryx Comprehensive E2E Test Matrix

> **Version**: 1.0
> **Status**: Active Test Plan
> **Last Updated**: 2026-01-29
> **Document Purpose**: Complete test coverage matrix for all Pryx interfaces and features

---

## Executive Summary

This document provides a **comprehensive E2E and Integration test matrix** for all Pryx interfaces:
- **CLI** - Command-line interface
- **TUI** - Terminal UI (TypeScript + Solid + OpenTUI)
- **Web** - Local dashboard + Auth flows
- **Channels** - Telegram, Discord, Slack, Webhooks

**Test Coverage Goals**:
- Critical user workflows: 100% coverage
- Cross-interface flows: 100% coverage
- Error handling: 90% coverage
- Edge cases: 70% coverage

---

## Current State Assessment

### ✅ Existing Tests (35 test files)

| Category | Test Files | Coverage |
|----------|------------|----------|
| **Unit Tests** | 33 | Internal services (memory, cost, audit, MCP, skills, etc.) |
| **E2E Tests** | 1 | CLI basic commands (skills, MCP) |
| **Integration Tests** | 1 | Memory management |

### 📋 Test Gaps Identified

| Gap | Impact | Priority |
|-----|--------|----------|
| Missing CLI E2E for Cost, Audit, Config, Doctor | High | P0 |
| Missing Web E2E tests (OAuth, Dashboard, Settings) | High | P0 |
| Missing Cross-interface tests (CLI ↔ Web, Auth flows) | High | P0 |
| Missing Channel E2E tests (Telegram, Discord, Slack) | High | P1 |
| Missing TUI tests | Medium | P1 |
| Missing Integration tests for core services | Medium | P2 |

---

## Test Matrix by Interface

### 1. CLI E2E Tests

#### 1.1 Skills CLI (Partially Complete)

| Test Case | Status | File | Priority |
|-----------|--------|-------|----------|
| `skills list` | ✅ Complete | `e2e/cli_test.go` | P0 |
| `skills list --eligible` | ❌ Missing | `e2e/skills_cli_test.go` | P0 |
| `skills list --json` | ✅ Complete | `e2e/cli_test.go` | P0 |
| `skills info <name>` | ❌ Missing | `e2e/skills_cli_test.go` | P0 |
| `skills check` | ✅ Complete | `e2e/cli_test.go` | P0 |
| `skills enable <name>` | ❌ Missing | `e2e/skills_cli_test.go` | P0 |
| `skills disable <name>` | ❌ Missing | `e2e/skills_cli_test.go` | P0 |
| `skills install <name>` | ❌ Missing | `e2e/skills_cli_test.go` | P0 |
| `skills uninstall <name>` | ❌ Missing | `e2e/skills_cli_test.go` | P0 |
| Error handling (non-existent skill) | ✅ Complete | `e2e/cli_test.go` | P1 |

**Gap**: Need dedicated `e2e/skills_cli_test.go` for remaining commands

#### 1.2 MCP CLI (Partially Complete)

| Test Case | Status | File | Priority |
|-----------|--------|-------|----------|
| `mcp list` | ✅ Complete | `e2e/cli_test.go` | P0 |
| `mcp list --json` | ✅ Complete | `e2e/cli_test.go` | P0 |
| `mcp add <name> --url <url>` | ✅ Complete | `e2e/cli_test.go` | P0 |
| `mcp add <name> --cmd <command>` | ❌ Missing | `e2e/mcp_cli_test.go` | P0 |
| `mcp add <name> --auth bearer --token-ref <ref>` | ❌ Missing | `e2e/mcp_cli_test.go` | P0 |
| `mcp remove <name>` | ✅ Complete | `e2e/cli_test.go` | P0 |
| `mcp test <name>` | ❌ Missing | `e2e/mcp_cli_test.go` | P0 |
| `mcp auth <name>` | ❌ Missing | `e2e/mcp_cli_test.go` | P0 |
| Error handling (unknown server) | ✅ Complete | `e2e/cli_test.go` | P1 |

**Gap**: Need dedicated `e2e/mcp_cli_test.go` for auth and test commands

#### 1.3 Cost CLI (MISSING - Critical)

| Test Case | Status | File | Priority |
|-----------|--------|-------|----------|
| `cost summary` | ❌ Missing | `e2e/cost_cli_test.go` | **P0** |
| `cost daily` | ❌ Missing | `e2e/cost_cli_test.go` | **P0** |
| `cost monthly` | ❌ Missing | `e2e/cost_cli_test.go` | **P0** |
| `cost budget set --daily 10.00 --monthly 100.00` | ❌ Missing | `e2e/cost_cli_test.go` | P0 |
| `cost budget status` | ❌ Missing | `e2e/cost_cli_test.go` | P0 |
| `cost pricing` | ❌ Missing | `e2e/cost_cli_test.go` | P1 |
| `cost optimize` | ❌ Missing | `e2e/cost_cli_test.go` | P1 |

**Gap**: Need complete `e2e/cost_cli_test.go`

#### 1.4 Audit CLI (MISSING - Critical)

| Test Case | Status | File | Priority |
|-----------|--------|-------|----------|
| `audit list` | ❌ Missing | `e2e/audit_cli_test.go` | **P0** |
| `audit export --json` | ❌ Missing | `e2e/audit_cli_test.go` | **P0** |
| `audit export --csv` | ❌ Missing | `e2e/audit_cli_test.go` | P0 |
| `audit query --session <id>` | ❌ Missing | `e2e/audit_cli_test.go` | P0 |
| `audit query --tool <name>` | ❌ Missing | `e2e/audit_cli_test.go` | P0 |
| `audit cost` | ❌ Missing | `e2e/audit_cli_test.go` | P1 |

**Gap**: Need complete `e2e/audit_cli_test.go`

#### 1.5 Config CLI (MISSING)

| Test Case | Status | File | Priority |
|-----------|--------|-------|----------|
| `config get <key>` | ❌ Missing | `e2e/config_cli_test.go` | P0 |
| `config set <key> <value>` | ❌ Missing | `e2e/config_cli_test.go` | P0 |
| `config list` | ❌ Missing | `e2e/config_cli_test.go` | P0 |
| Config persistence | ❌ Missing | `e2e/config_cli_test.go` | P1 |
| Config validation | ❌ Missing | `e2e/config_cli_test.go` | P1 |

**Gap**: Need complete `e2e/config_cli_test.go`

#### 1.6 Doctor CLI (MISSING)

| Test Case | Status | File | Priority |
|-----------|--------|-------|----------|
| `doctor` - All checks | ❌ Missing | `e2e/doctor_cli_test.go` | **P0** |
| `doctor` - Database check | ❌ Missing | `e2e/doctor_cli_test.go` | P1 |
| `doctor` - Dependencies check | ❌ Missing | `e2e/doctor_cli_test.go` | P1 |
| `doctor` - Network check | ❌ Missing | `e2e/doctor_cli_test.go` | P1 |

**Gap**: Need complete `e2e/doctor_cli_test.go`

---

### 2. Web E2E Tests (MISSING - Critical)

#### 2.1 Web-only Features

| Feature | Test Cases | Status | File | Priority |
|---------|-----------|--------|-------|----------|
| **OAuth Device Flow** | Complete auth flow: CLI → Web → Authorization | ❌ Missing | `apps/web/e2e/oauth_test.ts` | **P0** |
| **Dashboard** | Load dashboard, verify metrics | ❌ Missing | `apps/web/e2e/dashboard_test.ts` | **P0** |
| **Settings** | Configure via web, verify persistence | ❌ Missing | `apps/web/e2e/settings_test.ts` | **P0** |
| **Skills Management** | Install/enable/disable via web | ❌ Missing | `apps/web/e2e/skills_web_test.ts` | P1 |
| **Cost Analytics** | View cost charts and breakdowns | ❌ Missing | `apps/web/e2e/cost_test.ts` | P1 |
| **Audit Log Viewer** | Browse audit logs, export | ❌ Missing | `apps/web/e2e/audit_test.ts` | P1 |

**Gap**: Need complete Web E2E test suite using Playwright

#### 2.2 Web Infrastructure

| Component | Status | Notes |
|-----------|--------|-------|
| Playwright installed | ✅ Ready | In `package.json` |
| Test framework setup | ⚠️ Partial | Has `Dashboard.test.tsx` (unit test) |
| E2E test runner | ❌ Missing | Need Playwright config for E2E |

---

### 3. Cross-Interface Workflow Tests (MISSING - Critical)

#### 3.1 Authentication Flows

| Workflow | Description | Status | File | Priority |
|----------|-------------|--------|-------|----------|
| **CLI → Web Auth** | User runs CLI command → Web browser opens → User approves → CLI receives token | ❌ Missing | `e2e/cross_auth_test.go` | **P0** |
| **Web → CLI Settings** | Configure via web → Verify CLI uses new settings | ❌ Missing | `e2e/cross_settings_test.go` | **P0** |
| **Session Continuity** | Start session in CLI → Continue in Web → Messages synced | ❌ Missing | `e2e/cross_session_test.go` | **P0** |

#### 3.2 Data Synchronization

| Workflow | Description | Status | File | Priority |
|----------|-------------|--------|-------|----------|
| **Cost Data Sync** | CLI cost usage → Web analytics view | ❌ Missing | `e2e/cross_cost_test.go` | P1 |
| **Skills Sync** | Install skill via Web → CLI recognizes it | ❌ Missing | `e2e/cross_skills_test.go` | P1 |
| **MCP Config Sync** | Configure MCP via Web → CLI uses config | ❌ Missing | `e2e/cross_mcp_test.go` | P1 |

#### 3.3 Multi-Device Flows

| Workflow | Description | Status | File | Priority |
|----------|-------------|--------|-------|----------|
| **Device Handoff** | Start session on laptop → Continue on server | ❌ Missing | `e2e/cross_device_test.go` | P1 |
| **Pryx Mesh** | Device discovery and pairing | ❌ Missing | `e2e/mesh_test.go` | P1 |

**Gap**: Need complete cross-interface test suite

---

### 4. Channel E2E Tests (MISSING - High Priority)

#### 4.1 Telegram

| Test Case | Description | Status | File | Priority |
|-----------|-------------|--------|-------|----------|
| Message reception | Receive message from Telegram | ❌ Missing | `e2e/channel_telegram_test.go` | **P0** |
| Message delivery | Send response to Telegram | ❌ Missing | `e2e/channel_telegram_test.go` | **P0** |
| Session management | Telegram message → Session created | ❌ Missing | `e2e/channel_telegram_test.go` | P1 |
| File handling | Handle file uploads via Telegram | ❌ Missing | `e2e/channel_telegram_test.go` | P2 |

#### 4.2 Discord

| Test Case | Description | Status | File | Priority |
|-----------|-------------|--------|-------|----------|
| Bot registration | Register Discord bot | ❌ Missing | `e2e/channel_discord_test.go` | P1 |
| Message flow | Receive/send messages | ❌ Missing | `e2e/channel_discord_test.go` | P1 |
| Slash commands | Handle Discord slash commands | ❌ Missing | `e2e/channel_discord_test.go` | P2 |

#### 4.3 Slack

| Test Case | Description | Status | File | Priority |
|-----------|-------------|--------|-------|----------|
| Bot registration | Register Slack app | ❌ Missing | `e2e/channel_slack_test.go` | P1 |
| Message flow | Receive/send messages | ❌ Missing | `e2e/channel_slack_test.go` | P1 |
| Slash commands | Handle Slack slash commands | ❌ Missing | `e2e/channel_slack_test.go` | P2 |

#### 4.4 Webhooks

| Test Case | Description | Status | File | Priority |
|-----------|-------------|--------|-------|----------|
| Webhook registration | Register webhook endpoint | ❌ Missing | `e2e/channel_webhook_test.go` | P1 |
| Payload handling | Process webhook payload | ❌ Missing | `e2e/channel_webhook_test.go` | P1 |
| Webhook authentication | Verify webhook signature | ❌ Missing | `e2e/channel_webhook_test.go` | P2 |

#### 4.5 Channel Integration Tests

| Test Case | Description | Status | File | Priority |
|-----------|-------------|--------|-------|----------|
| Multi-channel session | Same session across Telegram + CLI | ❌ Missing | `integration/channel_session_test.go` | P1 |
| Channel failover | Channel unavailable → Fallback | ❌ Missing | `integration/channel_failover_test.go` | P2 |

**Gap**: Need complete channel E2E test suite

---

### 5. Integration Tests (Partial Coverage)

#### 5.1 Existing Integration Tests

| Test | Status | Coverage | Priority |
|------|--------|----------|----------|
| Memory Management | ✅ Complete | Memory manager + store + bus | P0 |

#### 5.2 Missing Integration Tests

| Integration | Description | Status | File | Priority |
|-------------|-------------|--------|-------|----------|
| **LLM + Audit** | LLM request → Audit entry created with cost | ❌ Missing | `integration/llm_audit_test.go` | **P0** |
| **LLM + Cost** | LLM request → Cost tracked | ❌ Missing | `integration/llm_cost_test.go` | **P0** |
| **MCP + Policy** | Tool request → Policy enforced → Approval flow | ❌ Missing | `integration/mcp_policy_test.go` | **P0** |
| **Skills + Dependencies** | Skill install → Dependencies checked | ❌ Missing | `integration/skills_deps_test.go` | P1 |
| **Session + Memory + Cost** | Session lifecycle → Memory + Cost tracking | ❌ Missing | `integration/session_memory_cost_test.go` | **P0** |
| **Channels + Sessions** | Channel message → Session stored | ❌ Missing | `integration/channel_session_test.go` | P1 |

**Gap**: Need core service integration tests

---

## Test Infrastructure Requirements

### 6.1 Test Helpers & Fixtures

| Component | Status | Priority |
|-----------|--------|----------|
| `testdata/` directory structure | ⚠️ Partial | P1 |
| Test fixtures for sessions | ❌ Missing | P1 |
| Mock MCP servers | ❌ Missing | P1 |
| Test skill bundles | ❌ Missing | P1 |
| Test data for LLM calls | ❌ Missing | P2 |

### 6.2 Test Execution Matrix

| Test Type | Execution Time | Frequency | Environment |
|-----------|----------------|-----------|-------------|
| Unit Tests | < 1s | Every commit | Local |
| CLI E2E | 1-5s | Every PR | CI |
| Integration | 5-30s | Every PR | CI |
| Web E2E | 30-60s | Every PR | CI (with Playwright) |
| Channel E2E | 60-120s | Nightly | CI (requires secrets) |
| Full E2E | 2-5min | Nightly | CI/CD |
| Performance | Variable | Weekly | Performance CI |

### 6.3 Test Data Management

```
testdata/
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

---

## Implementation Priority

### Phase 1: Critical Path (P0) - 2-3 days

1. ✅ **Review existing tests** (DONE)
2. ⏳ **Complete CLI E2E tests**:
   - Skills CLI remaining commands
   - MCP CLI remaining commands
   - Cost CLI (all commands)
   - Audit CLI (all commands)
   - Doctor CLI
   - Config CLI
3. ⏳ **Web E2E - OAuth flow**:
   - CLI → Web → Authorization complete flow
   - Token exchange verification
4. ⏳ **Core Integration tests**:
   - LLM + Audit + Cost
   - MCP + Policy
   - Session + Memory + Cost

### Phase 2: Cross-Interface (P0-P1) - 2-3 days

5. ⏳ **Cross-interface workflows**:
   - Auth flow (CLI ↔ Web)
   - Settings sync (Web → CLI)
   - Session continuity (CLI ↔ Web)
   - Cost data sync
   - Skills sync
6. ⏳ **Web E2E - Dashboard & Settings**:
   - Dashboard load and metrics
   - Settings page
   - Cost analytics view
   - Audit log viewer

### Phase 3: Channel Tests (P1-P2) - 2-3 days

7. ⏳ **Channel E2E tests**:
   - Telegram (message flow)
   - Discord (message flow)
   - Slack (message flow)
   - Webhooks
8. ⏳ **Channel integration**:
   - Multi-channel sessions
   - Channel failover

### Phase 4: TUI Tests (P1) - 1-2 days

9. ⏳ **TUI E2E tests**:
   - TUI launch and navigation
   - TUI skills management
   - TUI session management

### Phase 5: Polish & Coverage (P2) - 1-2 days

10. ⏳ **Test infrastructure**:
    - Complete testdata structure
    - Test helpers and fixtures
    - Mock servers
11. ⏳ **Edge cases and error handling**:
    - Network failures
    - Invalid inputs
    - Timeout scenarios

---

## Success Metrics

| Metric | Target | Status |
|--------|--------|--------|
| **CLI E2E Coverage** | 100% of commands | 20% |
| **Web E2E Coverage** | 100% of web-only features | 0% |
| **Cross-interface Coverage** | 100% of critical workflows | 0% |
| **Channel E2E Coverage** | 80% of basic flows | 0% |
| **Integration Test Coverage** | 100% of core integrations | 15% |
| **Total Test Execution Time** | < 5min for full suite | TBD |

---

## Next Steps

1. ✅ **Create this test matrix document** (DONE)
2. ⏳ **Implement Phase 1: Critical Path**
   - Start with missing CLI E2E tests (Cost, Audit, Config, Doctor)
   - Then Web OAuth flow
   - Then core integration tests
3. ⏳ **Set up Playwright for Web E2E**
   - Configure Playwright in `apps/web/`
   - Create test fixtures and helpers
4. ⏳ **Implement test infrastructure**
   - Create `testdata/` directory
   - Add test helpers
   - Add mock servers

---

## Appendix: Test File Mapping

### E2E Test Files to Create

| File | Tests | Priority |
|------|-------|----------|
| `e2e/skills_cli_test.go` | Skills CLI commands | P0 |
| `e2e/mcp_cli_test.go` | MCP CLI commands | P0 |
| `e2e/cost_cli_test.go` | Cost CLI commands | P0 |
| `e2e/audit_cli_test.go` | Audit CLI commands | P0 |
| `e2e/config_cli_test.go` | Config CLI commands | P0 |
| `e2e/doctor_cli_test.go` | Doctor CLI commands | P0 |
| `apps/web/e2e/oauth_test.ts` | OAuth device flow | P0 |
| `apps/web/e2e/dashboard_test.ts` | Dashboard | P0 |
| `apps/web/e2e/settings_test.ts` | Settings | P0 |
| `e2e/cross_auth_test.go` | Cross-interface auth | P0 |
| `e2e/cross_session_test.go` | Session continuity | P0 |
| `e2e/channel_telegram_test.go` | Telegram channel | P1 |
| `e2e/channel_discord_test.go` | Discord channel | P1 |
| `e2e/channel_slack_test.go` | Slack channel | P1 |
| `e2e/channel_webhook_test.go` | Webhook channel | P1 |

### Integration Test Files to Create

| File | Tests | Priority |
|------|-------|----------|
| `integration/llm_audit_test.go` | LLM + Audit | P0 |
| `integration/llm_cost_test.go` | LLM + Cost | P0 |
| `integration/mcp_policy_test.go` | MCP + Policy | P0 |
| `integration/session_memory_cost_test.go` | Session + Memory + Cost | P0 |
| `integration/channel_session_test.go` | Channel + Sessions | P1 |

---

*This test matrix will be updated as tests are implemented and coverage is achieved.*
