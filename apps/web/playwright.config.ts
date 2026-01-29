import { defineConfig } from '@playwright/test';

export default defineConfig({
    testDir: './e2e',
    fullyParallel: true,
    webServer: {
        command: 'bun run dev -- --port 4321 --host 127.0.0.1',
        port: 4321,
        reuseExistingServer: !process.env.CI,
        timeout: 120_000,
    },
    use: {
        baseURL: 'http://127.0.0.1:4321',
    },
});

