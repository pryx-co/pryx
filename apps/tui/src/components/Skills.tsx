// @ts-nocheck
import { createSignal, onMount, For, Show } from "solid-js";
import { Effect } from "effect";
import { useEffectService, AppRuntime } from "../lib/hooks";
import { SkillsService } from "../services/skills-api";

export default function Skills() {
  const skillsService = useEffectService(SkillsService);
  const [skills, setSkills] = createSignal<any[]>([]);
  const [selectedIndex] = createSignal(0);
  const [loading, setLoading] = createSignal(true);
  const [error, setError] = createSignal("");
  const [detailView] = createSignal(false);

  onMount(() => {
    const service = skillsService();
    if (!service) return;

    AppRuntime.runFork(
      service.fetchSkills.pipe(
        Effect.tap(skills => Effect.sync(() => setSkills(skills))),
        Effect.catchAll(err => Effect.sync(() => {
          setError(err.message || "Failed to load skills");
          setLoading(false);
        }))
      )
    );
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
        <Show
          when={!detailView()}
          fallback={
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
          }
        >
          <box marginTop={1} flexDirection="column" borderStyle="round" padding={1}>
            <Show when={skills().length === 0}>
              <text fg="gray">No skills available</text>
            </Show>
            <For each={skills()}>
              {(skill, index) => {
                const isSelected = index() === selectedIndex();
                return (
                  <box flexDirection="row" marginBottom={0}>
                    <text fg={isSelected ? "cyan" : "gray"}>{isSelected ? "❯ " : "  "}</text>
                    <box width={30}>
                      <text>{skill.name}</text>
                    </box>
                    <text fg={skill.enabled ? "green" : "gray"}>{skill.enabled ? "✓" : "○"}</text>
                  </box>
                );
              }}
            </For>
          </box>
        </Show>
      </Show>

      <box marginTop={1}>
        <text fg="gray">
          {detailView() ? "Esc: Back" : "↑↓ Navigate │ Enter: Details │ R: Refresh"}
        </text>
      </box>
    </box>
  );
}
