import { Hono } from 'hono';
import { cors } from 'hono/cors';

// API Routes Sub-App (defined first to avoid forward reference)
const apiApp = new Hono<{ Bindings: CloudflareBindings }>();

// Device code auth flow
apiApp.post('/auth/device/code', async (c) => {
    try {
        const formData = await c.req.formData();
        const deviceId = formData.get('device_id')?.toString() || '';
        const scopeStr = formData.get('scope')?.toString() || 'telemetry.write';

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
    } catch (error) {
        console.error('Device code error:', error);
        return c.json({ error: String(error) }, 500);
    }
});

// Device token exchange
apiApp.post('/auth/device/token', async (c) => {
    try {
        const formData = await c.req.formData();
        const deviceCode = formData.get('device_code')?.toString() || '';

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
    } catch (error) {
        console.error('Token error:', error);
        return c.json({ error: String(error) }, 500);
    }
});

// Token refresh
apiApp.post('/auth/token/refresh', async (c) => {
    try {
        const formData = await c.req.formData();
        const refreshToken = formData.get('refresh_token')?.toString() || '';

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
    } catch (error) {
        console.error('Refresh error:', error);
        return c.json({ error: String(error) }, 500);
    }
});

// QR Code pairing
apiApp.post('/auth/qr/pairing', async (c) => {
    try {
        const body = await c.req.json().catch(() => ({}));
        const deviceId = body.device_id || '';

        const pairingCode = generateCode(8, 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789');
        const pairingToken = generateCode(40, 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789');

        const entry = {
            deviceId,
            pairingCode,
            status: 'pending',
            created_at: Date.now(),
        };

        await c.env.DEVICE_CODES.put(`qr:${pairingCode}`, JSON.stringify(entry), { expirationTtl: 300 });
        await c.env.DEVICE_CODES.put(`ptr:${pairingToken}`, pairingCode, { expirationTtl: 300 });

        return c.json({
            pairing_code: pairingCode,
            pairing_token: pairingToken,
            expires_in: 300,
        });
    } catch (error) {
        console.error('QR pairing error:', error);
        return c.json({ error: String(error) }, 500);
    }
});

apiApp.get('/auth/qr/status', async (c) => {
    const token = c.req.query('token');
    if (!token) return c.json({ error: 'missing_token' }, 400);

    const pairingCode = await c.env.DEVICE_CODES.get(`ptr:${token}`);
    if (!pairingCode) return c.json({ error: 'expired_token' }, 400);

    const entryStr = await c.env.DEVICE_CODES.get(`qr:${pairingCode}`);
    if (!entryStr) return c.json({ error: 'expired_pairing' }, 400);

    return c.json(JSON.parse(entryStr));
});

// Session broadcast
apiApp.post('/sessions/broadcast', async (c) => {
    try {
        const body = await c.req.json().catch(() => ({}));
        const { device_id, session_id, payload, timestamp } = body;

        if (!device_id || !session_id) return c.json({ error: 'missing_fields' }, 400);

        const key = `session:${session_id}`;
        const update = {
            device_id,
            payload,
            timestamp: timestamp || Date.now(),
        };

        await c.env.SESSIONS.put(key, JSON.stringify(update), { expirationTtl: 86400 });

        return c.json({ status: 'broadcasted', key });
    } catch (error) {
        console.error('Broadcast error:', error);
        return c.json({ error: String(error) }, 500);
    }
});

apiApp.get('/sessions/:id', async (c) => {
    const id = c.req.param('id');
    const entry = await c.env.SESSIONS.get(`session:${id}`);
    if (!entry) return c.json({ error: 'not_found' }, 404);
    return c.json(JSON.parse(entry));
});

// Tauri update manifest
apiApp.get('/update/manifest', async (c) => {
    const platform = c.req.query('platform') || 'darwin-aarch64';

    const manifest = {
        version: '0.1.1',
        notes: 'Unified worker architecture.',
        pub_date: new Date().toISOString(),
        platforms: {
            'darwin-aarch64': {
                signature: '...',
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

// Telemetry ingest
apiApp.post('/telemetry/ingest', async (c) => {
    try {
        const body = await c.req.json();
        const events = Array.isArray(body) ? body : [body];
        return c.json({ accepted: events.length });
    } catch (e) {
        return c.json({ error: 'Invalid JSON' }, 400);
    }
});

// Main App
const app = new Hono<{ Bindings: CloudflareBindings }>();

// CORS for all routes
app.use('/*', cors());

// Root route - serves landing page
app.get('/', async (c) => {
    return c.html(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pryx - Sovereign AI Agent</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: white;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container { text-align: center; padding: 2rem; }
        h1 { font-size: 3rem; margin-bottom: 1rem; }
        p { font-size: 1.2rem; color: #a0a0a0; margin-bottom: 2rem; }
        .status { 
            display: inline-block;
            padding: 0.5rem 1rem;
            background: #10b981;
            color: white;
            border-radius: 9999px;
            font-size: 0.875rem;
        }
        .api-link {
            display: inline-block;
            margin-top: 1rem;
            padding: 0.75rem 1.5rem;
            background: #3b82f6;
            color: white;
            text-decoration: none;
            border-radius: 0.5rem;
            font-weight: 500;
        }
        .api-link:hover { background: #2563eb; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 Pryx</h1>
        <p>Sovereign AI Agent with Local-First Control Center</p>
        <span class="status">● Operational</span>
        <br>
        <a href="/api" class="api-link">Explore API</a>
    </div>
</body>
</html>
    `);
});

// Mount API routes under /api
app.route('/api', apiApp);

// SPA fallback - serves landing page for non-API routes
app.get('/*', async (c) => {
    // Don't override API routes
    if (c.req.path.startsWith('/api')) {
        return c.json({ error: 'Not Found' }, 404);
    }
    
    return c.html(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pryx - Sovereign AI Agent</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: white;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container { text-align: center; padding: 2rem; }
        h1 { font-size: 3rem; margin-bottom: 1rem; }
        p { font-size: 1.2rem; color: #a0a0a0; margin-bottom: 2rem; }
        .status { 
            display: inline-block;
            padding: 0.5rem 1rem;
            background: #10b981;
            color: white;
            border-radius: 9999px;
            font-size: 0.875rem;
        }
        .api-link {
            display: inline-block;
            margin-top: 1rem;
            padding: 0.75rem 1.5rem;
            background: #3b82f6;
            color: white;
            text-decoration: none;
            border-radius: 0.5rem;
            font-weight: 500;
        }
        .api-link:hover { background: #2563eb; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 Pryx</h1>
        <p>Sovereign AI Agent with Local-First Control Center</p>
        <span class="status">● Operational</span>
        <br>
        <a href="/api" class="api-link">Explore API</a>
    </div>
</body>
</html>
    `);
});

// Export for Cloudflare Workers
export default {
    async fetch(request: Request, env: CloudflareBindings, ctx: ExecutionContext): Promise<Response> {
        return app.fetch(request, env, ctx);
    },
};

// Type definition for Cloudflare bindings
type CloudflareBindings = {
    DEVICE_CODES: KVNamespace;
    TOKENS: KVNamespace;
    SESSIONS: KVNamespace;
    RATE_LIMITER?: RateLimit;
    ENVIRONMENT?: string;
};

// Utility function
function generateCode(length: number, charset: string): string {
    const array = new Uint8Array(length);
    crypto.getRandomValues(array);
    return Array.from(array, b => charset[b % charset.length]).join('');
}
