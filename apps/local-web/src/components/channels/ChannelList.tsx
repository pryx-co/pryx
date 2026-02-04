import { useState, useEffect } from 'react';

interface Channel {
    id: string;
    type: string;
    name: string;
    status: string;
}

export default function ChannelList() {
    const [channels, setChannels] = useState<Channel[]>([]);
    const [loading, setLoading] = useState(true);
    const [showForm, setShowForm] = useState(false);
    
    // Form state
    const [type, setType] = useState('telegram');
    const [name, setName] = useState('');
    const [token, setToken] = useState('');

    const fetchChannels = () => {
        setLoading(true);
        fetch('/api/channels')
            .then(res => res.json())
            .then(data => {
                setChannels(data.channels || []);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setLoading(false);
            });
    };

    useEffect(() => {
        fetchChannels();
    }, []);

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this channel?')) return;
        await fetch(`/api/channels/${id}`, { method: 'DELETE' });
        fetchChannels();
    };

    const handleTest = async (id: string) => {
        const res = await fetch(`/api/channels/${id}/test`, { method: 'POST' });
        const data = await res.json();
        alert(data.message || (data.success ? 'Test passed' : 'Test failed'));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const res = await fetch('/api/channels', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ type, name, token })
        });
        if (res.ok) {
            setShowForm(false);
            setName('');
            setToken('');
            fetchChannels();
        } else {
            alert('Failed to create channel');
        }
    };

    return (
        <div style={{ padding: '1rem', maxWidth: '800px', margin: '0 auto', color: '#eee' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                <h1>Channels</h1>
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
                    {showForm ? 'Cancel' : 'Add Channel'}
                </button>
            </div>

            {showForm && (
                <form onSubmit={handleSubmit} style={{ background: '#111', padding: '1.5rem', borderRadius: '8px', marginBottom: '2rem', border: '1px solid #333' }}>
                    <div style={{ marginBottom: '1rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Type</label>
                        <select 
                            value={type} 
                            onChange={e => setType(e.target.value)}
                            style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                        >
                            <option value="telegram">Telegram</option>
                            <option value="discord">Discord</option>
                            <option value="slack">Slack</option>
                            <option value="webhook">Webhook</option>
                        </select>
                    </div>
                    <div style={{ marginBottom: '1rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Name</label>
                        <input 
                            type="text" 
                            value={name} 
                            onChange={e => setName(e.target.value)}
                            placeholder="my-bot"
                            required
                            style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                        />
                    </div>
                    <div style={{ marginBottom: '1rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Token / URL</label>
                        <input 
                            type="password" 
                            value={token} 
                            onChange={e => setToken(e.target.value)}
                            placeholder="Secret token or Webhook URL"
                            required
                            style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                        />
                    </div>
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
                        Save Channel
                    </button>
                </form>
            )}

            {loading ? (
                <div>Loading...</div>
            ) : (
                <div style={{ display: 'grid', gap: '1rem' }}>
                    {channels.map(channel => (
                        <div key={channel.id} style={{ 
                            background: '#111', 
                            padding: '1rem', 
                            borderRadius: '8px', 
                            border: '1px solid #333',
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center'
                        }}>
                            <div>
                                <h3 style={{ margin: '0 0 0.5rem 0' }}>{channel.name}</h3>
                                <div style={{ display: 'flex', gap: '0.5rem', fontSize: '0.875rem' }}>
                                    <span style={{ 
                                        padding: '2px 6px', 
                                        borderRadius: '4px', 
                                        background: '#222', 
                                        color: '#9ca3af',
                                        textTransform: 'uppercase'
                                    }}>
                                        {channel.type}
                                    </span>
                                    <span style={{ 
                                        padding: '2px 6px', 
                                        borderRadius: '4px', 
                                        background: channel.status === 'active' ? 'rgba(16, 185, 129, 0.2)' : 'rgba(239, 68, 68, 0.2)',
                                        color: channel.status === 'active' ? '#10b981' : '#ef4444'
                                    }}>
                                        {channel.status}
                                    </span>
                                </div>
                            </div>
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                <button 
                                    onClick={() => handleTest(channel.id)}
                                    style={{
                                        padding: '0.25rem 0.75rem',
                                        background: '#222',
                                        border: '1px solid #444',
                                        color: '#eee',
                                        borderRadius: '4px',
                                        cursor: 'pointer'
                                    }}
                                >
                                    Test
                                </button>
                                <button 
                                    onClick={() => handleDelete(channel.id)}
                                    style={{
                                        padding: '0.25rem 0.75rem',
                                        background: 'rgba(239, 68, 68, 0.1)',
                                        border: '1px solid #ef4444',
                                        color: '#ef4444',
                                        borderRadius: '4px',
                                        cursor: 'pointer'
                                    }}
                                >
                                    Delete
                                </button>
                            </div>
                        </div>
                    ))}
                    {channels.length === 0 && (
                        <div style={{ padding: '2rem', textAlign: 'center', color: '#6b7280', background: '#111', borderRadius: '8px' }}>
                            No channels configured. Add one above.
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}