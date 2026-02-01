import { createSignal, For, Show, onMount } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";

type ApprovalAction = "allow" | "deny" | "require_review";
type ActionType = "file_ops" | "shell" | "network" | "channel_message" | "credential_access";

interface PolicyRule {
  id: string;
  name: string;
  actionType: ActionType;
  action: ApprovalAction;
  tools?: string[];
  channels?: string[];
  autoApprove: boolean;
  maxCost?: number;
  createdAt: string;
  active: boolean;
}

interface PolicyApprovalsProps {
  onClose: () => void;
}

export default function PolicyApprovals(props: PolicyApprovalsProps) {
  const keyboard = useKeyboard();
  const [policies, setPolicies] = createSignal<PolicyRule[]>([]);
  const [requests, setRequests] = createSignal<ApprovalRequest[]>([]);
  const [view, setView] = createSignal<"policies" | "requests">("policies");
  const [selectedIndex, setSelectedIndex] = createSignal(0);
  const [showCreateModal, setShowCreateModal] = createSignal(false);
  const [newPolicyName, setNewPolicyName] = createSignal("");
  const [newActionType, setNewActionType] = createSignal<ActionType>("shell");
  const [newApprovalAction, setNewApprovalAction] = createSignal<ApprovalAction>("require_review");
  const [newAutoApprove, setNewAutoApprove] = createSignal(false);
  const [newMaxCost, setNewMaxCost] = createSignal("");
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");

  onMount(() => {
    loadPolicies();
    loadRequests();
    setupKeyboard();
    startPolling();
  });

  const setupKeyboard = () => {
    keyboard.bind("1", () => setView("policies"));
    keyboard.bind("2", () => setView("requests"));
    keyboard.bind("n", () => {
      setShowCreateModal(true);
      setNewPolicyName("");
      setNewMaxCost("");
    });
    keyboard.bind("e", () => editPolicy());
    keyboard.bind("d", () => deletePolicy());
    keyboard.bind("a", () => togglePolicy());
    keyboard.bind("r", () => reviewRequest());
    keyboard.bind("esc", () => {
      if (showCreateModal()) {
        setShowCreateModal(false);
      }
    });
    keyboard.bind("q", () => {
      props.onClose();
    });
  };

  const loadPolicies = async () => {
    setLoading(true);
    try {
      const response = await fetch("http://localhost:3000/api/policies");
      if (!response.ok) {
        throw new Error("Failed to load policies");
      }
      const data = await response.json();
      setPolicies(data.policies || []);
    } catch (err) {
      setError(`Failed to load policies: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const loadRequests = async () => {
    try {
      const response = await fetch("http://localhost:3000/api/approvals/pending");
      if (!response.ok) {
        throw new Error("Failed to load approval requests");
      }
      const data = await response.json();
      setRequests(data.requests || []);
    } catch (err) {
      setError(`Failed to load requests: ${err.message}`);
    }
  };

  const startPolling = () => {
    setInterval(() => {
      loadRequests();
    }, 5000);
  };

  const createPolicy = async () => {
    if (!newPolicyName()) {
      setError("Policy name is required");
      return;
    }

    const policy: Partial<PolicyRule> = {
      name: newPolicyName(),
      actionType: newActionType(),
      action: newApprovalAction(),
      autoApprove: newAutoApprove(),
      maxCost: newMaxCost() ? parseFloat(newMaxCost()) : undefined,
      active: true,
    };

    try {
      const response = await fetch("http://localhost:3000/api/policies", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(policy),
      });

      if (!response.ok) {
        throw new Error("Failed to create policy");
      }

      setShowCreateModal(false);
      loadPolicies();
    } catch (err) {
      setError(`Failed to create policy: ${err.message}`);
    }
  };

  const editPolicy = () => {
    const policy = policies()[selectedIndex()];
    if (!policy) return;

    console.log("Edit policy:", policy);
  };

  const deletePolicy = async () => {
    const policy = policies()[selectedIndex()];
    if (!policy) return;

    try {
      const response = await fetch(`http://localhost:3000/api/policies/${policy.id}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error("Failed to delete policy");
      }

      loadPolicies();
    } catch (err) {
      setError(`Failed to delete policy: ${err.message}`);
    }
  };

  const togglePolicy = async () => {
    const policy = policies()[selectedIndex()];
    if (!policy) return;

    try {
      const response = await fetch(`http://localhost:3000/api/policies/${policy.id}/toggle`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to toggle policy");
      }

      loadPolicies();
    } catch (err) {
      setError(`Failed to toggle policy: ${err.message}`);
    }
  };

  const reviewRequest = async () => {
    const request = requests()[selectedIndex()];
    if (!request) return;

    console.log("Review request:", request);
  };

  const approveRequest = async (requestId: string) => {
    try {
      const response = await fetch(`http://localhost:3000/api/approvals/${requestId}/approve`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to approve request");
      }

      loadRequests();
    } catch (err) {
      setError(`Failed to approve request: ${err.message}`);
    }
  };

  const denyRequest = async (requestId: string) => {
    try {
      const response = await fetch(`http://localhost:3000/api/approvals/${requestId}/deny`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to deny request");
      }

      loadRequests();
    } catch (err) {
      setError(`Failed to deny request: ${err.message}`);
    }
  };

  const getActionTypeLabel = (type: ActionType) => {
    switch (type) {
      case "file_ops":
        return "File Ops";
      case "shell":
        return "Shell";
      case "network":
        return "Network";
      case "channel_message":
        return "Channel Message";
      case "credential_access":
        return "Credential Access";
    }
  };

  const getApprovalActionLabel = (action: ApprovalAction) => {
    switch (action) {
      case "allow":
        return "Allow";
      case "deny":
        return "Deny";
      case "require_review":
        return "Require Review";
    }
  };

  const getRequestStatusLabel = (status: string) => {
    switch (status) {
      case "pending":
        return "Pending";
      case "approved":
        return "Approved";
      case "denied":
        return "Denied";
    }
  };

  const getRequestStatusColor = (status: string) => {
    switch (status) {
      case "pending":
        return palette.accent;
      case "approved":
        return palette.success;
      case "denied":
        return palette.error;
    }
  };

  return (
    <Box flexDirection="column" width="100%" height="100%">
      <Box
        flexDirection="row"
        padding={1}
        backgroundColor={palette.primary}
        color={palette.background}
      >
        <Text bold>🛡️ Policies & Approvals</Text>
        <Box flexGrow={1} />
        <Text>
          View: <Text bold>[1]</Text> Policies <Text bold>[2]</Text> Requests
        </Text>
        <Text>
          Quit: <Text bold>[Q]</Text>
        </Text>
      </Box>

      <Show when={error()}>
        <Box padding={1} backgroundColor={palette.error}>
          <Text color={palette.background}>{error()}</Text>
        </Box>
      </Show>

      <Show when={!loading()}>
        <Box flexDirection="column" padding={1} flexGrow={1}>
          <Show when={view() === "policies"}>
            <Box flexDirection="row" padding={1} backgroundColor={palette.bgSecondary}>
              <Box flexGrow={1}>
                <Text bold>Total Policies</Text>
                <Text fontSize={2}>{policies().length}</Text>
              </Box>
              <Box flexGrow={1}>
                <Text bold>Active</Text>
                <Text fontSize={2} color={palette.success}>
                  {policies().filter(p => p.active).length}
                </Text>
              </Box>
              <Box flexGrow={1}>
                <Text bold>Auto-Approve</Text>
                <Text fontSize={2} color={palette.accent}>
                  {policies().filter(p => p.autoApprove).length}
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
                <Text bold>Create New Policy</Text>
                <Box marginTop={1}>
                  <Text width={20}>Name:</Text>
                  <Box flexGrow={1}>
                    <TextInput
                      value={newPolicyName()}
                      onInput={(e: any) => setNewPolicyName(e.target.value)}
                      placeholder="Policy name"
                    />
                  </Box>
                </Box>
                <Box marginTop={1}>
                  <Text width={20}>Action Type:</Text>
                  <Box flexGrow={1}>
                    <Select
                      value={newActionType()}
                      onChange={(e: any) => setNewActionType(e.target.value)}
                    >
                      <option value="file_ops">File Ops</option>
                      <option value="shell">Shell</option>
                      <option value="network">Network</option>
                      <option value="channel_message">Channel Message</option>
                      <option value="credential_access">Credential Access</option>
                    </Select>
                  </Box>
                </Box>
                <Box marginTop={1}>
                  <Text width={20}>Approval:</Text>
                  <Box flexGrow={1}>
                    <Select
                      value={newApprovalAction()}
                      onChange={(e: any) => setNewApprovalAction(e.target.value)}
                    >
                      <option value="allow">Allow</option>
                      <option value="deny">Deny</option>
                      <option value="require_review">Require Review</option>
                    </Select>
                  </Box>
                </Box>
                <Box marginTop={1}>
                  <Text width={20}>Max Cost:</Text>
                  <Box flexGrow={1}>
                    <TextInput
                      value={newMaxCost()}
                      onInput={(e: any) => setNewMaxCost(e.target.value)}
                      placeholder="Optional max cost"
                    />
                  </Box>
                </Box>
                <Box flexDirection="row" marginTop={1}>
                  <Box flexGrow={1}>
                    <Button onClick={createPolicy}>Create</Button>
                  </Box>
                  <Box flexGrow={1}>
                    <Button onClick={() => setShowCreateModal(false)}>Cancel</Button>
                  </Box>
                </Box>
              </Box>
            </Show>

            <Box padding={1} marginTop={1} backgroundColor={palette.background}>
              <Text bold>Policy Rules</Text>
            </Box>

            <Box
              flexDirection="column"
              flexGrow={1}
              padding={1}
              backgroundColor={palette.background}
            >
              <For each={policies()}>
                {(policy, index) => (
                  <Box
                    flexDirection="row"
                    padding={0.5}
                    backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
                    onClick={() => setSelectedIndex(index())}
                  >
                    <Text width={30}>{policy.name}</Text>
                    <Text width={20}>{getActionTypeLabel(policy.actionType)}</Text>
                    <Text width={20}>{getApprovalActionLabel(policy.action)}</Text>
                    <Text width={15} color={policy.active ? palette.success : palette.dim}>
                      {policy.active ? "Active" : "Inactive"}
                    </Text>
                    <Text width={15}>{policy.autoApprove ? "Auto" : "Manual"}</Text>
                  </Box>
                )}
              </For>

              <Show when={policies().length === 0}>
                <Box padding={2} textAlign="center">
                  <Text color={palette.dim}>No policies configured. Press [N] to create one.</Text>
                </Box>
              </Show>
            </Box>

            <Box
              flexDirection="row"
              padding={1}
              marginTop={1}
              backgroundColor={palette.bgSecondary}
            >
              <Text>
                Create: <Text bold>[N]</Text>
              </Text>
              <Box flexGrow={1} />
              <Text>
                Edit: <Text bold>[E]</Text> Delete: <Text bold>[D]</Text> Toggle:{" "}
                <Text bold>[A]</Text>
              </Text>
            </Box>
          </Show>

          <Show when={view() === "requests"}>
            <Box flexDirection="row" padding={1} backgroundColor={palette.bgSecondary}>
              <Box flexGrow={1}>
                <Text bold>Pending Requests</Text>
                <Text fontSize={2}>{requests().filter(r => r.status === "pending").length}</Text>
              </Box>
              <Box flexGrow={1}>
                <Text bold>Approved Today</Text>
                <Text fontSize={2} color={palette.success}>
                  {requests().filter(r => r.status === "approved").length}
                </Text>
              </Box>
              <Box flexGrow={1}>
                <Text bold>Denied Today</Text>
                <Text fontSize={2} color={palette.error}>
                  {requests().filter(r => r.status === "denied").length}
                </Text>
              </Box>
            </Box>

            <Box padding={1} marginTop={1} backgroundColor={palette.background}>
              <Text bold>Approval Requests</Text>
            </Box>

            <Box
              flexDirection="column"
              flexGrow={1}
              padding={1}
              backgroundColor={palette.background}
            >
              <For each={requests()}>
                {(request, index) => (
                  <Box
                    flexDirection="row"
                    padding={0.5}
                    backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
                    onClick={() => setSelectedIndex(index())}
                  >
                    <Text width={25}>{request.agentName}</Text>
                    <Text width={20}>{getActionTypeLabel(request.actionType)}</Text>
                    <Text width={20}>{request.tool}</Text>
                    <Text width={15} color={getRequestStatusColor(request.status)}>
                      {getRequestStatusLabel(request.status)}
                    </Text>
                    <Text width={20}>{request.timestamp}</Text>
                    <Box flexGrow={1}>
                      <Show when={request.status === "pending"}>
                        <Button
                          style={{ padding: "0.25 0.5", marginRight: "0.5" }}
                          onClick={() => approveRequest(request.id)}
                        >
                          ✓
                        </Button>
                        <Button
                          style={{
                            padding: "0.25 0.5",
                            backgroundColor: palette.error,
                          }}
                          onClick={() => denyRequest(request.id)}
                        >
                          ✗
                        </Button>
                      </Show>
                    </Box>
                  </Box>
                )}
              </For>

              <Show when={requests().length === 0}>
                <Box padding={2} textAlign="center">
                  <Text color={palette.dim}>No pending requests.</Text>
                </Box>
              </Show>
            </Box>

            <Box
              flexDirection="row"
              padding={1}
              marginTop={1}
              backgroundColor={palette.bgSecondary}
            >
              <Text>
                Review: <Text bold>[R]</Text>
              </Text>
              <Box flexGrow={1} />
              <Text>
                Approve: <Text bold>✓</Text> Deny: <Text bold>✗</Text>
              </Text>
            </Box>
          </Show>
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
      backgroundColor: props.style?.backgroundColor || palette.primary,
      color: palette.background,
      border: "none",
      cursor: "pointer",
      ...props.style,
    }}
  >
    {props.children}
  </button>
);
