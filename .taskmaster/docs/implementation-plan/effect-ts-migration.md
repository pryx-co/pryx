# Effect-TS Migration Plan

## Overview

Migrate all TypeScript code to use Effect-TS for:
- **Type Safety**: Strongly typed error handling
- **Race Condition Prevention**: Structured concurrency via Fibers
- **Memory Leak Prevention**: Proper resource management with Scope
- **Composability**: Effect composition for complex async flows

## Background and Motivation

Current issues:
- Multiple WebSocket connections from different components
- Manual cleanup that may be missed
- No structured error handling
- Race conditions possible in async operations
- Memory leaks from uncleared listeners

Effect-TS provides solutions through:
- `Effect<A, E, R>` type for all operations with explicit errors
- `Scope` for automatic resource cleanup
- `Fiber` for structured concurrency
- `Layer` for dependency injection
- `Stream` for reactive data flows

## Key Challenges

1. **Integration with SolidJS/React**: Need hooks to bridge Effect with UI frameworks
2. **WebSocket Lifecycle**: Managing connection lifecycle with proper cleanup
3. **File I/O**: Making config operations safe with Effect
4. **Migration Strategy**: Gradual migration without breaking existing code
5. **Testing**: Ensuring no race conditions or memory leaks post-migration

## High-level Task Breakdown

### Phase 1: Setup and Infrastructure
- [ ] Install effect package in apps/tui
- [ ] Install effect package in apps/web
- [ ] Create Effect configuration (tsconfig, eslint)
- [ ] Create base Effect utilities and helpers

### Phase 2: Core Services Migration (TUI)
- [ ] Migrate ws.ts to Effect-based WebSocket service
- [ ] Migrate config.ts to Effect-based file operations
- [ ] Create Effect Layers for DI
- [ ] Add WebSocket resource management with Scope

### Phase 3: TUI Components Migration
- [ ] Create useEffectRunner hook for SolidJS
- [ ] Migrate App.tsx to use Effect
- [ ] Migrate Chat.tsx WebSocket handling
- [ ] Migrate SessionExplorer.tsx
- [ ] Migrate Settings.tsx
- [ ] Migrate Channels.tsx
- [ ] Migrate Skills.tsx
- [ ] Migrate Notifications.tsx

### Phase 4: Web App Migration
- [ ] Install and configure effect in apps/web
- [ ] Create React hooks for Effect
- [ ] Migrate Dashboard components
- [ ] Migrate SkillCard component
- [ ] Migrate DeviceCard/DeviceList components

### Phase 5: Testing and Validation
- [ ] Test WebSocket connection handling
- [ ] Memory profiling to verify no leaks
- [ ] Race condition testing
- [ ] Error handling verification
- [ ] Performance benchmarks

## Implementation Details

### WebSocket Effect Service

```typescript
// Conceptual design
import { Effect, Stream, Layer, Context, Scope } from "effect"

interface WebSocketClient {
  connect: Effect<void, ConnectionError, Scope>
  send: (msg: Message) => Effect<void, SendError>
  messages: Stream<Message, WebSocketError>
}

// Automatic cleanup via Scope
const program = Effect.gen(function* () {
  const ws = yield* WebSocketClient
  yield* ws.connect // Auto-cleanup when scope closes
  const messages = yield* ws.messages
  // ...
}).pipe(Effect.scoped) // Ensures cleanup
```

### Config Service with Effect

```typescript
// Safe file operations
const loadConfig = Effect.tryPromise({
  try: () => fs.readFile(CONFIG_PATH, "utf-8"),
  catch: (e) => new ConfigError(String(e))
}).pipe(
  Effect.map((content) => yaml.load(content)),
  Effect.orElseSucceed(() => ({})) // Return empty on error
)
```

### SolidJS Integration

```typescript
// Custom hook for running effects
function useEffect<A, E>(effect: Effect<A, E>) {
  const [state, setState] = createSignal<A>()
  const [error, setError] = createSignal<E>()
  
  onMount(() => {
    const fiber = Effect.runFork(effect)
    onCleanup(() => Fiber.interrupt(fiber))
  })
  
  return { state, error }
}
```

## Technical Considerations

1. **Bundle Size**: Effect is ~30KB gzipped, acceptable for our use case
2. **Learning Curve**: Team needs to understand Effect patterns
3. **Interop**: Need clean boundaries between Effect and non-Effect code
4. **Testing**: Effect provides excellent testability with TestClock, TestContext

## Dependencies

```json
{
  "effect": "^3.0.0",
  "@effect/platform": "^0.50.0",
  "@effect/platform-node": "^0.47.0"
}
```

## Success Criteria

- [ ] All async operations use Effect
- [ ] No memory leaks detected in profiling
- [ ] No race conditions in WebSocket handling
- [ ] All errors explicitly typed and handled
- [ ] Tests pass with proper cleanup
- [ ] No regression in performance

## Related Resources

- [Effect-TS Documentation](https://effect.website/)
- [Effect GitHub](https://github.com/Effect-TS/effect)
- [Effect Discord Community](https://discord.gg/effect-ts)
