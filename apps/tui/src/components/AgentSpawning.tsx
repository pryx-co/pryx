import { createSignal, For, Show, onMount } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { useEffectService, AppRuntime } from "../lib/hooks";
import { palette } from "../theme";

type AgentStatus = "running" | "stopped" | "idle" | "error";
type AgentType = "chat" | "task" | "code" | "analysis";

interface Agent {
  id: string;
  name: string;
  type: AgentType;
  status: AgentStatus;
  session: string | null;
  created: string;
  lastActivity: string;
  pid?: number;
}

interface CreateAgentRequest {
  name: string;
  type: AgentType;
  session: string | null;
  prompt?: string;
  tools?: string[];
}

export default function AgentSpawning() {
  const keyboard = useKeyboard();
  const [agents, setAgents] = createSignal<Agent[]>([]);
  const [selectedIndex, setSelectedIndex] = createSignal(0);
  const [showCreateModal, setShowCreateModal] = createSignal(false);
  const [newAgentName, setNewAgentName] = createSignal("");
  const [newAgentType, setNewAgentType] = createSignal<AgentType>("chat");
  const [newAgentPrompt, setNewAgentPrompt] = createSignal("");
  const [newAgentSession, setNewAgentSession] = createSignal<string | null>(null);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");

  onMount(() => {
    loadAgents();
    setupKeyboard();
    startPolling();
  });

  const setupKeyboard = () => {
    keyboard.bind("c", () => {
      setShowCreateModal(true);
      setNewAgentName("");
      setNewAgentPrompt("");
    });
    keyboard.bind("s", () => stopAgent());
    keyboard.bind("k", () => killAgent());
    keyboard.bind("v", () => viewAgent());
    keyboard.bind("a", () => attachToSession());
    keyboard.bind("r", () => restartAgent());
    keyboard.bind("l", () => viewLogs());
    keyboard.bind("q", () => {
      props.onClose();
    });
    keyboard.bind("esc", () => {
      if (showCreateModal()) {
        setShowCreateModal(false);
      }
    });
  };

  const loadAgents = async () => {
    setLoading(true);
    try {
      const response = await fetch("http://localhost:3000/api/agents");
      if (!response.ok) {
        throw new Error("Failed to load agents");
      }
      const data = await response.json();
      setAgents(data.agents || []);
    } catch (err) {
      setError(`Failed to load agents: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const startPolling = () => {
    setInterval(() => {
      loadAgents();
    }, 5000);
  };

  const createAgent = async () => {
    if (!newAgentName()) {
      setError("Agent name is required");
      return;
    }

    const request: CreateAgentRequest = {
      name: newAgentName(),
      type: newAgentType(),
      session: newAgentSession(),
      prompt: newAgentPrompt() || undefined,
    };

    try {
      const response = await fetch("http://localhost:3000/api/agents", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        throw new Error("Failed to create agent");
      }

      setShowCreateModal(false);
      loadAgents();
    } catch (err) {
      setError(`Failed to create agent: ${err.message}`);
    }
  };

  const stopAgent = async () => {
    const agent = agents()[selectedIndex()];
    if (!agent) return;

    try {
      const response = await fetch(`http://localhost:3000/api/agents/${agent.id}/stop`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to stop agent");
      }

      loadAgents();
    } catch (err) {
      setError(`Failed to stop agent: ${err.message}`);
    }
  };

  const killAgent = async () => {
    const agent = agents()[selectedIndex()];
    if (!agent) return;

    try {
      const response = await fetch(`http://localhost:3000/api/agents/${agent.id}/kill`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to kill agent");
      }

      loadAgents();
    } catch (err) {
      setError(`Failed to kill agent: ${err.message}`);
    }
  };

  const restartAgent = async () => {
    const agent = agents()[selectedIndex()];
    if (!agent) return;

    try {
      const response = await fetch(`http://localhost:3000/api/agents/${agent.id}/restart`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to restart agent");
      }

      loadAgents();
    } catch (err) {
      setError(`Failed to restart agent: ${err.message}`);
    }
  };

  const viewAgent = () => {
    const agent = agents()[selectedIndex()];
    if (!agent) return;

    console.log("View agent:", agent);
  };

  const viewLogs = () => {
    const agent = agents()[selectedIndex()];
    if (!agent) return;

    console.log("View logs:", agent);
  };

  const attachToSession = () => {
    const agent = agents()[selectedIndex()];
    if (!agent || !agent.session) return;

    console.log("Attach to session:", agent.session);
  };

  const getAgentTypeLabel = (type: AgentType) => {
    switch (type) {
      case "chat":
        return "Chat";
      case "task":
        return "Task";
      case "code":
        return "Code";
      case "analysis":
        return "Analysis";
    }
  };

  const getStatusColor = (status: AgentStatus) => {
    switch (status) {
      case "running":
        return palette.success;
      case "stopped":
        return palette.dim;
      case "idle":
        return palette.accent;
      case "error":
        return palette.error;
    }
  };

  const getStatusLabel = (status: AgentStatus) => {
    switch (status) {
      case "running":
        return "Running";
      case "stopped":
        return "Stopped";
      case "idle":
        return "Idle";
      case "error":
        return "Error";
    }
  };

  const getAgentCount = (status: AgentStatus) => {
    return agents().filter(a => a.status === status).length;
  };

  return (
    <Box flexDirection="column" width="100%" height="100%">
      <Box
        flexDirection="row"
        padding={1}
        backgroundColor={palette.primary}
        color={palette.background}
      >
        <Text bold>🤖 Agent Spawning</Text>
        <Box flexGrow={1} />
        <Text>
          <Text bold>[C]</Text>reate <Text bold>[S]</Text>top <Text bold>[K]</Text>ill{" "}
          <Text bold>[R]</Text>estart <Text bold>[V]</Text>iew <Text bold>[A]</Text>ttach{" "}
          <Text bold>[L]</Text>ogs
        </Text>
        <Text>
          Quit: <Text bold>[Q]</Text>
        </Text>
      </Box>

      <Show when={loading()}>
        <Box padding={2}>
          <Text>Loading agents...</Text>
        </Box>
      </Show>

      <Show when={error()}>
        <Box padding={1} backgroundColor={palette.error}>
          <Text color={palette.background}>{error()}</Text>
        </Box>
      </Show>

      <Show when={!loading() && !error()}>
        <Box flexDirection="column" padding={1}>
          <Box flexDirection="row" padding={1} backgroundColor={palette.bgSecondary}>
            <Box flexGrow={1}>
              <Text bold>Total Agents</Text>
              <Text fontSize={2}>{agents().length}</Text>
            </Box>
            <Box flexGrow={1}>
              <Text bold>Running</Text>
              <Text fontSize={2} color={palette.success}>
                {getAgentCount("running")}
              </Text>
            </Box>
            <Box flexGrow={1}>
              <Text bold>Idle</Text>
              <Text fontSize={2} color={palette.accent}>
                {getAgentCount("idle")}
              </Text>
            </Box>
            <Box flexGrow={1}>
              <Text bold>Stopped</Text>
              <Text fontSize={2} color={palette.dim}>
                {getAgentCount("stopped")}
              </Text>
            </Box>
          </Box>

          <Show when={showCreateModal()}>
            <Box
              flexDirection="column"
              padding={1}
              marginTop={1}
              backgroundColor={palette.bgSelected}
              border={`1px solid ${palette.border}`}
            >
              <Text bold>Create New Agent</Text>
              <Box marginTop={1}>
                <Text width={20}>Name:</Text>
                <Box flexGrow={1}>
                  <TextInput
                    value={newAgentName()}
                    onInput={(e: any) => setNewAgentName(e.target.value)}
                    placeholder="Agent name"
                  />
                </Box>
              </Box>
              <Box marginTop={1}>
                <Text width={20}>Type:</Text>
                <Box flexGrow={1}>
                  <Select
                    value={newAgentType()}
                    onChange={(e: any) => setNewAgentType(e.target.value)}
                  >
                    <option value="chat">Chat</option>
                    <option value="task">Task</option>
                    <option value="code">Code</option>
                    <option value="analysis">Analysis</option>
                  </Select>
                </Box>
              </Box>
              <Box marginTop={1}>
                <Text width={20}>Prompt:</Text>
                <Box flexGrow={1}>
                  <TextInput
                    value={newAgentPrompt()}
                    onInput={(e: any) => setNewAgentPrompt(e.target.value)}
                    placeholder="System prompt (optional)"
                    multiline
                  />
                </Box>
              </Box>
              <Box flexDirection="row" marginTop={1}>
                <Box flexGrow={1}>
                  <Button onClick={createAgent}>Create</Button>
                </Box>
                <Box flexGrow={1}>
                  <Button onClick={() => setShowCreateModal(false)}>Cancel</Button>
                </Box>
              </Box>
            </Box>
          </Show>

          <Box padding={1} marginTop={1} backgroundColor={palette.background}>
            <Text bold>Active Agents</Text>
          </Box>

          <Box flexDirection="column" flexGrow={1} padding={1} backgroundColor={palette.background}>
            <For each={agents()}>
              {(agent, index) => (
                <Box
                  flexDirection="row"
                  padding={0.5}
                  backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
                  color={index() === selectedIndex() ? palette.text : undefined}
                  onClick={() => setSelectedIndex(index())}
                >
                  <Text width={25}>{agent.name}</Text>
                  <Text width={15}>{getAgentTypeLabel(agent.type)}</Text>
                  <Text width={15} color={getStatusColor(agent.status)}>
                    {getStatusLabel(agent.status)}
                  </Text>
                  <Text width={20}>{agent.created}</Text>
                  <Text width={20}>{agent.lastActivity}</Text>
                </Box>
              )}
            </For>

            <Show when={agents().length === 0}>
              <Box padding={2} textAlign="center">
                <Text color={palette.dim}>No agents running. Press [C] to create one.</Text>
              </Box>
            </Show>
          </Box>

          <Box flexDirection="row" padding={1} marginTop={1} backgroundColor={palette.bgSecondary}>
            <Text>
              Create: <Text bold>[C]</Text>
            </Text>
            <Box flexGrow={1} />
            <Text>
              Stop: <Text bold>[S]</Text> Kill: <Text bold>[K]</Text> Restart: <Text bold>[R]</Text>
            </Text>
          </Box>
        </Box>
      </Show>
    </Box>
  );
}

const Box: any = (props: any) => props.children;
const Text: any = (props: any) => {
  const content =
    typeof props.children === "string" ? props.children : props.children?.join?.("") || "";
  return <span style={props}>{content}</span>;
};
const TextInput: any = (props: any) => (
  <input
    type={props.multiline ? "textarea" : "text"}
    value={props.value}
    onInput={props.onInput}
    placeholder={props.placeholder}
    style={{
      width: "100%",
      padding: "0.5",
      backgroundColor: palette.bgSecondary,
      border: `1px solid ${palette.border}`,
      color: palette.text,
      ...props.style,
    }}
  />
);
const Select: any = (props: any) => (
  <select
    value={props.value}
    onChange={props.onChange}
    style={{
      width: "100%",
      padding: "0.5",
      backgroundColor: palette.bgSecondary,
      border: `1px solid ${palette.border}`,
      color: palette.text,
      ...props.style,
    }}
  >
    {props.children}
  </select>
);
const Button: any = (props: any) => (
  <button
    onClick={props.onClick}
    style={{
      padding: "0.5 1",
      backgroundColor: palette.primary,
      color: palette.background,
      border: "none",
      cursor: "pointer",
      ...props.style,
    }}
  >
    {props.children}
  </button>
);
