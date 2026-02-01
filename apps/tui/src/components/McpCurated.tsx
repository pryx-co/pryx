import { createSignal, createEffect, For, Show, onMount } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";
import { McpService, CuratedServer } from "../services/mcp";

interface McpCuratedProps {
  onSelect: (server: CuratedServer) => void;
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

// Category icons
const categoryIcons: Record<string, string> = {
  filesystem: "📁",
  web: "🌐",
  database: "🗄️",
  ai: "🤖",
  utility: "🛠️",
  search: "🔍",
  communication: "💬",
};

export default function McpCurated(props: McpCuratedProps) {
  const [servers, setServers] = createSignal<CuratedServer[]>([]);
  const [filteredServers, setFilteredServers] = createSignal<CuratedServer[]>([]);
  const [selectedIndex, setSelectedIndex] = createSignal(0);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");
  const [success, setSuccess] = createSignal("");
  const [searchQuery, setSearchQuery] = createSignal("");
  const [selectedCategory, setSelectedCategory] = createSignal<string | null>(null);
  const [mcpService] = createSignal(new McpService());
  const [showConfirm, setShowConfirm] = createSignal(false);
  const [selectedServer, setSelectedServer] = createSignal<CuratedServer | null>(null);

  const categories = () => {
    const cats = new Set(servers().map(s => s.category));
    return Array.from(cats).sort();
  };

  onMount(() => {
    loadServers();
  });

  createEffect(() => {
    let filtered = servers();

    if (selectedCategory()) {
      filtered = filtered.filter(s => s.category === selectedCategory());
    }

    if (searchQuery()) {
      const query = searchQuery().toLowerCase();
      filtered = filtered.filter(
        s =>
          s.name.toLowerCase().includes(query) ||
          s.description.toLowerCase().includes(query) ||
          s.tools.some(t => t.name.toLowerCase().includes(query))
      );
    }

    setFilteredServers(filtered);
    setSelectedIndex(0);
  });

  const loadServers = async () => {
    setLoading(true);
    setError("");
    try {
      const data = await mcpService().getCuratedServers();
      setServers(data);
      setFilteredServers(data);
    } catch (_e) {
      setError("Failed to load curated servers");
    } finally {
      setLoading(false);
    }
  };

  const handleAddServer = async (server: CuratedServer) => {
    setLoading(true);
    setError("");
    try {
      await mcpService().addCuratedServer(server.id);
      setSuccess(`✓ ${server.name} added successfully`);
      setTimeout(() => {
        setSuccess("");
        props.onSelect(server);
      }, 1500);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to add server");
    } finally {
      setLoading(false);
      setShowConfirm(false);
    }
  };

  useKeyboard(evt => {
    if (showConfirm()) {
      switch (evt.name) {
        case "y":
          if (selectedServer()) {
            handleAddServer(selectedServer()!);
          }
          break;
        case "n":
        case "escape":
          evt.preventDefault();
          setShowConfirm(false);
          setSelectedServer(null);
          break;
      }
      return;
    }

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
        setSelectedIndex(i => Math.min(filteredServers().length - 1, i + 1));
        break;
      }
      case "return":
      case "enter": {
        evt.preventDefault();
        const server = filteredServers()[selectedIndex()];
        if (server) {
          setSelectedServer(server);
          setShowConfirm(true);
        }
        break;
      }
      case "escape": {
        evt.preventDefault();
        if (selectedCategory()) {
          setSelectedCategory(null);
        } else if (searchQuery()) {
          setSearchQuery("");
        } else {
          props.onClose();
        }
        break;
      }
      case "tab": {
        evt.preventDefault();
        const cats = categories();
        if (cats.length > 0) {
          const currentIdx = selectedCategory() ? cats.indexOf(selectedCategory()!) : -1;
          const nextIdx = (currentIdx + 1) % (cats.length + 1);
          if (nextIdx === cats.length) {
            setSelectedCategory(null);
          } else {
            setSelectedCategory(cats[nextIdx]);
          }
        }
        break;
      }
    }
  });

  return (
    <box flexDirection="column" flexGrow={1}>
      <text fg={palette.accent} marginBottom={1}>
        Browse Curated Servers
      </text>

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

      <Show when={!showConfirm()}>
        <box flexDirection="column" marginBottom={1}>
          <box flexDirection="row" gap={1}>
            <text fg={palette.dim}>Filter: Tab</text>
            <For each={categories()}>
              {cat => (
                <box
                  borderStyle="single"
                  borderColor={selectedCategory() === cat ? palette.accent : palette.border}
                  padding={{ left: 1, right: 1 }}
                >
                  <text fg={selectedCategory() === cat ? palette.accent : palette.dim}>
                    {categoryIcons[cat] || "•"} {cat}
                  </text>
                </box>
              )}
            </For>
            <Show when={selectedCategory()}>
              <box
                borderStyle="single"
                borderColor={palette.accent}
                padding={{ left: 1, right: 1 }}
              >
                <text fg={palette.accent}>✓ All</text>
              </box>
            </Show>
          </box>

          <box marginTop={1}>
            <text fg={palette.dim}>
              Showing {filteredServers().length} of {servers().length} servers
            </text>
          </box>
        </box>

        <Show when={loading()}>
          <box marginBottom={1}>
            <text fg={palette.accent}>Loading...</text>
          </box>
        </Show>

        <box flexDirection="column" flexGrow={1} overflow="hidden">
          <For each={filteredServers()}>
            {(server, index) => (
              <box
                flexDirection="column"
                padding={1}
                marginBottom={1}
                borderStyle="single"
                borderColor={index() === selectedIndex() ? palette.accent : palette.border}
                backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
              >
                <box flexDirection="row" marginBottom={1}>
                  <box width={3}>
                    <text>{categoryIcons[server.category] || "•"}</text>
                  </box>
                  <box width={25}>
                    <text fg={index() === selectedIndex() ? palette.accent : palette.text}>
                      {server.name}
                    </text>
                  </box>
                  <box width={8}>
                    <text fg={ratingColors[server.securityRating] || palette.dim}>
                      [{server.securityRating}]
                    </text>
                  </box>
                  <box flexGrow={1}>
                    <text fg={palette.dim}>{server.author}</text>
                  </box>
                  <box width={10}>
                    <text fg={palette.dim}>v{server.version}</text>
                  </box>
                </box>

                <box marginLeft={3}>
                  <text fg={palette.dim}>{server.description}</text>
                </box>

                <box flexDirection="row" marginLeft={3} marginTop={1}>
                  <text fg={palette.dim}>Tools: </text>
                  <For each={server.tools.slice(0, 3)}>
                    {(tool, toolIndex) => (
                      <>
                        <text fg={palette.dim}>{tool.name}</text>
                        <Show when={toolIndex() < Math.min(server.tools.length, 3) - 1}>
                          <text fg={palette.dim}>, </text>
                        </Show>
                      </>
                    )}
                  </For>
                  <Show when={server.tools.length > 3}>
                    <text fg={palette.dim}> +{server.tools.length - 3} more</text>
                  </Show>
                </box>
              </box>
            )}
          </For>
        </box>

        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>↑↓ Navigate | Enter Add | Tab Filter | Esc Back</text>
        </box>
      </Show>

      <Show when={showConfirm() && selectedServer()}>
        <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center">
          <text fg={palette.accent} marginBottom={1}>
            Add Server?
          </text>

          <box flexDirection="row" marginBottom={1}>
            <text fg={palette.text}>{selectedServer()!.name}</text>
          </box>

          <box flexDirection="row" marginBottom={1}>
            <text fg={palette.dim}>Security Rating: </text>
            <text fg={ratingColors[selectedServer()!.securityRating]}>
              {selectedServer()!.securityRating}
            </text>
          </box>

          <box flexDirection="column" marginBottom={1} alignItems="center">
            <text fg={palette.dim}>Tools ({selectedServer()!.tools.length}):</text>
            <box flexDirection="column" marginTop={1}>
              <For each={selectedServer()!.tools}>
                {tool => <text fg={palette.dim}>• {tool.name}</text>}
              </For>
            </box>
          </box>

          <box flexDirection="row" gap={2} marginTop={2}>
            <box borderStyle="single" borderColor={palette.accent} padding={1}>
              <text fg={palette.accent}>Y - Add Server</text>
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
