import { createSignal, For, onMount, onCleanup } from "solid-js";

interface Command {
    id: string;
    label: string;
    shortcut: string;
    action: () => void;
}

interface CommandPaletteProps {
    commands: Command[];
    onClose: () => void;
}

export default function CommandPalette(props: CommandPaletteProps) {
    const [selectedIndex, setSelectedIndex] = createSignal(0);

    onMount(() => {
        const handleKey = (data: Buffer) => {
            const seq = data.toString();
            
            // Arrow up: ESC[A or ESCOA
            if (seq === '\u001b[A' || seq === '\u001bOA') {
                setSelectedIndex(i => Math.max(0, i - 1));
            }
            // Arrow down: ESC[B or ESCOB
            else if (seq === '\u001b[B' || seq === '\u001bOB') {
                setSelectedIndex(i => Math.min(props.commands.length - 1, i + 1));
            }
            // Enter
            else if (seq === '\r' || seq === '\n') {
                const cmd = props.commands[selectedIndex()];
                if (cmd) {
                    cmd.action();
                }
            }
            // Escape
            else if (seq === '\u001b') {
                props.onClose();
            }
            // Number keys 1-9 for direct selection
            else if (seq >= '1' && seq <= '9') {
                const idx = parseInt(seq) - 1;
                if (idx < props.commands.length) {
                    setSelectedIndex(idx);
                }
            }
        };

        if (typeof process !== "undefined" && process.stdin.isTTY) {
            process.stdin.on("data", handleKey);
        }

        onCleanup(() => {
            if (typeof process !== "undefined" && process.stdin) {
                process.stdin.off("data", handleKey);
            }
        });
    });

    return (
        <box
            position="absolute"
            top={3}
            left="20%"
            width="60%"
            borderStyle="single"
            borderColor="cyan"
            backgroundColor="#1a1a1a"
            flexDirection="column"
            padding={1}
        >
            <box marginBottom={1}>
                <text fg="cyan">Commands</text>
            </box>
            
            <For each={props.commands}>
                {(cmd, index) => (
                    <box 
                        flexDirection="row" 
                        padding={1}
                        backgroundColor={index() === selectedIndex() ? "cyan" : undefined}
                    >
                        <text 
                            fg={index() === selectedIndex() ? "black" : "white"}
                        >
                            {index() + 1}. {cmd.label}
                        </text>
                        <box flexGrow={1} />
                        <text fg="gray">{cmd.shortcut}</text>
                    </box>
                )}
            </For>
            
            <box marginTop={1} flexDirection="column" gap={0}>
                <text fg="gray">↑↓ Navigate | Enter Select | Esc Close | 1-9 Quick Select</text>
            </box>
        </box>
    );
}
