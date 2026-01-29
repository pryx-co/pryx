// @ts-nocheck
import { Box, Text, Input } from "@opentui/core";
import { createSignal, onMount, onCleanup, Show, For } from "solid-js";
import { loadConfig, saveConfig } from "../services/config";

interface ChannelField {
    id: string;
    type: "header" | "toggle" | "input";
    key?: string;
    label: string;
    placeholder?: string;
}

export default function Channels() {
    const [config, setConfig] = createSignal<any>({});
    const [fields] = createSignal<ChannelField[]>([
        { id: "h1", type: "header", label: "TELEGRAM BOT" },
        { id: "tg_status", type: "toggle", key: "telegram_enabled", label: "Status" },
        { id: "tg_token", type: "input", key: "telegram_token", label: "Bot Token", placeholder: "123456:ABC-..." },
        { id: "sep1", type: "header", label: " " }, // Spacer
        { id: "h2", type: "header", label: "GENERIC WEBHOOK" },
        { id: "wh_status", type: "toggle", key: "webhook_enabled", label: "Status" }, // Future
    ]);
    const [selectedIndex, setSelectedIndex] = createSignal(1); // Start at status
    const [isEditing, setIsEditing] = createSignal(false);
    const [status, setStatus] = createSignal("");

    // Helper to skip headers
    const moveSelection = (dir: 1 | -1) => {
        let next = selectedIndex();
        const len = fields().length;
        // Max attempts to find non-header
        for (let i = 0; i < len; i++) {
            next = (next + dir + len) % len;
            if (fields()[next].type !== "header") break;
        }
        setSelectedIndex(next);
    };

    const handleInput = (data: Buffer) => {
        if (isEditing()) return;

        const key = data.toString();
        if (key === '\u001B\u005B\u0041') { // Up
            moveSelection(-1);
        } else if (key === '\u001B\u005B\u0042') { // Down
            moveSelection(1);
        } else if (key === '\r' || key === '\n') { // Enter
            const field = fields()[selectedIndex()];
            if (field.type === "input") {
                setIsEditing(true);
            } else if (field.type === "toggle") {
                // Toggle immediately
                toggleValue(field.key!);
            }
        } else if (key === ' ') { // Space
            const field = fields()[selectedIndex()];
            if (field.type === "toggle") {
                toggleValue(field.key!);
            }
        }
    };

    const toggleValue = async (key: string) => {
        const val = !config()[key];
        await handleSave(key, val);
    };

    const handleSave = async (key: string, value: any) => {
        const newConfig = { ...config(), [key]: value };
        setConfig(newConfig);
        setIsEditing(false);
        saveConfig(newConfig);
        setStatus("Saved!");
        setTimeout(() => setStatus(""), 2000);
    };

    onMount(async () => {
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

    const renderValue = (field: ChannelField) => {
        const val = config()[field.key!];
        if (field.type === "toggle") {
            return val ? <Text color="green">ENABLED</Text> : <Text color="gray">DISABLED</Text>;
        }
        if (!val) return <Text color="gray" italic>empty</Text>;
        // Mask token partially
        if (field.key?.includes("token")) {
            return <Text>{val.substring(0, 4)}...{val.substring(val.length - 4)}</Text>;
        }
        return <Text>{val}</Text>;
    };

    return (
        <Box flexDirection="column" flexGrow={1}>
            <Text bold color="magenta">Channel Setup</Text>
            <Text color="gray">Config Path: ~/.pryx/config.yaml</Text>

            <Box marginTop={1} flexDirection="column" borderStyle="round" padding={1}>
                <For each={fields()}>
                    {(field, index) => {
                        if (field.type === "header") {
                            return <Box marginTop={field.label === " " ? 0 : 1} marginBottom={0}>
                                <Text bold color="cyan">{field.label}</Text>
                            </Box>;
                        }

                        const isSelected = index() === selectedIndex();
                        return (
                            <Box flexDirection="row" marginBottom={0}>
                                <Text color={isSelected ? "cyan" : "gray"}>
                                    {isSelected ? "❯ " : "  "}
                                </Text>
                                <Box width={15}>
                                    <Text bold={isSelected}>
                                        {field.label}:
                                    </Text>
                                </Box>

                                <Show when={isEditing() && isSelected} fallback={renderValue(field)}>
                                    <Input
                                        value={config()[field.key!] || ""}
                                        placeholder={field.placeholder}
                                        onSubmit={(val) => handleSave(field.key!, val)}
                                    />
                                </Show>
                            </Box>
                        );
                    }}
                </For>
            </Box>

            <Box marginTop={1}>
                <Text color="green">{status()}</Text>
            </Box>
            <Text color="gray">↑↓ Select │ Enter/Space Toggle │ Enter Edit</Text>
        </Box>
    );
}
