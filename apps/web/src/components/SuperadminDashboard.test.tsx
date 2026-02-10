import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import SuperadminDashboard from './SuperadminDashboard';

const statsPayload = {
    totalUsers: 1247,
    activeUsers: 892,
    newUsersToday: 23,
    totalDevices: 3421,
    onlineDevices: 2187,
    offlineDevices: 1234,
    totalSessions: 15432,
    totalCost: 2847.5,
    avgCostPerUser: 2.28,
};

const usersPayload = [
    {
        id: 'user-1',
        email: 'admin@pryx.dev',
        createdAt: '2026-01-15T10:30:00Z',
        lastActive: '2026-02-03T14:22:00Z',
        deviceCount: 3,
        sessionCount: 156,
        totalCost: 45.2,
        status: 'active',
    },
];

const devicesPayload = [
    {
        id: 'dev-1',
        userId: 'user-1',
        userEmail: 'admin@pryx.dev',
        name: 'MacBook Pro',
        platform: 'macos',
        version: '1.0.0',
        status: 'online',
        lastSeen: '2026-02-03T14:22:00Z',
    },
];

const healthPayload = {
    runtimeStatus: 'healthy',
    apiLatency: 45,
    errorRate: 0.001,
    dbStatus: 'connected',
    queueDepth: 12,
    activeConnections: 456,
};

function createJsonResponse(payload: unknown): Response {
    return {
        ok: true,
        json: async () => payload,
    } as Response;
}

describe('SuperadminDashboard', () => {
    const mockOnLogout = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();

        vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
            const url = typeof input === 'string' ? input : input.toString();

            if (url.includes('/api/admin/stats')) return createJsonResponse(statsPayload);
            if (url.includes('/api/admin/users')) return createJsonResponse(usersPayload);
            if (url.includes('/api/admin/devices')) return createJsonResponse(devicesPayload);
            if (url.includes('/api/admin/health')) return createJsonResponse(healthPayload);

            return {
                ok: false,
                json: async () => ({ error: 'not_found' }),
            } as Response;
        });
    });

    it('renders dashboard header with title', async () => {
        render(<SuperadminDashboard onLogout={mockOnLogout} />);
        expect(await screen.findByText('Pryx Superadmin Dashboard')).toBeInTheDocument();
    });

    it('renders navigation tabs', async () => {
        render(<SuperadminDashboard onLogout={mockOnLogout} />);
        await screen.findByText('Pryx Superadmin Dashboard');
        expect(screen.getByRole('button', { name: /^Overview$/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /^Users/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /^Devices/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /^Costs$/ })).toBeInTheDocument();
        expect(screen.getAllByRole('button', { name: /^System Health$/ }).length).toBeGreaterThan(0);
    });

    it('displays logout button', async () => {
        render(<SuperadminDashboard onLogout={mockOnLogout} />);
        await screen.findByText('Pryx Superadmin Dashboard');
        expect(screen.getByRole('button', { name: 'Logout' })).toBeInTheDocument();
    });

    it('calls onLogout when logout button is clicked', async () => {
        render(<SuperadminDashboard onLogout={mockOnLogout} />);
        await screen.findByText('Pryx Superadmin Dashboard');
        fireEvent.click(screen.getByRole('button', { name: 'Logout' }));
        expect(mockOnLogout).toHaveBeenCalledTimes(1);
    });

    it('switches to users tab when clicked', async () => {
        render(<SuperadminDashboard onLogout={mockOnLogout} />);
        await screen.findByText('Pryx Superadmin Dashboard');
        fireEvent.click(screen.getByRole('button', { name: /^Users/ }));
        expect(await screen.findByText('User Management')).toBeInTheDocument();
    });

    it('switches to devices tab when clicked', async () => {
        render(<SuperadminDashboard onLogout={mockOnLogout} />);
        await screen.findByText('Pryx Superadmin Dashboard');
        fireEvent.click(screen.getByRole('button', { name: /^Devices/ }));
        expect(await screen.findByText('Device Fleet')).toBeInTheDocument();
    });

    it('switches to system health tab when clicked', async () => {
        render(<SuperadminDashboard onLogout={mockOnLogout} />);
        await screen.findByText('Pryx Superadmin Dashboard');
        const healthButtons = screen.getAllByRole('button', { name: 'System Health' });
        fireEvent.click(healthButtons[0]);
        expect(await screen.findByText('HEALTHY')).toBeInTheDocument();
    });
});
