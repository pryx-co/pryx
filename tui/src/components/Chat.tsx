import { Box, Text, Input } from "@opentui/core";
import { createSignal } from "solid-js";

export default function Chat() {
    const [messages, setMessages] = createSignal<string[]>([]);
    const [input, setInput] = createSignal("");

    const handleSubmit = (value: string) => {
        setMessages((prev) => [...prev, `You: ${value}`]);
        setInput("");
        // TODO: Send to websocket
    };

    return (
        <Box flexDirection="column" flexGrow={1}>
            <Box flexDirection="column" flexGrow={1} borderStyle="round" padding={1}>
                {messages().map((msg) => (
                    <Text>{msg}</Text>
                ))}
            </Box>
            <Box borderStyle="single" marginTop={1}>
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
