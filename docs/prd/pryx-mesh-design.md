# Pryx Mesh Architecture Design

> **Status**: Draft v1
> **Parent**: `docs/prd/prd.md`
> **Context**: Defines the multi-device coordination layer ("Pryx Mesh")

---

## 1. Overview

Pryx Mesh is the distributed coordination layer that enables **"Control Everything from Anywhere."** It connects multiple Pryx instances (Nodes) into a secure, synchronized personal botnet.

**Core Philosophy**:
- **Hub-and-Spoke Topology**: A Cloudflare Durable Object acts as the always-on "Coordinator" (Session Bus).
- **Local Execution**: Tools run locally on Nodes; the Coordinator only routes commands and events.
- **Defense-in-Depth**: Every cross-device command requires cryptographically signed authorization.

---

## 2. Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                  Cloudflare Edge (Coordinator)               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Mesh Durable Object (per user)                       │  │
│  │  - Session Bus (Pub/Sub)                              │  │
│  │  - Device Registry (State)                            │  │
│  │  - Presence Monitor (Heartbeat)                       │  │
│  └───────────────────────────────────────────────────────┘  │
└──────────────────────────────▲──────────────────────────────┘
                               │
             WebSocket (Secure, Persistent)
                               │
      ┌────────────────────────┴────────────────────────┐
      ▼                                                 ▼
┌─────────────┐                                   ┌─────────────┐
│   Node A    │                                   │   Node B    │
│  (Laptop)   │                                   │  (Server)   │
│             │                                   │             │
│ [Generated  │                                   │ [Generated  │
│  Identity]  │                                   │  Identity]  │
└─────────────┘                                   └─────────────┘
```

### 2.1 Device Identity (The Trust Root)
Every Pryx installation generates a permanent identity on first run.
- **Algorithm**: ED25519
- **Storage**: OS Keychain (private key), D1 Registry (public key)
- **ID Format**: `did:pryx:<sha256-fingerprint>`

### 2.2 The Coordinator (Cloudflare Durable Object)
The "Hub" that devices connect to. It does **not** store history or heavy data.
- **Responsibilities**: 
  - Routes events between devices (`tool.request`, `session.sync`)
  - Stores the "Active Session" pointer
  - Manages device presence (Online/Offline)

---

## 3. Pairing Protocol (The Handshake)

We use a **Device Flow (RFC 8628)** adaptation for pairing new devices to the Mesh.

**Scenario**: User has Laptop (logged in) and wants to add Server.

1.  **Server (New Device)**:
    - Generates Ephemeral Keypair.
    - Displays `user_code` (e.g., "ABCD-1234").
    - Polls Coordinator for auth token.

2.  **Laptop (Authenticated Device)**:
    - User enters `user_code`.
    - Laptop signs `approve_request` with its Private Key.
    - Sends signed approval to Coordinator.

3.  **Coordinator**:
    - Validates Laptop's signature.
    - Issues Long-Lived Mesh Token (encrypted with Server's public key) to Server.
    - Adds Server to Device Registry.

---

## 4. Multi-Device Session Sync

**Problem**: User chats on Phone, switches to Laptop. History must follow.

### 4.1 Sync Strategy: Hybrid Hot/Warm
- **Hot State (Active Session)**: Synced via WebSocket Broadcast in real-time (latency < 100ms).
- **Warm State (History)**: Stored in SQLite (local) + synced to Cloud D1 (encrypted blob) on idle.

### 4.2 Conflict Resolution: "Coordinator Time"
The Durable Object acts as the logical clock.
- Every event gets a monotonic `sequence_id` from the Coordinator.
- Devices apply events in sequence.
- If a device reconnects, it requests "all events since `last_sequence_id`".

### 4.3 Detailed Conflict Resolution Scenarios

#### Scenario A: Both Devices Edit Same Setting (Last-Write-Wins)

**Situation**: User on Laptop changes model to "claude-sonnet-4", while Server changes it to "claude-opus-3" at the same time.

**Resolution Algorithm**:
```
┌─────────────────────────────────────────────────────────┐
│  Conflict: Simultaneous Model Setting Changes        │
├─────────────────────────────────────────────────────────┤
│                                                     │
│  Event Flow:                                       │
│                                                     │
│  T=00:00:01.123 - Laptop sends:               │
│    {event: "config.update",                         │
│     model: "claude-sonnet-4",                      │
│     sequence_id: 1001}                             │
│                                                     │
│  T=00:00:01.087 - Server sends:                │
│    {event: "config.update",                         │
│     model: "claude-opus-3",                        │
│     sequence_id: 1002}                             │
│                                                     │
│  Coordinator Process:                                 │
│  1. Receive Laptop event (seq 1001)                │
│  2. Apply to Coordinator state                        │
│  3. Broadcast to all devices                         │
│  4. Receive Server event (seq 1002)                 │
│  5. Apply to Coordinator state (overrides seq 1001)   │
│  6. Broadcast to all devices                         │
│                                                     │
│  Laptop Receives:                                   │
│  - Event seq 1002 (from Server)                    │
│  - Applies change: model = "claude-opus-3"         │
│  - Shows notification: "Server overwrote your change"  │
│                                                     │
│  Server Receives:                                   │
│  - No action needed (its event won)                    │
│                                                     │
└─────────────────────────────────────────────────────────┘
```

**Resolution**: Last-write-wins based on `sequence_id` from Coordinator.

**User Notification**: Toast message on losing device:
```
⚠️ Model setting changed on another device

Your change was overwritten:
  Your setting: claude-sonnet-4
  Current: claude-opus-3 (changed by Server)

[View change history]
```

---

#### Scenario B: Both Devices Try to Execute Conflicting Commands (Lock Acquisition)

**Situation**: Laptop attempts to run `docker build` while Server is already running a conflicting `docker-compose down`.

**Resolution Algorithm**:
```
┌─────────────────────────────────────────────────────────┐
│  Conflict: Conflicting Docker Operations            │
├─────────────────────────────────────────────────────────┤
│                                                     │
│  T=00:00:01.100 - Server initiates:          │
│    shell.exec("docker-compose down", target=server)   │
│    → Coordinator grants lock for "docker:server"    │
│                                                     │
│  T=00:00:01.150 - Laptop initiates:          │
│    shell.exec("docker build", target=server)         │
│    → Coordinator checks lock "docker:server"          │
│    → Lock held, queue request                     │
│                                                     │
│  Server completes at T=00:00:05.234:             │
│    → Releases lock "docker:server"                   │
│    → Coordinator dequeues Laptop's request           │
│    → Grants lock to Laptop                          │
│                                                     │
│  Laptop receives lock at T=00:00:05.256:          │
│    → Executes "docker build"                       │
│    → Completes at T=00:00:45.123                 │
│                                                     │
└─────────────────────────────────────────────────────────┘
```

**Lock Schema** (in Durable Object):
```typescript
interface DistributedLock {
    resource_id: string;     // "docker:server", "file:/path/to/file"
    holder_device_id: string; // Which device holds the lock
    acquired_at: number;     // Timestamp
    expires_at: number;      // Lock timeout (prevents deadlock)
}

// Lock acquisition
function acquireLock(resourceId: string, deviceId: string): boolean {
    const currentLock = getLock(resourceId);

    if (!currentLock || isExpired(currentLock)) {
        setLock(resourceId, {
            holder_device_id: deviceId,
            acquired_at: Date.now(),
            expires_at: Date.now() + 300000 // 5 min timeout
        });
        return true;
    }

    return false;
}
```

**Resolution**: Lock acquisition at Coordinator level. One device queues while other holds lock.

**User Experience**:
- Initiating device sees "Waiting for lock..." indicator
- Completing device releases lock automatically
- Queued command executes when lock available

---

#### Scenario C: Disconnected Devices Reconnect with Divergent State (Three-Way Merge)

**Situation**: Laptop and Server both disconnected. Both made changes independently. Reconnect with different state.

**Resolution Algorithm**:
```
┌─────────────────────────────────────────────────────────┐
│  Conflict: Divergent State After Disconnect      │
├─────────────────────────────────────────────────────────┤
│                                                     │
│  Timeline:                                         │
│                                                     │
│  T=00:00:00 - Both devices online, synced          │
│  T=00:00:10 - Coordinator goes offline (outage)    │
│  T=00:00:15 - Laptop adds session: "Code Review"   │
│    [Local DB: session_5 = {title: "Code Review"}]  │
│  T=00:00:20 - Server adds session: "Bug Fix"       │
│    [Local DB: session_6 = {title: "Bug Fix"}]      │
│  T=00:00:25 - Laptop modifies session_2 title        │
│    [Local DB: session_2.title = "Refactor"]       │
│  T=00:00:30 - Server modifies session_2 title        │
│    [Local DB: session_2.title = "Optimize"}       │
│                                                     │
│  T=00:05:00 - Coordinator back online                │
│  T=00:05:01 - Both devices reconnect               │
│                                                     │
│  Sync Process (Laptop):                             │
│  1. Request events since last_sequence_id=500         │
│  2. Coordinator returns:                             │
│     - Server events: [501, 502, 503]            │
│     - Laptop events: [501, 502, 503]            │
│  3. Apply Server events locally:                       │
│     - Add session_6 ("Bug Fix")                       │
│     - session_2.title = "Optimize" (Server's wins)  │
│  4. Notify user: "3 changes synced from Server"     │
│                                                     │
│  Sync Process (Server):                             │
│  1. Request events since last_sequence_id=500         │
│  2. Coordinator returns same events                   │
│  3. Apply Laptop events locally:                      │
│     - Add session_5 ("Code Review")                   │
│     - session_2.title = "Optimize" (Server's wins)  │
│     [Server was source of truth for this conflict]   │
│  4. Notify user: "2 changes synced from Laptop"     │
│                                                     │
└─────────────────────────────────────────────────────────┘
```

**Merge Algorithm (Three-Way)**:
```typescript
interface MergeConflict {
    base: any;      // State at disconnect
    local: any;      // Device A's changes
    remote: any;     // Device B's changes
}

function resolveConflict(conflict: MergeConflict): any {
    // Strategy 1: Last-Write-Wins (by timestamp)
    if (conflict.remote.timestamp > conflict.local.timestamp) {
        return conflict.remote;
    }
    return conflict.local;

    // Strategy 2: Conflict markers (for text fields)
    if (typeof conflict.local === 'string') {
        return `<<<<<<< LOCAL\n${conflict.local}\n=======\n${conflict.remote}\n>>>>>>> REMOTE`;
    }

    // Strategy 3: Union (for array fields)
    if (Array.isArray(conflict.local)) {
        return [...new Set([...conflict.local, ...conflict.remote])];
    }

    // Strategy 4: User choice (via UI)
    return promptUserForResolution(conflict);
}
```

**Resolution**: Three-way merge algorithm using:
1. **Coordinator Time**: Last-write-wins by `sequence_id` (default)
2. **Conflict Markers**: For text fields (similar to Git)
3. **Union Merge**: For array/collection fields
4. **User Choice**: Prompt user to resolve manual conflicts

**User Experience**:
```
┌─────────────────────────────────────────────────────────┐
│  ⚠️ Merge Conflict Detected                       │
├─────────────────────────────────────────────────────────┤
│                                                     │
│  Session: "Project Planning"                      │
│                                                     │
│  Conflict in: "tasks" field                        │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │  Your version (Laptop):                  │   │
│  │  • [x] Write documentation             │   │
│  │  • [ ] Fix bug #123                   │   │
│  │  • [ ] Deploy to staging              │   │
│  │                                       │   │
│  │  Server version:                       │   │
│  │  • [x] Write documentation             │   │
│  │  • [x] Fix bug #123                   │   │
│  │  • [x] Deploy to staging              │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  How to resolve?                                │
│                                                     │
│  ○ Use Server version                              │
│  ○ Use Laptop version                              │
│  ○ Merge both (union)                             │
│  ○ Manual merge                                    │
│                                                     │
│  [Resolve] [View Diff] [Mark for Later]          │
│                                                     │
└─────────────────────────────────────────────────────────┘
```

---

#### Scenario D: Simultaneous Message in Same Session (Message Ordering)

**Situation**: User on Laptop and Server both type in same session at the same time.

**Resolution Algorithm**:
```
┌─────────────────────────────────────────────────────────┐
│  Conflict: Simultaneous Messages                 │
├─────────────────────────────────────────────────────────┤
│                                                     │
│  T=00:00:01.000 - Coordinator receives from Laptop:  │
│    {event: "message.new",                           │
│     session_id: "abc123",                            │
│     content: "Hello from laptop",                      │
│     sequence_id: 2001}                             │
│                                                     │
│  T=00:00:01.005 - Coordinator receives from Server:   │
│    {event: "message.new",                           │
│     session_id: "abc123",                            │
│     content: "Hello from server",                      │
│     sequence_id: 2002}                             │
│                                                     │
│  Coordinator Broadcasts:                             │
│  - Broadcast message seq 2001 to all devices          │
│  - Broadcast message seq 2002 to all devices          │
│                                                     │
│  Laptop Receives:                                  │
│  - Own message (seq 2001) appears                    │
│  - Server message (seq 2002) appears                 │
│  → Both displayed in order                            │
│                                                     │
│  Server Receives:                                  │
│  - Laptop message (seq 2001) appears                 │
│  - Own message (seq 2002) appears                    │
│  → Both displayed in order                            │
│                                                     │
└─────────────────────────────────────────────────────────┘
```

**Resolution**: FIFO ordering based on `sequence_id` from Coordinator. No conflict - both messages preserved.

**User Experience**: Both messages appear in correct order on both devices.

---

#### Scenario E: Integration Conflict (Both Devices Host Same Channel)

**Situation**: Both Laptop and Server try to register as Telegram bot handler.

**Resolution Algorithm**:
```
┌─────────────────────────────────────────────────────────┐
│  Conflict: Duplicate Integration Registration        │
├─────────────────────────────────────────────────────────┤
│                                                     │
│  T=00:00:01.000 - Laptop registers:              │
│    {event: "integration.register",                     │
│     channel: "telegram",                              │
│     device_id: "laptop-uuid"}                         │
│                                                     │
│  Coordinator State:                                  │
│    integrations.telegram = "laptop-uuid"              │
│                                                     │
│  T=00:00:05.000 - Server registers:               │
│    {event: "integration.register",                     │
│     channel: "telegram",                              │
│     device_id: "server-uuid"}                         │
│                                                     │
│  Coordinator Response to Server:                       │
│    {error: "integration_already_registered",             │
│     current_handler: "laptop-uuid",                   │
│     message: "Telegram bot already hosted by Laptop"}    │
│                                                     │
│  Server UI Shows:                                  │
│    ⚠️ Telegram bot already active on Laptop          │
│    Options:                                         │
│      ○ Force takeover (disconnect Laptop)            │
│      ○ Register different channel (e.g., Discord)     │
│      ○ Wait and try again later                     │
│                                                     │
└─────────────────────────────────────────────────────────┘
```

**Resolution**: First-to-register wins. Subsequent registrations are rejected with current handler info.

**User Experience**: Clear error message with options to resolve (force takeover, try different channel, wait).

---

### 4.4 Conflict Resolution Summary Table

| Conflict Type | Resolution Strategy | User Notification | UX Priority |
|----------------|---------------------|-------------------|---------------|
| Setting conflict (simultaneous edit) | Last-write-wins (sequence_id) | Toast + history link | High |
| Command execution (conflicting operations) | Lock acquisition + queue | "Waiting for lock..." indicator | High |
| State divergence (offline merge) | Three-way merge (time-based, markers, union, manual) | Conflict resolution dialog | Critical |
| Simultaneous messages | FIFO ordering (sequence_id) | None - both display in order | Low |
| Integration conflict | First-to-register wins | Error with resolution options | Medium |
| Scheduled task conflict | First-created wins + version bump | "Task already exists, would you like to edit?" | Medium |

---

## 5. Cross-Device Command Execution

**Scenario**: User on Laptop says "Run build on Server".

**Protocol**:
1.  **Laptop**:
    - LLM generates tool call: `shell.exec(cmd="make", target="server")`
    - Policy Engine checks: "Is this safe?"
    - Sends `command.routed` event to Coordinator.

2.  **Coordinator**:
    - Verifies Laptop's signature.
    - Routes event to **Server** via WebSocket.

3.  **Server**:
    - Receives `command.routed`.
    - **Local Policy Check**: "Do I allow remote execution from Laptop?"
    - Executes command.
    - Streams stdout/stderr back to Coordinator -> Laptop.

---

## 6. API Key Management (Federation)

**Goal**: User adds API Key once, works everywhere.

**Implementation**: **Encrypted Vault Sync**.
1.  **Vault**: A JSON blob containing API keys (OpenAI, Anthropic).
2.  **Encryption**: Encrypted with a symmetric "Master Key" (AES-256-GCM).
3.  **Distribution**:
    - The Encrypted Vault is stored in Cloud KV.
    - The "Master Key" is **never** sent to the cloud.
    - When pairing, Device A securely shares the Master Key with Device B (via E2EE channel established during pairing).

---

## 7. Integration Sharing

Integrations (Telegram Bot, GitHub) are **Mesh-Global** but **Device-Hosted**.

**Clarification**: This section describes the sovereignty-first mode where the integration is hosted on a user device (e.g., pryx-core long-polls Telegram). Pryx also supports a cloud-hosted webhook mode where the "device" is Pryx Edge (no user-side install) and Mesh only needs to coordinate routing and permissions, not webhook ownership.

- **Registration**: Device A registers "I host the Telegram Bot".
- **Registry**: Coordinator marks Device A as the handler for `channel:telegram`.
- **Routing**: 
    - If Device B wants to send a Telegram message, it sends an event to Coordinator.
    - Coordinator routes it to Device A.
    - Device A executes the API call.
- **Failover**: If Device A goes offline, the integration becomes unavailable (shown in UI).

---

## 8. Offline Behavior

- **Queue**: Devices queue outbound events when offline.
- **Replay**: On reconnect, they flush the queue.
- **Optimization**: "Snapshot" events (e.g., "Current Context") replace older intermediate events to save bandwidth.
