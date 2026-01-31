# Agent Registry Service API Documentation

## Overview

The Agent Registry Service provides a central registry for managing AI agent identities, capabilities, and endpoints within the Pryx interoperability system.

**Purpose**: Enable Pryx to discover, authenticate, and communicate with any external AI agent system (OpenClaw, OpenCode, Clawdbot, Moltbot, and future agents).

**API Base Path**: `/api/v1/agents`

---

## Data Models

### Agent

```json
{
  "id": "string - Unique identifier for the agent",
  "name": "string - Human-readable agent name",
  "description": "string - Detailed description of the agent's purpose",
  "version": "string - Version string for compatibility checking",
  "capabilities": [
    {
      "type": "string - Type of capability (tool, skill, model)",
      "name": "string - Name of the capability",
      "version": "string - Version of the capability",
      "description": "string - Description of what this capability does",
      "permissions": ["string"] - List of permissions required for this capability"
    }
  ],
  "endpoint": {
    "type": "string - Either 'http' or 'websocket'",
    "host": "string - Hostname or IP address",
    "port": "string - Port number (optional)",
    "path": "string - URL path (optional)",
    "url": "string - Full constructed URL"
  },
  "trust_level": "string - One of: 'untrusted', 'sandboxed', 'trusted'",
  "health": "string - One of: 'online', 'offline', 'degraded', 'unknown'",
  "registered_at": "string - ISO 8601 timestamp when agent was registered",
  "last_seen": "string - ISO 8601 timestamp when agent was last seen/healthy"
}
```

### Endpoint Types

- **http**: HTTP endpoint for REST APIs
- **websocket**: WebSocket endpoint for real-time communication

### Trust Levels

- **untrusted**: Unknown agent, no prior trust, require explicit approval for all actions
- **sandboxed**: Agent running in isolated environment, limited permissions
- **trusted**: Known agent with established trust relationship

### Health Status

- **online**: Agent is responding to health checks
- **offline**: Agent is not responding
- **degraded**: Agent is responding but degraded performance
- **unknown**: Health status has not been checked yet

---

## API Endpoints

### POST /api/v1/agents

**Description**: Register a new agent with the registry

**Request Body**:
```json
{
  "id": "string - Required: Unique agent ID",
  "name": "string - Required: Human-readable name",
  "description": "string - Optional: Agent description",
  "version": "string - Optional: Version string",
  "capabilities": "array - Required: List of agent capabilities",
  "endpoint": "object - Required: How to contact the agent",
  "trust_level": "string - Optional: One of 'untrusted', 'sandboxed', 'trusted' (default: 'untrusted')"
}
```

**Response**: `201 Created`  
**Response Body**: The created Agent object

**Errors**:
- `400 Bad Request`: Invalid request body or missing required fields
- `409 Conflict`: Agent ID already exists
- `500 Internal Server Error`: Server error during registration

**Example**:
```bash
curl -X POST http://localhost:3000/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my-agent-001",
    "name": "Code Assistant",
    "description": "Helps with coding tasks",
    "version": "1.0.0",
    "capabilities": [
      {
        "type": "tool",
        "name": "execute",
        "description": "Execute shell commands",
        "permissions": ["shell", "read"]
      }
    ],
    "endpoint": {
      "type": "http",
      "host": "localhost",
      "port": "8080"
    },
    "trust_level": "trusted"
  }'
```

---

### GET /api/v1/agents/{id}

**Description**: Retrieve agent information by ID

**Response**: `200 OK`  
**Response Body**: The Agent object

**Errors**:
- `404 Not Found`: Agent with specified ID does not exist

**Example**:
```bash
curl http://localhost:3000/api/v1/agents/my-agent-001
```

---

### GET /api/v1/agents

**Description**: List all registered agents

**Query Parameters**:
- `capability_type` (string, optional): Filter by capability type
- `capability_name` (string, optional): Filter by capability name
- `trust_level` (string, optional): Filter by trust level
- `min_version` (string, optional): Minimum version requirement
- `max_version` (string, optional): Maximum version constraint

**Response**: `200 OK`  
**Response Body**:
```json
{
  "agents": [Agent objects]
}
```

**Example**:
```bash
# List all agents
curl http://localhost:3000/api/v1/agents

# Filter by capability type
curl http://localhost:3000/api/v1/agents?capability_type=tool

# Filter by trust level
curl http://localhost:3000/api/v1/agents?trust_level=trusted

# Find agents with specific capability
curl http://localhost:3000/api/v1/agents?capability_name=execute
```

---

### DELETE /api/v1/agents/{id}

**Description**: Unregister (remove) an agent from the registry

**Response**: `204 No Content`  

**Errors**:
- `404 Not Found`: Agent with specified ID does not exist
- `500 Internal Server Error`: Server error during removal

**Example**:
```bash
curl -X DELETE http://localhost:3000/api/v1/agents/my-agent-001
```

---

### GET /api/v1/agents/discover

**Description**: Discover agents matching specified criteria

**Query Parameters**: Same as GET /api/v1/agents

**Response**: `200 OK`  
**Response Body**:
```json
{
  "agents": [Matching Agent objects]
}
```

**Example**:
```bash
# Find all agents with tool capability
curl http://localhost:3000/api/v1/agents/discover?capability_type=tool

# Find trusted agents
curl http://localhost:3000/api/v1/agents/discover?trust_level=trusted

# Find agents with version >= 1.0.0
curl "http://localhost:3000/api/v1/agents/discover?min_version=1.0.0"
```

---

### PUT /api/v1/agents/{id}/health

**Description**: Update health status for an agent

**Request Body**:
```json
{
  "status": "string - Required: New health status (online, offline, degraded, unknown)"
}
```

**Response**: `204 No Content`  

**Errors**:
- `404 Not Found`: Agent with specified ID does not exist
- `400 Bad Request`: Invalid status value
- `500 Internal Server Error`: Server error during update

**Example**:
```bash
# Mark agent as offline
curl -X PUT http://localhost:3000/api/v1/agents/my-agent-001 \
  -H "Content-Type: application/json" \
  -d '{"status": "offline"}'

# Mark agent as online
curl -X PUT http://localhost:3000/api/v1/agents/my-agent-001 \
  -H "Content-Type: application/json" \
  -d '{"status": "online"}'
```

---

## Events

The Agent Registry Service publishes events to the event bus for important agent lifecycle events.

### agent.registered

Emitted when a new agent is successfully registered.

**Event Type**: `agent.registered`  
**Payload**:
```json
{
  "agent_id": "string - The registered agent ID",
  "name": "string - Agent name",
  "version": "string - Agent version",
  "endpoint": "object - Agent's endpoint information"
}
```

### agent.unregistered

Emitted when an agent is removed from the registry.

**Event Type**: `agent.unregistered`  
**Payload**:
```json
{
  "agent_id": "string - The unregistered agent ID"
}
```

### agent.health_updated

Emitted when an agent's health status changes.

**Event Type**: `agent.health_updated`  
**Payload**:
```json
{
  "agent_id": "string - The agent ID",
  "status": "string - New health status (online, offline, degraded, unknown)"
  "timestamp": "string - When the status was last seen/checked"
}
```

---

## Integration with Event Bus

The Agent Registry Service integrates with the Pryx event bus (`internal/bus`):

1. **Register**: Service subscribes to event bus for lifecycle events
2. **Publish**: Service publishes agent registration/unregistration events
3. **Health Checks**: Background goroutine monitors agent health and publishes updates

### Integration Points

- **Agent Discovery**: Discovery results are used by messaging layer to find agents
- **Message Exchange**: Agent endpoints are included in agent routing
- **Policy Engine**: Trust levels from registry inform policy decisions

---

## Security Considerations

### Registration Validation

- **Agent ID**: Must be unique, validated against existing registrations
- **Endpoint**: Must be valid (http or websocket with proper host/port)
- **Trust Level**: Must be valid value (untrusted, sandboxed, trusted)
- **Capabilities**: Each capability must specify required permissions

### Discovery Permissions

- No authentication required for public discovery
- Filtered results never expose agent metadata beyond what's published
- Health status may be stale (background checks)

### Authentication & Authorization (Future Enhancement)

Future versions of the agent registry will support:
- OAuth 2.0 authentication for agent-to-agent communication
- API key management for direct agent access
- Shared secret establishment for trusted agent federation
- Dynamic trust level updates based on agent behavior

---

## Usage Examples

### Example 1: Register a Code Assistant Agent

```bash
# Register a new agent
curl -X POST http://localhost:3000/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "code-assistant-001",
    "name": "Code Assistant",
    "description": "Specializes in code review and suggestions",
    "version": "1.0.0",
    "capabilities": [
      {
        "type": "tool",
        "name": "code_review",
        "description": "Review code for bugs and improvements",
        "permissions": ["read"]
      },
      {
        "type": "skill",
        "name": "git",
        "description": "Manage git repositories",
        "permissions": ["read", "write"]
      }
    ],
    "endpoint": {
      "type": "http",
      "host": "code-assistant.internal",
      "port": "8080"
    },
    "trust_level": "trusted"
  }'
```

### Example 2: Discover All Agents

```bash
# List all registered agents
curl http://localhost:3000/api/v1/agents
```

Response:
```json
{
  "agents": [
    {
      "id": "code-assistant-001",
      "name": "Code Assistant",
      "health": "online",
      ...
    },
    {
      "id": "data-agent-001",
      "name": "Data Processor",
      "health": "online",
      ...
    }
  ]
}
```

### Example 3: Discover Agents with Tool Capability

```bash
# Find agents that have tool capabilities
curl "http://localhost:3000/api/v1/agents/discover?capability_type=tool"
```

---

## Testing

Run unit tests for the registry service:

```bash
# Run all tests
go test ./apps/runtime/internal/registry/... -v

# Run specific test
go test ./apps/runtime/internal/registry/... -run TestRegister
```

---

## Dependencies

- `pryx-core/internal/bus`: Event bus for publishing registry events
- `pryx-core/internal/config`: Configuration for registry settings (future)
- `pryx-core/internal/store`: Persistent storage for agent registry (future)

---

## Next Steps

This is **Phase 1, Task 1 (pryx-interop-001)** of the Agent Interoperability epic.

**Completed**:
- ✅ Agent Registry data structures designed
- ✅ Agent Registry service implemented
- ✅ HTTP API endpoints created
- ✅ Unit tests written
- ✅ API documentation created

**Next**: pryx-interop-002 (Agent Discovery Protocol specification)
