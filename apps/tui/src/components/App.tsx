import { Box, Text } from "@opentui/core";
import { createSignal, onMount, onCleanup, Switch, Match, createEffect } from "solid-js";
import { Effect, Stream } from "effect";
import { useEffectService, useEffectStream, TUIRuntime } from "../lib/hooks";
import { WebSocketService } from "../services/ws";
import AppHeader from "./AppHeader";
import Chat from "./Chat";
import SessionExplorer from "./SessionExplorer";
import Settings from "./Settings";
import Channels from "./Channels";
import Skills from "./Skills";
import Notifications, { Notification } from "./Notifications";

type View = "chat" | "sessions" | "settings" | "channels" | "skills";

export default function App() {
    const [status, setStatus] = createSignal("Connecting...");
    const [view, setView] = createSignal<View>("chat");
    // ... notifications code unchanged ...
    const [notifications, setNotifications] = createSignal<Notification[]>([]);
    const [pending, setPending] = createSignal<null | {
        approval_id: string;
        tool: string;
        reason?: string;
    }>(null);

    const addNotification = (message: string, type: Notification["type"] = "info") => {
        const id = Math.random().toString(36).substr(2, 9);
        const newNotif: Notification = { id, message, type, timestamp: Date.now() };
        setNotifications(prev => [...prev.slice(-4), newNotif]); // Keep max 5
        setTimeout(() => {
            setNotifications(prev => prev.filter(n => n.id !== id));
        }, 5000);
    };
    const handleGlobalKey = (data: Buffer) => {
        // Tab key navigation
        if (data.toString() === '\t') {
            setView(prev => {
                if (prev === "chat") return "sessions";
                if (prev === "sessions") return "channels";
                if (prev === "channels") return "skills";
                if (prev === "skills") return "settings";
                return "chat";
            });
        }
    };

    // Service access
    const ws = useEffectService(WebSocketService);

    // Subscribe to status
    const connectionStatus = useEffectStream(
        Stream.unwrap(
            Effect.if(
                Effect.sync(() => {
                    const s = ws();
                    return s ? s.status : Stream.empty;
                }),
                {
                    onTrue: (s) => s,
                    onFalse: () => Stream.empty
                }
            )
        )
    );

    // Update local status signal based on stream
    createEffect(() => {
        const s = connectionStatus();
        const last = s[s.length - 1];
        if (!last) return;

        if (last._tag === "Connected") setStatus("Connected \u2705");
        else if (last._tag === "Connecting") setStatus("Connecting...");
        else if (last._tag === "Error") setStatus(`Error: ${last.error.message}`);
        else setStatus("Disconnected \u274C");
    });

    // Handle messages (Notifications & Approval)
    const messages = useEffectStream(
        Stream.unwrap(
            Effect.if(
                Effect.sync(() => {
                    const s = ws();
                    return s ? s.messages : Stream.empty;
                }),
                {
                    onTrue: (s) => s,
                    onFalse: () => Stream.empty
                }
            )
        )
    );

    createEffect(() => {
        const msgs = messages();
        msgs.forEach(evt => {
            if (evt?.event === "approval.needed" && evt?.payload?.approval_id) {
                setPending({
                    approval_id: evt.payload.approval_id,
                    tool: evt.payload.tool ?? "unknown",
                    reason: evt.payload.reason,
                });
                addNotification(`Approval needed for ${evt.payload.tool}`, "warning");
            } else if (evt?.event === "notification.event") {
                addNotification(evt.payload.message, evt.payload.type || "info");
            }
        });
    });

    // Auto-connect on mount
    onMount(() => {
        // Run the connect effect
        Effect.runFork(
            Effect.if(
                Effect.sync(() => ws()),
                {
                    onTrue: () => ws()!.connect,
                    onFalse: () => Effect.void
                }
            )
        );

        if (typeof process !== "undefined" && process.stdin.isTTY) {
            process.stdin.on("data", handleGlobalKey);
        }
    });

    onCleanup(() => {
        // Disconnect effect
        Effect.runFork(
            Effect.if(
                Effect.sync(() => ws()),
                {
                    onTrue: () => ws()!.disconnect,
                    onFalse: () => Effect.void
                }
            )
        );

        if (typeof process !== "undefined" && process.stdin) {
            process.stdin.off("data", handleGlobalKey);
        }
    });

    return (
        <Box flexDirection="column" padding={1} borderStyle="double" borderColor="cyan">
            <AppHeader />

            <Box flexDirection="row" justifyContent="space-between" marginBottom={1} borderStyle="single" borderColor="white" padding={1}>
                <Box>
                    <Text
                        color={view() === "chat" ? "black" : "white"}
                        backgroundColor={view() === "chat" ? "cyan" : undefined}
                        bold={view() === "chat"}
                    >
                        {" Chat "}
                    </Text>
                    <Text color="gray"> │ </Text>
                    <Text
                        color={view() === "sessions" ? "black" : "white"}
                        backgroundColor={view() === "sessions" ? "cyan" : undefined}
                        bold={view() === "sessions"}
                    >
                        {" Sessions "}
                    </Text>
                    <Text color="gray"> │ </Text>
                    <Text
                        color={view() === "channels" ? "black" : "white"}
                        backgroundColor={view() === "channels" ? "cyan" : undefined}
                        bold={view() === "channels"}
                    >
                        {" Channels "}
                    </Text>
                    <Text color="gray"> │ </Text>
                    <Text
                        color={view() === "settings" ? "black" : "white"}
                        backgroundColor={view() === "settings" ? "cyan" : undefined}
                        bold={view() === "settings"}
                    >
                        {" Settings "}
                    </Text>
                    <Text color="gray"> │ </Text>
                    <Text
                        color={view() === "skills" ? "black" : "white"}
                        backgroundColor={view() === "skills" ? "cyan" : undefined}
                        bold={view() === "skills"}
                    >
                        {" Skills "}
                    </Text>
                </Box>
                <Text color={status().includes("Connected") ? "green" : "red"}>{status()}</Text>
            </Box>

            <Box flexDirection="column" flex={1} borderStyle="single" borderColor="gray" padding={1}>
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
            </Box>

            <Box marginTop={1} borderTopStyle="single" borderColor="gray" flexDirection="row" justifyContent="space-between">
                <Text color="gray">Tab: Switch View │ Ctrl+C: Quit</Text>
                <Text color="blue">v0.1.0-alpha</Text>
            </Box>

            <Notifications items={notifications()} />
        </Box>
    );
}
