![Pryx](https://github.com/irfndi/pryx/raw/develop/v1-production-ready/.github/assets/pryx-logo.png)

# Pryx

[![CI](https://github.com/irfndi/pryx/actions/workflows/ci.yml/badge.svg)](https://github.com/irfndi/pryx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/irfndi/pryx/branch/main/graph/badge.svg)](https://codecov.io/gh/irfndi/pryx)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/irfndi/pryx/blob/main/LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0-brightgreen)](https://github.com/irfndi/pryx/releases/tag/v1.0.0)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](https://github.com/irfndi/pryx/releases)

> **Sovereign AI agent with local-first control center**

Pryx is a second-generation autonomous agent platform: local-first, sovereign, and secure. It is a unified control center available via TUI, CLI, local web, channels, and Pryx Cloud, with explicit approvals for sensitive actions and clear observability without sacrificing data sovereignty or privacy.

---

## 🚀 Quick Start

### Installation (One-Liners)

Canonical installer source in this repository: `install.sh`

**macOS (Intel + Apple Silicon)**
```bash
brew install irfndi/pryx/pryx
```

**Linux (Ubuntu/Debian)**
```bash
curl -fsSL https://get.pryx.ai/install.sh | bash
```

**Linux (Fedora/RHEL)**
```bash
dnf install https://github.com/irfndi/pryx/releases/download/v1.0.0/pryx-1.0.0-1.x86_64.rpm
```

**Windows**
```powershell
winget install pryx
```

> **Or** download the binary from [Releases](https://github.com/irfndi/pryx/releases)

### First 5 Minutes

1. **Start Pryx**
   ```bash
   pryx
   ```
   This launches the Terminal UI (TUI) and runtime server automatically.

2. **Complete Onboarding**
   The onboarding wizard will guide you through:
   - Setting a master password (for vault encryption)
   - Adding your first AI provider
   - Creating your first agent

3. **Start Chatting**
   Press `/` to open command palette, select "Chat", and start interacting!

---

## ✨ Features

### 🎯 Multi-Channel Integration
- **Telegram** - Run your agent as a Telegram bot
- **Discord** - Deploy as a Discord bot with slash commands
- **Slack** - Connect to Slack channels and DMs
- **Webhooks** - Integrate with any HTTP endpoint

### Any MCP & Skills
- integration with any MCP and Skills as you needs

### 🔌 84+ AI Providers
Dynamic integration via [models.dev](https://models.dev) supporting:
- OpenAI, Anthropic, Google, xAI
- OpenRouter, Groq, Mistral, Cohere
- And 76+ more providers

Full list: `pryx provider list --available`

### 🔒 Sovereign Security
- **Vault with Argon2id** - Military-grade password derivation
- **OS Keychain Storage** - Secrets never stored in plaintext
- **Scope-Based Access Control** - Fine-grained permission management
- **Human-in-the-Loop Approvals** - Explicit approval for sensitive operations
- **Comprehensive Audit Logging** - Every action traceable

### 🎛️ Rich Terminal UI via CLI - TUI - Desktop Apps - Any Channel You Want
- **Provider Management** - Add, configure, and test providers of any compatible AI models
- **Channel Configuration** - Set up Telegram, Discord, Slack, Webhooks, and more
- **MCP Tool Management** - Discover and manage Model Context Protocol servers
- **Skill Management** - Add, configure, and test skills of any compatible MCP tools
- **Session Explorer** - Browse and resume conversations
- **Settings & Configuration** - All settings in one place

### 🤖 Agent Capabilities
- **Agent Spawning** - Create sub-agents for parallel task execution
- **Policy Engine** - Define approval rules for sensitive operations
- **Skills System** - Extensible capabilities via MCP tools
- **Natural Language Parser** - Intent recognition and command parsing
- **Context Management** - Maintain conversation context for multi-turn interactions
- **Cron Job Scheduler** - Schedule tasks with cron expressions

### 📊 Observability
- **Cost Tracking** - Monitor token usage and costs across all providers
- **Session Timeline** - Complete trace of conversations, tool calls, and approvals
- **Performance Profiling** - Memory and CPU usage monitoring
- **OTLP Telemetry** - Export to OpenTelemetry backends (optional)

### 🌐 Multi-Device Coordination
- **Pryx Mesh** - Secure sync across devices
- **Device Pairing** - QR code or 6-digit code pairing
- **WebSocket Mesh** - Real-time coordination without cloud dependency

### Memory
- **Long-Term Memory** - Store and retrieve information over sessions
- **Short-Term Memory** - Contextual understanding for current conversations

---

## 📖 Channel Integration Guides

### Telegram

1. **Create a Bot**
   - Chat with [@BotFather](https://t.me/botfather) on Telegram
   - Send `/newbot` and follow the instructions
   - Copy the bot token

2. **Configure in Pryx**
   ```bash
   pryx channel add telegram
   ```
   Enter your bot token when prompted.

3. **Enable the Channel**
   ```bash
   pryx channel enable telegram
   ```

4. **Start Chatting**
   Open your bot in Telegram and start the conversation!

### Discord

1. **Create a Discord Application**
   - Go to [Discord Developer Portal](https://discord.com/developers/applications)
   - Click "New Application" and name it (e.g., "Pryx Bot")
   - Create a bot in the "Bot" tab and copy the token

2. **Configure Bot Permissions**
   - Enable "Server Members Intent" and "Message Content Intent"
   - Save changes

3. **Invite Bot to Server**
   - Go to "OAuth2" → "URL Generator"
   - Select scopes: `bot`, `applications.commands`
   - Select bot permissions: Read Messages, Send Messages, Embed Links
   - Copy the generated URL and open it in a browser
   - Invite the bot to your server

4. **Configure in Pryx**
   ```bash
   pryx channel add discord
   ```
   Enter your bot token when prompted.

5. **Start Chatting**
   In your Discord server, use `/chat <your message>` to interact!

### Slack

1. **Create a Slack App**
   - Go to [Slack API](https://api.slack.com/apps)
   - Click "Create New App" → "From scratch"
   - Name your app and select your workspace

2. **Configure Bot Permissions**
   - Go to "Bot" → "Permissions"
   - Add scopes: `chat:write`, `channels:read`, `im:read`, `im:write`, `groups:read`, `groups:write`
   - Install the app to your workspace and copy the bot token

3. **Enable Events**
   - Go to "Event Subscriptions" → "Enable Events"
   - Add workspace URL (use `pryx channel test slack` to get the URL)
   - Subscribe to events: `message.channels`, `message.groups`, `message.im`

4. **Configure in Pryx**
   ```bash
   pryx channel add slack
   ```
   Enter your bot token when prompted.

5. **Start Chatting**
   Invite the bot to channels or DM the bot directly!

---

## 🔌 Provider Configuration

### Quick Setup

**OpenAI**
```bash
pryx provider add openai
pryx provider set-key openai
# Enter your API key when prompted
pryx provider use openai
```

**Anthropic**
```bash
pryx provider add anthropic
pryx provider set-key anthropic
# Enter your API key when prompted
pryx provider use anthropic
```

**Google AI (Gemini)**
```bash
# Option 1: API Key
pryx provider add google
pryx provider set-key google
# Enter your API key when prompted

# Option 2: OAuth
pryx provider oauth google
# Follow the browser prompt
pryx provider use google
```

**Ollama (Local)**
```bash
pryx provider add ollama
pryx provider use ollama
```

### List Available Providers
```bash
pryx provider list --available
```

### Test Provider Connection
```bash
pryx provider test openai
```

---

## 🏗️ Architecture

### Component Overview

Pryx uses a **polyglot architecture** designed for performance, security, and extensibility:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Desktop Host (Rust + Tauri)                   │
│  Port: 42424                                                         │
│  • HTTP server (axum)                                                │
│  • WebSocket for real-time TUI communication                         │
│  • Local web UI admin panel (apps/local-web/)                        │
│  • Go runtime sidecar management                                    │
│  • Native dialogs & system tray                                      │
└─────────────────────────────────────────────────────────────────────┘
                              │
                    Sidecar (Go Runtime)
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Runtime (Go)                                     │
│  • Agent execution & orchestration                                   │
│  • HTTP API + WebSocket server                                      │
│  • 84+ AI providers (models.dev)                                    │
│  • Channels (Telegram, Discord, Slack, Webhooks)                     │
│  • MCP integration                                                   │
│  • Memory & RAG                                                      │
│  • Vault (Argon2id encryption)                                      │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Data Storage                                    │
│  • SQLite database                                                   │
│  • OS Keychain (credentials)                                         │
│  • File system (sessions, logs)                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Local Web Admin UI

Access the local admin panel at **http://localhost:42424** when running in desktop mode:

- **Dashboard** - Overview of agent status, recent sessions, and cost tracking
- **Providers** - Configure AI providers (OpenAI, Anthropic, etc.)
- **Channels** - Set up Telegram, Discord, Slack integrations
- **MCP Tools** - Manage Model Context Protocol servers
- **Sessions** - Browse and resume conversations
- **Settings** - Configure preferences and permissions

### Key Design Principles

1. **Local-First by Default** - All data stays on your device unless explicitly enabled for sync
2. **Sovereign Security** - Keys stored in OS keychain, not plaintext files
3. **Sidecar Architecture** - UI and runtime are separate processes for crash isolation
4. **Port 42424** - Unified port for all desktop host services (HTTP, WebSocket, static files)
5. **Extensible via MCP** - Add tools without rebuilding the host
6. **Observable** - Every action is traceable with comprehensive audit logs

### Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Host** | Rust + Tauri v2 | Desktop wrapper, HTTP/WebSocket server (port 42424), native dialogs |
| **Runtime** | Go 1.24+ | Agent execution, AI providers, channels, MCP integration |
| **Local Web** | React + TypeScript + Vite | Admin UI served by host on port 42424 |
| **TUI** | TypeScript + Solid + OpenTUI | Terminal interface, keyboard-driven workflow |
| **Vault** | Argon2id + OS Keychain | Secure credential storage with scope-based access |
| **Channels** | Go native clients | Telegram, Discord, Slack, webhook integrations |
| **MCP** | Model Context Protocol | Extensible tool integration |

---

## 🎓 Common Workflows

### View Session History
1. Press `/` to open command palette
2. Type "sessions" or press `2`
3. Browse and resume past conversations

### Manage Multiple Providers
1. Press `/` → "Providers" or use `pryx provider list`
2. Add multiple providers: `pryx provider add <name>`
3. Switch between them: `pryx provider use <name>`

### Cost Monitoring
```bash
# View cost summary
pryx cost summary

# Daily breakdown
pryx cost daily 7

# Set budget
pryx cost budget set 100

# Get optimization suggestions
pryx cost optimize
```

### Enable MCP Tools
1. Press `/` → "MCP Servers" or use `pryx mcp list`
2. Browse available tools
3. Enable servers: `pryx mcp enable filesystem`

### Run System Diagnostics
```bash
pryx doctor
```

---

## 🛠️ Troubleshooting

### TUI Not Connecting to Runtime

**Symptom:** TUI shows "Disconnected" or "Runtime Error"

**Solutions:**
1. Check if host is running:
    ```bash
    curl http://localhost:42424/health
    ```

2. Start desktop host (automatically starts runtime sidecar):
    ```bash
    pryx
    ```

3. Check port file:
    ```bash
    cat ~/.pryx/runtime.port
    ```

4. Use explicit WebSocket URL:
    ```bash
    export PRYX_WS_URL=ws://localhost:42424/ws
    pryx
    ```

### Port Already in Use (42424)

**Symptom:** "Address already in use" error on port 42424

**Solutions:**
1. Find the process:
    ```bash
    lsof -i :42424
    ```

2. Kill the process:
    ```bash
    kill <PID>
    ```

3. Or configure different port:
    ```bash
    export PRYX_HOST_PORT=42425
    pryx
    ```

### Provider Connection Failed

**Symptom:** "Failed to connect to provider"

**Solutions:**
1. Test connection:
   ```bash
   pryx provider test openai
   ```

2. Verify API key:
   ```bash
   pryx provider set-key openai
   ```

3. Check provider status:
   ```bash
   pryx provider status
   ```

### Channel Bot Not Responding

**Symptom:** Bot added but no responses

**Solutions:**

**Telegram:**
1. Verify bot token: `pryx channel test telegram`
2. Check bot is enabled: `pryx channel status telegram`
3. Ensure bot has been started (send /start to the bot)

**Discord:**
1. Verify bot permissions (needs Message Content Intent)
2. Check slash commands are synced: `pryx channel sync discord`
3. Ensure bot is in the server and has permissions

**Slack:**
1. Verify event subscriptions are configured
2. Check bot has required scopes
3. Test webhook: `pryx channel test slack`

### Memory Issues

**Symptom:** High memory usage or slowdowns

**Solutions:**
1. Enable memory profiling:
   ```bash
   pryx config set enable_memory_profiling true
   pryx runtime
   ```

2. Check memory limits:
   ```bash
   pryx config get max_memory_mb
   ```

3. Adjust limits:
   ```bash
   pryx config set max_memory_mb 1024
   ```

### More Help

- **Documentation:** [docs/](https://github.com/irfndi/pryx/tree/main/docs)
- **GitHub Issues:** [Report a bug](https://github.com/irfndi/pryx/issues)
- **Community:** [Discord](https://discord.gg/pryx) | [Slack](https://join.slack.com/pryx/shared_invite/zt-...)

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/irfndi/pryx.git
   cd pryx
   ```

2. **Install development tools**
   ```bash
   make install-tools
   ```

3. **Install dependencies**
   ```bash
   make install-deps
   ```

4. **Start development stack**
   ```bash
   make dev
   ```

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# E2E tests
make test-e2e

# With coverage
make test-coverage
```

### Code Style

```bash
# Format all code
make format

# Lint all code
make lint

# Run comprehensive checks
make check
```

### Submitting Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'feat: add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Contribution Guidelines

- Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification
- Add tests for new features
- Update documentation as needed
- Ensure all CI checks pass

---

## 📚 Documentation

- **[PRD](docs/prd/prd.md)** - Product Requirements Document
- **[Architecture](docs/architecture/)** - Technical architecture and design
- **[API Reference](docs/api/)** - Runtime HTTP API documentation
- **[Security](docs/security/)** - Security audit and best practices
- **[Testing](docs/testing/TESTING.md)** - Test strategy and coverage
- **[Build System](BUILD_SYSTEM.md)** - Build and tooling guide

---

## 🗺️ Roadmap

### v1.0 (Current)
- ✅ Multi-channel support (Telegram, Discord, Slack)
- ✅ 84+ AI providers via models.dev
- ✅ Secure vault with Argon2id
- ✅ MCP tool integration
- ✅ Agent spawning
- ✅ Rich TUI interface
- ✅ Cost tracking and observability

### v1.5 (Planned)
- WhatsApp channel integration
- Auto-update mechanism
- Web UI for headless servers
- Local LLM inference support
- Plugin architecture

### v2.0 (Future)
- Mobile native apps (iOS, Android)
- Skills marketplace
- Advanced multi-device sync
- Voice interface
- Collaborative sessions

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🌟 Acknowledgments

- **[Tauri](https://tauri.app/)** - Desktop application framework
- **[SolidJS](https://www.solidjs.com/)** - Reactive UI library
- **[OpenTUI](https://github.com/opencodeproject/opentui)** - Terminal UI framework
- **[models.dev](https://models.dev)** - AI provider catalog
- **[Model Context Protocol](https://modelcontextprotocol.io/)** - Extensible tool integration

---

## 📞 Support

- **GitHub:** [irfndi/pryx](https://github.com/irfndi/pryx)
- **Documentation:** [docs.pryx.ai](https://docs.pryx.ai) (coming soon)
- **Discord:** [discord.gg/pryx](https://discord.gg/pryx)
- **Email:** support@pryx.ai

---

**Made with ❤️ by the [irfndi/pryx](https://github.com/irfndi/pryx) & Pryx Community**

*Take control of your AI. Be sovereign.*
