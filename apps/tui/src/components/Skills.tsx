// @ts-nocheck
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
            setError("");
            const apiUrl = process.env.PRYX_API_URL || "http://localhost:3000";
            const res = await fetch(`${apiUrl}/skills`);
            if (!res.ok) {
                throw new Error(`HTTP ${res.status}`);
            }
            const data = await res.json();
            setSkills(data.skills || []);
        } catch (e) {
            setError("Failed to load skills");
        } finally {
            setLoading(false);
        }
    };

    onMount(() => {
        fetchSkills().catch(() => {});
    });

    const selectedSkill = () => {
        const index = selectedIndex();
        const skillsList = skills();
        if (skillsList.length === 0) return null;
        return skillsList[index] || skillsList[0];
    };

    return (
        <box flexDirection="column" flexGrow={1}>
            <text fg="magenta">Skills Manager</text>
            <text fg="gray">Extend agent capabilities with skills</text>

            <Show when={loading()}>
                <box marginTop={1}>
                    <text fg="yellow">Loading skills...</text>
                </box>
            </Show>

            <Show when={error()}>
                <box marginTop={1}>
                    <text fg="red">{error()}</text>
                </box>
            </Show>

            <Show when={!loading() && !error()}>
                <Show when={!detailView()} fallback={
                    <box marginTop={1} flexDirection="column" borderStyle="round" padding={1}>
                        <text fg="cyan">{selectedSkill()?.name}</text>
                        <box marginTop={1}>
                            <text fg="gray">ID: </text>
                            <text>{selectedSkill()?.id}</text>
                        </box>
                        <box marginTop={1}>
                            <text fg="gray">Description:</text>
                        </box>
                        <box marginTop={0}>
                            <text>{selectedSkill()?.description || "No description available"}</text>
                        </box>
                        <box marginTop={1}>
                            <text fg="gray">Status: </text>
                            <text fg={selectedSkill()?.enabled ? "green" : "gray"}>
                                {selectedSkill()?.enabled ? "ENABLED" : "DISABLED"}
                            </text>
                        </box>
                        <box marginTop={2}>
                            <text fg="gray">Press Esc to go back</text>
                        </box>
                    </box>
                }>
                    <box marginTop={1} flexDirection="column" borderStyle="round" padding={1}>
                        <Show when={skills().length === 0}>
                            <text fg="gray">No skills available</text>
                        </Show>
                        <For each={skills()}>
                            {(skill, index) => {
                                const isSelected = index() === selectedIndex();
                                return (
                                    <box flexDirection="row" marginBottom={0}>
                                        <text fg={isSelected ? "cyan" : "gray"}>
                                            {isSelected ? "❯ " : "  "}
                                        </text>
                                        <box width={30}>
                                            <text>{skill.name}</text>
                                        </box>
                                        <text fg={skill.enabled ? "green" : "gray"}>
                                            {skill.enabled ? "✓" : "○"}
                                        </text>
                                    </box>
                                );
                            }}
                        </For>
                    </box>
                </Show>
            </Show>

            <box marginTop={1}>
                <text fg="gray">
                    {detailView()
                        ? "Esc: Back"
                        : "↑↓ Navigate │ Enter: Details │ R: Refresh"}
                </text>
            </box>
        </box>
    );
}
