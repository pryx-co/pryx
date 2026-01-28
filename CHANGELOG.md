# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] - 2026-01-28

### Added
- **Host Native Integration**:
    - `pryx-ntf`: Native OS Notifications support via JSON-RPC (`notification.show`).
    - `pryx-clp`: Clipboard Read/Write support via JSON-RPC (`clipboard.writeText`, `clipboard.readText`).
    - `pryx-td7`: Auto-Update mechanism via JSON-RPC (`updater.check`, `updater.install`).
    - `pryx-l27`: Deep Linking support. Host forwards `deeplink.opened` events to Runtime.
- **Sidecar Management**:
    - Robust process lifecycle management (start, stop, monitor).
    - Port discovery and health checking.
    - JSON-RPC over Stdio transport layer.

### Fixed
- Resolved concurrency (`Send`/`Sync`) issues in `SidecarProcess` to ensure thread-safe async operation.
- Fixed `tauri.conf.json` implementation to support new plugins.
- Refactored `sidecar.rs` to cleaner, modular RPC handling.
