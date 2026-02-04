import { Effect, Context, Layer } from "effect";
import { readFileSync, existsSync } from "node:fs";
import { join } from "node:path";
import { homedir } from "node:os";

const getHostHost = (): string => {
  return process.env.PRYX_HOST || "localhost";
};

const getDefaultHostPort = (): string => {
  return process.env.PRYX_HOST_PORT || "42424";
};

export function getRuntimeHttpUrl(): string {
  if (process.env.PRYX_API_URL) return process.env.PRYX_API_URL;
  const host = getHostHost();
  try {
    const port = readFileSync(join(homedir(), ".pryx", "runtime.port"), "utf-8").trim();
    return `http://${host}:${port}`;
  } catch {
    return `http://${host}:${getDefaultHostPort()}`;
  }
}

export function describeRuntimeConnectionFailure(): string | null {
  const url = getRuntimeHttpUrl();
  const portFile = join(homedir(), ".pryx", "runtime.port");

  if (!process.env.PRYX_API_URL && !existsSync(portFile)) {
    return "Pryx host not running. Start it with `pryx` (desktop mode) or `pryx-core` (headless mode).";
  }

  return `Pryx host not reachable at ${url}. Start it with \`pryx\` or \`pryx-core\`.`;
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
  constructor(
    readonly message: string,
    readonly cause?: unknown
  ) {}
}

export interface SkillsService {
  readonly fetchSkills: Effect.Effect<Skill[], SkillsFetchError>;
  readonly toggleSkill: (
    skillId: string,
    enabled: boolean
  ) => Effect.Effect<void, SkillsFetchError>;
  readonly installSkill: (skillId: string) => Effect.Effect<void, SkillsFetchError>;
  readonly uninstallSkill: (skillId: string) => Effect.Effect<void, SkillsFetchError>;
}

export const SkillsService = Context.GenericTag<SkillsService>("@pryx/tui/SkillsService");

const makeSkillsService = Effect.sync(() => {
  const fetchSkills = Effect.gen(function* () {
    const result = yield* Effect.tryPromise({
      try: async () => {
        const res = await fetch(`${getRuntimeHttpUrl()}/skills`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = (await res.json()) as SkillsResponse;
        return data.skills || [];
      },
      catch: error => new SkillsFetchError("Failed to fetch skills", error),
    });
    return result;
  });

  const toggleSkill = (skillId: string, enabled: boolean) =>
    Effect.gen(function* () {
      yield* Effect.tryPromise({
        try: async () => {
          const endpoint = enabled ? "/skills/enable" : "/skills/disable";
          const res = await fetch(`${getRuntimeHttpUrl()}${endpoint}`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id: skillId }),
          });
          if (!res.ok) throw new Error(`HTTP ${res.status}`);
        },
        catch: error =>
          new SkillsFetchError(`Failed to ${enabled ? "enable" : "disable"} skill`, error),
      });
    });

  const installSkill = (skillId: string) =>
    Effect.gen(function* () {
      yield* Effect.tryPromise({
        try: async () => {
          const res = await fetch(`${getRuntimeHttpUrl()}/skills/install`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id: skillId }),
          });
          if (!res.ok) throw new Error(`HTTP ${res.status}`);
        },
        catch: error => new SkillsFetchError("Failed to install skill", error),
      });
    });

  const uninstallSkill = (skillId: string) =>
    Effect.gen(function* () {
      yield* Effect.tryPromise({
        try: async () => {
          const res = await fetch(`${getRuntimeHttpUrl()}/skills/uninstall`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id: skillId }),
          });
          if (!res.ok) throw new Error(`HTTP ${res.status}`);
        },
        catch: error => new SkillsFetchError("Failed to uninstall skill", error),
      });
    });

  return {
    fetchSkills,
    toggleSkill,
    installSkill,
    uninstallSkill,
  } as SkillsService;
});

export const SkillsServiceLive = Layer.effect(SkillsService, makeSkillsService);
