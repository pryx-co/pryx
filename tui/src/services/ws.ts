import WebSocket from "ws";

let socket: WebSocket | null = null;
const URL = "ws://localhost:8081/ws"; // Pryx-core WS URL

export function connect(onStatus: (status: string) => void) {
    try {
        socket = new WebSocket(URL);

        socket.onopen = () => {
            onStatus("Connected");
        };

        socket.onclose = () => {
            onStatus("Disconnected");
        };

        socket.onerror = (err: any) => {
            onStatus("Error");
        };
    } catch (e) {
        onStatus("Failed");
    }
}

export function send(msg: any) {
    if (socket?.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(msg));
    }
}
