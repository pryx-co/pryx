import { createSignal, createEffect, For, Show, onMount, onCleanup } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";
import { McpServer, McpService } from "../services/mcp";
import McpCurated from "./McpCurated";
import McpAddServer from "./McpAddServer";

type ViewMode = "list" | "add" | "curated" | "details" | "delete_confirm";

interface McpServersProps {
  onClose: () => void;
}

// Security rating colors
const ratingColors: Record<string, string> = {
  A: palette.success,
  B: palette.info,
  C: palette.accent,
  D: palette.warning,
  F: palette.error,
};

// Status colors
const statusColors: Record<string, string> = {
  connected: palette.success,
  error: palette.error,
  disabled: palette.dim,
  connecting: palette.accent,
};

export default function McpServers(props: McpServersProps) {
  const [viewMode, setViewMode] = createSignal<ViewMode>("list");
  const [servers, setServers] = createSignal<McpServer[]>([]);
  const [selectedIndex, setSelectedIndex] = createSignal(0);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");
  const [success, setSuccess] = createSignal("");
  const [selectedServer, setSelectedServer] = createSignal<McpServer | null>(null);
  const [mcpService] = createSignal(new McpService());

  onMount(() => {
    loadServers();
  });

  const loadServers = async () => {
    setLoading(true);
    setError("");
    try {
      const data = await mcpService().getServers();
      setServers(data);
    } catch (e) {
      setError("Failed to load MCP servers");
    } finally {
      setLoading(false);
    }
  };

  const handleToggleEnable = async (server: McpServer) => {
    try {
      await mcpService().toggleServer(server.id, !server.enabled);
      setSuccess(server.enabled ? `✓ ${server.name} disabled` : `✓ ${server.name} enabled`);
      await loadServers();
      setTimeout(() => setSuccess(""), 2000);
    } catch (e) {
      setError("Failed to toggle server");
    }
  };

  const handleDeleteServer = async (serverId: string) => {
    try {
      await mcpService().deleteServer(serverId);
      setSuccess("✓ Server removed");
      setViewMode("list");
      await loadServers();
      setTimeout(() => setSuccess(""), 2000);
    } catch (e) {
      setError("Failed to delete server");
    }
  };

  const handleServerAdded = () => {
    setSuccess("✓ Server added successfully");
    setViewMode("list");
    loadServers();
    setTimeout(() => setSuccess(""), 2000);
  };

  useKeyboard(evt => {
    if (viewMode() === "list") {
      switch (evt.name) {
        case "up":
        case "arrowup": {
          evt.preventDefault();
          setSelectedIndex(i => Math.max(0, i - 1));
          break;
        }
        case "down":
        case "arrowdown": {
          evt.preventDefault();
          const maxIndex = servers().length + 2;
          setSelectedIndex(i => Math.min(maxIndex, i + 1));
          break;
        }
        case "return":
        case "enter": {
          evt.preventDefault();
          const idx = selectedIndex();
          const srvs = servers();

          if (idx < srvs.length) {
            setSelectedServer(srvs[idx]);
            setViewMode("details");
          } else if (idx === srvs.length) {
            setViewMode("curated");
          } else if (idx === srvs.length + 1) {
            setViewMode("add");
          } else {
            props.onClose();
          }
          break;
        }
        case "escape":
          evt.preventDefault();
          props.onClose();
          break;
        case "e": {
          const idx = selectedIndex();
          const srvs = servers();
          if (idx < srvs.length) {
            evt.preventDefault();
            handleToggleEnable(srvs[idx]);
          }
          break;
        }
        case "d": {
          const idx = selectedIndex();
          const srvs = servers();
          if (idx < srvs.length) {
            evt.preventDefault();
            setSelectedServer(srvs[idx]);
            setViewMode("delete_confirm");
          }
          break;
        }
      }
    } else if (viewMode() === "details") {
      switch (evt.name) {
        case "escape":
          evt.preventDefault();
          setViewMode("list");
          break;
        case "e":
          if (selectedServer()) {
            handleToggleEnable(selectedServer()!);
          }
          break;
        case "d":
          if (selectedServer()) {
            setViewMode("delete_confirm");
          }
          break;
      }
    } else if (viewMode() === "delete_confirm") {
      switch (evt.name) {
        case "y":
          if (selectedServer()) {
            handleDeleteServer(selectedServer()!.id);
          }
          break;
        case "n":
        case "escape":
          evt.preventDefault();
          setViewMode(selectedServer() ? "details" : "list");
          break;
      }
    }
  });

  createEffect(() => {
    if (viewMode() === "list") {
      setSelectedIndex(0);
    }
  });

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
      <box flexDirection="row" marginBottom={1}>
        <text fg={palette.accent}>MCP Server Management</text>
        <box flexGrow={1} />
        <text fg={palette.dim}>[Esc to close]</text>
      </box>

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

      <Show when={viewMode() === "list"}>
        <box flexDirection="column" flexGrow={1}>
          <text fg={palette.dim} marginBottom={1}>
            Configured Servers ({servers().length})
          </text>

          <Show when={loading()}>
            <box marginBottom={1}>
              <text fg={palette.accent}>Loading...</text>
            </box>
          </Show>

          <box flexDirection="column" flexGrow={1}>
            <For each={servers()}>
              {(server, index) => (
                <box
                  flexDirection="row"
                  padding={1}
                  backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
                >
                  <box width={3}>
                    <text
                      fg={server.enabled ? statusColors[server.status] || palette.dim : palette.dim}
                    >
                      {server.enabled ? "●" : "○"}
                    </text>
                  </box>
                  <box width={20}>
                    <text fg={index() === selectedIndex() ? palette.accent : palette.text}>
                      {server.name}
                    </text>
                  </box>
                  <box width={12}>
                    <text fg={statusColors[server.status] || palette.dim}>{server.status}</text>
                  </box>
                  <box width={8}>
                    <text fg={ratingColors[server.securityRating] || palette.dim}>
                      [{server.securityRating}]
                    </text>
                  </box>
                  <box flexGrow={1}>
                    <text fg={palette.dim}>{server.tools.length} tools</text>
                  </box>
                  <box width={10}>
                    <Show when={!server.enabled}>
                      <text fg={palette.dim}>[DISABLED]</text>
                    </Show>
                  </box>
                </box>
              )}
            </For>

            <box
              flexDirection="row"
              padding={1}
              marginTop={1}
              backgroundColor={
                selectedIndex() === servers().length ? palette.bgSelected : undefined
              }
            >
              <box width={3}>
                <text fg={palette.accent}>+</text>
              </box>
              <box>
                <text fg={selectedIndex() === servers().length ? palette.accent : palette.text}>
                  Browse Curated Servers
                </text>
              </box>
            </box>

            <box
              flexDirection="row"
              padding={1}
              backgroundColor={
                selectedIndex() === servers().length + 1 ? palette.bgSelected : undefined
              }
            >
              <box width={3}>
                <text fg={palette.accent}>+</text>
              </box>
              <box>
                <text fg={selectedIndex() === servers().length + 1 ? palette.accent : palette.text}>
                  Add Custom Server
                </text>
              </box>
            </box>

            <box
              flexDirection="row"
              padding={1}
              backgroundColor={
                selectedIndex() === servers().length + 2 ? palette.bgSelected : undefined
              }
            >
              <box width={3}>
                <text fg={palette.dim}>×</text>
              </box>
              <box>
                <text fg={selectedIndex() === servers().length + 2 ? palette.accent : palette.dim}>
                  Close
                </text>
              </box>
            </box>
          </box>
        </box>

        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>↑↓ Navigate | Enter Select | E Toggle | D Delete | Esc Close</text>
        </box>
      </Show>

      <Show when={viewMode() === "curated"}>
        <McpCurated
          onSelect={server => {
            handleServerAdded();
          }}
          onClose={() => setViewMode("list")}
        />
      </Show>

      <Show when={viewMode() === "add"}>
        <McpAddServer onAdded={handleServerAdded} onCancel={() => setViewMode("list")} />
      </Show>

      <Show when={viewMode() === "details" && selectedServer()}>
        <box flexDirection="column" flexGrow={1}>
          <text fg={palette.accent} marginBottom={1}>
            {selectedServer()!.name}
          </text>

          <box flexDirection="column" gap={1}>
            <box flexDirection="row">
              <text fg={palette.dim}>Status: </text>
              <text fg={statusColors[selectedServer()!.status] || palette.dim}>
                {selectedServer()!.status}
              </text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>Security Rating: </text>
              <text fg={ratingColors[selectedServer()!.securityRating] || palette.dim}>
                {selectedServer()!.securityRating}
              </text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>Enabled: </text>
              <text fg={selectedServer()!.enabled ? palette.success : palette.dim}>
                {selectedServer()!.enabled ? "Yes" : "No"}
              </text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>Transport: </text>
              <text fg={palette.text}>{selectedServer()!.transport}</text>
            </box>

            <box flexDirection="row">
              <text fg={palette.dim}>URL: </text>
              <text fg={palette.dim}>{selectedServer()!.url}</text>
            </box>

            <box marginTop={1}>
              <text fg={palette.dim}>Tools ({selectedServer()!.tools.length}):</text>
              <box flexDirection="column" marginLeft={2} marginTop={1}>
                <For each={selectedServer()!.tools}>
                  {tool => <text fg={palette.dim}>• {tool.name}</text>}
                </For>
              </box>
            </box>
          </box>

          <box flexGrow={1} />

          <box flexDirection="column" marginTop={1}>
            <text fg={palette.dim}>E Toggle Enable | D Delete | Esc Back</text>
          </box>
        </box>
      </Show>

      <Show when={viewMode() === "delete_confirm"}>
        <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center">
          <text fg={palette.error} marginBottom={1}>
            ⚠ Delete Server?
          </text>
          <text fg={palette.text} marginBottom={1}>
            Are you sure you want to remove {selectedServer()?.name}?
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
    </box>
  );
}
