import { createSignal, For, Show, onMount } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { palette } from "../theme";

type DeviceStatus = "online" | "offline" | "syncing" | "error";
type DeviceRole = "primary" | "secondary";

interface Device {
  id: string;
  name: string;
  status: DeviceStatus;
  role: DeviceRole;
  lastSeen: string;
  ipAddress?: string;
  platform: string;
  version: string;
  sessionCount: number;
}

interface SyncEvent {
  id: string;
  deviceId: string;
  deviceName: string;
  timestamp: string;
  type: "session_sync" | "config_sync" | "heartbeat";
  success: boolean;
  error?: string;
}

interface MeshStatusProps {
  onClose: () => void;
}

export default function MeshStatus(props: MeshStatusProps) {
  const keyboard = useKeyboard();
  const [devices, setDevices] = createSignal<Device[]>([]);
  const [events, setEvents] = createSignal<SyncEvent[]>([]);
  const [selectedIndex, setSelectedIndex] = createSignal(0);
  const [view, setView] = createSignal<"devices" | "events">("devices");
  const [showPairModal, setShowPairModal] = createSignal(false);
  const [pairingCode, setPairingCode] = createSignal("");
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");
  const [pairingStatus, setPairingStatus] = createSignal<"idle" | "pairing" | "success" | "failed">(
    "idle"
  );

  onMount(() => {
    loadDevices();
    loadEvents();
    setupKeyboard();
    startPolling();
  });

  const setupKeyboard = () => {
    keyboard.bind("1", () => setView("devices"));
    keyboard.bind("2", () => setView("events"));
    keyboard.bind("p", () => {
      setShowPairModal(true);
      setPairingCode("");
      setPairingStatus("idle");
    });
    keyboard.bind("r", () => refreshDevices());
    keyboard.bind("u", () => unpairDevice());
    keyboard.bind("v", () => viewDevice());
    keyboard.bind("s", () => syncDevice());
    keyboard.bind("q", () => {
      props.onClose();
    });
    keyboard.bind("esc", () => {
      if (showPairModal()) {
        setShowPairModal(false);
      }
    });
  };

  const loadDevices = async () => {
    setLoading(true);
    try {
      const response = await fetch("http://localhost:3000/api/mesh/devices");
      if (!response.ok) {
        throw new Error("Failed to load devices");
      }
      const data = await response.json();
      setDevices(data.devices || []);
    } catch (err) {
      setError(`Failed to load devices: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const loadEvents = async () => {
    try {
      const response = await fetch("http://localhost:3000/api/mesh/events");
      if (!response.ok) {
        throw new Error("Failed to load events");
      }
      const data = await response.json();
      setEvents(data.events || []);
    } catch (err) {
      setError(`Failed to load events: ${err.message}`);
    }
  };

  const startPolling = () => {
    setInterval(() => {
      loadDevices();
      loadEvents();
    }, 5000);
  };

  const refreshDevices = () => {
    loadDevices();
  };

  const startPairing = async () => {
    if (!pairingCode()) {
      setError("Pairing code is required");
      return;
    }

    setPairingStatus("pairing");

    try {
      const response = await fetch("http://localhost:3000/api/mesh/pair", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          code: pairingCode(),
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to pair device");
      }

      const data = await response.json();
      setPairingStatus("success");

      setTimeout(() => {
        setShowPairModal(false);
        loadDevices();
      }, 2000);
    } catch (err) {
      setError(`Failed to pair device: ${err.message}`);
      setPairingStatus("failed");
    }
  };

  const unpairDevice = async () => {
    const device = devices()[selectedIndex()];
    if (!device) return;

    try {
      const response = await fetch(`http://localhost:3000/api/mesh/devices/${device.id}/unpair`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to unpair device");
      }

      loadDevices();
    } catch (err) {
      setError(`Failed to unpair device: ${err.message}`);
    }
  };

  const syncDevice = async () => {
    const device = devices()[selectedIndex()];
    if (!device) return;

    try {
      const response = await fetch(`http://localhost:3000/api/mesh/devices/${device.id}/sync`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to sync device");
      }

      loadDevices();
    } catch (err) {
      setError(`Failed to sync device: ${err.message}`);
    }
  };

  const viewDevice = () => {
    const device = devices()[selectedIndex()];
    if (!device) return;

    console.log("View device:", device);
  };

  const getStatusColor = (status: DeviceStatus) => {
    switch (status) {
      case "online":
        return palette.success;
      case "offline":
        return palette.dim;
      case "syncing":
        return palette.accent;
      case "error":
        return palette.error;
    }
  };

  const getStatusLabel = (status: DeviceStatus) => {
    switch (status) {
      case "online":
        return "Online";
      case "offline":
        return "Offline";
      case "syncing":
        return "Syncing";
      case "error":
        return "Error";
    }
  };

  const getRoleLabel = (role: DeviceRole) => {
    switch (role) {
      case "primary":
        return "Primary";
      case "secondary":
        return "Secondary";
    }
  };

  const getEventTypeLabel = (type: string) => {
    switch (type) {
      case "session_sync":
        return "Session Sync";
      case "config_sync":
        return "Config Sync";
      case "heartbeat":
        return "Heartbeat";
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
        <Text bold>🔗 Mesh Status</Text>
        <Box flexGrow={1} />
        <Text>
          View: <Text bold>[1]</Text> Devices <Text bold>[2]</Text> Events
        </Text>
        <Text>
          Quit: <Text bold>[Q]</Text>
        </Text>
      </Box>

      <Show when={loading()}>
        <Box padding={2}>
          <Text>Loading mesh status...</Text>
        </Box>
      </Show>

      <Show when={error()}>
        <Box padding={1} backgroundColor={palette.error}>
          <Text color={palette.background}>{error()}</Text>
        </Box>
      </Show>

      <Show when={!loading() && !error()}>
        <Box flexDirection="column" padding={1} flexGrow={1}>
          <Show when={showPairModal()}>
            <Box
              flexDirection="column"
              padding={1}
              marginTop={1}
              backgroundColor={palette.bgSelected}
              border={`1px solid ${palette.border}`}
            >
              <Text bold>Pair New Device</Text>
              <Show when={pairingStatus() === "idle"}>
                <Box marginTop={1}>
                  <Text>Enter the 6-digit pairing code from the other device:</Text>
                  <Box marginTop={1}>
                    <TextInput
                      value={pairingCode()}
                      onInput={(e: any) => setPairingCode(e.target.value)}
                      placeholder="000000"
                      maxLength={6}
                    />
                  </Box>
                </Box>
                <Box flexDirection="row" marginTop={1}>
                  <Box flexGrow={1}>
                    <Button onClick={startPairing}>Pair</Button>
                  </Box>
                  <Box flexGrow={1}>
                    <Button onClick={() => setShowPairModal(false)}>Cancel</Button>
                  </Box>
                </Box>
              </Show>
              <Show when={pairingStatus() === "pairing"}>
                <Box marginTop={1} textAlign="center">
                  <Text color={palette.accent}>Pairing...</Text>
                </Box>
              </Show>
              <Show when={pairingStatus() === "success"}>
                <Box marginTop={1} textAlign="center">
                  <Text color={palette.success}>✓ Device Paired Successfully!</Text>
                </Box>
              </Show>
              <Show when={pairingStatus() === "failed"}>
                <Box marginTop={1} textAlign="center">
                  <Text color={palette.error}>✗ Pairing Failed</Text>
                </Box>
              </Show>
            </Box>
          </Show>

          <Show when={view() === "devices"}>
            <Box flexDirection="row" padding={1} backgroundColor={palette.bgSecondary}>
              <Box flexGrow={1}>
                <Text bold>Total Devices</Text>
                <Text fontSize={2}>{devices().length}</Text>
              </Box>
              <Box flexGrow={1}>
                <Text bold>Online</Text>
                <Text fontSize={2} color={palette.success}>
                  {devices().filter(d => d.status === "online").length}
                </Text>
              </Box>
              <Box flexGrow={1}>
                <Text bold>Offline</Text>
                <Text fontSize={2} color={palette.dim}>
                  {devices().filter(d => d.status === "offline").length}
                </Text>
              </Box>
            </Box>

            <Box padding={1} marginTop={1} backgroundColor={palette.background}>
              <Text bold>Mesh Devices</Text>
            </Box>

            <Box
              flexDirection="column"
              flexGrow={1}
              padding={1}
              backgroundColor={palette.background}
            >
              <For each={devices()}>
                {(device, index) => (
                  <Box
                    flexDirection="row"
                    padding={0.5}
                    backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
                    onClick={() => setSelectedIndex(index())}
                  >
                    <Text width={25}>{device.name}</Text>
                    <Text width={15} color={getStatusColor(device.status)}>
                      {getStatusLabel(device.status)}
                    </Text>
                    <Text width={15}>{getRoleLabel(device.role)}</Text>
                    <Text width={15}>{device.platform}</Text>
                    <Text width={15}>{device.version}</Text>
                    <Text width={20}>{device.lastSeen}</Text>
                  </Box>
                )}
              </For>

              <Show when={devices().length === 0}>
                <Box padding={2} textAlign="center">
                  <Text color={palette.dim}>No devices paired. Press [P] to pair a device.</Text>
                </Box>
              </Show>

              <Show when={devices().length === 1}>
                <Box padding={2} textAlign="center">
                  <Text color={palette.dim}>
                    This is your primary device. Pair other devices to enable mesh sync.
                  </Text>
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
                Pair: <Text bold>[P]</Text>
              </Text>
              <Box flexGrow={1} />
              <Text>
                Refresh: <Text bold>[R]</Text> Unpair: <Text bold>[U]</Text> Sync:{" "}
                <Text bold>[S]</Text> View: <Text bold>[V]</Text>
              </Text>
            </Box>
          </Show>

          <Show when={view() === "events"}>
            <Box flexDirection="row" padding={1} backgroundColor={palette.bgSecondary}>
              <Box flexGrow={1}>
                <Text bold>Total Events</Text>
                <Text fontSize={2}>{events().length}</Text>
              </Box>
              <Box flexGrow={1}>
                <Text bold>Successful</Text>
                <Text fontSize={2} color={palette.success}>
                  {events().filter(e => e.success).length}
                </Text>
              </Box>
              <Box flexGrow={1}>
                <Text bold>Failed</Text>
                <Text fontSize={2} color={palette.error}>
                  {events().filter(e => !e.success).length}
                </Text>
              </Box>
            </Box>

            <Box padding={1} marginTop={1} backgroundColor={palette.background}>
              <Text bold>Sync Events</Text>
            </Box>

            <Box
              flexDirection="column"
              flexGrow={1}
              padding={1}
              backgroundColor={palette.background}
            >
              <For each={events()}>
                {(event, index) => (
                  <Box
                    flexDirection="row"
                    padding={0.5}
                    backgroundColor={index() === selectedIndex() ? palette.bgSelected : undefined}
                    onClick={() => setSelectedIndex(index())}
                  >
                    <Text width={25}>{event.deviceName}</Text>
                    <Text width={20}>{getEventTypeLabel(event.type)}</Text>
                    <Text width={15} color={event.success ? palette.success : palette.error}>
                      {event.success ? "✓" : "✗"}
                    </Text>
                    <Text width={20}>{event.timestamp}</Text>
                  </Box>
                )}
              </For>

              <Show when={events().length === 0}>
                <Box padding={2} textAlign="center">
                  <Text color={palette.dim}>No sync events yet.</Text>
                </Box>
              </Show>
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
    type="text"
    value={props.value}
    onInput={props.onInput}
    placeholder={props.placeholder}
    maxLength={props.maxLength}
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
