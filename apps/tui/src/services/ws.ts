import WebSocket from "ws";

let socket: WebSocket | null = null;
const URL = process.env.PRYX_WS_URL ?? "ws://localhost:3000/ws";

export type RuntimeEvent = {
    event?: string;
    type?: string;
    session_id?: string;
    payload?: any;
};

export function connect(
    onStatus: (status: string) => void,
    onEvent: (evt: RuntimeEvent) => void
) {
    try {
        socket = new WebSocket(URL);

        socket.onopen = () => {
            onStatus("Connected");
        };

        socket.onmessage = (msg: WebSocket.MessageEvent) => {
            const raw = typeof msg.data === "string" ? msg.data : msg.data.toString();
            try {
                const evt = JSON.parse(raw);
                onEvent(evt);
            } catch {
                // ignore
            }
        };

        socket.onclose = () => {
            onStatus("Disconnected");
        };

        socket.onerror = (err: any) => {
            console.error("WebSocket error:", err);
            onStatus("Error");
        };
    } catch (e) {
        console.error("WebSocket connection failed:", e);
        onStatus("Failed");
    }
}

export function send(msg: any) {
    if (socket?.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(msg));
    }
}

export function sendChat(sessionId: string, message: string) {
    send({
        event: "chat.send",
        session_id: sessionId,
        payload: { content: message }
    });
}

export function sendApproval(sessionId: string, approvalId: string, approved: boolean) {
    send({
        event: "approval.resolve",
        session_id: sessionId,
        payload: { approval_id: approvalId, approved }
    });
}
