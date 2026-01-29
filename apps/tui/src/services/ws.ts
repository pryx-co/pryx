
import { Effect, Context, Layer, Stream, PubSub, Ref, Console, Schedule } from "effect";
import WebSocket from "ws";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { homedir } from "node:os";

// Define errors
export class ConnectionError {
    readonly _tag = "ConnectionError";
    constructor(readonly message: string, readonly originalError?: unknown) { }
}

// Define connection states
export type ConnectionStatus =
    | { readonly _tag: "Disconnected" }
    | { readonly _tag: "Connecting" }
    | { readonly _tag: "Connected" }
    | { readonly _tag: "Error"; readonly error: ConnectionError };

export type RuntimeEvent = {
    event?: string;
    type?: string;
    session_id?: string;
    payload?: any;
};

// Define the service interface
export interface WebSocketService {
    readonly status: Stream.Stream<ConnectionStatus>;
    readonly messages: Stream.Stream<RuntimeEvent>;
    readonly connect: Effect.Effect<void, ConnectionError>;
    readonly send: (msg: any) => Effect.Effect<void, ConnectionError>;
    readonly disconnect: Effect.Effect<void>;
}

export const WebSocketService = Context.GenericTag<WebSocketService>("@pryx/tui/WebSocketService");

// Implementation
const make = Effect.gen(function* (_) {
    // State management
    const statusHub = yield* PubSub.unbounded<ConnectionStatus>();
    const messageHub = yield* PubSub.unbounded<RuntimeEvent>();
    const socketRef = yield* Ref.make<WebSocket | null>(null);

    // Initial status
    yield* PubSub.publish(statusHub, { _tag: "Disconnected" });

    const getRuntimeURL = () => {
        if (process.env.PRYX_WS_URL) return process.env.PRYX_WS_URL;
        try {
            const port = readFileSync(join(homedir(), ".pryx", "runtime.port"), "utf-8").trim();
            return `ws://localhost:${port}/ws`;
        } catch {
            return "ws://localhost:3000/ws";
        }
    };

    const connect = Effect.gen(function* (_) {
        yield* PubSub.publish(statusHub, { _tag: "Connecting" });
        const url = getRuntimeURL();

        yield* Effect.async<void, ConnectionError>((resume) => {
            let ws: WebSocket;
            try {
                ws = new WebSocket(url);
            } catch (e) {
                const err = new ConnectionError("Failed to create WebSocket", e);
                Effect.runSync(PubSub.publish(statusHub, { _tag: "Error", error: err }));
                resume(Effect.fail(err));
                return;
            }

            ws.onopen = () => {
                Effect.runSync(Ref.set(socketRef, ws));
                Effect.runSync(PubSub.publish(statusHub, { _tag: "Connected" }));
                resume(Effect.void);
            };

            ws.onmessage = (event) => {
                try {
                    const raw = event.data.toString();
                    const parsed = JSON.parse(raw);
                    Effect.runSync(PubSub.publish(messageHub, parsed));
                } catch (e) {
                    Effect.runSync(Console.error("Failed to parse message", e));
                }
            };

            ws.onerror = (err) => {
                const error = new ConnectionError(err.message, err);
                Effect.runSync(PubSub.publish(statusHub, { _tag: "Error", error }));
            };

            ws.onclose = () => {
                Effect.runSync(Ref.set(socketRef, null));
                Effect.runSync(PubSub.publish(statusHub, { _tag: "Disconnected" }));
            };
        });
    });

    const send = (msg: any) => Effect.gen(function* (_) {
        const ws = yield* Ref.get(socketRef);
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            return yield* Effect.fail(new ConnectionError("Not connected"));
        }
        try {
            ws.send(JSON.stringify(msg));
        } catch (e) {
            return yield* Effect.fail(new ConnectionError("Send failed", e));
        }
    });

    const disconnect = Effect.gen(function* (_) {
        const ws = yield* Ref.get(socketRef);
        if (ws) {
            ws.close();
            yield* Ref.set(socketRef, null);
            yield* PubSub.publish(statusHub, { _tag: "Disconnected" });
        }
    });

    return {
        status: Stream.fromPubSub(statusHub),
        messages: Stream.fromPubSub(messageHub),
        connect,
        send,
        disconnect
    } as WebSocketService;
});

export const WebSocketServiceLive = Layer.effect(WebSocketService, make);
