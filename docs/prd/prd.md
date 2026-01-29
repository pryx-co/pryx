# Product Requirements Document (PRD): Pryx

> **Version**: 1.0  
> **Status**: Canonical  
> **Last Updated**: 2026-01-27  

---

## 0) Product Principles (Non-Negotiables)

These principles are inviolable and apply to every feature decision:

| Principle | Description |
|-----------|-------------|
| **Sovereign-by-default** | User data stays on user-owned machines unless explicitly enabled for export/backup |
| **Telemetry opt-out** | Analytics/telemetry ON by default (helps improvement); prompts/PII NEVER sent; user can disable anytime |
| **Zero-friction onboarding** | No Node.js, Python, or external runtime dependency for end users |
| **Safe execution** | Human-in-the-loop approvals + policy engine for all sensitive operations |
| **Observable** | Every action is traceable locally; optional sanitized export to user-chosen backends |
| **Simple + reusable** | Centralized logic; avoid duplicated flows and config surfaces |

---

## 1) Executive Summary

**Pryx** is a second-generation "sovereign AI agent" that combines the power of a local-first control center with the polish of a modern desktop application. It replaces fragile CLI + config-file assistants with:

- A first-class UI for onboarding, integrations, sessions, and telemetry
- Secure permission gating with human-in-the-loop approvals
- Transparent observability without sacrificing privacy
- A resilient sidecar architecture for crash isolation

The system ships as compiled binaries with one-liner install. A thin Cloudflare Workers edge layer handles OAuth device flows and telemetry ingestion without persisting user conversations.

**Key Differentiators**:
- Local machine owns: conversations, files, indices, policies, secrets
- Cloud edge owns: auth handshakes, pairing, optional telemetry relay (stateless)
- Polyglot architecture: Rust (host), Go (runtime), TypeScript/React (UI)

---

## 2) Problem Statement

Existing local AI agents commonly fail on:

| Problem | Impact | Current State |
|---------|--------|---------------|
| **Installation friction** | Users abandon setup | Dependency hell (Node.js versioning, Python conflicts) |
| **No management UI** | Poor discoverability | Setup scattered across CLI + JSON config files |
| **Opaque security posture** | Trust erosion | Keys and data flows unclear to users |
| **Lack of observability** | Debugging impossible | Hard to know what agent did, why, and at what cost |
| **Unsafe tool execution** | Security risk | Overbroad privileges, weak approval gates |
| **No multi-device story** | Fragmented experience | Server + laptop coordination is ad-hoc |

---

## 3) Competitive Analysis

### 3.1 Market Landscape

| Competitor | Type | Strengths | Weaknesses | Pryx Advantage |
|------------|------|-----------|------------|----------------|
| **Cursor** | IDE-integrated | Deep editor integration, good UX | Closed source, subscription model, IDE-locked | Open, standalone, no vendor lock-in |
| **Continue.dev** | IDE extension | Open source, multi-IDE | Still requires IDE, limited automation | Headless-capable, channel integrations |
| **OpenAI Desktop** | Desktop app | Official, polished | Cloud-only, no local sovereignty | Local-first, data sovereignty |
| **Clawdbot** | CLI agent | Multi-channel, flexible | Node.js dependency, no GUI, config complexity | Zero-dep binary, first-class UI |
| **Aider** | CLI tool | Git-aware, solid | CLI-only, no integrations | Full UI, channel mux, observability |
| **GPT Pilot** | Development agent | Full project generation | Heavy, complex setup | Lightweight, modular |

### 3.2 Competitive Positioning

```
                    Cloud-First
                         |
         OpenAI Desktop  |  Cursor
                         |
    --------------------|--------------------
    Local-First         |         IDE-Integrated
                         |
         Pryx (target)   |  Continue.dev
         Clawdbot        |  Aider
                         |
                    Standalone
```

**Pryx Target Quadrant**: Local-First + Standalone with optional cloud sync

### 3.3 Key Differentiators

1. **Zero-dependency install**: Single binary, no runtime requirements
2. **Local-first sovereignty**: All data on-device by default
3. **Multi-channel**: Telegram, Discord, Slack, webhooks from day one
4. **Observable**: Built-in telemetry with cost tracking
5. **Safe by default**: Policy engine with explicit approvals

---

## 4) Target Users & Personas

### 4.1 Primary Personas

| Persona | Profile | Needs | Pain Points |
|---------|---------|-------|-------------|
| **Solo Builder** | Independent developer, CLI-comfortable | Speed, keyboard workflow, automation with guardrails | Too much config, unclear security |
| **Operator** | DevOps/SRE running headless agents | Server deployment, remote UI access, reliability | Process management, monitoring gaps |
| **Power User** | Multi-channel integrations | Clear governance, audit trails | Scattered tools, no unified control |
| **Security-Conscious** | Enterprise/regulated environment | Explicit approvals, audit logs, keychain storage | Opaque data flows, weak permissions |

### 4.2 Primary Use Cases

| Use Case | Description | Success Criteria |
|----------|-------------|------------------|
| **Code/Workspace Agent** | File ops, shell, git, HTTP tools with approvals | Complete task with <3 approval prompts |
| **Channel Agent** | Telegram/Discord/Slack/webhooks as inbound triggers | <500ms response latency p95 |
| **Multi-Device** | Laptop UI controls server agent | Seamless handoff, <2s sync delay |
| **Observability** | Timeline, costs, errors, tool calls, approvals | Answer "what happened?" in <30s |

---

## 5) Goals & Non-Goals

### 5.1 Goals (MVP)

| ID | Goal | Success Metric |
|----|------|----------------|
| G1 | Zero-friction install | <2 minutes from download to first chat |
| G2 | First-class UI | Onboarding wizard completion rate >90% |
| G3 | Crash-resilient architecture | 0 zombie processes after app exit |
| G4 | Transparent observability | >90% sessions have complete trace timeline |
| G5 | Safe execution | 100% sensitive tool calls require approval |
| G6 | Extensible tools via MCP | Add new tool without rebuilding host |

### 5.2 Non-Goals (Initial Release)

| Non-Goal | Rationale | Future Consideration |
|----------|-----------|---------------------|
| Mobile native apps | Focus on desktop/server first | v2+ |
| Skills marketplace monetization | Validate core product first | v2+ |
| Mandatory cloud sync | Conflicts with sovereignty principle | Never mandatory |
| Local LLM inference | API-first for MVP simplicity | v1.5+ |
| Mobile native apps | Focus on desktop/server first | v2+ |
| Skills marketplace monetization | Validate core product first | v2+ |
| Mandatory cloud sync | Conflicts with sovereignty principle | Never mandatory |
| Local LLM inference | API-first for MVP simplicity | v1.5+ |
| Voice wake word | Complexity vs value | v2+ |

### 5.3 Technical Constraints
- **Development Runtime**: Bun v1.3.7+ required for all TypeScript/JS tooling.
- **Node.js**: Explicitly avoided for development and runtime.

---

## 6) Scope: MVP vs v1

### 6.1 MVP (Ship fast, validate product)

| Component | Deliverables |
|-----------|--------------|
| **Installation** | One-liner installer (macOS/Linux + Windows), no external deps |
| **CLI/TUI** | `pryx` TUI, `pryx run "..."` for scripts, `pryx serve` for headless |
| **Web UI** | Localhost dashboard for onboarding/config/telemetry |
| **Runtime** | Go sidecar with HTTP+WebSocket API, MCP tools |
| **Auth** | Cloudflare worker with OAuth Device Flow (RFC 8628) |
| **Telemetry** | OTLP ingest + PII redaction worker |
| **Channels** | Telegram + generic webhooks |
| **Health** | `pryx doctor` diagnostics, `/health` endpoint |
| **Skills** | Managed skills (`~/.pryx/skills`) + workspace skills |
| **MCP** | Native MCP client (stdio + HTTP transport) |

### 6.2 v1 (Make it durable)

| Component | Additions |
|-----------|-----------|
| **Desktop App** | Tauri host for native dialogs, tray, deep links |
| **Multi-device** | "Pryx Mesh" with device pairing and registry |
| **AI Gateway** | Cloudflare-based model routing, quotas, key management |
| **Channels** | Discord, Slack; WhatsApp strategy (Cloud API vs subprocess) |
| **Persistence** | Optional cloud config backup, device metadata sync |

---

## 7) Product Surfaces

### 7.1 CLI & TUI (Required for MVP)

```
pryx              # Launch TUI (default)
pryx run "..."    # Non-interactive, single prompt
pryx serve        # Headless daemon mode
pryx doctor       # Health diagnostics
pryx install-service   # Install as system service
pryx uninstall-service # Remove system service

# Skills management
pryx skills list [--eligible] [--json]  # List all skills
pryx skills info <name>                  # Show skill details
pryx skills check                        # Validate skill configs
pryx skills enable <name>                # Enable a skill
pryx skills disable <name>               # Disable a skill
pryx skills install <name>               # Install skill dependencies

# MCP management
pryx mcp list                            # List configured MCP servers
pryx mcp add <name> --url <url>          # Add HTTP MCP server
pryx mcp add <name> --cmd <command>      # Add stdio MCP server
pryx mcp remove <name>                   # Remove MCP server
pryx mcp test <name>                     # Test MCP server connection
pryx mcp auth <name>                     # Authenticate with MCP server (OAuth)
```

**TUI Screens**:
| Screen | Purpose |
|--------|---------|
| Sessions | List, resume, export past conversations |
| Chat | Streaming responses, tool approval prompts |
| Integrations | Connect/disconnect channels, status indicators |
| Skills | Browse, enable/disable, install dependencies |
| MCP Servers | Status, available tools, connection health |
| Telemetry | Timeline + costs (summary; deep dive in web UI) |

### 7.2 Localhost Web UI (Required for MVP)

| View | Features |
|------|----------|
| **Onboarding Wizard** | Connect integrations via QR/device codes, 3-step flow |
| **Channel Manager** | Add/remove channels, routing rules, status badges |
| **Skills Manager** | Browse skills, enable/disable, install dependencies, view requirements |
| **MCP Servers** | Add/remove servers, test connections, browse available tools |
| **Observability Dashboard** | Trace timeline (Gantt), approvals, errors, costs |
| **Admin Settings** | Model routing, telemetry toggle (opt-in), policy management |

### 7.3 Desktop Wrapper (Optional in MVP, Required for v1)

| Feature | Implementation |
|---------|----------------|
| Native permission dialogs | Tauri system dialog API |
| System tray | Persistent background presence |
| Deep linking | `pryx://` protocol for sessions/settings/auth |
| Auto-updates | Background check + rollback capability |

### 7.4 Cloud Web Dashboard (Account + Devices)

Pryx’s “web platform” dashboard is intentionally privacy-preserving by default:

- It manages account login and device linking (OAuth Device Flow / pairing flows).
- It shows device list + health + sanitized telemetry and error summaries.
- It does not expose raw secrets, prompts, or full local logs unless the user explicitly exports.

Detailed contract: [auth-sync-dashboard.md](../auth-sync-dashboard.md).

---

## 8) Architecture

### 8.1 Component Diagram

**Core Pattern**: Pryx-core acts as the **Gateway** (single source of truth). All surfaces and channels connect to it.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CONTROL SURFACES                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │   TUI    │ │  Web UI  │ │ Desktop  │ │ Telegram │ │ WhatsApp │       │
│  │          │ │          │ │ (Tauri)  │ │   Bot    │ │   Bot    │       │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘       │
│       │            │            │            │            │              │
│       └────────────┴────────────┴────────────┴────────────┘              │
│                              │                                            │
│                              │ WebSocket (real-time sync)                 │
│                              ▼                                            │
│  ┌───────────────────────────────────────────────────────────────────┐   │
│  │                pryx-core (GATEWAY - Source of Truth)              │   │
│  │                                                                   │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │   │
│  │  │Session Bus │  │  Policy    │  │   MCP      │  │ Telemetry  │  │   │
│  │  │(pub/sub)   │  │  Engine    │  │  Client    │  │  Export    │  │   │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │   │
│  │                                                                   │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │   │
│  │  │   LLM      │  │  Channel   │  │  Session   │  │   Audit    │  │   │
│  │  │ Orchestrate│  │  Adapters  │  │   Store    │  │    Log     │  │   │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │   │
│  │                                                                   │   │
│  └───────────────────────────────────────────────────────────────────┘   │
│                              │                                            │
│         ┌────────────────────┼────────────────────────┐                  │
│         │                    │                        │                  │
│         ▼                    ▼                        ▼                  │
│  ┌────────────┐      ┌────────────┐          ┌────────────────┐         │
│  │ MCP Servers│      │   SQLite   │          │ Rust Host      │         │
│  │            │      │ + Keychain │          │ (Lifecycle)    │         │
│  │ - filesystem      │            │          │ - Native dialog│         │
│  │ - shell   │      │ Sessions   │          │ - Sidecar mgmt │         │
│  │ - browser │      │ Audit log  │          │ - Auto-update  │         │
│  │ - http    │      │ Policies   │          └────────────────┘         │
│  └────────────┘      └────────────┘                                     │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                   │
                                   │ HTTPS (stateless)
                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Cloudflare Workers (Edge)                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐       │
│  │ Auth Worker  │  │Telemetry Wkr │  │   Web (Docs/Changelog)   │       │
│  │              │  │              │  │                          │       │
│  │ - RFC 8628   │  │ - OTLP intake│  │  - Static pages          │       │
│  │ - OAuth flows│  │ - PII redact │  │  - Version info          │       │
│  │ - QR pairing │  │ - 3-tier log │  │  - Update manifest       │       │
│  └──────────────┘  └──────────────┘  └──────────────────────────┘       │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │              Stateful Modules (for Pryx Mesh & Admin)            │   │
│  │  - Durable Objects: session coordinator, device pairing          │   │
│  │  - D1: user registry, device mesh, referrals                     │   │
│  │  - KV: config metadata backup, feature flags                     │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**Key Architectural Patterns** (learned from Clawdbot Gateway):

| Pattern | Implementation |
|---------|----------------|
| **Single Source of Truth** | pryx-core owns all session state; UIs query Gateway |
| **Session Bus** | Pub/sub event bus for real-time multi-surface sync |
| **Role-based Clients** | Surfaces subscribe with scopes (approvals, pairing, etc.) |
| **Defense in Depth** | 5-layer security model for remote control |
| **Event Broadcast** | All surfaces receive state changes simultaneously |
| **Offline Queue** | Each surface maintains local queue; sync on reconnect |

### 8.2 Technology Stack

| Layer | Technology | Rationale |
|-------|------------|-----------|
| **Host** | Rust + Tauri v2 | Native performance, security, sidecar supervision |
| **Runtime** | Go | High concurrency, single binary, fast compilation |
| **TUI** | SolidJS + OpenTUI | Rich terminal UI, component-based, compiled via Bun (OpenTUI build toolchain requires Zig) |
| **Web UI** | Astro + React + TypeScript + Vite | (Optional) Modern DX, SPA |
| **Edge** | Cloudflare Workers (TypeScript) | Global distribution, serverless, low ops burden |
| **Storage** | SQLite + OS Keychain | Local-first, secure secrets, no external DB |
| **Protocol** | MCP (Model Context Protocol) | Standard tool interface, dynamic discovery |
| **Sync** | WebSocket + Event Bus | Real-time multi-surface synchronization |

### 8.3 Session Bus Architecture (Multi-Surface Sync)

**Core Pattern**: Pryx-core acts as the **Gateway** (single source of truth). All surfaces and channels connect to it.

> **Note**: This architecture scales to **Multi-Device** via the Pryx Mesh Coordinator (see `docs/prd/pryx-mesh-design.md`).

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         pryx-core (Gateway)                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                      Session Bus (pub/sub)                        │   │
│  │                                                                   │   │
│  │   Events:                                                         │   │
│  │   - session.message      (new message in session)                 │   │
│  │   - session.typing       (typing indicator)                       │   │
│  │   - tool.request         (tool approval needed)                   │   │
│  │   - tool.executing       (tool running)                           │   │
│  │   - tool.complete        (tool finished)                          │   │
│  │   - approval.needed      (cross-surface approval prompt)          │   │
│  │   - approval.resolved    (user approved/denied)                   │   │
│  │   - trace.event          (telemetry data)                         │   │
│  │   - error.occurred       (error notification)                     │   │
│  │                                                                   │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                              │                                           │
│          ┌───────────────────┼───────────────────────────┐              │
│          ▼                   ▼                           ▼              │
│  ┌──────────────┐   ┌──────────────┐            ┌──────────────┐       │
│  │  CLI/TUI     │   │   Web UI     │            │   Telegram   │       │
│  │  Subscriber  │   │  Subscriber  │    ...     │  Subscriber  │       │
│  └──────────────┘   └──────────────┘            └──────────────┘       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**Sync Guarantees**:

| State | Sync Strategy | Latency Target |
|-------|---------------|----------------|
| Messages | Event broadcast + SQLite persist | <100ms |
| Tool calls | Event broadcast (real-time) | <50ms |
| Approvals | Event broadcast + push notification | <50ms |
| Typing | Ephemeral broadcast (no persistence) | <50ms |
| Session metadata | Polling on reconnect | <1s |

**Event Schema**:
```json
{
  "type": "event",
  "event": "tool.complete",
  "session_id": "session-uuid",
  "surface": "telegram",
  "payload": {
    "tool": "shell.exec",
    "result": "success",
    "duration_ms": 1234
  },
  "timestamp": "2026-01-27T12:00:00Z",
  "version": 42
}
```

**Offline Queue**: Each surface maintains local queue. On reconnect:
1. Get server's last known version
2. Replay queued events with version > server version
3. Request missed events from server
4. Apply missed events locally
5. Clear queue

### 8.4 Constraint Management & Multi-Device Orchestration

**Why This Matters**: Users will ask:
- "How do I handle model limits when agent wants 300k tokens?"
- "Can agent search files on laptop and download to my phone?"
- "How do I run command on device X from device Y?"
- "Can agent send files via WhatsApp to other device?"
- "Which model should I use for this task? There are 600+ options!"

Pryx needs a robust system for handling constraints across providers, devices, and tasks.

---

## 9) Functional Requirements

### FR1: Agent Runtime (pryx-core)

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR1.1 | Headless operation | N/A | `pryx-core` runs without UI on headless servers |
| FR1.2 | Tool extensibility via MCP | N/A | New MCP tool works without host rebuild |
| FR1.3 | Concurrent event handling | ≥50 concurrent | Handle 50 webhooks with <100ms latency p95 |
| FR1.4 | WebSocket event stream | <50ms latency | Tokens, tool calls, approvals streamed in real-time |
| FR1.5 | Local persistence | N/A | Sessions, logs stored in SQLite |
| FR1.6 | Policy-aware execution | 100% coverage | Every tool call evaluated by policy engine |
| FR1.7 | Constraint awareness | Model limits auto-detected and enforced | Max tokens, thinking budget handled |

**FR1 Acceptance Tests**:
```gherkin
Scenario: Headless server deployment
  Given pryx-core is started with --headless flag
  When no UI process is attached
  Then pryx-core serves HTTP API on configured port
  And responds to /health with 200 OK

Scenario: Concurrent webhook handling
  Given pryx-core is running
  When 50 webhook requests arrive simultaneously
  Then all requests complete within 5 seconds
  And p95 latency is under 100ms

Scenario: Model constraint enforcement
  Given user has Claude Sonnet with 200k token limit
  And agent requests 300k tokens for response
  Then request is split into 2 API calls (150k each)
  And total tokens counted accurately across split calls

Scenario: Auto-compaction on thinking budget
  Given thinking token budget is 100k
  And agent uses 80k thinking tokens
  Then system auto-compacts context and frees 20k for next task
```

### FR2: Approvals & Policy Engine

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR2.1 | Policy decision on sensitive tools | 100% | Every fs.write, shell.exec triggers policy check |
| FR2.2 | Approval scopes | N/A | workspace, host, network, integration scopes supported |
| FR2.3 | Approval duration | N/A | once, session, forever (with configurable expiry) |
| FR2.4 | Default policies | N/A | read-only, workspace-only, network-allowlist modes |
| FR2.5 | Native approval dialogs | <500ms | System dialog appears within 500ms of request |

**FR2 Acceptance Tests**:
```gherkin
Scenario: Dangerous tool requires approval
  Given policy is set to "ask" for shell.exec
  When agent attempts shell.exec("rm -rf /tmp/test")
  Then approval dialog appears within 500ms
  And tool execution blocks until user responds

Scenario: Workspace scope enforcement
  Given workspace is set to /home/user/project
  When agent attempts fs.write("/etc/passwd", "...")
  Then request is denied with "outside workspace" error
  And no file modification occurs
```

### FR3: Host (Desktop Shell / Supervisor)

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR3.1 | Sidecar lifecycle | <3s startup | Start pryx-core within 3s of app launch |
| FR3.2 | Crash restart | <5s recovery | Restart crashed sidecar within 5s with backoff |
| FR3.3 | Graceful shutdown | 0 zombies | SIGTERM propagated, no orphan processes |
| FR3.4 | Port discovery | <1s | Read sidecar port from stdout/metadata within 1s |
| FR3.5 | Deep linking | N/A | `pryx://session/abc` opens specific session |
| FR3.6 | Permission gating | N/A | Critical tools require native dialog approval |
| FR3.7 | Cross-device initiation | N/A | Pair new device via QR/device code from CLI |
| FR3.8 | Remote command routing | <2s | Route commands to device via SSH/WebDAV within 2s |

**FR3 Acceptance Tests**:
```gherkin
Scenario: No zombie processes after exit
  Given Pryx app is running with sidecar
  When user quits application
  Then main process exits within 2 seconds
  And no pryx or pryx-core processes remain
  And all child processes are terminated

Scenario: Cross-device command routing
  Given device A (laptop) paired with device B (server)
  When user on device A requests file operation on device B
  Then command is routed via SSH/WebDAV within 2 seconds
  And operation completes on device B with response streamed back
  And audit log records cross-device action with timestamps

Scenario: Pryx Mesh session continuation
  Given user has active session on laptop
  And initiates session on phone via WhatsApp
  Then Pryx Mesh discovers and connects both devices
  And session context is synchronized between devices
  And tool calls can be initiated from either device
```

### FR4: UI (Control Center)

**FR3 Acceptance Tests**:
```gherkin
Scenario: No zombie processes after exit
  Given Pryx app is running with sidecar
  When user quits the application
  Then main process exits within 2 seconds
  And no pryx or pryx-core processes remain
  And all child processes are terminated

Scenario: Crash restart with backoff
  Given pryx-core crashes 3 times in 60 seconds
  Then restart delay increases: 1s, 2s, 4s
  And after 5 failures, user is notified
```

### FR4: UI (Control Center)

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR4.1 | Onboarding wizard | <3 min | First integration connected in under 3 minutes |
| FR4.2 | Visual configuration | 0 JSON edits | All settings configurable via UI |
| FR4.3 | Session explorer | <1s search | Search 1000 sessions returns in <1s |
| FR4.4 | Trace timeline | Real-time | Gantt chart updates within 100ms of events |
| FR4.5 | Cost display | Per-session | Token count and estimated cost shown |
| FR4.6 | Policy management | N/A | Define scopes, allowlists, approval modes in UI |
| FR4.7 | Long-running task status | Indeterminate progress | When agent executes long operation (>10s), show persistent status indicator with: current step, progress bar (if applicable), stop button, notification on complete |

**FR4 Acceptance Tests**:
```gherkin
Scenario: Onboarding wizard completion
  Given user opens Pryx for first time
  When user follows onboarding wizard
  Then workspace is configured in step 1
  And model provider is selected in step 2
  And first integration connects in step 3
  And total time is under 3 minutes

Scenario: Trace timeline visualization
  Given agent is executing a multi-tool task
  When tool calls are made
  Then Gantt chart shows each tool as a bar
  And bars update in real-time (<100ms delay)
```

### FR5: Integrations & Channels

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR5.1 | Channel abstraction | N/A | Unified interface for all channels |
| FR5.2 | MVP channels | 2 channels | Telegram (cloud-hosted webhook mode) + generic webhooks functional |
| FR5.3 | Routing rules | N/A | Map channels to agents/workspaces |
| FR5.4 | Rate limiting | Configurable | Per-channel rate limits enforced |
| FR5.5 | Status indicators | Real-time | Connection state visible in UI |
| FR5.6 | Reconnect logic | <30s | Auto-reconnect on disconnect within 30s |

**FR5 Acceptance Tests**:
```gherkin
Scenario: Telegram integration (cloud-hosted)
  Given user has completed registration
  And user has configured a model provider key (e.g., OpenRouter for GLM-4.x)
  And user has a Telegram bot token (from BotFather)
  When user connects Telegram integration in onboarding wizard
  Then Pryx verifies bot token using Telegram getMe
  And Pryx sets a webhook to Pryx Edge with a per-bot secret
  And user can link a chat using /start
  And incoming messages route to configured agent and produce replies

Scenario: Telegram integration (device-hosted, no key stored in cloud)
  Given user has installed pryx-core locally
  And user configured their model provider via local env/OS keychain
  And user has a Telegram bot token
  When user enables Telegram channel in local pryx-core
  Then pryx-core polls Telegram updates and processes messages locally
  And no model API key is sent to Pryx cloud

Scenario: Rate limiting
  Given channel rate limit is 10 requests/minute
  When 15 requests arrive in 1 minute
  Then first 10 are processed
  And remaining 5 return 429 status
```

#### Telegram Bot Operational Requirements (Clarified)

Telegram does not “run” user code. It only delivers updates to a bot via either **webhook** (recommended) or **long polling**. This yields two viable execution models:

1) **Cloud-hosted webhook mode (recommended for MVP)**  
   - **User does not need pryx-core installed**.  
   - Pryx hosts the bot logic and receives updates from Telegram via HTTPS webhook.  
   - **Constraint**: Telegram webhooks cannot include a user-provided Authorization header, so Pryx must either:
     - store the user’s model provider key (encrypted at rest) and use it server-side, or
     - route processing to a user-hosted runtime (device-hosted mode).

2) **Device-hosted polling mode (privacy-first / offline-ish)**  
   - User runs pryx-core which polls Telegram (getUpdates) and calls the AI model locally using the user’s key.  
   - **Constraint**: polling must run on exactly one device per bot token to avoid duplicated handling; Mesh can coordinate “active bot host”.

**Minimum Pryx-hosted components for cloud-hosted mode**:
- Edge API that implements Telegram webhook receiver + integration management endpoints.
- Persistent storage for: bot tokens (encrypted), per-bot webhook secret, chat links, routing config, per-user model keys (encrypted) if BYOK-in-cloud is enabled.
- Worker-side message processing pipeline: dedupe updates, policy checks, call model, send reply.

**Minimum user-provided inputs**:
- Telegram bot token (BotFather)
- Linked chat (via /start, allowlist, or explicit chat binding)
- Model provider credentials (e.g., OpenRouter key for GLM-4.x / GLM-4.7) depending on chosen execution model

**Proposed Edge API surface (cloud-hosted mode)**:
- `POST /api/v1/integrations/telegram/bots` (user auth) create bot integration, verify token, persist encrypted token
- `POST /api/v1/integrations/telegram/bots/{bot_id}/sync-webhook` (user auth) setWebhook + secret token
- `POST /api/v1/integrations/telegram/bots/{bot_id}/link-code` (user auth) generate chat link code
- `POST /api/v1/integrations/telegram/webhook/{bot_id}` (Telegram webhook) verify secret header, ingest updates
- `POST /api/v1/integrations/telegram/bots/{bot_id}/pause|resume|rotate-secret|delete` (user auth) lifecycle ops

### FR6: Auth & Edge Control Plane

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR6.1 | OAuth Device Flow | RFC 8628 | Complete device flow with QR code |
| FR6.2 | Stateless workers | 0 persistence | Workers do not store conversation data |
| FR6.3 | Telemetry ingestion | OTLP | Accept OTLP batches, redact PII |
| FR6.4 | PII redaction | 100% | All PII fields redacted before forwarding |
| FR6.5 | Token rotation | 24h max | Refresh tokens expire within 24 hours |

**FR6 Acceptance Tests**:
```gherkin
Scenario: Device flow authentication
  Given user initiates OAuth for GitHub
  When device code is generated
  Then QR code displays in UI
  And user scans with phone
  And authorization completes within 5 minutes
  And tokens stored in OS keychain

Scenario: PII redaction
  Given telemetry contains email "user@example.com"
  When trace is sent to telemetry worker
  Then forwarded trace contains "[REDACTED_EMAIL]"
```

### FR7: Installation, Updates, and Ops

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR7.1 | One-liner install | <60s | `curl \| sh` completes in under 60 seconds |
| FR7.2 | No external deps | 0 deps | Works without Node.js, Python, Docker |
| FR7.3 | Native bundles | 3 platforms | macOS (.dmg), Windows (.exe), Linux (.AppImage) |
| FR7.4 | Auto-update | Background | Support Main/Beta/Alpha build channels; check on startup; background download; toast notifications |
| FR7.4.1 | Build channel selection | N/A | Users can select Main, Beta, or Alpha channel |
| FR7.4.2 | Version check on startup | <5s | Check for updates within 5 seconds of startup |
| FR7.4.3 | Background download | N/A | Download updates while user continues using Pryx |
| FR7.4.4 | Toast notifications | N/A | Show "Update Available", "Download Progress", "Update Ready" toasts |
| FR7.4.5 | What's New modal | N/A | Display release notes after successful update |
| FR7.4.6 | Graceful restart | <30s | Apply update and restart within 30 seconds of user approval |
| FR7.4.7 | Build channel switching | N/A | Allow users to switch between Main, Beta, Alpha channels |
| FR7.5 | Safe rollback | <30s | Rollback to previous version in under 30s |
| FR7.6 | Service mode | systemd/launchd | `pryx install-service` works on macOS/Linux |
| FR7.7 | Clean uninstall | Complete | Remove binaries, service, optionally data |

**FR7 Acceptance Tests**:
```gherkin
Scenario: One-liner installation
  Given fresh macOS/Linux machine
  When user runs "curl https://get.pryx.dev | sh"
  Then pryx binary is installed to ~/.local/bin or /usr/local/bin
  And pryx command is available in PATH
  And "pryx doctor" reports healthy
  And total time is under 60 seconds

Scenario: Auto-update with rollback
  Given Pryx v1.0.0 is installed
  And v1.1.0 is available
  When update is applied
  And v1.1.0 crashes on startup
  Then Pryx auto-rolls back to v1.0.0
  And user is notified of rollback
```

### FR8: Skills Management

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR8.1 | Skill discovery | <1s | Load all skills from managed + workspace dirs within 1s |
| FR8.2 | SKILL.md format | N/A | YAML frontmatter + Markdown body per clawdbot convention |
| FR8.3 | Skill precedence | N/A | workspace > managed > bundled (higher overrides lower) |
| FR8.4 | Skill CLI | N/A | `pryx skills list|info|check|enable|disable` |
| FR8.5 | Skill requirements | N/A | Declare bins, env vars, config dependencies |
| FR8.6 | Skill installers | N/A | Support brew, npm, go, uv, download installers |
| FR8.7 | Token efficiency | N/A | Only skill metadata in system prompt; body loaded on-demand |

**SKILL.md Format**:
```markdown
---
name: my-skill
description: What triggers this skill
metadata:
  pryx:
    emoji: "🔧"
    requires:
      bins: ["jq"]        # Required on PATH
      env: ["API_KEY"]    # Required env vars
    install:
      - id: brew
        kind: brew
        formula: jq
        bins: ["jq"]
---
# My Skill

Instructions for the agent when this skill is triggered...
```

**Skill Locations**:
| Location | Scope | Precedence |
|----------|-------|------------|
| `<workspace>/.pryx/skills/` | Per-workspace | Highest |
| `~/.pryx/skills/` | User-managed (all workspaces) | Medium |
| Bundled in binary | Built-in | Lowest |

**FR8 Acceptance Tests**:
```gherkin
Scenario: Skill discovery and loading
  Given skill "formatter" exists in ~/.pryx/skills/formatter/SKILL.md
  When pryx starts a session
  Then skill metadata is loaded within 1 second
  And skill appears in "pryx skills list"
  And skill can be triggered by agent

Scenario: Workspace skill override
  Given skill "linter" exists in both ~/.pryx/skills/ and <workspace>/.pryx/skills/
  When agent loads skills for workspace
  Then workspace version takes precedence
  And managed version is shadowed
```

### FR9: MCP Client Integration

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR9.1 | Native MCP client | N/A | pryx-core implements MCP client directly (no external CLI) |
| FR9.2 | Transport: stdio | N/A | Support MCP servers via stdin/stdout |
| FR9.3 | Transport: HTTP | N/A | Support MCP servers via HTTP/SSE |
| FR9.4 | Server discovery | <2s | List available tools from configured servers within 2s |
| FR9.5 | Tool execution | N/A | MCP tools exposed to agent like built-in tools |
| FR9.6 | Server config UI | N/A | Add/remove MCP servers via Web UI |
| FR9.7 | Auth support | N/A | OAuth for servers that require it |
| FR9.8 | Manual config | N/A | Users can edit `servers.json` directly if preferred |

**MCP Server Priorities** (from research on 50+ MCP servers):

**Tier 1: Must Have for MVP** (covers 80% of use cases)

| Server | Purpose | Use Cases |
|--------|---------|-----------|
| **filesystem** | Read/write files, directory operations | All use cases |
| **shell** | Execute commands, scripts | Trading, server mgmt, code dev |
| **browser** | Playwright-based browser control | Research, trading, automation |
| **clipboard** | Copy/paste between applications | All use cases |

**Tier 2: High Value for v1**

| Server | Purpose | Use Cases |
|--------|---------|-----------|
| **desktop-automation** | UI automation (clicks, keyboard) | Excel/Office, trading platforms |
| **screen-capture** | Screenshots, OCR | Debugging, research |
| **notifications** | System notifications | All use cases |
| **http** | Make HTTP requests | Trading APIs, webhooks |

**Tier 3: Specialized (User-installed)**

| Server | Purpose | Use Cases |
|--------|---------|-----------|
| **office** | Excel, Word, PowerPoint | Office automation |
| **git** | Git operations | Code development |
| **docker** | Container management | Server management |
| **ssh** | Remote server access | Server management |
| **trading** | Alpaca, Binance, etc. | Trading automation |

**MCP Server Hosting**:
- **Bundled** (Tier 1): Compiled into pryx-core binary
- **Sidecar** (Tier 2): Separate process started on-demand
- **User-installed** (Tier 3): Configured in `servers.json`

**MCP Server Configuration** (`~/.pryx/mcp/servers.json`):
```json
{
  "servers": {
    "linear": {
      "transport": "http",
      "url": "https://api.linear.app/mcp",
      "auth": {
        "type": "oauth",
        "token_ref": "keychain:linear-token"
      }
    },
    "custom-tool": {
      "transport": "stdio",
      "command": ["node", "./my-mcp-server.js"],
      "cwd": "/path/to/server"
    }
  }
}
```

**FR9 Acceptance Tests**:
```gherkin
Scenario: MCP server connection
  Given MCP server "linear" is configured with valid auth
  When pryx-core starts
  Then server connects within 5 seconds
  And tools from server appear in available tools list

Scenario: MCP tool execution
  Given MCP server exposes tool "list_issues"
  When agent calls list_issues with arguments
  Then request is routed to MCP server
  And response is returned to agent within 10 seconds
  And policy engine evaluates call before execution

Scenario: Stdio MCP server
  Given MCP server configured with transport: stdio
  When pryx-core needs to call a tool
  Then server process is spawned if not running
  And communication happens via stdin/stdout
  And process is kept alive for subsequent calls
```

### FR10: Multi-Device Coordination (Pryx Mesh)

> **Detailed Specification**: See `docs/prd/pryx-mesh-design.md` for full protocol and security model.

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR10.1 | Device Pairing | RFC 8628 | Pair new device via 6-digit code or QR code |
| FR10.2 | Session Handoff | <2s latency | Messages typed on Device A appear on Device B instantly |
| FR10.3 | Cross-Device Exec | 100% Policy | Executing command on Device B from Device A requires signed approval |
| FR10.4 | API Key Sync | E2EE | API keys synced securely (never decrypted in cloud) |
| FR10.5 | Integration Sharing | Mesh-Global | Telegram bot hosted on Server acts for Laptop user |

**FR10 Acceptance Tests**:
```gherkin
Scenario: Cross-Device Command Execution
  Given "Laptop" and "Server" are paired in Pryx Mesh
  When user on Laptop says "Run 'docker ps' on Server"
  Then Server receives execution request
  And Server policy engine prompts for approval (if configured)
  And Server executes command
  And Output streams back to Laptop UI
```

### FR11: Scheduled Tasks & Automation

> **Scope**: v1.1+ (deferred from MVP, first post-MVP release)

**User Stories**:
- "Set up a job that watches my stock every 4 hours and saves to Google Sheet"
- "Monitor log files continuously and alert me on errors"
- "Run tests every time I push to GitHub"
- "Debug my app continuously and auto-fix until no errors"

**Architecture**:
```
┌─────────────────────────────────────────────────────────┐
│           Task Scheduler (pryx-core)                   │
├─────────────────────────────────────────────────────────┤
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  │
│  │  Cron      │  │  Event     │  │  Manual    │  │
│  │  Trigger   │  │  Trigger   │  │  Trigger   │  │
│  └────────────┘  └────────────┘  └────────────┘  │
│         │              │              │              │
│         └──────────────┴──────────────┘              │
│                        ▼                             │
│  ┌─────────────────────────────────────┐            │
│  │     Persistent Task Queue          │            │
│  │  (SQLite + recover on restart)    │            │
│  └─────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────┘
```

| ID | Requirement | Metric | Acceptance Criteria |
|----|-------------|--------|---------------------|
| FR11.1 | Task Dashboard UI | <1s load | View all scheduled tasks with status, last run, next run |
| FR11.2 | Cron Expression Support | N/A | Every X hours, cron syntax, event-based triggers supported |
| FR11.3 | Task History | 100 runs per task | View logs of historical task runs with results |
| FR11.4 | Notification System | <5s delivery | Push notification on task completion/failure via all channels |
| FR11.5 | Cost Per Task | Per-run basis | Show token cost for each scheduled task execution |
| FR11.6 | Pause/Resume | Instant | User can pause individual tasks without deleting |
| FR11.7 | Task Templates | N/A | Pre-built templates (stock monitor, log watcher, etc.) |
| FR11.8 | Cross-Device Execution | <2s latency | Scheduled task on Device A can execute on Device B via Mesh |
| FR11.9 | Task Persistence | 100% | Scheduled tasks survive application restart/crash |
| FR11.10 | Retry Policies | Configurable | Exponential backoff, max retries, on-failure actions |

**FR11 Acceptance Tests**:
```gherkin
Scenario: Create scheduled task
  Given user opens Scheduled Tasks dashboard
  When user creates "Stock Watch" task with 4-hour trigger
  Then task appears in list with "active" status
  And next run time is calculated correctly
  And task persists after pryx restart

Scenario: Task execution and notification
  Given scheduled task triggers
  When task completes successfully
  Then execution is logged in task history
  And cost is recorded
  And notification sent to Telegram/Web UI
  And next run time is scheduled

Scenario: Cross-device scheduled task
  Given "Server" is paired in Pryx Mesh
  When user on Laptop creates task to run on Server
  Then task executes on Server
  And result streams back to Laptop UI
  And failure on Server triggers notification to Laptop

Scenario: Task with retry policy
  Given task has retry policy: 3 retries, exponential backoff
  When task fails on first attempt
  Then retry 1 after 1 min, retry 2 after 2 min, retry 3 after 4 min
  And after 3rd failure, mark task as failed and notify user

Scenario: Pause and resume task
  Given task is active and running every hour
  When user pauses task
  Then task status changes to "paused"
  And next run is cancelled
  When user resumes task
  Then task status changes to "active"
  And next run is recalculated from now
```

**UI Surfaces**:
```
┌─────────────────────────────────────────────────────────────┐
│  📊 Scheduled Tasks (Web UI/TUI)                      │
├─────────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  [+ New Task]  [Templates]  [Bulk Actions]    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  📈 Stock Watch - AAPL              Active   │   │
│  │    Trigger: Every 4 hours                      │   │
│  │    Next run: In 1 hour 23 minutes             │   │
│  │    Last run: 3 hours ago (✅ Success)         │   │
│  │    Cost this month: $0.45                      │   │
│  │    Target Device: Server                       │   │
│  │    [History] [Pause] [Edit] [Delete]         │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │  📝 Log Monitor - Server              Active   │   │
│  │    Trigger: Event-based (file change)         │   │
│  │    Next run: Waiting for event               │   │
│  │    Last run: 15 minutes ago (✅ Success)      │   │
│  │    Cost this month: $0.12                      │   │
│  │    [History] [Pause] [Edit] [Delete]         │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │  🧪 Test Runner - Workspace          Paused   │   │
│  │    Trigger: Cron: 0 */2 * * * (every 2h)   │   │
│  │    Paused by: user on 2026-01-27            │   │
│  │    [History] [Resume] [Edit] [Delete]         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                         │
└─────────────────────────────────────────────────────────────┘
```

**Task History View**:
```
┌─────────────────────────────────────────────────────────────┐
│  📊 Stock Watch - AAPL > Execution History            │
├─────────────────────────────────────────────────────────────┤
│                                                         │
│  Run #42 - 2026-01-27 13:00:00 - ✅ Success    │
│    Duration: 2.3s | Tokens: 1,245 | Cost: $0.02   │
│    Result: Stock price $178.42, news fetched          │
│                                                         │
│  Run #41 - 2026-01-27 09:00:00 - ✅ Success    │
│    Duration: 2.1s | Tokens: 1,198 | Cost: $0.02   │
│    Result: Stock price $176.89, news fetched          │
│                                                         │
│  Run #40 - 2026-01-27 05:00:00 - ❌ Failed      │
│    Duration: 1.2s | Tokens: 892 | Cost: $0.01     │
│    Error: Timeout fetching stock price               │
│    Retry #1: Failed (same error)                    │
│    Retry #2: Failed (same error)                    │
│    Action taken: Notified user, paused task          │
│                                                         │
│  [Download CSV] [View Full Logs] [Restart Task]     │
│                                                         │
└─────────────────────────────────────────────────────────────┘
```

**Task Configuration Schema** (SQLite):
```sql
CREATE TABLE scheduled_tasks (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    trigger_type TEXT NOT NULL, -- 'cron', 'interval', 'event'
    trigger_config JSONB NOT NULL,
    action_config JSONB NOT NULL,
    target_device_id UUID, -- NULL = local, otherwise remote device
    retry_policy JSONB,
    notification_config JSONB,
    status TEXT NOT NULL, -- 'active', 'paused', 'failed', 'disabled'
    next_run TIMESTAMP,
    last_run TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE task_runs (
    id UUID PRIMARY KEY,
    task_id UUID REFERENCES scheduled_tasks(id) ON DELETE CASCADE,
    run_number INTEGER,
    attempt_number INTEGER,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    status TEXT NOT NULL, -- 'success', 'failed', 'cancelled'
    result JSONB,
    error TEXT,
    tokens_used INTEGER,
    cost_usd DECIMAL(10,4),
    execution_device_id UUID
);
```

**Notification Channels**:
- **Desktop**: Native notification (Tauri API)
- **Telegram**: Message to chat
- **Discord/Slack**: Direct message to user
- **Email**: Email report (v2.0+)
- **Web UI**: Toast notification + bell icon badge

**Pre-built Task Templates** (v1.1):
```
┌─────────────────────────────────────────────────────────┐
│  📦 Task Templates                                 │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  💹 Stock Monitor                                     │
│     Watch stock price every X hours, save to sheet    │
│     Requires: Trading API or Web tool                 │
│                                                         │
│  📝 Log Monitor                                       │
│     Monitor log files, alert on errors/patterns        │
│     Requires: filesystem.read, regex matching           │
│                                                         │
│  🧪 Test Runner                                       │
│     Run tests on Git push or cron schedule            │
│     Requires: git, shell.exec                        │
│                                                         │
│  🔄 Continuous Debug                                   │
│     Run continuously, auto-fix until no errors         │
│     Requires: shell.exec, logs, error patterns        │
│                                                         │
│  📊 Data Sync                                         │
│     Sync data from API to database regularly           │
│     Requires: http, filesystem.write                  │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 10) Non-Functional Requirements

### 10.1 Security & Privacy

| ID | Requirement | Implementation |
|----|-------------|----------------|
| NFR-S1 | Local data sovereignty | All conversations/indices on local disk unless export enabled |
| NFR-S2 | Keychain secrets | API keys in OS keychain (never plaintext config) |
| NFR-S3 | Least privilege | Tool execution sandboxed to workspace by default |
| NFR-S4 | Audit logging | Immutable local audit log of tool calls + approvals |
| NFR-S5 | Signed artifacts | Release binaries signed, integrity verified on update |
| NFR-S6 | Network allowlists | Configurable domain allowlists for network tools |
| NFR-S7 | Defense in depth | 5-layer security model for remote control |

### 10.1.1 Security Layers (Defense in Depth)

**Remote control = RCE by design.** Must implement defense in depth:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       Security Layers                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Layer 1: Channel Authentication                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Telegram: Bot token + user allowlist                             │   │
│  │  WhatsApp: Verified account + number allowlist                    │   │
│  │  Web UI:   Localhost-only OR OAuth + device pairing               │   │
│  │  CLI/TUI:  OS user (implicit trust on local machine)              │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                              ▼                                           │
│  Layer 2: Session Authorization                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Each channel maps to a workspace + policy set                    │   │
│  │  "Telegram user X" → workspace:/trading, policy:trading-safe     │   │
│  │  "WhatsApp user Y" → workspace:/research, policy:read-only       │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                              ▼                                           │
│  Layer 3: Tool Policy Engine                                             │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Every tool call evaluated:                                       │   │
│  │  ├── Workspace scope: Is path within allowed workspace?          │   │
│  │  ├── Tool allowlist: Is this tool enabled for this policy?       │   │
│  │  ├── Argument validation: Are arguments within safe bounds?      │   │
│  │  └── Rate limiting: Is user within rate limits?                  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                              ▼                                           │
│  Layer 4: Human-in-the-Loop                                              │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Sensitive operations require explicit approval:                  │   │
│  │  ├── shell.exec (any command)                                     │   │
│  │  ├── filesystem.write (outside workspace)                         │   │
│  │  ├── http.request (external domains)                              │   │
│  │  ├── browser.navigate (sensitive domains)                         │   │
│  │  └── Any tool with side effects                                   │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                              ▼                                           │
│  Layer 5: Audit Trail                                                    │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Immutable local log of all actions:                              │   │
│  │  ├── What: tool_name, arguments (redacted sensitive)              │   │
│  │  ├── Who: channel, user_id                                        │   │
│  │  ├── When: timestamp                                              │   │
│  │  ├── Outcome: success/failure, result hash                        │   │
│  │  └── Approval: who approved, when, scope                          │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**Cross-Surface Approval Flow**:
```
User (Telegram): "Execute the trading bot script"

1. Telegram Adapter → pryx-core: tool_call(shell.exec, "python trade.py")

2. Policy Engine evaluates:
   - Is shell.exec allowed? → Yes, but only python
   - Is "python trade.py" matching pattern? → Yes
   - Is workspace valid? → Yes, ~/trading/trade.py exists
   
3. Policy says: ASK for shell.exec
   
4. Approval broadcast to ALL connected surfaces:
   - Telegram: "⚠️ Approve: python trade.py? Reply /approve or /deny"
   - Web UI:   Native dialog with command preview
   - TUI:      Inline prompt
   - Desktop:  System notification + dialog

5. User approves on ANY surface → approval recorded

6. Tool executes, result broadcast to all surfaces
```

**Policy Configuration Example** (`~/.pryx/policies/trading-safe.yaml`):
```yaml
name: trading-safe
description: Policy for trading via Telegram

workspace:
  root: ~/trading
  allow_outside: false

tools:
  allowed:
    - filesystem.read
    - http.request:
        domains: [api.binance.com, api.coinbase.com]
    - shell.exec:
        commands: [python, node]
  
  ask:
    - browser.*
    - clipboard.*
  
  denied:
    - shell.exec:
        commands: [rm, sudo, chmod]

rate_limits:
  requests_per_minute: 30
  tool_calls_per_minute: 10

approval:
  timeout: 5m
  default: deny
```

### 10.2 Reliability

| ID | Requirement | Metric |
|----|-------------|--------|
| NFR-R1 | No zombie processes | 0 orphans after app exit |
| NFR-R2 | Crash isolation | Sidecar crash does not crash host |
| NFR-R3 | Automatic restart | Crashed sidecar restarts within 5s |
| NFR-R4 | Graceful shutdown | All processes exit within 3s of quit signal |
| NFR-R5 | Offline resilience | Queue channel messages during sidecar restart |
| NFR-R6 | Session recovery | Resume session after crash/restart |

### 10.3 Performance

| ID | Requirement | Metric |
|----|-------------|--------|
| NFR-P1 | Time to first chat | <2 minutes from install |
| NFR-P2 | UI startup | Visible within 3 seconds |
| NFR-P3 | Sidecar startup | Ready within 2 seconds |
| NFR-P4 | Search latency | <1s for 1000 sessions |
| NFR-P5 | WebSocket latency | <50ms for event streaming |
| NFR-P6 | Memory footprint | <200MB idle (host + sidecar) |

### 10.4 Scalability

| ID | Requirement | Metric |
|----|-------------|--------|
| NFR-SC1 | Concurrent channels | ≥10 channels simultaneously |
| NFR-SC2 | Session storage | ≥10,000 sessions without degradation |
| NFR-SC3 | Trace storage | ≥1M trace events before rotation |

### 10.5 Memory Management (NEW)

| ID | Requirement | Implementation |
|----|-------------|----------------|
| NFR-M1.1 | Context window tracking | Monitor tokens per session, warn at 80% of model context limit |
| NFR-M1.2 | Automatic summarization | Summarize oldest 20% of messages when approaching limit (90%+) |
| NFR-M1.3 | Session archival | Archive completed sessions to disk, keep recent sessions in memory |
| NFR-M1.4 | RAG integration (optional) | Optional retrieval from vector DB for long-term memory beyond context window |
| NFR-M1.5 | Conversation branching | Create child sessions for distinct topics to preserve context |
| NFR-M1.6 | Token cost awareness | Estimated token cost shown before long-running operations |
| NFR-M1.7 | Cross-device memory sync | Hybrid sync: real-time for active session, encrypted blob for history |

**Memory Management Architecture**:
```go
// Pseudocode for context window management
type Session struct {
    id           string
    tokensUsed   int
    maxTokens    int // from model config (e.g., 200K for Claude)
    messages     []Message
}

func (s *Session) AddMessage(msg Message) error {
    // Check if we're approaching limit
    if s.tokensUsed + msg.Tokens > s.maxTokens {
        if s.tokensUsed + msg.Tokens > int(float64(s.maxTokens)*0.9) {
            // At 90%, summarize oldest 20%
            summarized := s.summarizeOldestMessages(0.2)
            s.messages = append(summarized, s.messages[len(summarized):]...)
            s.tokensUsed = calculateTokens(s.messages)
            // Notify user via UI
            s.notifyUser("Session summarized to fit context window")
        } else {
            // At 80%, just warn
            s.notifyUser("⚠️ Context window at 80% capacity")
        }
    }

    s.messages = append(s.messages, msg)
    s.tokensUsed += msg.Tokens

    // Update SQLite
    return s.persist()
}

func (s *Session) summarizeOldestMessages(ratio float64) []Message {
    numToSummarize := int(float64(len(s.messages)) * ratio)
    oldest := s.messages[:numToSummarize]
    summary := s.llm.Summarize(oldest)

    // Create summary message
    summaryMsg := Message{
        Role: "system",
        Content: fmt.Sprintf("[SUMMARY of %d messages]: %s", numToSummarize, summary),
        Tokens: estimateTokens(summary),
        IsSummary: true,
    }

    return []Message{summaryMsg}
}
```

**Multi-Device Memory Sync**:
```
┌─────────────────────────────────────────────────────────────────┐
│              MEMORY SYNC ACROSS DEVICES                       │
├─────────────────────────────────────────────────────────────────┤
│                                                            │
│  HOT SYNC (Active Session):                                  │
│  ┌────────────────────────────────────────────────────────┐   │
│  │  Real-time WebSocket broadcast                     │   │
│  │  - User types on Device A                        │   │
│  │  - Event: session.message                        │   │
│  │  - Device B receives <100ms latency              │   │
│  │  - Display appears instantly                        │   │
│  └────────────────────────────────────────────────────────┘   │
│                                                            │
│  WARM SYNC (Session History):                                │
│  ┌────────────────────────────────────────────────────────┐   │
│  │  Encrypted blob sync to Cloud D1 (on idle)      │   │
│  │  - Session stored locally (SQLite)               │   │
│  │  - On idle, encrypt session blob                  │   │
│  │  - Upload to Cloud KV (encrypted)                 │   │
│  │  - Device B downloads on connection              │   │
│  │  - Decrypts with shared Master Key              │   │
│  │  - Merges with local state                     │   │
│  └────────────────────────────────────────────────────────┘   │
│                                                            │
│  CONFLICT RESOLUTION:                                        │
│  ┌────────────────────────────────────────────────────────┐   │
│  │  Coordinator Time (sequence_id) is source of truth│   │
│  │  - Every event gets monotonic sequence from DO    │   │
│  │  - Devices apply events in order                │   │
│  │  - Divergent states merge: "Coordinator wins"   │   │
│  └────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 10.6 Task Queue Persistence (NEW)

| ID | Requirement | Metric |
|----|-------------|--------|
| NFR-M2.1 | Scheduled tasks survive restart | 100% of tasks persist across application restart |
| NFR-M2.2 | Task state recovery | Resume in-progress tasks after restart |
| NFR-M2.3 | Cron scheduler resilience | Wake on missed schedule within 5 min |
| NFR-M2.4 | Task execution isolation | Task failure doesn't crash scheduler |
| NFR-M2.5 | Cross-device task handoff | Task scheduled on Device A can execute on Device B |

### 10.7 Autocompletion & Long-Running Task Management (NEW)

**Why This Matters**: Long-running continuous tasks (e.g., "watch logs and auto-fix", "monitor POS transactions") present token efficiency challenges.

**User Concerns**:
- "Will continuous monitoring consume too many tokens?"
- "If I'm working in OpenCode while Pryx runs in background, how do updates work?"
- "What's the UX for agent waiting for user attention?"
- "How does the agent handle context window for multi-hour operations?"

---

#### 10.7.1 Background Process Management (OpenCode Pattern)

**Core Capability**: Long-running processes (log monitoring, file watching, cron jobs) run independently with controlled output streaming.

**Architecture**:
```
┌─────────────────────────────────────────────────────────────┐
│  User (OpenCode/IDE)                             │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Pryx Agent (User's Device)                         │
├─────────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Background Process Manager                  │   │
│  │  • createBackgroundProcess(...)            │   │
│  │  • listBackgroundProcesses()             │   │
│  │  • killProcesses()                        │   │
│  └──────────────────────────────────────────────────┘   │
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Token Optimization Layer                   │   │
│  │  • Context pruning                       │   │
│  │  • Cache TTL management                  │   │
│  │  • Output stream limiting               │   │
│  └──────────────────────────────────────────────────┘   │
│                                                         │
│  ▼                                                    │
┌─────────────────────────────────────────────────────┐
│  Log File Stream (Continuous Monitoring)          │
└─────────────────────────────────────────────────────────────┘
```

**Background Process Features**:
- **In-memory task tracking** with unique process IDs
- **Output stream limiting**: Last 100 lines for individual task, last 10 lines for task list
- **Tag-based filtering** for categorization (monitoring, debugging, deployment)
- **Global vs session-specific** processes
- **Automatic cleanup** on session end

---

#### 10.7.2 Token Efficiency Strategies (Clawdbot Pattern)

**Core Principles**:
1. **Compact history**: Use `/compact` to summarize long sessions
2. **Trim tool outputs**: Limit verbose tool results
3. **Short skill descriptions**: Reduce system prompt overhead
4. **Prefer smaller models**: For verbose, exploratory work

**Cache TTL Management**:
```yaml
agents:
  defaults:
    model:
      primary: "anthropic/claude-opus-4-5"
    models:
      "anthropic/claude-opus-4-5":
        params:
          cacheControlTtl: "1h"
  heartbeat:
      every: "55m"  # Keep cache warm just under TTL
```

**Session Pruning**:
- **Only prunes tool results** (not user/assistant messages)
- **Protects last N assistant messages** (default: 3)
- **Soft-trim**: Keeps head + tail, inserts `...`
- **Hard-clear**: Replaces entire result with placeholder `[Old tool result content cleared]`
- **Skips image blocks** (never trimmed)

---

#### 10.7.3 Streaming vs "Pump-and-Dump" Patterns

**Block Streaming** (coarse, efficient for long operations):
- Emits **completed blocks** as assistant writes
- Not token deltas, but coarse chunks
- Controlled by:
  - `blockStreamingDefault`: `"on"` or `"off"`
  - `blockStreamingBreak`: `"text_end"` or `"message_end"`
  - `blockStreamingChunk`: `{ minChars, maxChars }`

**Chunking Algorithm**:
- **Low bound**: Don't emit until buffer >= minChars
- **High bound**: Prefer splits before maxChars
- **Break preference**: `paragraph` → `newline` → `sentence` → `whitespace` → hard break
- **Code fences**: Never split inside ``` blocks

**Coalescing**:
- Wait for idle gaps before flushing chunks
- Reduces "single-line spam" while providing progressive output
- Default: min 150 chars, max 3000 chars, 200ms idle

---

#### 10.7.4 Heartbeat System for Continuous Operations

**Purpose**: Keep prompt cache warm across idle gaps, reducing cache write costs.

**Configuration**:
```yaml
heartbeat:
  every: "30m"  # Default, configurable
  ackMaxChars: 200  # Padding for HEARTBEAT_OK
  includeReasoning: false  # Set to true to send separate Reasoning message
```

**How It Works**:
```
1. Read HEARTBEAT.md if it exists (workspace context)
2. Follow it strictly - Do not infer or repeat old tasks
3. If nothing needs attention → reply HEARTBEAT_OK
4. Inline directives in heartbeat message apply as usual
5. Send heartbeat only (no delivery to avoid double delivery)
```

**Delivery**:
- Heartbeat probe body is configured to deliver to **final payload only**
- To send Reasoning as separate message, set `includeReasoning: true`
- If ack within limits, suppress delivery to avoid spam

**Benefits**:
- Reduces cache write costs during idle periods
- Prevents prompt cache evictions
- Maintains context between long-running operations

---

#### 10.7.5 Agent Waiting UX Patterns

**Scenario**: Agent completes task, waiting for user instruction/continue signal.

**Implementation**:
```typescript
agent.wait uses waitForAgentJob:
  - waits for **lifecycle end/error** for runId
  - returns: { 
      status: "ok" | "error" | "timeout", 
      startedAt, 
      endedAt, 
      error? 
    }
```

**Key Features**:
- **Separate timeout**: 30s wait timeout (default) vs 600s agent timeout
- **Gateway RPC endpoints**: `agent` and `agent.wait`
- **Serialized execution**: Per session + global queues

---

#### 10.7.6 Context Window Optimization Techniques

**What Counts in Context Window**:
- System prompt (tool descriptions, skills, bootstrap)
- Conversation history (user + assistant messages)
- Tool calls and tool results
- Attachments and transcripts
- Compaction summaries and pruning artifacts

**Optimization Hierarchy** (from highest to lowest impact):

1. **System Prompt Trimming**:
   - Limit bootstrap files to 20,000 chars (default)
   - Keep skill descriptions concise
   - Use tool allow/deny lists instead of verbose descriptions

2. **Session Pruning** (cache-ttl mode):
   - Prune old tool results after TTL expires
   - Protect last N assistant messages (default: 3)
   - **Soft-trim**: head + tail (30%/70% ratio)
   - **Hard-clear**: replace with placeholder (50% ratio)
   - **Skip image blocks** (never trimmed)

3. **Output Stream Limiting**:
   - **Background tasks**: Last 100 lines (individual)
   - **List view**: Last 10 lines (last N for overview)
   - **Circular buffer**: 500 lines for infinite streams

4. **Compaction**:
   - Summarize old messages
   - Archive to disk
   - Inject summary instead of full history

---

#### 10.7.7 Recommended Architecture for "Watch & Auto-Fix" Use Case

**User Scenario**: "Watch logs on server, detect errors, attempt fixes in loop until no more errors"

**Recommended Strategy 1: Pump-and-Dump with Periodic Summaries**

**Token Cost**: ~500 tokens per summary (every 10 minutes)
**User Benefits**:
- Minimal token usage
- User control (can stop/pause anytime)
- Persistent monitoring (background process continues)

**Flow**:
```
1. Create background process: tail -f -n 100 /var/log/app.log
2. Agent reads output in chunks (100 line buffer)
3. Pattern matching in-memory:
   - Detect errors
   - Identify recurring patterns
   - Maintain compact state
4. Every N iterations OR on error detection:
   - Generate compact summary of findings
   - Send to user
   - Wait for "continue" signal
   - Clear memory except critical state
   - Resume monitoring
```

**Configuration**:
```json
{
  "backgroundTask": {
    "name": "log-monitor",
    "command": "tail -f -n 100 /var/log/app.log",
    "tags": ["monitoring", "logs", "auto-fix"],
    "global": true,
    "mode": "pump-dump",
    "outputConfig": {
      "bufferLines": 100,
      "errorPatterns": ["ERROR", "FATAL", "Exception"],
      "summaryInterval": "10m"
    }
  },
  "contextOptimization": {
    "mode": "cache-ttl",
    "ttl": "5m",
    "keepLastAssistants": 3
    "softTrimRatio": 0.3,
    "hardClearRatio": 0.5
  }
}
```

**Recommended Strategy 2: Streaming with Coalescing** (Real-time user experience)

**Token Cost**: ~2,000-5,000 tokens per hour (varies by output)
**User Benefits**:
- Real-time feedback
- Progressive output (feels natural)
- Immediate error alerts
- User can interrupt/pause anytime

**Flow**:
```
1. Create background process: tail -f -n 100 /var/log/app.log
2. Agent reads output continuously
3. Streaming output to user:
   - Block streaming (coarse chunks)
   - Coalescing (merge chunks before sending)
   - Min 500 chars, max 3000 chars, idle 200ms
4. When error detected:
   - Send immediate alert chunk (breaks coalescing)
   - Attempt auto-fix
   - Stream fix attempt output
5. Continue monitoring until user stops
```

**Configuration**:
```json
{
  "backgroundTask": {
    "name": "log-monitor",
    "command": "tail -f -n 100 /var/log/app.log",
    "tags": ["monitoring", "logs", "streaming"],
    "global": true,
    "mode": "streaming"
  },
  "streaming": {
    "blockStreamingDefault": "on",
    "blockStreamingBreak": "text_end",
    "blockStreamingChunk": {
      "minChars": 500,
      "maxChars": 3000,
      "breakPreference": "paragraph"
    },
    "blockStreamingCoalesce": {
      "minChars": 1500,
      "maxChars": 3000,
      "idleMs": 200
    }
  },
  "contextOptimization": {
    "mode": "cache-ttl",
    "ttl": "5m",
    "keepLastAssistants": 3
    "softTrimRatio": 0.3,
    "hardClearRatio": 0.5
  }
}
```

**Recommended Strategy 3: Hybrid - Adaptive Based on State** (Best of both worlds)

**Token Cost**: Variable (pump-dump when idle, streaming when active)
**User Benefits**:
- Optimized for both scenarios
- Automatic switching based on urgency
- Best user experience + efficiency balance

**Flow**:
```
State Machine:
┌─────────────┐
│ IDLE        │ → Normal: pump-and-dump with 10m summaries
└──────┬──────┘
       ▼
┌─────────────┐
│ MONITORING  │ → Active: streaming with coalescing (min: 500, idle: 200ms)
└──────┬──────┘
       ▼
┌─────────────┐
│ ACTION      │ → Critical: streaming immediate (no coalescing, sentence breaks)
└─────────────┘
       ▼
┌─────────────┐
│ ERROR       │ → Continue with same state as IDLE
└─────────────┘
```

**State Definitions**:
- **IDLE**: No errors detected in last N minutes
- **MONITORING**: Error patterns detected OR user requested watch mode
- **ACTION**: Auto-fix in progress OR critical error
- Use pump-and-dump (10m summaries) when idle
- Use streaming (coalesced) when monitoring
- Use streaming immediate when in action state

**Configuration**:
```json
{
  "backgroundTask": {
    "name": "log-monitor",
    "command": "tail -f -n 100 /var/log/app.log",
    "tags": ["monitoring", "logs", "auto-fix", "adaptive"],
    "global": true,
    "mode": "hybrid",
    "outputConfig": {
      "bufferLines": 100,
      "circularBufferSize": 500,
      "errorPatterns": ["ERROR", "FATAL", "Exception"],
      "summaryInterval": "10m"
    }
  },
  "stateMachine": {
    "idleThresholdMs": 5 * 60 * 1000,  // 5 minutes
    "monitoringThresholdMs": 60 * 1000,  // 1 minute since error
    "actionThresholdMs": 30 * 1000,  // 30 seconds since last error
  }
}
```

---

#### 10.7.8 When User is Working in Other Tools (Your Scenario)

**Scenario**: User is actively working in OpenCode/Cursor while Pryx agent runs background task on their device.

**Recommended Strategy**: **Pump-and-Dump with Periodic Summaries**

**Rationale**:
1. User is actively working - they don't need real-time streaming
2. Pryx agent runs autonomously - token cost should be minimal
3. User can always request latest summary by typing `continue` or `status`
4. Compact summaries every 10 minutes = excellent awareness vs cost trade-off

**UX Pattern**:
```
Initial Setup:
User: "Watch logs on server, detect errors, attempt fixes in loop until no more errors"

Pryx Agent:
"✅ Started background log monitor (pid: 12345)
✅ Monitoring /var/log/app.log
✅ Will send summaries every 10 minutes
✅ Type 'continue' for next summary, 'status' for current state, or 'stop' to end"

[10 minutes later...]
Pryx Agent:
"📊 Monitoring Summary (10:00 - 10:10)
• Lines processed: 2,543
• Errors detected: 3
  - ERROR: Database connection failed (10:02:15)
  - FATAL: Out of memory (10:05:22)
  - ERROR: Payment gateway timeout (10:07:01)
• Auto-fix attempts: 2
  - [10:02:18] Restarted database service (SUCCESS)
  - [10:05:30] Increased memory limit (PENDING)
  - [10:07:05] Payment gateway health check (FAILED)

Status: 2 errors remaining
Next summary: 10:20

Type 'continue' for next summary:
OR
Type 'fix [error-id]' to attempt manual fix for specific error:
OR
Type 'status' to see full recent output:
OR
Type 'stop' to end monitoring:"
```

**User in OpenCode** - sees notification in Slack/Discord:
```
[Channel: #dev-ops]
[Bot: @Pryx]
[Notification] Found 3 errors in server logs. See summary above.
```

User can then:
- Check in Pryx UI for full logs
- Take action on specific error
- Continue monitoring or stop task

---

#### 10.7.9 Implementation Recommendations

**Phase 1: Background Process Manager** (Week 1-2)
- Implement `PryxBackgroundProcessManager` class
- Support in-memory task tracking with unique IDs
- Implement output buffering with circular buffer (500 lines)
- Add tag-based filtering
- Add global vs session-specific processes

**Phase 2: Token Optimization Layer** (Week 2-3)
- Implement session pruning with cache-ttl mode
- Add output stream limiting (100 lines individual, 10 lines list)
- Implement soft-trim and hard-clear strategies
- Add context window tracking

**Phase 3: Adaptive Streaming Controller** (Week 3-4)
- Implement state machine (idle → monitoring → action → completed)
- Add streaming with coalescing support
- Implement immediate streaming for critical actions
- Add error pattern detection and consecutive error tracking

**Phase 4: Heartbeat Integration** (Week 4)
- Implement heartbeat configuration system
- Add support for custom heartbeat intervals
- Implement HEARTBEAT.md workspace file reading
- Add ack suppression to avoid double delivery

**Phase 5: Configuration UI** (Week 4-5)
- Add UI for selecting autocompletion mode (pump-dump/streaming/hybrid)
- Add UI for configuring output limits (buffer size, coalescing settings)
- Add UI for error pattern configuration
- Add UI for context optimization settings (TTL, trim ratios)

**Phase 6: Documentation & Examples** (Week 5)
- Document all three strategies with pros/cons
- Provide configuration examples for each use case
- Create templates for common background tasks (log monitor, file watcher, cron job)

---

## 11) Data Model

**Task Queue Architecture**:
```go
type TaskScheduler struct {
    db        *sql.DB
    executor  *TaskExecutor
    mesh      *PryxMeshClient
}

func (ts *TaskScheduler) Run() {
    ticker := time.NewTicker(1 * time.Minute)

    for {
        select {
        case <-ticker.C:
            // Find tasks due to run
            tasks := ts.getDueTasks()

            for _, task := range tasks {
                // Check target device
                if task.TargetDeviceID != nil {
                    // Remote execution via Mesh
                    go ts.executeRemote(task)
                } else {
                    // Local execution
                    go ts.executeLocal(task)
                }
            }

        case err := <-ts.executor.ErrorChan():
            // Handle task failures
            ts.handleTaskError(err)
        }
    }
}

func (ts *TaskScheduler) getDueTasks() []Task {
    var tasks []Task
    ts.db.Select(&tasks,
        "SELECT * FROM scheduled_tasks WHERE status = 'active' AND next_run <= ?",
        time.Now())
    return tasks
}

func (ts *TaskScheduler) executeRemote(task Task) error {
    // Send task to target device via Mesh Coordinator
    meshReq := MeshTaskRequest{
        TaskID:    task.ID,
        Task:      task,
        Timestamp: time.Now(),
    }

    response, err := ts.mesh.SendRequest(task.TargetDeviceID, meshReq)
    if err != nil {
        // Mark task as failed, schedule retry
        ts.handleTaskError(task, err)
        return err
    }

    // Update task status in local DB
    ts.updateTaskRun(task.ID, response)
    return nil
}

func (ts *TaskScheduler) handleTaskError(task Task, err error) {
    retryPolicy := task.RetryPolicy

    if task.Attempts < retryPolicy.MaxRetries {
        // Schedule retry with exponential backoff
        backoff := time.Duration(math.Pow(2, float64(task.Attempts))) * time.Minute
        nextRun := time.Now().Add(backoff)

        task.NextRun = &nextRun
        task.Attempts++

        ts.db.Update(task)
    } else {
        // Max retries exceeded, mark as failed
        task.Status = "failed"
        ts.db.Update(task)

        // Send notification
        ts.notifyUser(task, fmt.Sprintf("Task failed after %d attempts", retryPolicy.MaxRetries))
    }
}
```

**Persistence Schema** (already defined in FR11):
```sql
-- Task definition
CREATE TABLE scheduled_tasks (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    trigger_type TEXT NOT NULL,
    trigger_config JSONB NOT NULL,
    action_config JSONB NOT NULL,
    target_device_id UUID,
    retry_policy JSONB,
    notification_config JSONB,
    status TEXT NOT NULL,
    next_run TIMESTAMP,
    last_run TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Task execution history
CREATE TABLE task_runs (
    id UUID PRIMARY KEY,
    task_id UUID REFERENCES scheduled_tasks(id) ON DELETE CASCADE,
    run_number INTEGER,
    attempt_number INTEGER,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    status TEXT NOT NULL,
    result JSONB,
    error TEXT,
    tokens_used INTEGER,
    cost_usd DECIMAL(10,4),
    execution_device_id UUID
);
```

**Recovery After Crash**:
```go
func (ts *TaskScheduler) Recover() {
    // Find tasks that were running when crashed
    var incompleteRuns []TaskRun
    ts.db.Select(&incompleteRuns,
        "SELECT * FROM task_runs WHERE status = 'running' AND completed_at IS NULL")

    for _, run := range incompleteRuns {
        // Mark as failed, notify user
        run.Status = "failed"
        run.Error = "Task interrupted by application crash"
        ts.db.Update(run)

        ts.notifyUser(run.TaskID, "Task was interrupted, check logs")
    }

    // Reschedule due tasks
    now := time.Now()
    ts.db.Exec("UPDATE scheduled_tasks SET next_run = ? WHERE next_run < ? AND status = 'active'",
        now, now)
}
```

---

## 11) Data Model

### 11.1 Core Entities

```
┌─────────────────────────────────────────────────────────────────┐
│                         SQLite Schema                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  sessions                     messages                           │
│  ┌────────────────────┐      ┌────────────────────────────┐     │
│  │ id (UUID, PK)      │      │ id (UUID, PK)              │     │
│  │ workspace_id       │──┐   │ session_id (FK)            │     │
│  │ created_at         │  │   │ role (user|assistant|tool) │     │
│  │ updated_at         │  │   │ content (TEXT)             │     │
│  │ title              │  │   │ tool_call_id               │     │
│  │ status             │  │   │ created_at                 │     │
│  │ model              │  │   │ tokens_in                  │     │
│  │ total_tokens       │  │   │ tokens_out                 │     │
│  │ total_cost_usd     │  │   └────────────────────────────┘     │
│  └────────────────────┘  │                                       │
│           │              │   tool_calls                          │
│           │              │   ┌────────────────────────────┐     │
│           │              │   │ id (UUID, PK)              │     │
│           │              │   │ session_id (FK)            │     │
│           │              │   │ tool_name                  │     │
│           │              │   │ arguments (JSON)           │     │
│           │              │   │ result (JSON)              │     │
│           │              │   │ status (pending|approved|  │     │
│           │              │   │         denied|completed)  │     │
│           │              │   │ approval_scope             │     │
│           │              │   │ started_at                 │     │
│           │              │   │ completed_at               │     │
│           │              │   │ duration_ms                │     │
│           │              │   └────────────────────────────┘     │
│           │              │                                       │
│  workspaces              │   integrations                        │
│  ┌────────────────────┐  │   ┌────────────────────────────┐     │
│  │ id (UUID, PK)      │◄─┘   │ id (UUID, PK)              │     │
│  │ name               │      │ type (telegram|discord|    │     │
│  │ path               │      │       slack|webhook)       │     │
│  │ created_at         │      │ name                       │     │
│  │ policies (JSON)    │      │ config (JSON, encrypted)   │     │
│  └────────────────────┘      │ status                     │     │
│                              │ workspace_id (FK)          │     │
│  audit_log                   │ created_at                 │     │
│  ┌────────────────────┐      │ last_connected_at          │     │
│  │ id (UUID, PK)      │      └────────────────────────────┘     │
│  │ timestamp          │                                         │
│  │ action             │      policies                           │
│  │ resource_type      │      ┌────────────────────────────┐     │
│  │ resource_id        │      │ id (UUID, PK)              │     │
│  │ actor              │      │ workspace_id (FK)          │     │
│  │ details (JSON)     │      │ tool_pattern               │     │
│  │ outcome            │      │ action (allow|deny|ask)    │     │
│  └────────────────────┘      │ scope                      │     │
│                              │ duration                   │     │
│                              │ expires_at                 │     │
│                              └────────────────────────────┘     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 11.2 Configuration Files

```
~/.pryx/
├── config.json           # User preferences
├── metadata.json         # Runtime metadata (ports, PIDs)
├── pryx.db              # SQLite database
├── keychain.json        # Reference to keychain entries (not secrets)
├── skills/              # Managed skills (global, user-installed)
│   └── <skill-name>/
│       ├── SKILL.md     # Skill definition (YAML frontmatter + instructions)
│       ├── scripts/     # Optional executable code
│       └── references/  # Optional docs loaded on-demand
├── mcp/
│   └── servers.json     # MCP server configurations
└── exports/             # Exported sessions/reports
```

**Workspace Skills** (per-workspace, higher precedence):
```
<workspace>/
└── .pryx/
    └── skills/
        └── <skill-name>/SKILL.md
```

**config.json Schema**:
```json
{
  "version": "1.0",
  "telemetry": {
    "enabled": true,
    "endpoint": "https://telemetry.pryx.dev",
    "sampling_ratio": 1.0,
    "redact_pii": true
  },
  "defaults": {
    "model": "anthropic/claude-sonnet-4-20250514",
    "workspace": "~/projects"
  },
  "ui": {
    "theme": "system",
    "port": null
  },
  "updates": {
    "channel": "stable",
    "auto_check": true
  },
  "skills": {
    "enabled": true,
    "extraDirs": [],
    "entries": {}
  },
  "mcp": {
    "enabled": true,
    "servers": {}
  }
}
```

**Telemetry Data Classification** (applies when enabled):

| Data Type | Sent | Policy |
|-----------|------|--------|
| Anonymized usage metrics (errors, latency, feature usage) | When opted-in | Aggregated, no session context |
| Prompts/conversations | NEVER | Sovereignty principle |
| PII (emails, names, API keys) | NEVER | Redacted before any processing |
| Tool call patterns | When opted-in | Tool names only, no arguments |

**Telemetry Tiers** (for Cloudflare Workers logging):

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      Telemetry Tiers                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Tier 1: Always (Core Metrics) - <1KB per session                       │
│  ├── session_started {timestamp, surface, model}                         │
│  ├── session_ended {duration_ms, message_count, tool_count}             │
│  ├── error_occurred {error_code, component, trace_id}                    │
│  └── No content, no arguments, no user data                              │
│                                                                          │
│  Tier 2: Detailed (Opt-in Analytics) - ~5KB per session                 │
│  ├── tool_called {tool_name, duration_ms, success, approval_type}       │
│  ├── model_request {model, tokens_in, tokens_out, latency_ms}           │
│  ├── channel_event {channel_type, event_type, latency_ms}               │
│  └── Still no content, no arguments                                      │
│                                                                          │
│  Tier 3: Debug (Explicit Enable) - Variable size                        │
│  ├── Full tool arguments (PII-redacted)                                  │
│  ├── LLM request/response metadata                                       │
│  └── Only during active debugging, auto-disables after 1 hour           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**Telemetry Event Schema** (for Cloudflare Workers):
```typescript
interface TelemetryEvent {
  // Identity (anonymized)
  device_id: string;       // SHA256(machine_id + install_date)
  session_id: string;      // Random, not correlatable to user
  
  // Event
  event_type: string;
  timestamp: number;
  
  // Context (no PII)
  pryx_version: string;
  os: "macos" | "linux" | "windows";
  surface: "cli" | "tui" | "web" | "telegram" | "whatsapp" | "discord";
  
  // Payload (varies by event_type)
  payload: Record<string, unknown>;
}
```

**What NEVER to log** (even in Tier 3):
- Prompts/messages (sovereignty principle)
- Tool arguments that could contain secrets/paths
- API keys (obviously)
- File contents
- User identifiers (email, name, IP)

---

## 12) Auto-Update Mechanism

### 12.1 Build Channel Architecture

**Different Build Channels**:

| Channel | Use Case | Update Mechanism |
|---------|-----------|------------------|
| **Main/Stable** (Production) | Production users on `main` branch | Auto-update enabled by default |
| **Beta/Development** | Beta testers on development branch | Auto-update for beta builds only |
| **Alpha/Canary** | Early access users | Manual updates only, notifications available |

### 12.2 User Experience - Production Build (Main Pool)

**Update Flow**:
1. **Version Check (Startup)**: Pryx checks for updates on startup, compares current version with latest
2. **Update Available Notification**: Toast notification appears with actions: Update Now, Remind Me Later, Skip
3. **Background Download**: Update downloaded in background while user continues using Pryx
4. **Update Installation**: Toast shows "Update ready! Restart to apply", user clicks "Restart Now"
5. **Graceful Restart**: Pryx shuts down gracefully, applies update, restarts automatically

**Toast Notification Types**:

| Toast Type | When Shown | Actions | Auto-Dismiss |
|-----------|-------------|---------|--------------|
| Update Available | New version detected | Update Now, Remind Me Later, Skip | 30s |
| Download Progress | Update downloading | Show Progress, Cancel | Never |
| Update Ready | Download complete | Restart Now | 60s |
| Update Installed | Restart complete | What's New, Changelog | Never (shows modal) |

**What's New Modal**: Displays after successful update with new features, improvements, and fixes.

### 12.3 User Experience - Beta Build Channel

**Update Flow** (similar to production, but with beta-specific features):
1. **Beta Version Check**: Same version check mechanism
2. **Beta Update Available**: Toast shows 🧪 icon, "Beta build - may contain bugs" warning
3. **User Actions**: Update to beta, Report bug, Stable only mode toggle

**Beta-Specific UX**:
- 🧪 Icon distinguishes beta from stable
- Warning message on beta updates
- "Stable only" mode to disable beta updates
- Direct link to bug reporting for beta builds

### 12.4 User Experience - Switching Build Channels

**User Flow**:
1. User clicks "Switch to Beta Channel" (or similar)
2. Confirmation dialog shows warning and available beta versions
3. Pryx restarts with new channel config
4. Future update checks target the selected channel

### 12.5 Background Update Mechanism

**Key Requirements**:

| Requirement | Implementation | Notes |
|------------|----------------|--------|
| Silent downloads | Downloads in background, user can continue using Pryx |  |
| Progress indicators | Show download progress in UI toast |  |
| Graceful shutdown | App restarts cleanly after update |  |
| Rollback capability | If update fails, rollback to previous version |  |
| Update history | Track update history for debugging |  |
| Update scheduling | Respect "remind me later" user preference |  |

**Implementation Details** (TypeScript interface):
```typescript
interface UpdateConfig {
  currentVersion: string;
  latestVersion: string;
  buildChannel: 'main' | 'beta' | 'alpha';
  autoUpdateEnabled: boolean;
  lastCheckAt: Date;
  downloadProgress?: {
    downloadedBytes: number;
    totalBytes: number;
    percentage: number;
  };
}
```

### 12.6 Configuration API

**User Config Structure**:
```json
{
  "update": {
    "enabled": true,
    "buildChannel": "main",
    "autoCheckOnStartup": true,
    "checkIntervalHours": 24,
    "downloadInBackground": true,
    "allowBetaUpdates": false,
    "remindMeLater": "1h"
  }
}
```

**Configuration UI Settings**:
- Automatic Updates: Check on startup, download in background, ask before downloading
- Build Channel selector: Main (Stable), Beta (Development), Alpha (Early Access)
- Update Frequency: Daily (recommended), Weekly, Manual only
- Notification preferences: New stable releases, Beta releases, Security updates only

### 12.7 Success Metrics

| Metric | v1 Target | v1.1 Target | v2 Target |
|--------|-----------|-------------|----------|
| Update success rate | >98% | >99% | >99.5% |
| Rollback success rate | N/A | >95% | >98% |
| Update adoption rate | 60% of users | 70% of users | 80% of users |
| Update awareness | Users know update available in <1 day | Users aware within 1 day | Users aware within 1 day |

---

## 13) API Specifications

### 12.1 REST API (pryx-core)

Base URL: `http://localhost:{port}/api/v1`

#### Sessions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/sessions` | List all sessions |
| POST | `/sessions` | Create new session |
| GET | `/sessions/{id}` | Get session details |
| DELETE | `/sessions/{id}` | Delete session |
| POST | `/sessions/{id}/messages` | Send message to session |

**Example: Create Session**
```http
POST /api/v1/sessions
Content-Type: application/json

{
  "workspace_id": "uuid",
  "model": "anthropic/claude-sonnet-4-20250514",
  "system_prompt": "You are a helpful assistant."
}

Response: 201 Created
{
  "id": "session-uuid",
  "status": "active",
  "created_at": "2026-01-27T12:00:00Z"
}
```

#### Health & Diagnostics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/health/ready` | Readiness probe |
| GET | `/diagnostics` | Detailed diagnostics |

**Health Response**:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "components": {
    "database": "ok",
    "keychain": "ok",
    "integrations": {
      "telegram": "connected",
      "webhook": "connected"
    }
  }
}
```

#### Integrations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/integrations` | List integrations |
| POST | `/integrations` | Add integration |
| DELETE | `/integrations/{id}` | Remove integration |
| POST | `/integrations/{id}/test` | Test connection |

### 12.2 WebSocket API

Connect: `ws://localhost:{port}/api/v1/ws`

**Event Types**:

| Event | Direction | Payload |
|-------|-----------|---------|
| `session.message` | Server→Client | New message in session |
| `tool.request` | Server→Client | Tool approval requested |
| `tool.response` | Client→Server | Approval decision |
| `tool.result` | Server→Client | Tool execution result |
| `trace.event` | Server→Client | Trace/telemetry event |
| `error` | Server→Client | Error notification |

**Example: Tool Approval Flow**
```json
// Server sends approval request
{
  "type": "tool.request",
  "id": "req-123",
  "session_id": "session-uuid",
  "tool": "shell.exec",
  "arguments": {
    "command": "rm -rf /tmp/test"
  },
  "risk_level": "high"
}

// Client sends approval decision
{
  "type": "tool.response",
  "request_id": "req-123",
  "decision": "approve",
  "scope": "once"
}
```

---

## 14) Error Handling

### 14.1 Error Categories

| Category | HTTP Code | Retry Strategy |
|----------|-----------|----------------|
| Validation | 400 | No retry |
| Authentication | 401 | Refresh token, retry once |
| Authorization | 403 | No retry |
| Not Found | 404 | No retry |
| Rate Limited | 429 | Exponential backoff (1s, 2s, 4s, 8s, max 60s) |
| Server Error | 5xx | Exponential backoff, max 3 retries |
| Network Error | N/A | Exponential backoff, max 5 retries |

### 14.2 Error Response Format

```json
{
  "error": {
    "code": "TOOL_DENIED",
    "message": "Tool execution denied by policy",
    "details": {
      "tool": "shell.exec",
      "policy": "workspace-only",
      "attempted_path": "/etc/passwd"
    },
    "request_id": "req-123",
    "timestamp": "2026-01-27T12:00:00Z"
  }
```

### 14.3 Circuit Breaker

For external integrations (Telegram, Discord, etc.):

| State | Behavior |
|-------|----------|
| Closed | Normal operation |
| Open | Fail fast (after 5 consecutive failures) |
| Half-Open | Allow 1 request every 30s to test recovery |

### 14.4 Error Surface Architecture (Rust + Go)

**Why two languages?** Error handling needs to work at every layer:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        User's Machine                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                      Rust Host (pryx)                         │   │
│  │                                                               │   │
│  │  Responsibilities:                                            │   │
│  │  - Catch sidecar crashes → display error in TUI/native dialog │   │
│  │  - Handle deep link errors → show in UI                       │   │
│  │  - Manage update failures → rollback + notify                 │   │
│  │  - Forward errors to telemetry (if opted-in)                  │   │
│  │                                                               │   │
│  │  Error Display:                                               │   │
│  │  - TUI: inline error panel with trace ID                      │   │
│  │  - Desktop: native toast/dialog                               │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ IPC (WebSocket/HTTP)                  │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Go Sidecar (pryx-core)                     │   │
│  │                                                               │   │
│  │  Responsibilities:                                            │   │
│  │  - LLM errors → structured error in session timeline          │   │
│  │  - Tool execution errors → approval UI with error details     │   │
│  │  - Channel errors → retry + circuit breaker                   │   │
│  │  - MCP server errors → reconnect + notify host                │   │
│  │                                                               │   │
│  │  Error Display:                                               │   │
│  │  - Web UI: error toast + timeline entry                       │   │
│  │  - WebSocket: error event to all connected clients            │   │
│  │  - API: structured JSON error response                        │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ Channel-specific                      │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Channel Adapters                           │   │
│  │                                                               │   │
│  │  Telegram: reply with error message to user                   │   │
│  │  Discord: error embed in channel                              │   │
│  │  Webhook: 4xx/5xx response with error body                    │   │
│  │  Slack: ephemeral error message                               │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ OpenTelemetry (opt-in)
                              ▼
                     Telemetry Worker
                              │
                              │ PII redacted
                              ▼
                     Error Aggregation
                     (Internal Dashboard)
```

**Error Response includes trace_id for correlation**:
```json
{
  "error": {
    "code": "LLM_RATE_LIMITED",
    "message": "Claude API rate limit exceeded",
    "trace_id": "abc123-def456",
    "surface": "telegram",
    "retry_after_ms": 5000
  }
}
```

**User can reference trace_id when reporting issues** → correlates to internal dashboard.

---

## 15) Deployment Architecture

### 15.1 End-User Installation

```bash
# macOS/Linux
curl -fsSL https://get.pryx.dev | sh

# Windows (PowerShell)
irm https://get.pryx.dev/install.ps1 | iex
```

**Installation Directory Structure**:
```
# macOS/Linux
~/.local/bin/pryx          # Main binary
~/.local/bin/pryx-core     # Sidecar binary
~/.pryx/                   # Data directory

# Windows
%LOCALAPPDATA%\Pryx\bin\   # Binaries
%APPDATA%\Pryx\            # Data directory
```

### 14.2 Server Deployment (Docker)

```yaml
# docker-compose.yml
version: '3.8'
services:
  pryx:
    image: ghcr.io/irfndi/pryx:latest
    ports:
      - "3001:3001"
    volumes:
      - pryx_data:/data
      - /var/run/docker.sock:/var/run/docker.sock  # Optional: for Docker tools
    environment:
      - PRYX_PORT=3001
      - PRYX_HEADLESS=true
      - PRYX_DATA_DIR=/data
    restart: unless-stopped

volumes:
  pryx_data:
```

### 14.3 Cloudflare Workers

**Architecture Decision: Separate Workers with Service Bindings**

We keep auth, telemetry, and web as separate workers connected via service bindings for:
- **Zero-latency communication**: Service bindings run on same thread, no network hop
- **Independent deployments**: Security patches to auth don't require web redeploy
- **Better warm-start**: Smaller workers evicted less frequently (99.99% warm rate)
- **Security isolation**: Auth worker not exposed to public internet

```
┌─────────────────────────────────────────────────────────────┐
│                    Public Internet                          │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
               ┌───────────────────┐
               │    web-worker     │ ◄── Entry point (docs, changelog)
               │    (public)       │
               └────────┬──────────┘
                        │ Service Bindings (zero-cost RPC)
          ┌─────────────┴─────────────┐
          ▼                           ▼
┌───────────────────┐      ┌───────────────────┐
│   auth-worker     │      │ telemetry-worker  │
│   (internal)      │      │   (internal)      │
│   RFC 8628 OAuth  │      │   OTLP + redact   │
└───────────────────┘      └───────────────────┘
```

```
workers/
├── auth/
│   ├── src/index.ts       # OAuth Device Flow (WorkerEntrypoint)
│   └── wrangler.toml
├── telemetry/
│   ├── src/index.ts       # OTLP ingestion (WorkerEntrypoint)
│   └── wrangler.toml
└── web/
    ├── src/index.ts       # Docs/changelog + service binding orchestration
    └── wrangler.toml
```

**Service Binding Configuration** (web-worker's wrangler.toml):
```toml
[[services]]
binding = "AUTH"
service = "auth-worker"
entrypoint = "AuthEntrypoint"

[[services]]
binding = "TELEMETRY"
service = "telemetry-worker"
entrypoint = "TelemetryEntrypoint"
```

### 14.4 Internal Admin Dashboard (Superadmin)

**Separate from user-installed Pryx.** This is our internal platform for monitoring all users.

**Architecture**:
```
┌────────────────────────────────────────────────────────────────────┐
│                    Pryx Internal Admin Dashboard                    │
│                    (superadmin.pryx.dev)                           │
├────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────┐ │
│  │ User Management │  │ Error Dashboard │  │ Analytics Dashboard │ │
│  │                 │  │                 │  │                     │ │
│  │ - All users     │  │ - Aggregated    │  │ - Usage metrics     │ │
│  │ - Device mesh   │  │   errors        │  │ - Feature adoption  │ │
│  │ - Referrals     │  │ - AI analysis   │  │ - Retention         │ │
│  │ - Subscriptions │  │ - GitHub sync   │  │ - Costs             │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────────┘ │
│                                                                     │
└────────────────────────────────────────────────────────────────────┘
                                │
                                │ Cloudflare Workers
                                ▼
┌────────────────────────────────────────────────────────────────────┐
│  ┌─────────────┐  ┌─────────────────┐  ┌───────────────────────┐  │
│  │ admin-worker│  │ error-agg-worker│  │ analytics-worker      │  │
│  │             │  │                 │  │                       │  │
│  │ - Auth/RBAC │  │ - Sentry ingest │  │ - Usage aggregation   │  │
│  │ - User CRUD │  │ - PII redaction │  │ - Grafana/Loki export │  │
│  │ - Referrals │  │ - GitHub issues │  │ - Trend analysis      │  │
│  └─────────────┘  └─────────────────┘  └───────────────────────┘  │
│                           │                                        │
│                     ┌─────▼─────┐                                  │
│                     │ D1/KV/DO  │  Durable storage                 │
│                     └───────────┘                                  │
└────────────────────────────────────────────────────────────────────┘
```

**Components**:

| Component | Purpose | Technology |
|-----------|---------|------------|
| **User Registry** | All registered users, device mesh, referrals | Cloudflare D1 |
| **Error Aggregation** | Collect errors from all user devices | Sentry (self-hosted or cloud) |
| **GitHub Integration** | Auto-create issues, AI fix suggestions | Sentry Seer + GitHub API |
| **Analytics** | Usage metrics, feature adoption | Grafana + Loki (multi-tenant) |
| **Admin UI** | Internal dashboard (not user-facing) | React SPA on Cloudflare Pages |

**Error Flow (User → Admin)**:
```
User Device (CLI/Bot/TUI)
         │
         │ OpenTelemetry SDK (errors + traces)
         ▼
Telemetry Worker (PII redaction)
         │
         │ Forward (anonymized)
         ▼
Error Aggregation Worker
         │
    ┌────┴────┐
    ▼         ▼
 Sentry    Loki
    │         │
    │    Multi-tenant
    │    Dashboard
    ▼
AI Analysis (Seer)
    │
    ▼
GitHub Issue (auto-created with CODEOWNERS)
    │
    ▼
Slack/Discord Alert
```

**User-Side Error Visibility**:

| Surface | Error Display | Implementation |
|---------|--------------|----------------|
| **CLI/TUI** | Inline error message + trace ID | Rust Host displays pryx-core errors |
| **Web UI** | Error toast + timeline entry | React error boundary + WS events |
| **Telegram/Discord Bot** | Error reply to user | Channel adapter formats error |
| **Headless/API** | JSON error response | HTTP 4xx/5xx with structured body |

All surfaces include a **trace ID** that correlates to internal dashboards.

**Admin-Side Capabilities**:

| Feature | Description |
|---------|-------------|
| **Aggregated Error View** | All errors across all users, grouped/deduplicated |
| **User Context** | Which user/device/workspace triggered error (anonymized) |
| **AI Root Cause** | Sentry Seer analyzes stack trace + traces + logs |
| **Auto GitHub Issue** | Critical errors auto-create issues with CODEOWNERS routing |
| **Reproduction Hints** | Aggregated context helps reproduce issues |
| **Fix Suggestions** | AI generates draft PRs for common issues |
| **Trend Analysis** | Error rate over time, regression detection |

### 14.5 Website & Login (pryx.dev)

**Same domain serves docs, changelog, and user authentication**:

```
pryx.dev/              → Landing page
pryx.dev/docs/         → Documentation
pryx.dev/changelog/    → Release notes
pryx.dev/login/        → OAuth login (GitHub, Google)
pryx.dev/dashboard/    → User dashboard (devices, settings, referrals)
pryx.dev/mesh/         → Device mesh management
```

**User Dashboard (post-login)**:

| Feature | Description |
|---------|-------------|
| **Devices** | List of paired devices (laptop, server, etc.) |
| **Mesh Status** | Connection status of device mesh |
| **Referrals** | Invite link, referral stats, rewards |
| **Settings** | Telemetry opt-out, notification preferences |
| **Billing** (v2) | Subscription management (if applicable) |

**Login persists to local Pryx via Device Flow**:
1. User clicks "Connect Device" in pryx.dev/dashboard
2. Shows QR code / device code (RFC 8628)
3. User scans in Pryx CLI/TUI
4. Device authenticated and linked to account

---

## 16) Success Metrics

### 16.1 Quantitative Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Install-to-first-chat | <2 minutes | Timer from install start to first response |
| Install-to-first-integration | <5 minutes | Timer from install to connected channel |
| Zombie processes | 0 | Count after app exit |
| Crash restart success | ≥99% | Successful restarts / total crashes |
| Trace completeness | ≥90% | Sessions with full trace timeline |
| UI startup time | <3 seconds | Time to first paint |
| Memory footprint (idle) | <200MB | Host + sidecar combined |
| Concurrent channel handling | ≥50 events | Without degradation |
| Update success rate | >98% (v1), >99% (v1.1), >99.5% (v2) | Successful updates / total update attempts |
| Rollback success rate | >95% (v1.1), >98% (v2) | Successful rollbacks / total rollback attempts |
| Update adoption rate | 60% (v1), 70% (v1.1), 80% (v2) | Users who update within 1 week |
| Update awareness | Users aware within 1 day | Time from release to user awareness |

### 16.2 Qualitative Metrics

| Metric | Target | Method |
|--------|--------|--------|
| User satisfaction | NPS ≥50 | Post-install survey |
| Onboarding completion | ≥90% | Funnel analytics |
| Feature discoverability | ≥80% | Can find setting without help |
| Trust in security | High | User interviews |

---

## 16) Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| WhatsApp integration complexity | High | High | Prefer Cloud API; isolate Baileys in subprocess |
| OAuth edge complexity | Medium | Medium | Stateless workers, strict token TTL, key rotation |
| Observability privacy concerns | High | Medium | Default sampling off, redact all PII |
| Sidecar crash instability | High | Low | Crash isolation, backoff restart, session recovery |
| Auto-update breaks user | High | Low | Staged rollouts, instant rollback |
| Port conflicts | Medium | Medium | Random port default, clear override options |

---

## 17) Milestones

### M1: Foundation (Weeks 1-3)
- [ ] One-liner installer (macOS/Linux)
- [ ] Go sidecar with HTTP/WS API
- [ ] Basic MCP tool execution
- [ ] TUI with chat and approvals

### M2: Integration (Weeks 4-6)
- [ ] Telegram integration (cloud-hosted webhook mode)
  - [ ] Bot token verification (getMe) + webhook registration (setWebhook + secret token)
  - [ ] Chat linking (/start + link code) + allowlist
  - [ ] Message ingest pipeline (dedupe update_id, retries, timeouts)
  - [ ] Per-user model key isolation + rate limits
  - [ ] User controls: pause/resume, rotate secret, revoke token
- [ ] Telegram integration (device-hosted polling mode, optional)
  - [ ] Local channels.json wiring + auto-connect in pryx-core
  - [ ] Single-active-host enforcement (Mesh coordinator integration)
- [ ] Webhook channel adapter
- [ ] Web UI dashboard
- [ ] Onboarding wizard
- [ ] Skills management (managed + workspace)
- [ ] Native MCP client (stdio + HTTP)

### M3: Control Plane (Weeks 7-9)
- [ ] Cloudflare auth worker (with service bindings)
- [ ] Cloudflare telemetry worker (opt-in, PII redaction)
- [ ] Device flow OAuth
- [ ] Trace visualization
- [ ] Skills CLI (`pryx skills list|info|check`)
- [ ] MCP CLI (`pryx mcp add|list|test`)

### M4: Polish (Weeks 10-12)
- [ ] Tauri desktop wrapper
- [ ] Native permission dialogs
- [ ] Auto-updates with rollback
- [ ] Documentation and guides

### M5: v1 (Post-MVP)
- [ ] Multi-device Pryx Mesh
- [ ] AI Gateway mode
- [ ] Additional channels (Discord, Slack)
- [ ] Advanced policy engine

---

## 18) Glossary

| Term | Definition |
|------|------------|
| **Host** | The Rust/Tauri application managing lifecycle and native features |
| **Sidecar** | The Go runtime (pryx-core) handling LLM and tool execution |
| **MCP** | Model Context Protocol - standard interface for tool discovery and execution |
| **MCP Server** | External service providing tools via MCP (stdio or HTTP transport) |
| **Skill** | Agent capability defined in SKILL.md with instructions, requirements, and installers |
| **Policy Engine** | Component evaluating tool calls against user-defined rules |
| **Channel** | An integration adapter (Telegram, Discord, webhook, etc.) |
| **Workspace** | A scoped directory for agent file operations |
| **Device Flow** | OAuth 2.0 Device Authorization Grant (RFC 8628) |
| **Pryx Mesh** | Multi-device connectivity and coordination (v1) |
| **Service Binding** | Cloudflare Workers zero-latency inter-worker communication |
| **Node** | A single instance of Pryx running on a device (Laptop, Server, Phone) |
| **Coordinator** | Cloudflare Durable Object managing the Pryx Mesh state and routing |
| **Session Bus** | The event-driven pub/sub layer inside pryx-core |
| **Node** | A single instance of Pryx running on a device (Laptop, Server, Phone) |
| **Coordinator** | Cloudflare Durable Object managing the Pryx Mesh state and routing |
| **Session Bus** | The event-driven pub/sub layer inside pryx-core |

---

## 19) Appendix

### A. Referenced Standards

- [RFC 8628: OAuth 2.0 Device Authorization Grant](https://tools.ietf.org/html/rfc8628)
- [OpenTelemetry Protocol (OTLP)](https://opentelemetry.io/docs/specs/otlp/)
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)

### 8.3 Session Bus Architecture (Multi-Surface Sync)

### 8.4 Constraint Management & Multi-Device Orchestration

**Why This Matters**: Users will ask:
- "How do I handle model limits when agent wants 300k tokens?"
- "Can agent search files on laptop and download to my phone?"
- "How do I run command on device X from device Y?"
- "Can agent send files via WhatsApp to other device?"
- "Which model should I use? There are 600+ options!"
- "How do I manage API keys for 10+ providers?"

Pryx needs a robust system for handling constraints across providers, devices, and tasks.

#### 8.4.1 Unified Model Constraint Management

**Challenge**: 600+ models across 50+ providers (OpenRouter, Together, OpenAI, Anthropic, etc.) with varying constraints.

**Solution**: Dynamic constraint catalog sourced from [models.dev](https://models.dev) + runtime enforcement.

> **External Dependency**: Pryx integrates with [models.dev](https://github.com/anomalyco/models.dev) — an open-source database of AI model specifications.
> - **API Endpoint**: `https://models.dev/api.json`
> - **Refresh Strategy**: Fetch on startup, cache for 24 hours
> - **Fallback**: Use cached version if API unavailable
> - **Auto-Update**: New providers/models appear automatically without Pryx updates

**Constraint Catalog Schema** (mapped from models.dev API):
```typescript
interface ModelConstraints {
  // Core identity
  id: string;                    // "anthropic/claude-sonnet-4-20250514"
  canonical_slug: string;          // Permanent slug, never changes
  name: string;                  // Human-readable: "Claude Sonnet 4"
  
  // Context & Tokenization
  context_length: number;          // Max tokens in context (e.g., 200K)
  tokenizer: string;               // "cl100k", "tiktoken", "llama3"
  
  // Capabilities
  architecture: {
    input_modalities: string[];   // ["text", "image", "file"]
    output_modalities: string[];  // ["text", "json"]
    supports_tools: boolean;
    supports_vision: boolean;
    supports_reasoning: boolean;
    supports_streaming: boolean;
    max_completion_tokens: number;
  };
  
  // Rate limiting
  per_request_limits: {
    max_tokens_per_request: number | null;
    max_tools_per_request: number | null;
    max_parallel_tool_calls: number | null;
    max_images_per_request: number | null;
  };
  
  // Pricing (cost-aware routing)
  pricing: {
    prompt_per_million: number;    // Cost for input tokens
    completion_per_million: number; // Cost for output tokens
    request_fixed: number | null;      // Fixed cost per API call
    supports_caching: boolean;        // Can cache input tokens?
  };
  
  // Provider-specific overrides
  provider_overrides: {
    [provider_id: string]: {
      context_length?: number;          // Provider-specific limit
      max_completion_tokens?: number;
      rate_limit_rpm?: number;        // Requests per minute
      rate_limit_rph?: number;        // Requests per hour
    }
  };
}

interface ProviderConfig {
  id: string;                     // "anthropic", "openai", "together"
  name: string;
  base_url: string;
  supports_byok: boolean;           // User can bring own key
  supports_bundled: boolean;       // Pryx can provide access (paid)
  default_model: string | null;
  constraint_sources: string[];       // ["catalog", "api_headers", "user_override"]
}
```

**Constraint Detection Strategies**:

| Constraint Type | Detection Method | Enforcement Point | Example |
|---------------|------------------|-------------------|----------|
| **Context window** | Model catalog + API response headers | Pre-request validation | Prompt >200K → error before API call |
| **Token limit per request** | Model catalog (`per_request_limits`) | LLM request builder | Cap `max_tokens` in API request |
| **Rate limiting (RPM/RPH)** | API 429/429 error | HTTP client middleware | Exponential backoff + queue |
| **Tool calling limits** | Model catalog (`supports_tools`) + API error | Tool executor | Cap parallel calls, handle errors |
| **Image/video constraints** | Model catalog (`architecture.input_modalities`) | Content validator | Reject unsupported input types |
| **Streaming support** | Model catalog (`supports_streaming`) | Response handler | Use streaming vs polling |
| **Reasoning token budget** | Model catalog + API response (`usage.reasoning_tokens`) | LLM orchestration | Track and budget thinking tokens |
| **Caching support** | Model catalog (`pricing.supports_caching`) | Request builder | Include `cache_control` for savings |

**Runtime Enforcement**:
```go
type ConstraintEnforcer struct {
    catalog      *ModelCatalog
    rateLimiter  *RateLimiter
    budget       *BudgetTracker
}

func (ce *ConstraintEnforcer) PreValidateRequest(req LLMRequest) error {
    model, err := ce.catalog.GetModel(req.ModelID)
    if err != nil {
        return fmt.Errorf("unknown model: %s", req.ModelID)
    }
    
    // Check context window
    inputTokens := ce.tokenizer.Count(req.Prompt)
    if inputTokens > model.ContextLength {
        return fmt.Errorf("prompt too large: %d > %d", inputTokens, model.ContextLength)
    }
    
    // Check tool count
    if len(req.Tools) > model.MaxToolsPerRequest {
        return fmt.Errorf("too many tools: %d > %d", len(req.Tools), model.MaxToolsPerRequest)
    }
    
    // Check image limits
    if model.MaxImagesPerRequest > 0 && len(req.Images) > model.MaxImagesPerRequest {
        return fmt.Errorf("too many images: %d > %d", len(req.Images), model.MaxImagesPerRequest)
    }
    
    // Apply provider-specific overrides
    if req.ProviderID != "" {
        providerOverrides := model.ProviderOverrides[req.ProviderID]
        if providerOverrides != nil && providerOverrides.MaxTokens != 0 {
            if inputTokens > providerOverrides.MaxTokens {
                return fmt.Errorf("provider limit exceeded: %d > %d", inputTokens, providerOverrides.MaxTokens)
            }
        }
    }
    
    return nil
}

func (ce *ConstraintEnforcer) PostProcessResponse(resp LLMResponse) {
    // Track costs
    cost := ce.calculateCost(resp)
    ce.budget.Record(resp.SessionID, cost)
    
    // Check rate limits
    if resp.StatusCode == 429 {
        delay := ce.calculateBackoff(resp.RetryAfter)
        ce.rateLimiter.Delay(delay)
    }
    
    // Check remaining budget
    if ce.budget.IsExceeded(resp.SessionID) {
        ce.notifyUser("Budget exceeded, pausing session")
        ce.pauseSession(resp.SessionID)
    }
}
```

**Model Catalog Synchronization**:
```
┌─────────────────────────────────────────────────────────┐
│           Model Catalog Management                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Catalog Sources:                                      │
│  ┌─────────────────────────────────────────────┐   │
│  │  • Static bundled catalog (100+ models)    │   │
│  │  • Dynamic API fetch (OpenRouter Models API)  │   │
│  │  • User-defined model overrides             │   │
│  │  • Provider-specific metadata            │   │
│  └─────────────────────────────────────────────┘   │
│                                                         │
│  Update Frequency:                                      │
│  ┌─────────────────────────────────────────────┐   │
│  │  • Bundle: On app update                │   │
│  │  • Dynamic: Daily (or manual refresh)     │   │
│  │  • User overrides: Immediate               │   │
│  └─────────────────────────────────────────────┘   │
│                                                         │
│  Caching:                                             │
│  • Edge-cached (1 hour TTL)                      │
│  • Local SQLite cache (7 days)                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**OpenRouter Integration Example** (600+ models):
```go
// Fetch model catalog from OpenRouter
func fetchOpenRouterModels(apiKey string) ([]ModelConstraints, error) {
    resp, err := http.Get(
        "https://openrouter.ai/api/v1/models",
        "Authorization", "Bearer " + apiKey,
    )
    
    var models []OpenRouterModel
    json.Unmarshal(resp.Body, &models)
    
    // Normalize to Pryx schema
    var constraints []ModelConstraints
    for _, model := range models {
        constraints = append(constraints, ModelConstraints{
            ID: model.ID,
            CanonicalSlug: model.Pricing.ID,
            Name: model.Name,
            ContextLength: model.ContextLength,
            Tokenizer: mapTokenizer(model.Pricing.ID),
            Architecture: Architecture{
                InputModalities:  model.Architecture.InputModalities,
                OutputModalities: model.Architecture.OutputModalities,
                SupportsTools:     contains("tools", model.SupportedParameters),
                SupportsVision:    contains("file", model.Architecture.InputModalities),
                SupportsStreaming:  model.Streaming,
            },
            Pricing: Pricing{
                PromptPerMillion:     model.Pricing.Prompt,
                CompletionPerMillion: model.Pricing.Completion,
                SupportsCaching:     model.Pricing.InputCacheRead != "0",
            },
            PerRequestLimits: PerRequestLimits{
                MaxTokensPerRequest:     model.Pricing.MaxCompletionTokens,
                MaxToolsPerRequest:     10, // OpenRouter default
                MaxImagesPerRequest:     countImages(model.Architecture.InputModalities),
                MaxParallelToolCalls: 5,  // OpenRouter constraint
            },
        })
    }
    
    return constraints, nil
}
```

#### 8.4.2 Multi-Provider Model Routing

**Goal**: Automatically select optimal model based on constraints, costs, and capabilities.

**Routing Strategies**:

| Strategy | Use Case | Example |
|----------|------------|----------|
| **Cost optimization** | Minimize spend | Route to cheapest model meeting requirements |
| **Performance optimization** | Minimize latency | Route to fastest model (low percentile) |
| **Capability matching** | Ensure features | Route to model with vision/tools if needed |
| **Fallback chain** | Handle failures | Primary → Secondary → Tertiary on errors |
| **Provider diversity** | Avoid single point of failure | Distribute across providers |
| **User preference** | Respect choices | Always prefer user's default model |

**Cost-Aware Routing Example**:
```go
type ModelRouter struct {
    catalog      *ModelCatalog
    preferences  UserPreferences
}

func (mr *ModelRouter) SelectModel(req LLMRequest) (ModelConstraints, error) {
    // Get candidates matching requirements
    candidates := mr.catalog.FindModels(req.Requirements)
    
    // Filter by user preference
    if mr.preferences.PreferredModel != "" {
        candidates = filterPreferred(candidates, mr.preferences.PreferredModel)
    }
    
    // Sort by cost (user's budget preference)
    if mr.preferences.BudgetMode == "minimize_cost" {
        sort(candidates, byCostAscending)
        return candidates[0], nil
    }
    
    // Sort by performance (user's speed preference)
    if mr.preferences.BudgetMode == "maximize_speed" {
        sort(candidates, byLatencyAscending)
        return candidates[0], nil
    }
    
    // Default: balanced (cost + speed)
    sort(candidates, byScoreFunction)
    return candidates[0], nil
}

func ScoreFunction(model ModelConstraints, req LLMRequest) float64 {
    // Normalize cost (0-1 scale, lower is better)
    costScore := normalize(model.Pricing.CostPerRequest, 0, 0.1)
    
    // Normalize latency (0-1 scale, lower is better)
    latencyScore := normalize(model.AvgLatency, 0, 5) // 5s = worst
    
    // Weighted score
    return 0.6 * costScore + 0.4 * latencyScore
}
```

**Provider-Specific Constraint Handling**:

| Provider | Unique Constraints | Handling Strategy |
|-----------|-------------------|-------------------|
| **Anthropic** | Max 200K context, thinking tokens budget | Track `usage.reasoning_tokens` separately |
| **OpenAI** | 4K context (gpt-4o), 128K (gpt-4-turbo) | Model-specific `context_length` in catalog |
| **Together AI** | Variable by model, rate limits per tier | Provider-level overrides in catalog |
| **OpenRouter** | 600+ models, per-model limits, credits system | Dynamic fetch + per-request limits |
| **Local (Ollama)** | System memory, VRAM constraint | Hardware detection + model compatibility |

**Constraint Violation Handling**:
```go
type ViolationHandler struct {}

func (vh *ViolationHandler) Handle(err error, ctx RequestContext) error {
    switch err {
    case ContextLengthExceeded:
        // Option 1: Auto-chunk prompt
        if supportsAutoChunk(ctx.Model) {
            return vh.autoChunkAndRetry(ctx)
        }
        
        // Option 2: Truncate with summary
        summary := summarizeEarly(ctx.Prompt)
        return vh.retryWithTruncated(ctx, summary)
        
    case RateLimitExceeded:
        // Option 1: Queue for retry
        vh.rateLimiter.Queue(ctx.SessionID, ctx.Request)
        
        // Option 2: Switch model (fallback)
        fallbackModel := findCheaperModel(ctx.Requirements)
        return vh.retryWithFallback(ctx, fallbackModel)
        
    case ToolLimitExceeded:
        // Option 1: Reduce parallelism
        newConcurrency := min(ctx.Concurrency - 1, 1)
        return vh.retryWithConcurrency(ctx, newConcurrency)
        
    case BudgetExceeded:
        // Option 1: Pause session
        vh.pauseSession(ctx.SessionID)
        vh.notifyUser("Monthly budget reached, top up to continue")
        
        // Option 2: Downgrade model
        cheaperModel := findCheapestModel(ctx.Requirements)
        return vh.retryWithModel(ctx, cheaperModel)
        
    default:
        return err
    }
}
```

#### 8.4.3 Multi-Device Orchestration (Pryx Mesh)

**Core Principle**: Every Pryx instance is first-class citizen. Any device can initiate operations.

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Pryx Mesh (Multi-Device Coordination)             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │Device A  │  │Device B  │  │Device C  │  │Device D  │   │
│  │(Laptop)  │  │(Server)  │  │(Phone)  │  │(RPi)  │   │
│  │          │  │          │  │          │  │          │   │
│  │  Full    │  │  Full    │  │  Full    │  │  Remote   │   │
│  │  Pryx    │  │  Pryx    │  │  Pryx    │  │  Pryx    │   │
│  │  Runtime  │  │  Runtime  │  │  Runtime  │  │  Runtime  │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
│       │                │                │                │        │   │
│       └────────────────┴────────────────┘        │   │        │
│                          │                                  │        │   │
│                    ┌─────┴──────┐                        │   │
│                    │  Mesh Bus (Event)                        │   │
│                    │  - Session sync                          │   │
│                    │  - Device discovery                       │   │
│                    │  - Operation routing                    │   │
│                    │  - Constraint awareness                    │   │
│                    │  - Model routing                         │   │
│                    └───────────────────────────────────────────┘   │
│                                                                      │
│  Events:                                                          │
│  - session.continued_on_device[device_id]                         │
│  - operation.route[operation_id, target_device, source_device]      │
│  - device.pair_requested[device_id, pairing_code]             │
│  - file.transfer_requested[device_id, target_device, file_id]        │
│  - model.switch_request[source_device, target_device, model_id]        │
│  - constraint_violation[session_id, constraint_type, resolution]    │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

**Mesh Operations**:

| Operation | Initiated From | Handled On | Approval Required? | Model Consideration |
|-----------|----------------|-------------|------------------|-------------------|
| **Chat continuation** | Laptop → Phone | Phone | No | Use same model (session sync) |
| **File search** | Phone → Server | Server | No | Use device's default model |
| **File download** | Server → Phone | Server | Yes (policy check) | N/A |
| **Command execution** | Phone → Server (via SSH) | Server | Yes (policy check) | Use server's model config |
| **Web browse** | Phone → Server | Server | Yes (policy check) | Use browser-compatible model |
| **Multi-model task** | Any → Any | Any | No | Auto-select optimal model per subtask |

#### 8.4.2 Multi-Device Orchestration (Pryx Mesh)

**Core Principle**: Every Pryx instance is first-class citizen. Any device can initiate operations.

```
┌─────────────────────────────────────────────────────────────┐
│                     Pryx Mesh (Multi-Device Coordination)             │
├─────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │Device A  │  │Device B  │  │Device C  │  │Device D  │   │
│  │ (Laptop) │  │ (Server)  │  │ (Phone)  │  │  (RPi)  │   │
│  │          │  │          │  │          │  │          │   │
│  │  Full    │  │  Full    │  │  Full    │  │  Remote   │   │
│  │  Pryx    │  │  Pryx    │  │  Pryx    │  │  Pryx    │   │
│  │  Runtime  │  │  Runtime  │  │  Runtime  │  │  Runtime  │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
│       │                │                │        │   │        │
│       └────────────────┴────────────────┘        │   │        │
│                          │                                  │        │   │
│                    ┌─────┴──────┐                        │   │
│                    │  Mesh Bus (Event)                        │   │
│                    │  - Session sync                          │   │
│                    │  - Device discovery                       │   │
│                    │  - Operation routing                    │   │
│                    │  - Constraint awareness                    │   │
│                    └───────────────────────────────────────────┘   │
│                                                                      │
│  Events:                                                          │
│  - session.continued_on_device[device_id]                         │
│  - operation.route[operation_id, target_device, source_device]      │
│  - device.pair_requested[device_id, pairing_code]             │
│  - file.transfer_requested[device_id, target_device, file_id]        │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

**Mesh Operations**:

| Operation | Initiated From | Handled On | Approval Required? |
|-----------|----------------|-------------|------------------|
| Chat continuation | Laptop → Phone | Phone | No |
| File search | Phone → Server | Server | No |
| File download | Server → Phone | Server | Yes (policy check) |
| Command execution | Phone → Server (via SSH) | Server | Yes (policy check) |

#### 8.4.3 Cross-Device File Operations

**Use Cases**:
1. **Search on laptop, send to phone**: "Find /home/user/report.pdf and send to my WhatsApp"
2. **Browse on phone, download to laptop**: "Open Google Drive, find file, save to laptop"
3. **Transfer via WhatsApp**: "Download from server, send to device Y"
4. **Remote file access**: "Browser on Device X, access Drive Y, transfer to laptop"

**Implementation Patterns**:

**Pattern 1: Staged Transfer via Pryx Mesh**
```
User on Phone: "Send /data/photo.jpg to laptop"

Flow:
1. Phone agent uploads to Pryx Mesh cloud (optional storage)
2. Laptop agent discovers available file
3. User on laptop receives notification: "File available, download?"
4. User approves download
5. File transfers to laptop via Pryx Mesh
6. Phone receives confirmation
```

**Pattern 2: Direct Device-to-Device Access**
```
User on Laptop: "Browse my phone's gallery"

Flow:
1. Laptop agent requests file listing from phone device
2. Phone agent checks local permissions
3. If authorized: files listed (metadata only, no content)
4. Laptop agent requests specific file
5. Phone transfers file (content) via Pryx Mesh
6. Laptop receives file and saves
```

**Pattern 3: SSH-Based Remote Execution**
```
User on Phone: "Run git status on server"

Flow:
1. Phone agent initiates: pryx.device.request("ssh_command", {
     device_id: "server-device",
     command: "git status"
   })
2. Pryx Mesh routes to server device
3. Server device checks policy: workspace="/home/user/project", allow_shell=true
4. If approved: command executes in workspace directory
5. Results streamed back to phone via Pryx Mesh
6. Audit log records cross-device operation
```

**Pattern 4: WebDAV/SMB Network Mount**
```
User on Laptop: "I want files from phone's cloud storage"

Flow:
1. Laptop agent configures WebDAV mount point (phone-cloud as network device)
2. Phone's cloud storage appears as standard fs.read/fs.write tools
3. Standard fs tools work seamlessly across network mount
4. No content copied to Pryx Mesh (stream directly)
```

**Pattern 5: WhatsApp File Transfer**
```
User on Server: "Download this file and send to my WhatsApp"

Flow:
1. Server agent initiates: file.transfer("download", { file_id })
2. Pryx Mesh checks WhatsApp file size limits
3. If within limits: generates temporary download URL
4. Phone agent downloads from URL
5. File sent via WhatsApp message
6. URL expires (security)
```

**Security Considerations**:

| Operation | Security Measure | Implementation |
|-----------|------------------|-------------------|
| File download | User approval | Native dialog prompt on target device |
| Command execution | Scoped execution | Run in workspace directory only |
| File access | Policy check | Verify workspace/host/network scope |
| Data transfer | Encryption | End-to-end encryption via Pryx Mesh |
| Audit logging | Immutable record | All cross-device ops logged locally |

**Constraints Handled**:

| Constraint | Handling |
|-----------|------------|
| File size limits | Pre-transfer check, chunked transfer for large files |
| Storage quotas | Check available space before initiating |
| Network timeouts | 30s default, user-configurable |
| Concurrent operations | Queue-based, max 3 concurrent transfers |
| Device offline | Queue locally, sync on reconnect |

#### 8.4.4 Multi-Task Orchestration

**Scenarios**:
1. **Parallel independent tasks**: "Analyze logs AND summarize files"
2. **Sequential dependent tasks**: "Search file, THEN analyze content"
3. **Multi-surface monitoring**: "Watch file changes AND notify on all devices"
4. **Cross-device spawning**: "Spawn sub-agent on remote server"

**Implementation Strategy**:

```
┌─────────────────────────────────────────────────────────────┐
│                    Task Orchestration Engine                        │
├─────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Task Queue & Scheduler                           │   │
│  │                                                      │   │
│  │  - Priority Queue (High, Medium, Low)               │   │
│  │  - Task Dependencies (DAG-based)                   │   │
│  │  - Resource Allocation (CPU, memory, API rate)       │   │
│  │  - Concurrency Limits                             │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                  │   │
│  ▼                                  │   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Task Executor (Worker Pool)                     │   │
│  │                                                      │   │
│  │  - Spawn sub-tasks (go, async)            │   │
│  │  - Collect results                                 │   │
│  │  - Handle failures with retry                    │   │
│  │  - Update state in session store                │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                  │   │
│  ▼                                  │   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Session & State Management                         │   │
│  │                                                      │   │
│  │  - Track task status per session                     │   │
│  │  - Track cross-device operations                     │   │
│  │  - Constraint awareness per provider                 │   │
│  │  - Multi-surface event broadcasting                 │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

**Queue Prioritization**:

| Priority | Task Type | Example |
|----------|-----------|----------|
| **High** | User-initiated (chats, file ops) | "Send file to phone" |
| **Medium** | Background tasks (indexing, analysis) | "Reindex workspace" |
| **Low** | Scheduled tasks (backups, sync) | "Daily backup at 2am" |

**Constraint Propagation**:

```
Task Queue
  ├── task_a (chat_with_laptop)
  │   └── requires: { model: claude, tokens: 50k }
  │   └── provider_limits: { max_tokens: 200k, thinking: 10k }
  │   └── auto: { chunk: true, compact: thinking_budget }
  │   └── task_b (analyze_files_on_laptop)
      └── requires: { fs_ops: true, workspace: "/home/user" }
```

**Failure Handling**:

| Failure Mode | Recovery Strategy | User Notification |
|-------------|------------------|-------------------|
| Rate limit hit | Exponential backoff, queue for retry | "Rate limited, retrying in 10s" |
| Context exceeded | Chunk request, summarize context | "Context limit, split into 2 calls" |
| Device offline | Queue locally, sync on reconnect | "Phone offline, queued locally" |
| Provider error | Fallback to alternative provider | "Claude unavailable, trying GPT-4" |

### 10.7 Autocompletion & Long-Running Task Management (NEW)

**Why This Matters**: Long-running continuous tasks (e.g., "watch logs and auto-fix", "monitor POS transactions", "debug in loop") present token efficiency challenges. Users need to understand:
- "Will continuous monitoring consume too many tokens?"
- "If I'm working in OpenCode while Pryx runs in background, how do updates work?"
- "What's the UX for agent waiting for user attention?"
- "How does the agent handle context window during multi-hour operations?"

---

#### 10.7.1 Background Process Management (OpenCode Pattern)

**Core Capability**: Long-running processes (log monitoring, file watching, cron jobs, debugging loops) run independently with controlled output streaming.

**Architecture**:
```
┌─────────────────────────────────────────────────────────────┐
│  User (OpenCode/Cursor/IDE)                             │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Pryx Agent (User's Device)                         │
├─────────────────────────────────────────────────────────────┤
│                                                         │
│  ┌───────────────────────────────────────────────────┐   │
│  │  Background Process Manager                  │   │
│  │  • createBackgroundProcess(...)            │   │
│  │  • listBackgroundProcesses()             │   │
│  │  • killProcesses()                        │   │
│  └───────────────────────────────────────────────────┘   │
│                                                         │
│  ┌───────────────────────────────────────────────────┐   │
│  │  Token Optimization Layer                   │   │
│  │  • Context pruning                       │   │
│  │  • Cache TTL management                  │   │
│  │  • Output stream limiting               │   │
│  └───────────────────────────────────────────────────┘   │
│                                                         │
│  ▼                                                    │
┌─────────────────────────────────────────────────────┐
│  Log/File Stream (Continuous Monitoring)          │
└─────────────────────────────────────────────────────┘
```

**Background Process Features**:
- **In-memory task tracking** with unique process IDs
- **Output stream limiting**: Last 100 lines for individual task, last 10 lines for task list
- **Tag-based filtering** for categorization (monitoring, debugging, deployment)
- **Global vs session-specific** processes
- **Automatic cleanup** on session end

**Implementation (TypeScript)**:
```typescript
class PryxBackgroundProcessManager {
  private processes: Map<string, PryxBackgroundProcess> = new Map();

  createProcess(input: {
    command: string;
    name?: string;
    tags?: string[];
    global?: boolean;
    sessionId?: string;
    mode?: 'pump-dump' | 'streaming' | 'hybrid';
    outputConfig?: {
      bufferLines?: number;        // Default: 100
      circularBufferSize?: number; // For infinite streams: 500
      errorPatterns?: string[]; // For state machine transitions
      summaryInterval?: string;   // For pump-dump mode
    };
  }): string {
    const subprocess = execa(input.command, {
      shell: true,
      stdout: 'pipe',
      stderr: 'pipe',
    });

    const process = new PryxBackgroundProcess({
      id: 'proc-' + nanoid(),
      command: input.command,
      name: input.name || input.command,
      tags: input.tags || [],
      global: input.global ?? false,
      mode: input.mode ?? 'hybrid',
      sessionId: input.sessionId,
      pid: subprocess.pid,
      status: 'running',
      startedAt: new Date(),
      outputConfig: input.outputConfig || {},
    });

    this.processes.set(process.id, process);

    // Implement output buffering with circular buffer
    this.setupOutputCapture(process, subprocess);
    
    // Subscribe to subprocess events
    subprocess.stdout?.on('data', (chunk: Buffer) => {
      const line = chunk.toString().trim();
      if (line) process.recordOutput(line);
    });
    // ... error handling and completion
  }

  listProcesses(filters: { tags?: string[], status?: string[] }): ProcessSummary[] {
    return Array.from(this.processes.values())
      .filter(p => this.matchesFilters(p, filters))
      .map(p => ({
        id: p.id,
        name: p.name,
        status: p.status,
        mode: p.mode,
        tags: p.tags,
        currentState: p.currentState,
        outputStream: p.outputStream.slice(-10), // Last 10 lines
        pid: p.pid,
        startedAt: p.startedAt,
        completedAt: p.completedAt,
      }));
  }
}
```

---

#### 10.7.2 Token Efficiency Strategies (Clawdbot Pattern)

**Core Principles**:
1. **Compact history**: Use `/compact` to summarize long sessions
2. **Trim tool outputs**: Limit verbose tool results
3. **Short skill descriptions**: Reduce system prompt overhead
4. **Prefer smaller models**: For verbose, exploratory work

**Cache TTL Management**:
```yaml
agents:
  defaults:
    model:
      primary: "anthropic/claude-opus-4-5"
    models:
      "anthropic/claude-opus-4-5":
        params:
          cacheControlTtl: "1h"
  heartbeat:
      every: "55m"  # Keep cache warm just under TTL
```

**Session Pruning**:
- **Only prunes tool results** (not user/assistant messages)
- **Protects last N assistant messages** (default: 3)
- **Soft-trim**: Keeps head + tail, inserts `...`
- **Hard-clear**: Replaces entire result with placeholder `[Old tool result content cleared]`
- **Skips image blocks** (never trimmed)

---

#### 10.7.3 Streaming vs "Pump-and-Dump" Patterns

**Block Streaming** (Channel Messages):
- Emits **completed blocks** as assistant writes
- Not token deltas, but coarse chunks
- Controlled by: `blockStreamingDefault`, `blockStreamingBreak`, `blockStreamingChunk`

**Streaming Configuration**:
```typescript
{
  blockStreamingDefault: "on" | "off",
  blockStreamingBreak: "text_end" | "message_end",
  blockStreamingChunk: {
    minChars: 500,
    maxChars: 2000,
    breakPreference: "paragraph" | "newline" | "sentence" | "whitespace"
  },
  blockStreamingCoalesce: {
    minChars: 1500,
    maxChars: 3000,
    idleMs: 200  // Wait before flushing
  }
}
```

**Coalescing**: Merges consecutive block chunks before sending, waits for idle gaps to reduce "single-line spam".

**Key Insight**: Coalescing waits for idle gaps before flushing, reducing token spam while providing progressive output.

---

#### 10.7.4 Heartbeat System for Continuous Operations

**Purpose**: Keep prompt cache warm across idle gaps, reducing cache write costs.

**Configuration**:
```yaml
heartbeat:
  every: "30m"  # Default, configurable
  includeReasoning: false  # Set to true to send separate Reasoning message
  ackMaxChars: 200  # Padding for HEARTBEAT_OK
```

**How It Works**:
1. Read HEARTBEAT.md if it exists (workspace context)
2. Follow it strictly - Do not infer or repeat old tasks
3. If nothing needs attention → reply HEARTBEAT_OK
4. Inline directives in heartbeat message apply as usual
5. Send heartbeat only (no delivery to avoid double delivery)

**Benefits**:
- Reduces cache write costs during idle periods
- Prevents prompt cache evictions
- Maintains context between long-running operations

---

#### 10.7.5 Agent Waiting UX Patterns

**Scenario**: Agent completes task, waiting for user instruction/continue signal.

**Implementation**:
```typescript
agent.wait uses waitForAgentJob:
  - waits for **lifecycle end/error** for runId
  - returns: { 
      status: "ok" | "error" | "timeout", 
      startedAt, 
      endedAt, 
      error? 
    }
```

**Key Features**:
- **Separate timeout**: 30s wait timeout (default) vs 600s agent timeout
- **Gateway RPC endpoints**: `agent` and `agent.wait`
- **Serialized execution**: Per session + global queues

---

#### 10.7.6 Context Window Optimization Techniques

**What Counts in Context Window**:
- System prompt (tool descriptions, skills, bootstrap)
- Conversation history (user + assistant messages)
- Tool calls and tool results
- Attachments and transcripts
- Compaction summaries and pruning artifacts

**Optimization Hierarchy** (from highest to lowest impact):

1. **System Prompt Trimming**:
   - Limit bootstrap files to 20,000 chars (default)
   - Keep skill descriptions concise
   - Use tool allow/deny lists

2. **Session Pruning** (cache-ttl mode):
   - Prune old tool results after TTL expires
   - Protect last N assistant messages (default: 3)
   - **Soft-trim**: head + tail (30% ratio)
   - **Hard-clear**: replace with placeholder (50% ratio)
   - **Skip image blocks** (never trimmed)

3. **Output Stream Limiting**:
   - Background tasks: Last 100 lines (individual)
   - List view: Last 10 lines
   - Implement circular buffer for infinite streams

4. **Compaction**:
   - Summarize old messages
   - Archive to disk
   - Inject summary instead of full history

---

#### 10.7.7 Recommended Architecture for "Watch & Auto-Fix" Use Case

**User Scenario**: "Watch logs on server, detect errors, attempt fixes in loop until no more errors"

**Recommended Strategy**: **Pump-and-Dump with Periodic Summaries**

**Token Cost**: ~500 tokens per summary (every 10 minutes)
**User Benefits**:
- Minimal token usage (stateful agent, only summaries sent)
- User control (can stop/pause anytime)
- Persistent monitoring (background process continues)
- Context window optimization (compact summaries only)

**Flow**:
```
1. Create background process: tail -f -n 100 /var/log/app.log
2. Agent reads output in chunks (100 line buffer)
3. Pattern matching in-memory:
   - Detect errors
   - Identify recurring patterns
   - Maintain compact state
4. Every N iterations OR on error detection:
   - Generate compact summary of findings
   - Send to user
   - Wait for "continue" signal
   - Clear memory except critical state
   - Resume monitoring
```

**Configuration**:
```json
{
  "backgroundTask": {
    "name": "log-monitor",
    "command": "tail -f -n 100 /var/log/app.log",
    "tags": ["monitoring", "logs", "auto-fix"],
    "global": true,
    "mode": "pump-dump",
    "outputConfig": {
      "bufferLines": 100,
      "errorPatterns": ["ERROR", "FATAL", "Exception"],
      "summaryInterval": "10m"
    }
  },
  "contextOptimization": {
    "mode": "cache-ttl",
    "ttl": "5m",
    "keepLastAssistants": 3,
    "softTrimRatio": 0.3,
    "hardClearRatio": 0.5
  }
}
```

---

#### 10.7.8 When User is Working in Other Tools (Your Scenario)

**Scenario**: User is actively working in OpenCode/Cursor while Pryx agent runs background task on their device.

**Recommended Strategy**: **Pump-and-Dump with Periodic Summaries**

**Rationale**:
1. User is actively working - they don't need real-time streaming
2. Pryx agent runs autonomously - token cost should be minimal
3. User can always request latest summary by typing `continue`
4. Compact summaries every 10 minutes = excellent awareness vs cost trade-off

**UX Pattern**:
```
Initial Setup:
User: "Watch logs on server, detect errors, attempt fixes in loop until no more errors"

Pryx Agent:
"✅ Started background log monitor (pid: 12345)"
"✅ Monitoring /var/log/app.log"
"✅ Will send summaries every 10 minutes"
"✅ Auto-fix attempts enabled"
Type 'continue' for next summary, 'status' for current state,
or 'stop' to end monitoring"

[10 minutes later...]
Pryx Agent:
"📊 Monitoring Summary (10:00 - 10:10)"
• Lines processed: 2,543
• Errors detected: 3
  - ERROR: Database connection failed (10:02:15)
  - FATAL: Out of memory (10:05:22)
  - ERROR: Payment gateway timeout (10:07:01)
• Auto-fix attempts: 2
  - [10:02:18] Restarted database service (SUCCESS)
  - [10:06:33] Increased memory limit (PENDING)
  - [10:08:00] Payment gateway health check (FAILED)
• Status: 2 errors remaining
Next summary: 10:20

Type 'continue' to continue monitoring with detailed logs,
or 'fix [error-id]' to attempt manual fix for specific error,
or 'status' to see full recent output,
or 'stop' to end."
```

**User in OpenCode** - sees notification in Slack/Discord:
```
[Channel: #dev-ops]
[Bot: @Pryx]
[Notification] Found 3 errors in server logs. See summary above.
```

User can then:
- Check in Pryx UI for full logs
- Take action on specific error
- Continue monitoring or stop task

---

#### 10.7.9 Implementation Recommendations

**Phase 1: Background Process Manager** (Week 1-2)
- Implement `PryxBackgroundProcessManager` class
- Support in-memory task tracking with unique IDs
- Implement output buffering with circular buffer (500 lines)
- Add tag-based filtering for categorization
- Add global vs session-specific processes

**Phase 2: Token Optimization Layer** (Week 2-3)
- Implement session pruning with cache-ttl mode
- Add output stream limiting (100 lines individual, 10 lines list)
- Implement soft-trim and hard-clear strategies
- Add context window tracking

**Phase 3: Adaptive Streaming Controller** (Week 3-4)
- Implement state machine (idle/monitoring/action/completed)
- Add streaming with coalescing support
- Add immediate streaming for critical actions
- Add error pattern detection and consecutive error tracking

**Phase 4: Heartbeat Integration** (Week 4)
- Implement heartbeat configuration system
- Add support for HEARTBEAT.md workspace file
- Implement ack suppression to avoid double delivery
- Add customizable heartbeat intervals

**Phase 5: Configuration UI** (Week 4-5)
- Add UI for selecting autocompletion mode (pump-dump/streaming/hybrid)
- Add UI for configuring output limits (buffer size, coalescing settings)
- Add UI for error pattern configuration
- Add UI for context optimization settings (TTL, trim ratios)
- Add UI for heartbeat configuration

**Phase 6: Documentation & Examples** (Week 5)
- Document all three strategies with pros/cons
- Provide configuration examples for each use case
- Create templates for common background tasks
- Document best practices for token efficiency

**Configuration Example**:
```json
{
  "backgroundTasks": {
    "defaultMode": "hybrid",
    "defaultGlobal": false,
    "outputBuffer": {
      "bufferLines": 100,
      "circularBufferSize": 500
    }
  },
  "contextOptimization": {
    "mode": "cache-ttl",
    "ttl": "5m",
    "keepLastAssistants": 3,
    "softTrimRatio": 0.3,
    "hardClearRatio": 0.5
  },
  "streaming": {
    "blockStreamingDefault": "on",
    "blockStreamingBreak": "text_end",
    "blockStreamingChunk": {
      "minChars": 500,
      "maxChars": 2000,
      "breakPreference": "paragraph"
    },
    "blockStreamingCoalesce": {
      "minChars": 1500,
      "maxChars": 3000,
      "idleMs": 200
    }
  },
  "heartbeat": {
    "enabled": true,
    "interval": "30m",
    "includeReasoning": false,
    "ackMaxChars": 200
  }
}
```

---

#### 10.7.10 Token Efficiency Heuristics

| Scenario | Mode | Token Cost (per hour) | Recommended |
|----------|------|------------------------|-------------|
| Idle monitoring | Pump-dump (10m summaries) | ~500 tokens | ✅ Best for background work |
| Error detected | Streaming (coalesced) | ~2,000-5,000 tokens | ✅ Necessary for awareness |
| Auto-fix active | Streaming immediate | ~1,000-5,000 tokens | ✅ Required for transparency |
| User working in other tool | Pump-dump (10m summaries) | ~500 tokens | ✅ Perfect balance |
| User actively watching | Streaming with coalescing | ~2,000-5,000 tokens | ✅ Real-time feedback |

### B. Related Documents

- `docs/prd/PRD-UPDATES.md` - Auto-Update mechanism specification (research complete, integrated into this PRD)
- `docs/plan/plan-a.md` - Original Pryx PRD (superseded)
- `docs/plan/plan-b.md` - Clawdbot Reborn strategy (superseded)
- `docs/final-prd.md` - Unified PRD draft v0.1 (superseded)
- `docs/prd/prd-v2.md` - Roadmap (post-v1 planning)
- `docs/prd/pryx-mesh-design.md` - Multi-device coordination architecture
- `docs/prd/autocompletion-background-tasks.md` - Autocompletion & background tasks (NEW, based on OpenCode/Clawdbot research) - **NEW**
- `docs/prd/plugin-architecture.md` - Plugin architecture & third-party integration (research in progress) - **TODO**
- `docs/prd/scheduled-tasks.md` - Scheduled tasks platform (NEW, v1.1+)
- `docs/prd/ai-assisted-setup.md` - AI-assisted configuration flows & 3-phase implementation plan (NEW) - **NEW**
- `docs/prd/implementation-roadmap.md` - Complete 3-phase implementation timeline and task breakdown (NEW) - **NEW**
- `docs/templates/SYSTEM.md` - AI system prompt with cron capability documentation (NEW) - **NEW**
- `docs/guides/cron-natural-language.md` - User guide for natural language scheduling (NEW) - **NEW**
- `docs/prd/cron-anti-hallucination-summary.md` - Anti-hallucination strategy for natural language features (NEW) - **NEW**

---

*This document is the canonical source of truth for Pryx product requirements. All implementation decisions should reference this PRD.*
