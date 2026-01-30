import { Effect, Context, Layer, Schedule, Console } from "effect";

export interface HealthCheckResponse {
  status: "ok" | "error";
  providers?: string[];
  error?: string;
}

export class HealthCheckError {
  readonly _tag = "HealthCheckError";
  constructor(
    readonly message: string,
    readonly cause?: unknown
  ) {}
}

export interface HealthCheckService {
  readonly checkHealth: Effect.Effect<HealthCheckResponse, HealthCheckError>;
  readonly pollHealth: (intervalMs: number, callback: (result: HealthCheckResponse) => void) => Effect.Effect<void, never>;
}

export const HealthCheckService = Context.GenericTag<HealthCheckService>("@pryx/tui/HealthCheckService");

const getApiUrl = (): string => {
  return process.env.PRYX_API_URL || "http://localhost:3000";
};

const makeHealthCheckService = Effect.gen(function* () {
  const checkHealth = Effect.gen(function* () {
    const result = yield* Effect.tryPromise({
      try: async () => {
        const apiUrl = getApiUrl();
        const res = await fetch(`${apiUrl}/health`, { method: "GET" });
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        return (await res.json()) as HealthCheckResponse;
      },
      catch: error => new HealthCheckError("Failed to check health", error),
    });
    return result;
  });

  const pollHealth = (intervalMs: number, callback: (result: HealthCheckResponse) => void) =>
    Effect.gen(function* () {
      const check = Effect.gen(function* () {
        const result = yield* checkHealth;
        yield* Effect.sync(() => callback(result));
      }).pipe(
        Effect.catchAll(error => Effect.sync(() => {
          callback({ status: "error", error: "Health check failed" });
        }))
      );

      yield* check.pipe(
        Effect.repeat(Schedule.spaced(intervalMs))
      );
    });

  return {
    checkHealth,
    pollHealth,
  } as HealthCheckService;
});

export const HealthCheckServiceLive = Layer.effect(HealthCheckService, makeHealthCheckService);
