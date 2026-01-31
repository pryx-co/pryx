# Pryx

[![CI](https://github.com/irfndi/pryx/actions/workflows/ci.yml/badge.svg)](https://github.com/irfndi/pryx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/irfndi/pryx/branch/main/graph/badge.svg)](https://codecov.io/gh/irfndi/pryx)
[![Host Coverage](https://img.shields.io/codecov/c/github/irfndi/pryx?flag=host&label=host%20coverage)](https://codecov.io/gh/irfndi/pryx)
[![Runtime Coverage](https://img.shields.io/codecov/c/github/irfndi/pryx?flag=runtime&label=runtime%20coverage)](https://codecov.io/gh/irfndi/pryx)
[![TUI Coverage](https://img.shields.io/codecov/c/github/irfndi/pryx?flag=tui&label=tui%20coverage)](https://codecov.io/gh/irfndi/pryx)

Sovereign AI agent with local-first control center.

## Quick Start

This project is in early development. Build system is set up.

### Installation

1. **Install development tools**:
   ```bash
   make install-tools
   ```

2. **Install dependencies** (once component directories exist):
   ```bash
   make install-deps
   ```

### Development

See [BUILD_SYSTEM.md](BUILD_SYSTEM.md) for complete build system documentation.

### Architecture

- **Host** (Rust + Tauri v2) - Desktop wrapper with native dialogs
- **Runtime** (Go) - Core agent runtime with HTTP+WebSocket API
- **TUI** (TypeScript + Solid + OpenTUI) - Terminal UI surface
- **Web Apps** (Astro + React + Bun) - Planned edge-deployed services (telemetry/auth/installer)
- **Edge** (Cloudflare Workers) - Planned OAuth and telemetry workers

### Documentation

- [PRD](docs/prd/prd.md) - Product Requirements Document
- [PRD v2](docs/prd/prd-v2.md) - Roadmap and future features
- [Mesh Design](docs/prd/pryx-mesh-design.md) - Multi-device architecture
- [Build System](BUILD_SYSTEM.md) - Build and tooling documentation

### Status

- ✅ Build system & tooling set up
- ⏳ Component initialization pending
- ⏳ Implementation pending

## License

MIT
