import { createSignal, onCleanup, onMount } from "solid-js";
import type { Accessor } from "solid-js";
import { Effect, Fiber, Runtime, Stream, Context, ManagedRuntime, Layer } from "effect";
import { WebSocketServiceLive } from "../services/ws";
import { HealthCheckServiceLive } from "../services/health-check";
import { ProviderServiceLive } from "../services/provider-service";
import { SkillsServiceLive } from "../services/skills-api";
import { appendFileSync } from "fs";

function log(msg: string) {
  appendFileSync("debug.log", `[hooks] ${msg}\n`);
  // console.error(`[hooks] ${msg}`); // Fallback
}

log("MODULE LOADED");

// Create a managed runtime that includes our Live services
export const AppRuntime = ManagedRuntime.make(
  Layer.mergeAll(WebSocketServiceLive, HealthCheckServiceLive, ProviderServiceLive, SkillsServiceLive)
);

/**
 * Run an Effect and expose result as SolidJS signal
 */
export function useEffectSignal<A, E = never>(
  effect: Effect.Effect<A, E>
): Accessor<A | undefined> {
  const [value, setValue] = createSignal<A | undefined>();
  const [error, setError] = createSignal<E | undefined>();

  onMount(() => {
    // Run with our managed runtime
    AppRuntime.runFork(
      effect.pipe(
        Effect.tap(a => Effect.sync(() => setValue(() => a))),
        Effect.tapError(e => Effect.sync(() => setError(() => e)))
      )
    );
  });

  return value;
}

/**
 * Subscribe to an Effect Stream as SolidJS signal
 */
export function useEffectStream<A, E = never>(stream: Stream.Stream<A, E>): Accessor<A[]> {
  const [items, setItems] = createSignal<A[]>([]);

  onMount(() => {
    const fiber = AppRuntime.runFork(
      stream.pipe(Stream.runForEach(item => Effect.sync(() => setItems(prev => [...prev, item]))))
    );

    onCleanup(() => {
      Effect.runFork(Fiber.interrupt(fiber));
    });
  });

  return items;
}

/**
 * Access the WebSocketService
 */
export function useEffectService<I, S>(tag: Context.Tag<I, S>): Accessor<S | undefined> {
  log("useEffectService: start");
  try {
    log("useEffectService: calling createSignal");
    const [service, setService] = createSignal<S | undefined>();
    log("useEffectService: createSignal done");

    log("useEffectService: scheduling onMount");
    onMount(() => {
      log("useEffectService: onMount running");
      // Run an effect to extract the service
      // This is safe because we use ManagedRuntime which keeps services alive
      AppRuntime.runPromise(tag as any)
        .then(svc => {
          log("useEffectService: service resolved");
          setService(() => svc as S);
        })
        .catch(err => {
          log(`useEffectService: service error ${err}`);
          console.error("Failed to get service:", err);
        });
    });
    log("useEffectService: onMount scheduled");

    return service;
  } catch (e) {
    log(`useEffectService: CRASHED ${e}`);
    throw e;
  }
}

// Global runtime for ad-hoc usage
export const TUIRuntime = Runtime.defaultRuntime;
