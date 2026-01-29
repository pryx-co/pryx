import { z } from 'zod';
export const ChannelType = z.enum(['telegram', 'discord', 'slack', 'email', 'whatsapp', 'webhook']);
export const RateLimitConfigSchema = z.object({
    requestsPerMinute: z.number().int().positive().optional(),
    tokensPerMinute: z.number().int().positive().optional(),
    requestsPerDay: z.number().int().positive().optional(),
});
export const TelegramConfigSchema = z.object({
    botToken: z.string().min(1),
    chatId: z.string().optional(),
    parseMode: z.enum(['HTML', 'Markdown', 'MarkdownV2']).optional(),
    disableNotification: z.boolean().optional(),
});
export const DiscordConfigSchema = z.object({
    botToken: z.string().min(1),
    applicationId: z.string().min(1),
    guildId: z.string().optional(),
    channelId: z.string().optional(),
    intents: z.array(z.string()).default([]),
});
export const SlackConfigSchema = z.object({
    botToken: z.string().min(1),
    appToken: z.string().optional(),
    signingSecret: z.string().optional(),
    channelId: z.string().optional(),
    socketMode: z.boolean().default(false),
});
export const EmailServerConfigSchema = z.object({
    host: z.string().min(1),
    port: z.number().int().positive(),
    secure: z.boolean(),
    username: z.string().min(1),
    password: z.string().min(1),
});
export const EmailConfigSchema = z.object({
    imap: EmailServerConfigSchema.optional(),
    smtp: EmailServerConfigSchema.optional(),
    checkInterval: z.number().int().positive().default(60000),
    markAsRead: z.boolean().default(true),
});
export const WhatsAppConfigSchema = z.object({
    sessionName: z.string().min(1),
    phoneNumber: z.string().optional(),
    qrTimeout: z.number().int().positive().default(60000),
    pairingCode: z.boolean().default(false),
});
export const WebhookAuthSchema = z.object({
    type: z.enum(['bearer', 'basic', 'api-key']),
    token: z.string().optional(),
    username: z.string().optional(),
    password: z.string().optional(),
});
export const WebhookRetryPolicySchema = z.object({
    maxRetries: z.number().int().min(0).default(3),
    backoffMs: z.number().int().positive().default(1000),
});
export const WebhookConfigSchema = z.object({
    url: z.string().url(),
    method: z.enum(['GET', 'POST', 'PUT', 'DELETE']).default('POST'),
    headers: z.record(z.string()).default({}),
    auth: WebhookAuthSchema.optional(),
    retryPolicy: WebhookRetryPolicySchema.default({}),
});
export const ChannelSettingsSchema = z.object({
    allowCommands: z.boolean().default(true),
    autoReply: z.boolean().default(false),
    filterPatterns: z.array(z.string()).default([]),
    allowedUsers: z.array(z.string()).default([]),
    blockedUsers: z.array(z.string()).default([]),
    rateLimit: RateLimitConfigSchema.optional(),
});
export const WebhookSettingsSchema = z.object({
    url: z.string().url(),
    secret: z.string().optional(),
    enabled: z.boolean().default(false),
});
export const ConnectionStatusSchema = z.object({
    connected: z.boolean(),
    lastConnected: z.string().datetime().optional(),
    lastError: z.string().optional(),
    errorCount: z.number().int().min(0).default(0),
    messageCount: z.number().int().min(0).default(0),
});
export const ChannelConfigSchema = z.object({
    id: z.string().min(1).max(64).regex(/^[a-z0-9_-]+$/),
    name: z.string().min(1).max(128),
    type: ChannelType,
    enabled: z.boolean().default(true),
    config: z.union([
        TelegramConfigSchema,
        DiscordConfigSchema,
        SlackConfigSchema,
        EmailConfigSchema,
        WhatsAppConfigSchema,
        WebhookConfigSchema,
    ]),
    settings: ChannelSettingsSchema.default({}),
    webhook: WebhookSettingsSchema.optional(),
    status: ConnectionStatusSchema.optional(),
});
export const ChannelsConfigSchema = z.object({
    version: z.number().int().default(1),
    channels: z.array(ChannelConfigSchema),
});
export const ValidationResultSchema = z.object({
    valid: z.boolean(),
    errors: z.array(z.string()),
});
export const ConnectionTestResultSchema = z.object({
    success: z.boolean(),
    latency: z.number().optional(),
    error: z.string().optional(),
});
export class ChannelError extends Error {
    constructor(message) {
        super(message);
        this.name = 'ChannelError';
    }
}
export class ChannelNotFoundError extends ChannelError {
    constructor(id) {
        super(`Channel not found: ${id}`);
        this.name = 'ChannelNotFoundError';
    }
}
export class ChannelValidationError extends ChannelError {
    errors;
    constructor(errors) {
        super(`Validation failed: ${errors.join(', ')}`);
        this.errors = errors;
        this.name = 'ChannelValidationError';
    }
}
export class ChannelAlreadyExistsError extends ChannelError {
    constructor(id) {
        super(`Channel already exists: ${id}`);
        this.name = 'ChannelAlreadyExistsError';
    }
}
export const CURRENT_VERSION = 1;
//# sourceMappingURL=types.js.map