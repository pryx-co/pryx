# Contributing to Pryx

Thank you for your interest in contributing to Pryx! We welcome contributions from the community.

## Quick Start

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes
4. Commit with conventional commits: `git commit -m "feat: add awesome feature"`
5. Push to your fork: `git push origin feature/your-feature-name`
6. Open a pull request

## Development Setup

### Prerequisites

- Go 1.24+
- Node.js 20+
- Bun for package management
- Rust 1.70+ (for host development)
- Make

### Install Dependencies

```bash
make install
```

### Running the Development Stack

```bash
make dev          # Run TUI + Runtime + Host (local development)
make dev-tui      # Run TUI + Runtime
make dev-tail     # Tail runtime logs while TUI is running
```

## Project Structure

```
neon-star/
├── apps/
│   ├── host/          # Rust + Tauri desktop wrapper
│   ├── runtime/       # Go runtime with agents, channels, MCP
│   ├── tui/           # TypeScript TUI (SolidJS)
│   ├── web/           # Astro web app (cloud deployment)
│   └── local-web/     # Local web UI (served by host)
├── docs/              # Project documentation
├── packages/          # Shared TypeScript packages
└── skills/           # External skills
```

## Code Style

### Go
- Follow [Effective Go](https://effectivego.com/) guidelines
- Use `gofmt` for formatting
- Run `make lint-runtime` before committing
- Write tests for new features

### TypeScript/TUI
- Use solid-js patterns
- Use `make lint-tui` before committing
- Follow existing component structure

### Rust
- Use `cargo fmt` for formatting
- Run `make lint-host` before committing

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Test additions/changes
- `chore:` Maintenance tasks

Example:
```
feat(channels): add Discord channel support
fix(runtime): resolve memory leak in agent spawning
docs(readme): update installation instructions
```

## Testing

Run the test suite before submitting:

```bash
make test           # Run all tests
make test-coverage # Run tests with coverage
```

## Pull Request Guidelines

- Keep PRs focused and atomic
- Include tests for new features
- Update documentation for API changes
- Link related issues
- Ensure CI checks pass

## Architecture Notes

Pryx follows a **polyglot architecture**:
- **Rust Host** - Desktop wrapper, local web UI, IPC bridge
- **Go Runtime** - Agents, channels, MCP, memory, vault
- **TypeScript TUI** - Terminal interface

### Key Design Principles

1. **Local-First**: All data stays local unless explicitly synced
2. **Sovereign Security**: Credentials encrypted in vault/keychain
3. **Event-Driven**: Communication via event bus
4. **Extensible**: MCP protocol for tools

### Component Boundaries

| Component | Responsibility | Protocol |
|-----------|---------------|----------|
| Host | HTTP server, local web UI, updates | IPC (JSON-RPC) to Runtime |
| Runtime | Agents, LLM, channels, MCP, memory, vault | HTTP API to Host |
| TUI | Terminal interface, keyboard navigation | WebSocket to Runtime |

## Communication

- Join [Discord](https://discord.gg/pryx) for questions
- Check [Issues](https://github.com/irfndi/pryx/issues) before starting work
- Discuss large changes in an issue first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
