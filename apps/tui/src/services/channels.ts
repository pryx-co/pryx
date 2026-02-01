/**
 * Channel API Service
 *
 * Provides HTTP API calls to the Pryx runtime for channel management.
 */

import type {
  Channel,
  ChannelConfig,
  HealthStatus,
  ChannelTestResult,
  ChannelActivity,
} from "../types/channels";

const API_BASE = "/api";

/**
 * Get all configured channels
 */
export const getChannels = async (): Promise<Channel[]> => {
  const response = await fetch(`${API_BASE}/channels`);

  if (!response.ok) {
    throw new Error(`Failed to fetch channels: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Get a specific channel by ID
 */
export const getChannel = async (id: string): Promise<Channel> => {
  const response = await fetch(`${API_BASE}/channels/${id}`);

  if (!response.ok) {
    throw new Error(`Failed to fetch channel: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Create a new channel
 */
export const createChannel = async (
  config: ChannelConfig & { name: string; type: string }
): Promise<Channel> => {
  const response = await fetch(`${API_BASE}/channels`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(config),
  });

  if (!response.ok) {
    throw new Error(`Failed to create channel: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Update an existing channel
 */
export const updateChannel = async (
  id: string,
  config: Partial<ChannelConfig> & { name?: string; enabled?: boolean }
): Promise<Channel> => {
  const response = await fetch(`${API_BASE}/channels/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(config),
  });

  if (!response.ok) {
    throw new Error(`Failed to update channel: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Delete a channel
 */
export const deleteChannel = async (id: string): Promise<void> => {
  const response = await fetch(`${API_BASE}/channels/${id}`, {
    method: "DELETE",
  });

  if (!response.ok) {
    throw new Error(`Failed to delete channel: ${response.statusText}`);
  }
};

/**
 * Test channel connection
 */
export const testConnection = async (id: string): Promise<ChannelTestResult> => {
  const response = await fetch(`${API_BASE}/channels/${id}/test`);

  if (!response.ok) {
    throw new Error(`Failed to test connection: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Get channel health status
 */
export const getChannelHealth = async (id: string): Promise<HealthStatus> => {
  const response = await fetch(`${API_BASE}/channels/${id}/health`);

  if (!response.ok) {
    throw new Error(`Failed to fetch health: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Toggle channel enabled/disabled state
 */
export const toggleChannel = async (id: string, enabled: boolean): Promise<Channel> => {
  return updateChannel(id, { enabled });
};

/**
 * Get recent channel activity
 */
export const getChannelActivity = async (
  id: string,
  limit: number = 50
): Promise<ChannelActivity[]> => {
  const response = await fetch(`${API_BASE}/channels/${id}/activity?limit=${limit}`);

  if (!response.ok) {
    throw new Error(`Failed to fetch activity: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Connect a channel (establish connection)
 */
export const connectChannel = async (id: string): Promise<Channel> => {
  const response = await fetch(`${API_BASE}/channels/${id}/connect`, {
    method: "POST",
  });

  if (!response.ok) {
    throw new Error(`Failed to connect channel: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Disconnect a channel
 */
export const disconnectChannel = async (id: string): Promise<Channel> => {
  const response = await fetch(`${API_BASE}/channels/${id}/disconnect`, {
    method: "POST",
  });

  if (!response.ok) {
    throw new Error(`Failed to disconnect channel: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Get available channel types
 */
export const getChannelTypes = async (): Promise<
  Array<{ type: string; name: string; description: string }>
> => {
  const response = await fetch(`${API_BASE}/channels/types`);

  if (!response.ok) {
    throw new Error(`Failed to fetch channel types: ${response.statusText}`);
  }

  return response.json();
};
