# Repository Structure Assessment (Polyglot Monorepo)

This repo is a polyglot monorepo for Pryx:

- Host (Rust/Tauri)
- Runtime (Go)
- TUI (TypeScript + Solid + OpenTUI)
- Planned: Cloudflare Worker (edge API) + a single web frontend (Astro/React)

## Current Layout (Observed)

Top-level modules:

- `apps/`: product surfaces (host/runtime/tui/web)
- `docs/`: PRD and supporting documents
- `packages/`: planned shared cross-surface libraries (protocol/config)
- `workers/`: planned Cloudflare Worker project(s)
- `deploy/`: planned deployment assets (compose + edge)
- `Makefile`, `.github/workflows/ci.yml`: repo orchestration

## Assessment

### Separation of Concerns

Strengths:

- Clear component boundaries at the directory level (`host/`, `runtime/`, `tui/`).
- Runtime is isolated as its own Go module (`runtime/go.mod`), which scales well as runtime grows.

Issues / risks:

- Docs and root metadata reference web/edge concepts that do not exist yet. This creates confusion for contributors and makes CI/build expectations ambiguous.
- Shared cross-surface concepts (event schemas, protocols, types) are not yet centralized. As web/edge surfaces arrive, drift is likely without a shared “protocol” package.

### Dependency Management

Strengths:

- Rust, Go, and TUI dependencies are scoped to their components.

Issues / risks:

- JavaScript tooling can drift if entrypoints use different runners. Prefer Bun everywhere (Makefile + pre-commit via bunx) to avoid version skew.
- OpenTUI’s upstream build requires Zig when building OpenTUI itself. Consumers typically only need Bun, but contributors working on OpenTUI forks or vendoring should plan for Zig.

### Build / CI Organization

Strengths:

- `make build/test/lint` provide a single conceptual entrypoint.
- CI runs platform builds and separate test jobs.

Issues / risks:

- CI should prefer `make lint/test/build` as the single source of truth to reduce drift.
- Artifact and coverage paths must match what the build/test commands produce.

### Deployment Workflow Fit

Planned architecture suggests:

- Go runtime runs locally / on servers
- Web apps and Cloudflare Workers deploy to edge/cloud infrastructure
- Installer and update distribution also lives in “web/edge” surface area

This repo currently lacks `deploy/` or `infra/` scaffolding to standardize those flows.

## Recommended Long-Term Structure

If Pryx stays a monorepo, the most scalable structure is “apps + packages”, plus a single Cloudflare Worker project and a single web frontend.

This matches a simpler product shape:

- One worker (edge API) that exposes multiple routes: auth, telemetry ingest, installer/updates
- One web frontend that serves both user dashboard and superadmin/management dashboard (role-gated/RBAC), and can be deployed to (temporary URL using cloudflare workers in future maybe `pryx.dev` or `pryxbot.com` or `any domain` and/or served on localhost for onboarding/config
- One local “gateway” runtime that all UIs and channels talk to (TUI is optional)

```
/
  apps/
    host/                 # (move from host/) Rust/Tauri desktop app
    runtime/              # (move from runtime/) Go pryx-core service
    tui/                  # (move from tui/) Terminal UI
    web/                  # Astro/React app: user dashboard + superadmin dashboard
  workers/
    edge/                 # single Cloudflare Worker (routing + bindings)
  packages/
    protocol/             # shared event schema, types, versioning
    config/               # shared config conventions (env keys, defaults)
    ui-kit/               # shared UI primitives (optional)
  deploy/
    compose/              # docker-compose stacks for dev/prod parity
    edge/                 # wrangler configs, deploy scripts
  docs/
  scripts/
```

Notes:

- Keep each “app” self-contained with its own package/module manager.
- Put shared cross-surface artifacts in `packages/` (especially the event schema and protocol types).
- Keep all deployment assets in `deploy/` so operational workflows are discoverable and consistent.

Distribution note:

- The one-liner installer should ship a cohesive “Pryx bundle”: `pryx` CLI entrypoint + `pryx-core` runtime binary (and optionally the `host/` desktop wrapper where supported). Repo structure should optimize for building and packaging these together.

### Local Gateway Pattern (OpenCode-like)

Architecture pattern (similar to OpenCode’s client/server approach):

- `apps/runtime/` (`pryx-core`) is the always-on localhost backend (“gateway”) that owns sessions, tools, policy, routing, and integrations.
- `apps/tui/` is one client; it can be absent (headless mode).
- External UIs (Telegram/Discord/Slack/webhooks) act as additional clients via channel adapters in the runtime.
- `apps/host/` is optional and focuses on lifecycle (tray, deep links, native dialogs, auto-update), not core logic.
- `workers/edge/` is the internet-facing edge backend for cross-device auth and cloud concerns (telemetry ingest, update metadata, optional AI gateway/model routing).

### Core Language Choice (Go vs Rust)

Both options can be future-proof if the system is designed around stable boundaries (HTTP/WebSocket API + versioned `packages/protocol` schemas) and treats integrations as modular adapters.

When Go is the runtime core:

- Fast iteration for networking-heavy gateways (HTTP/WS, JSON, concurrency, integrations).
- Simple operational story (single static-ish binary, strong standard library, straightforward profiling/debugging).
- Plays well with “many adapters” architecture (channels, webhooks, MCP clients, telemetry exporters).

When Rust is the runtime core:

- Strong memory safety + predictable performance for long-running agents.
- Tight control of sandboxing and low-level primitives (process mgmt, resource limits, IPC).
- Shared language with Tauri host, which can reduce cognitive load for core systems engineers.

Recommended default for Pryx (pragmatic + durable):

- Keep Rust primarily for lifecycle/UI host responsibilities and any security-sensitive primitives that benefit from tight control.
- Keep the always-on gateway runtime in Go as the integration hub.
- Make the core swappable by enforcing a strict gateway API contract and keeping “clients” (TUI/web/channels) decoupled from runtime internals.

### Edge + Web Responsibilities (Single Worker + Single Frontend)

Recommended ownership split:

- `workers/edge/` is the only public API entrypoint (auth + telemetry + installer/update metadata).
- `apps/web/` is the only browser surface (user dashboard + superadmin dashboard).
- `apps/runtime/` stays local-first and should not require the web to function beyond optional auth/sync.

Within `workers/edge/`, keep code modular even if it is “one worker”:

- Route modules by domain (`auth/*`, `telemetry/*`, `installer/*`)
- Shared libraries for signing, token validation, role checks, and schema validation

### Install + Login + Onboarding Flow (Target UX)

Target UX mirrors the best parts of Moltbot’s “wizard + doctor” approach, but adapted for Pryx:

1. One-liner install (curl | sh) installs the CLI + runtime and sets up a local service.
2. `pryx auth login` (or `pryx login`) opens a browser to the Pryx web app and completes a device-code flow.
3. After login, `pryx onboard` (TUI-first, web-assisted) configures model/provider, integrations, MCP servers, skills/plugins, channels, etc.
4. `pryx doctor` continuously checks for risky/misconfigured settings and gives actionable fixes.

Dashboard contract:

- Pryx has two dashboards by design: a full-fidelity localhost dashboard for the device, and a cloud web dashboard that shows a curated, privacy-preserving subset unless the user explicitly enables sync/backup.
- Details: see [auth-sync-dashboard.md](./auth-sync-dashboard.md).

Implementation notes (behavioral contract, not a mandate on tech):

- `pryx auth login` should be a device authorization flow:
  - Use OAuth Device Flow (RFC 8628) semantics for device codes + polling
  - CLI requests a short device code from `workers/edge` (auth route)
  - CLI opens `apps/web` (deployed at pryx.dev) to a URL like `/auth/device` and the user logs in (or is prompted to create an account)
  - User enters/confirms the code; edge issues tokens
  - CLI polls until success, then stores refresh token securely (keychain where available)
- On success, CLI returns to TUI and continues configuration without forcing the user to copy secrets manually.

## If Restructuring Is Not Desired Yet

If you want to keep the current top-level layout for now, the most important actions are:

1. Make docs and root metadata match reality (keep web/edge as planned until implemented).
2. Centralize shared “protocol” types before adding multiple web apps/workers.
3. Standardize JS toolchain choices (Bun vs Node) to reduce drift.
4. Standardize Docker + Make workflows before shipping server/edge components.
