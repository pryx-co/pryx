import { useState, useEffect } from 'react';

interface Policy {
    id: string;
    name: string;
    description: string;
    status: 'active' | 'inactive';
}

export default function PolicyList() {
    const [policies, setPolicies] = useState<Policy[]>([]);
    const [loading, setLoading] = useState(true);
    const [showForm, setShowForm] = useState(false);
    
    // Form state
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');

    const fetchPolicies = () => {
        setLoading(true);
        fetch('/api/policies')
            .then(res => res.json())
            .then(data => {
                setPolicies(data.policies || []);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setLoading(false);
            });
    };

    useEffect(() => {
        fetchPolicies();
    }, []);

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this policy?')) return;
        await fetch(`/api/policies/${id}`, { method: 'DELETE' });
        fetchPolicies();
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const res = await fetch('/api/policies', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, description })
        });
        if (res.ok) {
            setShowForm(false);
            setName('');
            setDescription('');
            fetchPolicies();
        } else {
            alert('Failed to create policy');
        }
    };

    return (
        <div style={{ padding: '1rem', maxWidth: '800px', margin: '0 auto', color: '#eee' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                <h1>Policies</h1>
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
                    {showForm ? 'Cancel' : 'Create Policy'}
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
                            placeholder="trading-safe"
                            required
                            style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                        />
                    </div>
                    <div style={{ marginBottom: '1rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Description</label>
                        <input 
                            type="text" 
                            value={description} 
                            onChange={e => setDescription(e.target.value)}
                            placeholder="Policy for trading bot operations"
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
                        Save Policy
                    </button>
                </form>
            )}

            {loading ? (
                <div>Loading...</div>
            ) : (
                <div style={{ display: 'grid', gap: '1rem' }}>
                    {policies.map(policy => (
                        <div key={policy.id} style={{ 
                            background: '#111', 
                            padding: '1rem', 
                            borderRadius: '8px', 
                            border: '1px solid #333',
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center'
                        }}>
                            <div>
                                <h3 style={{ margin: '0 0 0.5rem 0' }}>{policy.name}</h3>
                                <p style={{ margin: 0, color: '#9ca3af', fontSize: '0.9rem' }}>{policy.description}</p>
                            </div>
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                <button 
                                    onClick={() => handleDelete(policy.id)}
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
                    {policies.length === 0 && (
                        <div style={{ padding: '2rem', textAlign: 'center', color: '#6b7280', background: '#111', borderRadius: '8px' }}>
                            No policies defined. Security defaults to 'ask for everything'.
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
