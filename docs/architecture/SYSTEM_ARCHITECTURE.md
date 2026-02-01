# System Architecture

## 1. High-Level Overview

Pryx is a sovereign AI agent system designed with a local-first architecture. It consists of three main components running on the user's machine, plus an optional mesh layer for multi-device coordination.

```mermaid
graph TD
    subgraph "Local Machine (User Device)"
        Host[Host App (Rust/Tauri)]
        Runtime[Core Runtime (Go)]
        TUI[Terminal UI (TypeScript)]
        Browser[Web Browser (Optional)]
        
        Host -- "Spawns & Manages" --> Runtime
        Host -- "JSON-RPC (Stdio)" --> Runtime
        Runtime -- "TCP (HTTP/WS)" --> TUI
        Runtime -- "TCP (HTTP/WS)" --> Browser
    end

    subgraph "Pryx Mesh (Optional)"
        Coordinator[Cloudflare Durable Object]
        Runtime -- "Secure WebSocket" --> Coordinator
    end
```

## 2. Components

### 2.1 Host (Apps/Host)
**Stack**: Rust, Tauri v2
**Responsibility**: Native OS integration and process management.
- Acts as the system supervisor.
- Spawns the Runtime as a child process (Sidecar).
- Provides native dialogs (permissions, file pickers) via JSON-RPC.
- Manages system tray and global shortcuts.
- Handles auto-updates.

### 2.2 Runtime (Apps/Runtime)
**Stack**: Go (Golang)
**Responsibility**: The brain of the agent.
- Runs the AI logic, tool execution, and provider management.
- Exposes a local HTTP/WebSocket API.
- Manages the SQLite database (`pryx.db`).
- Handles encryption and vault storage.
- Connects to the Mesh network.

### 2.3 TUI (Apps/TUI)
**Stack**: TypeScript, SolidJS, Ink/OpenTUI
**Responsibility**: The primary user interface.
- Connects to the Runtime via WebSocket.
- Renders a rich terminal interface.
- Completely decoupled from the Runtime (can run in a separate terminal or split pane).

## 3. Communication Protocols

### 3.1 Host ↔ Runtime (IPC)
**Mechanism**: Standard Input/Output (Stdio)
**Protocol**: JSON-RPC 2.0
- **Runtime → Host**: Requests permissions, notifications, clipboard access.
- **Host → Runtime**: Lifecycle signals (SIGTERM), configuration injection.

**Key RPC Methods**:
- `permission.request`: Ask user for sensitive action approval.
- `notification.show`: Display native OS notification.
- `clipboard.writeText` / `clipboard.readText`: Access system clipboard.
- `updater.check` / `updater.install`: Trigger self-update.

### 3.2 Runtime ↔ TUI (Client API)
**Mechanism**: TCP (Localhost)
**Protocol**: HTTP (REST) + WebSocket
- **Discovery**: Runtime prints its dynamic port to stdout; Host captures it (and TUI reads it from config/logs).
- **Events**: Real-time updates (logs, thinking process, status) via WebSocket.
- **Commands**: REST API for actions (start agent, stop, configure).

### 3.3 Runtime ↔ Mesh (Sync)
**Mechanism**: WebSocket (wss://)
**Protocol**: Custom Sync Protocol (see [Mesh Design](../../product/prd/pryx-mesh-design.md))
- Encrypted, authenticated connection to Cloudflare Durable Objects.
- Used for multi-device synchronization and remote control.

## 4. Data Flow

### 4.1 Startup Sequence
1. User launches **Host**.
2. Host reads configuration and finds **Runtime** binary.
3. Host spawns Runtime with dynamic port configuration (`PRYX_LISTEN_ADDR=127.0.0.1:0`).
4. Runtime starts, binds to a random port, and prints `PRYX_CORE_LISTEN_ADDR=127.0.0.1:54321`.
5. Host regex-parses stdout to find the port.
6. Host launches TUI (if configured) or User opens TUI manually pointing to that port.

### 4.2 Tool Execution Flow
1. **Agent** decides to run a tool (e.g., `fs.write`).
2. **Runtime** checks permissions in `pryx.db`.
3. If permission not granted:
   - Runtime sends `permission.request` JSON-RPC to **Host**.
   - Host shows Native Dialog to user.
   - User approves/denies.
   - Host replies with JSON-RPC response.
4. If approved, Runtime executes the tool.

## 5. Security Architecture

- **Local-First**: All data stored locally in SQLite (`pryx.db`).
- **Sandboxing**: Runtime runs as a standard user process; permissions are enforced by logic, not OS (though Host can kill Runtime).
- **Encryption**: Sensitive credentials (API keys) stored in `Vault`, encrypted with a master key derived from user password/keychain.
- **Network**: Runtime binds to `127.0.0.1` by default. Remote access only via secure Mesh.
