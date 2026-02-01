import { createSignal, createEffect, For, Show, onMount } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";
import type { Channel, ChannelActivity, HealthStatus } from "../types/channels";
import {
  CHANNEL_TYPE_LABELS,
  CHANNEL_STATUS_LABELS,
  CHANNEL_STATUS_COLORS,
} from "../types/channels";
import {
  getChannel,
  updateChannel,
  deleteChannel,
  testConnection,
  getChannelHealth,
  getChannelActivity,
  toggleChannel,
} from "../services/channels";

interface ChannelDetailProps {
  channelId: string;
  onBack: () => void;
  onDelete: () => void;
}

type ViewMode = "view" | "edit" | "delete_confirm" | "test_result";

const PLATFORM_ICONS: Record<string, string> = {
  webhook: "🔌",
  telegram: "✈️",
  discord: "🎮",
  slack: "💬",
  email: "📧",
  whatsapp: "📱",
};

export default function ChannelDetail(props: ChannelDetailProps) {
  const [viewMode, setViewMode] = createSignal<ViewMode>("view");
  const [channel, setChannel] = createSignal<Channel | null>(null);
  const [health, setHealth] = createSignal<HealthStatus | null>(null);
  const [activities, setActivities] = createSignal<ChannelActivity[]>([]);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");
  const [success, setSuccess] = createSignal("");
  const [testResult, setTestResult] = createSignal<{
    success: boolean;
    message: string;
  } | null>(null);
  const [editData, setEditData] = createSignal<Partial<Channel>>({});
  const [focusedTab, setFocusedTab] = createSignal(0);

  const tabs = ["Overview", "Configuration", "Activity", "Health"];

  onMount(() => {
    loadChannelData();
  });

  const loadChannelData = async () => {
    setLoading(true);
    setError("");
    try {
      const [channelData, healthData, activityData] = await Promise.all([
        getChannel(props.channelId),
        getChannelHealth(props.channelId).catch(() => null),
        getChannelActivity(props.channelId, 20).catch(() => []),
      ]);
      setChannel(channelData);
      setEditData({
        name: channelData.name,
        enabled: channelData.enabled,
      });
      setHealth(healthData);
      setActivities(activityData);
    } catch (e) {
      setError("Failed to load channel data");
    } finally {
      setLoading(false);
    }
  };

  const handleToggleEnable = async () => {
    if (!channel()) return;
    try {
      await toggleChannel(channel()!.id, !channel()!.enabled);
      setSuccess(
        channel()!.enabled ? `✓ ${channel()!.name} disabled` : `✓ ${channel()!.name} enabled`
      );
      await loadChannelData();
      setTimeout(() => setSuccess(""), 2000);
    } catch (e) {
      setError("Failed to toggle channel");
    }
  };

  const handleTestConnection = async () => {
    setLoading(true);
    setError("");
    try {
      const result = await testConnection(props.channelId);
      setTestResult({
        success: result.success,
        message: result.message,
      });
      setViewMode("test_result");
      await loadChannelData();
    } catch (e) {
      setError("Failed to test connection");
    } finally {
      setLoading(false);
    }
  };

  const handleSaveEdit = async () => {
    if (!channel()) return;
    setLoading(true);
    setError("");
    try {
      await updateChannel(channel()!.id, editData());
      setSuccess("✓ Changes saved");
      setViewMode("view");
      await loadChannelData();
      setTimeout(() => setSuccess(""), 2000);
    } catch (e) {
      setError("Failed to save changes");
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    setLoading(true);
    setError("");
    try {
      await deleteChannel(props.channelId);
      props.onDelete();
    } catch (e) {
      setError("Failed to delete channel");
      setLoading(false);
    }
  };

  useKeyboard(evt => {
    if (viewMode() === "view") {
      switch (evt.name) {
        case "escape":
        case "b": {
          evt.preventDefault();
          props.onBack();
          break;
        }
        case "left":
        case "arrowleft": {
          evt.preventDefault();
          setFocusedTab(i => Math.max(0, i - 1));
          break;
        }
        case "right":
        case "arrowright": {
          evt.preventDefault();
          setFocusedTab(i => Math.min(tabs.length - 1, i + 1));
          break;
        }
        case "e": {
          evt.preventDefault();
          handleToggleEnable();
          break;
        }
        case "t": {
          evt.preventDefault();
          handleTestConnection();
          break;
        }
        case "d": {
          evt.preventDefault();
          setViewMode("delete_confirm");
          break;
        }
        case "r": {
          evt.preventDefault();
          loadChannelData();
          break;
        }
      }
    } else if (viewMode() === "edit") {
      switch (evt.name) {
        case "escape": {
          evt.preventDefault();
          setViewMode("view");
          setEditData({
            name: channel()?.name,
            enabled: channel()?.enabled,
          });
          break;
        }
        case "return":
        case "enter": {
          evt.preventDefault();
          handleSaveEdit();
          break;
        }
      }
    } else if (viewMode() === "delete_confirm") {
      switch (evt.name) {
        case "y": {
          handleDelete();
          break;
        }
        case "n":
        case "escape": {
          evt.preventDefault();
          setViewMode("view");
          break;
        }
      }
    } else if (viewMode() === "test_result") {
      switch (evt.name) {
        case "escape":
        case "return":
        case "enter": {
          evt.preventDefault();
          setViewMode("view");
          break;
        }
      }
    }
  });

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString();
  };

  const renderOverview = () => (
    <box flexDirection="column" gap={1}>
      <box flexDirection="row">
        <text fg={palette.dim}>Type: </text>
        <text fg={palette.text}>
          {PLATFORM_ICONS[channel()?.type || ""]}{" "}
          {CHANNEL_TYPE_LABELS[channel()?.type || "webhook"]}
        </text>
      </box>

      <box flexDirection="row">
        <text fg={palette.dim}>Status: </text>
        <text fg={CHANNEL_STATUS_COLORS[channel()?.status || "disconnected"]}>
          {CHANNEL_STATUS_LABELS[channel()?.status || "disconnected"]}
        </text>
      </box>

      <box flexDirection="row">
        <text fg={palette.dim}>Enabled: </text>
        <text fg={channel()?.enabled ? palette.success : palette.dim}>
          {channel()?.enabled ? "✓ Yes" : "✗ No"}
        </text>
      </box>

      <box flexDirection="row">
        <text fg={palette.dim}>ID: </text>
        <text fg={palette.dim}>{channel()?.id}</text>
      </box>

      <box flexDirection="row">
        <text fg={palette.dim}>Created: </text>
        <text fg={palette.dim}>{formatDate(channel()?.createdAt || "")}</text>
      </box>

      <box flexDirection="row">
        <text fg={palette.dim}>Updated: </text>
        <text fg={palette.dim}>{formatDate(channel()?.updatedAt || "")}</text>
      </box>

      <Show when={health()}>
        <box marginTop={1} borderStyle="single" borderColor={palette.border} padding={1}>
          <box flexDirection="row" marginBottom={1}>
            <text fg={palette.accent}>Health Status</text>
          </box>
          <box flexDirection="row">
            <text fg={palette.dim}>Status: </text>
            <text fg={health()?.healthy ? palette.success : palette.error}>
              {health()?.healthy ? "✓ Healthy" : "✗ Unhealthy"}
            </text>
          </box>
          <Show when={health()?.message}>
            <box flexDirection="row">
              <text fg={palette.dim}>Message: </text>
              <text fg={palette.text}>{health()?.message}</text>
            </box>
          </Show>
          <Show when={health()?.lastError}>
            <box flexDirection="row">
              <text fg={palette.dim}>Last Error: </text>
              <text fg={palette.error}>{health()?.lastError}</text>
            </box>
          </Show>
        </box>
      </Show>
    </box>
  );

  const renderConfiguration = () => {
    const config = channel()?.config;
    if (!config) return null;

    return (
      <box flexDirection="column" gap={1}>
        <For each={Object.entries(config)}>
          {([key, value]) => (
            <box flexDirection="row">
              <box width={20}>
                <text fg={palette.dim}>{key}:</text>
              </box>
              <box flexGrow={1}>
                <Show
                  when={typeof value === "object" && value !== null}
                  fallback={
                    <text fg={palette.text}>
                      {key.includes("token") || key.includes("password")
                        ? "•".repeat(String(value).length)
                        : String(value)}
                    </text>
                  }
                >
                  <text fg={palette.dim}>[Object]</text>
                </Show>
              </box>
            </box>
          )}
        </For>
      </box>
    );
  };

  const renderActivity = () => (
    <box flexDirection="column" gap={1}>
      <Show when={activities().length === 0}>
        <text fg={palette.dim}>No recent activity</text>
      </Show>
      <For each={activities()}>
        {activity => (
          <box flexDirection="row" borderStyle="single" borderColor={palette.border} padding={1}>
            <box width={20}>
              <text fg={palette.dim}>{formatDate(activity.timestamp)}</text>
            </box>
            <box width={20}>
              <text
                fg={
                  activity.type === "error"
                    ? palette.error
                    : activity.type === "message_received"
                      ? palette.success
                      : palette.text
                }
              >
                {activity.type}
              </text>
            </box>
            <box flexGrow={1}>
              <text fg={palette.text}>{activity.content}</text>
            </box>
          </box>
        )}
      </For>
    </box>
  );

  const renderHealth = () => (
    <box flexDirection="column" gap={1}>
      <Show when={!health()}>
        <text fg={palette.dim}>No health data available</text>
      </Show>
      <Show when={health()}>
        <box flexDirection="row">
          <text fg={palette.dim}>Health: </text>
          <text fg={health()?.healthy ? palette.success : palette.error}>
            {health()?.healthy ? "✓ Healthy" : "✗ Unhealthy"}
          </text>
        </box>
        <box flexDirection="row">
          <text fg={palette.dim}>Message: </text>
          <text fg={palette.text}>{health()?.message}</text>
        </box>
        <Show when={health()?.lastError}>
          <box flexDirection="row">
            <text fg={palette.dim}>Last Error: </text>
            <text fg={palette.error}>{health()?.lastError}</text>
          </box>
        </Show>
        <Show when={health()?.timestamp}>
          <box flexDirection="row">
            <text fg={palette.dim}>Last Check: </text>
            <text fg={palette.dim}>{formatDate(health()?.timestamp || "")}</text>
          </box>
        </Show>
      </Show>
    </box>
  );

  return (
    <box
      position="absolute"
      top={2}
      left="10%"
      width="80%"
      height="80%"
      borderStyle="single"
      borderColor={palette.border}
      backgroundColor={palette.bgPrimary}
      flexDirection="column"
      padding={1}
    >
      {/* Header */}
      <box flexDirection="row" marginBottom={1}>
        <box flexDirection="row">
          <text fg={palette.accent}>{PLATFORM_ICONS[channel()?.type || ""]}</text>
          <box width={1} />
          <text fg={palette.accent}>{channel()?.name}</text>
        </box>
        <box flexGrow={1} />
        <text fg={palette.dim}>[Esc to go back]</text>
      </box>

      {/* Error/Success */}
      <Show when={error()}>
        <box marginBottom={1}>
          <text fg={palette.error}>✗ {error()}</text>
        </box>
      </Show>
      <Show when={success()}>
        <box marginBottom={1}>
          <text fg={palette.success}>{success()}</text>
        </box>
      </Show>

      <Show when={viewMode() === "view"}>
        {/* Tabs */}
        <box flexDirection="row" marginBottom={1}>
          <For each={tabs}>
            {(tab, index) => (
              <box
                borderStyle={focusedTab() === index() ? "double" : "single"}
                borderColor={focusedTab() === index() ? palette.accent : palette.border}
                padding={{ left: 1, right: 1 }}
                marginRight={1}
              >
                <text fg={focusedTab() === index() ? palette.accent : palette.dim}>{tab}</text>
              </box>
            )}
          </For>
        </box>

        {/* Content */}
        <box flexDirection="column" flexGrow={1}>
          <Show when={loading()}>
            <text fg={palette.accent}>Loading...</text>
          </Show>

          <Show when={!loading() && channel()}>
            <Show when={focusedTab() === 0}>{renderOverview()}</Show>
            <Show when={focusedTab() === 1}>{renderConfiguration()}</Show>
            <Show when={focusedTab() === 2}>{renderActivity()}</Show>
            <Show when={focusedTab() === 3}>{renderHealth()}</Show>
          </Show>
        </box>

        {/* Footer */}
        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>
            ←→ Tabs | E Toggle | T Test | D Delete | R Refresh | Esc Back
          </text>
        </box>
      </Show>

      {/* Delete Confirmation */}
      <Show when={viewMode() === "delete_confirm"}>
        <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center">
          <text fg={palette.error} marginBottom={1}>
            ⚠ Delete Channel?
          </text>
          <text fg={palette.text} marginBottom={1}>
            Are you sure you want to remove {channel()?.name}?
          </text>
          <text fg={palette.dim} marginBottom={2}>
            This action cannot be undone.
          </text>
          <box flexDirection="row" gap={2}>
            <box borderStyle="single" borderColor={palette.error} padding={1}>
              <text fg={palette.error}>Y - Yes, Delete</text>
            </box>
            <box borderStyle="single" borderColor={palette.border} padding={1}>
              <text fg={palette.dim}>N - Cancel</text>
            </box>
          </box>
        </box>
      </Show>

      {/* Test Result */}
      <Show when={viewMode() === "test_result" && testResult()}>
        <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center">
          <Show when={testResult()?.success}>
            <text fg={palette.success} marginBottom={1}>
              ✓ Connection Test Passed
            </text>
          </Show>
          <Show when={!testResult()?.success}>
            <text fg={palette.error} marginBottom={1}>
              ✗ Connection Test Failed
            </text>
          </Show>
          <text fg={palette.text} marginBottom={2}>
            {testResult()?.message}
          </text>
          <text fg={palette.dim}>Press Enter or Esc to continue</text>
        </box>
      </Show>
    </box>
  );
}
