// @ts-nocheck
import { Box, Text } from "@opentui/core";
import { createSignal, onMount, Switch, Match } from "solid-js";
import Chat from "./Chat";
import SessionExplorer from "./SessionExplorer";
import { connect, send } from "../services/ws";

type View = "chat" | "sessions";

export default function App() {
    const [status, setStatus] = createSignal("Connecting...");
    const [view, setView] = createSignal<View>("chat");
    const [pending, setPending] = createSignal<null | {
        approval_id: string;
        tool: string;
        reason?: string;
    }>(null);

    onMount(() => {
        connect(
            (s) => setStatus(s),
            (evt) => {
                if (evt?.event === "approval.needed" && evt?.payload?.approval_id) {
                    setPending({
                        approval_id: evt.payload.approval_id,
                        tool: evt.payload.tool ?? "unknown",
                        reason: evt.payload.reason,
                    });
                }
            }
        );
    });

    return (
        <Box flexDirection="column" padding={1} borderStyle="single">
            <Box flexDirection="row" justifyContent="space-between" marginBottom={1}>
                <Box>
                    <Text bold>Pryx TUI</Text>
                    <Text color="gray"> │ </Text>
                    <Text
                        color={view() === "chat" ? "cyan" : "gray"}
                        bold={view() === "chat"}
                    >
                        Chat
                    </Text>
                    <Text color="gray"> │ </Text>
                    <Text
                        color={view() === "sessions" ? "cyan" : "gray"}
                        bold={view() === "sessions"}
                    >
                        Sessions
                    </Text>
                </Box>
                <Text color="green">{status()}</Text>
            </Box>

            <Switch>
                <Match when={view() === "chat"}>
                    <Chat />
                </Match>
                <Match when={view() === "sessions"}>
                    <SessionExplorer />
                </Match>
            </Switch>

            <Box marginTop={1} borderStyle="single" padding={1}>
                <Text color="gray">Tab: Switch view │ q: Quit</Text>
            </Box>
        </Box>
    );
}
