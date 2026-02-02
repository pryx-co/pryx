# OAuth Implementation Architecture

## Overview

Pryx requires **TWO separate OAuth implementations**:

1. **Device Flow OAuth** (RFC 8628) - Device registration with Pryx Cloud
2. **Provider OAuth** (OAuth 2.0) - AI provider authentication (Google, etc.)

---

## 1. Device Flow OAuth (Pryx Cloud)

**Purpose**: Register local device with user's Pryx account

**Flow**:
```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌────────────┐
│   Runtime   │────▶│  Edge Worker │────▶│  Web (Astro)│────▶│   User     │
│  (pryx-core)│◄────│(Cloudflare)  │◄────│  Dashboard  │◄────│  Browser   │
└─────────────┘     └──────────────┘     └─────────────┘     └────────────┘
      │                      │                   │
      │ 1. Request device code│                   │
      │──────────────────────▶│                   │
      │ 2. Return device_code │                   │
      │◄──────────────────────│                   │
      │ 3. Open browser       │                   │
      │    /auth/device       │                   │
      │──────────────────────▶│───────────────────▶
      │ 4. Poll for token     │                   │
      │──────────────────────▶│                   │
      │ 5. User logs in       │                   │
      │                       │◄──────────────────│
      │ 6. Authorize device   │                   │
      │                       │───────────────────▶
      │ 7. Return tokens      │                   │
      │◄──────────────────────│                   │
```

**Endpoints Required**:
- `POST /oauth/device/code` - Request device code
- `POST /oauth/token` - Exchange device code for tokens (polling)

**Runtime Components**:
- `internal/auth/device/` - Device flow client
- `cmd/pryx-core login` - CLI command
- Local callback server (for development/testing)

---

## 2. Provider OAuth (AI Providers)

**Purpose**: Connect to AI providers using OAuth instead of API keys

**Supported Providers**:
- Google (Gemini API)
- Future: Azure OpenAI, etc.

**Flow**:
```
┌─────────────┐     ┌──────────────────┐     ┌────────────┐
│   Runtime   │────▶│  Google OAuth    │────▶│   User     │
│  (pryx-core)│◄────│  (accounts.google)│◄────│  Browser   │
└──────────────┘     └──────────────────┘     └────────────┘
      │
      │ 1. Start local callback server
      │    (localhost:random-port)
      │
      │ 2. Open browser with OAuth URL
      │    scope: https://www.googleapis.com/auth/generative-language
      │
      │ 3. User grants permission
      │
      │ 4. Callback to localhost
      │    with authorization code
      │
      │ 5. Exchange code for tokens
      │
      │ 6. Store in keychain:
      │    - access_token
      │    - refresh_token
      │    - expires_at
```

**Runtime Components**:
- `internal/auth/provider/` - Provider OAuth client
- `internal/auth/callback/` - Local HTTP callback server
- PKCE support for security
- Token refresh before expiry (5 min buffer)

---

## Implementation Phases

### Phase 1: Runtime-Side OAuth (Can work independently)
- [x] Device flow client (polling, state management)
- [ ] Provider OAuth client
- [ ] Local callback server
- [ ] Token storage/refresh
- [ ] CLI commands

### Phase 2: Edge Worker (Required for cloud sync)
- [ ] Device code endpoint
- [ ] Token exchange endpoint
- [ ] User authorization UI
- [ ] Device linking logic

### Phase 3: Integration
- [ ] TUI OAuth flows
- [ ] Web dashboard OAuth UI
- [ ] Multi-device sync

---

## Storage Schema

**Keychain Entries**:
```
oauth_token_pryx           - Pryx cloud access token
oauth_refresh_pryx         - Pryx cloud refresh token
oauth_expires_pryx         - Token expiry timestamp
oauth_token_google         - Google access token
oauth_refresh_google       - Google refresh token
oauth_expires_google       - Google token expiry
device_id                  - Unique device identifier
```

---

## Security Considerations

1. **PKCE** (Proof Key for Code Exchange) - Required for provider OAuth
2. **State parameter** - Prevent CSRF attacks
3. **Token storage** - OS keychain only, never plaintext
4. **Short-lived codes** - Device codes expire in 5 minutes
5. **Localhost binding** - Callback server binds to 127.0.0.1 only
6. **TLS verification** - All HTTPS connections verified

---

## CLI Commands

```bash
# Device Flow (Pryx Cloud)
pryx auth login              # Initiate device flow
pryx auth logout             # Remove tokens
pryx auth status             # Show auth status

# Provider OAuth
pryx provider connect google    # OAuth connect to Google
pryx provider disconnect google # Remove Google tokens
```

---

## API Endpoints

### Device Flow (Edge Worker)

**POST /oauth/device/code**
```json
Request:
{
  "client_id": "pryx-cli",
  "scope": "device sync telemetry"
}

Response:
{
  "device_code": "GmRhmhcxhwAzkoEqiMEg_DnyEysO...",
  "user_code": "WDJBMJHT",
  "verification_uri": "https://pryx.dev/auth/device",
  "expires_in": 300,
  "interval": 5
}
```

**POST /oauth/token**
```json
Request:
{
  "grant_type": "urn:ietf:params:oauth:grant-type:device_code",
  "device_code": "GmRhmhcxhwAzkoEqiMEg_DnyEysO...",
  "client_id": "pryx-cli"
}

Response (success):
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6..."
}

Response (pending):
{
  "error": "authorization_pending"
}
```
