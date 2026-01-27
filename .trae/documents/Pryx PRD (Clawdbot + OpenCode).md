# Product Requirements Document (PRD): Pryx (v0.2)

## 0) Product Principles (Non-Negotiables)
- Sovereign-by-default: user data stays on user-owned machines unless explicitly enabled.
- Zero-friction onboarding: no Node/Python dependency requirement for end users.
- Safe execution: human-in-the-loop approvals + policy engine.
- Observable: every action is traceable locally; optional sanitized export.
- Simple + reusable: centralized logic; avoid duplicated flows and config surfaces.

## 1) Executive Summary
Pryx is a second-generation sovereign AI agent inspired by Clawdbot and OpenCode. It provides a “localhost-first” control center and a fast CLI/TUI workflow, shipped as compiled binaries with one-liner install. The cloud layer (Cloudflare) handles device-flow auth, telemetry ingress, docs/changelog web, optional device registry/config backup, and can also operate as an AI gateway for model routing and key management.

## 2) Problem Statement
Local AI agents today commonly suffer from:
- Installation/runtime friction (dependency hell, Node versioning).
- Blind state (headless operation without strong health/telemetry/UX).
- Unsafe tool execution (overbroad privileges, weak approvals).
- No coherent multi-device story (server + laptop coordination is ad-hoc).

## 3) Target Users  Use Cases
### Personas
- Builder (CLI-first): wants speed, keyboard workflow, automation.
- Operator (server-first): wants headless agent on a server, with local UI access.
- Power user: wants multi-channel integrations and clear governance.

### Primary Use Cases
- Code/workspace agent: file ops, shell, git, HTTP tools with approvals.
- Channel agent: Telegram/Discord/Slack/webhooks as inbound triggers.
- Multi-device: laptop UI controls server agent; local agent can delegate.
- Observability: timeline, costs, errors, tool calls, approvals.

## 4) Scope: MVP vs v1
### MVP (Ship fast, validate product)
- One-liner installer (macOS/Linux + Windows option).
- CLI + TUI as first UI surface.
- Localhost UI (web dashboard) for onboarding/config/telemetry.
- Agent runtime sidecar (Go) with HTTP+WebSocket API.
- Cloudflare auth worker (OAuth Device Flow).
- Cloudflare telemetry worker (OTLP ingest + PII redaction).
- Health checks + diagnostics (`pryx doctor`).

### v1 (Make it durable)
- Desktop app wrapper (optional): Tauri host for native dialogs + tray + deep links.
- Multi-device pairing (“Pryx Mesh”) with device registry.
- AI Gateway mode on Cloudflare: routing, quotas, key mgmt, redaction.
- Docs + changelog web fully integrated into the control plane.
- Optional cloudstore configuration + device-level metadata.

## 5) Product Surfaces
### 5.1 CLI  TUI (Required)
- Default command: `pryx` launches TUI (OpenCode-like).
- Key screens:
  - Sessions: list, resume, export.
  - Chat: streaming responses, tool approval prompts.
  - Integrations: connect/disconnect, status.
  - Telemetry: timeline + costs (summary in TUI; deep dive in web UI).
- Non-interactive mode:
  - `pryx run "..."` for scripts/CI.
  - `pryx serve` for headless.

### 5.2 Localhost Web UI (Required)
- Onboarding wizard: connect integrations, show QR/device codes.
- Channel manager: add/remove channels, routing rules.
- Observability dashboard: trace timeline, approvals, errors, costs.
- Admin settings: model routing, telemetry toggle, policies.

### 5.3 Desktop Wrapper (Optional in MVP)
- Adds native permission dialogs and deep links.
- If deferred, approvals are handled in TUI and web UI.

## 6) Port Strategy (Must avoid conflicts)
- Default: bind to a random free port (no hard-coded 3000/3001 required).
- Discovery: print bound ports on stdout + write a local metadata file (for UI/TUI attach).
- Override:
  - `--port` / `PRYX_PORT` for API.
  - Separate UI port if UI is served independently.
- Make/Docker: reserved port computation to avoid collisions with known tools.

## 7) Installation  Updates  Uninstall (Critical Improvement)
### 7.1 One-liner install
- `curl https://get.pryx.dev | sh` (plus PowerShell equivalent).
- Installs a single `pryx` launcher + `pryx-core` runtime as needed.
- No external runtime dependencies.

### 7.2 Service management
- Optional: install as user service (launchd/systemd) for server mode.
- Clear commands:
  - `pryx install-service`
  - `pryx uninstall-service`

### 7.3 Auto-updates
- Default: prompt-based updates.
- Optional: background update channel (stable/beta).
- Signed releases and integrity verification.

### 7.4 Uninstall
- Documented and reliable removal (binaries, service, local data optional).

## 8) Architecture (Refined)
### 8.1 Local components
- `pryx` (Supervisor/CLI/TUI):
  - Spawns and monitors `pryx-core`.
  - Presents approvals in TUI; later can delegate to native dialogs.
- `pryx-core` (Agent runtime, Go):
  - LLM orchestration.
  - Tool protocol (MCP).
  - Channel/webhook ingestion.
  - Local HTTP+WebSocket server.
- Web UI (TS/React):
  - Either bundled as static assets served by `pryx-core` OR run separately in dev.

### 8.2 Cloudflare control plane
- Auth Worker: RFC 8628 device flow; third-party OAuth as needed.
- Telemetry Worker: OTLP intake; redaction; forwarding.
- Docs/Changelog Web: static pages + worker routing.
- Optional stateful modules (only if user enables):
  - Durable Objects for short-lived sessions/pairing and rate limiting.
  - KV/D1 for minimal device registry/config metadata.

### 8.3 AI Gateway mode (Optional but high impact)
- Central model routing + policy enforcement.
- Key vault abstraction (users store provider keys locally; gateway issues short-lived tokens).
- Rate limiting, quotas, caching, redaction.
- Strongly separate from chat history (no chat storage).

## 9) Functional Requirements (Expanded)
### FR1: Agent Runtime
- Headless: `pryx-core` runs without UI.
- Concurrency: handle =50 webhook/channel events concurrently.
- Tool execution via MCP; tools discoverable dynamically.
- WebSocket event stream: tokens, tool calls, approvals, errors.

### FR2: Approvals  Policy Engine (Product-defining)
- Every sensitive tool call produces a “policy decision”:
  - allow/deny/ask
  - scope: workspace, host, network domain, integration
  - duration: once, session, forever (with expiry)
- Default policies:
  - read-only safe mode
  - workspace-only mode
  - network allowlist mode

### FR3: Integrations  Channels (Channel Mux)
- Unified channel abstraction:
  - inbound messages/events
  - outbound replies
  - rate limiting and retry
  - identity mapping
- MVP channels: Telegram + generic webhooks.
- v1 channels: Discord, Slack; WhatsApp strategy defined explicitly (Cloud API vs subprocess).

### FR4: Routing Rules
- Map channels to agents/workspaces.
- Controls:
  - mention-only activation
  - allowlist/denylist users
  - per-channel budgets/token limits
  - group vs DM behavior

### FR5: Observability
- Local traces always available.
- Optional export:
  - PII redaction before leaving device
  - sampling controls
  - cost/latency breakdown per model/channel

### FR6: Multi-device “Pryx Mesh”
- Pair devices via QR/device code.
- Device registry includes:
  - capabilities, version, status
  - approved policies (summary)
- Connectivity modes:
  - LAN direct
  - tunneled relay (Cloudflare-based) when needed

## 10) Data  Storage Requirements
- Local:
  - sqlite for sessions, audit log, configuration.
  - vector store optional, local only.
- Cloud (optional):
  - minimal config and device metadata only.
  - explicit opt-in for backups.

## 11) Security Requirements (Strengthened)
- OS keychain for secrets.
- Scoped permissions:
  - workspace root enforcement
  - network domain allowlists
- Signed artifacts and integrity checks on install/update.
- Audit log separate from chat (tamper-evident goal for v1).
- Clear threat model section in docs.

## 12) Reliability Requirements
- No zombie processes; deterministic shutdown.
- Crash restart with session recovery.
- `/health` endpoint and `pryx doctor`.
- Offline resilience:
  - queue inbound channel messages during restart

## 13) Developer/Operator Experience
- Docker + Make standardization:
  - Make targets for lint/test/build for local and server.
  - Docker Compose example for headless server deployments.
- Submodule boundaries: excluded from general CI/lint unless explicitly needed.

## 14) Success Metrics
- Install-to-first-chat 2 minutes.
- Install-to-first-channel-connected 5 minutes.
- Zero orphan processes after exit.
- Telemetry clarity: users can answer “what happened?” in 30 seconds.
- Multi-device pairing success rate and stability.

## 15) “What Else Makes It Great” (Added)
- Cost controls:
  - daily/weekly budgets
  - per-agent and per-channel limits
  - model routing by price/performance
- Profile system:
  - Work/Personal/Server policies and keys
- Safe mode:
  - enforce read-only tools until explicitly enabled
- Export/import:
  - encrypted export of config + sessions
- Update/rollback:
  - stable/beta channels and rollback path
- Extensible tool registry:
  - local registry first; optional cloud catalog later

---

# Updated Execution Plan (after confirmation)
## Milestone A: Freeze PRD  Acceptance Criteria
- Turn this PRD into a versioned doc + acceptance criteria checklist per FR/SR.

## Milestone B: Installer + Lifecycle
- One-liner install, service install/uninstall, updates, and uninstall flow.

## Milestone C: Core Runtime + APIs
- Go runtime with localhost API, WebSocket events, tool approvals.

## Milestone D: CLI/TUI
- TUI UX and approvals; connect flows.

## Milestone E: Cloudflare Control Plane
- Auth worker, telemetry worker, docs/changelog web, optional config/device registry.

## Milestone F: Mesh + Gateway (v1)
- Device pairing and connectivity; AI gateway routing/quotas.
