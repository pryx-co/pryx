# Effect-TS Migration PRD

## Problem Statement

Current TUI TypeScript code has several critical issues:

1. **Race Conditions**: WebSocket reconnection, concurrent state updates
2. **Memory Leaks**: Event listeners not properly cleaned up
3. **Error Handling**: Try/catch blocks lose type safety
4. **Resource Management**: No guarantees of cleanup on interruption
5. **Testability**: Side effects make testing difficult

## Goals

Migrate TUI codebase to Effect-TS for:

- **Type-safe error handling** with `Effect<Success, Error, Requirements>`
- **Resource safety** with `Scope` and automatic cleanup
- **Race condition prevention** with `Ref`, `Deferred`, and `Queue`
- **Interruption safety** for graceful shutdown
- **Better testability** with dependency injection via `Layer`

## Scope

### In Scope
1. **WebSocket Service** (`apps/tui/src/services/ws.ts`)
   - Convert to Effect Resource
   - Use Queue for message handling
   - Ref for connection state
   - Proper cleanup with Scope

2. **State Management** (SolidJS signals)
   - Integrate Effect with SolidJS reactivity
   - Use Ref for shared state
   - Prevent race conditions in updates

3. **Component Lifecycle**
   - Effect-based onMount/onCleanup
   - Scope management per component
   - Automatic resource cleanup

### Out of Scope
- Runtime (Go) - stays as is
- Build system changes
- OpenTUI library changes

## Technical Design

### 1. WebSocket as Managed Resource

```typescript
import { Effect, Queue, Ref, Scope, Stream } from "effect"

interface RuntimeEvent {
  event?: string
  payload?: any
}

class WebSocketService extends Effect.Tag("WebSocketService")<
  WebSocketService,
  {
    readonly connect: Effect.Effect<void, WebSocketError>
    readonly messages: Stream.Stream<RuntimeEvent>
    readonly send: (msg: any) => Effect.Effect<void, WebSocketError>
    readonly disconnect: Effect.Effect<void>
  }
>() {}

const makeWebSocketService = Effect.gen(function* () {
  const messageQueue = yield* Queue.unbounded<RuntimeEvent>()
  const socket = yield* Ref.make<WebSocket | null>(null)
  
  // Resource with cleanup
  yield* Effect.addFinalizer(() => 
    Effect.gen(function* () {
      const ws = yield* Ref.get(socket)
      if (ws) ws.close()
    })
  )
  
  return {
    connect: /* ... */,
    messages: Stream.fromQueue(messageQueue),
    send: /* ... */,
    disconnect: /* ... */
  }
})

export const WebSocketServiceLive = Layer.scoped(
  WebSocketService,
  makeWebSocketService
)
```

### 2. SolidJS Integration

```typescript
import { createSignal, onCleanup } from "solid-js"
import { Effect, Runtime } from "effect"

function useEffectSignal<A>(
  effect: Effect.Effect<A>
): () => A | undefined {
  const runtime = Runtime.defaultRuntime
  const [value, setValue] = createSignal<A>()
  
  const fiber = Effect.runFork(
    effect.pipe(Effect.tap(setValue))
  )(runtime)
  
  onCleanup(() => {
    Effect.runFork(Fiber.interrupt(fiber))(runtime)
  })
  
  return value
}
```

### 3. Component Pattern

```typescript
export default function MyComponent() {
  const ws = useEffectService(WebSocketService)
  
  const messages = useEffectStream(
    ws.messages.pipe(
      Stream.take(100), // Backpressure
      Stream.debounce("100 millis")
    )
  )
  
  return <div>{messages().map(/* ... */)}</div>
}
```

## Migration Strategy

### Phase 1: Foundation
- Install Effect-TS
- Create adapter utilities for SolidJS integration
- Setup Layer-based dependency injection

### Phase 2: WebSocket Service
- Migrate `ws.ts` to Effect Resource
- Use Queue for messages
- Ref for connection state
- Test reconnection logic

### Phase 3: Component Migration
- Migrate one component at a time
- Start with `MinimalApp.tsx` (already simple)
- Then `Chat.tsx`, `Skills.tsx`, etc.

### Phase 4: Cleanup & Verification
- Remove old Promise-based code
- Verify no memory leaks with profiling
- Load testing for race conditions

## Success Criteria

1. ✅ No memory leaks after 1000 connect/disconnect cycles
2. ✅ No race conditions in concurrent message handling
3. ✅ Graceful shutdown with Ctrl+C (all resources cleaned)
4. ✅ 100% type safety (no `any` or `try/catch`)
5. ✅ Test coverage >80% for Effect code

## Dependencies

- `effect@latest` (~3.10.x)
- Integration with existing Bun, SolidJS, OpenTUI
- No breaking changes to users

## Timeline

- Phase 1: 1 day
- Phase 2: 2 days  
- Phase 3: 3 days
- Phase 4: 1 day

Total: ~1 week

## Risks

1. **Learning curve**: Team needs Effect-TS knowledge
   - *Mitigation*: Pair programming, documentation
   
2. **SolidJS compatibility**: Effect primitives vs signals
   - *Mitigation*: Create adapter hooks early

3. **Bundle size**: Effect adds ~50kb
   - *Mitigation*: Acceptable for TUI, tree-shaking helps
