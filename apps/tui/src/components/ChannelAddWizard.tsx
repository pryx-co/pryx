import { createSignal, createEffect, For, Show, onMount } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";
import type { ChannelType, ChannelConfig, ChannelFormData } from "../types/channels";
import { CHANNEL_TYPE_LABELS, DEFAULT_CHANNEL_CONFIGS } from "../types/channels";
import { createChannel, testConnection } from "../services/channels";

interface ChannelAddWizardProps {
  onComplete: () => void;
  onCancel: () => void;
}

type WizardStep = "type" | "config" | "test" | "confirm";

const PLATFORM_ICONS: Record<ChannelType, string> = {
  webhook: "🔌",
  telegram: "✈️",
  discord: "🎮",
  slack: "💬",
  email: "📧",
  whatsapp: "📱",
};

const AVAILABLE_TYPES: ChannelType[] = [
  "webhook",
  "telegram",
  "discord",
  "slack",
  "email",
  "whatsapp",
];

export default function ChannelAddWizard(props: ChannelAddWizardProps) {
  const [step, setStep] = createSignal<WizardStep>("type");
  const [selectedTypeIndex, setSelectedTypeIndex] = createSignal(0);
  const [formData, setFormData] = createSignal<ChannelFormData>({
    name: "",
    type: "webhook",
    enabled: true,
    config: {},
  });
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");
  const [success, setSuccess] = createSignal("");
  const [testResult, setTestResult] = createSignal<{
    success: boolean;
    message: string;
  } | null>(null);
  const [focusedField, setFocusedField] = createSignal(0);

  const currentType = () => AVAILABLE_TYPES[selectedTypeIndex()];

  createEffect(() => {
    setFormData(prev => ({
      ...prev,
      type: currentType(),
      config: { ...DEFAULT_CHANNEL_CONFIGS[currentType()] },
    }));
  });

  useKeyboard(evt => {
    if (step() === "type") {
      switch (evt.name) {
        case "up":
        case "arrowup": {
          evt.preventDefault();
          setSelectedTypeIndex(i => Math.max(0, i - 1));
          break;
        }
        case "down":
        case "arrowdown": {
          evt.preventDefault();
          setSelectedTypeIndex(i => Math.min(AVAILABLE_TYPES.length - 1, i + 1));
          break;
        }
        case "return":
        case "enter": {
          evt.preventDefault();
          setStep("config");
          setFocusedField(0);
          break;
        }
        case "escape": {
          evt.preventDefault();
          props.onCancel();
          break;
        }
      }
    } else if (step() === "config") {
      const fields = getConfigFields();
      switch (evt.name) {
        case "up":
        case "arrowup": {
          evt.preventDefault();
          setFocusedField(i => Math.max(0, i - 1));
          break;
        }
        case "down":
        case "arrowdown": {
          evt.preventDefault();
          setFocusedField(i => Math.min(fields.length - 1, i + 1));
          break;
        }
        case "tab": {
          evt.preventDefault();
          setFocusedField(i => (i + 1) % fields.length);
          break;
        }
        case "return":
        case "enter": {
          evt.preventDefault();
          if (validateConfig()) {
            setStep("test");
            handleTest();
          }
          break;
        }
        case "escape": {
          evt.preventDefault();
          setStep("type");
          break;
        }
      }
    } else if (step() === "test") {
      switch (evt.name) {
        case "return":
        case "enter": {
          evt.preventDefault();
          if (testResult()?.success) {
            setStep("confirm");
          } else {
            setStep("config");
          }
          break;
        }
        case "escape": {
          evt.preventDefault();
          setStep("config");
          break;
        }
      }
    } else if (step() === "confirm") {
      switch (evt.name) {
        case "return":
        case "enter": {
          evt.preventDefault();
          handleSave();
          break;
        }
        case "escape": {
          evt.preventDefault();
          setStep("config");
          break;
        }
      }
    }
  });

  const getConfigFields = () => {
    const type = formData().type;
    const baseFields = [{ key: "name", label: "Channel Name", type: "text", required: true }];

    switch (type) {
      case "webhook":
        return [
          ...baseFields,
          { key: "url", label: "Webhook URL", type: "text", required: true },
          {
            key: "method",
            label: "HTTP Method",
            type: "select",
            options: ["POST", "GET", "PUT", "DELETE", "PATCH"],
            required: true,
          },
          { key: "secret", label: "Secret (optional)", type: "password" },
        ];
      case "telegram":
        return [
          ...baseFields,
          {
            key: "token",
            label: "Bot Token",
            type: "password",
            required: true,
          },
          {
            key: "mode",
            label: "Mode",
            type: "select",
            options: ["polling", "webhook"],
            required: true,
          },
        ];
      case "discord":
        return [
          ...baseFields,
          {
            key: "token",
            label: "Bot Token",
            type: "password",
            required: true,
          },
        ];
      case "slack":
        return [
          ...baseFields,
          {
            key: "appToken",
            label: "App Token",
            type: "password",
            required: true,
          },
          {
            key: "botToken",
            label: "Bot Token",
            type: "password",
            required: true,
          },
          {
            key: "mode",
            label: "Mode",
            type: "select",
            options: ["socket", "webhook"],
            required: true,
          },
        ];
      case "email":
        return [
          ...baseFields,
          {
            key: "imap.host",
            label: "IMAP Host",
            type: "text",
            required: true,
          },
          {
            key: "imap.port",
            label: "IMAP Port",
            type: "number",
            required: true,
          },
          {
            key: "imap.username",
            label: "IMAP Username",
            type: "text",
            required: true,
          },
          {
            key: "imap.password",
            label: "IMAP Password",
            type: "password",
            required: true,
          },
        ];
      case "whatsapp":
        return [
          ...baseFields,
          {
            key: "sessionData",
            label: "Session Data (optional)",
            type: "text",
          },
        ];
      default:
        return baseFields;
    }
  };

  const validateConfig = () => {
    const data = formData();
    if (!data.name.trim()) {
      setError("Channel name is required");
      return false;
    }

    const fields = getConfigFields();
    for (const field of fields) {
      if (field.required) {
        const value = getNestedValue(data.config, field.key);
        if (!value || (typeof value === "string" && !value.trim())) {
          setError(`${field.label} is required`);
          return false;
        }
      }
    }

    setError("");
    return true;
  };

  const getNestedValue = (obj: any, path: string) => {
    return path.split(".").reduce((o, p) => (o || {})[p], obj);
  };

  const setNestedValue = (obj: any, path: string, value: any) => {
    const parts = path.split(".");
    const last = parts.pop()!;
    const target = parts.reduce((o, p) => {
      if (!o[p]) o[p] = {};
      return o[p];
    }, obj);
    target[last] = value;
  };

  const handleFieldChange = (key: string, value: any) => {
    setFormData(prev => {
      const newConfig = { ...prev.config };
      setNestedValue(newConfig, key, value);
      return { ...prev, config: newConfig };
    });
  };

  const handleNameChange = (value: string) => {
    setFormData(prev => ({ ...prev, name: value }));
  };

  const handleTest = async () => {
    setLoading(true);
    setError("");
    try {
      const result = await testConnection("new");
      setTestResult({
        success: result.success,
        message: result.message,
      });
    } catch (e) {
      setTestResult({
        success: false,
        message: "Connection test failed: " + (e as Error).message,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setLoading(true);
    setError("");
    try {
      await createChannel({
        ...formData(),
        type: formData().type,
      } as any);
      setSuccess("Channel created successfully!");
      setTimeout(() => {
        props.onComplete();
      }, 1500);
    } catch (e) {
      setError("Failed to create channel: " + (e as Error).message);
      setStep("config");
    } finally {
      setLoading(false);
    }
  };

  const renderProgressBar = () => {
    const steps: WizardStep[] = ["type", "config", "test", "confirm"];
    const currentIdx = steps.indexOf(step());
    return (
      <box flexDirection="row" marginBottom={1}>
        <For each={steps}>
          {(s, idx) => (
            <>
              <text fg={idx() <= currentIdx ? palette.accent : palette.dim}>
                {idx() <= currentIdx ? "●" : "○"} {s.charAt(0).toUpperCase() + s.slice(1)}
              </text>
              <Show when={idx() < steps.length - 1}>
                <text fg={palette.dim}> → </text>
              </Show>
            </>
          )}
        </For>
      </box>
    );
  };

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
      {/* Header */}
      <box flexDirection="row" marginBottom={1}>
        <text fg={palette.accent}>Add New Channel</text>
        <box flexGrow={1} />
        <text fg={palette.dim}>[Esc to cancel]</text>
      </box>

      {/* Progress */}
      {renderProgressBar()}

      {/* Error/Success */}
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

      {/* Step 1: Select Type */}
      <Show when={step() === "type"}>
        <box flexDirection="column" flexGrow={1}>
          <text fg={palette.dim} marginBottom={1}>
            Select channel type:
          </text>
          <For each={AVAILABLE_TYPES}>
            {(type, index) => (
              <box
                flexDirection="row"
                padding={1}
                backgroundColor={index() === selectedTypeIndex() ? palette.bgSelected : undefined}
              >
                <box width={4}>
                  <text>{PLATFORM_ICONS[type]}</text>
                </box>
                <box width={15}>
                  <text fg={index() === selectedTypeIndex() ? palette.accent : palette.text}>
                    {CHANNEL_TYPE_LABELS[type]}
                  </text>
                </box>
                <box flexGrow={1}>
                  <text fg={palette.dim}>
                    {type === "webhook" && "Custom HTTP webhooks"}
                    {type === "telegram" && "Telegram Bot API"}
                    {type === "discord" && "Discord Bot Gateway"}
                    {type === "slack" && "Slack App/Bot"}
                    {type === "email" && "IMAP/SMTP Email"}
                    {type === "whatsapp" && "WhatsApp Web"}
                  </text>
                </box>
              </box>
            )}
          </For>
        </box>
        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>↑↓ Select | Enter Continue | Esc Cancel</text>
        </box>
      </Show>

      {/* Step 2: Configuration */}
      <Show when={step() === "config"}>
        <box flexDirection="column" flexGrow={1}>
          <text fg={palette.dim} marginBottom={1}>
            Configure {CHANNEL_TYPE_LABELS[formData().type]}:
          </text>
          <box flexDirection="column">
            <For each={getConfigFields()}>
              {(field, index) => (
                <box
                  flexDirection="row"
                  padding={1}
                  backgroundColor={index() === focusedField() ? palette.bgSelected : undefined}
                >
                  <box width={20}>
                    <text fg={index() === focusedField() ? palette.accent : palette.text}>
                      {field.label}
                      {field.required && <text fg={palette.error}>*</text>}:
                    </text>
                  </box>
                  <box flexGrow={1}>
                    <Show
                      when={field.type === "select"}
                      fallback={
                        <text fg={palette.text}>
                          {field.type === "password"
                            ? "•".repeat(
                                String(getNestedValue(formData().config, field.key) || "").length
                              )
                            : String(getNestedValue(formData().config, field.key) || "")}
                          {index() === focusedField() && <text fg={palette.accent}>▌</text>}
                        </text>
                      }
                    >
                      <text fg={palette.text}>
                        {String(getNestedValue(formData().config, field.key) || field.options?.[0])}
                        <text fg={palette.dim}> ▼</text>
                      </text>
                    </Show>
                  </box>
                </box>
              )}
            </For>
          </box>
        </box>
        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>↑↓ Navigate | Tab Next | Enter Test | Esc Back</text>
        </box>
      </Show>

      {/* Step 3: Test */}
      <Show when={step() === "test"}>
        <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center">
          <Show when={loading()}>
            <text fg={palette.accent} marginBottom={1}>
              Testing connection...
            </text>
          </Show>
          <Show when={!loading() && testResult()}>
            <Show when={testResult()?.success}>
              <text fg={palette.success} marginBottom={1}>
                ✓ Connection Test Passed
              </text>
            </Show>
            <Show when={!testResult()?.success}>
              <text fg={palette.error} marginBottom={1}>
                ✗ Connection Test Failed
              </text>
            </Show>
            <text fg={palette.text}>{testResult()?.message}</text>
          </Show>
        </box>
        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>
            {testResult()?.success ? "Enter Continue | Esc Back" : "Enter Retry | Esc Back"}
          </text>
        </box>
      </Show>

      {/* Step 4: Confirm */}
      <Show when={step() === "confirm"}>
        <box flexDirection="column" flexGrow={1}>
          <text fg={palette.accent} marginBottom={1}>
            Confirm Channel Creation
          </text>
          <box flexDirection="column" gap={1}>
            <box flexDirection="row">
              <text fg={palette.dim}>Name: </text>
              <text fg={palette.text}>{formData().name}</text>
            </box>
            <box flexDirection="row">
              <text fg={palette.dim}>Type: </text>
              <text fg={palette.text}>
                {PLATFORM_ICONS[formData().type]} {CHANNEL_TYPE_LABELS[formData().type]}
              </text>
            </box>
            <box flexDirection="row">
              <text fg={palette.dim}>Status: </text>
              <text fg={palette.success}>Ready to connect</text>
            </box>
          </box>
          <box marginTop={2}>
            <text fg={palette.dim}>
              The channel will be created and connection will be established.
            </text>
          </box>
        </box>
        <box flexDirection="column" marginTop={1}>
          <text fg={palette.dim}>Enter Save | Esc Back</text>
        </box>
      </Show>
    </box>
  );
}
