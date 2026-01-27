import { Box, Text } from "@opentui/core";
import { createSignal, onMount } from "solid-js";
import Chat from "./Chat";
import { connect } from "../services/ws";

export default function App() {
    const [status, setStatus] = createSignal("Connecting...");

    onMount(() => {
        connect((s) => setStatus(s));
    });

    return (
        <Box flexDirection="column" padding={1} borderStyle="single">
            <Box flexDirection="row" justifyContent="space-between" marginBottom={1}>
                <Text bold>Pryx TUI</Text>
                <Text color="green">{status()}</Text>
            </Box>
            <Chat />
        </Box>
    );
}
