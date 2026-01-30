import { createSignal, createEffect, onMount } from "solid-js";
import { Effect } from "effect";
import { useEffectService, AppRuntime } from "../lib/hooks";
import { ProviderService, Provider as ProviderType, Model as ModelType } from "../services/provider-service";
import { saveConfig } from "../services/config";

interface SetupRequiredProps {
  onSetupComplete: () => void;
}

const API_BASE = "http://localhost:3000";

export default function SetupRequired(props: SetupRequiredProps) {
  const providerService = useEffectService(ProviderService);
  const [step, setStep] = createSignal(1);
  const [provider, setProvider] = createSignal("");
  const [apiKey, setApiKey] = createSignal("");
  const [modelName, setModelName] = createSignal("");
  const [error, setError] = createSignal("");
  const [providers, setProviders] = createSignal<ProviderType[]>([]);
  const [models, setModels] = createSignal<ModelType[]>([]);
  const [loading, setLoading] = createSignal(false);
  const [fetchError, setFetchError] = createSignal("");

  onMount(() => {
    const service = providerService();
    if (!service) return;

    AppRuntime.runFork(
      service.fetchProviders.pipe(
        Effect.tap(providers => Effect.sync(() => setProviders(providers))),
        Effect.catchAll(err => Effect.sync(() => {
          setFetchError(err.message || "Failed to connect to runtime");
          setProviders([
            { id: "openai", name: "OpenAI", requires_api_key: true },
            { id: "anthropic", name: "Anthropic", requires_api_key: true },
            { id: "google", name: "Google AI", requires_api_key: true },
            { id: "ollama", name: "Ollama (Local)", requires_api_key: false },
          ]);
        }))
      )
    );
  });

  const handleProviderSelect = (providerId: string) => {
    const service = providerService();
    if (!service) return;

    setProvider(providerId);
    const selectedProvider = providers().find(p => p.id === providerId);

    AppRuntime.runFork(
      service.fetchModels(providerId).pipe(
        Effect.tap(availableModels => {
          const defaultModel = availableModels.length > 0 ? availableModels[0].id : "";
          setModels(availableModels);
          setModelName(defaultModel);
          setStep(2);
          setError("");
        }),
        Effect.catchAll(() => Effect.sync(() => setModels([])))
      )
    );
  };

  const handleSubmit = () => {
    const selectedProvider = providers().find(p => p.id === provider());

    if (selectedProvider?.requires_api_key && !apiKey().trim()) {
      setError("API key is required");
      return;
    }

    const config: any = {
      model_provider: provider(),
      model_name: modelName(),
    };

    const keyMapping: Record<string, string> = {
      openai: "openai_key",
      anthropic: "anthropic_key",
      google: "google_key",
    };

    const configKey = keyMapping[provider()];
    if (configKey && apiKey().trim()) {
      config[configKey] = apiKey();
    }

    if (provider() === "ollama") {
      config.ollama_endpoint = apiKey().trim() || "http://localhost:11434";
    }

    try {
      saveConfig(config);
      setStep(3);
      setTimeout(() => {
        props.onSetupComplete();
      }, 1500);
    } catch (e) {
      setError("Failed to save configuration");
    }
  };

  const selectedProvider = () => providers().find(p => p.id === provider());

  return (
    <box flexDirection="column" flexGrow={1} padding={2}>
      <box marginBottom={2} flexDirection="column">
        <text fg="cyan">Welcome to Pryx!</text>
        <text fg="white">Setup Required</text>
        <text fg="gray">To start chatting, you need to configure an AI provider.</text>
      </box>

      {fetchError() && (
        <box marginBottom={1}>
          <text fg="yellow">⚠ {fetchError()}</text>
        </box>
      )}

      <box flexDirection="column">
        <box flexDirection="row" marginBottom={1}>
          <text fg={step() >= 1 ? "cyan" : "gray"}>Step 1: Choose Provider</text>
          {step() > 1 && <text fg="green"> ✓</text>}
        </box>

        {step() === 1 && (
          <box flexDirection="column" marginLeft={2}>
            {loading() ? (
              <text fg="gray">Loading providers...</text>
            ) : (
              providers().map(p => (
                <box
                  borderStyle="single"
                  borderColor={provider() === p.id ? "cyan" : "gray"}
                  padding={1}
                  flexDirection="column"
                >
                  <text fg="white">{p.name}</text>
                  <text fg="gray">
                    {p.requires_api_key ? "Requires API key" : "No API key required"}
                  </text>
                </box>
              ))
            )}
          </box>
        )}

        <box flexDirection="row" marginTop={1} marginBottom={1}>
          <text fg={step() >= 2 ? "cyan" : "gray"}>Step 2: API Configuration</text>
          {step() > 2 && <text fg="green"> ✓</text>}
        </box>

        {step() === 2 && (
          <box flexDirection="column" marginLeft={2}>
            <box>
              <text fg="gray">Selected: {selectedProvider()?.name}</text>
            </box>

            <box marginTop={1}>
              <text fg="gray">Model:</text>
              <box flexDirection="column">
                {models().length > 0 ? (
                  models().map(m => (
                    <box
                      borderStyle="single"
                      borderColor={modelName() === m.id ? "cyan" : "gray"}
                      padding={1}
                    >
                      <text fg={modelName() === m.id ? "cyan" : "white"}>{m.name}</text>
                    </box>
                  ))
                ) : (
                  <text fg="gray">No models available</text>
                )}
              </box>
            </box>

            <box marginTop={1}>
              <text fg="gray">
                {selectedProvider()?.requires_api_key ? "API Key:" : "Endpoint (optional):"}
              </text>
              <box
                borderStyle="single"
                borderColor={error() ? "red" : "gray"}
                padding={1}
                flexDirection="row"
              >
                <text fg="white">
                  {apiKey() ||
                    (selectedProvider()?.requires_api_key
                      ? "Enter API key..."
                      : "http://localhost:11434")}
                </text>
                <box flexGrow={1} />
                <text fg="cyan">▌</text>
              </box>
              {error() && <text fg="red">{error()}</text>}
            </box>

            <box marginTop={1}>
              <box borderStyle="single" borderColor="cyan" padding={1}>
                <text fg="cyan">Save Configuration</text>
              </box>
            </box>
          </box>
        )}

        {step() === 3 && (
          <box flexDirection="column" alignItems="center" marginTop={2}>
            <text fg="green">✓ Configuration Saved!</text>
            <text fg="gray">Starting Pryx...</text>
          </box>
        )}
      </box>

      <box flexGrow={1} />

      <box flexDirection="row">
        <text fg="gray">Need help? docs.pryx.dev</text>
      </box>
    </box>
  );
}
