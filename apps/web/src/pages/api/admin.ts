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

/**
 * Layer context containing authentication and authorization information
 */
interface LayerContext {
  layer: Layer;
  userId?: string;
  isLocalhost: boolean;
}

/**
 * Extended context with layer information attached
 */
interface AdminContext extends LayerContext {
  // Additional admin-specific context if needed
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
function extractLayer(authHeader: string | null, env: Record<string, string | undefined>): LayerContext {
  // Check for localhost bypass via environment variable
  const localhostKey = env.LOCALHOST_ADMIN_KEY || process.env.LOCALHOST_ADMIN_KEY;

  if (!authHeader) {
    // No auth header means localhost layer (local control panel)
    return { layer: 'localhost', isLocalhost: true };
  }

  const token = authHeader.replace('Bearer ', '').trim();

  // Check for localhost bypass key
  if (token === localhostKey || token === 'localhost') {
    return { layer: 'localhost', isLocalhost: true };
  }

  // Check for superadmin token
  if (token.startsWith('superadmin:')) {
    const adminId = token.replace('superadmin:', '');
    return { layer: 'superadmin', userId: adminId, isLocalhost: false };
  }

  // Check for regular user token
  if (token.startsWith('user:')) {
    const userId = token.replace('user:', '');
    return { layer: 'user', userId, isLocalhost: false };
  }

  // Fallback: treat as user with token as userId
  return { layer: 'user', userId: token, isLocalhost: false };
}

/**
 * Require a specific layer or higher for access
 * @param requiredLayers - Layers that are allowed access
 */
function requireLayer(...requiredLayers: Layer[]) {
  return async (c: any, next: () => Promise<void>) => {
    const authHeader = c.req.header('Authorization');
    const env = c.env as Record<string, string | undefined>;
    const layerContext = extractLayer(authHeader, env);

    // Store layer context in request for downstream handlers
    (c as any).req.layerContext = layerContext;

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

// Admin API router
const adminApi = new Hono();

// Enable CORS for admin API
adminApi.use('/*', cors({
  origin: ['http://localhost:4321', 'https://pryx.dev'],
  allowMethods: ['GET', 'POST', 'PUT', 'DELETE'],
  allowHeaders: ['Content-Type', 'Authorization'],
}));

// Middleware to verify admin authentication
adminApi.use('/*', async (c, next) => {
  const authHeader = c.req.header('Authorization');

  // TODO: Implement proper admin authentication
  // For now, check for admin API key
  const env = c.env as Record<string, string | undefined>;
  const adminKey = env.ADMIN_API_KEY || process.env.ADMIN_API_KEY;

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return c.json({ error: 'Unauthorized - Missing bearer token' }, 401);
  }

  const token = authHeader.replace('Bearer ', '');

  // Validate admin token
  if (token !== adminKey) {
    return c.json({ error: 'Unauthorized - Invalid credentials' }, 401);
  }

  await next();
});

// GET /api/admin/stats - Statistics (layer-aware)
adminApi.get('/stats', requireLayer('superadmin', 'localhost'), async (c) => {
  // Access layer context stored by middleware
  const layerContext = (c as any).req.layerContext as LayerContext;
  const filters = getLayerFilters(layerContext);

  // Layer-aware statistics based on access level
  let stats: Record<string, any>;

  if (filters.global) {
    // Superadmin or localhost: return global statistics
    stats = {
      totalUsers: 1247,
      activeUsers: 892,
      newUsersToday: 23,
      totalDevices: 3421,
      onlineDevices: 2187,
      offlineDevices: 1234,
      totalSessions: 15432,
      totalCost: 2847.50,
      avgCostPerUser: 2.28,
      timestamp: new Date().toISOString(),
    };
  } else {
    // Regular user: return user-specific statistics
    stats = {
      deviceCount: 3,
      sessionCount: 156,
      totalCost: 45.20,
      avgCostPerSession: 0.29,
      lastActive: new Date().toISOString(),
      timestamp: new Date().toISOString(),
    };
  }

  return c.json(stats);
});

// GET /api/admin/users - List users (layer-aware)
adminApi.get('/users', requireLayer('superadmin', 'localhost', 'user'), async (c) => {
  const layerContext = (c as any).req.layerContext as LayerContext;
  const range = c.req.query('range') || '7d';
  const filters = getLayerFilters(layerContext);

  // Layer-aware user listing
  let users: Array<Record<string, any>>;

  if (filters.userId && layerContext.layer === 'user') {
    // Regular user: return only their own user data
    users = [
      {
        id: filters.userId,
        email: 'user@example.com',
        createdAt: '2026-01-15T10:30:00Z',
        lastActive: '2026-02-03T14:22:00Z',
        deviceCount: 3,
        sessionCount: 156,
        totalCost: 45.20,
        status: 'active',
      },
    ];
  } else {
    // Superadmin or localhost: return all users
    users = [
      {
        id: 'user-001',
        email: 'admin@pryx.dev',
        createdAt: '2026-01-15T10:30:00Z',
        lastActive: '2026-02-03T14:22:00Z',
        deviceCount: 3,
        sessionCount: 156,
        totalCost: 45.20,
        status: 'active',
      },
      {
        id: 'user-002',
        email: 'demo@example.com',
        createdAt: '2026-01-20T08:15:00Z',
        lastActive: '2026-02-03T09:45:00Z',
        deviceCount: 2,
        sessionCount: 89,
        totalCost: 12.50,
        status: 'active',
      },
    ];
  }

  return c.json(users);
});

// GET /api/admin/users/:id - Get detailed user info
adminApi.get('/users/:id', async (c) => {
  const userId = c.req.param('id');

  // TODO: Fetch from D1 database
  const user = {
    id: userId,
    email: 'user@example.com',
    createdAt: '2026-01-15T10:30:00Z',
    lastActive: '2026-02-03T14:22:00Z',
    deviceCount: 3,
    sessionCount: 156,
    totalCost: 45.20,
    status: 'active',
    devices: [
      {
        id: 'dev-001',
        name: 'MacBook Pro',
        platform: 'macos',
        version: '1.0.0',
        status: 'online',
      },
    ],
    providers: ['openai', 'anthropic'],
    channels: ['telegram'],
  };

  return c.json(user);
});

// PUT /api/admin/users/:id - Update user (suspend/activate)
adminApi.put('/users/:id', async (c) => {
  const userId = c.req.param('id');
  const body = await c.req.json();

  // TODO: Update user status in database
  return c.json({
    id: userId,
    status: body.status,
    updatedAt: new Date().toISOString(),
  });
});

// GET /api/admin/devices - List devices (layer-aware)
adminApi.get('/devices', requireLayer('superadmin', 'localhost', 'user'), async (c) => {
  const layerContext = (c as any).req.layerContext as LayerContext;
  const filters = getLayerFilters(layerContext);

  // Layer-aware device listing
  let devices: Array<Record<string, any>>;

  if (filters.userId && layerContext.layer === 'user') {
    // Regular user: return only their own devices
    devices = [
      {
        id: 'dev-001',
        userId: filters.userId,
        userEmail: 'user@example.com',
        name: 'MacBook Pro',
        platform: 'macos',
        version: '1.0.0',
        status: 'online',
        lastSeen: '2026-02-03T14:22:00Z',
        ipAddress: '192.168.1.100',
      },
    ];
  } else {
    // Superadmin or localhost: return all devices
    devices = [
      {
        id: 'dev-001',
        userId: 'user-001',
        userEmail: 'admin@pryx.dev',
        name: 'MacBook Pro',
        platform: 'macos',
        version: '1.0.0',
        status: 'online',
        lastSeen: '2026-02-03T14:22:00Z',
        ipAddress: '192.168.1.100',
      },
      {
        id: 'dev-002',
        userId: 'user-001',
        userEmail: 'admin@pryx.dev',
        name: 'iPhone 15',
        platform: 'ios',
        version: '1.0.0',
        status: 'offline',
        lastSeen: '2026-02-03T10:15:00Z',
        ipAddress: null,
      },
    ];
  }

  return c.json(devices);
});

// POST /api/admin/devices/:id/sync - Force sync a device
adminApi.post('/devices/:id/sync', async (c) => {
  const deviceId = c.req.param('id');

  // TODO: Trigger device sync
  return c.json({
    id: deviceId,
    syncStatus: 'initiated',
    timestamp: new Date().toISOString(),
  });
});

// POST /api/admin/devices/:id/unpair - Unpair a device
adminApi.post('/devices/:id/unpair', async (c) => {
  const deviceId = c.req.param('id');

  // TODO: Unpair device in database
  return c.json({
    id: deviceId,
    status: 'unpaired',
    timestamp: new Date().toISOString(),
  });
});

// GET /api/admin/costs - Cost analytics (layer-aware)
adminApi.get('/costs', requireLayer('superadmin', 'localhost', 'user'), async (c) => {
  const layerContext = (c as any).req.layerContext as LayerContext;
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

// GET /api/admin/health - System health status
adminApi.get('/health', async (c) => {
  // TODO: Check actual system health
  const health = {
    runtimeStatus: 'healthy',
    apiLatency: 45,
    errorRate: 0.001,
    dbStatus: 'connected',
    queueDepth: 12,
    activeConnections: 456,
    timestamp: new Date().toISOString(),
  };

  return c.json(health);
});

// GET /api/admin/telemetry - Real-time telemetry stream (SSE)
adminApi.get('/telemetry', async (c) => {
  // TODO: Implement Server-Sent Events for real-time telemetry
  return c.json({
    message: 'SSE endpoint for real-time telemetry - To be implemented',
  });
});

// GET /api/admin/telemetry/config - Get telemetry configuration (superadmin only)
adminApi.get('/telemetry/config', requireLayer('superadmin', 'localhost'), async (c) => {
  // Return telemetry configuration settings
  const config = {
    enabled: true,
    retentionDays: 7,
    exportBackend: 'https://api.pryx.dev',
    samplingRate: 0.1,
    logLevel: 'info',
   敏感数据过滤: true,
    batchSize: 100,
    flushInterval: 5000,
  };

  return c.json(config);
});

// PUT /api/admin/telemetry/config - Update telemetry configuration (superadmin only)
adminApi.put('/telemetry/config', requireLayer('superadmin', 'localhost'), async (c) => {
  const body = await c.req.json();

  // Validate and update telemetry configuration
  const config = {
    enabled: body.enabled ?? true,
    retentionDays: body.retentionDays ?? 7,
    exportBackend: body.exportBackend ?? 'https://api.pryx.dev',
    samplingRate: body.samplingRate ?? 0.1,
    logLevel: body.logLevel ?? 'info',
    敏感数据过滤: body.敏感数据过滤 ?? true,
    batchSize: body.batchSize ?? 100,
    flushInterval: body.flushInterval ?? 5000,
    updatedAt: new Date().toISOString(),
    updatedBy: (c as any).req.layerContext?.userId || 'localhost',
  };

  return c.json(config);
});

// GET /api/admin/logs - System logs
adminApi.get('/logs', async (c) => {
  const level = c.req.query('level') || 'info';
  const limit = parseInt(c.req.query('limit') || '100');

  // TODO: Fetch logs from Cloudflare Workers analytics
  const logs = [
    {
      timestamp: '2026-02-03T14:22:00Z',
      level: 'info',
      message: 'User login successful',
      userId: 'user-001',
    },
    {
      timestamp: '2026-02-03T14:20:00Z',
      level: 'error',
      message: 'Device sync failed',
      deviceId: 'dev-002',
      error: 'Connection timeout',
    },
  ];

  return c.json({ logs, total: logs.length });
});

export default adminApi;
