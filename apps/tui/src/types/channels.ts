/**
 * Channel type definitions for Pryx TUI
 */

export type ChannelType = "webhook" | "telegram" | "discord" | "slack" | "email" | "whatsapp";

export type ChannelStatus = "connected" | "disconnected" | "error" | "pending_setup";

export interface Channel {
  id: string;
  type: ChannelType;
  name: string;
  status: ChannelStatus;
  enabled: boolean;
  config: ChannelConfig;
  createdAt: string;
  updatedAt: string;
}

export type ChannelConfig =
  | WebhookConfig
  | TelegramConfig
  | DiscordConfig
  | SlackConfig
  | EmailConfig
  | WhatsAppConfig;

export interface WebhookConfig {
  url: string;
  secret?: string;
  method: "POST" | "GET" | "PUT" | "DELETE" | "PATCH";
  headers?: Record<string, string>;
}

export interface TelegramConfig {
  token: string;
  mode: "webhook" | "polling";
  webhookUrl?: string;
  allowedChats?: string[];
  parseMode?: "Markdown" | "MarkdownV2" | "HTML";
}

export interface DiscordConfig {
  token: string;
  applicationId?: string;
  allowedGuilds?: string[];
  allowedChannels?: string[];
  intents?: number;
}

export interface SlackConfig {
  appToken: string;
  botToken: string;
  signingSecret?: string;
  mode: "socket" | "webhook";
  allowedChannels?: string[];
}

export interface EmailConfig {
  imap: {
    host: string;
    port: number;
    secure: boolean;
    username: string;
    password: string;
  };
  smtp: {
    host: string;
    port: number;
    secure: boolean;
    username: string;
    password: string;
  };
  pollingInterval?: number;
  allowedSenders?: string[];
  folders?: string[];
}

export interface WhatsAppConfig {
  sessionData?: string;
  qrCode?: string;
  allowedContacts?: string[];
  allowedGroups?: string[];
}

export interface HealthStatus {
  healthy: boolean;
  message: string;
  lastError?: string;
  timestamp?: string;
}

export interface ChannelFormData {
  name: string;
  type: ChannelType;
  enabled: boolean;
  config: Partial<ChannelConfig>;
}

export interface ChannelTestResult {
  success: boolean;
  message: string;
  details?: Record<string, unknown>;
}

export interface ChannelActivity {
  id: string;
  channelId: string;
  type: "message_received" | "message_sent" | "error" | "status_change";
  content: string;
  timestamp: string;
  metadata?: Record<string, unknown>;
}

export const CHANNEL_TYPE_LABELS: Record<ChannelType, string> = {
  webhook: "Webhook",
  telegram: "Telegram",
  discord: "Discord",
  slack: "Slack",
  email: "Email",
  whatsapp: "WhatsApp",
};

export const CHANNEL_STATUS_LABELS: Record<ChannelStatus, string> = {
  connected: "Connected",
  disconnected: "Disconnected",
  error: "Error",
  pending_setup: "Pending Setup",
};

export const CHANNEL_STATUS_COLORS: Record<ChannelStatus, string> = {
  connected: "#4CAF50",
  disconnected: "#9E9E9E",
  error: "#F44336",
  pending_setup: "#FF9800",
};

export const DEFAULT_CHANNEL_CONFIGS: Record<ChannelType, Record<string, unknown>> = {
  webhook: {
    method: "POST",
    headers: {},
  },
  telegram: {
    mode: "polling",
    parseMode: "Markdown",
    allowedChats: [],
  },
  discord: {
    allowedGuilds: [],
    allowedChannels: [],
    intents: 0,
  },
  slack: {
    mode: "socket",
    allowedChannels: [],
  },
  email: {
    imap: {
      host: "",
      port: 993,
      secure: true,
      username: "",
      password: "",
    },
    smtp: {
      host: "",
      port: 587,
      secure: true,
      username: "",
      password: "",
    },
    pollingInterval: 30000,
    folders: ["INBOX"],
  },
  whatsapp: {
    allowedContacts: [],
    allowedGroups: [],
  },
};
