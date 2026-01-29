// @ts-nocheck
import { Box, Text } from "@opentui/core";
import { createSignal, onMount, For, Show } from "solid-js";

interface Skill {
    id: string;
    name: string;
    description: string;
    enabled?: boolean;
    installed?: boolean;
}

export default function Skills() {
    const [skills, setSkills] = createSignal<Skill[]>([]);
    const [selectedIndex, setSelectedIndex] = createSignal(0);
    const [loading, setLoading] = createSignal(true);
    const [error, setError] = createSignal("");
    const [detailView, setDetailView] = createSignal(false);

    const fetchSkills = async () => {
        try {
            setLoading(true);
            const apiUrl = process.env.PRYX_API_URL || "http://localhost:3000";
            const res = await fetch(`${apiUrl}/skills`);
            const data = await res.json();
            setSkills(data.skills || []);
            setLoading(false);
        } catch (e) {
            setError("Failed to load skills");
            setLoading(false);
        }
    };


    onMount(() => {
        fetchSkills();
    });

    const selectedSkill = () => {
        const index = selectedIndex();
        const skillsList = skills();
        if (skillsList.length === 0) return null;
        return skillsList[index] || skillsList[0];
    };

    return (
        <Box flexDirection="column" flexGrow={1}>
            <Text bold color="magenta">Skills Manager</Text>
            <Text color="gray">Extend agent capabilities with skills</Text>

            <Show when={loading()}>
                <Box marginTop={1}>
                    <Text color="yellow">Loading skills...</Text>
                </Box>
            </Show>

            <Show when={error()}>
                <Box marginTop={1}>
                    <Text color="red">{error()}</Text>
                </Box>
            </Show>

            <Show when={!loading() && !error()}>
                <Show when={!detailView()} fallback={
                    <Box marginTop={1} flexDirection="column" borderStyle="round" padding={1}>
                        <Text bold color="cyan">{selectedSkill()?.name}</Text>
                        <Box marginTop={1}>
                            <Text color="gray">ID: </Text>
                            <Text>{selectedSkill()?.id}</Text>
                        </Box>
                        <Box marginTop={1}>
                            <Text color="gray">Description:</Text>
                        </Box>
                        <Box marginTop={0}>
                            <Text>{selectedSkill()?.description || "No description available"}</Text>
                        </Box>
                        <Box marginTop={1}>
                            <Text color="gray">Status: </Text>
                            <Text color={selectedSkill()?.enabled ? "green" : "gray"}>
                                {selectedSkill()?.enabled ? "ENABLED" : "DISABLED"}
                            </Text>
                        </Box>
                        <Box marginTop={2}>
                            <Text color="gray">Press Esc to go back</Text>
                        </Box>
                    </Box>
                }>
                    <Box marginTop={1} flexDirection="column" borderStyle="round" padding={1}>
                        <Show when={skills().length === 0}>
                            <Text color="gray">No skills available</Text>
                        </Show>
                        <For each={skills()}>
                            {(skill, index) => {
                                const isSelected = index() === selectedIndex();
                                return (
                                    <Box flexDirection="row" marginBottom={0}>
                                        <Text color={isSelected ? "cyan" : "gray"}>
                                            {isSelected ? "❯ " : "  "}
                                        </Text>
                                        <Box width={30}>
                                            <Text bold={isSelected}>
                                                {skill.name}
                                            </Text>
                                        </Box>
                                        <Text color={skill.enabled ? "green" : "gray"}>
                                            {skill.enabled ? "✓" : "○"}
                                        </Text>
                                    </Box>
                                );
                            }}
                        </For>
                    </Box>
                </Show>
            </Show>

            <Box marginTop={1}>
                <Text color="gray">
                    {detailView()
                        ? "Esc: Back"
                        : "↑↓ Navigate │ Enter: Details │ R: Refresh"}
                </Text>
            </Box>
        </Box>
    );
}
