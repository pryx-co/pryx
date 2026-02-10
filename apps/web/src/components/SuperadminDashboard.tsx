// Superadmin Dashboard for Pryx - Global telemetry and fleet management
// This component provides maintainers with visibility into all users and devices

import { useState, useEffect } from 'react';

// Types for global telemetry
type GlobalStats = {
  totalUsers: number;
  activeUsers: number;
  newUsersToday: number;
  totalDevices: number;
  onlineDevices: number;
  offlineDevices: number;
  totalSessions: number;
  totalCost: number;
  avgCostPerUser: number;
};

type UserSummary = {
  id: string;
  email: string;
  createdAt: string;
  lastActive: string;
  deviceCount: number;
  sessionCount: number;
  totalCost: number;
  status: 'active' | 'inactive' | 'suspended';
};

type DeviceFleet = {
  id: string;
  userId: string;
  userEmail: string;
  name: string;
  platform: string;
  version: string;
  status: 'online' | 'offline' | 'syncing';
  lastSeen: string;
  ipAddress?: string;
};

type SystemHealth = {
  runtimeStatus: 'healthy' | 'degraded' | 'critical';
  apiLatency: number;
  errorRate: number;
  dbStatus: 'connected' | 'disconnected';
  queueDepth: number;
  activeConnections: number;
};

// Props for the dashboard
interface SuperadminDashboardProps {
  onLogout: () => void;
}

export default function SuperadminDashboard(props: SuperadminDashboardProps) {
  const [activeTab, setActiveTab] = useState<'overview' | 'users' | 'devices' | 'costs' | 'health'>('overview');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Data states
  const [globalStats, setGlobalStats] = useState<GlobalStats | null>(null);
  const [users, setUsers] = useState<UserSummary[]>([]);
  const [devices, setDevices] = useState<DeviceFleet[]>([]);
  const [health, setHealth] = useState<SystemHealth | null>(null);
  const [timeRange, setTimeRange] = useState<'24h' | '7d' | '30d' | '90d'>('7d');

  const loadDashboardData = async () => {
    setLoading(true);
    setError(null);

    try {
      const authToken = getCookieValue('auth_token');
      const headers = authToken
        ? { Authorization: `Bearer ${decodeURIComponent(authToken)}` }
        : undefined;

      // Fetch global stats
      const statsRes = await fetch('/api/admin/stats', { headers });
      if (!statsRes.ok) throw new Error('Failed to load global stats');
      setGlobalStats(await statsRes.json());

      // Fetch users list
      const usersRes = await fetch(`/api/admin/users?range=${timeRange}`, { headers });
      if (!usersRes.ok) throw new Error('Failed to load users');
      setUsers(await usersRes.json());

      // Fetch device fleet
      const devicesRes = await fetch('/api/admin/devices', { headers });
      if (!devicesRes.ok) throw new Error('Failed to load devices');
      setDevices(await devicesRes.json());

      // Fetch system health
      const healthRes = await fetch('/api/admin/health', { headers });
      if (!healthRes.ok) throw new Error('Failed to load health status');
      setHealth(await healthRes.json());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  // Intentionally not adding loadDashboardData to deps - it's recreated on each render
  // and timeRange dependency ensures data is refetched when it changes
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    loadDashboardData();
    const interval = setInterval(loadDashboardData, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, [timeRange]);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(value);
  };

  const formatNumber = (value: number) => {
    return new Intl.NumberFormat('en-US').format(value);
  };

  return (
    <div className="superadmin-dashboard">
      {/* Header */}
      <header className="dashboard-header">
        <div className="header-left">
          <h1>Pryx Superadmin Dashboard</h1>
          <span className="environment-badge">Production</span>
        </div>
        <div className="header-right">
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value as any)}
            className="time-range-select"
          >
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="90d">Last 90 Days</option>
          </select>
          <button type="button" onClick={props.onLogout} className="logout-btn">
            Logout
          </button>
        </div>
      </header>

      {/* Navigation */}
      <nav className="dashboard-nav">
        <button type="button"
          className={activeTab === 'overview' ? 'active' : ''}
          onClick={() => setActiveTab('overview')}
        >
          Overview
        </button>
        <button type="button"
          className={activeTab === 'users' ? 'active' : ''}
          onClick={() => setActiveTab('users')}
        >
          Users ({users.length})
        </button>
        <button type="button"
          className={activeTab === 'devices' ? 'active' : ''}
          onClick={() => setActiveTab('devices')}
        >
          Devices ({devices.length})
        </button>
        <button type="button"
          className={activeTab === 'costs' ? 'active' : ''}
          onClick={() => setActiveTab('costs')}
        >
          Costs
        </button>
        <button type="button"
          className={activeTab === 'health' ? 'active' : ''}
          onClick={() => setActiveTab('health')}
        >
          System Health
        </button>
      </nav>

      {/* Error Display */}
      {error && (
        <div className="error-banner">
          <span className="error-icon">⚠️</span>
          {error}
          <button type="button" onClick={loadDashboardData} className="retry-btn">
            Retry
          </button>
        </div>
      )}

      {/* Loading State */}
      {loading && !globalStats && (
        <div className="loading-overlay">
          <div className="spinner"></div>
          <p>Loading dashboard data...</p>
        </div>
      )}

      {/* Overview Tab */}
      {activeTab === 'overview' && globalStats && (
        <div className="tab-content overview">
          <div className="stats-grid">
            <div className="stat-card users">
              <h3>Total Users</h3>
              <div className="stat-value">{formatNumber(globalStats.totalUsers)}</div>
              <div className="stat-detail">
                {formatNumber(globalStats.activeUsers)} active
                <span className="separator">|</span>
                {formatNumber(globalStats.newUsersToday)} new today
              </div>
            </div>

            <div className="stat-card devices">
              <h3>Total Devices</h3>
              <div className="stat-value">{formatNumber(globalStats.totalDevices)}</div>
              <div className="stat-detail">
                {formatNumber(globalStats.onlineDevices)} online
                <span className="separator">|</span>
                {formatNumber(globalStats.offlineDevices)} offline
              </div>
            </div>

            <div className="stat-card sessions">
              <h3>Total Sessions</h3>
              <div className="stat-value">{formatNumber(globalStats.totalSessions)}</div>
            </div>

            <div className="stat-card costs">
              <h3>Total Cost</h3>
              <div className="stat-value">{formatCurrency(globalStats.totalCost)}</div>
              <div className="stat-detail">
                {formatCurrency(globalStats.avgCostPerUser)} avg/user
              </div>
            </div>
          </div>

            <div className="quick-links">
              <h3>Quick Actions</h3>
              <div className="action-buttons">
                <button type="button" onClick={() => setActiveTab('users')}>
                  View All Users
                </button>
                <button type="button" onClick={() => setActiveTab('devices')}>
                  View Device Fleet
                </button>
                <button type="button" onClick={() => setActiveTab('health')}>
                  System Health
                </button>
                <button type="button" onClick={loadDashboardData}>
                  Refresh Data
                </button>
              </div>
            </div>
        </div>
      )}

      {/* Users Tab */}
      {activeTab === 'users' && (
        <div className="tab-content users-list">
          <div className="section-header">
            <h2>User Management</h2>
            <div className="filters">
              <input
                type="text"
                placeholder="Search users..."
                className="search-input"
              />
              <select className="status-filter">
                <option value="all">All Status</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="suspended">Suspended</option>
              </select>
            </div>
          </div>

          <table className="data-table">
            <thead>
              <tr>
                <th>Email</th>
                <th>Created</th>
                <th>Last Active</th>
                <th>Devices</th>
                <th>Sessions</th>
                <th>Cost</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id}>
                  <td>{user.email}</td>
                  <td>{new Date(user.createdAt).toLocaleDateString()}</td>
                  <td>{new Date(user.lastActive).toLocaleString()}</td>
                  <td>{user.deviceCount}</td>
                  <td>{user.sessionCount}</td>
                  <td>{formatCurrency(user.totalCost)}</td>
                  <td>
                    <span className={`status-badge ${user.status}`}>
                      {user.status}
                    </span>
                  </td>
                   <td>
                     <button type="button" className="action-btn">View</button>
                     <button type="button" className="action-btn">Edit</button>
                   </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Devices Tab */}
      {activeTab === 'devices' && (
        <div className="tab-content devices-list">
          <div className="section-header">
            <h2>Device Fleet</h2>
            <div className="fleet-summary">
              <span className="online">
                {devices.filter(d => d.status === 'online').length} Online
              </span>
              <span className="offline">
                {devices.filter(d => d.status === 'offline').length} Offline
              </span>
              <span className="syncing">
                {devices.filter(d => d.status === 'syncing').length} Syncing
              </span>
            </div>
          </div>

          <table className="data-table">
            <thead>
              <tr>
                <th>Device Name</th>
                <th>User</th>
                <th>Platform</th>
                <th>Version</th>
                <th>Status</th>
                <th>Last Seen</th>
                <th>IP Address</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {devices.map((device) => (
                <tr key={device.id}>
                  <td>{device.name}</td>
                  <td>{device.userEmail}</td>
                  <td>{device.platform}</td>
                  <td>{device.version}</td>
                  <td>
                    <span className={`status-badge ${device.status}`}>
                      {device.status}
                    </span>
                  </td>
                  <td>{new Date(device.lastSeen).toLocaleString()}</td>
                  <td>{device.ipAddress || 'N/A'}</td>
                   <td>
                     <button type="button" className="action-btn">View</button>
                     <button type="button" className="action-btn">Sync</button>
                     <button type="button" className="action-btn danger">Unpair</button>
                   </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Costs Tab */}
      {activeTab === 'costs' && (
        <div className="tab-content costs-analysis">
          <div className="section-header">
            <h2>Cost Analytics</h2>
          </div>
          <div className="cost-summary">
            <div className="cost-card">
              <h4>Total Spend ({timeRange})</h4>
              <div className="cost-value">{formatCurrency(globalStats?.totalCost || 0)}</div>
            </div>
            <div className="cost-card">
              <h4>Average per User</h4>
              <div className="cost-value">{formatCurrency(globalStats?.avgCostPerUser || 0)}</div>
            </div>
            <div className="cost-card">
              <h4>Top Spenders</h4>
              <ul className="top-spenders">
                {users
                  .sort((a, b) => b.totalCost - a.totalCost)
                  .slice(0, 5)
                  .map((user) => (
                    <li key={user.id}>
                      {user.email}: {formatCurrency(user.totalCost)}
                    </li>
                  ))}
              </ul>
            </div>
          </div>
        </div>
      )}

      {/* Health Tab */}
      {activeTab === 'health' && health && (
        <div className="tab-content system-health">
          <div className="section-header">
            <h2>System Health</h2>
            <span className={`health-status ${health.runtimeStatus}`}>
              {health.runtimeStatus.toUpperCase()}
            </span>
          </div>

          <div className="health-metrics">
            <div className="metric">
              <span className="metric-label">API Latency</span>
              <div className="metric-value">{health.apiLatency}ms</div>
              <div className="metric-bar">
                <div
                  className="metric-fill"
                  style={{ width: `${Math.min(health.apiLatency / 100, 100)}%` }}
                />
              </div>
            </div>

            <div className="metric">
              <span className="metric-label">Error Rate</span>
              <div className="metric-value">{(health.errorRate * 100).toFixed(2)}%</div>
              <div className="metric-bar">
                <div
                  className="metric-fill error"
                  style={{ width: `${Math.min(health.errorRate * 100, 100)}%` }}
                />
              </div>
            </div>

            <div className="metric">
              <span className="metric-label">Database</span>
              <div className={`metric-value ${health.dbStatus}`}>
                {health.dbStatus}
              </div>
            </div>

            <div className="metric">
              <span className="metric-label">Queue Depth</span>
              <div className="metric-value">{health.queueDepth}</div>
            </div>

            <div className="metric">
              <span className="metric-label">Active Connections</span>
              <div className="metric-value">{health.activeConnections}</div>
            </div>
          </div>
        </div>
      )}

      <style>{`
        .superadmin-dashboard {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
          max-width: 1400px;
          margin: 0 auto;
          padding: 20px;
        }

        .dashboard-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 20px;
          padding-bottom: 20px;
          border-bottom: 1px solid #e0e0e0;
        }

        .header-left {
          display: flex;
          align-items: center;
          gap: 15px;
        }

        .header-left h1 {
          margin: 0;
          font-size: 24px;
        }

        .environment-badge {
          background: #4caf50;
          color: white;
          padding: 4px 8px;
          border-radius: 4px;
          font-size: 12px;
          font-weight: bold;
        }

        .header-right {
          display: flex;
          gap: 10px;
        }

        .time-range-select {
          padding: 8px 12px;
          border: 1px solid #ddd;
          border-radius: 4px;
          background: white;
        }

        .logout-btn {
          padding: 8px 16px;
          background: #f44336;
          color: white;
          border: none;
          border-radius: 4px;
          cursor: pointer;
        }

        .dashboard-nav {
          display: flex;
          gap: 10px;
          margin-bottom: 20px;
          padding: 10px 0;
          border-bottom: 1px solid #e0e0e0;
        }

        .dashboard-nav button {
          padding: 10px 20px;
          background: #f5f5f5;
          border: 1px solid #ddd;
          border-radius: 4px;
          cursor: pointer;
          transition: all 0.2s;
        }

        .dashboard-nav button.active {
          background: #2196f3;
          color: white;
          border-color: #2196f3;
        }

        .error-banner {
          background: #ffebee;
          border: 1px solid #f44336;
          padding: 12px;
          margin-bottom: 20px;
          border-radius: 4px;
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .retry-btn {
          margin-left: auto;
          padding: 6px 12px;
          background: #f44336;
          color: white;
          border: none;
          border-radius: 4px;
          cursor: pointer;
        }

        .stats-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
          gap: 20px;
          margin-bottom: 30px;
        }

        .stat-card {
          background: white;
          padding: 20px;
          border-radius: 8px;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .stat-card h3 {
          margin: 0 0 10px 0;
          color: #666;
          font-size: 14px;
          text-transform: uppercase;
        }

        .stat-value {
          font-size: 32px;
          font-weight: bold;
          color: #333;
        }

        .stat-detail {
          margin-top: 8px;
          color: #666;
          font-size: 14px;
        }

        .separator {
          margin: 0 8px;
          color: #ddd;
        }

        .data-table {
          width: 100%;
          border-collapse: collapse;
          background: white;
          border-radius: 8px;
          overflow: hidden;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .data-table th,
        .data-table td {
          padding: 12px;
          text-align: left;
          border-bottom: 1px solid #e0e0e0;
        }

        .data-table th {
          background: #f5f5f5;
          font-weight: 600;
        }

        .status-badge {
          padding: 4px 8px;
          border-radius: 4px;
          font-size: 12px;
          font-weight: 500;
        }

        .status-badge.active {
          background: #e8f5e9;
          color: #4caf50;
        }

        .status-badge.inactive {
          background: #fff3e0;
          color: #ff9800;
        }

        .status-badge.suspended {
          background: #ffebee;
          color: #f44336;
        }

        .status-badge.online {
          background: #e8f5e9;
          color: #4caf50;
        }

        .status-badge.offline {
          background: #fafafa;
          color: #9e9e9e;
        }

        .status-badge.syncing {
          background: #e3f2fd;
          color: #2196f3;
        }

        .action-btn {
          padding: 6px 12px;
          margin-right: 5px;
          background: #f5f5f5;
          border: 1px solid #ddd;
          border-radius: 4px;
          cursor: pointer;
          font-size: 12px;
        }

        .action-btn.danger {
          background: #ffebee;
          color: #f44336;
          border-color: #f44336;
        }

        .section-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 20px;
        }

        .filters {
          display: flex;
          gap: 10px;
        }

        .search-input {
          padding: 8px 12px;
          border: 1px solid #ddd;
          border-radius: 4px;
          min-width: 200px;
        }

        .fleet-summary {
          display: flex;
          gap: 20px;
        }

        .fleet-summary span {
          font-weight: 500;
        }

        .fleet-summary .online { color: #4caf50; }
        .fleet-summary .offline { color: #9e9e9e; }
        .fleet-summary .syncing { color: #2196f3; }

        .health-metrics {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 20px;
        }

        .metric {
          background: white;
          padding: 20px;
          border-radius: 8px;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .metric label {
          display: block;
          color: #666;
          font-size: 12px;
          margin-bottom: 8px;
          text-transform: uppercase;
        }

        .metric-value {
          font-size: 24px;
          font-weight: bold;
        }

        .metric-value.connected {
          color: #4caf50;
        }

        .metric-value.disconnected {
          color: #f44336;
        }

        .metric-bar {
          height: 4px;
          background: #e0e0e0;
          border-radius: 2px;
          margin-top: 10px;
          overflow: hidden;
        }

        .metric-fill {
          height: 100%;
          background: #4caf50;
          transition: width 0.3s;
        }

        .metric-fill.error {
          background: #f44336;
        }

        .health-status {
          padding: 8px 16px;
          border-radius: 4px;
          font-weight: bold;
        }

        .health-status.healthy {
          background: #e8f5e9;
          color: #4caf50;
        }

        .health-status.degraded {
          background: #fff3e0;
          color: #ff9800;
        }

        .health-status.critical {
          background: #ffebee;
          color: #f44336;
        }

        .quick-links {
          background: white;
          padding: 20px;
          border-radius: 8px;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .action-buttons {
          display: flex;
          gap: 10px;
          margin-top: 15px;
        }

        .action-buttons button {
          padding: 10px 20px;
          background: #2196f3;
          color: white;
          border: none;
          border-radius: 4px;
          cursor: pointer;
        }

        .loading-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(255,255,255,0.9);
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          z-index: 1000;
        }

        .spinner {
          width: 40px;
          height: 40px;
          border: 4px solid #f3f3f3;
          border-top: 4px solid #2196f3;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}

function getCookieValue(name: string): string | null {
  if (typeof document === 'undefined') return null;

  const needle = `${name}=`;
  const cookies = document.cookie.split('; ');
  for (const entry of cookies) {
    if (entry.indexOf(needle) === 0) {
      return entry.slice(needle.length);
    }
  }

  return null;
}
