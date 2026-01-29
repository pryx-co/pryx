import { createSignal, createEffect, onMount, onCleanup, Switch, Match, Show } from "solid-js";
import { useRenderer } from "@opentui/solid";
import { useEffectService } from "../lib/hooks";
import { WebSocketService } from "../services/ws";
import AppHeader from "./AppHeader";
import Chat from "./Chat";
import SessionExplorer from "./SessionExplorer";
import Settings from "./Settings";
import Channels from "./Channels";
import Skills from "./Skills";
import CommandPalette from "./CommandPalette";
import KeyboardShortcuts from "./KeyboardShortcuts";

type View = "chat" | "sessions" | "settings" | "channels" | "skills";

export default function App() {
    const renderer = useRenderer();
    renderer.disableStdoutInterception();
    
    const ws = useEffectService(WebSocketService);
    const [view, setView] = createSignal<View>("chat");
    const [showCommands, setShowCommands] = createSignal(false);
    const [showHelp, setShowHelp] = createSignal(false);
    const [connectionStatus, setConnectionStatus] = createSignal("Connecting...");
    const [hasProvider, setHasProvider] = createSignal(false);

    createEffect(() => {
        const service = ws();
        if (!service) {
            setConnectionStatus("Runtime Error");
            return;
        }

        const checkStatus = async () => {
            try {
                const apiUrl = process.env.PRYX_API_URL || "http://localhost:3000";
                const res = await fetch(`${apiUrl}/health`, { method: "GET" });
                if (res.ok) {
                    const data = await res.json();
                    if (data.providers?.length > 0) {
                        setHasProvider(true);
                        setConnectionStatus("Ready");
                    } else {
                        setHasProvider(false);
                        setConnectionStatus("No Provider");
                    }
                } else {
                    setConnectionStatus("Runtime Error");
                }
            } catch {
                setConnectionStatus("Disconnected");
            }
        };

        checkStatus();
        const interval = setInterval(checkStatus, 5000);
        return () => clearInterval(interval);
    });

    const commands = [
        { id: "chat", label: "Chat", shortcut: "c", action: () => { setView("chat"); setShowCommands(false); } },
        { id: "sessions", label: "Sessions", shortcut: "s", action: () => { setView("sessions"); setShowCommands(false); } },
        { id: "channels", label: "Channels", shortcut: "n", action: () => { setView("channels"); setShowCommands(false); } },
        { id: "settings", label: "Settings", shortcut: ",", action: () => { setView("settings"); setShowCommands(false); } },
        { id: "skills", label: "Skills", shortcut: "k", action: () => { setView("skills"); setShowCommands(false); } },
        { id: "help", label: "Help", shortcut: "?", action: () => { setShowHelp(true); setShowCommands(false); } },
        { id: "quit", label: "Quit", shortcut: "q", action: () => process.exit(0) },
    ];

    const views: View[] = ["chat", "sessions", "channels", "skills", "settings"];

    const handleKey = (data: Buffer) => {
        const key = data.toString();
        
        // Handle help screen
        if (showHelp()) {
            if (key === '\u001b' || key === 'q') {
                setShowHelp(false);
            }
            return;
        }

        // Handle command palette
        if (showCommands()) {
            if (key === '\u001b') {
                setShowCommands(false);
                return;
            }
            const cmd = commands.find(c => c.shortcut === key);
            if (cmd) {
                cmd.action();
            }
            return;
        }

        // Global shortcuts
        switch (key) {
            case '/':
                setShowCommands(true);
                break;
            case '?':
                setShowHelp(true);
                break;
            case '\t':
                // Tab - next view
                setView(prev => {
                    const idx = views.indexOf(prev);
                    return views[(idx + 1) % views.length];
                });
                break;
            case '\u001b[Z':
                // Shift+Tab - previous view
                setView(prev => {
                    const idx = views.indexOf(prev);
                    return views[(idx - 1 + views.length) % views.length];
                });
                break;
            case '\u0003':
                // Ctrl+C
                process.exit(0);
                break;
            case '\u000c':
                break;
            case '1':
            case '2':
            case '3':
            case '4':
            case '5': {
                const idx = parseInt(key) - 1;
                if (idx < views.length) {
                    setView(views[idx]);
                }
                break;
            }
        }
    };

    onMount(() => {
        if (typeof process !== "undefined" && process.stdin.isTTY) {
            process.stdin.on("data", handleKey);
        }
    });

    onCleanup(() => {
        if (typeof process !== "undefined" && process.stdin) {
            process.stdin.off("data", handleKey);
        }
    });

    const getStatusColor = () => {
        if (connectionStatus() === "Ready") return "green";
        if (connectionStatus() === "Connecting...") return "yellow";
        return "red";
    };

    return (
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
                        <Chat />
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
                <CommandPalette 
                    commands={commands} 
                    onClose={() => setShowCommands(false)} 
                />
            </Show>

            <Show when={showHelp()}>
                <KeyboardShortcuts onClose={() => setShowHelp(false)} />
            </Show>
        </box>
    );
}
