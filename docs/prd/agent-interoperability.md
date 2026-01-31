# Epic: Agent Interoperability System

**Status**: Planning  
**Priority**: P2  
**Parent PRD**: v1 (docs/prd/prd.md)  
**Related**: Skills System (pryx-l3q), MCP Integration (pryx-gf8)  

---

## Executive Summary

Design and implement a comprehensive agent interoperability system that enables Pryx to discover, authenticate, and communicate with any external AI agent system (not limited to Clawdbot/Moltbot/OpenClaw but extensible to future agents).

## Vision Statement

**"Universal Agent Federation"** - Pryx becomes a sovereign-first agent platform that can:

- Discover and register other AI agents dynamically
- Exchange messages and coordinate workflows across agent boundaries
- Share tools, skills, and capabilities securely
- Maintain sovereignty (local execution, local data) while enabling collaboration

---

## Scope Breakdown

### Phase 1: Foundation (P0)

#### Task: pryx-interop-003 - Authentication & Authorization Layer
**Description**: Implement multi-method auth system supporting OAuth 2.0, API keys, shared secrets, and MTLS for mutual TLS.
**Acceptance Criteria**:
- OAuth 2.0 flow for external agent authentication
- API key management for direct agent communication
- Shared secret establishment for trusted agent federation
- Token validation and refresh mechanisms

**Previous**: pryx-interop-001 - Agent Registry Service ✅ COMPLETED

**Files Created**:
- apps/runtime/internal/registry/registry.go
- apps/runtime/internal/registry/handler.go
- apps/runtime/internal/registry/agent_registry_test.go
- apps/runtime/internal/registry/README.md

**Next**: pryx-interop-004 - Message Exchange Protocol

**Status**: 🔄 In Progress

---

### Phase 1: Foundation (P0)

#### Task: pryx-interop-001 - Agent Registry Service ✅ COMPLETED

#### Task: pryx-interop-002 - Agent Discovery Protocol
**Description**: Define protocol for agents to advertise their capabilities and discover other agents.
**Acceptance Criteria**:
- Agent registration endpoint
- Capability advertisement format
- Discovery query mechanisms (by type, capability, name)
- Version compatibility checking

#### Task: pryx-interop-003 - Authentication & Authorization Layer
**Description**: Implement multi-method auth system supporting OAuth 2.0, API keys, shared secrets, and MTLS for mutual TLS.
**Acceptance Criteria**:
- OAuth 2.0 flow for external agent authentication
- API key management for direct agent communication
- Shared secret establishment for trusted agent federation
- Token validation and refresh mechanisms

#### Task: pryx-interop-004 - Message Exchange Protocol
**Description**: Define standardized message format and transport for agent-to-agent communication.
**Acceptance Criteria**:
- JSON schema for message payloads
- Support for HTTP and WebSocket transports
- Request/response correlation IDs
- Streaming message support
- Error handling and retry strategies

---

## Key Requirements

### FR_INTEROP_001: Agent Registry API
- Create, read, update, delete agent registrations
- Query agents by ID, name, capabilities
- Health check endpoint per agent

### FR_INTEROP_002: Agent Discovery
- Agents can register their capabilities (tools, skills, models)
- Discovery by capability type and name
- Version constraint checking

### FR_INTEROP_003: Agent-to-Agent Messaging
- Send messages to external agents
- Receive messages from external agents
- Async message queue with correlation IDs
- Streaming support for real-time collaboration

### FR_INTEROP_004: Authentication & Authorization
- Support OAuth 2.0 for external agents
- API key authentication
- Shared secret establishment for trusted federation
- Token validation and automatic refresh

### FR_INTEROP_005: Capability Exchange Protocol
- Tool/skill advertisement format
- Permission request/response flow
- Capability compatibility checking before interaction

### FR_INTEROP_006: Policy-Based Authorization
- Per-connection trust levels
- Action allow/deny/ask decisions
- Resource constraints
- Time-based limits

### FR_INTEROP_007: Session Handoff
- Context transfer between agent instances
- Session state serialization
- Handoff initiation and acceptance flows

### FR_INTEROP_008: Tool Federation
- Tool discovery across agent boundaries
- Tool invocation with proper authorization
- Result streaming back to caller

---

## References

### Internal Patterns
- **OpenClaw multi-agent.md**: Bindings system, agent-to-agent tools (`sessions_list`, `sessions_history`, `sessions_send`)
- **OpenCode agent.ts**: Delegation system (`@agent` pattern), native vs external agents
- **Pryx event bus**: `apps/runtime/internal/bus/bus.go` - extensible pub/sub system

### External Standards
- **MCP (Model Context Protocol)**: Tool sharing standard
- **OAuth 2.0**: Delegated authorization framework
- **WebSocket**: Real-time communication transport
- **JSON Schema**: Standard message format

---

## Anti-Patterns

### ❌ Hardcoded Agent Integrations
**Don't**: Build specific integrations for clawdbot/moltbot/openclaw
**Instead**: Build a generic discovery and registration system

### ❌ Global Trust Model
**Don't**: Trust all external agents equally
**Instead**: Per-connection trust levels that evolve over time

### ❌ Monolithic Auth System
**Don't**: Use only one authentication method
**Instead**: Support OAuth 2.0, API keys, and shared secrets

### ❌ Direct Tool Access Without Policy
**Don't**: Allow external agents direct access to tools
**Instead**: All tool access must go through policy engine with approval

### ❌ Blocking Communication
**Don't**: Force synchronous request/response patterns
**Instead**: Async message queue with correlation tracking

---

## Success Criteria

- [ ] Pryx can register any agent with HTTP/WebSocket endpoint
- [ ] Pryx agents can discover each other's capabilities
- [ ] Pryx can send messages to external agents with routing
- [ ] External agents can send messages to Pryx agents
- [ ] All communication uses standardized JSON protocol with authentication
- [ ] Policy engine enforces permissions per connection
- [ ] Sessions can be handed off between different instances
- [ ] Health monitoring detects and reports issues

---

## Dependencies

- **Phase 1**: Event bus extension, agent registry service, HTTP client
- **Phase 2**: Messaging queue, policy engine integration, MCP extension
- **Phase 3**: Health monitoring system, trust graph, marketplace client

---

## Notes

### Current State Assessment

**What Pryx Has Now**:
- ✅ Event bus for internal communication
- ✅ Simple sub-agent spawning (max 10 concurrent)
- ✅ Session store with basic metadata
- ✅ MCP tool integration planned
- ❌ No multi-agent routing/binding system
- ❌ No agent discovery capability
- ❌ No agent-to-agent messaging protocol
- ❌ No external agent authentication

**What's Missing for Interoperability**:
- No agent registry service
- No agent discovery protocol
- No external authentication layer
- No standardized messaging protocol
- No policy engine for cross-agent boundaries
- No session handoff capability
- No tool/skill federation

**Gap Analysis**: Pryx is currently a single-agent system with basic sub-agent spawning. To achieve universal agent federation, we need to build a complete interoperability architecture from the ground up.

### Design Philosophy

**Sovereign-First Foundation**:
- Local agents run in their own processes with full sovereignty
- External agents are untrusted by default
- All cross-agent interactions go through policy engine
- Data never leaves the device without explicit consent
- Discovery-based architecture (no hardcoded integrations)

**Standardized Protocols**:
- Versioned JSON schemas for all messages
- WebSocket for real-time communication
- HTTP for request/response
- Correlation IDs for async operations
- Retry and error handling built-in

**Extensibility First**:
- Plugin architecture for new agent types
- Capability advertisement system
- Version compatibility checking
- Backward compatibility guarantees

---

## Next Steps

1. Review and approve this epic with stakeholders
2. Phase 1 tasks can be implemented in parallel with current v1 features
3. Phase 2 requires Foundation components from Phase 1
4. Phase 3 builds on Phase 2 core

**Timeline Estimate**: 
- Phase 1: 4-6 weeks
- Phase 2: 6-8 weeks  
- Phase 3: 8-12 weeks
