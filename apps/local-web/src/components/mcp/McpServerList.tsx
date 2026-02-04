import { useState, useEffect } from 'react';

interface McpServer {
    id: string;
    name: string;
    transport: 'stdio' | 'sse';
    status: 'connected' | 'disconnected' | 'error';
    tools_count: number;
}

export default function McpServerList() {
    const [servers, setServers] = useState<McpServer[]>([]);
    const [loading, setLoading] = useState(true);
    const [showForm, setShowForm] = useState(false);
    
    // Form state
    const [name, setName] = useState('');
    const [transport, setTransport] = useState('stdio');
    const [command, setCommand] = useState('');
    const [args, setArgs] = useState('');
    const [url, setUrl] = useState('');

    const fetchServers = () => {
        setLoading(true);
        fetch('/api/mcp')
            .then(res => res.json())
            .then(data => {
                setServers(data.servers || []);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setLoading(false);
            });
    };

    useEffect(() => {
        fetchServers();
    }, []);

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to remove this MCP server?')) return;
        await fetch(`/api/mcp/${id}`, { method: 'DELETE' });
        fetchServers();
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        
        const payload: any = { name, transport };
        if (transport === 'stdio') {
            payload.command = command;
            payload.args = args.split(' ').filter(Boolean);
        } else {
            payload.url = url;
        }

        const res = await fetch('/api/mcp', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (res.ok) {
            setShowForm(false);
            setName('');
            setCommand('');
            setArgs('');
            setUrl('');
            fetchServers();
        } else {
            alert('Failed to add MCP server');
        }
    };

    return (
        <div style={{ padding: '1rem', maxWidth: '800px', margin: '0 auto', color: '#eee' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                <h1>MCP Servers</h1>
                <button 
                    onClick={() => setShowForm(!showForm)}
                    style={{
                        padding: '0.5rem 1rem',
                        background: '#3b82f6',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer'
                    }}
                >
                    {showForm ? 'Cancel' : 'Add Server'}
                </button>
            </div>

            {showForm && (
                <form onSubmit={handleSubmit} style={{ background: '#111', padding: '1.5rem', borderRadius: '8px', marginBottom: '2rem', border: '1px solid #333' }}>
                    <div style={{ marginBottom: '1rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Name</label>
                        <input 
                            type="text" 
                            value={name} 
                            onChange={e => setName(e.target.value)}
                            placeholder="filesystem"
                            required
                            style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                        />
                    </div>
                    <div style={{ marginBottom: '1rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Transport</label>
                        <select 
                            value={transport} 
                            onChange={e => setTransport(e.target.value)}
                            style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                        >
                            <option value="stdio">Stdio (Command)</option>
                            <option value="sse">SSE (HTTP)</option>
                        </select>
                    </div>

                    {transport === 'stdio' ? (
                        <>
                            <div style={{ marginBottom: '1rem' }}>
                                <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Command</label>
                                <input 
                                    type="text" 
                                    value={command} 
                                    onChange={e => setCommand(e.target.value)}
                                    placeholder="npx"
                                    required
                                    style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                                />
                            </div>
                            <div style={{ marginBottom: '1rem' }}>
                                <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Arguments</label>
                                <input 
                                    type="text" 
                                    value={args} 
                                    onChange={e => setArgs(e.target.value)}
                                    placeholder="-y @modelcontextprotocol/server-filesystem /path/to/allow"
                                    style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                                />
                            </div>
                        </>
                    ) : (
                        <div style={{ marginBottom: '1rem' }}>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>URL</label>
                            <input 
                                type="url" 
                                value={url} 
                                onChange={e => setUrl(e.target.value)}
                                placeholder="http://localhost:8000/sse"
                                required
                                style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                            />
                        </div>
                    )}

                    <button 
                        type="submit"
                        style={{
                            padding: '0.5rem 1rem',
                            background: '#10b981',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer'
                        }}
                    >
                        Connect Server
                    </button>
                </form>
            )}

            {loading ? (
                <div>Loading...</div>
            ) : (
                <div style={{ display: 'grid', gap: '1rem' }}>
                    {servers.map(server => (
                        <div key={server.id} style={{ 
                            background: '#111', 
                            padding: '1rem', 
                            borderRadius: '8px', 
                            border: '1px solid #333',
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center'
                        }}>
                            <div>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                    <h3 style={{ margin: 0 }}>{server.name}</h3>
                                    <span style={{ 
                                        width: 8, 
                                        height: 8, 
                                        borderRadius: '50%', 
                                        background: server.status === 'connected' ? '#10b981' : '#ef4444'
                                    }} />
                                </div>
                                <div style={{ display: 'flex', gap: '0.5rem', fontSize: '0.875rem', marginTop: '0.5rem', color: '#9ca3af' }}>
                                    <span>{server.transport.toUpperCase()}</span>
                                    <span>•</span>
                                    <span>{server.tools_count} Tools</span>
                                </div>
                            </div>
                            <button 
                                onClick={() => handleDelete(server.id)}
                                style={{
                                    padding: '0.25rem 0.75rem',
                                    background: 'rgba(239, 68, 68, 0.1)',
                                    border: '1px solid #ef4444',
                                    color: '#ef4444',
                                    borderRadius: '4px',
                                    cursor: 'pointer'
                                }}
                            >
                                Remove
                            </button>
                        </div>
                    ))}
                    {servers.length === 0 && (
                        <div style={{ padding: '2rem', textAlign: 'center', color: '#6b7280', background: '#111', borderRadius: '8px' }}>
                            No MCP servers connected. Add one to extend capabilities.
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}