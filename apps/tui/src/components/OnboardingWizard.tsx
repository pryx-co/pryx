// @ts-nocheck
import { Box, Text, Input } from "@opentui/core";
import { createSignal, Show, Switch, Match } from "solid-js";
import { send } from "../services/ws";

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
    const [step, setStep] = createSignal<Step>(1);
    const [workspace, setWorkspace] = createSignal<WorkspaceConfig>({ name: "", path: "" });
    const [provider, setProvider] = createSignal<ProviderConfig>({ provider: "", apiKey: "" });
    const [integration, setIntegration] = createSignal<IntegrationConfig>({ botToken: "" });
    const [input, setInput] = createSignal("");
    const [field, setField] = createSignal<"name" | "path" | "provider" | "apiKey" | "botToken">("name");

    const handleSubmit = (value: string) => {
        const currentStep = step();
        const currentField = field();

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
            send({
                event: "config.save",
                payload: {
                    workspace: workspace(),
                    provider: provider(),
                    integration: { type: "telegram", ...integration() }
                }
            });
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

            <Box borderStyle="round" padding={1} flexGrow={1}>
                <Switch>
                    <Match when={step() === 1}>
                        <Box flexDirection="column">
                            <Text bold>Step 1: Configure Workspace</Text>
                            <Text color="gray">Set up your project workspace</Text>
                            <Show when={workspace().name}>
                                <Text color="green">✓ Name: {workspace().name}</Text>
                            </Show>
                        </Box>
                    </Match>
                    <Match when={step() === 2}>
                        <Box flexDirection="column">
                            <Text bold>Step 2: Select Model Provider</Text>
                            <Text color="gray">Connect your AI provider</Text>
                            <Show when={provider().provider}>
                                <Text color="green">✓ Provider: {provider().provider}</Text>
                            </Show>
                        </Box>
                    </Match>
                    <Match when={step() === 3}>
                        <Box flexDirection="column">
                            <Text bold>Step 3: Connect Telegram</Text>
                            <Text color="gray">Enable mobile access via Telegram</Text>
                        </Box>
                    </Match>
                    <Match when={step() === "done"}>
                        <Box flexDirection="column">
                            <Text bold color="green">✓ Setup Complete!</Text>
                            <Text color="gray">Starting Pryx...</Text>
                        </Box>
                    </Match>
                </Switch>
            </Box>

            <Show when={step() !== "done"}>
                <Box borderStyle="single" marginTop={1}>
                    <Input
                        placeholder={getPlaceholder()}
                        value={input()}
                        onChange={setInput}
                        onSubmit={handleSubmit}
                    />
                </Box>
            </Show>

            <Box marginTop={1}>
                <Text color="gray">Enter to continue │ Esc to skip │ Ctrl+C to exit</Text>
            </Box>
        </Box>
    );
}
