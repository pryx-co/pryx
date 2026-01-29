import { createSignal, For } from "solid-js";
import { KEYBINDINGS } from "../lib/keybindings";

interface KeyboardShortcutsProps {
    onClose: () => void;
}

export default function KeyboardShortcuts(props: KeyboardShortcutsProps) {
    const [selectedCategory, setSelectedCategory] = createSignal<string | null>(null);

    const categories = [
        { id: "application", label: "Application", color: "cyan" },
        { id: "navigation", label: "Navigation", color: "green" },
        { id: "editing", label: "Editing", color: "yellow" },
        { id: "history", label: "History", color: "magenta" },
        { id: "scroll", label: "Scroll", color: "blue" },
    ];

    const filteredBindings = () => {
        if (!selectedCategory()) return KEYBINDINGS;
        return KEYBINDINGS.filter(b => b.category === selectedCategory());
    };

    const groupedByCategory = () => {
        const groups: Record<string, typeof KEYBINDINGS> = {};
        filteredBindings().forEach(binding => {
            if (!groups[binding.category]) {
                groups[binding.category] = [];
            }
            groups[binding.category].push(binding);
        });
        return groups;
    };

    const getCategoryColor = (cat: string) => {
        const c = categories.find(c => c.id === cat);
        return c?.color || "white";
    };

    const getCategoryLabel = (cat: string) => {
        const c = categories.find(c => c.id === cat);
        return c?.label || cat;
    };

    return (
        <box
            position="absolute"
            top={2}
            left="10%"
            width="80%"
            height="90%"
            borderStyle="double"
            borderColor="cyan"
            backgroundColor="#0a0a0a"
            flexDirection="column"
            padding={1}
        >
            <box flexDirection="row" marginBottom={1}>
                <text fg="cyan">Keyboard Shortcuts</text>
                <box flexGrow={1} />
                <text fg="gray">Press Esc to close</text>
            </box>

            <box flexDirection="row" gap={1} marginBottom={1}>
                <box
                    padding={1}
                    backgroundColor={!selectedCategory() ? "cyan" : undefined}
                >
                    <text fg={!selectedCategory() ? "black" : "white"}>All</text>
                </box>
                <For each={categories}>
                    {(cat) => (
                        <box
                            padding={1}
                            backgroundColor={selectedCategory() === cat.id ? cat.color : undefined}
                        >
                            <text fg={selectedCategory() === cat.id ? "black" : cat.color}>
                                {cat.label}
                            </text>
                        </box>
                    )}
                </For>
            </box>

            <box flexDirection="column" flexGrow={1}>
                <For each={Object.entries(groupedByCategory())}>
                    {([category, bindings]) => (
                        <box flexDirection="column" marginBottom={1}>
                            <box marginBottom={0.5}>
                                <text fg={getCategoryColor(category)}>
                                    {getCategoryLabel(category)}
                                </text>
                            </box>
                            
                            <For each={bindings}>
                                {(binding) => (
                                    <box flexDirection="row" padding={0.5}>
                                        <box width={20}>
                                            <text fg="yellow">{binding.key}</text>
                                        </box>
                                        <box flexGrow={1}>
                                            <text fg="white">{binding.description}</text>
                                        </box>
                                    </box>
                                )}
                            </For>
                        </box>
                    )}
                </For>
            </box>

            <box flexDirection="row" marginTop={1} gap={2}>
                <text fg="gray">Total: {KEYBINDINGS.length} shortcuts</text>
                <box flexGrow={1} />
                <text fg="gray">? anytime for help</text>
            </box>
        </box>
    );
}
