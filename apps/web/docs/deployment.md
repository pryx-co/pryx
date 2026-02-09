# Web Deployment (Cloudflare)

This document defines the reproducible deployment flow for `apps/web`.

## Installer Source of Truth

- Canonical installer: repository root `install.sh`
- `scripts/install.sh` delegates to canonical installer
- Web `/install` endpoint imports canonical installer directly

Verify this contract at any time:

```bash
./scripts/check-installer-sync.sh
```

## 1) Prepare

```bash
bun install
bun run build
```

## 2) Configure Secrets (per environment)

Set secrets in Cloudflare instead of committing them into `wrangler.toml`.

```bash
wrangler secret put ADMIN_API_KEY --env staging
wrangler secret put LOCALHOST_ADMIN_KEY --env staging

wrangler secret put ADMIN_API_KEY --env production
```

`LOCALHOST_ADMIN_KEY` is intended for non-production environments only.

## 3) Deploy

From `apps/web`:

```bash
wrangler d1 migrations apply pryx-db-staging --env staging
wrangler d1 migrations apply pryx-db --env production

wrangler deploy --env staging
wrangler deploy --env production
```

Migration files live in `apps/web/migrations/`.

## 4) Post-Deploy Checks

Run endpoint checks against the deployed host:

```bash
./scripts/post-deploy-checks.sh https://pryx.dev
```

This verifies:
- `/` returns a successful response
- `/install` returns shell script content type and installer body
- `/api` returns API health payload
- `/api/admin/health` is reachable (401/403/200 accepted, but not 404)
