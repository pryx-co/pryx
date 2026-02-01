import { Effect, Stream, Console } from "effect";
import { WebSocketService, WebSocketServiceLive } from "./services/ws";

const program = Effect.gen(function* () {
  const ws = yield* WebSocketService;

  yield* Console.log("--- Starting WS Test ---");

  // 1. Connect
  yield* Console.log("Connecting...");
  yield* Effect.fork(ws.connect);

  // 2. Monitor Status
  yield* Effect.fork(
    ws.status.pipe(
      Stream.runForEach(status => Console.log(`STATUS CHANGE: ${JSON.stringify(status)}`))
    )
  );

  // 3. Wait for connection (or timeout)
  yield* Effect.sleep("2 seconds");

  // 4. Send a test message
  yield* Console.log("Sending test message...");
  const testMsg = { type: "PING", timestamp: Date.now() };
  yield* ws.send(testMsg);

  // 5. Wait for messages
  yield* Effect.fork(
    ws.messages.pipe(Stream.runForEach(msg => Console.log(`RECEIVED: ${JSON.stringify(msg)}`)))
  );

  yield* Effect.sleep("5 seconds");
  yield* Console.log("--- Test Complete ---");
});

// Run with Live dependencies
const runnable = program.pipe(Effect.provide(WebSocketServiceLive));

Effect.runPromise(runnable).catch(console.error);
