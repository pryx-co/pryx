import { beforeEach, describe, expect, it } from 'vitest';
import adminApi from './admin-api';

type UserRecord = {
    id: string;
    email: string;
    createdAt: string;
    lastActive: string;
    totalCost: number;
    status: 'active' | 'inactive' | 'suspended';
};

type DeviceRecord = {
    id: string;
    userId: string;
    name: string;
    platform: string;
    version: string;
    status: 'online' | 'offline' | 'syncing';
    lastSeen: string;
    ipAddress: string | null;
    isPaired: number;
};

type SessionRecord = {
    id: string;
    userId: string;
};

class FakePreparedStatement {
    private bindings: Array<string | number | null> = [];

    constructor(private readonly db: FakeD1Database, private readonly sql: string) {}

    bind(...values: Array<string | number | null>) {
        this.bindings = values;
        return this;
    }

    async all<T>() {
        return { results: this.db.selectAll<T>(this.sql, this.bindings) };
    }

    async first<T>() {
        return this.db.selectFirst<T>(this.sql, this.bindings);
    }

    async run() {
        this.db.execute(this.sql, this.bindings);
        return { success: true };
    }
}

class FakeD1Database {
    users: UserRecord[] = [];
    devices: DeviceRecord[] = [];
    sessions: SessionRecord[] = [];
    adminActions: Array<{ actionType: string; targetType: string; targetId: string }> = [];

    prepare(sql: string) {
        return new FakePreparedStatement(this, sql);
    }

    selectAll<T>(sql: string, bindings: Array<string | number | null>): T[] {
        const query = sql.replace(/\s+/g, ' ').trim().toLowerCase();

        if (query.includes('from users u left join devices')) {
            const userFilter = bindings[0] ? String(bindings[0]) : null;
            const scopedUsers = userFilter ? this.users.filter((u) => u.id === userFilter) : this.users;

            return scopedUsers.map((user) => {
                const deviceCount = this.devices.filter((d) => d.userId === user.id && d.isPaired === 1).length;
                const sessionCount = this.sessions.filter((s) => s.userId === user.id).length;
                return {
                    id: user.id,
                    email: user.email,
                    createdAt: user.createdAt,
                    lastActive: user.lastActive,
                    deviceCount,
                    sessionCount,
                    totalCost: user.totalCost,
                    status: user.status,
                } as unknown as T;
            });
        }

        if (query.includes('from devices d inner join users u')) {
            const userFilter = bindings[0] ? String(bindings[0]) : null;
            return this.devices
                .filter((d) => d.isPaired === 1)
                .filter((d) => (userFilter ? d.userId === userFilter : true))
                .map((device) => {
                    const owner = this.users.find((u) => u.id === device.userId);
                    return {
                        id: device.id,
                        userId: device.userId,
                        userEmail: owner?.email ?? 'unknown@example.com',
                        name: device.name,
                        platform: device.platform,
                        version: device.version,
                        status: device.status,
                        lastSeen: device.lastSeen,
                        ipAddress: device.ipAddress,
                    } as unknown as T;
                });
        }

        if (query.includes('select id, name, platform, version, status from devices')) {
            const userId = String(bindings[0]);
            return this.devices
                .filter((d) => d.userId === userId && d.isPaired === 1)
                .map((device) => ({
                    id: device.id,
                    name: device.name,
                    platform: device.platform,
                    version: device.version,
                    status: device.status,
                } as unknown as T));
        }

        return [];
    }

    selectFirst<T>(sql: string, bindings: Array<string | number | null>): T | null {
        const query = sql.replace(/\s+/g, ' ').trim().toLowerCase();

        if (query === 'select id from users where id = ?') {
            const userId = String(bindings[0]);
            const user = this.users.find((u) => u.id === userId);
            return user ? ({ id: user.id } as unknown as T) : null;
        }

        if (query.includes('from users u where u.id = ?')) {
            const userId = String(bindings[0]);
            const user = this.users.find((u) => u.id === userId);
            if (!user) return null;

            const deviceCount = this.devices.filter((d) => d.userId === user.id && d.isPaired === 1).length;
            const sessionCount = this.sessions.filter((s) => s.userId === user.id).length;

            return {
                id: user.id,
                email: user.email,
                createdAt: user.createdAt,
                lastActive: user.lastActive,
                deviceCount,
                sessionCount,
                totalCost: user.totalCost,
                status: user.status,
            } as unknown as T;
        }

        if (query === 'select id from devices where id = ? and is_paired = 1') {
            const deviceId = String(bindings[0]);
            const device = this.devices.find((d) => d.id === deviceId && d.isPaired === 1);
            return device ? ({ id: device.id } as unknown as T) : null;
        }

        return null;
    }

    execute(sql: string, bindings: Array<string | number | null>) {
        const query = sql.replace(/\s+/g, ' ').trim().toLowerCase();

        if (query.startsWith('update users set status = ?, last_active = ? where id = ?')) {
            const status = String(bindings[0]) as UserRecord['status'];
            const lastActive = String(bindings[1]);
            const userId = String(bindings[2]);
            const user = this.users.find((u) => u.id === userId);
            if (user) {
                user.status = status;
                user.lastActive = lastActive;
            }
            return;
        }

        if (query.startsWith('update devices set status = ?, last_seen = ? where id = ?')) {
            const status = String(bindings[0]) as DeviceRecord['status'];
            const lastSeen = String(bindings[1]);
            const deviceId = String(bindings[2]);
            const device = this.devices.find((d) => d.id === deviceId);
            if (device) {
                device.status = status;
                device.lastSeen = lastSeen;
            }
            return;
        }

        if (query.startsWith('update devices set is_paired = 0, status = ?, last_seen = ? where id = ?')) {
            const status = String(bindings[0]) as DeviceRecord['status'];
            const lastSeen = String(bindings[1]);
            const deviceId = String(bindings[2]);
            const device = this.devices.find((d) => d.id === deviceId);
            if (device) {
                device.isPaired = 0;
                device.status = status;
                device.lastSeen = lastSeen;
            }
            return;
        }

        if (query.startsWith('insert into admin_actions')) {
            this.adminActions.push({
                actionType: String(bindings[0]),
                targetType: String(bindings[1]),
                targetId: String(bindings[2]),
            });
        }
    }
}

const telemetryStore = {
    list: async () => ({ keys: [] as Array<{ name: string }> }),
    get: async () => null,
};

let fakeDb: FakeD1Database;

beforeEach(() => {
    fakeDb = new FakeD1Database();
    fakeDb.users = [
        {
            id: 'user-1',
            email: 'admin@pryx.dev',
            createdAt: '2026-01-15T10:30:00Z',
            lastActive: '2026-02-03T14:22:00Z',
            totalCost: 45.2,
            status: 'active',
        },
        {
            id: 'user-2',
            email: 'demo@example.com',
            createdAt: '2026-01-20T08:15:00Z',
            lastActive: '2026-02-03T09:45:00Z',
            totalCost: 12.5,
            status: 'active',
        },
    ];
    fakeDb.devices = [
        {
            id: 'dev-1',
            userId: 'user-1',
            name: 'MacBook Pro',
            platform: 'macos',
            version: '1.0.0',
            status: 'online',
            lastSeen: '2026-02-03T14:22:00Z',
            ipAddress: '192.168.1.100',
            isPaired: 1,
        },
        {
            id: 'dev-2',
            userId: 'user-2',
            name: 'iPhone 15',
            platform: 'ios',
            version: '1.0.0',
            status: 'offline',
            lastSeen: '2026-02-03T10:15:00Z',
            ipAddress: null,
            isPaired: 1,
        },
    ];
    fakeDb.sessions = [
        { id: 'session-1', userId: 'user-1' },
        { id: 'session-2', userId: 'user-1' },
        { id: 'session-3', userId: 'user-2' },
    ];
});

function createEnv(overrides: Record<string, unknown> = {}) {
    return {
        ENVIRONMENT: 'production',
        ADMIN_API_KEY: 'admin-secret',
        LOCALHOST_ADMIN_KEY: 'localhost-secret',
        ENABLE_UNSAFE_USER_LAYER: 'false',
        TELEMETRY: telemetryStore,
        DB: fakeDb,
        ...overrides,
    };
}

async function request(path: string, token?: string, env: Record<string, unknown> = createEnv()) {
    const headers = token ? { Authorization: `Bearer ${token}` } : undefined;
    return adminApi.request(`http://localhost${path}`, { headers }, env as never);
}

describe('admin API RBAC', () => {
    it('rejects missing auth in production', async () => {
        const response = await request('/stats');
        expect(response.status).toBe(401);
    });

    it('allows superadmin token from env binding', async () => {
        const response = await request('/stats', 'admin-secret');
        expect(response.status).toBe(200);
    });

    it('allows localhost key token', async () => {
        const response = await request('/stats', 'localhost-secret');
        expect(response.status).toBe(200);
    });

    it('allows user layer only when explicitly enabled and forbids global stats', async () => {
        const env = createEnv({ ENABLE_UNSAFE_USER_LAYER: 'true' });

        const usersResponse = await request('/users', 'user:user-1', env);
        expect(usersResponse.status).toBe(200);

        const statsResponse = await request('/stats', 'user:user-1', env);
        expect(statsResponse.status).toBe(403);
    });

    it('rejects user layer tokens when unsafe user layer is disabled', async () => {
        const response = await request('/users', 'user:user-1');
        expect(response.status).toBe(401);
    });

    it('allows localhost layer without auth outside production', async () => {
        const response = await request('/stats', undefined, createEnv({ ENVIRONMENT: 'development' }));
        expect(response.status).toBe(200);
    });

    it('rejects invalid bearer token without env fallback', async () => {
        const response = await request('/users', 'not-a-valid-token');
        expect(response.status).toBe(401);
    });
});

describe('admin API persistence', () => {
    it('reads users and devices from D1-backed records', async () => {
        const usersResponse = await request('/users', 'admin-secret');
        expect(usersResponse.status).toBe(200);
        const users = await usersResponse.json();
        expect(users).toHaveLength(2);

        const devicesResponse = await request('/devices', 'admin-secret');
        expect(devicesResponse.status).toBe(200);
        const devices = await devicesResponse.json();
        expect(devices).toHaveLength(2);
    });

    it('writes user status updates to D1-backed records', async () => {
        const updateResponse = await adminApi.request(
            'http://localhost/users/user-2',
            {
                method: 'PUT',
                headers: {
                    Authorization: 'Bearer admin-secret',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ status: 'suspended' }),
            },
            createEnv() as never,
        );

        expect(updateResponse.status).toBe(200);
        expect(fakeDb.users.find((u) => u.id === 'user-2')?.status).toBe('suspended');
        expect(fakeDb.adminActions.some((a) => a.actionType === 'user.status.updated' && a.targetId === 'user-2')).toBe(true);
    });

    it('writes unpair actions and hides unpaired devices', async () => {
        const unpairResponse = await adminApi.request(
            'http://localhost/devices/dev-2/unpair',
            {
                method: 'POST',
                headers: {
                    Authorization: 'Bearer admin-secret',
                },
            },
            createEnv() as never,
        );

        expect(unpairResponse.status).toBe(200);
        expect(fakeDb.devices.find((d) => d.id === 'dev-2')?.isPaired).toBe(0);

        const devicesResponse = await request('/devices', 'admin-secret', createEnv());
        const devices = await devicesResponse.json();
        expect(devices).toHaveLength(1);
    });

    it('filters telemetry query results for admin dashboard usage', async () => {
        const telemetryEntries = new Map<string, string>([
            ['telemetry:1:a', JSON.stringify({ correlation_id: 'corr-a', level: 'info', received_at: 1 })],
            ['telemetry:2:b', JSON.stringify({ correlation_id: 'corr-b', level: 'error', received_at: 2 })],
        ]);

        const telemetryStoreForTest = {
            list: async () => ({ keys: Array.from(telemetryEntries.keys()).map((name) => ({ name })) }),
            get: async (key: string) => telemetryEntries.get(key) ?? null,
        };

        const response = await request('/telemetry?level=error', 'admin-secret', createEnv({ TELEMETRY: telemetryStoreForTest }));
        expect(response.status).toBe(200);

        const payload = await response.json();
        expect(payload.count).toBe(1);
        expect(payload.events[0].correlation_id).toBe('corr-b');
    });
});
