/**
 * MCP Types and Schemas
 *
 * Defines Zod schemas and TypeScript types for Model Context Protocol (MCP)
 * server configurations, including transports, capabilities, and settings.
 */

import { z } from 'zod';

/**
 * Enum of supported transport types for MCP servers
 */
export const TransportType = z.enum(['stdio', 'sse', 'websocket']);

/**
 * Enum of server source types
 */
export const ServerSource = z.enum(['manual', 'curated', 'marketplace']);

/**
 * Schema for tool definitions
 */
export const ToolDefinitionSchema = z.object({
  name: z.string().min(1),
  description: z.string(),
  inputSchema: z.record(z.unknown()),
});

/**
 * Schema for resource definitions
 */
export const ResourceDefinitionSchema = z.object({
  uri: z.string().url(),
  name: z.string().min(1),
  mimeType: z.string().optional(),
});

/**
 * Schema for argument definitions
 */
export const ArgumentDefinitionSchema = z.object({
  name: z.string().min(1),
  description: z.string(),
  required: z.boolean().default(false),
});

/**
 * Schema for prompt definitions
 */
export const PromptDefinitionSchema = z.object({
  name: z.string().min(1),
  description: z.string(),
  arguments: z.array(ArgumentDefinitionSchema).optional(),
});

/**
 * Schema for server capabilities
 */
export const CapabilitiesSchema = z.object({
  tools: z.array(ToolDefinitionSchema).default([]),
  resources: z.array(ResourceDefinitionSchema).default([]),
  prompts: z.array(PromptDefinitionSchema).default([]),
});

/**
 * Schema for stdio transport configuration
 */
export const StdioTransportSchema = z.object({
  type: z.literal('stdio'),
  command: z.string().min(1),
  args: z.array(z.string()).default([]),
  env: z.record(z.string()).default({}),
  cwd: z.string().optional(),
});

/**
 * Schema for SSE transport configuration
 */
export const SSETransportSchema = z.object({
  type: z.literal('sse'),
  url: z.string().url(),
  headers: z.record(z.string()).default({}),
});

/**
 * Schema for WebSocket transport configuration
 */
export const WebSocketTransportSchema = z.object({
  type: z.literal('websocket'),
  url: z.string().url().regex(/^wss?:\/\//),
  headers: z.record(z.string()).default({}),
});

/**
 * Schema for transport configuration (union of all types)
 */
export const TransportConfigSchema = z.union([
  StdioTransportSchema,
  SSETransportSchema,
  WebSocketTransportSchema,
]);

/**
 * Schema for server settings
 */
export const ServerSettingsSchema = z.object({
  autoConnect: z.boolean().default(true),
  timeout: z.number().int().positive().default(30000),
  reconnect: z.boolean().default(true),
  maxReconnectAttempts: z.number().int().min(0).default(3),
  fallbackServers: z.array(z.string()).default([]),
});

/**
 * Schema for connection status
 */
export const ConnectionStatusSchema = z.object({
  connected: z.boolean(),
  lastConnected: z.string().datetime().optional(),
  lastError: z.string().optional(),
  reconnectAttempts: z.number().int().min(0).default(0),
});

/**
 * Schema for MCP server configuration
 */
export const MCPServerConfigSchema = z.object({
  id: z.string().min(1).max(64).regex(/^[a-z0-9_-]+$/),
  name: z.string().min(1).max(128),
  enabled: z.boolean().default(true),
  transport: TransportConfigSchema,
  capabilities: CapabilitiesSchema.optional(),
  source: ServerSource.default('manual'),
  settings: ServerSettingsSchema.default({}),
  status: ConnectionStatusSchema.optional(),
});

/**
 * Schema for complete MCP servers configuration
 */
export const MCPServersConfigSchema = z.object({
  version: z.number().int().default(1),
  servers: z.array(MCPServerConfigSchema),
});

/**
 * Schema for validation results
 */
export const ValidationResultSchema = z.object({
  valid: z.boolean(),
  errors: z.array(z.string()),
});

/**
 * Schema for connection test results
 */
export const ConnectionTestResultSchema = z.object({
  success: z.boolean(),
  latency: z.number().optional(),
  error: z.string().optional(),
  capabilities: CapabilitiesSchema.optional(),
});

/**
 * Inferred TypeScript type for TransportType
 */
export type TransportType = z.infer<typeof TransportType>;

/**
 * Inferred TypeScript type for ServerSource
 */
export type ServerSource = z.infer<typeof ServerSource>;

/**
 * Inferred TypeScript type for ToolDefinition
 */
export type ToolDefinition = z.infer<typeof ToolDefinitionSchema>;

/**
 * Inferred TypeScript type for ResourceDefinition
 */
export type ResourceDefinition = z.infer<typeof ResourceDefinitionSchema>;

/**
 * Inferred TypeScript type for ArgumentDefinition
 */
export type ArgumentDefinition = z.infer<typeof ArgumentDefinitionSchema>;

/**
 * Inferred TypeScript type for PromptDefinition
 */
export type PromptDefinition = z.infer<typeof PromptDefinitionSchema>;

/**
 * Inferred TypeScript type for Capabilities
 */
export type Capabilities = z.infer<typeof CapabilitiesSchema>;

/**
 * Inferred TypeScript type for StdioTransport
 */
export type StdioTransport = z.infer<typeof StdioTransportSchema>;

/**
 * Inferred TypeScript type for SSETransport
 */
export type SSETransport = z.infer<typeof SSETransportSchema>;

/**
 * Inferred TypeScript type for WebSocketTransport
 */
export type WebSocketTransport = z.infer<typeof WebSocketTransportSchema>;

/**
 * Inferred TypeScript type for TransportConfig
 */
export type TransportConfig = z.infer<typeof TransportConfigSchema>;

/**
 * Inferred TypeScript type for ServerSettings
 */
export type ServerSettings = z.infer<typeof ServerSettingsSchema>;

/**
 * Inferred TypeScript type for ConnectionStatus
 */
export type ConnectionStatus = z.infer<typeof ConnectionStatusSchema>;

/**
 * Inferred TypeScript type for MCPServerConfig
 */
export type MCPServerConfig = z.infer<typeof MCPServerConfigSchema>;

/**
 * Inferred TypeScript type for MCPServersConfig
 */
export type MCPServersConfig = z.infer<typeof MCPServersConfigSchema>;

/**
 * Inferred TypeScript type for ValidationResult
 */
export type ValidationResult = z.infer<typeof ValidationResultSchema>;

/**
 * Inferred TypeScript type for ConnectionTestResult
 */
export type ConnectionTestResult = z.infer<typeof ConnectionTestResultSchema>;

/**
 * Base error class for MCP-related errors
 */
export class MCPError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'MCPError';
  }
}

/**
 * Error thrown when an MCP server is not found
 */
export class MCPServerNotFoundError extends MCPError {
  constructor(id: string) {
    super(`MCP server not found: ${id}`);
    this.name = 'MCPServerNotFoundError';
  }
}

/**
 * Error thrown when MCP configuration validation fails
 */
export class MCPValidationError extends MCPError {
  constructor(public errors: string[]) {
    super(`Validation failed: ${errors.join(', ')}`);
    this.name = 'MCPValidationError';
  }
}

/**
 * Error thrown when attempting to add an MCP server that already exists
 */
export class MCPServerAlreadyExistsError extends MCPError {
  constructor(id: string) {
    super(`MCP server already exists: ${id}`);
    this.name = 'MCPServerAlreadyExistsError';
  }
}

/**
 * Current configuration version number
 */
export const CURRENT_VERSION = 1;
