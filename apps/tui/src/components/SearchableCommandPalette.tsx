import { createSignal, createMemo, createEffect } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";

export interface Command {
    id: string;
    name: string;
    description: string;
    shortcut?: string;
    category: string;
    action: () => void;
    keywords?: string[];
}

interface SearchableCommandPaletteProps {
    commands: Command[];
    onClose: () => void;
    placeholder?: string;
}

export default function SearchableCommandPalette(props: SearchableCommandPaletteProps) {
    const [searchQuery, setSearchQuery] = createSignal("");
    const [selectedIndex, setSelectedIndex] = createSignal(0);
    const [inputMode, setInputMode] = createSignal<"keyboard" | "mouse">("keyboard");

    const filterCommands = (query: string) => {
        if (!query.trim()) {
            return props.commands;
        }
        
        const lowerQuery = query.toLowerCase();
        return props.commands.filter(cmd => {
            const nameMatch = cmd.name.toLowerCase().includes(lowerQuery);
            const descMatch = cmd.description.toLowerCase().includes(lowerQuery);
            const categoryMatch = cmd.category.toLowerCase().includes(lowerQuery);
            const keywordMatch = cmd.keywords?.some(k => k.toLowerCase().includes(lowerQuery));
            return nameMatch || descMatch || categoryMatch || keywordMatch;
        });
    };

    const filteredCommands = createMemo(() => filterCommands(searchQuery()));

    const groupedByCategory = createMemo(() => {
        const groups: Record<string, Command[]> = {};
        filteredCommands().forEach(cmd => {
            if (!groups[cmd.category]) {
                groups[cmd.category] = [];
            }
            groups[cmd.category].push(cmd);
        });
        return groups;
    });

    const commandsWithIndices = createMemo(() => {
        let globalIdx = 0;
        const result: Array<{cmd: Command; globalIdx: number; category: string}> = [];
        Object.entries(groupedByCategory()).forEach(([category, commands]) => {
            commands.forEach(cmd => {
                result.push({cmd, globalIdx: globalIdx++, category});
            });
        });
        return result;
    });

    const getCategoryColor = (category: string) => {
        const colors: Record<string, string> = {
            "Navigation": palette.accent,
            "Chat": palette.success,
            "Skills": palette.accentSoft,
            "MCP": palette.info,
            "Settings": palette.dim,
            "System": palette.error,
            "Help": palette.dim
        };
        return colors[category] || palette.text;
    };

    const executeCommand = (cmd: Command) => {
        cmd.action();
        props.onClose();
    };

    useKeyboard((evt) => {
        setInputMode("keyboard");
        
        const preventDefaultKeys = ["up", "down", "return", "enter", "escape", "tab", "backspace", "delete", "space"];
        if (preventDefaultKeys.includes(evt.name)) {
            evt.preventDefault?.();
        }
        
        switch (evt.name) {
            case "up":
            case "arrowup":
                setSelectedIndex(i => Math.max(0, i - 1));
                return;

            case "down":
            case "arrowdown":
                setSelectedIndex(i => Math.min(filteredCommands().length - 1, i + 1));
                return;

            case "return":
            case "enter": {
                const commands = filteredCommands();
                if (commands.length > 0) {
                    const idx = selectedIndex();
                    if (idx >= 0 && idx < commands.length) {
                        executeCommand(commands[idx]);
                    }
                }
                return;
            }

            case "escape":
                props.onClose();
                return;

            case "backspace":
            case "delete":
                setSearchQuery(q => q.slice(0, -1));
                return;

            case "space":
                setSearchQuery(q => q + " ");
                return;

            case "tab":
                return;
        }

        if (evt.name.length === 1) {
            setSearchQuery(q => q + evt.name);
        }
    });

    const handleMouseMove = () => {
        setInputMode("mouse");
    };

    const handleMouseOver = (globalIdx: number) => {
        if (inputMode() !== "mouse") return;
        setSelectedIndex(globalIdx);
    };

    const handleMouseUp = (cmd: Command) => {
        executeCommand(cmd);
    };

    const handleMouseDown = (globalIdx: number) => {
        setSelectedIndex(globalIdx);
    };

    createEffect(() => {
        const idx = selectedIndex();
        const total = filteredCommands().length;
        if (idx >= total && total > 0) {
            setSelectedIndex(0);
        }
    });

    const totalCommands = createMemo(() => props.commands.length);

    return (
        <box
            position="absolute"
            top={3}
            left="10%"
            width="80%"
            height="80%"
            borderStyle="double"
            borderColor={palette.border}
            backgroundColor={palette.bgPrimary}
            flexDirection="column"
            padding={1}
        >
            <box flexDirection="row" marginBottom={1} gap={1}>
                <text fg={palette.accent}>/</text>
                <box 
                    flexGrow={1} 
                    borderStyle="single" 
                    borderColor={searchQuery() ? palette.accent : palette.dim} 
                    padding={0.5}
                >
                    {searchQuery() ? (
                        <text fg={palette.text}>{searchQuery()}</text>
                    ) : (
                        <text fg={palette.dim}>{props.placeholder || "Type to search..."}</text>
                    )}
                </box>
                <box flexGrow={1} />
                <text fg={palette.dim}>{filteredCommands().length} / {totalCommands()}</text>
            </box>

            <box flexDirection="column" flexGrow={1} overflow="scroll">
                {commandsWithIndices().map(({cmd, globalIdx, category}, idx) => {
                    const isFirstInCategory = idx === 0 || 
                        commandsWithIndices()[idx - 1]?.category !== category;
                    const isSelected = globalIdx === selectedIndex();
                    
                    return (
                        <box flexDirection="column">
                            {isFirstInCategory && (
                                <box marginTop={idx > 0 ? 1 : 0} marginBottom={0.5} paddingLeft={0.5}>
                                    <text fg={getCategoryColor(category)}>{category}</text>
                                </box>
                            )}
                            
                            <box 
                                flexDirection="row" 
                                padding={0.5}
                                backgroundColor={isSelected ? palette.bgSelected : undefined}
                                onMouseMove={handleMouseMove}
                                onMouseOver={() => handleMouseOver(globalIdx)}
                                onMouseUp={() => handleMouseUp(cmd)}
                                onMouseDown={() => handleMouseDown(globalIdx)}
                            >
                                <box width={25}>
                                    <text fg={isSelected ? palette.accent : palette.accentSoft}>
                                        {cmd.name}
                                    </text>
                                </box>
                                <box flexGrow={1}>
                                    <text fg={isSelected ? palette.text : palette.dim}>
                                        {cmd.description}
                                    </text>
                                </box>
                                {cmd.shortcut && (
                                    <box width={10}>
                                        <text fg={isSelected ? palette.accentSoft : palette.dim}>
                                            {cmd.shortcut}
                                        </text>
                                    </box>
                                )}
                            </box>
                        </box>
                    );
                })}

                {filteredCommands().length === 0 && (
                    <box flexDirection="column" alignItems="center" marginTop={2}>
                        <text fg={palette.dim}>No commands found matching "{searchQuery()}"</text>
                    </box>
                )}
            </box>

            <box flexDirection="row" marginTop={1} gap={2}>
                <text fg={palette.dim}>↑↓ Navigate | Enter Select | Esc Close | Type to filter</text>
            </box>
        </box>
    );
}

