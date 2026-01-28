# Product Roadmap: Pryx v2 and Beyond

> **Version**: 2.0 (Roadmap)  
> **Status**: Planning  
> **Last Updated**: 2026-01-27  
> **Depends On**: `docs/prd/prd.md` (v1 PRD - must ship first)

---

## Document Purpose

This document defines the **post-v1 roadmap** for Pryx, covering features and capabilities planned after the initial release. Unlike the v1 PRD (which contains detailed requirements), this is a **strategic roadmap** that:

1. Prioritizes features deferred from v1
2. Defines new capabilities based on user feedback
3. Establishes long-term product vision
4. Outlines monetization and sustainability strategy

**Relationship to v1 PRD**: This document assumes all v1 milestones (M1-M5) are complete.

---

## 0) Strategic Principles (Carried Forward)

These principles from v1 remain inviolable:

| Principle | v2 Implications |
|-----------|-----------------|
| **Sovereign-by-default** | Local LLM support, encrypted sync, no forced cloud |
| **Zero-friction onboarding** | Chat integrations work seamlessly on any device |
| **Safe execution** | Enterprise policy engine, compliance features |
| **Observable** | Advanced analytics, team dashboards |
| **Simple + reusable** | Plugin architecture, community ecosystem |

---

## 1) Executive Summary: The v2 Vision

**v1 established Pryx as**: A reliable, local-first AI agent with excellent UX.

**v2 expands Pryx into**: A complete AI operations platform with:
- **Local AI**: Run models locally (Ollama, llama.cpp, MLX)
- **Ecosystem**: Skills marketplace and community plugins
- **Enterprise**: Team features, compliance, advanced policies
- **Everywhere**: More channels (WhatsApp, Email, SMS) = mobile access via chat apps
- **Sustainable**: Clear monetization without compromising sovereignty

### Target State by End of v2

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Pryx v2 Platform                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │      Desktop        │  │    CLI      │  │       Server        │ │
│  │      (Tauri)        │  │   (TUI)     │  │     (Headless)      │ │
│  └──────────┬──────────┘  └──────┬──────┘  └──────────┬──────────┘ │
│             │                    │                    │             │
│             └────────────────────┴────────────────────┘             │
│                                   │                                  │
│                    ┌──────────────┴──────────────┐                  │
│                    │        Pryx Core            │                  │
│                    │  ┌────────────────────────┐ │                  │
│                    │  │ Local LLM  │ Cloud LLM │ │                  │
│                    │  │ (Ollama)   │ (Anthropic)│ │                  │
│                    │  └────────────────────────┘ │                  │
│                    │  ┌────────────────────────┐ │                  │
│                    │  │    Skills Ecosystem    │ │                  │
│                    │  │  Community │ Official  │ │                  │
│                    │  └────────────────────────┘ │                  │
│                    └─────────────────────────────┘                  │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │          Channels = Mobile Interface (via chat apps)           ││
│  │  Telegram │ WhatsApp │ Discord │ Slack │ Email │ SMS │ Voice   ││
│  │     📱        📱         📱        📱      📱      📱           ││
│  │  (Users control Pryx from their phones via these apps)         ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

> **Key Insight (Clawdbot Pattern)**: No native mobile app needed. Users already have 
> Telegram/WhatsApp/Discord on their phones. These chat integrations ARE the mobile interface.

**Telegram execution model (clarified)**:
- **Cloud-hosted webhook** is the default “zero-install” path (Pryx Channels Cloud): Pryx hosts the webhook receiver + bot logic; users only provide bot token and link a chat. This is a monetizable cloud convenience feature.
- **Device-hosted polling** remains a sovereignty option: pryx-core can run the bot locally so model API keys never leave the user’s device; Mesh ensures only one active poller per bot token.

---

## 2) Features Deferred from v1

These were explicitly marked as "Non-Goals (Initial)" in v1 PRD:

| Feature | v1 Status | v2 Plan | Priority |
|---------|-----------|---------|----------|
| ~~Mobile native apps~~ | Deferred | **Cancelled** - Chat integrations are the mobile interface | N/A |
| Skills marketplace | Deferred | v2.1 | High |
| Local LLM inference | Deferred | v2.0 | Critical |
| Voice wake word | Deferred | v2.1 | Medium |
| Mandatory cloud sync | Never | Remains opt-in only | N/A |

> **Why no mobile app?** Users control their Pryx agent via Telegram, WhatsApp, Discord, 
> or Slack from their phones. These apps are already installed, work offline, and provide 
> push notifications. Building a native mobile app would duplicate functionality that 
> chat integrations already provide. This is the proven Clawdbot pattern.

---

## 3) Release Timeline

### Overview

| Version | Timeframe | Theme | Key Deliverables |
|---------|-----------|-------|------------------|
| **v2.0** | Months 1-4 post-v1 | Local AI & Foundation | Local LLM, plugin SDK, channel polish |
| **v2.1** | Months 5-8 post-v1 | Ecosystem & Channels | Marketplace, voice, WhatsApp, Email, SMS |
| **v2.2** | Months 9-12 post-v1 | Enterprise & Scale | Team features, compliance, performance |
| **v3.0** | Year 2+ | Platform | Agent-to-agent, autonomous workflows |

---

## 4) v2.0: Local AI & Foundation

**Theme**: Enable fully offline operation and establish plugin ecosystem foundation.

**Timeline**: Months 1-4 post-v1

### 4.1 Local LLM Integration (Critical)

**Goal**: Run AI models entirely on-device without cloud dependency.

| Requirement | Description | Success Metric |
|-------------|-------------|----------------|
| Ollama integration | Auto-detect local Ollama, model selection UI | Works on first try for 90% users |
| llama.cpp support | Native integration for maximum performance | Comparable speed to Ollama |
| MLX support (macOS) | Apple Silicon optimization | 2x faster than llama.cpp on M-series |
| Model management | Download, switch, delete models via UI | All operations in <3 clicks |
| Hybrid routing | Auto-fallback: local → cloud based on capability | Seamless, user-configured rules |
| Offline mode | Full functionality with no internet | 100% core features work offline |

**Architecture**:
```
┌─────────────────────────────────────────────────────────────────┐
│                     Model Router (pryx-core)                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │  Local Provider │  │  Local Provider │  │  Cloud Provider │ │
│  │     Ollama      │  │    llama.cpp    │  │    Anthropic    │ │
│  │                 │  │      MLX        │  │     OpenAI      │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘ │
│           │                    │                    │          │
│           └────────────────────┴────────────────────┘          │
│                              │                                  │
│                    ┌─────────┴─────────┐                       │
│                    │   Routing Rules   │                       │
│                    │ - Cost thresholds │                       │
│                    │ - Capability match│                       │
│                    │ - Latency targets │                       │
│                    │ - Privacy level   │                       │
│                    └───────────────────┘                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**User Stories**:
```gherkin
Scenario: First-time local LLM setup
  Given user has Ollama installed with llama3 model
  When user opens Pryx settings
  Then Ollama is auto-detected with available models
  And user can select llama3 as default model
  And subsequent chats use local model with no network calls

Scenario: Offline operation
  Given user has local model configured
  When network is disconnected
  Then Pryx continues working normally
  And status bar shows "Offline (Local Model)"

Scenario: Hybrid routing
  Given user has both local and cloud models configured
  When task requires advanced reasoning (detected heuristically)
  Then system prompts: "This task may benefit from Claude. Use cloud model?"
  And user can approve or continue with local
```

### 4.2 Plugin SDK & Architecture

**Goal**: Enable third-party tool development with first-class developer experience.

| Requirement | Description | Success Metric |
|-------------|-------------|----------------|
| Plugin SDK | TypeScript/JavaScript SDK for tool development | Hello World in <5 minutes |
| Plugin manifest | Declarative tool definition (JSON/YAML) | No code for simple tools |
| Sandboxed execution | Plugins run in isolated environment | Zero host access by default |
| Permission model | Granular permissions (network, fs, shell) | Per-plugin approval |
| Hot reload | Live reload during development | <1s reload time |
| Plugin CLI | `pryx plugin init/dev/build/publish` | Full development lifecycle |

**Permission Approval UX (v1 → v2)**:
- **Primary surface (CLI/TUI)**: approvals render inline in the terminal UI (fast `y/n` prompt) because this is the first interaction surface.
- **Desktop host (optional)**: if a Tauri host app is running, approvals may use native OS dialogs for stronger modality.
- **Pluggable transport**: runtime emits `approval.needed` events and accepts `approval.resolve` responses over its local WS bus so any surface (TUI, desktop, web) can drive approvals.

**Plugin Structure**:
```
my-plugin/
├── manifest.json       # Metadata, permissions, entry points
├── src/
│   └── index.ts       # Plugin code
├── README.md          # Documentation
└── package.json       # Dependencies (bundled)
```

**manifest.json Example**:
```json
{
  "name": "github-issues",
  "version": "1.0.0",
  "description": "Create and manage GitHub issues",
  "author": "pryx-community",
  "permissions": [
    "network:api.github.com",
    "storage:local"
  ],
  "tools": [
    {
      "name": "github.create_issue",
      "description": "Create a new GitHub issue",
      "parameters": {
        "repo": { "type": "string", "required": true },
        "title": { "type": "string", "required": true },
        "body": { "type": "string" }
      }
    }
  ]
}
```

### 4.3 v2.0 Success Metrics

| Metric | Target |
|--------|--------|
| Local LLM adoption | 40% of users try local model |
| Offline usage | 20% sessions are fully offline |
| Plugin SDK satisfaction | NPS ≥60 among plugin developers |
| Channel usage from mobile | 50% of Telegram/WhatsApp messages from mobile devices |

---

## 5) v2.1: Ecosystem & Channels

**Theme**: Build thriving community ecosystem and expand reach.

**Timeline**: Months 5-8 post-v1

### 5.1 Skills Marketplace

**Goal**: Create discovery and distribution platform for community plugins.

| Component | Description |
|-----------|-------------|
| **Registry** | Central catalog of verified plugins |
| **Discovery** | Search, categories, trending, recommended |
| **Installation** | One-click install from UI |
| **Updates** | Automatic updates with rollback |
| **Reviews** | Community ratings and reviews |
| **Verification** | Security audit for "verified" badge |

**Monetization Options**:
| Model | Description | Implementation |
|-------|-------------|----------------|
| Free tier | Core plugins free forever | Default |
| Donations | Tip jar for plugin authors | Optional |
| Premium plugins | Paid plugins (author keeps 85%) | v2.2+ |
| Sponsored plugins | Featured placement | v2.2+ |

**Registry API**:
```
GET  /plugins                    # List/search plugins
GET  /plugins/{id}               # Plugin details
POST /plugins/{id}/install       # Install plugin
GET  /plugins/{id}/reviews       # Reviews
POST /plugins/{id}/reviews       # Submit review
```

### 5.2 Voice Interface

**Goal**: Enable hands-free interaction with Pryx.

| Feature | Description | Priority |
|---------|-------------|----------|
| Voice input | Speech-to-text for commands | High |
| Voice output | Text-to-speech for responses | High |
| Wake word | "Hey Pryx" activation | Medium |
| Continuous conversation | Multi-turn voice dialog | Medium |
| Voice approvals | "Approve" / "Deny" by voice | High |

**Implementation Strategy**:
- Use system APIs (macOS Speech, Windows SAPI, Web Speech API)
- Optional Whisper integration for better accuracy
- Wake word via Porcupine or similar (offline capable)
- Voice output via system TTS or ElevenLabs (optional)

### 5.3 Expanded Channels

| Channel | v1 Status | v2.1 Scope |
|---------|-----------|------------|
| **WhatsApp** | Planned | Cloud API integration (official) |
| **Email** | Not planned | IMAP/SMTP with smart filtering |
| **SMS** | Not planned | Twilio integration |
| **Matrix** | Not planned | Protocol support |
| **IRC** | Not planned | Basic integration |
| **Microsoft Teams** | Not planned | Enterprise channel |

**WhatsApp Strategy (Refined)**:
- Use official WhatsApp Cloud API (not Baileys)
- Business account required
- Rate limits: 1,000 messages/day (free tier)
- Full media support (images, documents, voice)

### 5.4 v2.1 Success Metrics

| Metric | Target |
|--------|--------|
| Marketplace plugins | 100 published plugins |
| Plugin installs | 10,000 total installs |
| Voice adoption | 15% of users enable voice |
| WhatsApp connections | 1,000 active bots |
| Channel diversity | Average user has 2.5 channels |

---

## 6) v2.2: Enterprise & Scale

**Theme**: Enterprise readiness and performance at scale.

**Timeline**: Months 9-12 post-v1

### 6.1 Team & Organization Features

| Feature | Description |
|---------|-------------|
| **Team workspaces** | Shared workspaces with role-based access |
| **SSO/SAML** | Enterprise identity provider integration |
| **Audit logging** | Exportable, tamper-evident audit trails |
| **Admin console** | Centralized team management |
| **Usage analytics** | Team-wide cost and usage dashboards |
| **Policy templates** | Pre-built compliance policies |

**Team Roles**:
| Role | Permissions |
|------|-------------|
| Owner | Full access, billing, member management |
| Admin | Configuration, policy, but not billing |
| Member | Use Pryx, personal sessions |
| Viewer | Read-only access to shared sessions |

### 6.2 Compliance & Security

| Feature | Description | Target Compliance |
|---------|-------------|-------------------|
| **SOC 2 Type II** | Security controls audit | Enterprise |
| **GDPR tooling** | Data export, deletion, consent management | EU |
| **HIPAA mode** | Enhanced encryption, audit logging | Healthcare |
| **Data residency** | Choose where data is stored | Regulated industries |
| **Key management** | BYOK (Bring Your Own Key) for encryption | Enterprise |

### 6.3 Performance & Scale

| Optimization | Target |
|--------------|--------|
| Session storage | Support 100,000+ sessions |
| Concurrent channels | 100+ simultaneous connections |
| Memory optimization | <150MB idle (down from 200MB) |
| Startup time | <1.5s (down from 3s) |
| Response latency | <30ms local processing overhead |

### 6.4 v2.2 Success Metrics

| Metric | Target |
|--------|--------|
| Enterprise customers | 10 paying teams |
| Team size average | 5 members per team |
| SOC 2 certification | Achieved |
| Uptime (server mode) | 99.9% |
| Enterprise NPS | ≥70 |

---

## 7) v3.0 Vision: The Platform

**Theme**: From tool to platform - autonomous agents and workflows.

**Timeline**: Year 2+

### 7.1 Agent-to-Agent Communication

Enable Pryx instances to communicate and delegate tasks.

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│   Pryx A    │◄───────►│   Pryx B    │◄───────►│   Pryx C    │
│  (Laptop)   │  Mesh   │  (Server)   │  Mesh   │  (Mobile)   │
│             │         │             │         │             │
│  Personal   │         │  Shared     │         │  On-the-go  │
│  Tasks      │         │  Workloads  │         │  Access     │
└─────────────┘         └─────────────┘         └─────────────┘
```

### 7.2 Autonomous Workflows

**Capabilities**:
- Scheduled tasks (cron-like)
- Event-driven automation
- Multi-step workflows with conditional logic
- Long-running background tasks
- Self-healing workflows (retry, fallback)

#### 7.2.1 Scheduled Tasks Platform (Enhanced from v1.1)

**Built on v1.1 foundation with advanced capabilities**:

| Feature | v1.1 Capability | v2 Enhancement |
|---------|-----------------|----------------|
| Task Dashboard | View tasks, history | Real-time monitoring, bulk operations |
| Triggers | Cron, interval | Cron, interval, event-based, webhook |
| Task History | Per-task logs | Aggregated analytics, trends, cost reports |
| Notifications | Push notifications | Custom notification rules, escalations |
| Retry Policies | Basic retry | Configurable backoff, on-failure actions |
| Cross-Device | Local/remote | Device orchestration, load balancing |

**v2 New Capabilities**:

**1. Task Chaining**
- Define dependencies between tasks
- Pass output from Task A to Task B
- Create DAG (Directed Acyclic Graph) workflows
- Visual workflow editor in UI

```
Example: Stock Watch → Trade Decision → Email Alert
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Stock Watch │────►│ Trade Decision│────►│ Email Alert  │
│ (Every 4h) │     │ (If price > │     │ (Notify me) │
│             │     │  $X)       │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
```

**2. Conditional Logic**
- IF/THEN/ELSE branching in workflows
- Wait conditions (wait for API response, wait for file change)
- Loop constructs (repeat N times, repeat until condition)
- Parallel execution (run multiple tasks simultaneously)

**3. Advanced Triggers**
- **Event-based**: File system events (file created/modified), Git events (push/PR), API webhooks
- **Webhook triggers**: Receive HTTP webhook to trigger task
- **State triggers**: Trigger when system state changes (CPU > 80%, disk space < 10%)
- **Custom triggers**: User-defined scripts that return boolean

**4. Task Templates Marketplace**
- Community-shared task templates
- Verified templates (official review)
- User-generated templates (community rating)
- One-click template installation

**Template Categories**:
```
┌─────────────────────────────────────────────────────────┐
│  📦 Task Templates Marketplace                      │
├─────────────────────────────────────────────────────────┤
│                                                   │
│  🔥 Trending This Week                            │
│     • Stock Alert Bot (2.3k installs)            │
│     • Daily Report Generator (1.8k installs)       │
│     • Auto-Commit Bot (1.5k installs)             │
│                                                   │
│  💹 Finance & Trading                             │
│     • Crypto Price Monitor                           │
│     • Portfolio Rebalancer                          │
│     • Trade Signal Notifier                          │
│                                                   │
│  📝 Development                                  │
│     • CI/CD Pipeline Runner                         │
│     • Code Review Assistant                          │
│     • Bug Tracker Auto-Triage                       │
│                                                   │
│  📊 Data & Analytics                             │
│     • ETL Job Scheduler                            │
│     • Database Backup Automator                       │
│     • Report Generator                              │
│                                                   │
│  🔧 Operations                                   │
│     • Log Monitor & Alert                           │
│     • Server Health Checker                         │
│     • Auto-Deployment on Push                       │
│                                                   │
│  🔌 Security                                     │
│     • SSL Certificate Expiry Monitor                 │
│     • Vulnerability Scanner                          │
│     • Access Log Analyzer                           │
│                                                   │
└─────────────────────────────────────────────────────────┘
```

**Template Installation Flow**:
```gherkin
Scenario: Install task template
  Given user opens Task Templates Marketplace
  When user clicks "Stock Alert Bot"
  Then template details page shows:
    - Description and use cases
    - Required permissions (API keys, tools)
    - Estimated cost per run
    - Community rating and reviews
    - Last updated date
  When user clicks "Install"
  Then template is added to Scheduled Tasks
  And user can customize triggers and parameters
  And template is ready to run
```

**5. Monitoring Dashboards**
- Real-time task execution visualization
- Gantt charts for multi-step workflows
- Success/failure rates over time
- Cost trends and budget alerts
- Anomaly detection (unusual patterns)

**Dashboard Views**:
```
┌─────────────────────────────────────────────────────────────┐
│  📊 Scheduled Tasks Dashboard (v2)                     │
├─────────────────────────────────────────────────────────────┤
│                                                         │
│  Overview (Last 30 Days):                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Total Runs: 1,247 | Success: 98.2%       │   │
│  │  Total Cost: $12.34 | Avg Cost/Run: $0.01   │   │
│  │  Active Tasks: 23 | Paused: 5               │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                         │
│  Live Executions (Right Now):                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  ▶ Stock Watch (Step 2/3: Analyzing...)     │   │
│  │     Device: Server | Est. remaining: 45s        │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │  ▶ Log Monitor (Running for 2h 34m)         │   │
│  │     Device: Desktop | Events processed: 1,234    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                         │
│  Success/Failure Trends:                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  ████░░░░░░░░░░░░░  85% Success Rate     │   │
│  │  Last 7 days | Last 30 days | All time         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                         │
└─────────────────────────────────────────────────────────────┘
```

**6. Advanced Retry Policies**
- **Exponential backoff**: 1s, 2s, 4s, 8s, ...
- **Linear backoff**: Fixed increment (1s, 2s, 3s, ...)
- **Custom backoff**: User-defined delay sequence
- **Max retry limit**: Configurable per task or globally
- **On-failure actions**:
  - Notify immediately
  - Wait and retry
  - Mark task as failed
  - Fallback to alternative task
  - Escalate to human (send message to Telegram)

**Retry Policy Configuration**:
```json
{
  "name": "Stock Watch",
  "retry_policy": {
    "strategy": "exponential_backoff",
    "max_retries": 5,
    "initial_delay": "1s",
    "max_delay": "300s",
    "on_final_failure": {
      "action": "escalate_to_human",
      "channel": "telegram",
      "message": "🚨 Stock Watch failed after 5 retries. Manual intervention required."
    },
    "on_intermittent_failure": {
      "action": "continue",
      "log_level": "warning"
    }
  }
}
```

**7. Task Cost Management**
- **Per-task budgets**: Set max cost per task execution
- **Monthly budgets**: Set total monthly spend limit
- **Cost alerts**: Notify at 50%, 75%, 90% of budget
- **Auto-pause**: Pause tasks when budget exceeded
- **Cost optimization**: Suggest cheaper model alternatives

**Budget Configuration**:
```json
{
  "budgets": {
    "monthly_total": {
      "limit_usd": 50.00,
      "alerts": [25.00, 37.50, 45.00],
      "action_on_exceed": "pause_all"
    },
    "per_task": {
      "default_limit_usd": 0.10,
      "task_limits": {
        "Stock Watch": 0.05,
        "Data ETL": 0.50
      }
    }
  }
}
```

**8. Device Orchestration**
- **Load balancing**: Distribute tasks across multiple devices
- **Device capabilities**: Match task requirements to device capabilities
- **Health-aware routing**: Route tasks away from unhealthy devices
- **Affinity rules**: Certain tasks prefer specific devices (e.g., server for Docker)

**Device Selection Algorithm**:
```go
func SelectDevice(task Task, devices []Device) (Device, error) {
    // Filter by requirements
    candidates := filterByRequirements(devices, task.Requirements)

    // Sort by health (available > busy > offline)
    sort(candidates, byHealthDescending)

    // Check budget/limits on each candidate
    for _, device := range candidates {
        if device.HasCapacityFor(task) {
            return device, nil
        }
    }

    return nil, errors.New("no available device with sufficient capacity")
}
```

**9. Task Debugging & Replay**
- **Debug mode**: Enable detailed logging for specific tasks
- **Dry run**: Execute task without side effects
- **Replay**: Re-run historical task execution with same inputs
- **Diff view**: Compare outputs across different runs

**Debug UI**:
```
┌─────────────────────────────────────────────────────────────┐
│  🔍 Task Debug: Stock Watch (Run #42)                │
├─────────────────────────────────────────────────────────────┤
│                                                         │
│  Execution Log:                                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  [14:00:00] Task started                   │   │
│  │  [14:00:01] Fetching stock price from API     │   │
│  │  [14:00:02] API response: $178.42          │   │
│  │  [14:00:03] Checking condition: price > $175   │   │
│  │  [14:00:03] Condition TRUE, proceeding        │   │
│  │  [14:00:04] Searching news articles          │   │
│  │  [14:00:05] Found 3 articles                │   │
│  │  [14:00:06] Saving to Google Sheet          │   │
│  │  [14:00:07] Task completed successfully     │   │
│  │  Tokens used: 1,245 | Cost: $0.02          │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                         │
│  Actions:                                                │
│  [🔄 Replay with same inputs]  [🧪 Dry run]          │
│  [📝 Edit task]  [📊 Compare with other runs]       │
│  [🐛 Report bug]                                       │
│                                                         │
└─────────────────────────────────────────────────────────────┘
```

**10. Task Export & Import**
- **Export tasks**: Backup all tasks as JSON/YAML
- **Import tasks**: Restore tasks from backup file
- **Share tasks**: Generate shareable URL or QR code
- **Version control**: Track task configurations in Git

**Export Format**:
```json
{
  "version": "2.0",
  "exported_at": "2026-01-27T12:00:00Z",
  "tasks": [
    {
      "id": "task-uuid",
      "name": "Stock Watch",
      "description": "Monitor stock price and alert",
      "trigger": {
        "type": "interval",
        "config": {"hours": 4}
      },
      "action": {
        "type": "workflow",
        "steps": [...]
      },
      "retry_policy": {...},
      "budget": {...}
    }
  ]
}
```

**Success Metrics**:
| Metric | v2.0 Target | v2.1 Target |
|--------|-------------|--------------|
| Active scheduled tasks | 10,000 | 50,000 |
| Task marketplace templates | 100 | 500 |
| Task success rate | 95% | 97% |
| Avg task setup time | <2 min | <1 min |
| Template installs | 5,000 | 25,000 |

### 7.3 Specialized Agents

| Agent Type | Purpose |
|------------|---------|
| **Code Agent** | Deep IDE integration, PR reviews, refactoring |
| **Research Agent** | Web research, summarization, knowledge management |
| **Ops Agent** | Monitoring, alerting, incident response |
| **Data Agent** | Analysis, visualization, reporting |
| **Personal Agent** | Calendar, email, task management |

### 7.4 Platform Marketplace

Beyond plugins - full agent templates:
- Pre-built agent configurations
- Workflow templates
- Integration bundles
- Enterprise solutions

---

## 8) Monetization & Sustainability

### 8.1 Understanding Our Cost Structure

**Key Insight**: Pryx runs on user's hardware, not ours. This fundamentally changes monetization.

| Cost Type | Who Pays | Notes |
|-----------|----------|-------|
| **LLM API costs** | User (BYOK) | Users bring their own Anthropic/OpenAI keys |
| **Compute & Storage** | User | Runs on user's desktop/server |
| **Auth Workers** | Pryx | Cloudflare Workers (minimal cost) |
| **Telemetry Storage** | Pryx | Optional, low volume |
| **Update Distribution** | Pryx | CDN/GitHub releases |
| **Plugin Registry** | Pryx | Hosting and verification |

**Pryx's actual costs are minimal** → We can be generous with free tier.

### 8.2 Principles (Revised)

1. **Core is 100% free forever**: All local functionality, unlimited channels, all plugins
2. **Pay only for cloud services**: Things that require our infrastructure (sync, hosted webhooks, hosted routing, hosted key vault)
3. **BYOK always supported**: Users who bring their own model keys can stay local-only, or optionally use BYOK-in-cloud for hosted integrations
4. **Optional convenience layers**: Pryx-hosted models and Pryx-hosted channel webhooks for users who don't want to run always-on devices
5. **Compete with free**: Be more generous than Clawdbot to win users

### 8.3 Revenue Streams

| Stream | Description | Timeline | Margin |
|--------|-------------|----------|--------|
| **Pryx Gateway** | Hosted model access (no BYOK needed) | v2.0 | ~10% on API costs |
| **Pryx Channels Cloud** | Hosted channel webhooks + routing (Telegram webhook, managed secrets, delivery retries) | v2.0 | High |
| **Pryx Sync** | Cloud backup of config, sessions, policies | v2.1 | High (storage is cheap) |
| **Pryx Publish** | Public dashboards for observability | v2.2 | High |
| **Team Features** | Admin console, shared workspaces, audit export | v2.2 | High |
| **Enterprise** | SSO, compliance, support SLA | v2.2 | Very high |
| **Marketplace Cut** | 15% of premium plugin sales | v2.2 | Pure margin |
| **Donations/Sponsors** | GitHub Sponsors, Open Collective | v2.0 | 100% |

### 8.4 Pricing Model (Revised)

#### Free Tier (Generous - Compete with Clawdbot)

| Feature | Limit | Notes |
|---------|-------|-------|
| **All core features** | Unlimited | Chat, channels, tools, plugins |
| **Channels (device-hosted)** | Unlimited | Telegram polling, Discord polling, local webhooks, etc. (no Pryx hosting) |
| **Channels Cloud (hosted webhooks)** | Limited | 1 Telegram bot / 1 linked chat to demo zero-install onboarding |
| **Sessions** | Unlimited | Stored locally |
| **Plugins** | All community plugins | Free forever |
| **Local LLM** | Unlimited | Ollama, llama.cpp, MLX |
| **BYOK Cloud Models** | Unlimited | Bring your own API keys |
| **Workspaces** | 3 | Upgrade for more |
| **Devices (Mesh)** | 2 | Upgrade for more |
| **Telemetry retention** | 7 days | Local only |

> **Why so generous?** Clawdbot is 100% free with large ecosystem. We need to win on features AND price.

#### Pryx Gateway (Pay-as-you-go)

For users who don't want to manage API keys:

| Feature | Price |
|---------|-------|
| **Model Access** | Provider cost + 10% Pryx fee |
| **No API key management** | Included |
| **Usage dashboard** | Included |
| **Cost alerts** | Included |
| **Spending limits** | Included |

**Example**: Claude Sonnet at $3/M input tokens → $3.30/M through Pryx Gateway

> **Why 10%?** Lower than OpenRouter (5.5% + fees), competitive with Vercel (0% but less features). Our value: integrated experience, no key management, unified billing.

#### Pryx Pro ($8/month)

For power users who want cloud features:

| Feature | Description |
|---------|-------------|
| **Pryx Sync** | Cloud backup of config, policies, session metadata |
| **Pryx Channels Cloud** | Hosted Telegram webhook mode (no always-on device), bot token vault, retries, integration status |
| **Unlimited workspaces** | No limit on workspace count |
| **Unlimited devices** | Full Pryx Mesh access |
| **Telemetry retention** | 90 days (cloud-stored) |
| **Priority support** | Email support with 24h response |
| **Early access** | Beta features and models |
| **Profile badge** | "Pro" badge in community |

> **Why would they pay?** Cloud sync, more devices, longer telemetry. These require OUR infrastructure.

#### Pryx Team ($12/user/month)

For organizations:

| Feature | Description |
|---------|-------------|
| **Everything in Pro** | All Pro features included |
| **Shared workspaces** | Team-wide workspace sharing |
| **Admin console** | Centralized management |
| **Role-based access** | Owner, Admin, Member, Viewer |
| **Usage analytics** | Team-wide cost and usage |
| **Audit log export** | Compliance-ready exports |
| **Shared policies** | Org-wide policy templates |

#### Pryx Enterprise (Custom)

For large organizations:

| Feature | Description |
|---------|-------------|
| **Everything in Team** | All Team features included |
| **SSO/SAML** | Identity provider integration |
| **SCIM provisioning** | Automated user management |
| **SOC 2 compliance** | Security controls |
| **HIPAA mode** | Enhanced encryption |
| **Data residency** | Choose storage region |
| **Dedicated support** | Named account manager |
| **SLA** | 99.9% uptime guarantee |
| **Custom contracts** | Invoice/PO billing |

### 8.5 Comparison: Pryx vs Competitors

| Feature | Pryx Free | Pryx Pro | Cursor Pro | Continue.dev | Clawdbot |
|---------|-----------|----------|------------|--------------|----------|
| **Price** | $0 | $8/mo | $20/mo | $0 | $0 |
| **Local LLM** | ✅ | ✅ | ❌ | ✅ | ❌ |
| **BYOK** | ✅ | ✅ | Limited | ✅ | ✅ |
| **Channels** | Unlimited | Unlimited | N/A | N/A | Unlimited |
| **Cloud Sync** | ❌ | ✅ | ✅ | ❌ | ❌ |
| **Multi-device** | 2 | Unlimited | N/A | N/A | ❌ |
| **Plugins** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Open Source** | ✅ | ✅ | ❌ | ✅ | ✅ |

### 8.6 Revenue Projections

| Metric | v2.0 | v2.1 | v2.2 |
|--------|------|------|------|
| Free Users | 4,000 | 12,000 | 25,000 |
| Gateway MAU | 200 | 1,000 | 3,000 |
| Pro Subscribers | 50 | 300 | 800 |
| Team Seats | 0 | 50 | 200 |
| Enterprise Seats | 0 | 0 | 50 |
| **MRR** | $400 | $4,000 | $15,000 |

### 8.7 Why This Model Works

1. **Low costs = generous free tier**: We don't pay for compute, so we can give away more
2. **Gateway captures convenience seekers**: Some users will pay 10% to avoid key management
3. **Cloud features justify subscription**: Sync, multi-device, telemetry require our infra
4. **Enterprise is pure upside**: High margins on compliance/support
5. **Community builds ecosystem**: Free users create plugins, documentation, word-of-mouth

### 8.8 Anti-Patterns to Avoid

| Anti-Pattern | Why It Fails | Our Approach |
|--------------|--------------|--------------|
| Charging for local features | Users know it runs on their hardware | Only charge for cloud services |
| Limiting BYOK users | Punishes power users who bring value | BYOK unlimited, Gateway for convenience |
| Tight free tier limits | Can't compete with Clawdbot | Generous free tier wins market share |
| Complex pricing | Confuses users | Simple: Free, Pro ($8), Team ($12), Enterprise |
| Forced subscriptions | Alienates open-source community | Core always free, subscriptions optional |

---

## 9) Competitive Evolution

### 9.1 Competitive Threats

| Threat | Likelihood | Impact | Mitigation |
|--------|------------|--------|------------|
| Cursor adds local mode | High | High | Local-first by default, better UX |
| OpenAI Desktop goes local | Medium | High | Open source, sovereignty focus |
| New entrants | High | Medium | Community moat, plugin ecosystem |
| Cloud-only becomes norm | Low | High | Strong local-first positioning |

### 9.2 Differentiation Strategy

| Competitor Move | Pryx Response |
|-----------------|---------------|
| Cursor local mode | Emphasize openness, no subscription lock-in |
| Continue.dev adds UI | Better offline, more channels |
| OpenAI local mode | Data sovereignty, no vendor lock-in |
| New CLI tools | Full UI/UX, plugin ecosystem |

### 9.3 Moat Building

| Moat Type | v2 Investment |
|-----------|---------------|
| **Community** | Plugin ecosystem, contributor program |
| **Switching cost** | Session history, trained policies, integrations |
| **Network effects** | Skill marketplace, shared workflows |
| **Brand** | "Local-first" category leader |

---

## 10) Technical Debt & Infrastructure

### 10.1 Known Technical Debt

| Debt | Impact | v2 Plan |
|------|--------|---------|
| SQLite at scale | Performance ceiling | Consider LibSQL/Turso for sync |
| Single-binary size | Large downloads (~50MB) | Modular downloads, delta updates |
| WebView bundle | Slow first load | Code splitting, lazy loading |
| Go sidecar memory | ~100MB baseline | Profiling, optimization pass |

### 10.2 Infrastructure Evolution

| Current | v2 Evolution |
|---------|--------------|
| Manual releases | CI/CD with staged rollouts |
| Self-managed telemetry | Consider third-party (optional) |
| GitHub-only hosting | Mirror to GitLab, Codeberg |
| Cloudflare-only edge | Multi-provider option |

---

## 11) Dependencies & Risks

### 11.1 External Dependencies

| Dependency | Risk | Mitigation |
|------------|------|------------|
| Ollama | Project abandonment | Also support llama.cpp directly |
| Anthropic API | Pricing changes | Local-first, multi-provider |
| Cloudflare Workers | Vendor lock-in | Abstraction layer for edge |

### 11.2 Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Local LLM quality insufficient | Medium | High | Clear capability documentation |
| Plugin ecosystem doesn't grow | Medium | High | Official plugins, developer grants |
| Monetization fails | Medium | Critical | Multiple revenue streams |
| Team velocity | High | High | Prioritize ruthlessly |

---

## 12) Success Metrics (Overall v2)

### 12.1 User Growth

| Metric | v2.0 | v2.1 | v2.2 |
|--------|------|------|------|
| Monthly Active Users | 5,000 | 15,000 | 30,000 |
| Mobile Channel Users | 1,000 | 5,000 | 10,000 |
| Enterprise Teams | 0 | 5 | 20 |
| GitHub Stars | 2,000 | 5,000 | 10,000 |

> Note: "Mobile Channel Users" = users interacting via Telegram/WhatsApp/Discord from mobile devices

### 12.2 Ecosystem Health

| Metric | v2.0 | v2.1 | v2.2 |
|--------|------|------|------|
| Published Plugins | 20 | 100 | 300 |
| Plugin Developers | 10 | 50 | 150 |
| Community Contributors | 20 | 50 | 100 |

### 12.3 Revenue (if monetizing)

| Metric | v2.0 | v2.1 | v2.2 |
|--------|------|------|------|
| MRR | $0 | $2,000 | $10,000 |
| Paying Users | 0 | 150 | 500 |
| Enterprise ARR | $0 | $0 | $50,000 |

---

## 13) Open Questions

| Question | Options | Decision Needed By |
|----------|---------|-------------------|
| Marketplace hosting? | Self-hosted vs npm-like registry | v2.1 start |
| Voice: System API vs Whisper? | System (simple), Whisper (better) | v2.1 start |
| Enterprise pricing model? | Per-seat vs usage-based | v2.2 start |
| Open source license change? | Stay MIT vs AGPL for monetization | v2.1 |

---

## 14) Appendix

### A. Feature Prioritization Framework

Using RICE scoring:
- **R**each: How many users affected?
- **I**mpact: How much does it improve experience? (3=massive, 2=high, 1=medium, 0.5=low)
- **C**onfidence: How sure are we? (100%, 80%, 50%)
- **E**ffort: Person-months to build

RICE Score = (Reach × Impact × Confidence) / Effort

### B. Glossary Additions

| Term | Definition |
|------|------------|
| **Local LLM** | AI model running entirely on user's device |
| **Plugin** | Third-party extension providing tools to Pryx |
| **Marketplace** | Discovery and distribution platform for plugins |
| **RICE** | Prioritization framework (Reach, Impact, Confidence, Effort) |
| **Moat** | Competitive advantage that's hard to replicate |

### C. Related Documents

- `docs/prd/prd.md` - Pryx v1 PRD (prerequisite)
- Future: `docs/prd/plugin-sdk.md` - Plugin SDK specification
- Future: `docs/prd/enterprise.md` - Enterprise features PRD

---

*This roadmap is a living document. Features and timelines will evolve based on user feedback, market conditions, and resource availability. The v1 PRD remains the immediate priority.*
