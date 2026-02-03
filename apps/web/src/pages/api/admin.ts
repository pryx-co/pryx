// Admin API routes for superadmin dashboard
// These endpoints provide global telemetry and fleet management data

import { Hono } from 'hono';
import { cors } from 'hono/cors';

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
  const adminKey = c.env.ADMIN_API_KEY || process.env.ADMIN_API_KEY;

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

// GET /api/admin/stats - Global statistics
adminApi.get('/stats', async (c) => {
  // TODO: Connect to D1 database to fetch real stats
  const stats = {
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

  return c.json(stats);
});

// GET /api/admin/users - List all users with summary data
adminApi.get('/users', async (c) => {
  const range = c.req.query('range') || '7d';

  // TODO: Fetch from D1 database
  const users = [
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

// GET /api/admin/devices - List all devices across all users
adminApi.get('/devices', async (c) => {
  // TODO: Fetch from D1 database
  const devices = [
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

// GET /api/admin/costs - Cost analytics
adminApi.get('/costs', async (c) => {
  const range = c.req.query('range') || '7d';

  // TODO: Fetch cost data from D1 database
  const costs = {
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
