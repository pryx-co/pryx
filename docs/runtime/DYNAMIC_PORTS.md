# Dynamic Port Allocation

Following the patterns used by **OpenCode** and **Moltbot**, Pryx uses **dynamic port allocation** to avoid conflicts with other services.

## How It Works

### 1. Runtime Dynamic Port Assignment

The Pryx runtime uses port `:0` by default, which tells the OS to assign an available random port:

```go
// config.go - Default configuration
cfg := &Config{
    ListenAddr: ":0",  // Dynamic port allocation
    // ...
}
```

When the runtime starts:
1. If `listen_addr` is `:0` or `:3000`, it finds an available port
2. The actual port is written to `~/.pryx/runtime.port`
3. Clients read this file to discover the runtime port
4. Port file is cleaned up on shutdown

### 2. Port Discovery

**Services automatically discover the runtime port:**

```typescript
// WebSocket service
const getRuntimeURL = () => {
    if (process.env.PRYX_WS_URL) return process.env.PRYX_WS_URL;
    try {
        const port = readFileSync(join(homedir(), ".pryx", "runtime.port"), "utf-8").trim();
        return `ws://localhost:${port}/ws`;
    } catch {
        return "ws://localhost:3000/ws";  // Fallback
    }
};
```

```typescript
// API service
function getApiUrl(): string {
    if (process.env.PRYX_API_URL) return process.env.PRYX_API_URL;
    try {
        const port = readFileSync(join(homedir(), ".pryx", "runtime.port"), "utf-8").trim();
        return `http://localhost:${port}`;
    } catch {
        return "http://localhost:3000";
    }
}
```

### 3. Environment Variables (Override)

You can override the dynamic port allocation:

```bash
# Force a specific port
export PRYX_LISTEN_ADDR=:8080

# Or override discovery for clients
export PRYX_API_URL=http://localhost:8080
export PRYX_WS_URL=ws://localhost:8080/ws
```

### 4. Dev Scripts

The `make dev-tui` script handles dynamic ports automatically:

1. Starts runtime (which picks a random port)
2. Waits for `~/.pryx/runtime.port` to be created
3. Reads the actual port
4. Sets `PRYX_API_URL` and `PRYX_WS_URL` for TUI
5. Starts TUI with correct configuration

## Why Dynamic Ports?

### Problems with Fixed Ports (3000, 4321, 8080):
- **Commonly used** by Next.js, React dev server, Astro, etc.
- **Conflicts** when running multiple projects
- **Manual port changes** required for each environment
- **Hardcoded** in configs and scripts

### Benefits of Dynamic Ports:
- **Always available** - OS assigns free port
- **No conflicts** with other services
- **Automatic discovery** via port file
- **Same approach** as OpenCode and Moltbot
- **Works in CI/CD** without port management

## Port File Location

```
~/.pryx/runtime.port
```

Contains just the port number:
```
58873
```

## Troubleshooting

### Port file not created
```bash
# Check if runtime is running
ps aux | grep pryx-core

# Check logs
ls -la ~/.pryx/
cat ~/.pryx/runtime.port
```

### Manual port configuration
```bash
# Set explicit port
export PRYX_LISTEN_ADDR=:8080
./bin/pryx-core

# Connect TUI to explicit port
export PRYX_API_URL=http://localhost:8080
export PRYX_WS_URL=ws://localhost:8080/ws
./pryx-tui
```

### Find what port runtime is using
```bash
# Method 1: Check port file
cat ~/.pryx/runtime.port

# Method 2: Check process
lsof -i -P | grep pryx-core

# Method 3: Check runtime logs
./bin/pryx-core 2>&1 | grep "Starting server"
```

## Reference

### OpenCode Pattern
```typescript
// From opencode sdks/vscode/src/extension.ts
const port = Math.floor(Math.random() * (65535 - 16384 + 1)) + 16384
```

### Moltbot Pattern
```typescript
// From moltbot src/test-utils/ports.ts
async function getOsFreePort(): Promise<number> {
  return await new Promise((resolve, reject) => {
    const server = createServer();
    server.listen(0, "127.0.0.1", () => {
      const addr = server.address();
      const port = (addr as AddressInfo).port;
      server.close((err) => (err ? reject(err) : resolve(port)));
    });
  });
}
```

Pryx uses a simpler approach: Go's `http.ListenAndServe(":0", ...)` lets the OS assign the port, then we write it to a file for discovery.
