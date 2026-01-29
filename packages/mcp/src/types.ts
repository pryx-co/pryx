import { z } from 'zod';

export const TransportType = z.enum(['stdio', 'sse', 'websocket']);
export const ServerSource = z.enum(['manual', 'curated', 'marketplace']);

export const ToolDefinitionSchema = z.object({
  name: z.string().min(1),
  description: z.string(),
  inputSchema: z.record(z.unknown()),
});

export const ResourceDefinitionSchema = z.object({
  uri: z.string().url(),
  name: z.string().min(1),
  mimeType: z.string().optional(),
});

export const ArgumentDefinitionSchema = z.object({
  name: z.string().min(1),
  description: z.string(),
  required: z.boolean().default(false),
});

export const PromptDefinitionSchema = z.object({
  name: z.string().min(1),
  description: z.string(),
  arguments: z.array(ArgumentDefinitionSchema).optional(),
});

export const CapabilitiesSchema = z.object({
  tools: z.array(ToolDefinitionSchema).default([]),
  resources: z.array(ResourceDefinitionSchema).default([]),
  prompts: z.array(PromptDefinitionSchema).default([]),
});

export const StdioTransportSchema = z.object({
  type: z.literal('stdio'),
  command: z.string().min(1),
  args: z.array(z.string()).default([]),
  env: z.record(z.string()).default({}),
  cwd: z.string().optional(),
});

export const SSETransportSchema = z.object({
  type: z.literal('sse'),
  url: z.string().url(),
  headers: z.record(z.string()).default({}),
});

export const WebSocketTransportSchema = z.object({
  type: z.literal('websocket'),
  url: z.string().url().regex(/^wss?:\/\//),
  headers: z.record(z.string()).default({}),
});

export const TransportConfigSchema = z.union([
  StdioTransportSchema,
  SSETransportSchema,
  WebSocketTransportSchema,
]);

export const ServerSettingsSchema = z.object({
  autoConnect: z.boolean().default(true),
  timeout: z.number().int().positive().default(30000),
  reconnect: z.boolean().default(true),
  maxReconnectAttempts: z.number().int().min(0).default(3),
  fallbackServers: z.array(z.string()).default([]),
});

export const ConnectionStatusSchema = z.object({
  connected: z.boolean(),
  lastConnected: z.string().datetime().optional(),
  lastError: z.string().optional(),
  reconnectAttempts: z.number().int().min(0).default(0),
});

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

export const MCPServersConfigSchema = z.object({
  version: z.number().int().default(1),
  servers: z.array(MCPServerConfigSchema),
});

export const ValidationResultSchema = z.object({
  valid: z.boolean(),
  errors: z.array(z.string()),
});

export const ConnectionTestResultSchema = z.object({
  success: z.boolean(),
  latency: z.number().optional(),
  error: z.string().optional(),
  capabilities: CapabilitiesSchema.optional(),
});

export type TransportType = z.infer<typeof TransportType>;
export type ServerSource = z.infer<typeof ServerSource>;
export type ToolDefinition = z.infer<typeof ToolDefinitionSchema>;
export type ResourceDefinition = z.infer<typeof ResourceDefinitionSchema>;
export type ArgumentDefinition = z.infer<typeof ArgumentDefinitionSchema>;
export type PromptDefinition = z.infer<typeof PromptDefinitionSchema>;
export type Capabilities = z.infer<typeof CapabilitiesSchema>;
export type StdioTransport = z.infer<typeof StdioTransportSchema>;
export type SSETransport = z.infer<typeof SSETransportSchema>;
export type WebSocketTransport = z.infer<typeof WebSocketTransportSchema>;
export type TransportConfig = z.infer<typeof TransportConfigSchema>;
export type ServerSettings = z.infer<typeof ServerSettingsSchema>;
export type ConnectionStatus = z.infer<typeof ConnectionStatusSchema>;
export type MCPServerConfig = z.infer<typeof MCPServerConfigSchema>;
export type MCPServersConfig = z.infer<typeof MCPServersConfigSchema>;
export type ValidationResult = z.infer<typeof ValidationResultSchema>;
export type ConnectionTestResult = z.infer<typeof ConnectionTestResultSchema>;

export class MCPError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'MCPError';
  }
}

export class MCPServerNotFoundError extends MCPError {
  constructor(id: string) {
    super(`MCP server not found: ${id}`);
    this.name = 'MCPServerNotFoundError';
  }
}

export class MCPValidationError extends MCPError {
  constructor(public errors: string[]) {
    super(`Validation failed: ${errors.join(', ')}`);
    this.name = 'MCPValidationError';
  }
}

export class MCPServerAlreadyExistsError extends MCPError {
  constructor(id: string) {
    super(`MCP server already exists: ${id}`);
    this.name = 'MCPServerAlreadyExistsError';
  }
}

export const CURRENT_VERSION = 1;
