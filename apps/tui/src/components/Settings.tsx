// @ts-nocheck
import { Box, Text, Input } from "@opentui/core";
import { createSignal, onMount, onCleanup, Show, For } from "solid-js";
import { loadConfig, saveConfig, Config } from "../services/config";

export default function Settings() {
    const [config, setConfig] = createSignal<any>({});
    const [fields] = createSignal([
        { key: "model_provider", label: "Model Provider", placeholder: "ollama, openai, anthropic" },
        { key: "model_name", label: "Model Name", placeholder: "llama3, gpt-4" },
        { key: "openai_key", label: "OpenAI Key", placeholder: "sk-..." },
        { key: "anthropic_key", label: "Anthropic Key", placeholder: "sk-ant-..." },
        { key: "ollama_endpoint", label: "Ollama URL", placeholder: "http://localhost:11434" },
    ]);
    const [selectedIndex, setSelectedIndex] = createSignal(0);
    const [isEditing, setIsEditing] = createSignal(false);
    const [status, setStatus] = createSignal("");

    onMount(() => {
        const loaded = loadConfig();
        setConfig(loaded);

        // Setup Key Listener
        if (typeof process !== "undefined" && process.stdin.isTTY) {
            process.stdin.on("data", handleInput);
        }
    });

    onCleanup(() => {
        if (typeof process !== "undefined" && process.stdin) {
            process.stdin.off("data", handleInput);
        }
    });

    // We use 'data' instead of 'keypress' to avoid readline dependency if possible, 
    // but raw mode handling might be complex. 
    // Actually, Input component likely puts stdin in raw mode.
    // If I attach 'data' listener, I might steal from Input?
    // Proper way: Input should be unmounted when not editing.

    const handleInput = (data: Buffer) => {
        if (isEditing()) return; // Let Input handle it

        const key = data.toString();
        // ANSI codes for arrow keys
        if (key === '\u001B\u005B\u0041') { // Up
            setSelectedIndex(prev => (prev - 1 + fields().length) % fields().length);
        } else if (key === '\u001B\u005B\u0042') { // Down
            setSelectedIndex(prev => (prev + 1) % fields().length);
        } else if (key === '\r' || key === '\n') { // Enter
            setIsEditing(true);
        }
    };

    const handleSave = async (key: string, value: string) => {
        const newConfig = { ...config(), [key]: value };
        setConfig(newConfig);
        setIsEditing(false);
        saveConfig(newConfig);
        setStatus("Saved!");
        setTimeout(() => setStatus(""), 2000);
    };

    return (
        <Box flexDirection="column" flexGrow={1}>
            <Text bold color="cyan">Configuration</Text>
            <Text color="gray">Config Path: ~/.pryx/config.yaml</Text>
            <Box marginTop={1} flexDirection="column" borderStyle="round" padding={1}>
                <For each={fields()}>
                    {(field, index) => (
                        <Box flexDirection="row" marginBottom={0}>
                            <Text color={index() === selectedIndex() ? "cyan" : "gray"}>
                                {index() === selectedIndex() ? "❯ " : "  "}
                            </Text>
                            <Box width={20}>
                                <Text bold={index() === selectedIndex()}>
                                    {field.label}:
                                </Text>
                            </Box>

                            <Show when={isEditing() && index() === selectedIndex()} fallback={
                                <Text color="white">{config()[field.key] || <Text color="gray" italic>empty</Text>}</Text>
                            }>
                                <Input
                                    value={config()[field.key] || ""}
                                    placeholder={field.placeholder}
                                    onSubmit={(val) => handleSave(field.key, val)}
                                />
                            </Show>
                        </Box>
                    )}
                </For>
            </Box>
            <Box marginTop={1}>
                <Text color="green">{status()}</Text>
            </Box>
            <Text color="gray">↑↓ Select │ Enter Edit/Save</Text>
        </Box>
    );
}
