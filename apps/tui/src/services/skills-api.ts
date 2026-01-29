import { Effect, Context, Layer } from "effect";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { homedir } from "node:os";

function getApiUrl(): string {
    if (process.env.PRYX_API_URL) return process.env.PRYX_API_URL;
    try {
        const port = readFileSync(join(homedir(), ".pryx", "runtime.port"), "utf-8").trim();
        return `http://localhost:${port}`;
    } catch {
        return "http://localhost:3000";
    }
}

export interface Skill {
    id: string;
    name: string;
    description: string;
    enabled?: boolean;
    installed?: boolean;
    eligible?: boolean;
    source?: string;
}

export interface SkillsResponse {
    skills: Skill[];
}

export class SkillsFetchError {
    readonly _tag = "SkillsFetchError";
    constructor(readonly message: string, readonly cause?: unknown) {}
}

export interface SkillsService {
    readonly fetchSkills: Effect.Effect<Skill[], SkillsFetchError>;
    readonly toggleSkill: (skillId: string, enabled: boolean) => Effect.Effect<void, SkillsFetchError>;
    readonly installSkill: (skillId: string) => Effect.Effect<void, SkillsFetchError>;
    readonly uninstallSkill: (skillId: string) => Effect.Effect<void, SkillsFetchError>;
}

export const SkillsService = Context.GenericTag<SkillsService>("@pryx/tui/SkillsService");

const makeSkillsService = Effect.gen(function* () {
    const fetchSkills = Effect.gen(function* () {
        const result = yield* Effect.tryPromise({
            try: async () => {
                const res = await fetch(`${getApiUrl()}/skills`);
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                const data = await res.json() as SkillsResponse;
                return data.skills || [];
            },
            catch: (error) => new SkillsFetchError("Failed to fetch skills", error)
        });
        return result;
    });

    const toggleSkill = (skillId: string, enabled: boolean) => Effect.gen(function* () {
        yield* Effect.tryPromise({
            try: async () => {
                const endpoint = enabled ? "/skills/enable" : "/skills/disable";
                const res = await fetch(`${getApiUrl()}${endpoint}`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ id: skillId })
                });
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
            },
            catch: (error) => new SkillsFetchError(`Failed to ${enabled ? "enable" : "disable"} skill`, error)
        });
    });

    const installSkill = (skillId: string) => Effect.gen(function* () {
        yield* Effect.tryPromise({
            try: async () => {
                const res = await fetch(`${getApiUrl()}/skills/install`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ id: skillId })
                });
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
            },
            catch: (error) => new SkillsFetchError("Failed to install skill", error)
        });
    });

    const uninstallSkill = (skillId: string) => Effect.gen(function* () {
        yield* Effect.tryPromise({
            try: async () => {
                const res = await fetch(`${getApiUrl()}/skills/uninstall`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ id: skillId })
                });
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
            },
            catch: (error) => new SkillsFetchError("Failed to uninstall skill", error)
        });
    });

    return {
        fetchSkills,
        toggleSkill,
        installSkill,
        uninstallSkill
    } as SkillsService;
});

export const SkillsServiceLive = Layer.effect(SkillsService, makeSkillsService);
