# Contributing to Pryx

## Prerequisites

- Rust 1.70+
- Go 1.22+
- Node.js 22+ (for web components)
- Bun 1.3.7

## Initial Setup

1. Clone repository
2. Run `make install-tools`
3. Run `make install-deps`
4. Configure your LLM provider (see CONFIGURATION.md)

## Development Workflow

### Running Locally

```bash
# Start TUI + Runtime
make dev-tui

# Start all components
make dev

# View logs
make logs
```

### Testing

```bash
# Run all tests
make test

# Run specific component tests
make test-runtime
make test-tui
make test-host
```

### Code Style

- **Go**: Follows standard Go formatting (`gofmt`).
- **TypeScript**: Uses Prettier and ESLint.
- **Rust**: Uses `cargo fmt`.
