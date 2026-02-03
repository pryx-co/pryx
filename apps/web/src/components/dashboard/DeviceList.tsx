
import { useState, useEffect } from 'react';
import DeviceCard from './DeviceCard.tsx';

interface Device {
    id: string;
    name: string;
    type: 'host' | 'mobile' | 'cli' | 'web';
    status: 'online' | 'offline' | 'syncing';
    lastSeen: number;
}

export default function DeviceList() {
    const [devices, setDevices] = useState<Device[]>([]);

    useEffect(() => {
        // Mock data for now
        // In real implementation, this would fetch from /api/devices or /api/sessions
        const mockDevices: Device[] = [
            { id: 'dev-12345678', name: 'MacBook Pro', type: 'host', status: 'online', lastSeen: Date.now() },
            { id: 'dev-87654321', name: 'iPhone 15', type: 'mobile', status: 'syncing', lastSeen: Date.now() - 120000 },
            { id: 'dev-cli-001', name: 'Dev Server', type: 'cli', status: 'offline', lastSeen: Date.now() - 86400000 },
        ];
        setDevices(mockDevices);
    }, []);

    return (
        <section style={{ marginBottom: '2rem' }}>
            <h2 style={{ fontSize: '1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <span>☁️</span> Cloud Devices
            </h2>
            <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
                gap: '1rem'
            }}>
                {devices.map(dev => (
                    <DeviceCard key={dev.id} device={dev} />
                ))}
                {devices.length === 0 && (
                    <div style={{ color: '#6b7280', fontStyle: 'italic' }}>No devices found.</div>
                )}
            </div>
        </section>
    );
}
