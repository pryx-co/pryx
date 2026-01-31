import { createSignal, createEffect, For, Show, onMount, onCleanup } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";
import type { Channel, ChannelType, ChannelStatus } from "../types/channels";
import {
  getChannels,
  deleteChannel,
  toggleChannel,
  testConnection,
} from "../services/channels";
import {
  CHANNEL_TYPE_LABELS,
  CHANNEL_STATUS_LABELS,
  CHANNEL_STATUS_COLORS,
} from "../types/channels";

type ViewMode = "list" | "add" | "details" | "delete_confirm" | "test_result";

interface ChannelManagerProps {
  onClose: () => void;
}

// Platform icons
const PLATFORM_ICONS: Record<ChannelType, string> = {
  webhook: "🔌",
  telegram: "✈️",
  discord: "🎮",
  slack: "💬",
  email: "📧",
  whatsapp: "📱",
};

// Status colors using palette
const STATUS_COLORS: Record<ChannelStatus, string> = {
  connected: palette.success,
  disconnected: palette.warning,
  error: palette.error,
  pending_setup: palette.dim,
};

export default function ChannelManager(props: ChannelManagerProps) {
  const [viewMode, setViewMode] = createSignal<ViewMode>("list");
  const [channels, setChannels] = createSignal<Channel[]>([]);
  const [selectedIndex, setSelectedIndex] = createSignal(0);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");
  const [success, setSuccess] = createSignal("");
  const [selectedChannel, setSelectedChannel] = createSignal<Channel | null>(null);
  const [testResult, setTestResult] = createSignal<{ success: boolean; message: string } | null>(
    null
  );

  onMount(() => {
    loadChannels();
  });

  const loadChannels = async () => {
    setLoading(true);
    setError("");
    try {
      const data = await getChannels();
      setChannels(data);
    } catch (e) {
      setError("Failed to load channels");
    } finally {
      setLoading(false);
    }
  };

  const handleToggleEnable = async (channel: Channel) => {
    try {
      await toggleChannel(channel.id, !channel.enabled);
      setSuccess(
        channel.enabled
          ? `✓ ${channel.name} disabled`
          : `✓ ${channel.name} enabled`
      );
      await loadChannels();
      setTimeout(() => setSuccess(""), 2000);
    } catch (e) {
      setError("Failed to toggle channel");
    }
  };

  const handleDeleteChannel = async (channelId: string) => {
    try {
      await deleteChannel(channelId);
      setSuccess("✓ Channel removed");
      setViewMode("list");
      await loadChannels();
      setTimeout(() => setSuccess(""), 2000);
    } catch (e) {
      setError("Failed to delete channel");
    }
  };

  const handleTestConnection = async (channel: Channel) => {
    setLoading(true);
    setError("");
    try {
      const result = await testConnection(channel.id);
      setTestResult({
        success: result.success,
        message: result.message,
      });
      setViewMode("test_result");
    } catch (e) {
      setError("Failed to test connection");
    } finally {
      setLoading(false);
    }
  };

  useKeyboard((evt) => {
    if (viewMode() === "list") {
      switch (evt.name) {
        case "up":
        case "arrowup": {
          evt.preventDefault();
          setSelectedIndex((i) => Math.max(0, i - 1));
          break;
        }
        case "down":
        case "arrowdown": {
          evt.preventDefault();
          const maxIndex = channels().length + 1;
          setSelectedIndex((i) => Math.min(maxIndex, i + 1));
          break;
        }
        case "return":
        case "enter": {
          evt.preventDefault();
          const idx = selectedIndex();
          const chans = channels();

          if (idx < chans.length) {
            setSelectedChannel(chans[idx]);
            setViewMode("details");
          } else if (idx === chans.length) {
            setViewMode("add");
          } else {
            props.onClose();
          }
          break;
        }
        case "escape":
        case "b": {
          evt.preventDefault();
          props.onClose();
          break;
        }
        case "e": {
          const idx = selectedIndex();
          const chans = channels();
          if (idx < chans.length) {
            evt.preventDefault();
            handleToggleEnable(chans[idx]);
          }
          break;
        }
        case "t": {
          const idx = selectedIndex();
          const chans = channels();
          if (idx < chans.length) {
            evt.preventDefault();
            handleTestConnection(chans[idx]);
          }
          break;
        }
        case "d": {
          const idx = selectedIndex();
          const chans = channels();
          if (idx < chans.length) {
            evt.preventDefault();
            setSelectedChannel(chans[idx]);
            setViewMode("delete_confirm");
          }
          break;
        }
        case "a": {
          evt.preventDefault();
          setViewMode("add");
          break;
        }
      }
    } else if (viewMode() === "details") {
      switch (evt.name) {
        case "escape":
        case "b": {
          evt.preventDefault();
          setViewMode("list");
          break;
        }
        case "e": {
          if (selectedChannel()) {
            handleToggleEnable(selectedChannel()!);
          }
          break;
        }
        case "t": {
          if (selectedChannel()) {
            handleTestConnection(selectedChannel()!);
          }
          break;
        }
        case "d": {
          if (selectedChannel()) {
            setViewMode("delete_confirm");
          }
          break;
        }
      }
    } else if (viewMode() === "delete_confirm") {
      switch (evt.name) {
        case "y": {
          if (selectedChannel()) {
            handleDeleteChannel(selectedChannel()!.id);
          }
          break;
        }
        case "n":
        case "escape": {
          evt.preventDefault();
          setViewMode(selectedChannel() ? "details" : "list");
          break;
        }
      }
    } else if (viewMode() === "test_result") {
      switch (evt.name) {
        case "escape":
        case "return":
        case "enter": {
          evt.preventDefault();
          setViewMode("list");
          break;
        }
      }
    }
  });

  createEffect(() => {
    if (viewMode() === "list") {
      setSelectedIndex(0);
    }
  });

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString();
  };

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
        <box flexDirection="column">
          <text fg={palette.accent}>Channel Management</text>
          <text fg={palette.dim}>{channels().length} channels configured</text>
        </box>
        <box flexGrow={1} />
        <box flexDirection="column" alignItems="flex-end">
          <box
            borderStyle="single"
            borderColor={palette.accent}
            padding={{ left: 1, right: 1 }}
          >
            <text fg={palette.accent}>A - Add Channel</text>
          </box>
          <text fg={palette.dim}>[Esc to close]</text>
        </box>
      </box>

      {/* Error/Success Messages */}
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

      {/* List View */}
      <Show when={viewMode() === "list"}>
        <box flexDirection="column" flexGrow={1}>
          <Show when={loading()}>
            <box marginBottom={1}>
              <text fg={palette.accent}>Loading channels...</text>
            </box>
          </Show>

          <Show when={channels().length === 0 && !loading()}>
            <box
              flexDirection="column"
              flexGrow={1}
              alignItems="center"
              justifyContent="center"
            >
              <text fg={palette.dim}>No channels configured</text>
              <box marginTop={1}>
                <text fg={palette.accent}>Press 'A' to add your first channel</text>
              </box>
            </box>
          </Show>

          <Show when={channels().length > 0}>
            <box flexDirection="column" flexGrow={1}>
              {/* Column Headers */}
              <box flexDirection="row" padding={1} borderStyle="single" borderColor={palette.border}
              >
                <box width={4}>
                  <text fg={palette.dim}></text>
                </box>
                <box width={20}>
                  <text fg={palette.dim}>Name</text>
                </box>
                <box width={12}>
                  <text fg={palette.dim}>Type</text>
                </box>
                <box width={14}>
                  <text fg={palette.dim}>Status</text>
                </box>
                <box width={10}>
                  <text fg={palette.dim}>Enabled</text>
                </box>
                <box flexGrow={1}>
                  <text fg={palette.dim}>Last Updated</text>
                </box>
              </box>

              {/* Channel List */}
              <For each={channels()}>
                {(channel, index) => (
                  <box
                    flexDirection="row"
                    padding={1}
                    backgroundColor={
                      index() === selectedIndex() ? palette.bgSelected : undefined
                    }
                  >
                    <box width={4}>
                      <text>{PLATFORM_ICONS[channel.type]}</text>
                    </box>
                    <box width={20}>
                      <text
                        fg={
                          index() === selectedIndex() ? palette.accent : palette.text
                        }
                      >
                        {channel.name}
                      </text>
                    </box>
                    <box width={12}>
                      <text fg={palette.dim}>
                        {CHANNEL_TYPE_LABELS[channel.type]}
                      </text>
                    </box>
                    <box width={14}>
                      <box flexDirection="row">
                        <text fg={STATUS_COLORS[channel.status]}>● </text>
                        <text fg={STATUS_COLORS[channel.status]}>
                          {CHANNEL_STATUS_LABELS[channel.status]}
                        </text>
                      </box>
                    </box>
                    <box width={10}>
                      <text fg={channel.enabled ? palette.success : palette.dim}>
                        {channel.enabled ? "✓ Yes" : "✗ No"}
                      </text>
                    </box>
                    <box flexGrow={1}>
                      <text fg={palette.dim}>{formatDate(channel.updatedAt)}</text>
                    </box>
                  </box>
                )}
              </For>
            </box>
          </Show>

          {/* Add Channel Option */}
          <box
            flexDirection="row"
            padding={1}
            marginTop={channels().length > 0 ? 1 : 0}
            backgroundColor={
              selectedIndex() === channels().length ? palette.bgSelected : undefined
            }
          >
            <box width={4}>
              <text fg={palette.accent}>+</text>
            </box>
            <box>
              <text
                fg={
                  selectedIndex() === channels().length
                    ? palette.accent
                    : palette.text
                }
              >
                Add New Channel
              </text>
            </box>
          </box>

          {/* Close Option */}
          <box
            flexDirection="row"
            padding={1}
            backgroundColor={
              selectedIndex() === channels().length + 1
                ? palette.bgSelected
                : undefined
            }
          >
            <box width={4}>
              <text fg={palette.dim}>×</text>
            </box>
            <box>
              <text
                fg={
                  selectedIndex() === channels().length + 1
                    ? palette.accent
                    : palette.dim
                }
              >
                Back to Main Menu
              </text>
            </box>
          </box>
        </box>

        {/* Footer */}
        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>
            ↑↓ Navigate | Enter Select | E Toggle | T Test | D Delete | A Add | Esc
            Back
          </text>
        </box>
      </Show>

      {/* Add Channel View */}
      <Show when={viewMode() === "add"}>
        <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center"
        >
          <text fg={palette.accent} marginBottom={1}>Add Channel</text>
          <text fg={palette.dim}>Channel wizard coming soon...</text>
          <box marginTop={2}>
            <text fg={palette.dim}>Press Esc to go back</text>
          </box>
        </box>
      </Show>

      {/* Details View */}
      <Show when={viewMode() === "details" && selectedChannel()}>
        <box flexDirection="column" flexGrow={1}>
          <box flexDirection="row" marginBottom={1}>
            <text fg={palette.accent}>{PLATFORM_ICONS[selectedChannel()!.type]}</text>
            <box width={1} />
            <text fg={palette.accent}>{selectedChannel()!.name}</text>
          </box>

          <box flexDirection="column" gap={1}>
            <box flexDirection="row">
              <text fg={palette.dim}>Type: </text>
              <text fg={palette.text}>
                {CHANNEL_TYPE_LABELS[selectedChannel()!.type]}
              </text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>Status: </text>
              <text fg={STATUS_COLORS[selectedChannel()!.status]}>
                {CHANNEL_STATUS_LABELS[selectedChannel()!.status]}
              </text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>Enabled: </text>
              <text
                fg={selectedChannel()!.enabled ? palette.success : palette.dim}
              >
                {selectedChannel()!.enabled ? "Yes" : "No"}
              </text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>ID: </text>
              <text fg={palette.dim}>{selectedChannel()!.id}</text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>Created: </text>
              <text fg={palette.dim}>
                {formatDate(selectedChannel()!.createdAt)}
              </text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>Updated: </text>
              <text fg={palette.dim}>
                {formatDate(selectedChannel()!.updatedAt)}
              </text>
            </box>
          </box>

          <box flexGrow={1} />

          <box flexDirection="column" marginTop={1}>
            <text fg={palette.dim}>E Toggle | T Test | D Delete | Esc Back</text>
          </box>
        </box>
      </Show>

      {/* Delete Confirmation */}
      <Show when={viewMode() === "delete_confirm"}>
        <box
          flexDirection="column"
          flexGrow={1}
          alignItems="center"
          justifyContent="center"
        >
          <text fg={palette.error} marginBottom={1}>⚠ Delete Channel?</text>
          <text fg={palette.text} marginBottom={1}>
            Are you sure you want to remove {selectedChannel()?.name}?
          </text>
          <box flexDirection="row" gap={2} marginTop={1}>
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
        <box
          flexDirection="column"
          flexGrow={1}
          alignItems="center"
          justifyContent="center"
        >
          <Show when={testResult()?.success}>
            <text fg={palette.success} marginBottom={1}>✓ Connection Test Passed</text>
          </Show>
          <Show when={!testResult()?.success}>
            <text fg={palette.error} marginBottom={1}>✗ Connection Test Failed</text>
          </Show>
          <text fg={palette.text} marginBottom={1}>{testResult()?.message}</text>
          <box marginTop={2}>
            <text fg={palette.dim}>Press Enter or Esc to continue</text>
          </box>
        </box>
      </Show>
    </box>
  );
}
