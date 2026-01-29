import { Box, Text, Input } from "@opentui/core";
import { createSignal, For, createEffect, onCleanup } from "solid-js";
import { Effect, Stream, Fiber } from "effect";
import { useEffectService, TUIRuntime } from "../lib/hooks";
import { WebSocketService } from "../services/ws";

// Define locally
type RuntimeEvent = any;

interface Session {
    id: string;
    title: string;
    createdAt: string;
    updatedAt: string;
    cost?: number;
    tokens?: number;
    duration?: number;
}

export default function SessionExplorer() {
    const ws = useEffectService(WebSocketService);
    const [sessions, setSessions] = createSignal<Session[]>([]);
    const [searchQuery, setSearchQuery] = createSignal("");
    const [selectedIndex, setSelectedIndex] = createSignal(0);
    const [loading, setLoading] = createSignal(true);

    createEffect(() => {
        const service = ws();
        if (!service) return;

        const fiber = Effect.runFork(service.messages.pipe(
            Stream.runForEach((evt: RuntimeEvent) => Effect.sync(() => {
                if (evt.event === "sessions.list") {
                    setSessions(evt.payload?.sessions ?? []);
                    setLoading(false);
                }
            }))
        ));

        onCleanup(() => {
            Effect.runFork(Fiber.interrupt(fiber));
        });

        // Request sessions on mount
        Effect.runFork(service.send({ event: "sessions.list", payload: {} }));
    });

    const filteredSessions = () => {
        const query = searchQuery().toLowerCase();
        if (!query) return sessions();
        return sessions().filter(s =>
            s.title.toLowerCase().includes(query) ||
            s.id.toLowerCase().includes(query)
        );
    };

    const handleSearch = (value: string) => {
        setSearchQuery(value);
        setSelectedIndex(0);
    };

    const handleSelect = () => {
        const session = filteredSessions()[selectedIndex()];
        const service = ws();
        if (session && service) {
            Effect.runFork(service.send({ event: "session.resume", payload: { session_id: session.id } }));
        }
    };

    const formatDate = (dateStr: string) => {
        const d = new Date(dateStr);
        return d.toLocaleDateString() + " " + d.toLocaleTimeString().slice(0, 5);
    };

    const formatCost = (cost?: number) => {
        if (!cost) return "-";
        return `$${cost.toFixed(4)}`;
    };

    const formatTokens = (tokens?: number) => {
        if (!tokens) return "-";
        return tokens > 1000 ? `${(tokens / 1000).toFixed(1)}k` : tokens.toString();
    };

    return (
        <Box flexDirection="column" flexGrow={1}>
            <Box marginBottom={1}>
                <Text bold color="cyan">Session Explorer</Text>
                <Text color="gray"> ({filteredSessions().length} sessions)</Text>
            </Box>

            <Box borderStyle="single" marginBottom={1}>
                <Text color="gray">🔍 </Text>
                <Input
                    placeholder="Search sessions..."
                    value={searchQuery()}
                    onChange={handleSearch}
                    onSubmit={handleSelect}
                />
            </Box>

            <Box flexDirection="column" flexGrow={1} borderStyle="rounded" padding={1}>
                {loading() ? (
                    <Text color="gray">Loading sessions...</Text>
                ) : filteredSessions().length === 0 ? (
                    <Text color="gray">No sessions found</Text>
                ) : (
                    <For each={filteredSessions()}>
                        {(session, index) => (
                            <Box>
                                <Text color={index() === selectedIndex() ? "cyan" : "white"}>
                                    {index() === selectedIndex() ? "▶ " : "  "}
                                </Text>
                                <Text color={index() === selectedIndex() ? "cyan" : "white"} bold>
                                    {session.title.slice(0, 40)}
                                </Text>
                                <Text color="gray"> │ </Text>
                                <Text color="gray">{formatDate(session.updatedAt)}</Text>
                                <Text color="gray"> │ </Text>
                                <Text color="yellow">{formatCost(session.cost)}</Text>
                                <Text color="gray"> │ </Text>
                                <Text color="green">{formatTokens(session.tokens)} tok</Text>
                            </Box>
                        )}
                    </For>
                )}
            </Box>

            <Box marginTop={1}>
                <Text color="gray">↑↓ Navigate │ Enter Resume │ e Export │ d Delete</Text>
            </Box>
        </Box>
    );
}
