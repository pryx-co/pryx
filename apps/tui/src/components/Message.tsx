import { Box, Text } from "@opentui/core";

export type MessageType = "user" | "assistant" | "tool" | "approval" | "system";

export interface MessageProps {
    type: MessageType;
    content: string;
    toolName?: string;
    toolStatus?: "running" | "done" | "error";
    pending?: boolean;
}

const typeColors: Record<MessageType, string> = {
    user: "cyan",
    assistant: "white",
    tool: "yellow",
    approval: "magenta",
    system: "gray"
};

const typePrefixes: Record<MessageType, string> = {
    user: "You",
    assistant: "Pryx",
    tool: "⚙️",
    approval: "⚠️ Approval",
    system: "ℹ️"
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
            <Box>
                <Text color="yellow">{statusIcon} {props.toolName}: </Text>
                <Text color="gray">{props.content}</Text>
            </Box>
        );
    }

    return (
        <Box>
            <Text color={color} bold>{prefix}: </Text>
            <Text color={color}>{props.content}</Text>
            {props.pending && <Text color="gray"> ▌</Text>}
        </Box>
    );
}
