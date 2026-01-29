import { createSignal, For, createEffect, onCleanup } from "solid-js";
import { Effect, Stream, Fiber } from "effect";

import { useEffectService } from "../lib/hooks";
import { WebSocketService } from "../services/ws";
import Message, { MessageProps } from "./Message";

type RuntimeEvent = any;

export default function Chat() {
    const ws = useEffectService(WebSocketService);
    const [messages, setMessages] = createSignal<MessageProps[]>([]);
    const [inputValue, setInputValue] = createSignal("");
    const [sessionId] = createSignal(crypto.randomUUID());
    const [pendingApproval, setPendingApproval] = createSignal<{ id: string, description: string } | null>(null);
    const [isStreaming, setIsStreaming] = createSignal(false);
    const [streamingContent, setStreamingContent] = createSignal("");

    createEffect(() => {
        const service = ws();
        if (!service) return;

        const connectFiber = Effect.runFork(service.connect);
        
        const messageFiber = Effect.runFork(
            service.messages.pipe(
                Stream.runForEach((evt) => Effect.sync(() => handleEvent(evt as RuntimeEvent)))
            )
        );

        onCleanup(() => {
            Effect.runFork(Fiber.interrupt(connectFiber));
            Effect.runFork(Fiber.interrupt(messageFiber));
            Effect.runFork(service.disconnect);
        });
    });

    const handleEvent = (evt: RuntimeEvent) => {
        switch (evt.event) {
            case "message.delta":
                setIsStreaming(true);
                setStreamingContent(prev => prev + (evt.payload?.content ?? ""));
                break;
            case "message.done":
                setIsStreaming(false);
                setMessages(prev => [...prev, {
                    type: "assistant",
                    content: streamingContent(),
                    pending: false
                }]);
                setStreamingContent("");
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

    const handleSubmit = () => {
        const value = inputValue();
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
                setInputValue("");
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
                setInputValue("");
                return;
            }
        }

        setMessages((prev) => [...prev, { type: "user", content: value }]);
        Effect.runFork(service.send({
            type: "chat.message",
            sessionId: sessionId(),
            content: value
        }));
        setInputValue("");
        setIsStreaming(true);
    };



    const displayMessages = () => [...messages()];

    return (
        <box flexDirection="column" flexGrow={1}>
            <box 
                flexDirection="column" 
                flexGrow={1} 
                borderStyle="single" 
                borderColor="cyan" 
                padding={1}
                gap={1}
            >
                <For each={displayMessages()}>
                    {(msg) => <Message {...msg} />}
                </For>
                
                {isStreaming() && streamingContent() && (
                    <Message 
                        type="assistant" 
                        content={streamingContent()} 
                        pending={true}
                    />
                )}
            </box>

            {pendingApproval() && (
                <box 
                    borderStyle="double" 
                    borderColor="yellow" 
                    padding={1} 
                    marginTop={1}
                    flexDirection="row"
                >
                    <text fg="yellow">⚠️ {pendingApproval()!.description}</text>
                    <box flexGrow={1} />
                    <text fg="gray">(y/n)</text>
                </box>
            )}

            <box marginTop={1}>
                <input
                    placeholder="Type a message... (Enter to send)"
                    value={inputValue()}
                    onChange={(v: string) => setInputValue(v)}
                    onSubmit={handleSubmit}
                />
            </box>
        </box>
    );
}
