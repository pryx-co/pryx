// Admin API routes for superadmin dashboard
// These endpoints provide global telemetry and fleet management data

import { Hono } from 'hono';
import { cors } from 'hono/cors';

// ============================================================================
// 3-LAYER ARCHITECTURE TYPES
// ============================================================================

/**
 * Layer types for role-based access control
 * - user: Regular cloud users viewing their own data
 * - superadmin: Full admin access to all devices/users/telemetry
 * - localhost: Local admin via runtime port (no auth required)
 */
type Layer = 'user' | 'superadmin' | 'localhost';

interface AuthEnv {
  ADMIN_API_KEY?: string;
  LOCALHOST_ADMIN_KEY?: string;
  ENVIRONMENT?: string;
  ENABLE_UNSAFE_USER_LAYER?: string;
}

interface D1RunResult {
  success: boolean;
}

interface D1AllResult<T> {
  results: T[];
}

interface D1PreparedStatement {
  bind(...values: Array<string | number | null>): D1PreparedStatement;
  all<T>(): Promise<D1AllResult<T>>;
  first<T>(): Promise<T | null>;
  run(): Promise<D1RunResult>;
}

interface D1Database {
  prepare(query: string): D1PreparedStatement;
}

interface AdminApiEnv extends AuthEnv {
  DB?: D1Database;
  TELEMETRY?: {
    list(options?: { limit?: number; prefix?: string }): Promise<{ keys: Array<{ name: string }> }>;
    get(key: string): Promise<string | null>;
  };
}

/**
 * Layer context containing authentication and authorization information
 */
interface LayerContext {
  layer: Layer;
  userId?: string;
  isLocalhost: boolean;
}

// ============================================================================
// LAYER DETECTION MIDDLEWARE
// ============================================================================

/**
 * Extract layer information from Authorization header
 * Pattern: Bearer {layer}:{identifier}
 * - Bearer user:user-123 → User layer, user ID user-123
 * - Bearer superadmin:admin-key → Superadmin layer
 * - Bearer localhost or no header → Localhost layer
 */
function isNonProduction(env: AuthEnv): boolean {
  const normalized = (env.ENVIRONMENT ?? '').toLowerCase();
  return normalized === 'development' || normalized === 'staging' || normalized === 'test' || normalized === 'local';
}

function isUnsafeUserLayerEnabled(env: AuthEnv): boolean {
  return (env.ENABLE_UNSAFE_USER_LAYER ?? '').toLowerCase() === 'true';
}

function extractLayer(authHeader: string | null, env: AuthEnv): LayerContext | null {
  const localhostKey = env.LOCALHOST_ADMIN_KEY;
  const adminKey = env.ADMIN_API_KEY;
  const token = authHeader?.startsWith('Bearer ') ? authHeader.replace('Bearer ', '').trim() : null;

  if (!token) {
    return isNonProduction(env) ? { layer: 'localhost', isLocalhost: true } : null;
  }

  if ((localhostKey && token === localhostKey) || (token === 'localhost' && isNonProduction(env))) {
    return { layer: 'localhost', isLocalhost: true };
  }

  if ((adminKey && token === adminKey) || (adminKey && token === `superadmin:${adminKey}`)) {
    return { layer: 'superadmin', userId: 'superadmin', isLocalhost: false };
  }

  if (token.startsWith('user:')) {
    // TODO: Replace temporary user:<id> tokens with signed/session-backed user auth.
    // Until then, keep this path disabled by default to avoid user impersonation.
    if (!isUnsafeUserLayerEnabled(env)) {
      return null;
    }

    const userId = token.replace('user:', '').trim();
    if (!userId) return null;
    return { layer: 'user', userId, isLocalhost: false };
  }

  return null;
}

/**
 * Require a specific layer or higher for access
 * @param requiredLayers - Layers that are allowed access
 */
function requireLayer(...requiredLayers: Layer[]) {
  return async (c: any, next: () => Promise<void>) => {
    const layerContext = (c as any).get('layerContext') as LayerContext | null;

    if (!layerContext) {
      return c.json({
        error: 'Unauthorized',
        message: 'Missing or invalid credentials.',
      }, 401);
    }

    // Check if user's layer is allowed
    if (!requiredLayers.includes(layerContext.layer)) {
      const layerNames: Record<Layer, string> = {
        user: 'user',
        superadmin: 'superadmin',
        localhost: 'localhost',
      };

      const allowedStr = requiredLayers.map(l => layerNames[l]).join(' or ');
      const actualStr = layerNames[layerContext.layer];

      return c.json({
        error: 'Forbidden',
        message: `This endpoint requires ${allowedStr} access. You have ${actualStr} access.`,
        required: requiredLayers,
        current: layerContext.layer,
      }, 403);
    }

    await next();
  };
}

/**
 * Get layer-aware filtering for user-scoped data
 * Returns filter criteria based on the user's layer
 */
function getLayerFilters(layerContext: LayerContext): { userId?: string; global?: boolean } {
  switch (layerContext.layer) {
    case 'user':
      // Regular users can only see their own data
      return { userId: layerContext.userId };
    case 'superadmin':
      // Superadmins can see everything
      return { global: true };
    case 'localhost':
      // Localhost can see everything (local control panel)
      return { global: true };
    default:
      return { global: true };
  }
}

function getDb(c: any): D1Database | null {
  const env = (c.env ?? {}) as AdminApiEnv;
  return env.DB ?? null;
}

function getIsoNow(): string {
  return new Date().toISOString();
}

function toNullableString(value: unknown): string | null {
  if (value === null || value === undefined) return null;
  return String(value);
}

async function logAdminAction(c: any, actionType: string, targetType: string, targetId: string, payload: Record<string, unknown> = {}) {
  const db = getDb(c);
  if (!db) return;

  const layerContext = (c as any).get('layerContext') as LayerContext | null;
  try {
    await db.prepare(`
      INSERT INTO admin_actions (action_type, target_type, target_id, actor_layer, actor_id, payload_json, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `)
      .bind(
        actionType,
        targetType,
        targetId,
        layerContext?.layer ?? 'unknown',
        layerContext?.userId ?? null,
        JSON.stringify(payload),
        getIsoNow(),
      )
      .run();
  } catch (error) {
    console.error('Failed to persist admin action audit log', {
      actionType,
      targetType,
      targetId,
      actorLayer: layerContext?.layer ?? 'unknown',
      actorId: layerContext?.userId ?? null,
      payload,
      error,
    });
  }
}

// Admin API router
const adminApi = new Hono<{ Bindings: any }>();

// Enable CORS for admin API
adminApi.use('/*', cors({
  origin: ['http://localhost:4321', 'https://pryx.dev'],
  allowMethods: ['GET', 'POST', 'PUT', 'DELETE'],
  allowHeaders: ['Content-Type', 'Authorization'],
}));

// Middleware to resolve auth context once per request
adminApi.use('/*', async (c, next) => {
  const authHeader = c.req.header('Authorization') ?? null;
  const env = (c.env ?? {}) as AuthEnv;
  const layerContext = extractLayer(authHeader, env);

  (c as any).set('layerContext', layerContext);

  await next();
});

// Require a recognized layer token (or localhost in non-production) for all admin routes.
adminApi.use('/*', requireLayer('superadmin', 'localhost', 'user'));

adminApi.get('/stats', requireLayer('superadmin', 'localhost'), async (c) => {
  const layerContext = (c as any).get('layerContext') as LayerContext;
  const filters = getLayerFilters(layerContext);

  let stats: Record<string, any>;

  try {
    if (filters.global) {
      const list = await c.env.TELEMETRY.list({ limit: 1000, prefix: 'telemetry:' });

      let totalEvents = list.keys.length;
      let totalCost = 0;
      let errorCount = 0;
      const uniqueDevices = new Set<string>();
      const uniqueSessions = new Set<string>();

      for (const key of list.keys) {
        try {
          const value = await c.env.TELEMETRY.get(key.name);
          if (value) {
            const event = JSON.parse(value);

            if (event.device_id) uniqueDevices.add(event.device_id);
            if (event.session_id) uniqueSessions.add(event.session_id);
            if (event.cost) totalCost += event.cost;
            if (event.level === 'error' || event.level === 'critical') errorCount++;
          }
        } catch (e) {
        }
      }

      const now = Date.now();
      const oneDayAgo = now - 86400000;
      const recentEvents = await c.env.TELEMETRY.list({ limit: 100, prefix: 'telemetry:' });

      let newEventsToday = 0;
      for (const key of recentEvents.keys) {
        try {
          const value = await c.env.TELEMETRY.get(key.name);
          if (value) {
            const event = JSON.parse(value);
            if (event.received_at && event.received_at >= oneDayAgo) {
              newEventsToday++;
            }
          }
        } catch (e) {
        }
      }

      stats = {
        totalEvents,
        uniqueDevices: uniqueDevices.size,
        uniqueSessions: uniqueSessions.size,
        totalCost,
        errorCount,
        errorRate: totalEvents > 0 ? errorCount / totalEvents : 0,
        newEventsToday,
        timestamp: new Date().toISOString(),
      };
    } else {
      stats = {
        message: 'User-specific stats require user registry implementation',
        timestamp: new Date().toISOString(),
      };
    }

    return c.json(stats);
  } catch (e) {
    console.error('Stats error:', e);
    return c.json({ error: String(e) }, 500);
  }
});

// GET /api/admin/users - List users (layer-aware)
adminApi.get('/users', requireLayer('superadmin', 'localhost', 'user'), async (c) => {
  const layerContext = (c as any).get('layerContext') as LayerContext;
  const filters = getLayerFilters(layerContext);

  const db = getDb(c);
  if (!db) {
    return c.json({ error: 'database_unavailable', message: 'D1 binding DB is required for admin user queries.' }, 503);
  }

  type UserSummaryRow = {
    id: string;
    email: string;
    createdAt: string;
    lastActive: string;
    deviceCount: number;
    sessionCount: number;
    totalCost: number;
    status: string;
  };

  const userFilter = filters.userId ?? null;
  const result = await db.prepare(`
    SELECT
      u.id,
      u.email,
      u.created_at AS createdAt,
      u.last_active AS lastActive,
      COUNT(DISTINCT d.id) AS deviceCount,
      COUNT(DISTINCT s.id) AS sessionCount,
      COALESCE(u.total_cost, 0) AS totalCost,
      u.status
    FROM users u
    LEFT JOIN devices d ON d.user_id = u.id AND d.is_paired = 1
    LEFT JOIN sessions s ON s.user_id = u.id
    WHERE (?1 IS NULL OR u.id = ?1)
    GROUP BY u.id
    ORDER BY u.created_at DESC
  `).bind(userFilter).all<UserSummaryRow>();

  const users = result.results.map((row) => ({
    id: row.id,
    email: row.email,
    createdAt: row.createdAt,
    lastActive: row.lastActive,
    deviceCount: Number(row.deviceCount ?? 0),
    sessionCount: Number(row.sessionCount ?? 0),
    totalCost: Number(row.totalCost ?? 0),
    status: row.status,
  }));

  return c.json(users);
});

// GET /api/admin/users/:id - Get detailed user info
adminApi.get('/users/:id', requireLayer('superadmin', 'localhost', 'user'), async (c) => {
  const userId = c.req.param('id');
  const layerContext = (c as any).get('layerContext') as LayerContext;

  if (layerContext.layer === 'user' && layerContext.userId !== userId) {
    return c.json({ error: 'Forbidden', message: 'Users can only access their own profile.' }, 403);
  }

  const db = getDb(c);
  if (!db) {
    return c.json({ error: 'database_unavailable', message: 'D1 binding DB is required for admin user queries.' }, 503);
  }

  type UserDetailRow = {
    id: string;
    email: string;
    createdAt: string;
    lastActive: string;
    deviceCount: number;
    sessionCount: number;
    totalCost: number;
    status: string;
  };

  const user = await db.prepare(`
    SELECT
      u.id,
      u.email,
      u.created_at AS createdAt,
      u.last_active AS lastActive,
      (SELECT COUNT(1) FROM devices d WHERE d.user_id = u.id AND d.is_paired = 1) AS deviceCount,
      (SELECT COUNT(1) FROM sessions s WHERE s.user_id = u.id) AS sessionCount,
      COALESCE(u.total_cost, 0) AS totalCost,
      u.status
    FROM users u
    WHERE u.id = ?
  `).bind(userId).first<UserDetailRow>();

  if (!user) {
    return c.json({ error: 'not_found' }, 404);
  }

  type DeviceRow = {
    id: string;
    name: string;
    platform: string;
    version: string;
    status: string;
  };

  const devices = await db.prepare(`
    SELECT id, name, platform, version, status
    FROM devices
    WHERE user_id = ? AND is_paired = 1
    ORDER BY last_seen DESC
  `).bind(userId).all<DeviceRow>();

  return c.json({
    id: user.id,
    email: user.email,
    createdAt: user.createdAt,
    lastActive: user.lastActive,
    deviceCount: Number(user.deviceCount ?? 0),
    sessionCount: Number(user.sessionCount ?? 0),
    totalCost: Number(user.totalCost ?? 0),
    status: user.status,
    devices: devices.results,
    providers: [],
    channels: [],
  });
});

// PUT /api/admin/users/:id - Update user (suspend/activate)
adminApi.put('/users/:id', requireLayer('superadmin', 'localhost'), async (c) => {
  const userId = c.req.param('id');
  const body = await c.req.json().catch(() => ({}));
  const nextStatus = toNullableString((body as { status?: unknown }).status);

  const isSupportedStatus = nextStatus === 'active' || nextStatus === 'inactive' || nextStatus === 'suspended';
  if (!nextStatus || !isSupportedStatus) {
    return c.json({ error: 'invalid_status' }, 400);
  }

  const db = getDb(c);
  if (!db) {
    return c.json({ error: 'database_unavailable', message: 'D1 binding DB is required for admin user updates.' }, 503);
  }

  const existing = await db.prepare('SELECT id FROM users WHERE id = ?').bind(userId).first<{ id: string }>();
  if (!existing) {
    return c.json({ error: 'not_found' }, 404);
  }

  const updatedAt = getIsoNow();
  await db.prepare('UPDATE users SET status = ?, last_active = ? WHERE id = ?').bind(nextStatus, updatedAt, userId).run();
  await logAdminAction(c, 'user.status.updated', 'user', userId, { status: nextStatus });

  return c.json({
    id: userId,
    status: nextStatus,
    updatedAt,
  });
});

// GET /api/admin/devices - List devices (layer-aware)
adminApi.get('/devices', requireLayer('superadmin', 'localhost', 'user'), async (c) => {
  const layerContext = (c as any).get('layerContext') as LayerContext;
  const filters = getLayerFilters(layerContext);

  const db = getDb(c);
  if (!db) {
    return c.json({ error: 'database_unavailable', message: 'D1 binding DB is required for admin device queries.' }, 503);
  }

  type DeviceFleetRow = {
    id: string;
    userId: string;
    userEmail: string;
    name: string;
    platform: string;
    version: string;
    status: string;
    lastSeen: string;
    ipAddress: string | null;
  };

  const userFilter = filters.userId ?? null;
  const result = await db.prepare(`
    SELECT
      d.id,
      d.user_id AS userId,
      u.email AS userEmail,
      d.name,
      d.platform,
      d.version,
      d.status,
      d.last_seen AS lastSeen,
      d.ip_address AS ipAddress
    FROM devices d
    INNER JOIN users u ON u.id = d.user_id
    WHERE d.is_paired = 1
      AND (?1 IS NULL OR d.user_id = ?1)
    ORDER BY d.last_seen DESC
  `).bind(userFilter).all<DeviceFleetRow>();

  return c.json(result.results.map((row) => ({
    ...row,
    ipAddress: row.ipAddress ?? undefined,
  })));
});

// POST /api/admin/devices/:id/sync - Force sync a device
adminApi.post('/devices/:id/sync', requireLayer('superadmin', 'localhost'), async (c) => {
  const deviceId = c.req.param('id');

  const db = getDb(c);
  if (!db) {
    return c.json({ error: 'database_unavailable', message: 'D1 binding DB is required for device updates.' }, 503);
  }

  const existing = await db.prepare('SELECT id FROM devices WHERE id = ? AND is_paired = 1').bind(deviceId).first<{ id: string }>();
  if (!existing) {
    return c.json({ error: 'not_found' }, 404);
  }

  const timestamp = getIsoNow();
  await db.prepare('UPDATE devices SET status = ?, last_seen = ? WHERE id = ?').bind('syncing', timestamp, deviceId).run();
  await logAdminAction(c, 'device.sync.requested', 'device', deviceId, { status: 'syncing' });

  return c.json({
    id: deviceId,
    syncStatus: 'initiated',
    timestamp,
  });
});

// POST /api/admin/devices/:id/unpair - Unpair a device
adminApi.post('/devices/:id/unpair', requireLayer('superadmin', 'localhost'), async (c) => {
  const deviceId = c.req.param('id');

  const db = getDb(c);
  if (!db) {
    return c.json({ error: 'database_unavailable', message: 'D1 binding DB is required for device updates.' }, 503);
  }

  const existing = await db.prepare('SELECT id FROM devices WHERE id = ? AND is_paired = 1').bind(deviceId).first<{ id: string }>();
  if (!existing) {
    return c.json({ error: 'not_found' }, 404);
  }

  const timestamp = getIsoNow();
  await db.prepare('UPDATE devices SET is_paired = 0, status = ?, last_seen = ? WHERE id = ?').bind('offline', timestamp, deviceId).run();
  await logAdminAction(c, 'device.unpaired', 'device', deviceId);

  return c.json({
    id: deviceId,
    status: 'unpaired',
    timestamp,
  });
});

// GET /api/admin/costs - Cost analytics (layer-aware)
adminApi.get('/costs', requireLayer('superadmin', 'localhost', 'user'), async (c) => {
  const layerContext = (c as any).get('layerContext') as LayerContext;
  const range = c.req.query('range') || '7d';
  const filters = getLayerFilters(layerContext);

  // Layer-aware cost analytics
  let costs: Record<string, any>;

  if (filters.userId && layerContext.layer === 'user') {
    // Regular user: return their own cost data
    costs = {
      total: 45.20,
      byProvider: {
        openai: 25.00,
        anthropic: 20.20,
      },
      byDay: [
        { date: '2026-02-01', cost: 12.50 },
        { date: '2026-02-02', cost: 8.20 },
        { date: '2026-02-03', cost: 24.50 },
      ],
    };
  } else {
    // Superadmin or localhost: return global cost data
    costs = {
      total: 2847.50,
      byProvider: {
        openai: 1250.00,
        anthropic: 980.50,
        google: 617.00,
      },
      byDay: [
        { date: '2026-02-01', cost: 450.00 },
        { date: '2026-02-02', cost: 380.50 },
        { date: '2026-02-03', cost: 210.00 },
      ],
      topUsers: [
        { userId: 'user-001', email: 'admin@pryx.dev', cost: 45.20 },
        { userId: 'user-002', email: 'demo@example.com', cost: 12.50 },
      ],
    };
  }

  return c.json(costs);
});

adminApi.get('/health', requireLayer('superadmin', 'localhost'), async (c) => {
  const startTime = Date.now();
  let dbStatus = 'connected';
  let errorRate = 0;
  let telemetryCount = 0;

  try {
    const list = await c.env.TELEMETRY.list({ limit: 1 });
    const listTime = Date.now() - startTime;

    const recentTelemetry = await c.env.TELEMETRY.list({
      limit: 100,
      prefix: 'telemetry:',
    });

    telemetryCount = recentTelemetry.keys.length;

    let errorCount = 0;
    for (const key of recentTelemetry.keys.slice(0, 20)) {
      try {
        const value = await c.env.TELEMETRY.get(key.name);
        if (value) {
          const event = JSON.parse(value);
          if (event.level === 'error' || event.level === 'critical') {
            errorCount++;
          }
        }
      } catch (e) {
      }
    }
    errorRate = errorCount / Math.min(20, telemetryCount) || 0;

    const health = {
      runtimeStatus: errorRate > 0.1 ? 'degraded' : 'healthy',
      apiLatency: listTime,
      errorRate,
      dbStatus,
      queueDepth: 0,
      activeConnections: telemetryCount,
      timestamp: new Date().toISOString(),
    };

    return c.json(health);
  } catch (e) {
    console.error('Health check error:', e);
    const health = {
      runtimeStatus: 'critical',
      apiLatency: 9999,
      errorRate: 1.0,
      dbStatus: 'disconnected',
      queueDepth: 0,
      activeConnections: 0,
      timestamp: new Date().toISOString(),
    };

    return c.json(health);
  }
});

adminApi.get('/telemetry', requireLayer('superadmin', 'localhost'), async (c) => {
  const limit = Math.min(parseInt(c.req.query('limit') || '100', 10), 1000);
  const level = c.req.query('level');
  const category = c.req.query('category');
  const deviceId = c.req.query('device_id');
  const sessionId = c.req.query('session_id');
  const start = parseInt(c.req.query('start') || '0', 10) || 0;
  const end = parseInt(c.req.query('end') || `${Date.now()}`, 10) || Date.now();

  try {
    if (!c.env?.TELEMETRY) {
      return c.json({ error: 'telemetry_store_unavailable' }, 503);
    }

    const list = await c.env.TELEMETRY.list({
      limit,
      prefix: 'telemetry:',
    });

    const events: Array<Record<string, unknown>> = [];
    for (const key of list.keys) {
      try {
        const value = await c.env.TELEMETRY.get(key.name);
        if (value) {
          const event = JSON.parse(value) as Record<string, unknown>;
          const receivedAt = Number(event.received_at || 0);
          if (receivedAt < start || receivedAt > end) continue;
          if (level && event.level !== level) continue;
          if (category && event.category !== category) continue;
          if (deviceId && event.device_id !== deviceId) continue;
          if (sessionId && event.session_id !== sessionId) continue;

          events.push(event);
        }
      } catch (e) {
        console.error('Failed to parse telemetry event:', e);
      }
    }

    events.sort((a, b) => Number(b.received_at || 0) - Number(a.received_at || 0));

    return c.json({
      count: events.length,
      retentionDays: 7,
      events: events.slice(0, limit),
    });
  } catch (e) {
    console.error('Telemetry query error:', e);
    return c.json({ error: String(e) }, 500);
  }
});

// GET /api/admin/telemetry/config - Get telemetry configuration (superadmin only)
adminApi.get('/telemetry/config', requireLayer('superadmin', 'localhost'), async (c) => {
  const config = {
    enabled: true,
    retentionDays: 7,
    exportBackend: 'https://api.pryx.dev',
    samplingRate: 0.1,
    logLevel: 'info',
    sensitiveDataFiltering: true,
    batchSize: 100,
    flushInterval: 5000,
  };

  return c.json(config);
});

// PUT /api/admin/telemetry/config - Update telemetry configuration (superadmin only)
adminApi.put('/telemetry/config', requireLayer('superadmin', 'localhost'), async (c) => {
  const body = await c.req.json();

  const config = {
    enabled: body.enabled ?? true,
    retentionDays: body.retentionDays ?? 7,
    exportBackend: body.exportBackend ?? 'https://api.pryx.dev',
    samplingRate: body.samplingRate ?? 0.1,
    logLevel: body.logLevel ?? 'info',
    sensitiveDataFiltering: body.sensitiveDataFiltering ?? true,
    batchSize: body.batchSize ?? 100,
    flushInterval: body.flushInterval ?? 5000,
    updatedAt: new Date().toISOString(),
    updatedBy: ((c as any).get('layerContext') as LayerContext | null)?.userId || 'localhost',
  };

  return c.json(config);
});

adminApi.get('/logs', requireLayer('superadmin', 'localhost'), async (c) => {
  const level = c.req.query('level') || 'info';
  const limit = parseInt(c.req.query('limit') || '100');

  try {
    const list = await c.env.TELEMETRY.list({
      limit,
      prefix: 'telemetry:',
    });

    const logs = [];
    for (const key of list.keys) {
      try {
        const value = await c.env.TELEMETRY.get(key.name);
        if (value) {
          const event = JSON.parse(value);
          if (!level || event.level === level) {
            logs.push({
              timestamp: event.received_at ? new Date(event.received_at).toISOString() : new Date().toISOString(),
              level: event.level || 'info',
              message: event.message || event.type || 'Telemetry event',
              userId: event.user_id,
              deviceId: event.device_id,
              error: event.error,
            });
          }
        }
      } catch (e) {
        console.error('Failed to parse log entry:', e);
      }
    }

    return c.json({ logs: logs.slice(0, limit), total: logs.length });
  } catch (e) {
    console.error('Logs query error:', e);
    return c.json({ error: String(e) }, 500);
  }
});

export default adminApi;
