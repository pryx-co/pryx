import { createSignal, onMount, onCleanup, Show, For } from "solid-js";
import { loadConfig, saveConfig } from "../services/config";

interface ChannelField {
  id: string;
  type: "header" | "toggle" | "input";
  key?: string;
  label: string;
  placeholder?: string;
}

export default function Channels() {
  const [config, setConfig] = createSignal<any>({});
  const [fields] = createSignal<ChannelField[]>([
    { id: "h1", type: "header", label: "TELEGRAM BOT" },
    { id: "tg_status", type: "toggle", key: "telegram_enabled", label: "Status" },
    {
      id: "tg_token",
      type: "input",
      key: "telegram_token",
      label: "Bot Token",
      placeholder: "123456:ABC-...",
    },
    { id: "sep1", type: "header", label: " " },
    { id: "h2", type: "header", label: "SLACK APP" },
    { id: "sl_status", type: "toggle", key: "slack_enabled", label: "Status" },
    {
      id: "sl_app_token",
      type: "input",
      key: "slack_app_token",
      label: "App Token",
      placeholder: "xoxb-...",
    },
    {
      id: "sl_bot_token",
      type: "input",
      key: "slack_bot_token",
      label: "Bot Token",
      placeholder: "xoxb-...",
    },
    {
      id: "sl_mode",
      type: "select",
      key: "slack_mode",
      label: "Mode",
    },
    { id: "sep2", type: "header", label: " " },
    { id: "h3", type: "header", label: "GENERIC WEBHOOK" },
    { id: "wh_status", type: "toggle", key: "webhook_enabled", label: "Status" },
  ]);
  const [selectedIndex, setSelectedIndex] = createSignal(1);
  const [isEditing, setIsEditing] = createSignal(false);
  const [status, setStatus] = createSignal("");

  const moveSelection = (dir: 1 | -1) => {
    let next = selectedIndex();
    const len = fields().length;
    for (let i = 0; i < len; i++) {
      next = (next + dir + len) % len;
      if (fields()[next].type !== "header") break;
    }
    setSelectedIndex(next);
  };

  const handleInput = (data: Buffer) => {
    if (isEditing()) return;

    const key = data.toString();
    if (key === "\u001B\u005B\u0041") {
      moveSelection(-1);
    } else if (key === "\u001B\u005B\u0042") {
      moveSelection(1);
    } else if (key === "\r" || key === "\n") {
      const field = fields()[selectedIndex()];
      if (field.type === "input") {
        setIsEditing(true);
      } else if (field.type === "toggle") {
        toggleValue(field.key!);
      }
    } else if (key === " ") {
      const field = fields()[selectedIndex()];
      if (field.type === "toggle") {
        toggleValue(field.key!);
      }
    }
  };

  const toggleValue = async (key: string) => {
    const val = !config()[key];
    await handleSave(key, val);
  };

  const handleSave = async (key: string, value: any) => {
    const newConfig = { ...config(), [key]: value };
    setConfig(newConfig);
    setIsEditing(false);
    saveConfig(newConfig);
    setStatus("Saved!");
    setTimeout(() => setStatus(""), 2000);
  };

  onMount(async () => {
    const loaded = loadConfig();
    setConfig(loaded);
    if (typeof process !== "undefined" && process.stdin.isTTY) {
      process.stdin.on("data", handleInput);
    }
  });

  onCleanup(() => {
    if (typeof process !== "undefined" && process.stdin) {
      process.stdin.off("data", handleInput);
    }
  });

  const renderValue = (field: ChannelField) => {
    const val = config()[field.key!];
    if (field.type === "toggle") {
      return val ? <text fg="green">ENABLED</text> : <text fg="gray">DISABLED</text>;
    }
    if (!val) return <text fg="gray">empty</text>;
    if (field.key?.includes("token")) {
      return (
        <text>
          {val.substring(0, 4)}...{val.substring(val.length - 4)}
        </text>
      );
    }
    return <text>{val}</text>;
  };

  return (
    <box flexDirection="column" flexGrow={1}>
      <text fg="magenta">Channel Setup</text>
      <text fg="gray">Config Path: ~/.pryx/config.yaml</text>

      <box marginTop={1} flexDirection="column" borderStyle="rounded" padding={1}>
        <For each={fields()}>
          {(field, index) => {
            if (field.type === "header") {
              return (
                <box marginTop={field.label === " " ? 0 : 1} marginBottom={0}>
                  <text fg="cyan">{field.label}</text>
                </box>
              );
            }

            const isSelected = index() === selectedIndex();
            return (
              <box flexDirection="row" marginBottom={0}>
                <text fg={isSelected ? "cyan" : "gray"}>{isSelected ? "❯ " : "  "}</text>
                <box width={15}>
                  <text>{field.label}:</text>
                </box>

                <Show when={isEditing() && isSelected} fallback={renderValue(field)}>
                  <box>
                    <text fg="cyan">▌{config()[field.key!] || ""}</text>
                  </box>
                </Show>
              </box>
            );
          }}
        </For>
      </box>

      <box marginTop={1}>
        <text fg="green">{status()}</text>
      </box>
      <text fg="gray">↑↓ Select │ Enter/Space Toggle │ Enter Edit</text>
    </box>
  );
}
