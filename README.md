# Pryx

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
- **UI** (React + TypeScript + Vite) - Web-based control center
- **Edge** (Cloudflare Workers) - OAuth and telemetry workers

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
