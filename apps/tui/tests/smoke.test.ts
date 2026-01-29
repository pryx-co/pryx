
import { describe, test, expect } from "bun:test";
import { spawn } from "bun";

describe("TUI Smoke Test", () => {
    test("App starts and renders without crashing", async () => {
        const proc = spawn({
            cmd: ["bun", "run", "tests/scripts/smoke-render.tsx"],
            // cwd is inferred as current (apps/tui)
            stdout: "ignore", // We just check exit code for smoke test
            stderr: "pipe",
            env: { ...process.env, FORCE_COLOR: "1" }
        });

        const exitCode = await new Promise((resolve) => {
            // Wait for exit
            const check = setInterval(() => {
                if (proc.exitCode !== null) {
                    clearInterval(check);
                    resolve(proc.exitCode);
                }
            }, 100);

            // Timeout safety
            setTimeout(() => {
                clearInterval(check);
                proc.kill();
                resolve("timeout");
            }, 5000);
        });

        if (exitCode !== 0) {
            const stderr = await new Response(proc.stderr).text();
            console.error("Smoke test failed. Stderr:", stderr);
        }

        expect(exitCode).toBe(0);
    }, 6000);
});
