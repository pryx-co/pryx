import { Hono } from 'hono';
import type { APIRoute } from 'astro';

/**
 * Unified API route for Pryx Cloud using Hono
 * Ported from vanilla Response logic for better scalability
 */

const app = new Hono<{ Bindings: any }>().basePath('/api');

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

// --- Middleware ---
app.use('*', async (c, next) => {
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
app.post('/auth/qr/pairing', async (c) => {
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

app.get('/auth/qr/status', async (c) => {
    const token = c.req.query('token');
    if (!token) return c.json({ error: 'missing_token' }, 400);

    const pairingCode = await c.env.DEVICE_CODES.get(`ptr:${token}`);
    if (!pairingCode) return c.json({ error: 'expired_token' }, 400);

    const entryStr = await c.env.DEVICE_CODES.get(`qr:${pairingCode}`);
    if (!entryStr) return c.json({ error: 'expired_pairing' }, 400);

    const entry = JSON.parse(entryStr);
    return c.json(entry);
});

app.post('/auth/device/code', async (c) => {
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

app.post('/auth/device/token', async (c) => {
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

app.post('/auth/token/refresh', async (c) => {
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
app.post('/telemetry/ingest', async (c) => {
    try {
        const body = await c.req.json();
        const events = Array.isArray(body) ? body : [body];
        const sanitized = events.map((event: any) => {
            const clean: any = {};
            for (const [key, value] of Object.entries(event)) {
                if (ALLOWED_TELEMETRY_FIELDS.has(key)) {
                    clean[key] = typeof value === 'string' ? redactPII(value) : value;
                }
            }
            return clean;
        }).filter((e: any) => e.correlation_id);

        return c.json({ accepted: sanitized.length });
    } catch (e) {
        return c.json({ error: 'Invalid JSON' }, 400);
    }
});

// --- Session / Mesh Routes ---
app.post('/sessions/broadcast', async (c) => {
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

app.get('/sessions/:id', async (c) => {
    const id = c.req.param('id');
    const entry = await c.env.SESSIONS.get(`session:${id}`);
    if (!entry) return c.json({ error: 'not_found' }, 404);
    return c.json(JSON.parse(entry));
});

// --- Update Routes ---
app.get('/update/manifest', async (c) => {
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
app.get('/', (c) => c.json({ name: 'Pryx Cloud API', status: 'operational', engine: 'hono' }));

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

    // Pass bindings through execution context for Hono
    const executionCtx = {
        ...ctx,
        env: env
    };

    return app.fetch(ctx.request, env, executionCtx);
};
