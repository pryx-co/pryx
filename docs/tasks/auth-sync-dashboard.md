# End-to-End: Install/Login → Sync → Dashboard

> **Status**: Canonical (design + UX contract)  
> **Last Updated**: 2026-01-28  
> **Applies To**: `docs/prd/prd.md` (v1), `docs/prd/pryx-mesh-design.md` (v1+)

This document clarifies what is visible and manageable from the web dashboard when users install/integrate Pryx from a device via CLI/installer, and how auth + synchronization work end-to-end.

## 0) Scope and Current Repo State

### What exists today in this repository

- A **local gateway runtime** (`pryx-core`) exposing HTTP + WebSocket endpoints on localhost (no auth middleware yet).
- A local **event bus** streaming runtime events over WebSocket (`/ws`) for “multi-surface” sync on the same machine.
- Integrations implemented today run **on the device** (e.g., Telegram polling, local webhook server/outgoing webhook).

### What “web dashboard” means in Pryx (two dashboards)

To match the “sovereign-by-default” principle, Pryx intentionally splits dashboard responsibilities:

1. **Local Dashboard (localhost)**: full-fidelity configuration and observability for the machine where `pryx-core` runs.  
2. **Cloud Web Dashboard (pryx.dev / web platform)**: account + device list + sanitized telemetry + (optional) encrypted config backup and multi-device control via Mesh.

As a result: **not all configuration details are visible in the cloud dashboard by default**. The cloud dashboard shows a curated subset and only expands when the user explicitly enables sync/backup features.

## 1) Authentication Flow (CLI / Installation)

### 1.1 Installation (one-liner)

1. User runs installer (e.g., `curl | sh`) and gets `pryx` + `pryx-core`.
2. Installer (or first run) creates a local **install record**:
   - `install_id`: random UUID
   - `device_name`: user-friendly default (e.g., “MacBook Pro”)
   - `platform`: OS/arch
   - `runtime_version`: Pryx version
3. `pryx-core` starts (service mode optional).

### 1.2 Login: OAuth Device Authorization Grant (RFC 8628)

The login is a **public-client** flow designed for CLI environments:

1. CLI requests a device code from Edge:
   - `POST /oauth/device/code`
   - Response: `device_code`, `user_code`, `verification_uri`, `verification_uri_complete`, `expires_in`, `interval`
2. CLI prints `user_code` and opens `verification_uri_complete` when possible.
3. User signs in on the cloud web dashboard and approves the device code.
4. CLI polls Edge until approved:
   - `POST /oauth/token` with `grant_type=urn:ietf:params:oauth:grant-type:device_code`
5. Edge issues tokens:
   - `access_token` (short-lived; bearer; scoped)
   - `refresh_token` (long-lived; rotates)
   - Optional: `id_token` (OIDC) if we use an OIDC provider
6. CLI stores credentials securely:
   - Refresh token stored in **OS keychain** when available
   - Access token kept in memory and refreshed as needed

### 1.3 Validation and session binding

On the server side:

- Access tokens are validated (JWT signature + expiry) or introspected (opaque token), and are bound to:
  - `user_id`
  - `device_id` / `install_id` (registered at first login)
  - `scopes` (telemetry.write, devices.read, devices.manage, backups.read/write, etc.)

On the client side:

- `pryx-core` treats the cloud as optional: local usage works without login, but cloud dashboard features (device list, telemetry, backups, remote control) require it.

## 2) Data Synchronization (Device → Dashboard)

There are three distinct data planes.

### 2.1 Local plane (same machine): WebSocket event stream

Used for localhost dashboard and local UI clients:

- Transport: WebSocket
- Endpoint: `GET /ws`
- Data format: JSON events (versioned)
- Semantics: “best effort real-time”; clients can filter by `session_id` and `event`

### 2.2 Cloud telemetry plane (sanitized): OTLP + minimal custom events

Used for cloud observability without sending prompts/PII:

- Transport: HTTPS
- Primary format: **OTLP/HTTP** (protobuf payloads)
- Endpoint (conceptual): `POST /otlp/v1/traces` and/or `/otlp/v1/logs`
- Data model:
  - OpenTelemetry resource attributes contain `install_id`, `device_id`, `runtime_version`, `os`, `arch`
  - Spans/log records include tool durations, integration health signals, error categories
- Redaction:
  - A worker-side redaction layer strips/obfuscates PII and forbids prompt contents

### 2.3 Cloud state plane (opt-in): device metadata + configuration snapshots

Used to make the cloud dashboard “manageable” without leaking secrets:

- Transport: HTTPS (REST) and/or WebSocket (Mesh Coordinator in v1+)
- Data format: JSON (versioned; schema lives in `packages/protocol` once implemented)
- What is transmitted:
  - Device metadata (always): `device_name`, `first_seen_at`, `last_seen_at`, versions, health
  - Integration metadata (opt-in by default, but recommended): list of integrations and status
  - Config snapshots (opt-in): non-secret configuration and “secret present” flags
  - Encrypted backups (opt-in): encrypted blobs (cloud cannot decrypt)

## 3) Dashboard Visibility (What Users See)

### 3.1 Cloud web dashboard (what is shown)

Cloud dashboard shows only what the user must know to manage their fleet and trust the system:

- **Devices**
  - Installation timestamp (first seen)
  - Last seen / heartbeat status (online/offline)
  - Device name, OS/arch, Pryx version
  - Stable device identifier (non-PII): `device_id` / `install_id`
- **Health**
  - `pryx doctor` summary: Healthy / Degraded / Unhealthy
  - Current blockers (e.g., “Telegram disconnected”, “Webhook failing”, “Provider key missing”)
- **Integrations**
  - Connected integrations list (Telegram, webhook endpoints, MCP servers as metadata only)
  - Status (connected / degraded / disconnected)
  - Last successful message / last error timestamp
  - Retry/backoff state (high-level)
- **Observability (sanitized)**
  - Error summaries by category (network/auth/provider/integration/tool policy)
  - Cost summary (token counts, estimated cost) without prompt contents
  - Reliability metrics (p95 tool duration, failed calls)
- **Audit (high level)**
  - “What changed?” timeline for config/integration state (not raw secrets)
  - Token revocations and device unlinks

Cloud dashboard does not display:

- Prompt text, file contents, full chat transcripts (unless user explicitly exports)
- Raw environment variables (e.g., `OPENAI_API_KEY`) or secret values
- Full local logs by default (only sanitized error summaries + correlation IDs)

### 3.2 Local dashboard (what is shown)

Local dashboard can show full-fidelity details because it is running on the user’s machine:

- Full configuration (including where secrets are stored, but never auto-revealing secret values)
- Full session timeline and tool call details
- Local logs and integration traces (with sensitive content masked by default)
- “Copy diagnostic bundle” action for support/export

## 4) Management Capabilities (What Users Can Do)

### 4.1 Cloud web dashboard actions

- Device management
  - Rename device
  - Unlink device (revokes refresh tokens, stops telemetry from that device)
  - Rotate device credentials (forces re-login)
- Integration management (safe operations)
  - Disconnect integration (revokes cloud-managed tokens; device stops using it)
  - Trigger re-check / re-sync (requests device to run health checks)
  - Rotate webhook secrets (cloud-hosted integrations) and push new desired state to device
- Telemetry and privacy controls
  - Toggle telemetry export and configure retention level
  - Download/export user data (GDPR-style export)
- Audit & activity
  - View config change history (metadata), device activity history, and incident history

### 4.2 Local dashboard actions (full control)

- Modify runtime config and integration config (including local-only integrations)
- Manage MCP servers and skills on the device
- View full logs, tool outputs, and debugging details
- Run `doctor` and apply suggested remediations

### 4.3 Remote “configuration changes” are desired-state, not direct mutation

To avoid accidental breakage and to preserve sovereignty:

- Cloud dashboard proposes a **desired state** update
- Device validates it locally, applies it, and reports status back
- If the device requires approvals for changes, the user must approve on at least one trusted surface

## 5) Security & Privacy

### 5.1 Data in transit

- All cloud communication uses HTTPS (TLS).
- WebSockets for Mesh use secure, persistent connections (wss).
- Telemetry export is authenticated and scoped; anonymous telemetry is not allowed for “managed devices”.

### 5.2 Data at rest

- On device:
  - Secrets stored in OS keychain when available
  - Local state stored in SQLite (permissions locked down)
- In cloud:
  - Device metadata stored per account (tenant isolation)
  - Telemetry stored with strict retention and redaction rules
  - Optional config backups stored as **encrypted blobs** (cloud cannot decrypt)

### 5.3 Data minimization and isolation

- The cloud dashboard only receives:
  - sanitized telemetry
  - metadata necessary for device/integration health and management
  - user-approved exports/backups
- Each request is authorized and scoped to the authenticated user/org.

### 5.4 Privacy compliance posture (baseline)

- Telemetry is user-controlled (toggle + retention levels).
- No prompt contents or raw user data is ingested by default.
- Provide data export and deletion flows (GDPR-aligned baseline).

## 6) Error Handling & Monitoring

### 6.1 Error capture (device-side)

Errors are classified before export:

- `auth`: login/token invalid, revoked device
- `network`: timeouts, DNS, TLS failures
- `integration`: Telegram/webhook failures, MCP server unavailable
- `provider`: LLM provider errors, rate limits, quota exceeded
- `policy`: user denied approval, policy blocked tool call
- `runtime`: panics/crashes, DB issues, schema migration errors

Each error report includes:

- `error_code` (stable, user-facing)
- `severity` (info/warn/error/critical)
- `timestamp`, `device_id`, `session_id` (if applicable)
- `correlation_id` to map cloud summary ↔ local logs
- redacted context (no secrets, no prompts)

### 6.2 Reporting and presentation (cloud dashboard)

Cloud dashboard shows:

- Top issues with clear “what to do next” remediation steps
- Integration status and last error reason
- Trend charts (error rate, disconnect frequency)
- A “download local diagnostics” prompt when deeper info is required

### 6.3 Reporting and presentation (local dashboard)

Local dashboard shows:

- Full error details + stack traces (where safe)
- Raw logs and recent events
- One-click “copy troubleshooting bundle” for support

## 7) UX Defaults (Best Experience Without Leaking Data)

- Cloud dashboard defaults to “must-know” summaries and guided fixes.
- Local dashboard is the place for deep debugging and full configuration.
- Telemetry export defaults ON, but:
  - is sanitized
  - is user-togglable
  - has clear retention controls
