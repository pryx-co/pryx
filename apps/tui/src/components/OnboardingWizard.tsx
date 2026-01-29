import { Box, Text, Input } from "@opentui/core";
import { createSignal, Show, Switch, Match } from "solid-js";
import { Effect } from "effect";
import { useEffectService } from "../lib/hooks";
import { WebSocketService } from "../services/ws";

type Step = 1 | 2 | 3 | "done";

interface WorkspaceConfig {
    name: string;
    path: string;
}

interface ProviderConfig {
    provider: string;
    apiKey: string;
}

interface IntegrationConfig {
    botToken: string;
}

export default function OnboardingWizard(props: { onComplete: () => void }) {
    const ws = useEffectService(WebSocketService);
    const [step, setStep] = createSignal<Step>(1);
    const [workspace, setWorkspace] = createSignal<WorkspaceConfig>({ name: "", path: "" });
    const [provider, setProvider] = createSignal<ProviderConfig>({ provider: "", apiKey: "" });
    const [integration, setIntegration] = createSignal<IntegrationConfig>({ botToken: "" });
    const [input, setInput] = createSignal("");
    const [field, setField] = createSignal<"name" | "path" | "provider" | "apiKey" | "botToken">("name");

    const handleSubmit = (value: string) => {
        const currentStep = step();
        const currentField = field();
        const service = ws();

        if (currentStep === 1) {
            if (currentField === "name") {
                setWorkspace(w => ({ ...w, name: value }));
                setField("path");
            } else {
                setWorkspace(w => ({ ...w, path: value }));
                setStep(2);
                setField("provider");
            }
        } else if (currentStep === 2) {
            if (currentField === "provider") {
                setProvider(p => ({ ...p, provider: value }));
                setField("apiKey");
            } else {
                setProvider(p => ({ ...p, apiKey: value }));
                setStep(3);
                setField("botToken");
            }
        } else if (currentStep === 3) {
            setIntegration({ botToken: value });
            // Save configuration
            if (service) {
                Effect.runFork(service.send({
                    event: "config.save",
                    payload: {
                        workspace: workspace(),
                        provider: provider(),
                        integration: { type: "telegram", botToken: value }
                    }
                }));
            }
            setStep("done");
            setTimeout(() => props.onComplete(), 1500);
        }
        setInput("");
    };

    const getPlaceholder = () => {
        const f = field();
        switch (f) {
            case "name": return "Workspace name (e.g., my-project)";
            case "path": return "Workspace path (e.g., ~/code/my-project)";
            case "provider": return "Model provider (openai, anthropic, google)";
            case "apiKey": return "API key (sk-...)";
            case "botToken": return "Telegram bot token (from @BotFather)";
        }
    };

    return (
        <Box flexDirection="column" flexGrow={1}>
            <Box marginBottom={1}>
                <Text bold color="cyan">Onboarding Wizard</Text>
                <Text color="gray"> - Step {step() === "done" ? "✓" : step()} of 3</Text>
            </Box>

            <Box flexDirection="row" marginBottom={1}>
                <Text color={step() === 1 ? "cyan" : step() === "done" || step() > 1 ? "green" : "gray"}>
                    ● Workspace
                </Text>
                <Text color="gray"> → </Text>
                <Text color={step() === 2 ? "cyan" : step() === "done" || step() > 2 ? "green" : "gray"}>
                    ● Provider
                </Text>
                <Text color="gray"> → </Text>
                <Text color={step() === 3 ? "cyan" : step() === "done" ? "green" : "gray"}>
                    ● Integration
                </Text>
            </Box>

            <Box flexDirection="column" flexGrow={1} borderStyle="round" padding={1}>
                <Switch>
                    <Match when={step() === 1}>
                        <Text bold>Workspace Setup</Text>
                        <Text color="gray">{field() === "name" ? "Enter a name for your workspace" : "Enter the path to your workspace"}</Text>
                    </Match>
                    <Match when={step() === 2}>
                        <Text bold>AI Provider Setup</Text>
                        <Text color="gray">{field() === "provider" ? "Choose your AI provider" : "Enter your API key"}</Text>
                    </Match>
                    <Match when={step() === 3}>
                        <Text bold>Integration Setup</Text>
                        <Text color="gray">Enter your Telegram bot token</Text>
                    </Match>
                    <Match when={step() === "done"}>
                        <Text bold color="green">✓ Setup Complete!</Text>
                        <Text color="gray">Redirecting to main interface...</Text>
                    </Match>
                </Switch>

                <Show when={step() !== "done"}>
                    <Box marginTop={1}>
                        <Input
                            placeholder={getPlaceholder()}
                            value={input()}
                            onChange={setInput}
                            onSubmit={handleSubmit}
                        />
                    </Box>
                </Show>
            </Box>
        </Box>
    );
}
