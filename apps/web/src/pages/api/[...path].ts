import { Hono } from 'hono';
import type { APIRoute } from 'astro';
import adminApi from '../../server/admin-api';

interface TelemetryAuthEnv {
    ADMIN_API_KEY?: string;
    TELEMETRY_API_KEY?: string;
}

/**
 * Unified API route for Pryx Cloud using Hono
 * Ported from vanilla Response logic for better scalability
 */

export const apiApp = new Hono<{ Bindings: any }>().basePath('/api');

apiApp.route('/admin', adminApi);

// --- Constants & Utilities ---
const ALLOWED_TELEMETRY_FIELDS = new Set([
    'correlation_id', 'timestamp', 'level', 'category', 'error_code', 'error_message',
    'duration_ms', 'model_id', 'token_count', 'cost_usd', 'tool_name', 'status',
    'device_id', 'session_id', 'version',
]);

const PII_PATTERNS = [
    /\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b/g,
    /\b(?:sk|pk|api)[-_][a-zA-Z0-9]{20,}\b/g,
    /\b[0-9]{4}[- ]?[0-9]{4}[- ]?[0-9]{4}[- ]?[0-9]{4}\b/g,
    /\b[0-9]{3}[-.]?[0-9]{3}[-.]?[0-9]{4}\b/g,
];

const TELEMETRY_RETENTION_SECONDS = 7 * 24 * 60 * 60;

function generateCode(length: number, charset: string): string {
    const array = new Uint8Array(length);
    crypto.getRandomValues(array);
    return Array.from(array, b => charset[b % charset.length]).join('');
}

function redactPII(value: string): string {
    let result = value;
    for (const pattern of PII_PATTERNS) {
        result = result.replace(pattern, '[REDACTED]');
    }
    return result;
}

type TelemetryEvent = Record<string, string | number | boolean | null>;

function sanitizeTelemetryEvent(event: Record<string, unknown>): TelemetryEvent | null {
    const clean: TelemetryEvent = {};

    for (const [key, value] of Object.entries(event)) {
        if (!ALLOWED_TELEMETRY_FIELDS.has(key)) continue;

        if (value === null || typeof value === 'number' || typeof value === 'boolean') {
            clean[key] = value;
            continue;
        }

        if (typeof value === 'string') {
            clean[key] = redactPII(value);
        }
    }

    if (!clean.correlation_id || typeof clean.correlation_id !== 'string') {
        return null;
    }

    clean.received_at = Date.now();
    return clean;
}

function getBearerToken(headerValue: string | undefined): string | null {
    if (!headerValue) return null;
    if (!headerValue.startsWith('Bearer ')) return null;
    const token = headerValue.slice('Bearer '.length).trim();
    return token || null;
}

function ensureTelemetryAuthorization(c: any): Response | null {
    const env = (c.env ?? {}) as TelemetryAuthEnv;
    const expectedToken = env.TELEMETRY_API_KEY || env.ADMIN_API_KEY;

    if (!expectedToken) {
        return c.json({ error: 'telemetry_auth_unconfigured' }, 503);
    }

    const token = getBearerToken(c.req.header('Authorization'));
    if (!token) {
        return c.json({ error: 'unauthorized' }, 401);
    }

    if (token !== expectedToken) {
        return c.json({ error: 'forbidden' }, 403);
    }

    return null;
}

function parseTelemetryKeyTimestamp(keyName: string): number | null {
    if (!keyName.startsWith('telemetry:')) return null;
    const [, rawTs] = keyName.split(':', 3);
    const parsed = Number(rawTs);
    if (!Number.isFinite(parsed) || parsed <= 0) return null;
    return parsed;
}

// --- Middleware ---
apiApp.use('*', async (c, next) => {
    const ip = c.req.header('CF-Connecting-IP') || 'unknown';
    const env = c.env;

    if (env?.RATE_LIMITER) {
        const { success } = await env.RATE_LIMITER.limit(ip);
        if (!success) {
            return c.json({ error: 'slow_down' }, 429);
        }
    }
    await next();
});

// --- Auth Routes ---
apiApp.post('/auth/qr/pairing', async (c) => {
    const body = await c.req.json().catch(() => ({}));
    const deviceId = body.device_id || '';

    // Generate a short 8-char pairing code
    const pairingCode = generateCode(8, 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789');
    const pairingToken = generateCode(40, 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789');

    const entry = {
        deviceId,
        pairingCode,
        status: 'pending',
        created_at: Date.now(),
    };

    // Store with 5-minute TTL
    await c.env.DEVICE_CODES.put(`qr:${pairingCode}`, JSON.stringify(entry), { expirationTtl: 300 });
    await c.env.DEVICE_CODES.put(`ptr:${pairingToken}`, pairingCode, { expirationTtl: 300 });

    return c.json({
        pairing_code: pairingCode,
        pairing_token: pairingToken,
        expires_in: 300,
    });
});

apiApp.get('/auth/qr/status', async (c) => {
    const token = c.req.query('token');
    if (!token) return c.json({ error: 'missing_token' }, 400);

    const pairingCode = await c.env.DEVICE_CODES.get(`ptr:${token}`);
    if (!pairingCode) return c.json({ error: 'expired_token' }, 400);

    const entryStr = await c.env.DEVICE_CODES.get(`qr:${pairingCode}`);
    if (!entryStr) return c.json({ error: 'expired_pairing' }, 400);

    const entry = JSON.parse(entryStr);
    return c.json(entry);
});

apiApp.post('/auth/device/code', async (c) => {
    const body = await c.req.formData().catch(() => new FormData());
    const deviceId = body.get('device_id')?.toString() || '';
    const scopeStr = body.get('scope')?.toString() || 'telemetry.write';

    const deviceCode = generateCode(40, 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789');
    const userCode = `${generateCode(4, 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789')}-${generateCode(4, 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789')}`;

    const entry = {
        user_code: userCode,
        device_id: deviceId,
        scopes: scopeStr.split(' '),
        created_at: Date.now(),
        expires_at: Date.now() + 600 * 1000,
        authorized: false,
    };

    await c.env.DEVICE_CODES.put(deviceCode, JSON.stringify(entry), { expirationTtl: 660 });
    await c.env.DEVICE_CODES.put(`user:${userCode}`, deviceCode, { expirationTtl: 660 });

    return c.json({
        device_code: deviceCode,
        user_code: userCode,
        verification_uri: '/link',
        verification_uri_complete: `/link?code=${userCode}`,
        expires_in: 600,
        interval: 5,
    });
});

apiApp.post('/auth/device/token', async (c) => {
    const body = await c.req.formData().catch(() => new FormData());
    const deviceCode = body.get('device_code')?.toString() || '';

    const entryStr = await c.env.DEVICE_CODES.get(deviceCode);
    if (!entryStr) return c.json({ error: 'expired_token' }, 400);

    const entry = JSON.parse(entryStr);
    if (!entry.authorized) return c.json({ error: 'authorization_pending' }, 400);

    const accessToken = `pryx_at_${generateCode(40, 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789')}`;
    const refreshToken = `pryx_rt_${generateCode(40, 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789')}`;

    await c.env.TOKENS.put(accessToken, JSON.stringify(entry), { expirationTtl: 3600 });
    await c.env.TOKENS.put(`refresh:${refreshToken}`, JSON.stringify(entry), { expirationTtl: 86400 * 30 });

    return c.json({
        access_token: accessToken,
        token_type: 'Bearer',
        expires_in: 3600,
        refresh_token: refreshToken,
        scope: entry.scopes.join(' '),
    });
});

apiApp.post('/auth/token/refresh', async (c) => {
    const body = await c.req.formData().catch(() => new FormData());
    const refreshToken = body.get('refresh_token')?.toString() || '';

    const entryStr = await c.env.TOKENS.get(`refresh:${refreshToken}`);
    if (!entryStr) return c.json({ error: 'invalid_grant' }, 400);

    const entry = JSON.parse(entryStr);
    const newAccessToken = `pryx_at_${generateCode(40, 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789')}`;

    await c.env.TOKENS.put(newAccessToken, JSON.stringify(entry), { expirationTtl: 3600 });

    return c.json({
        access_token: newAccessToken,
        token_type: 'Bearer',
        expires_in: 3600,
        refresh_token: refreshToken,
        scope: entry.scopes.join(' '),
    });
});

// --- Telemetry Routes ---
apiApp.post('/telemetry/ingest', async (c) => {
    const authFailure = ensureTelemetryAuthorization(c);
    if (authFailure) return authFailure;

    try {
        const body = await c.req.json();
        const events = Array.isArray(body) ? body : [body];

        if (!c.env?.TELEMETRY) {
            return c.json({ error: 'telemetry_store_unavailable' }, 503);
        }

        const accepted: Array<{ key: string }> = [];
        for (const rawEvent of events) {
            if (!rawEvent || typeof rawEvent !== 'object') {
                continue;
            }

            const sanitized = sanitizeTelemetryEvent(rawEvent as Record<string, unknown>);
            if (!sanitized) {
                continue;
            }

            const key = `telemetry:${sanitized.received_at}:${generateCode(8, 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789')}`;
            await c.env.TELEMETRY.put(key, JSON.stringify(sanitized), { expirationTtl: TELEMETRY_RETENTION_SECONDS });
            accepted.push({ key });
        }

        return c.json({
            accepted: accepted.length,
            total: events.length,
            retention_seconds: TELEMETRY_RETENTION_SECONDS,
            timestamp: Date.now(),
            results: accepted.length <= 25 ? accepted : undefined,
        });
    } catch (e) {
        return c.json({ error: 'Invalid JSON' }, 400);
    }
});

apiApp.get('/telemetry/query', async (c) => {
    const authFailure = ensureTelemetryAuthorization(c);
    if (authFailure) return authFailure;

    if (!c.env?.TELEMETRY) {
        return c.json({ error: 'telemetry_store_unavailable' }, 503);
    }

    const limit = Math.min(parseInt(c.req.query('limit') || '100', 10), 1000);
    const level = c.req.query('level');
    const category = c.req.query('category');
    const deviceId = c.req.query('device_id');
    const sessionId = c.req.query('session_id');
    const start = parseInt(c.req.query('start') || '0', 10) || 0;
    const end = parseInt(c.req.query('end') || `${Date.now()}`, 10) || Date.now();

    const listed = await c.env.TELEMETRY.list({ prefix: 'telemetry:', limit: 1000 });
    const events: TelemetryEvent[] = [];

    const candidateKeys = (listed.keys as Array<{ name: string }>)
        .map((key: { name: string }) => ({ keyName: key.name, timestamp: parseTelemetryKeyTimestamp(key.name) }))
        .filter((entry: { keyName: string; timestamp: number | null }): entry is { keyName: string; timestamp: number } => entry.timestamp !== null)
        .filter((entry: { keyName: string; timestamp: number }) => entry.timestamp >= start && entry.timestamp <= end)
        .sort((a, b) => b.timestamp - a.timestamp);

    for (const candidate of candidateKeys) {
        if (events.length >= limit && !level && !category && !deviceId && !sessionId) {
            break;
        }

        const raw = await c.env.TELEMETRY.get(candidate.keyName);
        if (!raw) continue;

        try {
            const parsed = JSON.parse(raw) as TelemetryEvent;
            const receivedAt = Number(parsed.received_at || 0);

            if (receivedAt < start || receivedAt > end) continue;
            if (level && parsed.level !== level) continue;
            if (category && parsed.category !== category) continue;
            if (deviceId && parsed.device_id !== deviceId) continue;
            if (sessionId && parsed.session_id !== sessionId) continue;

            events.push(parsed);
        } catch {
            continue;
        }
    }

    events.sort((a, b) => Number(b.received_at || 0) - Number(a.received_at || 0));

    return c.json({
        count: events.length,
        limit,
        retention_seconds: TELEMETRY_RETENTION_SECONDS,
        events: events.slice(0, limit),
    });
});

// --- Session / Mesh Routes ---
apiApp.post('/sessions/broadcast', async (c) => {
    const body = await c.req.json().catch(() => ({}));
    const { device_id, session_id, payload, timestamp } = body;

    if (!device_id || !session_id) return c.json({ error: 'missing_fields' }, 400);

    const key = `session:${session_id}`;
    const update = {
        device_id,
        payload,
        timestamp: timestamp || Date.now(),
    };

    // Store session update in KV
    await c.env.SESSIONS.put(key, JSON.stringify(update), { expirationTtl: 86400 });

    return c.json({ status: 'broadcasted', key });
});

apiApp.get('/sessions/:id', async (c) => {
    const id = c.req.param('id');
    const entry = await c.env.SESSIONS.get(`session:${id}`);
    if (!entry) return c.json({ error: 'not_found' }, 404);
    return c.json(JSON.parse(entry));
});

// --- Update Routes ---
apiApp.get('/update/manifest', async (c) => {
    const platform = c.req.query('platform') || 'darwin-aarch64';
    const arch = c.req.query('arch') || 'aarch64';

    // Tauri v2 expected manifest format
    const manifest = {
        version: '0.1.1',
        notes: 'Security fixes and Hono API refactor.',
        pub_date: new Date().toISOString(),
        platforms: {
            'darwin-aarch64': {
                signature: '...', // Placeholder
                url: `https://github.com/pryx-dev/pryx/releases/download/v0.1.1/Pryx_0.1.1_aarch64.app.tar.gz`
            },
            'darwin-x86_64': {
                signature: '...',
                url: `https://github.com/pryx-dev/pryx/releases/download/v0.1.1/Pryx_0.1.1_x64.app.tar.gz`
            },
            'windows-x86_64': {
                signature: '...',
                url: `https://github.com/pryx-dev/pryx/releases/download/v0.1.1/Pryx_0.1.1_x64_en-US.msi.zip`
            }
        }
    };

    return c.json(manifest);
});


// --- Health ---
apiApp.get('/', (c) => c.json({ name: 'Pryx Cloud API', status: 'operational', engine: 'hono' }));

// --- Astro Integration ---
export const ALL: APIRoute = async (ctx) => {
    // Direct access to platform bindings
    const platform = (ctx as any).platform;
    const env = platform?.env || {};

    console.log('API bindings check:', {
        hasPlatform: !!platform,
        hasEnv: !!env,
        deviceCodes: !!env.DEVICE_CODES,
        tokens: !!env.TOKENS,
        sessions: !!env.SESSIONS
    });

    // Pass minimal execution context required by Hono.
    const executionCtx = {
        waitUntil: (promise: Promise<unknown>) => {
            const waitUntil = (ctx as any).waitUntil;
            if (typeof waitUntil === 'function') {
                waitUntil(promise);
            }
        },
        passThroughOnException: () => {},
    };

    return apiApp.fetch(ctx.request, env, executionCtx as any);
};
