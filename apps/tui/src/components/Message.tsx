export type MessageType = "user" | "assistant" | "tool" | "approval" | "system" | "thinking";

export interface MessageProps {
    type: MessageType;
    content: string;
    toolName?: string;
    toolStatus?: "running" | "done" | "error";
    pending?: boolean;
    isLast?: boolean;
}

const typeColors: Record<MessageType, string> = {
    user: "cyan",
    assistant: "white",
    tool: "yellow",
    approval: "magenta",
    system: "gray",
    thinking: "gray"
};

const typePrefixes: Record<MessageType, string> = {
    user: "You",
    assistant: "Pryx",
    tool: "⚙️",
    approval: "⚠️ Approval",
    system: "ℹ️",
    thinking: "💭"
};

export default function Message(props: MessageProps) {
    const color = typeColors[props.type];
    const prefix = typePrefixes[props.type];

    if (props.type === "tool" && props.toolName) {
        const statusIcon = props.toolStatus === "running" ? "⏳"
            : props.toolStatus === "done" ? "✅"
                : props.toolStatus === "error" ? "❌"
                    : "⚙️";
        return (
            <box>
                <text fg="yellow">{statusIcon} {props.toolName}: </text>
                <text fg="gray">{props.content}</text>
            </box>
        );
    }

    if (props.type === "thinking") {
        return (
            <box borderStyle="single" borderColor="gray" padding={1}>
                <text fg="gray">_Thinking:_ {props.content}</text>
            </box>
        );
    }

    return (
        <box>
            <text fg={color}>{prefix}: </text>
            <text fg={color}>{props.content}</text>
            {props.pending && <text fg="gray"> ▌</text>}
        </box>
    );
}
