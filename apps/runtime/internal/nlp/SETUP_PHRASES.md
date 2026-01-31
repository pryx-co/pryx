# NLP Setup Phrases Documentation

This document describes all the supported natural language phrases for AI-assisted setup capabilities in the NLP parser.

## Setup Intents

### 1. Setup Intent (`IntentSetup`)
Initializing or setting up a new configuration, provider, or feature.

**Supported phrases:**
- `setup <provider>` - "setup openai provider", "setup google"
- `install <provider>` - "install anthropic", "install claude"
- `initialize <provider>` - "initialize gemini", "initialize mistral"
- `prepare <provider>` - "prepare openai", "prepare my openai setup"
- `get ready to use <provider>` - "get ready to use claude", "get ready"
- `help me set up <provider>` - "help me set up gpt"
- `help me get started` - "help me get started"
- `help me configure <provider>` - "help me configure openai"
- `i want to use <provider>` - "i want to use gemini"
- `i want to set up <provider>` - "i want to set up cohere"

**Extracted Entities:**
- `provider`: openai, anthropic, google, claude, gpt, gemini, palm, mistral, llama, ollama, cohere, azure

---

### 2. Connect Intent (`IntentConnect`)
Connecting or linking with external services, bots, or integrations.

**Supported phrases:**
- `connect <channel>` - "connect my telegram bot", "connect slack"
- `link <channel>` - "link slack workspace", "link discord"
- `integrate <channel>` - "integrate microsoft teams"
- `attach <integration> to <channel>` - "attach a webhook to discord"
- `join <channel>` - "join my telegram channel", "join irc"
- `hook up <integration>` - "hook up the api integration"
- `connect with <channel>` - "connect with whatsapp"
- `add my <channel>` - "add my telegram bot"
- `add a <channel>` - "add a discord integration"
- `add the <integration>` - "add the webhook integration"

**Extracted Entities:**
- `channel`: telegram, discord, slack, teams, whatsapp, messenger, signal, matrix, irc
- `integration`: mcp, webhook, api, rest, graphql, grpc, websocket, skill, tool, plugin, filesystem

---

### 3. Configure Intent (`IntentConfigure`)
Adjusting settings or configuration for existing setups.

**Supported phrases:**
- `configure <channel> <integration>` - "configure discord webhook"
- `adjust <channel>` - "adjust my config for telegram"
- `customize <channel>` - "customize my discord bot"
- `tweak <channel>` - "tweak the settings for slack"
- `set up my settings` - "set up my settings"
- `set up my config` - "set up my config"
- `set up my configuration` - "set up my configuration"
- `change my settings` - "change my openai settings"
- `change my config` - "change my config for telegram"
- `change my configuration` - "change my configuration"

**Extracted Entities:**
- `channel`: telegram, discord, slack, teams, whatsapp, messenger, signal, matrix, irc
- `integration`: mcp, webhook, api, rest, graphql, grpc, websocket, skill, tool, plugin, filesystem
- `provider`: openai, anthropic, google, claude, gpt, gemini, palm, mistral, llama, ollama, cohere, azure

---

### 4. Enable Intent (`IntentEnable`)
Activating a feature, tool, or service.

**Supported phrases:**
- `enable <integration>` - "enable the filesystem tool", "enable mcp"
- `turn on <integration>` - "turn on mcp integration", "turn on slack"
- `activate <channel>` - "activate slack channel", "activate discord"
- `use the <integration>` - "use the mcp tool", "use the tool"
- `use a <integration>` - "use a plugin"
- `use my <integration>` - "use my webhook"
- `start <integration>` - "start mcp"

**Extracted Entities:**
- `channel`: telegram, discord, slack, teams, whatsapp, messenger, signal, matrix, irc
- `integration`: mcp, webhook, api, rest, graphql, grpc, websocket, skill, tool, plugin, filesystem

---

### 5. Disable Intent (`IntentDisable`)
Deactivating or turning off a feature, tool, or service.

**Supported phrases:**
- `disable <integration>` - "disable the filesystem tool", "disable mcp"
- `turn off <integration>` - "turn off mcp integration", "turn off slack"
- `deactivate <channel>` - "deactivate slack channel", "deactivate discord"
- `turn my <channel> off` - "turn my discord bot off"
- `turn the <integration> off` - "turn the webhook tool off"
- `turn <channel> off` - "turn slack off"
- `turn off` - "turn off"
- `stop my <channel>` - "stop my discord bot"
- `stop the <integration>` - "stop the mcp integration"
- `stop <channel>` - "stop slack"

**Extracted Entities:**
- `channel`: telegram, discord, slack, teams, whatsapp, messenger, signal, matrix, irc
- `integration`: mcp, webhook, api, rest, graphql, grpc, websocket, skill, tool, plugin, filesystem

---

## Entity Types

### Provider
AI/LLM providers that can be configured.
- **Keywords**: openai, anthropic, google, claude, gpt, gemini, palm, mistral, llama, ollama, cohere, azure

### Channel
Communication channels or messaging platforms.
- **Keywords**: telegram, discord, slack, teams, whatsapp, messenger, signal, matrix, irc

### Integration
External tools, plugins, or system integrations.
- **Keywords**: mcp, webhook, api, rest, graphql, grpc, websocket, skill, tool, plugin, filesystem

### Token
API tokens, keys, or authentication credentials.
- **Patterns**: `token: <value>`, `api key: <value>`, `secret: <value>`, `auth token: <value>`

---

## Examples

### Setting up a provider
```
Input: "setup openai provider"
ParseResult: 
  Intent: "setup"
  Entities: [{Type: "provider", Value: "openai"}]
  Confidence: 0.7
```

### Connecting a channel
```
Input: "connect my telegram bot"
ParseResult:
  Intent: "connect"
  Entities: [{Type: "channel", Value: "telegram"}]
  Confidence: 0.7
```

### Configuring with multiple entities
```
Input: "configure discord webhook"
ParseResult:
  Intent: "configure"
  Entities: [
    {Type: "channel", Value: "discord"},
    {Type: "integration", Value: "webhook"}
  ]
  Confidence: 0.7
```

### Enabling a tool
```
Input: "enable the filesystem tool"
ParseResult:
  Intent: "enable"
  Entities: [{Type: "integration", Value: "filesystem"}]
  Confidence: 0.7
```

### Disabling a service
```
Input: "turn off mcp integration"
ParseResult:
  Intent: "disable"
  Entities: [{Type: "integration", Value: "mcp"}]
  Confidence: 0.7
```

---

## Suggestions

The `SuggestSetupAction()` method provides context-aware suggestions based on detected intent and entities:

### Setup suggestions
- `setup`, `init`, `install`, `provider <name>`

### Connect suggestions
- `connect`, `link`, `integrate`, `channel <name>`, `integration <name>`

### Configure suggestions
- `config`, `settings`, `customize`, `channel <name>`, `integration <name>`, `provider <name>`

### Enable suggestions
- `enable`, `activate`, `start`, `integration <name>`, `channel <name>`

### Disable suggestions
- `disable`, `deactivate`, `stop`, `integration <name>`, `channel <name>`

---

## Notes

- All patterns are **case-insensitive** (e.g., "SETUP", "Setup", "setup" all work)
- Entity values are normalized to **lowercase** for consistency
- Confidence scores range from 0.3 to 1.0, with values below 0.6 considered ambiguous
- The parser can handle multiple entities in a single input phrase
- Patterns are scored and the highest-scoring intent is selected
