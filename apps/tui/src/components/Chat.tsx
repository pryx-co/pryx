// @ts-nocheck
import { Box, Text, Input } from "@opentui/core";
import { createSignal, For, createEffect } from "solid-js";
import { connect, sendChat, sendApproval, RuntimeEvent } from "../services/ws";
import Message, { MessageProps, MessageType } from "./Message";

export default function Chat() {
    const [messages, setMessages] = createSignal<MessageProps[]>([]);
    const [input, setInput] = createSignal("");
    const [status, setStatus] = createSignal("Connecting...");
    const [sessionId] = createSignal(crypto.randomUUID());
    const [pendingApproval, setPendingApproval] = createSignal<{ id: string, description: string } | null>(null);

    createEffect(() => {
        connect(setStatus, handleEvent);
    });

    const handleEvent = (evt: RuntimeEvent) => {
        switch (evt.event) {
            case "message.delta":
                // Streaming text delta
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
                // Mark streaming complete
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

        // Handle approval shortcuts
        if (pendingApproval()) {
            const approval = pendingApproval()!;
            if (value.toLowerCase() === "y" || value.toLowerCase() === "yes") {
                sendApproval(sessionId(), approval.id, true);
                setMessages((prev) => [...prev, { type: "system", content: "✅ Approved" }]);
                setPendingApproval(null);
                setInput("");
                return;
            } else if (value.toLowerCase() === "n" || value.toLowerCase() === "no") {
                sendApproval(sessionId(), approval.id, false);
                setMessages((prev) => [...prev, { type: "system", content: "❌ Denied" }]);
                setPendingApproval(null);
                setInput("");
                return;
            }
        }

        setMessages((prev) => [...prev, { type: "user", content: value }]);
        sendChat(sessionId(), value);
        setInput("");
    };

    return (
        <Box flexDirection="column" flexGrow={1}>
            <Box flexDirection="column" flexGrow={1} borderStyle="round" padding={1}>
                <For each={messages()}>
                    {(msg) => <Message {...msg} />}
                </For>
            </Box>
            {pendingApproval() && (
                <Box borderStyle="double" borderColor="yellow" padding={1}>
                    <Text color="yellow">⚠️ {pendingApproval()!.description}</Text>
                    <Text color="gray"> (y/n): </Text>
                </Box>
            )}
            <Box borderStyle="single" marginTop={1}>
                <Text color="gray">[{status()}] </Text>
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
