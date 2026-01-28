# Pryx Build System & Tooling

This directory contains the build system infrastructure for the Pryx project.

## Quick Start

```bash
# Install development tools
make install

# Install dependencies
make install-deps

# Run all checks (lint + test)
make check

# Build all components
make build
```

## Available Targets

### Build Targets
- `make build` - Build all components (host, runtime, tui)
- `make build-host` - Build Rust/Tauri host
- `make build-runtime` - Build Go runtime
- `make build-tui` - Build TypeScript TUI

### Test Targets
- `make test` - Run all tests
- `make test-host` - Run Rust tests
- `make test-runtime` - Run Go tests
- `make test-tui` - Run TypeScript TUI checks

### Lint Targets
- `make lint` - Run all linters
- `make lint-host` - Run Rust linters (rustfmt, clippy)
- `make lint-runtime` - Run Go linters (gofmt, golangci-lint)
- `make lint-tui` - Run TypeScript TUI linters (oxlint, oxfmt)

### Format Targets
- `make format` - Format all code
- `make format-host` - Format Rust code (rustfmt)
- `make format-runtime` - Format Go code (gofmt)
- `make format-tui` - Format TypeScript TUI code (oxfmt)

### Clean Targets
- `make clean` - Clean all build artifacts
- `make clean-host` - Clean Rust build artifacts
- `make clean-runtime` - Clean Go build artifacts
- `make clean-tui` - Clean TypeScript TUI build artifacts

### Install Targets
- `make install` - Install all development tools
- `make install-tools` - Install development tools (pre-commit, Bun, golangci-lint, Tauri CLI)
- `make install-deps` - Install all dependencies

### Version Management
- `make version-bump-patch` - Bump patch version (0.1.0 → 0.1.1)
- `make version-bump-minor` - Bump minor version (0.1.0 → 0.2.0)
- `make version-bump-major` - Bump major version (0.1.0 → 1.0.0)
- `make version-tag` - Create git tag for current version
- `make version-push` - Push version tags

### Info Targets
- `make info` - Show project information (version, build date, tools)

## Continuous Integration

GitHub Actions workflow (`.github/workflows/ci.yml`) will:
1. Run linters on all components
2. Run tests on all components
3. Build artifacts for all platforms (Linux, macOS, Windows)
4. Run security audits

## Pre-commit Hooks

Pre-commit hooks are configured in `.pre-commit-config.yaml` and will run automatically on `git commit`.

Installed hooks:
- **General**: Check merge conflicts, large files, JSON/TOML/YAML syntax, trailing whitespace
- **Rust**: cargo fmt, cargo check, clippy
- **Go**: gofmt, golangci-lint
- **JavaScript/TypeScript**: oxfmt, oxlint

To manually run all pre-commit hooks:
```bash
pre-commit run --all-files
```

## Version Management

Version is managed through the `VERSION` file and the `scripts/bump-version.sh` script.

**DO NOT** edit the VERSION file manually. Use the make targets:
```bash
# Bump version (creates VERSION update and CHANGELOG entry)
make version-bump-patch
make version-bump-minor
make version-bump-major

# After bumping and updating CHANGELOG.md:
git add VERSION CHANGELOG.md
git commit -m "chore: bump version to X.Y.Z"

# Create git tag
make version-tag

# Push tags
make version-push
```

## Development Workflow

1. **Initial Setup**
   ```bash
   make install    # Install tools
   make install-deps   # Install dependencies
   ```

2. **During Development**
   ```bash
   # Before committing
   make lint       # Check all code
   make test        # Run all tests

   # Format if linter complains
   make format

   # Commit (pre-commit hooks will run automatically)
   git commit -m "feat: add new feature"
   ```

3. **Before Merging PR**
   ```bash
   make check      # Run all checks (lint + test)
   ```

4. **For Release**
   ```bash
   make clean      # Clean all artifacts
   make build      # Build all components
   make version-bump-minor  # Bump version
   git add VERSION CHANGELOG.md
   git commit -m "chore: bump version to X.Y.Z"
   make version-tag
   git push origin main --tags
   ```

## Tool Versions

Based on 2026-01-27 research:

| Tool | Version | Source |
|-------|---------|--------|
| Pre-commit | 4.5.1 | https://github.com/pre-commit/pre-commit/releases |
| Tauri CLI | 2.8.0 | https://github.com/tauri-apps/tauri/releases |
| Wrangler | latest | https://developers.cloudflare.com/workers/wrangler/ |
| Bun | 1.3.7 | https://bun.sh/ |
| Rust | stable | https://rust-lang.org/ |
| Go | 1.22 | https://go.dev/ |
| Node.js | 22 LTS | https://nodejs.org/ |

## Troubleshooting

### Pre-commit hooks not running
```bash
# Uninstall and reinstall pre-commit
pip uninstall pre-commit
make install-tools

# Install hooks
pre-commit install
```

### Clippy warnings
```bash
# Run clippy with suggestions
make lint-host

# Auto-fix simple issues
cd apps/host
cargo clippy --fix
```

### Go formatting issues
```bash
# Format all Go files
make format-runtime
```

## Project Structure

```
silent-river/
├── .github/
│   └── workflows/
│       └── ci.yml          # GitHub Actions CI pipeline
├── .pre-commit-config.yaml   # Pre-commit hooks configuration
├── Makefile              # Main build system
├── VERSION               # Current version (managed by bump script)
├── CHANGELOG.md          # Release notes (managed by bump script)
├── scripts/
│   └── bump-version.sh  # Version bumping script
├── apps/
│   ├── host/            # Rust + Tauri host
│   ├── runtime/         # Go runtime (pryx-core)
│   ├── tui/             # TypeScript + Solid + OpenTUI
│   └── web/             # (Planned) Web dashboard (user + superadmin)
├── workers/
│   └── edge/            # (Planned) Cloudflare Worker (auth/telemetry/installer)
├── packages/            # (Planned) Shared cross-surface libraries (protocol/config)
└── deploy/              # (Planned) Deployment assets (compose + edge)
```
