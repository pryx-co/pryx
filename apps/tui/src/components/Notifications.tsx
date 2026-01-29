// @ts-nocheck
import { For } from "solid-js";

export interface Notification {
    id: string;
    message: string;
    type: "info" | "success" | "warning" | "error";
    timestamp: number;
}

interface NotificationsProps {
    items: Notification[];
}

export default function Notifications(props: NotificationsProps) {
    const getColor = (type: Notification["type"]) => {
        switch (type) {
            case "success": return "green";
            case "error": return "red";
            case "warning": return "yellow";
            default: return "blue";
        }
    };

    return (
        <box
            flexDirection="column"
            position="absolute"
            top={1}
            right={1}
            width={40}
        >
            <For each={props.items}>
                {(item) => (
                    <box
                        borderStyle="single"
                        borderColor={getColor(item.type)}
                        padding={1}
                        marginBottom={1}
                        flexDirection="column"
                    >
                        <text color={getColor(item.type)} bold>{item.type.toUpperCase()}</text>
                        <text>{item.message}</text>
                    </box>
                )}
            </For>
        </box>
    );
}
