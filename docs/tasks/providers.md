# LLM Provider Configuration

Pryx supports a variety of LLM providers. You can configure them using environment variables.

## Secret Handling (Visibility + Storage)

- Provider API keys are treated as secrets and are local-only by default.
- Keys should be stored in OS keychain where available, or supplied via environment variables in service environments.
- Keys are not displayed in dashboards; UIs only show whether a key is present/configured.
- Cloud dashboards must not receive or persist provider keys unless an explicit end-to-end encrypted vault sync feature is enabled.

**Security note:** Never commit API keys to version control. Prefer environment variables or a local `.env`
file excluded from VCS. Use HTTPS for non-local endpoints.

## Supported Providers

### OpenAI
Standard usage for GPT-4, GPT-3.5, etc.

**Configuration:**
- `OPENAI_API_KEY`: Your OpenAI API Key.
- `OPENAI_BASE_URL` (Optional): Override the default API URL (`https://api.openai.com/v1`).

### Anthropic
Usage for Claude 3 Opus, Sonnet, Haiku.

**Configuration:**
- `ANTHROPIC_API_KEY`: Your Anthropic API Key.

### OpenRouter
Access to hundreds of models (Llama 3, Mistral, Command R+, etc.) via a unified API.

**Configuration:**
- `OPENROUTER_API_KEY`: Your OpenRouter API Key.
- Note: Pryx automatically uses `https://openrouter.ai/api/v1` as the base URL when using the OpenRouter provider.

## GLM & Custom Models (LocalAI/Ollama)
You can connect to any OpenAI-compatible API (like LocalAI, Ollama, or self-hosted GLM services) by configuring the OpenAI provider with a custom Base URL.

**Configuration:**
1. Set `OPENAI_API_KEY` to any non-empty string (or your service's key).
2. Set `OPENAI_BASE_URL` to your service endpoint.

**Example (Ollama):**
```bash
export OPENAI_API_KEY="ollama"
export OPENAI_BASE_URL="http://localhost:11434/v1"
```

**Example (GLM-4 via OpenRouter):**
Just use the OpenRouter configuration! OpenRouter hosts GLM models.

**Example (Self-Hosted GLM):**
If you are running GLM-4 locally with an OpenAI-compatible wrapper:
```bash
export OPENAI_API_KEY="your-local-key"
export OPENAI_BASE_URL="http://localhost:8000/v1"
```
