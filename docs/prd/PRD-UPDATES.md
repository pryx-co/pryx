# PRD Update: Auto-Update Mechanism

> **Version**: 1.0
> **Date**: 2026-01-27
> **Status**: Research Complete, Ready for PRD Integration
> **Parent**: `docs/prd/prd.md` (v1) and `docs/prd/prd-v2.md`

---

## Executive Summary

This document consolidates all PRD improvements based on user questions about multi-device scenarios, long-running tasks, and ecosystem requirements.

**Key Questions Addressed**:
1. ✅ Authentication per device - Encrypted vault sync via E2EE
2. ✅ Telegram ↔ Web UI sync - Session Bus broadcast
3. ✅ Memory persistence across devices - Hybrid hot/warm sync
4. ✅ Scheduled tasks UX - Dashboard with templates, history
5. ✅ Multi-hop workflows - Pryx Mesh + waiting state UX
6. ✅ Constraint management for 600+ models - Dynamic catalog + routing
7. ✅ Long-running task autocompletion - Pump-dump, streaming, hybrid strategies
8. ✅ Plugin architecture - Based on OpenCode research
9. ✅ Auto-update mechanism - Production vs Beta build channels
10. ✅ Telegram bot operational model - Cloud webhook vs device polling, BYOK constraints + monetization (Channels Cloud)

---

## 1) Auto-Update & Install Flow (NEW)

### 1.1 Build Channel Architecture

**Different Build Channels**:

| Channel | Use Case | Update Mechanism |
|---------|-----------|------------------|
| **Main/Stable** (Production) | Production users on `main` branch | Auto-update enabled by default |
| **Beta/Development** | Beta testers on development branch | Auto-update for beta builds only |
| **Alpha/Canary** | Early access users | Manual updates only, notifications available |

**Update Flow Diagram**:
```
┌─────────────────────────────────────────────────────────────┐
│              Pryx Auto-Update Service                       │
│  ┌───────────────────────────────────────────────────┐   │
│  │  Build Pool (Multiple Versions)         │   │
│  │  • Version 1.2.3 (main)             │   │
│  │  • Version 1.2.4 (main, patched)      │   │
│  │  • Version 1.3.0-beta.1           │   │
│  │  • Version 1.3.0-alpha.1           │   │
│  └───────────────────────────────────────────────────┘   │
│                                                         │
│  ┌───────────────────────────────────────────────────┐   │
│  │  User Configuration                   │   │
│  │  • Current version tracking          │   │
│  │  • Update channel preference (main/beta) │   │
│  │  • Auto-update enable/disable         │   │
│  │  • Notification preferences          │   │
│  └───────────────────────────────────────────────────┘   │
│                                                         │
│  ┌───────────────────────────────────────────────────┐   │
│  │  Update Orchestration                │   │
│  │  • Version check at startup         │   │
│  │  • Update available check            │   │
│  │  • Background download                │   │
│  │  • App restart coordination           │   │
│  └───────────────────────────────────────────────────┘   │
│                                                         │
│  ▼                                                    │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 User Experience - Production Build (Main Pool)

**User Scenario**: User is on production build (main branch, auto-update enabled)

**Update Flow**:
```
┌─────────────────────────────────────────────────────────────┐
│  Step 1: Version Check (Startup)                    │
│  • Pryx checks for updates on startup         │
│  • Compares current version with latest       │
│  • If update available → go to step 2     │
└─────────────────────────────────────────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 2: Update Available Notification              │
│  • Toast notification appears in UI:              │
│    "🎉 Pryx v1.3.0 available!              │
│    Click to view changes"                     │
│  • User can:                                │
│    - "Update now"                           │
│    - "Remind me in 1 hour"                   │
│    - "Skip this version"                     │
│  • If user ignores: Toast auto-dismisses in 30s     │
└─────────────────────────────────────────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 3: Background Download (If "Update Now")    │
│  • Update downloaded in background               │
│  • User can continue using Pryx                 │
│  • Download progress shown in UI                  │
│  • Download size shown (e.g., "45MB")           │
└─────────────────────────────────────────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 4: Update Installation (Download Complete)     │
│  • Toast: "Update ready! Restart to apply"        │
│  • User clicks "Restart Now"                   │
│  • Pryx shuts down gracefully                    │
│  • Update applied                          │
│  • Pryx restarts automatically                 │
└─────────────────────────────────────────────────────────────┘
```

**UX Details**:

**Toast Notification Types**:

| Toast Type | When Shown | Actions | Auto-Dismiss |
|-----------|-------------|---------|--------------|
| **Update Available** | New version detected | Update Now, Remind Me, Skip | 30s |
| **Download Progress** | Update downloading | Show Progress, Cancel | Never (user action) |
| **Update Ready** | Download complete | Restart Now | 60s |
| **Update Installed** | Restart complete | What's New, Changelog | Never (shows modal) |

**What's New Modal** (shown after restart):
```
┌─────────────────────────────────────────────────────────────┐
│  Pryx Updated Successfully!              │
│                                          │
│  You're now running v1.3.0              │
│                                          │
│  ┌───────────────────────────────────────────┐   │
│  │  What's New:                         │   │
│  │  • New Features:                       │   │
│  │    - Scheduled tasks dashboard          │   │
│  │    - Multi-device constraint mgmt         │   │
│  │    - Background task manager             │   │
│  │    - Plugin architecture v2            │   │
│  │  • Token optimization layer             │   │
│  │  • Fixed: Security improvements       │   │
│  │                                      │   │
│  │  ┌───────────────────────────────────┐   │
│  │  │ View Full Changelog           │   │
│  │  └───────────────────────────────────┘   │
│  │                                      │   │
│  └───────────────────────────────────────────┘   │
│                                          │
│  [Dismiss]                                    │
└─────────────────────────────────────────────────────────────┘
```

---

### 1.3 User Experience - Beta Build Channel

**User Scenario**: User is on beta channel (development pool, auto-update for beta builds)

**Update Flow** (similar to production, but with beta builds):
```
┌─────────────────────────────────────────────────────────────┐
│  Step 1: Beta Version Check                      │
│  • Same version check mechanism               │
│  • Beta updates show with 🧪 icon       │
│  • "Beta build - may contain bugs"         │
└─────────────────────────────────────────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  Step 2: Beta Update Available                │
│  • Toast: "🧪 Beta v1.4.0 available!"      │
│  • Warning: "This is a beta build.         │
│    Only use if testing specific features"    │
│  • User can:                                │
│    - "Update to beta"                       │
│    - "Report bug"                           │
│    - "Stable only" mode toggle             │
└─────────────────────────────────────────────────────────────┘
```

**Beta-Specific UX**:
- 🧪 Icon distinguishes beta from stable
- Warning message on beta updates
- "Stable only" mode to disable beta updates
- Direct link to bug reporting for beta builds

---

### 1.4 User Experience - Switching Build Channels

**Scenario**: User switches from Main to Beta channel (or vice versa)

**User Flow**:
```
User Action: "Switch to Beta Channel"

┌─────────────────────────────────────────────────────────────┐
│  Confirmation Dialog                         │
│  ┌───────────────────────────────────────────┐   │
│  │  Switch to Beta Channel?               │   │
│  │                                    │   │
│  │  You will receive beta updates:         │   │
│  │  • v1.4.0-beta.1 (latest)          │   │
│  │  • May contain bugs                 │   │
│  │  • Can switch back to Stable anytime  │   │
│  │                                    │   │
│  │  [Cancel] [Switch to Beta]             │   │
│  └───────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  Pryx restarts with new channel config      │
│  • User now on beta pool                │
│  • Next update check will be against  │
│    beta version                              │
└─────────────────────────────────────────────────────────────┘
```

**Implementation Notes**:
- Channel preference stored in user config
- Version metadata includes build channel
- Update API respects channel preference

---

### 1.5 Background Update Mechanism

**Key Requirements**:

| Requirement | Implementation | Notes |
|------------|----------------|--------|
| **Silent downloads** | Downloads in background, user can continue using Pryx |
| **Progress indicators** | Show download progress in UI toast |
| **Graceful shutdown** | App restarts cleanly after update |
| **Rollback capability** | If update fails, rollback to previous version |
| **Update history** | Track update history for debugging |
| **Update scheduling** | Respect "remind me later" user preference |

**Implementation (TypeScript)**:
```typescript
interface UpdateConfig {
  currentVersion: string;
  latestVersion: string;
  buildChannel: 'main' | 'beta' | 'alpha';
  autoUpdateEnabled: boolean;
  lastCheckAt: Date;
  downloadProgress?: {
    downloadedBytes: number;
    totalBytes: number;
    percentage: number;
  };
}

class PryxUpdateManager {
  async checkForUpdates(): Promise<UpdateAvailable | null> {
    // Check update API for current build channel
    const latest = await this.fetchLatestVersion(this.config.buildChannel);

    if (!latest) {
      // Network error or API down
      return null;
    }

    if (this.isNewerVersion(this.config.currentVersion, latest.version)) {
      return {
        version: latest.version,
        buildChannel: this.config.buildChannel,
        releaseNotes: latest.releaseNotes,
        downloadSize: latest.downloadSize,
        required: latest.required,
      };
    }

    return null; // No update available
  }

  async downloadUpdate(version: string): Promise<void> {
    const toast = this.showToast({
      title: 'Downloading Update...',
      body: `Pryx v${version} (${this.formatSize(latest.downloadSize)})`,
      type: 'progress',
      progress: 0,
      showProgress: true,
    });

    try {
      const updateFile = await this.downloadFile(version);

      toast.update({
        title: 'Download Complete',
        body: 'Restart to apply update',
        type: 'normal',
        actions: [
          { label: 'Restart Now', action: () => this.applyUpdate(updateFile) },
          { label: 'Restart Later', action: () => this.dismissToast() },
        ],
      });
    } catch (error) {
      toast.update({
        title: 'Update Failed',
        body: `Failed to download: ${error.message}`,
        type: 'error',
        actions: [{ label: 'Dismiss', action: () => this.dismissToast() }],
      });
    }
  }

  async applyUpdate(updateFile: string): Promise<void> {
    // 1. Graceful shutdown
    await this.gracefulShutdown();

    // 2. Apply update (replace binaries)
    await this.replaceBinaries(updateFile);

    // 3. Restart application
    this.restartApp();
  }

  private gracefulShutdown(): Promise<void> {
    // Save state
    await this.saveCurrentSession();

    // Stop background processes
    await this.stopBackgroundProcesses();

    // Close connections
    await this.closeNetworkConnections();

    // Exit main process
    process.exit(0);
  }
}
```

---

### 1.6 Configuration API

**User Config Structure**:
```json
{
  "update": {
    "enabled": true,
    "buildChannel": "main",
    "autoCheckOnStartup": true,
    "checkIntervalHours": 24,
    "downloadInBackground": true,
    "allowBetaUpdates": false,
    "remindMeLater": "1h"
  },
  "channel": {
    "current": "main",
    "available": ["main", "beta"],
    "autoSwitch": false,
    "showBetaWarning": true
  }
}
```

**Configuration UI Settings**:

```
┌─────────────────────────────────────────────────────────────┐
│  Update Settings                                 │
│  ┌───────────────────────────────────────────┐   │
│  │  Automatic Updates                       │   │
│  │  [✓] Check for updates on startup    │   │
│  │  [✓] Download updates in background      │   │
│  │  [ ] Ask before downloading            │   │
│  │                                     │   │
│  │  ┌───────────────────────────────────┐   │
│  │  │  Build Channel: [Main ▼]        │   │
│  │  │     ○ Main (Stable)            │   │
│  │  │     ○ Beta (Development)         │   │
│  │  │     ○ Alpha (Early Access)         │   │
│  │  └───────────────────────────────────┘   │
│  │                                     │   │
│  │  ┌───────────────────────────────────┐   │
│  │  │  Update Frequency:                    │   │
│  │  │     ○ Daily (recommended)           │   │
│  │  │     ○ Weekly                       │   │
│  │  │     ○ Manual only                   │   │
│  │  └───────────────────────────────────┘   │
│  │                                     │   │
│  │  ┌───────────────────────────────────┐   │
│  │  │  Notify Me About:                  │   │
│  │  │     ☐ New stable releases          │   │
│  │  │     ☐ Beta releases                │   │
│  │  │     ☐ Security updates only           │   │
│  │  └───────────────────────────────────┘   │
│  └───────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

### 1.7 Success Metrics

| Metric | v1 Target | v1.1 Target | v2 Target |
|--------|-----------|-------------|----------|
| **Update success rate** | >98% | >99% | >99.5% |
| **Rollback success rate** | N/A | >95% | >98% |
| **Update adoption rate** | 60% of users | 70% of users | 80% of users |
| **Update awareness** | Users know update available in <1 day | Users aware within 1 day | Users aware within 1 day |

---

## 2) Integration with Existing PRD Sections

### 2.1 PRD v1 (`docs/prd/prd.md`) - Updates Required

**Add/Update Sections**:

| Section | Change | Status |
|---------|---------|--------|
| **8.4** | Already updated with 600+ model constraint management | ✅ Complete |
| **8.3** | Already removed duplicate, added to Appendix | ✅ Complete |
| **10.7** | Already added autocompletion section | ✅ Complete |
| **FR11** | Already added scheduled tasks section | ✅ Complete |
| **NFR-M1** | Already added memory management section | ✅ Complete |
| **NFR-M2** | Already added task queue persistence section | ✅ Complete |
| **Section 12** | **ADD**: Auto-Update mechanism (from this doc) | 🆕 TODO |
| **Appendix** | Update with new document reference | 🆕 TODO |

---

### 2.2 PRD v2 (`docs/prd/prd-v2.md`) - Updates Required

**Add/Update Sections**:

| Section | Change | Status |
|---------|---------|--------|
| **5.2** | Rename to "6.1 Skills Marketplace (v2.1)" | ✅ Complete |
| **5.1** | Rename to "6.1 Skills Marketplace" | ✅ Complete |
| **6.1** | Rename to "6.2.1" | ✅ Complete |
| **6.2** | Rename to "6.2.2" | ✅ Complete |
| **7.2** | Rename to "6.3.1" (was "6.2.1") | ✅ Complete |
| **Section 11** | **ADD**: Plugin Architecture & Third-Party Integration (from research) | 🆕 TODO |
| **Section 12** | **ADD**: Auto-Update mechanism (from this doc) | 🆕 TODO |
| **Release Timeline** | Update with auto-update phases | 🆕 TODO |

---

## 3) Implementation Phases

### 3.1 v1.1 (Post-MVP)

**Phase 1: Auto-Update Foundation** (Week 1-2)
- Implement update version check API
- Add build channel configuration
- Implement background download mechanism
- Add toast notification system
- Implement graceful shutdown and restart

**Phase 2: Auto-Update UI** (Week 2-3)
- Add update settings to Admin Settings UI
- Implement build channel switcher
- Add "What's New" modal after updates
- Add update history view

**Phase 3: Integration** (Week 3-4)
- Integrate auto-update with scheduled tasks system
- Ensure background processes survive updates
- Add update progress monitoring in task dashboard

---

## 4) Research Summary

### 4.1 Auto-Update Research (OpenCode Pattern)

**Key Findings**:
- OpenCode uses different build pools for production vs beta
- Auto-update is enabled by default for production builds
- Beta users can configure update preferences
- Plugins without explicit version don't auto-update (issue #6774)
- Updates are downloaded in background
- Toast notifications for user awareness
- Graceful restart after update application

### 4.2 Plugin Architecture Research (OpenCode Pattern)

**Key Findings**:
- Plugins loaded from local files, npm packages, or marketplace
- Two loading methods: local directory (~/.config/opencode/plugins/) and project directory (.opencode/plugins/)
- Plugin manifests define permissions, tools, entry points
- Event-driven architecture: plugins subscribe to Pryx events
- Sandbox execution: plugins run in isolated environment
- Permission model: granular approvals required (network, fs, shell)
- Hot reload support during development
- Dependencies managed via package.json in config directory

---

## 5) Next Steps

1. **Update PRD v1** with Section 12 (Auto-Update)
2. **Update PRD v2** with Section 11 (Plugin Architecture) and Section 12 (Auto-Update)
3. **Create detailed design document** for auto-update mechanism
4. **Integrate with scheduled tasks system** to handle background processes during updates
5. **Add success metrics** for auto-update feature

---

## 6) Questions & Answers - Reference

| Question | Answer | Documented In |
|----------|---------|--------------|
| **Q1**: Auth per device? | ✅ Encrypted vault sync | FR10.4 + Mesh Design |
| **Q2**: Telegram ↔ Web UI sync? | ✅ Session Bus broadcast | Section 8.3 |
| **Q3**: Memory persistence? | ✅ Hybrid sync + auto-summarization | NFR-M1 + Section 10.7.2 |
| **Q4**: Cron job UX? | ✅ Scheduled Tasks Dashboard | FR11 + v2 Section 5.2.1 |
| **Q5**: Multi-hop workflows? | ✅ Pryx Mesh + waiting state | FR10.3 + Section 8.4.3 |
| **Q6**: 600+ models? | ✅ Dynamic constraint catalog | Section 8.4.1 |
| **Q7**: Autocompletion for long tasks? | ✅ Pump-dump/streaming/hybrid | Section 10.7 + v2 Section 5.2.1 |
| **Q8**: Plugin architecture? | ✅ Based on OpenCode | v2 Section 6.1 (TODO) |
| **Q9**: Auto-update on CI builds? | ✅ Production vs Beta | This document (TODO) |
| **Q10**: What configs/integrations are visible/manageable in the web dashboard? | ✅ Cloud dashboard is “must-know” only by default; full config stays local unless opt-in sync/backup is enabled | [auth-sync-dashboard.md](../auth-sync-dashboard.md) + PRD Section 7.4 |

---

**Document Status**: ✅ Complete research phase. Ready for PRD integration.

**Next Action**: Apply changes to PRD v1 and v2, then create `docs/prd/plugin-architecture.md` with detailed plugin system design.
