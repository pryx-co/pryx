import { useState, useEffect } from 'react';
import DeviceList from './DeviceList';

interface TraceEvent {
    id: string;
    type: 'tool_call' | 'approval' | 'message' | 'error';
    name: string;
    startTime: number;
    endTime?: number;
    status: 'running' | 'done' | 'error';
    duration?: number;
    correlationId?: string;
    error?: string;
}

interface SessionStats {
    sessionId: string;
    cost: number;
    tokens: number;
    duration: number;
    eventCount: number;
}

export default function Dashboard() {
    const [events, setEvents] = useState<TraceEvent[]>([]);
    const [stats, setStats] = useState<SessionStats | null>(null);
    const [selectedEvent, setSelectedEvent] = useState<TraceEvent | null>(null);
    const [connected, setConnected] = useState(false);

    const fetchStats = async () => {
        try {
            const res = await fetch('/api/cost/summary');
            if (res.ok) {
                const data = await res.json();
                setStats(data);
            }
        } catch (err) {
            console.error('Failed to fetch stats:', err);
        }
    };

    const fetchAudit = async () => {
        try {
            const res = await fetch('/api/audit/logs');
            if (res.ok) {
                const data = await res.json();
                if (data.entries) {
                    setEvents(data.entries.map((e: any) => ({
                        id: e.id,
                        type: e.action.split('.')[1] || 'message',
                        name: e.tool || e.description || e.action,
                        startTime: new Date(e.timestamp).getTime(),
                        endTime: e.duration_ms ? new Date(e.timestamp).getTime() + e.duration_ms : undefined,
                        status: e.success ? 'done' : 'error',
                        duration: e.duration_ms,
                        correlationId: e.session_id,
                        error: e.error
                    })));
                }
            }
        } catch (err) {
            console.error('Failed to fetch audit:', err);
        }
    };

    useEffect(() => {
        fetchStats();
        fetchAudit();

        const ws = new WebSocket('ws://localhost:42424/ws');

        ws.onopen = () => setConnected(true);
        ws.onclose = () => setConnected(false);

        ws.onmessage = (msg) => {
            try {
                const evt = JSON.parse(msg.data);
                if (evt.event === 'trace.event') {
                    setEvents(prev => [...prev.slice(-99), evt.payload]);
                } else if (evt.event === 'session.stats') {
                    setStats(evt.payload);
                }
            } catch { }
        };

        return () => ws.close();
    }, []);

    const formatDuration = (ms?: number) => {
        if (!ms) return '-';
        if (ms < 1000) return `${ms}ms`;
        return `${(ms / 1000).toFixed(2)}s`;
    };

    const formatCost = (cost: number) => `$${cost.toFixed(4)}`;

    const getEventColor = (event: TraceEvent) => {
        if (event.status === 'error') return '#ef4444';
        if (event.status === 'running') return '#f59e0b';
        switch (event.type) {
            case 'tool_call': return '#3b82f6';
            case 'approval': return '#8b5cf6';
            case 'message': return '#10b981';
            default: return '#6b7280';
        }
    };

    const timelineWidth = 600;
    const now = Date.now();
    const timeWindow = 60000;

    return (
        <div style={{ padding: '1rem', fontFamily: 'system-ui', maxWidth: '1200px', margin: '0 auto' }}>
            <header style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '2rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem' }}>
                    <h1 style={{ margin: 0, fontSize: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <span>⚡</span> Pryx Local
                    </h1>
                    <nav style={{ display: 'flex', gap: '1rem', fontSize: '0.9rem' }}>
                        <a href="/" style={{ color: '#fff', textDecoration: 'none', fontWeight: 'bold' }}>Dashboard</a>
                        <a href="/skills" style={{ color: '#9ca3af', textDecoration: 'none' }}>Skills</a>
                        <a href="/settings" style={{ color: '#9ca3af', textDecoration: 'none' }}>Settings</a>
                        <a href="/channels" style={{ color: '#9ca3af', textDecoration: 'none' }}>Channels</a>
                        <a href="/mcp" style={{ color: '#9ca3af', textDecoration: 'none' }}>MCP</a>
                        <a href="/policies" style={{ color: '#9ca3af', textDecoration: 'none' }}>Policies</a>
                    </nav>
                </div>
                <span style={{
                    color: connected ? '#10b981' : '#ef4444',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem'
                }}>
                    <span style={{
                        width: 8,
                        height: 8,
                        borderRadius: '50%',
                        backgroundColor: connected ? '#10b981' : '#ef4444'
                    }} />
                    {connected ? 'Live' : 'Offline'}
                </span>
            </header>

            <DeviceList />

            {
                stats && (
                    <div style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(4, 1fr)',
                        gap: '1rem',
                        marginBottom: '1.5rem'
                    }}>
                        <StatCard label="Cost" value={formatCost(stats.cost)} color="#f59e0b" />
                        <StatCard label="Tokens" value={stats.tokens.toLocaleString()} color="#3b82f6" />
                        <StatCard label="Duration" value={formatDuration(stats.duration)} color="#10b981" />
                        <StatCard label="Events" value={stats.eventCount.toString()} color="#8b5cf6" />
                    </div>
                )
            }

            <section>
                <h2 style={{ fontSize: '1rem', marginBottom: '0.5rem' }}>Trace Timeline</h2>
                <div style={{
                    border: '1px solid #333',
                    borderRadius: 8,
                    padding: '1rem',
                    backgroundColor: '#111'
                }}>
                    <svg width={timelineWidth} height={Math.max(events.length * 24 + 20, 100)} role="img" aria-label="Event timeline visualization">
                        {events.map((event, i) => {
                            const start = Math.max(0, (event.startTime - (now - timeWindow)) / timeWindow);
                            const end = event.endTime
                                ? Math.min(1, (event.endTime - (now - timeWindow)) / timeWindow)
                                : 1;
                            const x = start * timelineWidth;
                            const width = Math.max(4, (end - start) * timelineWidth);

                            return (
                                <g key={event.id}>
                                    <button
                                        style={{
                                            position: 'absolute',
                                            left: `${x}px`,
                                            top: `${i * 24 + 4}px`,
                                            width: `${width}px`,
                                            height: '18px',
                                            backgroundColor: getEventColor(event),
                                            border: 'none',
                                            borderRadius: '4px',
                                            cursor: 'pointer'
                                        }}
                                        type="button"
                                        onClick={() => setSelectedEvent(event)}
                                        aria-label={`View details for ${event.name}`}
                                    />
                                    <text
                                        x={x + 4}
                                        y={i * 24 + 16}
                                        fontSize={10}
                                        fill="#fff"
                                    >
                                        {event.name.slice(0, 20)}
                                    </text>
                                </g>
                            );
                        })}
                    </svg>
                </div>
            </section>

            {
                selectedEvent && (
                    <section style={{ marginTop: '1rem' }}>
                        <h2 style={{ fontSize: '1rem', marginBottom: '0.5rem' }}>Event Details</h2>
                        <div style={{
                            border: '1px solid #333',
                            borderRadius: 8,
                            padding: '1rem',
                            backgroundColor: '#111',
                            fontSize: '0.875rem'
                        }}>
                            <p><strong>ID:</strong> {selectedEvent.id}</p>
                            <p><strong>Type:</strong> {selectedEvent.type}</p>
                            <p><strong>Name:</strong> {selectedEvent.name}</p>
                            <p><strong>Status:</strong> {selectedEvent.status}</p>
                            <p><strong>Duration:</strong> {formatDuration(selectedEvent.duration)}</p>
                            {selectedEvent.correlationId && (
                                <p><strong>Correlation ID:</strong> {selectedEvent.correlationId}</p>
                            )}
                            {selectedEvent.error && (
                                <p style={{ color: '#ef4444' }}><strong>Error:</strong> {selectedEvent.error}</p>
                            )}
                        </div>
                    </section>
                )
            }

            <section style={{ marginTop: '1rem' }}>
                <h2 style={{ fontSize: '1rem', marginBottom: '0.5rem' }}>Recent Events</h2>
                <div style={{
                    border: '1px solid #333',
                    borderRadius: 8,
                    overflow: 'hidden',
                    backgroundColor: '#111'
                }}>
                    {events.slice(-10).reverse().map(event => (
                        <div
                            key={event.id}
                            style={{
                                padding: '0.5rem 1rem',
                                borderBottom: '1px solid #222',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '0.5rem',
                                fontSize: '0.875rem'
                            }}
                        >
                            <span style={{
                                width: 8,
                                height: 8,
                                borderRadius: '50%',
                                backgroundColor: getEventColor(event)
                            }} />
                            <span style={{ color: '#9ca3af' }}>{event.type}</span>
                            <span>{event.name}</span>
                            <span style={{ marginLeft: 'auto', color: '#6b7280' }}>
                                {formatDuration(event.duration)}
                            </span>
                        </div>
                    ))}
                </div>
            </section>
        </div>
    );
}

function StatCard({ label, value, color }: { label: string; value: string; color: string }) {
    return (
        <div style={{
            border: '1px solid #333',
            borderRadius: 8,
            padding: '1rem',
            backgroundColor: '#111'
        }}>
            <div style={{ fontSize: '0.75rem', color: '#9ca3af', marginBottom: '0.25rem' }}>{label}</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 'bold', color }}>{value}</div>
        </div>
    );
}
