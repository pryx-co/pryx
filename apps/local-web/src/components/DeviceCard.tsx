interface DeviceProps {
    id: string;
    name: string;
    type: 'host' | 'mobile' | 'cli' | 'web';
    status: 'online' | 'offline' | 'syncing';
    lastSeen: number;
}

export default function DeviceCard({ device }: { device: DeviceProps }) {
    const getStatusColor = (status: string) => {
        switch (status) {
            case 'online': return '#10b981';
            case 'syncing': return '#3b82f6';
            case 'offline': return '#6b7280';
            default: return '#6b7280';
        }
    };

    const getTypeIcon = (type: string) => {
        switch (type) {
            case 'host': return '🖥️';
            case 'mobile': return '📱';
            case 'cli': return '⌨️';
            case 'web': return '🌐';
            default: return '❓';
        }
    };

    const formatTime = (ts: number) => {
        const diff = Date.now() - ts;
        if (diff < 60000) return 'Just now';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
        return new Date(ts).toLocaleTimeString();
    };

    return (
        <div style={{
            border: '1px solid #333',
            borderRadius: 8,
            padding: '1rem',
            backgroundColor: '#111',
            display: 'flex',
            flexDirection: 'column',
            gap: '0.5rem',
            minWidth: '200px'
        }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '1.5rem' }}>{getTypeIcon(device.type)}</span>
                <span style={{
                    fontSize: '0.75rem',
                    color: getStatusColor(device.status),
                    border: `1px solid ${getStatusColor(device.status)}`,
                    padding: '2px 6px',
                    borderRadius: '12px'
                }}>
                    {device.status.toUpperCase()}
                </span>
            </div>
            <div style={{ fontWeight: 'bold' }}>{device.name}</div>
            <div style={{ fontSize: '0.75rem', color: '#9ca3af' }}>ID: {device.id.slice(0, 8)}...</div>
            <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>seen {formatTime(device.lastSeen)}</div>
        </div>
    );
}
