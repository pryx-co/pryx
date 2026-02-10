import { describe, expect, it } from 'vitest';
import { apiApp } from './[...path]';

class FakeKvNamespace {
    store = new Map<string, string>();
    ttls = new Map<string, number>();

    async put(key: string, value: string, options?: { expirationTtl?: number }) {
        this.store.set(key, value);
        if (options?.expirationTtl) {
            this.ttls.set(key, options.expirationTtl);
        }
    }

    async get(key: string) {
        return this.store.get(key) ?? null;
    }

    async list(options?: { prefix?: string; limit?: number }) {
        const prefix = options?.prefix ?? '';
        const limit = options?.limit ?? 1000;
        const keys = Array.from(this.store.keys())
            .filter((key) => key.startsWith(prefix))
            .slice(0, limit)
            .map((name) => ({ name }));
        return { keys };
    }
}

function createEnv() {
    return {
        TELEMETRY: new FakeKvNamespace(),
        ADMIN_API_KEY: 'telemetry-secret',
    };
}

describe('telemetry persistence pipeline', () => {
    it('persists sanitized events with retention TTL', async () => {
        const env = createEnv();
        const response = await apiApp.request(
            'http://localhost/api/telemetry/ingest',
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: 'Bearer telemetry-secret',
                },
                body: JSON.stringify({
                    correlation_id: 'corr-1',
                    level: 'error',
                    error_message: 'contact admin@example.com using api_key sk-abcdefghijklmnopqrstuv',
                    session_id: 'sess-1',
                }),
            },
            env as never,
        );

        expect(response.status).toBe(200);
        const data = await response.json();
        expect(data.accepted).toBe(1);
        expect(data.retention_seconds).toBe(604800);

        const persistedKey = data.results[0].key as string;
        const persistedRaw = await env.TELEMETRY.get(persistedKey);
        expect(persistedRaw).toBeTruthy();
        expect(env.TELEMETRY.ttls.get(persistedKey)).toBe(604800);

        const persisted = JSON.parse(persistedRaw as string);
        expect(persisted.error_message).toContain('[REDACTED]');
        expect(String(persisted.error_message)).not.toContain('admin@example.com');
    });

    it('returns filtered persisted events from query endpoint', async () => {
        const env = createEnv();

        await apiApp.request(
            'http://localhost/api/telemetry/ingest',
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: 'Bearer telemetry-secret',
                },
                body: JSON.stringify([
                    { correlation_id: 'corr-2', level: 'info', category: 'runtime', session_id: 'a' },
                    { correlation_id: 'corr-3', level: 'error', category: 'runtime', session_id: 'b' },
                ]),
            },
            env as never,
        );

        const queryResponse = await apiApp.request(
            'http://localhost/api/telemetry/query?level=error&limit=10',
            {
                headers: {
                    Authorization: 'Bearer telemetry-secret',
                },
            },
            env as never,
        );

        expect(queryResponse.status).toBe(200);
        const queryData = await queryResponse.json();
        expect(queryData.count).toBe(1);
        expect(queryData.events[0].correlation_id).toBe('corr-3');
    });
});
