import { Effect, Context, Layer } from "effect";
import { getRuntimeHttpUrl } from "./skills-api";

export interface Provider {
  id: string;
  name: string;
  requires_api_key: boolean;
}

export interface Model {
  id: string;
  name: string;
  provider: string;
  context_window?: number;
  max_output_tokens?: number;
  supports_tools?: boolean;
  supports_vision?: boolean;
  supports_reasoning?: boolean;
  input_price_1m?: number;
  output_price_1m?: number;
}

export interface ProvidersResponse {
  providers: Provider[];
}

export interface ModelsResponse {
  models: Model[];
}

export interface ProviderKeyStatusResponse {
  configured: boolean;
  provider_id?: string;
}

export class ProviderFetchError {
  readonly _tag = "ProviderFetchError";
  constructor(
    readonly message: string,
    readonly cause?: unknown
  ) {}
}

export interface ProviderService {
  readonly fetchProviders: Effect.Effect<Provider[], ProviderFetchError>;
  readonly fetchModels: (providerId: string) => Effect.Effect<Model[], ProviderFetchError>;
  readonly getProviderKeyStatus: (providerId: string) => Effect.Effect<boolean, ProviderFetchError>;
  readonly setProviderKey: (
    providerId: string,
    apiKey: string
  ) => Effect.Effect<void, ProviderFetchError>;
  readonly deleteProviderKey: (providerId: string) => Effect.Effect<void, ProviderFetchError>;
}

export const ProviderService = Context.GenericTag<ProviderService>("@pryx/tui/ProviderService");

const makeProviderService = Effect.gen(function* () {
  const fetchProviders = Effect.gen(function* () {
    const result = yield* Effect.tryPromise({
      try: async () => {
        const res = await fetch(`${getRuntimeHttpUrl()}/api/v1/providers`);
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const data = (await res.json()) as ProvidersResponse;
        return data.providers || [];
      },
      catch: error => new ProviderFetchError("Failed to fetch providers", error),
    });
    return result;
  });

  const fetchModels = (providerId: string) =>
    Effect.gen(function* () {
      const result = yield* Effect.tryPromise({
        try: async () => {
          const res = await fetch(`${getRuntimeHttpUrl()}/api/v1/providers/${providerId}/models`);
          if (!res.ok) {
            throw new Error(`HTTP ${res.status}`);
          }
          const data = (await res.json()) as ModelsResponse;
          return data.models || [];
        },
        catch: error => new ProviderFetchError("Failed to fetch models", error),
      });
      return result;
    });

  const getProviderKeyStatus = (providerId: string) =>
    Effect.gen(function* () {
      const result = yield* Effect.tryPromise({
        try: async () => {
          const res = await fetch(`${getRuntimeHttpUrl()}/api/v1/providers/${providerId}/key`);
          if (!res.ok) {
            throw new Error(`HTTP ${res.status}`);
          }
          const data = (await res.json()) as ProviderKeyStatusResponse;
          return !!data.configured;
        },
        catch: error => new ProviderFetchError("Failed to fetch provider key status", error),
      });
      return result;
    });

  const setProviderKey = (providerId: string, apiKey: string) =>
    Effect.gen(function* () {
      yield* Effect.tryPromise({
        try: async () => {
          const res = await fetch(`${getRuntimeHttpUrl()}/api/v1/providers/${providerId}/key`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ api_key: apiKey }),
          });
          if (!res.ok) {
            throw new Error(`HTTP ${res.status}`);
          }
        },
        catch: error => new ProviderFetchError("Failed to store provider key", error),
      });
    });

  const deleteProviderKey = (providerId: string) =>
    Effect.gen(function* () {
      yield* Effect.tryPromise({
        try: async () => {
          const res = await fetch(`${getRuntimeHttpUrl()}/api/v1/providers/${providerId}/key`, {
            method: "DELETE",
          });
          if (!res.ok) {
            throw new Error(`HTTP ${res.status}`);
          }
        },
        catch: error => new ProviderFetchError("Failed to delete provider key", error),
      });
    });

  return {
    fetchProviders,
    fetchModels,
    getProviderKeyStatus,
    setProviderKey,
    deleteProviderKey,
  } as ProviderService;
});

export const ProviderServiceLive = Layer.effect(ProviderService, makeProviderService);
