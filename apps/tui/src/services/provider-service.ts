import { Effect, Context, Layer } from "effect";

export interface Provider {
  id: string;
  name: string;
  requires_api_key: boolean;
}

export interface Model {
  id: string;
  name: string;
}

export interface ProvidersResponse {
  providers: Provider[];
}

export interface ModelsResponse {
  models: Model[];
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
}

export const ProviderService = Context.GenericTag<ProviderService>("@pryx/tui/ProviderService");

const getApiUrl = (): string => {
  return process.env.PRYX_API_URL || "http://localhost:3000";
};

const makeProviderService = Effect.gen(function* () {
  const fetchProviders = Effect.gen(function* () {
    const result = yield* Effect.tryPromise({
      try: async () => {
        const apiUrl = getApiUrl();
        const res = await fetch(`${apiUrl}/api/v1/providers`);
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
          const apiUrl = getApiUrl();
          const res = await fetch(`${apiUrl}/api/v1/providers/${providerId}/models`);
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

  return {
    fetchProviders,
    fetchModels,
  } as ProviderService;
});

export const ProviderServiceLive = Layer.effect(ProviderService, makeProviderService);
