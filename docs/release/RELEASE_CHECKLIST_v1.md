# Pryx v1.0.0 Release Checklist

> **Target Date**: 2026-02-15  
> **Version**: 1.0.0  
> **Status**: In Progress  
> **Release Manager**: TBD  

---

## ✅ Pre-Release Verification

### Core Functionality
- [x] Runtime builds successfully on all platforms
- [x] TUI builds and runs
- [x] Host (Rust) compiles
- [x] Provider management works (84 providers via models.dev)
- [x] Vault security (Argon2id, scopes, audit logging)
- [x] Agent spawning functional
- [x] Telegram, Discord, Slack channels implemented
- [x] Input validation & security audit passed
- [x] Secret scanning clean

### Testing
- [x] Unit tests passing (vault, channels, agent)
- [ ] E2E tests for critical paths (16 scenarios defined)
- [ ] Integration tests with real providers (optional for v1)
- [ ] Performance benchmarks (optional for v1)

### Documentation
- [ ] README.md updated with v1 features
- [ ] Installation guide (all platforms)
- [ ] Quick start tutorial
- [ ] Configuration reference
- [ ] Troubleshooting guide
- [ ] Changelog (v1.0.0)

---

## 🔧 Build & Distribution

### Binaries
- [ ] **macOS** (Intel)
  - [ ] Build: `GOOS=darwin GOARCH=amd64`
  - [ ] Notarize with Apple
  - [ ] Create .dmg installer
  
- [ ] **macOS** (Apple Silicon)
  - [ ] Build: `GOOS=darwin GOARCH=arm64`
  - [ ] Notarize with Apple
  - [ ] Create .dmg installer
  
- [ ] **Linux** (x64)
  - [ ] Build: `GOOS=linux GOARCH=amd64`
  - [ ] Create .deb package (Ubuntu/Debian)
  - [ ] Create .rpm package (Fedora/RHEL)
  - [ ] Create AppImage
  
- [ ] **Windows** (x64)
  - [ ] Build: `GOOS=windows GOARCH=amd64`
  - [ ] Sign with Windows certificate
  - [ ] Create .msi installer
  - [ ] Create .exe installer

### Distribution Channels
- [ ] GitHub Releases page
- [ ] Homebrew formula (macOS/Linux)
- [ ] Scoop bucket (Windows)
- [ ] APT repository (Ubuntu/Debian)
- [ ] YUM repository (Fedora/RHEL)
- [ ] Website download page (opencode.ai)

---

## 📝 Release Notes (v1.0.0)

### New Features
- **Multi-Provider Support**: Dynamic integration with 84 AI providers via models.dev
- **Secure Vault**: Argon2id password derivation with scope-based access control
- **Multi-Channel**: Telegram, Discord, and Slack bot integrations
- **Agent Spawning**: Create and manage sub-agents for parallel task execution
- **TUI Interface**: Rich terminal interface with provider management (`/connect`)
- **Observability**: OTLP telemetry export and comprehensive audit logging
- **NLP Parser**: Natural language command parsing with intent recognition

### Security
- Zero-friction onboarding with secure defaults
- Input validation and command injection prevention
- Secrets stored in OS keychain (not plaintext)
- Dependency vulnerability scanning
- Pre-commit secret scanning

### Architecture
- Polyglot: Rust (host), Go (runtime), TypeScript (TUI)
- Local-first: All data stays on device
- WebSocket mesh for multi-device coordination
- Sidecar architecture for crash isolation

---

## 🚀 Post-Release

### Immediate (Week 1)
- [ ] Monitor GitHub issues
- [ ] Respond to community feedback
- [ ] Fix critical bugs (if any)
- [ ] Update documentation based on feedback

### Short Term (Month 1)
- [ ] Implement top 5 requested features
- [ ] Improve E2E test coverage
- [ ] Add WhatsApp channel
- [ ] Create video tutorials

### Medium Term (Quarter 1)
- [ ] Plugin architecture
- [ ] Auto-update mechanism
- [ ] Web UI for headless servers
- [ ] Desktop wrapper (Tauri)

---

## 📊 Success Metrics

### Adoption
- Target: 1,000 downloads in first month
- Target: 100 GitHub stars in first month
- Target: 50 active users (telemetry opt-in)

### Quality
- Bug reports: < 10 critical in first week
- Crash rate: < 1%
- Test coverage: > 70%

### Community
- Discord/Slack community: 100 members
- Contributing guidelines followed
- First community PR merged

---

## 🆘 Emergency Contacts

| Role | Contact | Responsibility |
|------|---------|----------------|
| Release Manager | TBD | Overall coordination |
| Build Engineer | TBD | Binary builds & signing |
| Docs Lead | TBD | Documentation updates |
| Community Manager | TBD | User support |

---

## 📚 References

- PRD: `docs/prd/prd.md`
- Roadmap: `docs/prd/implementation-roadmap.md`
- Architecture: `docs/prd/pryx-mesh-design.md`
- Security Audit: `docs/security/`

---

**Last Updated**: 2026-01-30  
**Next Review**: 2026-02-01  
