import { createSignal, Show, For } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";
import { McpService, ValidationResult } from "../services/mcp";

interface McpAddServerProps {
  onAdded: () => void;
  onCancel: () => void;
}

type Step = "input" | "validating" | "confirm" | "adding";

// Security rating colors
const ratingColors: Record<string, string> = {
  A: palette.success,
  B: palette.info,
  C: palette.accent,
  D: palette.warning,
  F: palette.error,
};

export default function McpAddServer(props: McpAddServerProps) {
  const [step, setStep] = createSignal<Step>("input");
  const [url, setUrl] = createSignal("");
  const [transport, setTransport] = createSignal("stdio");
  const [name, setName] = createSignal("");
  const [validation, setValidation] = createSignal<ValidationResult | null>(null);
  const [error, setError] = createSignal("");
  const [success, setSuccess] = createSignal("");
  const [cursorPosition, setCursorPosition] = createSignal(0);
  const [mcpService] = createSignal(new McpService());

  const transports = ["stdio", "sse", "http"];

  const handleValidate = async () => {
    if (!url().trim()) {
      setError("Please enter a server URL");
      return;
    }

    setStep("validating");
    setError("");

    try {
      const result = await mcpService().validateUrl(url(), transport());
      setValidation(result);

      if (!result.valid) {
        setError(result.errors.join(", ") || "Validation failed");
        setStep("input");
        return;
      }

      if (!name()) {
        setName(`MCP Server ${url().split("/").pop() || "Custom"}`);
      }

      setStep("confirm");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Validation failed");
      setStep("input");
    }
  };

  const handleAdd = async () => {
    setStep("adding");
    setError("");

    try {
      await mcpService().addServer({
        name: name() || "Custom MCP Server",
        url: url(),
        transport: transport(),
      });

      setSuccess("✓ Server added successfully");
      setTimeout(() => {
        props.onAdded();
      }, 1000);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to add server");
      setStep("confirm");
    }
  };

  useKeyboard(evt => {
    if (step() === "input") {
      switch (evt.name) {
        case "escape":
          evt.preventDefault();
          props.onCancel();
          break;
        case "return":
        case "enter":
          evt.preventDefault();
          handleValidate();
          break;
        case "tab": {
          evt.preventDefault();
          const currentIdx = transports.indexOf(transport());
          const nextIdx = (currentIdx + 1) % transports.length;
          setTransport(transports[nextIdx]);
          break;
        }
        case "backspace": {
          evt.preventDefault();
          const pos = cursorPosition();
          if (pos > 0) {
            const current = url();
            setUrl(current.slice(0, pos - 1) + current.slice(pos));
            setCursorPosition(pos - 1);
          }
          break;
        }
        default: {
          if (evt.name.length === 1) {
            const pos = cursorPosition();
            const current = url();
            setUrl(current.slice(0, pos) + evt.name + current.slice(pos));
            setCursorPosition(pos + 1);
          }
          break;
        }
      }
    } else if (step() === "confirm") {
      switch (evt.name) {
        case "y":
          handleAdd();
          break;
        case "n":
        case "escape":
          evt.preventDefault();
          setStep("input");
          break;
      }
    } else if (step() === "validating" || step() === "adding") {
      // Block input during async operations
      evt.preventDefault();
    }
  });

  const renderInput = () => (
    <box flexDirection="column" flexGrow={1}>
      <text fg={palette.accent} marginBottom={1}>
        Add Custom MCP Server
      </text>

      <box flexDirection="column" marginBottom={1}>
        <text fg={palette.dim}>Transport:</text>
        <box flexDirection="row" gap={2} marginTop={1}>
          <For each={transports}>
            {t => (
              <box
                borderStyle="single"
                borderColor={transport() === t ? palette.accent : palette.border}
                padding={{ left: 1, right: 1 }}
              >
                <text fg={transport() === t ? palette.accent : palette.dim}>
                  {transport() === t ? "● " : "○ "}
                  {t}
                </text>
              </box>
            )}
          </For>
        </box>
      </box>

      <box flexDirection="column" marginTop={1}>
        <text fg={palette.dim}>Server URL:</text>
        <box
          borderStyle="single"
          borderColor={palette.border}
          padding={1}
          marginTop={1}
          flexDirection="row"
        >
          <text fg={palette.text}>{url()}</text>
          <box flexGrow={1} />
          <text fg={palette.accent}>▌</text>
        </box>
      </box>

      <box flexDirection="column" marginTop={1}>
        <text fg={palette.dim}>Server Name (optional):</text>
        <box
          borderStyle="single"
          borderColor={palette.border}
          padding={1}
          marginTop={1}
          flexDirection="row"
        >
          <text fg={palette.dim}>{name() || "Custom MCP Server"}</text>
        </box>
      </box>

      <Show when={error()}>
        <box marginTop={1}>
          <text fg={palette.error}>✗ {error()}</text>
        </box>
      </Show>

      <box flexGrow={1} />

      <box flexDirection="column" marginTop={1}>
        <text fg={palette.dim}>Tab: Change Transport | Enter: Validate | Esc: Cancel</text>
      </box>
    </box>
  );

  const renderValidating = () => (
    <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center">
      <text fg={palette.accent}>Validating server...</text>
      <box marginTop={1}>
        <text fg={palette.dim}>Checking URL and security...</text>
      </box>
    </box>
  );

  const renderConfirm = () => {
    const val = validation();
    if (!val) return null;

    return (
      <box flexDirection="column" flexGrow={1}>
        <text fg={palette.accent} marginBottom={1}>
          Confirm Add Server
        </text>

        <box flexDirection="column" gap={1}>
          <box flexDirection="row">
            <text fg={palette.dim}>Name: </text>
            <text fg={palette.text}>{name() || "Custom MCP Server"}</text>
          </box>

          <box flexDirection="row">
            <text fg={palette.dim}>URL: </text>
            <text fg={palette.dim}>{url()}</text>
          </box>

          <box flexDirection="row">
            <text fg={palette.dim}>Transport: </text>
            <text fg={palette.text}>{transport()}</text>
          </box>

          <box flexDirection="row">
            <text fg={palette.dim}>Security Rating: </text>
            <text fg={ratingColors[val.securityRating] || palette.dim}>[{val.securityRating}]</text>
          </box>

          <Show when={val.warnings.length > 0}>
            <box flexDirection="column" marginTop={1}>
              <text fg={palette.warning}>⚠ Warnings:</text>
              <box flexDirection="column" marginLeft={2} marginTop={1}>
                <For each={val.warnings}>
                  {warning => <text fg={palette.warning}>• {warning}</text>}
                </For>
              </box>
            </box>
          </Show>
        </box>

        <Show when={error()}>
          <box marginTop={1}>
            <text fg={palette.error}>✗ {error()}</text>
          </box>
        </Show>

        <box flexGrow={1} />

        <box flexDirection="row" gap={2} marginTop={1}>
          <box borderStyle="single" borderColor={palette.accent} padding={1}>
            <text fg={palette.accent}>Y - Add Server</text>
          </box>
          <box borderStyle="single" borderColor={palette.border} padding={1}>
            <text fg={palette.dim}>N - Cancel</text>
          </box>
        </box>
      </box>
    );
  };

  const renderAdding = () => (
    <box flexDirection="column" flexGrow={1} alignItems="center" justifyContent="center">
      <text fg={palette.accent}>Adding server...</text>
      <Show when={success()}>
        <box marginTop={1}>
          <text fg={palette.success}>{success()}</text>
        </box>
      </Show>
    </box>
  );

  return (
    <>
      <Show when={step() === "input"}>{renderInput()}</Show>
      <Show when={step() === "validating"}>{renderValidating()}</Show>
      <Show when={step() === "confirm"}>{renderConfirm()}</Show>
      <Show when={step() === "adding"}>{renderAdding()}</Show>
    </>
  );
}
