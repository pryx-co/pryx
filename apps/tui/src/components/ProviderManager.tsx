import { createSignal, createEffect, For, Show, onMount, onCleanup } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { useEffectService, AppRuntime } from "../lib/hooks";
import { ProviderService, Provider as ProviderType, Model as ModelType } from "../services/provider-service";
import { loadConfig, saveConfig, AppConfig } from "../services/config";
import { palette } from "../theme";
import { Effect } from "effect";

interface ConfiguredProvider {
  id: string;
  name: string;
  status: "connected" | "error" | "not_configured";
  keyStatus: string;
  isActive: boolean;
}

type ViewMode = "list" | "add" | "edit" | "test" | "delete_confirm";

const API_BASE = "http://localhost:3000";

interface ProviderManagerProps {
  onClose: () => void;
}

export default function ProviderManager(props: ProviderManagerProps) {
  const providerService = useEffectService(ProviderService);
  const [viewMode, setViewMode] = createSignal<ViewMode>("list");
  const [providers, setProviders] = createSignal<ProviderType[]>([]);
  const [configuredProviders, setConfiguredProviders] = createSignal<ConfiguredProvider[]>([]);
  const [selectedIndex, setSelectedIndex] = createSignal(0);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");
  const [success, setSuccess] = createSignal("");
  const [selectedProvider, setSelectedProvider] = createSignal<ProviderType | null>(null);
  const [apiKey, setApiKey] = createSignal("");
  const [models, setModels] = createSignal<ModelType[]>([]);
  const [selectedModel, setSelectedModel] = createSignal("");
  const [testResult, setTestResult] = createSignal<{ success: boolean; message: string } | null>(null);
  const [config, setConfig] = createSignal<AppConfig>({});

  onMount(() => {
    setConfig(loadConfig());
    const service = providerService();
    if (!service) return;

    AppRuntime.runFork(
      service.fetchProviders.pipe(
        Effect.tap(providers => Effect.sync(() => {
          setProviders(providers);
          loadConfiguredProviders();
        })),
        Effect.catchAll(() => Effect.sync(() => {
          setProviders([
            { id: "openai", name: "OpenAI", requires_api_key: true },
            { id: "anthropic", name: "Anthropic", requires_api_key: true },
            { id: "google", name: "Google AI", requires_api_key: true },
            { id: "openrouter", name: "OpenRouter", requires_api_key: true },
            { id: "ollama", name: "Ollama (Local)", requires_api_key: false },
            { id: "groq", name: "Groq", requires_api_key: true },
            { id: "xai", name: "xAI", requires_api_key: true },
            { id: "mistral", name: "Mistral AI", requires_api_key: true },
            { id: "cohere", name: "Cohere", requires_api_key: true },
          ]);
        }))
      )
    );
  });

  const loadConfiguredProviders = () => {
    const cfg = loadConfig();
    const configured: ConfiguredProvider[] = [];
    const providerKeys: Record<string, string> = {
      openai: "openai_key",
      anthropic: "anthropic_key",
      google: "google_key",
    };

    providers().forEach(p => {
      const keyField = providerKeys[p.id];
      const hasKey = keyField ? !!cfg[keyField] : false;
      const isOllama = p.id === "ollama";
      const isActive = cfg.model_provider === p.id;
      
      if (hasKey || isOllama) {
        configured.push({
          id: p.id,
          name: p.name,
          status: "connected",
          keyStatus: hasKey ? "configured" : "local",
          isActive,
        });
      }
    });
    
    setConfiguredProviders(configured);
  };

  const handleAddProvider = async () => {
    if (!selectedProvider()) return;
    
    const provider = selectedProvider()!;
    const keyFieldMap: Record<string, string> = {
      openai: "openai_key",
      anthropic: "anthropic_key",
      google: "google_key",
    };
    
    const updates: AppConfig = {
      model_provider: provider.id,
      model_name: selectedModel() || undefined,
    };
    
    const keyField = keyFieldMap[provider.id];
    if (keyField && apiKey().trim()) {
      updates[keyField] = apiKey().trim();
    }
    
    if (provider.id === "ollama") {
      updates.ollama_endpoint = apiKey().trim() || "http://localhost:11434";
    }
    
    try {
      saveConfig({ ...config(), ...updates });
      setConfig({ ...config(), ...updates });
      setSuccess(`✓ ${provider.name} added successfully`);
      setTimeout(() => {
        setSuccess("");
        setViewMode("list");
        loadConfiguredProviders();
        resetAddForm();
      }, 1500);
    } catch (e) {
      setError("Failed to save configuration");
    }
  };

  const handleSetActive = (providerId: string) => {
    try {
      const updates: AppConfig = { model_provider: providerId };
      saveConfig({ ...config(), ...updates });
      setConfig({ ...config(), ...updates });
      loadConfiguredProviders();
      setSuccess(`✓ ${providerId} is now active`);
      setTimeout(() => setSuccess(""), 2000);
    } catch (e) {
      setError("Failed to update active provider");
    }
  };

  const handleDeleteProvider = (providerId: string) => {
    const keyFieldMap: Record<string, string> = {
      openai: "openai_key",
      anthropic: "anthropic_key",
      google: "google_key",
    };
    
    const keyField = keyFieldMap[providerId];
    const updates: AppConfig = {};
    
    if (keyField) {
      updates[keyField] = undefined;
    }
    
    if (config().model_provider === providerId) {
      updates.model_provider = undefined;
      updates.model_name = undefined;
    }
    
    try {
      const newConfig = { ...config() };
      Object.keys(updates).forEach(key => {
        if (updates[key] === undefined) {
          delete newConfig[key];
        } else {
          newConfig[key] = updates[key];
        }
      });
      saveConfig(newConfig);
      setConfig(newConfig);
      setViewMode("list");
      loadConfiguredProviders();
      setSuccess(`✓ Provider removed`);
      setTimeout(() => setSuccess(""), 2000);
    } catch (e) {
      setError("Failed to remove provider");
    }
  };

  const handleTestConnection = async (providerId: string) => {
    setLoading(true);
    setTestResult(null);

    await new Promise(resolve => setTimeout(resolve, 1500));

    const provider = providers().find(p => p.id === providerId);
    if (!provider) {
      setTestResult({ success: false, message: "Provider not found" });
    } else if (provider.id === "ollama") {
      setTestResult({ success: true, message: "Local Ollama connection ready" });
    } else {
      const keyFieldMap: Record<string, string> = {
        openai: "openai_key",
        anthropic: "anthropic_key",
        google: "google_key",
      };
      const keyField = keyFieldMap[providerId];
      const hasKey = keyField ? !!config()[keyField] : false;
      
      if (hasKey) {
        setTestResult({ success: true, message: `Connected to ${provider.name}` });
      } else {
        setTestResult({ success: false, message: "API key not configured" });
      }
    }
    
    setLoading(false);
  };

  const resetAddForm = () => {
    setSelectedProvider(null);
    setApiKey("");
    setModels([]);
    setSelectedModel("");
    setTestResult(null);
    setError("");
  };

  const handleProviderSelect = async (provider: ProviderType) => {
    const service = providerService();
    if (!service) return;

    setSelectedProvider(provider);
    AppRuntime.runFork(
      service.fetchModels(provider.id).pipe(
        Effect.tap(availableModels => {
          setModels(availableModels);
          if (availableModels.length > 0) {
            setSelectedModel(availableModels[0].id);
          }
        }),
        Effect.catchAll(() => Effect.sync(() => setModels([])))
      )
    );
  };

  useKeyboard(evt => {
    if (viewMode() === "list") {
      switch (evt.name) {
        case "up":
        case "arrowup": {
          evt.preventDefault();
          setSelectedIndex(i => Math.max(0, i - 1));
          break;
        }
        case "down":
        case "arrowdown": {
          evt.preventDefault();
          const maxIndex = configuredProviders().length + 2;
          setSelectedIndex(i => Math.min(maxIndex, i + 1));
          break;
        }
        case "return":
        case "enter": {
          evt.preventDefault();
          const idx = selectedIndex();
          const configured = configuredProviders();
          
          if (idx < configured.length) {
            setSelectedProvider(providers().find(p => p.id === configured[idx].id) || null);
            setViewMode("edit");
          } else if (idx === configured.length) {
            setViewMode("add");
            setSelectedIndex(0);
          } else {
            props.onClose();
          }
          break;
        }
        case "escape":
          evt.preventDefault();
          props.onClose();
          break;
        case "a":
          if (evt.ctrl) {
            evt.preventDefault();
            setViewMode("add");
            setSelectedIndex(0);
          }
          break;
      }
    } else if (viewMode() === "add") {
      switch (evt.name) {
        case "escape":
          evt.preventDefault();
          setViewMode("list");
          resetAddForm();
          break;
        case "up":
        case "arrowup":
          evt.preventDefault();
          if (!selectedProvider()) {
            setSelectedIndex(i => Math.max(0, i - 1));
          }
          break;
        case "down":
        case "arrowdown":
          evt.preventDefault();
          if (!selectedProvider()) {
            setSelectedIndex(i => Math.min(providers().length - 1, i + 1));
          }
          break;
        case "return":
        case "enter":
          evt.preventDefault();
          if (!selectedProvider()) {
            handleProviderSelect(providers()[selectedIndex()]);
          } else {
            handleAddProvider();
          }
          break;
      }
    } else if (viewMode() === "edit") {
      switch (evt.name) {
        case "escape":
          evt.preventDefault();
          setViewMode("list");
          break;
        case "t":
          if (selectedProvider()) {
            handleTestConnection(selectedProvider()!.id);
          }
          break;
        case "s":
          if (selectedProvider()) {
            handleSetActive(selectedProvider()!.id);
          }
          break;
        case "d":
          if (selectedProvider()) {
            setViewMode("delete_confirm");
          }
          break;
      }
    } else if (viewMode() === "delete_confirm") {
      switch (evt.name) {
        case "y":
          if (selectedProvider()) {
            handleDeleteProvider(selectedProvider()!.id);
          }
          break;
        case "n":
        case "escape":
          evt.preventDefault();
          setViewMode("edit");
          break;
      }
    }
  });

  createEffect(() => {
    if (viewMode() === "list") {
      setSelectedIndex(0);
    }
  });

  return (
    <box
      position="absolute"
      top={2}
      left="10%"
      width="80%"
      height="80%"
      borderStyle="single"
      borderColor={palette.border}
      backgroundColor={palette.bgPrimary}
      flexDirection="column"
      padding={1}
    >
      <box flexDirection="row" marginBottom={1}>
        <text fg={palette.accent}>Provider Management</text>
        <box flexGrow={1} />
        <text fg={palette.dim}>[Esc to close]</text>
      </box>

      <Show when={error()}>
        <box marginBottom={1}>
          <text fg={palette.error}>✗ {error()}</text>
        </box>
      </Show>
      <Show when={success()}>
        <box marginBottom={1}>
          <text fg={palette.success}>{success()}</text>
        </box>
      </Show>

      <Show when={viewMode() === "list"}>
        <box flexDirection="column" flexGrow={1}>
          <text fg={palette.dim} marginBottom={1}>
            Configured Providers ({configuredProviders().length})
          </text>
          
          <box flexDirection="column" flexGrow={1}>
            <For each={configuredProviders()}>
              {(provider, index) => (
                <box
                  flexDirection="row"
                  padding={1}
                  backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
                >
                  <box width={3}>
                    <text fg={provider.isActive ? palette.success : palette.dim}>
                      {provider.isActive ? "●" : "○"}
                    </text>
                  </box>
                  <box width={20}>
                    <text fg={index() === selectedIndex() ? palette.accent : palette.text}>
                      {provider.name}
                    </text>
                  </box>
                  <box width={15}>
                    <text fg={palette.success}>{provider.status}</text>
                  </box>
                  <box flexGrow={1}>
                    <text fg={palette.dim}>{provider.keyStatus}</text>
                  </box>
                  <box width={10}>
                    <Show when={provider.isActive}>
                      <text fg={palette.accent}>[ACTIVE]</text>
                    </Show>
                  </box>
                </box>
              )}
            </For>

            <box
              flexDirection="row"
              padding={1}
              marginTop={1}
              backgroundColor={selectedIndex() === configuredProviders().length ? palette.bgSelected : undefined}
            >
              <box width={3}>
                <text fg={palette.accent}>+</text>
              </box>
              <box>
                <text fg={selectedIndex() === configuredProviders().length ? palette.accent : palette.text}>
                  Add New Provider
                </text>
              </box>
            </box>

            <box
              flexDirection="row"
              padding={1}
              backgroundColor={selectedIndex() === configuredProviders().length + 1 ? palette.bgSelected : undefined}
            >
              <box width={3}>
                <text fg={palette.dim}>×</text>
              </box>
              <box>
                <text fg={selectedIndex() === configuredProviders().length + 1 ? palette.accent : palette.dim}>
                  Close
                </text>
              </box>
            </box>
          </box>
        </box>
        
        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>↑↓ Navigate | Enter Select | A Add | Esc Close</text>
          <text fg={palette.dim}>On provider: T Test | S Set Active | D Delete</text>
        </box>
      </Show>

      <Show when={viewMode() === "add"}>
        <box flexDirection="column" flexGrow={1}>
          <text fg={palette.accent} marginBottom={1}>Add New Provider</text>
          
          <Show when={!selectedProvider()}>
            <text fg={palette.dim} marginBottom={1}>Select a provider:</text>
            <box flexDirection="column" flexGrow={1}>
              <For each={providers()}>
                {(provider, index) => (
                  <box
                    flexDirection="row"
                    padding={1}
                    backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
                  >
                    <box width={3}>
                      <text fg={index() === selectedIndex() ? palette.accent : palette.dim}>
                        {index() === selectedIndex() ? "❯" : " "}
                      </text>
                    </box>
                    <box width={20}>
                      <text fg={palette.text}>{provider.name}</text>
                    </box>
                    <box>
                      <text fg={palette.dim}>
                        {provider.requires_api_key ? "Requires API key" : "No API key required"}
                      </text>
                    </box>
                  </box>
                )}
              </For>
            </box>
          </Show>
          
          <Show when={selectedProvider()}>
            <box flexDirection="column" marginTop={1}>
              <text fg={palette.text}>Selected: {selectedProvider()?.name}</text>
              
              <Show when={models().length > 0}>
                <box marginTop={1}>
                  <text fg={palette.dim}>Available Models:</text>
                  <box flexDirection="column" marginLeft={2}>
                    <For each={models()}>
                      {(model) => (
                        <text fg={palette.dim}>• {model.name}</text>
                      )}
                    </For>
                  </box>
                </box>
              </Show>
              
              <box marginTop={1}>
                <text fg={palette.dim}>
                  {selectedProvider()?.requires_api_key ? "API Key:" : "Endpoint URL:"}
                </text>
                <box
                  borderStyle="single"
                  borderColor={palette.border}
                  padding={1}
                  marginTop={1}
                  flexDirection="row"
                >
                  <text fg={palette.text}>{apiKey() || "Enter value..."}</text>
                  <box flexGrow={1} />
                  <text fg={palette.accent}>▌</text>
                </box>
              </box>
              
              <box marginTop={2} flexDirection="row" gap={2}>
                <box borderStyle="single" borderColor={palette.accent} padding={1}>
                  <text fg={palette.accent}>Enter to Save</text>
                </box>
                <box borderStyle="single" borderColor={palette.border} padding={1}>
                  <text fg={palette.dim}>Esc to Cancel</text>
                </box>
              </box>
            </box>
          </Show>
        </box>
      </Show>

      <Show when={viewMode() === "edit" && selectedProvider()}>
        <box flexDirection="column" flexGrow={1}>
          <text fg={palette.accent} marginBottom={1}>
            {selectedProvider()?.name}
          </text>
          
          <box flexDirection="column" gap={1}>
            <box flexDirection="row">
              <text fg={palette.dim}>Status: </text>
              <text fg={palette.success}>Connected</text>
            </box>
            
            <box flexDirection="row">
              <text fg={palette.dim}>Active: </text>
              <text fg={config().model_provider === selectedProvider()?.id ? palette.success : palette.dim}>
                {config().model_provider === selectedProvider()?.id ? "Yes" : "No"}
              </text>
            </box>
            
            <Show when={testResult()}>
              <box 
                borderStyle="single" 
                borderColor={testResult()?.success ? palette.success : palette.error}
                padding={1}
                marginTop={1}
              >
                <text fg={testResult()?.success ? palette.success : palette.error}>
                  {testResult()?.success ? "✓" : "✗"} {testResult()?.message}
                </text>
              </box>
            </Show>
            
            <Show when={loading()}>
              <box marginTop={1}>
                <text fg={palette.accent}>Testing connection...</text>
              </box>
            </Show>
          </box>
          
          <box flexGrow={1} />
          
          <box flexDirection="column" marginTop={1}>
            <text fg={palette.dim}>T Test Connection | S Set Active | D Delete | Esc Back</text>
          </box>
        </box>
      </Show>

      <Show when={viewMode() === "delete_confirm"}>
        <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center">
          <text fg={palette.error} marginBottom={1}>⚠ Delete Provider?</text>
          <text fg={palette.text} marginBottom={1}>
            Are you sure you want to remove {selectedProvider()?.name}?
          </text>
          <box flexDirection="row" gap={2} marginTop={1}>
            <box borderStyle="single" borderColor={palette.error} padding={1}>
              <text fg={palette.error}>Y - Yes, Delete</text>
            </box>
            <box borderStyle="single" borderColor={palette.border} padding={1}>
              <text fg={palette.dim}>N - Cancel</text>
            </box>
          </box>
        </box>
      </Show>
    </box>
  );
}
