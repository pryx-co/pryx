import { createSignal, onMount, onCleanup, Show, For } from "solid-js";
import { loadConfig, saveConfig } from "../services/config";

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

        if (typeof process !== "undefined" && process.stdin.isTTY) {
            process.stdin.on("data", handleInput);
        }
    });

    onCleanup(() => {
        if (typeof process !== "undefined" && process.stdin) {
            process.stdin.off("data", handleInput);
        }
    });

    const handleInput = (data: Buffer) => {
        if (isEditing()) return;

        const key = data.toString();
        if (key === '\u001B\u005B\u0041') {
            setSelectedIndex(prev => (prev - 1 + fields().length) % fields().length);
        } else if (key === '\u001B\u005B\u0042') {
            setSelectedIndex(prev => (prev + 1) % fields().length);
        } else if (key === '\r' || key === '\n') {
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
        <box flexDirection="column" flexGrow={1}>
            <text fg="cyan">Configuration</text>
            <text fg="gray">Config Path: ~/.pryx/config.yaml</text>
            <box marginTop={1} flexDirection="column" borderStyle="rounded" padding={1}>
                <For each={fields()}>
                    {(field, index) => (
                        <box flexDirection="row" marginBottom={0}>
                            <text fg={index() === selectedIndex() ? "cyan" : "gray"}>
                                {index() === selectedIndex() ? "❯ " : "  "}
                            </text>
                            <box width={20}>
                                <text>{field.label}:</text>
                            </box>

                            <Show when={isEditing() && index() === selectedIndex()} fallback={
                                <box>
                                    {config()[field.key] ? (
                                        <text fg="white">{config()[field.key]}</text>
                                    ) : (
                                        <text fg="gray">empty</text>
                                    )}
                                </box>
                            }>
                                <box>
                                    <text fg="cyan">▌{config()[field.key] || ""}</text>
                                </box>
                            </Show>
                        </box>
                    )}
                </For>
            </box>
            <box marginTop={1}>
                <text fg="green">{status()}</text>
            </box>
            <text fg="gray">↑↓ Select │ Enter Edit/Save</text>
        </box>
    );
}
