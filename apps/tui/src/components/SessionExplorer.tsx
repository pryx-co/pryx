import { createSignal, For, createEffect, onCleanup } from "solid-js";
import { Effect, Stream, Fiber } from "effect";
import { useEffectService } from "../lib/hooks";
import { WebSocketService } from "../services/ws";

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
        <box flexDirection="column" flexGrow={1}>
            <box marginBottom={1}>
                <text fg="cyan">Session Explorer</text>
                <text fg="gray"> ({filteredSessions().length} sessions)</text>
            </box>

            <box borderStyle="single" marginBottom={1} padding={1}>
                <text fg="gray">🔍 </text>
                <box flexGrow={1}>
                    {searchQuery() ? (
                        <text fg="white">{searchQuery()}</text>
                    ) : (
                        <text fg="gray">Search sessions...</text>
                    )}
                </box>
            </box>

            <box flexDirection="column" flexGrow={1} borderStyle="rounded" padding={1}>
                {loading() ? (
                    <text fg="gray">Loading sessions...</text>
                ) : filteredSessions().length === 0 ? (
                    <text fg="gray">No sessions found</text>
                ) : (
                    <For each={filteredSessions()}>
                        {(session, index) => (
                            <box flexDirection="row" gap={1}>
                                <text fg={index() === selectedIndex() ? "cyan" : "white"}>
                                    {index() === selectedIndex() ? "▶" : " "}
                                </text>
                                <text fg={index() === selectedIndex() ? "cyan" : "white"}>
                                    {session.title.slice(0, 40)}
                                </text>
                                <text fg="gray">|</text>
                                <text fg="gray">{formatDate(session.updatedAt)}</text>
                                <text fg="gray">|</text>
                                <text fg="yellow">{formatCost(session.cost)}</text>
                                <text fg="gray">|</text>
                                <text fg="green">{formatTokens(session.tokens)} tok</text>
                            </box>
                        )}
                    </For>
                )}
            </box>

            <box marginTop={1}>
                <text fg="gray">↑↓ Navigate │ Enter Resume</text>
            </box>
        </box>
    );
}
