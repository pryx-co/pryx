## Goal
Create a single, modern PRD for Pryx that merges and strengthens the product + architecture from plan-a.md (Pryx PRD) and plan-b.md (Clawdbot Reborn strategy), resolving conflicts and adding robustness.

## Deliverables (After You Confirm)
- Add a new doc: `docs/prd/prd.md` (markdown) containing the unified PRD below.
- Keep existing `docs/plan/plan-a.md` and `docs/plan/plan-b.md` unchanged.

## Key Improvements Over Both Plans
- One coherent product definition (Pryx name, scope, personas, success metrics).
- Hard requirements for reliability/security (crash isolation, permission gating, secret storage, auditability).
- Clear separation of “End-user binary” vs “Dev/Server Docker” so we get both zero-config UX and consistent ops.
- Explicit extension story: MCP-first tool/plugin model, policy engine, and channel adapters.

## Unified PRD Draft (v0.1)

# Product Requirements Document (PRD): Pryx

## 1) Executive Summary
Pryx is a local-first “sovereign AI agent” that feels like a polished desktop app but runs like a resilient daemon. It replaces fragile CLI + config-file assistants with a first-class UI, secure permission gating, and transparent observability—while keeping user data on-device. Cloudflare Workers provide a thin edge control plane for OAuth device flows and telemetry ingestion without persisting user conversations.

## 2) Problem Statement
Existing local AI agents commonly fail on:
- Installation friction (runtime/dependency hell).
- No management UI (setup scattered across CLI + JSON edits).
- Opaque security posture (keys and data flows unclear).
- Lack of observability (hard to know what the agent did, why, and at what cost).

## 3) Product Vision
Deliver the “Localhost experience” of a modern SaaS dashboard—with local sovereignty.
- Local machine owns: conversations, files, indices, policies, secrets.
- Cloud edge owns: auth handshakes, pairing, optional telemetry relay.

## 4) Target Users & Personas
- Solo builder: wants automation + chat with safe guardrails.
- Operator/power user: runs Pryx headless on a server/RPi, manages via browser.
- Security-conscious user: demands explicit approvals, audit logs, and keychain storage.

## 5) Goals & Non-Goals
### Goals
- Zero/low-friction install for end-users (native bundle; no Node/Python required).
- First-class UI for onboarding, integrations, sessions, policies, and telemetry.
- Crash-resilient architecture via sidecar separation.
- Transparent observability (traces/cost/errors) without sacrificing privacy.
- Extensible tools/integrations via MCP.

### Non-Goals (Initial)
- Mobile native apps.
- Full “skills marketplace” monetization.
- Mandatory cloud sync of conversation history.

## 6) Core User Journeys
1. Install Pryx → open app (or start daemon).
2. Onboarding wizard → pick workspace + model provider.
3. Connect an integration (Telegram/Slack/GitHub) via QR/device code.
4. Run an agent task → see required permission prompts.
5. Inspect session timeline (traces) → export/share a redacted report.
6. Update Pryx safely (auto-update with rollback).

## 7) Functional Requirements
### FR1: Agent Runtime (Sidecar)
- FR1.1 Headless-capable: sidecar runs without UI for server deployments.
- FR1.2 Tool extensibility: implement MCP to discover/execute tools.
- FR1.3 Concurrency: handle ≥50 concurrent inbound events/webhooks.
- FR1.4 Policy-aware execution: every tool call is evaluated against a local policy engine (allow/deny/ask).
- FR1.5 Local persistence: session state and logs stored locally (sqlite by default).

### FR2: Host (Desktop Shell / Supervisor)
- FR2.1 Sidecar supervision: start/stop, health checks, port discovery, crash restart.
- FR2.2 Permission gating: privileged actions require explicit user approval in native dialogs.
- FR2.3 Deep linking: `pryx://` links to sessions/settings/auth callbacks.

### FR3: UI (Control Center)
- FR3.1 Visual configuration: integration setup via wizard (QR/device codes), no JSON editing.
- FR3.2 Session explorer: list sessions, search, export.
- FR3.3 Observability view: live trace timeline (Gantt), tool calls, model tokens, and cost.
- FR3.4 Policy management: define workspace scopes, allowlists/denylists, and approval modes.

### FR4: Integrations & Channels
- FR4.1 Channel abstraction: adapters for chat platforms + webhooks.
- FR4.2 Routing rules: map channels → agents/workspaces; DM vs group rules; allowlists.
- FR4.3 Status & recovery: per-channel health, reconnect, rate-limit handling.

### FR5: Auth & Edge Control Plane (Cloudflare Workers)
- FR5.1 OAuth Device Flow (RFC 8628) and QR pairing.
- FR5.2 Stateless by default: workers do not persist conversations.
- FR5.3 Telemetry ingestion: accept OTLP/OTel batches; sanitize PII; forward to user-chosen backend.

### FR6: Installation, Updates, and Ops
- FR6.1 End-user distribution: native bundle (macOS/Windows/Linux).
- FR6.2 Auto-update: background update checks + safe rollback.
- FR6.3 Dev + server consistency: Docker + Make targets for build/test/lint/run; production supported via Docker Compose.

## 8) Non-Functional Requirements
### Security & Privacy
- All conversations and indices stay local unless user enables export/backup.
- Secrets stored in OS keychain/credential manager (never plaintext config).
- Least privilege: sandbox tool execution to selected workspace by default.
- Auditability: immutable local audit log of tool calls + approvals.

### Reliability
- No zombie processes; graceful shutdown.
- Sidecar crash isolation; automatic restart with backoff.
- Offline-first: UI works with local daemon; queue events during brief restarts.

### Performance
- Time-to-first-chat: <2 minutes from install.
- Fast startup: UI visible quickly; sidecar ready within seconds.

## 9) Proposed Architecture
- Host: Rust + Tauri (desktop lifecycle, permission dialogs, sidecar supervisor).
- UI: React + TypeScript (bundled webview; optional standalone web UI for headless mode).
- Sidecar: Go (agent loop, MCP tools, channel adapters, websockets/HTTP).
- Edge: Cloudflare Workers (device auth, telemetry ingress; optional durable session metadata only if explicitly required).

## 10) Data & Storage Model (Local-First)
- SQLite for sessions, audit logs, traces (cached), integration metadata.
- File storage for exports and optional indexed artifacts.
- Keychain for provider API keys and OAuth refresh tokens.

## 11) Risks & Mitigations
- WhatsApp integration complexity: prefer official APIs; otherwise isolate Node subprocess behind adapter boundary.
- OAuth edge complexity: keep workers stateless, strict token lifetimes, rotate signing keys.
- Observability privacy: default sampling off/on-device; redact sensitive fields.

## 12) Success Metrics
- Installation: <2 minutes to first successful chat.
- Onboarding: connect first integration in <3 minutes.
- Reliability: 0 zombie sidecars after app exit; crash restart success rate ≥99%.
- Transparency: ≥90% of sessions have a complete trace timeline.
- Extensibility: add a new MCP tool without rebuilding the host.

## 13) Milestones (Scope-Oriented)
- M1: Sidecar MVP + UI onboarding + one integration + permission gating.
- M2: Device flow auth worker + OTLP ingest + basic dashboards.
- M3: Multi-channel + routing rules + export/import.
- M4: Hardening (encryption, rate limiting, offline queue) + docs.

## Implementation Steps (After You Confirm)
1. Create `docs/prd/` and write `docs/prd/prd.md` with the PRD above.
2. Add a short note at top of the PRD stating it supersedes plan-a/plan-b for product definition.
3. Quick consistency pass to align terminology with the repo (Pryx naming, sidecar naming).

If you confirm, I’ll generate the new PRD file exactly as drafted (or with minor copy-edits for consistency).