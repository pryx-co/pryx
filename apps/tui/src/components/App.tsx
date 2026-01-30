import { createSignal, createEffect, onMount, onCleanup, Switch, Match, Show } from "solid-js";
import { useRenderer, useKeyboard } from "@opentui/solid";
import { useEffectService, AppRuntime } from "../lib/hooks";
import { WebSocketService } from "../services/ws";
import { HealthCheckService } from "../services/health-check";
import { loadConfig } from "../services/config";
import AppHeader from "./AppHeader";
import Chat from "./Chat";
import SessionExplorer from "./SessionExplorer";
import Settings from "./Settings";
import Channels from "./Channels";
import Skills from "./Skills";
import SearchableCommandPalette, { Command } from "./SearchableCommandPalette";
import KeyboardShortcuts from "./KeyboardShortcuts";
import SetupRequired from "./SetupRequired";
import ProviderManager from "./ProviderManager";

type View = "chat" | "sessions" | "settings" | "channels" | "skills";

export default function App() {
  const renderer = useRenderer();
  renderer.disableStdoutInterception();

  const ws = useEffectService(WebSocketService);
  const healthCheck = useEffectService(HealthCheckService);
  const [view, setView] = createSignal<View>("chat");
  const [showCommands, setShowCommands] = createSignal(false);
  const [showHelp, setShowHelp] = createSignal(false);
  const [showProviderManager, setShowProviderManager] = createSignal(false);
  const [connectionStatus, setConnectionStatus] = createSignal("Connecting...");
  const [hasProvider, setHasProvider] = createSignal(false);
  const [setupRequired, setSetupRequired] = createSignal(false);

  onMount(() => {
    const config = loadConfig();
    const hasValidProvider =
      config.model_provider &&
      (config.openai_key || config.anthropic_key || config.glm_key || config.ollama_endpoint);
    if (!hasValidProvider) {
      setSetupRequired(true);
    }
  });

  const handleSetupComplete = () => {
    setSetupRequired(false);
    setHasProvider(true);
    setConnectionStatus("Ready");
  };

  createEffect(() => {
    const service = healthCheck();
    if (!service) {
      setConnectionStatus("Runtime Error");
      return;
    }

    const pollInterval = 5000;

    AppRuntime.runFork(
      service.pollHealth(pollInterval, (result) => {
        if (result.status === "ok") {
          if (result.providers && result.providers.length > 0) {
            setHasProvider(true);
            setConnectionStatus("Ready");
          } else {
            setHasProvider(false);
            setConnectionStatus("No Provider");
          }
        } else {
          setConnectionStatus("Disconnected");
        }
      })
    );
  });

  const allCommands: Command[] = [
    {
      id: "chat",
      name: "Chat",
      description: "Open chat interface",
      category: "Navigation",
      shortcut: "1",
      keywords: ["chat", "talk", "message", "conversation"],
      action: () => {
        setView("chat");
        setShowCommands(false);
      },
    },
    {
      id: "sessions",
      name: "Sessions",
      description: "Browse and manage sessions",
      category: "Navigation",
      shortcut: "2",
      keywords: ["sessions", "history", "conversations", "browse"],
      action: () => {
        setView("sessions");
        setShowCommands(false);
      },
    },
    {
      id: "channels",
      name: "Channels",
      description: "Manage channel integrations",
      category: "Navigation",
      shortcut: "3",
      keywords: ["channels", "telegram", "discord", "slack", "webhooks", "integrations"],
      action: () => {
        setView("channels");
        setShowCommands(false);
      },
    },
    {
      id: "skills",
      name: "Skills",
      description: "Browse and manage skills",
      category: "Navigation",
      shortcut: "4",
      keywords: ["skills", "abilities", "tools", "capabilities"],
      action: () => {
        setView("skills");
        setShowCommands(false);
      },
    },
    {
      id: "settings",
      name: "Settings",
      description: "Configure Pryx",
      category: "Navigation",
      shortcut: "5",
      keywords: ["settings", "config", "preferences", "options"],
      action: () => {
        setView("settings");
        setShowCommands(false);
      },
    },
    {
      id: "new-chat",
      name: "New Chat",
      description: "Start a new conversation",
      category: "Chat",
      keywords: ["new", "chat", "conversation", "start", "fresh"],
      action: () => {
        setView("chat");
        setShowCommands(false);
      },
    },
    {
      id: "clear-chat",
      name: "Clear Chat",
      description: "Clear current conversation",
      category: "Chat",
      keywords: ["clear", "reset", "clean", "chat"],
      action: () => {
        setShowCommands(false);
      },
    },
    {
      id: "help",
      name: "Keyboard Shortcuts",
      description: "Show all keyboard shortcuts",
      category: "System",
      shortcut: "?",
      keywords: ["help", "shortcuts", "keys", "commands", "?"],
      action: () => {
        setShowHelp(true);
        setShowCommands(false);
      },
    },
    {
      id: "quit",
      name: "Quit",
      description: "Exit Pryx",
      category: "System",
      shortcut: "q",
      keywords: ["quit", "exit", "close", "stop"],
      action: () => process.exit(0),
    },
    {
      id: "reload",
      name: "Reload",
      description: "Refresh connection",
      category: "System",
      keywords: ["reload", "refresh", "reconnect", "restart"],
      action: () => {
        setShowCommands(false);
      },
    },
    {
      id: "providers",
      name: "Manage Providers",
      description: "Add, edit, or remove AI providers",
      category: "System",
      shortcut: "p",
      keywords: ["providers", "connect", "api", "keys", "models", "ai"],
      action: () => {
        setShowProviderManager(true);
        setShowCommands(false);
      },
    },
  ];

  const views: View[] = ["chat", "sessions", "channels", "skills", "settings"];

  useKeyboard(evt => {
    if (showHelp() || showCommands() || showProviderManager()) {
      return;
    }

    switch (evt.name) {
      case "/":
        evt.preventDefault();
        setShowCommands(true);
        break;
      case "?":
        evt.preventDefault();
        setShowHelp(true);
        break;
      case "tab":
        evt.preventDefault();
        setView(prev => {
          const idx = views.indexOf(prev);
          return views[(idx + 1) % views.length];
        });
        break;
      case "1":
      case "2":
      case "3":
      case "4":
      case "5": {
        evt.preventDefault();
        const idx = parseInt(evt.name) - 1;
        if (idx < views.length) {
          setView(views[idx]);
        }
        break;
      }
      case "c":
        if (evt.ctrl) {
          evt.preventDefault();
          process.exit(0);
        }
        break;
    }
  });

  const getStatusColor = () => {
    if (connectionStatus() === "Ready") return "green";
    if (connectionStatus() === "Connecting...") return "yellow";
    return "red";
  };

  return (
    <Show
      when={!setupRequired()}
      fallback={<SetupRequired onSetupComplete={handleSetupComplete} />}
    >
      <box flexDirection="column" backgroundColor="#0a0a0a" flexGrow={1}>
        <AppHeader />

        <box flexDirection="row" padding={1} gap={1}>
          <text fg="gray">/</text>
          <text fg="gray">commands</text>
          <box flexGrow={1} />
          <Show when={!hasProvider()}>
            <text fg="yellow">⚠️ No Provider</text>
          </Show>
          <text fg={getStatusColor()}>{connectionStatus()}</text>
        </box>

        <box flexGrow={1} padding={1}>
          <Switch>
            <Match when={view() === "chat"}>
              <Chat 
                disabled={showCommands() || showHelp() || showProviderManager()} 
                onConnectCommand={() => setShowProviderManager(true)}
              />
            </Match>
            <Match when={view() === "sessions"}>
              <SessionExplorer />
            </Match>
            <Match when={view() === "channels"}>
              <Channels />
            </Match>
            <Match when={view() === "settings"}>
              <Settings />
            </Match>
            <Match when={view() === "skills"}>
              <Skills />
            </Match>
          </Switch>
        </box>

        <box flexDirection="row" padding={1}>
          <text fg="gray">/: Commands | Tab: Switch | 1-5: Views | ?: Help | Ctrl+C: Quit</text>
          <box flexGrow={1} />
          <text fg="blue">v0.1.0-alpha</text>
        </box>

        <Show when={showCommands()}>
          <SearchableCommandPalette
            commands={allCommands}
            onClose={() => setShowCommands(false)}
            placeholder="Type to search commands..."
          />
        </Show>

        <Show when={showHelp()}>
          <KeyboardShortcuts onClose={() => setShowHelp(false)} />
        </Show>

        <Show when={showProviderManager()}>
          <ProviderManager onClose={() => setShowProviderManager(false)} />
        </Show>
      </box>
    </Show>
  );
}
