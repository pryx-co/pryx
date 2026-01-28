# Telegram Bot Integration — Operational Requirements & Tasks

## Current Repo Status (Reality Check)

- A Telegram channel adapter exists in the local runtime (polling + send), but it is not wired into runtime startup/config or any public API, so it is not end-to-end operational for real users yet.
- There is no Pryx Edge/Worker implementation of Telegram webhooks in this repo yet.

## Core Question: Can the bot work without user-side core installation?

Yes, **if** Pryx hosts the Telegram webhook receiver and message-processing logic (“cloud-hosted webhook mode”), and the user either:

- stores their model provider key (e.g., OpenRouter key for GLM-4.x/GLM-4.7) in Pryx cloud (encrypted at rest), or
- uses a Pryx-managed gateway/plan that does not require the user’s model key.

If the user does **not** want their model key stored server-side, then **no**: Telegram cannot attach a per-user Authorization header to webhooks, so the user must run a local runtime (“device-hosted polling mode”) that holds the model key and processes updates locally.

## Execution Models

### A) Cloud-Hosted Webhook Mode (Recommended MVP)

**Who hosts what**
- Telegram: delivers updates to Pryx via webhook.
- Pryx (us): stores bot token + webhook secret; receives updates; calls AI model with the user’s key; sends replies via Telegram API.
- User: creates bot via BotFather; provides bot token; links a chat; provides model provider key.

**Pros**
- No installation required for end users (pure web onboarding).
- Works reliably on mobile without “always-on” user device.

**Constraints**
- Must store bot tokens and (optionally) model provider keys securely.
- Must implement strict isolation, redaction, and per-user limits.

### B) Device-Hosted Polling Mode (Sovereignty/Privacy)

**Who hosts what**
- Telegram: updates pulled by user’s runtime via getUpdates (polling).
- User device: runs pryx-core, holds model provider key locally, processes updates locally.
- Pryx (us): optional coordination (Mesh), optional auth/device management.

**Pros**
- User’s model key never leaves their device.

**Constraints**
- Requires a continuously running process on at least one user device.
- Must ensure only one active poller per bot token (avoid duplicate message handling).

## User Integration Flow (End-to-End)

### 1) Registration + Provider Key (GLM-4.x / GLM-4.7)

1. User registers on Pryx Web Dashboard.
2. User adds a model provider:
   - Recommended path for GLM models: OpenRouter (since it exposes OpenAI-compatible APIs).
3. User saves provider credentials:
   - Key is encrypted at rest; never logged; only decrypted inside request handling for that user.

### 2) Telegram Bot Setup (BotFather)

1. User creates a Telegram bot via BotFather.
2. User receives a bot token.

### 3) Connect Telegram Bot to Pryx (Web Dashboard or CLI)

1. User pastes bot token into Pryx onboarding wizard.
2. Pryx verifies token (Telegram `getMe`).
3. Pryx sets webhook (Telegram `setWebhook`) to Pryx Edge endpoint and configures `secret_token`.
4. Pryx presents a “Link chat” step:
   - Either show a deep-link URL, or instruct: “Send `/start <link_code>` to the bot”.

### 4) Runtime Operation

1. Telegram sends updates to Pryx webhook.
2. Pryx validates:
   - webhook secret header
   - update_id dedupe
   - chat allowlist / link mapping
3. Pryx processes message:
   - policy checks
   - calls AI model using the user’s provider key (GLM-4.x / GLM-4.7)
4. Pryx replies via Telegram API.

## Pryx API Requirements (Endpoints, Auth, Formats)

### Auth Model

- **Management endpoints**: require Pryx user auth (web session or OAuth device-flow token).
- **Webhook endpoint**: authenticated via Telegram `X-Telegram-Bot-Api-Secret-Token` matching per-bot stored secret.
- **Never** authenticate Telegram webhook via the user’s model provider key (Telegram can’t attach it safely).

### 1) Create Telegram Integration

`POST /api/v1/integrations/telegram/bots`

Headers:
- `Authorization: Bearer <pryx_user_access_token>`

Body:
```json
{
  "name": "My Personal Bot",
  "bot_token": "123456:ABC-DEF...",
  "default_agent_id": "agent_default",
  "default_workspace_id": "ws_default",
  "mode": "cloud_webhook"
}
```

Response:
```json
{
  "bot_id": "bot_01H...",
  "telegram_bot_id": 123456789,
  "telegram_username": "my_pryx_bot",
  "webhook_url": "https://api.pryx.dev/api/v1/integrations/telegram/webhook/bot_01H...",
  "webhook_secret_preview": "****",
  "status": "created"
}
```

### 2) (Re)Sync Webhook

`POST /api/v1/integrations/telegram/bots/{bot_id}/sync-webhook`

Headers:
- `Authorization: Bearer <pryx_user_access_token>`

Response:
```json
{ "status": "ok", "telegram": { "webhook_set": true } }
```

### 3) Link a Telegram Chat

`POST /api/v1/integrations/telegram/bots/{bot_id}/link-code`

Headers:
- `Authorization: Bearer <pryx_user_access_token>`

Response:
```json
{ "link_code": "LK-7R3J9S", "expires_in_seconds": 600 }
```

Chat linking is completed when the bot receives `/start LK-7R3J9S` via webhook and Pryx stores:
- `user_id` ↔ `bot_id` ↔ `chat_id`
- optional allowlist by Telegram user id / username

### 4) Telegram Webhook Receiver (Public)

`POST /api/v1/integrations/telegram/webhook/{bot_id}`

Headers:
- `X-Telegram-Bot-Api-Secret-Token: <per-bot-secret>`

Body: Telegram Update JSON (as sent by Telegram).

Response:
- `200 OK` quickly (ack), processing can be async; retries handled by Telegram.

## Security, Isolation, and Reliability Requirements

- **Key isolation**: per-user storage + per-bot access control; server never mixes keys across tenants.
- **Secret handling**:
  - never log bot tokens or model keys
  - redact secrets from error messages and traces
  - rotate per-bot webhook secret + bot token revoke flow
- **Rate limiting**:
  - per-user and per-bot limits for inbound updates
  - per-user limits for outbound model calls
- **Idempotency**:
  - dedupe by `update_id` (Telegram) and message id per chat
- **Timeouts + retries**:
  - model call timeouts, exponential backoff, circuit breaker per provider
  - Telegram send retry on 429/5xx with respect to `retry_after`
- **Storage**:
  - record minimal metadata for debugging (timestamp, update_id, error class) without message bodies by default
  - optional opt-in debugging that stores message bodies with retention window

## Edge Cases / Product Decisions to Encode

- Multiple bots per user: supported (each has its own token, secret, routing).
- Multiple chats per bot: supported via chat linking; allowlist required.
- Bot paused: webhook stays set but processing returns quick “paused” behavior (optional) or ignores messages.
- Key revocation: immediately disables model calls; bot replies with actionable error message.
- Telegram API changes: versioned adapter layer; contract tests against Telegram schema.
- AI provider outages: graceful fallback message; retries; clear user-visible status in dashboard.

## Actionable Technical Subtasks (Tickets)

### P0 — Make Telegram functional (Cloud Webhook MVP)
- [ ] Design DB schema for `telegram_bots`, `telegram_chat_links`, `secrets`, `update_dedup`.
- [ ] Implement management API endpoints (create bot, sync webhook, pause/resume, rotate secret, delete).
- [ ] Implement webhook receiver with secret validation and update_id dedupe.
- [ ] Implement chat linking (/start link code) + allowlist enforcement.
- [ ] Implement message routing to agent/workspace and reply formatting.
- [ ] Implement per-user model key resolution (OpenRouter for GLM-4.x/GLM-4.7) + redaction.
- [ ] Add rate limits and retry policies (Telegram send + model call).
- [ ] Add audit events and user-visible integration status page.

### P1 — Device-Hosted Polling Mode (Optional)
- [ ] Implement runtime config loading for channels (e.g., `.pryx/channels.json`) and register Telegram channel at startup.
- [ ] Implement single-active-host coordination using Mesh coordinator.
- [ ] Implement “handoff”/failover policy for bot host device.

