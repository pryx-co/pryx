import { Box, Text, Input } from "@opentui/core";
import { createSignal, For, createEffect, onCleanup } from "solid-js";
import { Effect, Stream, Fiber } from "effect";
import { useEffectService, TUIRuntime } from "../lib/hooks";
import { WebSocketService } from "../services/ws";
import Message, { MessageProps } from "./Message";

// Define event types locally or import from ws service types if available
type RuntimeEvent = any;

export default function Chat() {
    const ws = useEffectService(WebSocketService);
    const [messages, setMessages] = createSignal<MessageProps[]>([]);
    const [input, setInput] = createSignal("");
    // Status is managed by App.tsx, but Chat can display if needed. 
    // For now we assume connected if we are receiving messages.
    const [sessionId] = createSignal(crypto.randomUUID());
    const [pendingApproval, setPendingApproval] = createSignal<{ id: string, description: string } | null>(null);

    createEffect(() => {
        const service = ws();
        if (!service) return;

        console.log("Chat: Subscribing to messages");
        const fiber = Effect.runFork(
            service.messages.pipe(
                Stream.runForEach((evt) => Effect.sync(() => handleEvent(evt as RuntimeEvent)))
            )
        ); // Removed improper TUIRuntime arg here too if Effect.runFork doesn't take it curried

        onCleanup(() => {
            Effect.runFork(Fiber.interrupt(fiber));
        });
    });

    const handleEvent = (evt: RuntimeEvent) => {
        switch (evt.event) {
            case "message.delta":
                setMessages((prev) => {
                    const last = prev[prev.length - 1];
                    if (last && last.type === "assistant" && last.pending) {
                        return [...prev.slice(0, -1), {
                            ...last,
                            content: last.content + (evt.payload?.content ?? "")
                        }];
                    }
                    return [...prev, {
                        type: "assistant",
                        content: evt.payload?.content ?? "",
                        pending: true
                    }];
                });
                break;
            case "message.done":
                setMessages((prev) => {
                    const last = prev[prev.length - 1];
                    if (last && last.pending) {
                        return [...prev.slice(0, -1), { ...last, pending: false }];
                    }
                    return prev;
                });
                break;
            case "tool.start":
                setMessages((prev) => [...prev, {
                    type: "tool",
                    content: "Running...",
                    toolName: evt.payload?.name,
                    toolStatus: "running"
                }]);
                break;
            case "tool.end":
                setMessages((prev) => {
                    const idx = prev.findLastIndex(m => m.toolName === evt.payload?.name && m.toolStatus === "running");
                    if (idx >= 0) {
                        const updated = [...prev];
                        updated[idx] = {
                            ...updated[idx],
                            content: evt.payload?.result ?? "Done",
                            toolStatus: evt.payload?.error ? "error" : "done"
                        };
                        return updated;
                    }
                    return prev;
                });
                break;
            case "approval.request":
                setPendingApproval({
                    id: evt.payload?.approval_id,
                    description: evt.payload?.description ?? "Action requires approval"
                });
                break;
        }
    };

    const handleSubmit = (value: string) => {
        if (!value.trim()) return;
        const service = ws();
        if (!service) return;

        if (pendingApproval()) {
            const approval = pendingApproval()!;
            if (value.toLowerCase() === "y" || value.toLowerCase() === "yes") {
                Effect.runFork(service.send({
                    type: "approval.response",
                    sessionId: sessionId(),
                    approvalId: approval.id,
                    approved: true
                }));
                setMessages((prev) => [...prev, { type: "system", content: "✅ Approved" }]);
                setPendingApproval(null);
                setInput("");
                return;
            } else if (value.toLowerCase() === "n" || value.toLowerCase() === "no") {
                Effect.runFork(service.send({
                    type: "approval.response",
                    sessionId: sessionId(),
                    approvalId: approval.id,
                    approved: false
                }));
                setMessages((prev) => [...prev, { type: "system", content: "❌ Denied" }]);
                setPendingApproval(null);
                setInput("");
                return;
            }
        }

        setMessages((prev) => [...prev, { type: "user", content: value }]);
        Effect.runFork(service.send({
            type: "chat.message",
            sessionId: sessionId(),
            content: value
        }));
        setInput("");
    };

    return (
        <Box flexDirection="column" flexGrow={1}>
            <Box flexDirection="column" flexGrow={1} borderStyle="single" borderColor="cyan" padding={1}>
                <For each={messages()}>
                    {(msg) => <Message {...msg} />}
                </For>
            </Box>
            {pendingApproval() && (
                <Box borderStyle="double" borderColor="yellow" padding={1} marginTop={1}>
                    <Text color="yellow">⚠️ {pendingApproval()!.description}</Text>
                    <Text color="gray"> (y/n): </Text>
                </Box>
            )}
            <Box borderStyle="single" borderColor="gray" marginTop={0} paddingLeft={1}>
                <Text color="cyan">❯ </Text>
                <Input
                    placeholder="Type a message..."
                    value={input()}
                    onChange={setInput}
                    onSubmit={handleSubmit}
                />
            </Box>
        </Box>
    );
}
