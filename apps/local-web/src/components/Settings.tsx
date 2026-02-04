import { useState, useEffect } from 'react';

// Interfaces matching Runtime API
interface Provider {
    id: string;
    name: string;
    requires_api_key: boolean;
}

interface Model {
    id: string;
    name: string;
    provider: string;
    context_window: number;
}

interface RuntimeConfig {
    model_provider: string;
    model_name: string;
    ollama_endpoint: string;
    telemetry_enabled?: boolean; // May be missing in some runtime versions
}

export default function Settings() {
    const [config, setConfig] = useState<RuntimeConfig | null>(null);
    const [providers, setProviders] = useState<Provider[]>([]);
    const [models, setModels] = useState<Model[]>([]);
    
    const [activeTab, setActiveTab] = useState<'general' | 'models' | 'telemetry'>('general');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [configRes, providersRes, modelsRes] = await Promise.all([
                    fetch('/api/config'),
                    fetch('/api/providers'),
                    fetch('/api/models')
                ]);

                if (!configRes.ok) throw new Error('Failed to load config');
                if (!providersRes.ok) throw new Error('Failed to load providers');
                if (!modelsRes.ok) throw new Error('Failed to load models');

                const configData = await configRes.json();
                const providersData = await providersRes.json();
                const modelsData = await modelsRes.json();

                setConfig(configData);
                setProviders(providersData.providers || []);
                setModels(modelsData.models || []);
            } catch (err) {
                console.error(err);
                setError(err instanceof Error ? err.message : 'Failed to load settings');
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    if (loading) return <div style={{ padding: '2rem' }}>Loading settings...</div>;
    if (error) return <div style={{ padding: '2rem', color: '#ef4444' }}>Error: {error}</div>;
    if (!config) return <div style={{ padding: '2rem' }}>No configuration found</div>;

    const availableModels = config.model_provider 
        ? models.filter(m => m.provider === config.model_provider)
        : [];

    return (
        <div style={{ padding: '1rem', maxWidth: '800px', margin: '0 auto', color: '#eee' }}>
            <h1 style={{ marginBottom: '2rem' }}>Settings</h1>

            <div style={{ display: 'flex', gap: '1rem', marginBottom: '2rem', borderBottom: '1px solid #333' }}>
                {['general', 'models', 'telemetry'].map(tab => (
                    <button
                        key={tab}
                        onClick={() => setActiveTab(tab as any)}
                        style={{
                            background: 'none',
                            border: 'none',
                            borderBottom: activeTab === tab ? '2px solid #3b82f6' : '2px solid transparent',
                            color: activeTab === tab ? '#3b82f6' : '#9ca3af',
                            padding: '0.5rem 1rem',
                            cursor: 'pointer',
                            textTransform: 'capitalize'
                        }}
                    >
                        {tab}
                    </button>
                ))}
            </div>

            <div style={{ background: '#111', padding: '2rem', borderRadius: '8px', border: '1px solid #333' }}>
                {activeTab === 'general' && (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                        <div>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Theme</label>
                            <select 
                                style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                                defaultValue="dark"
                            >
                                <option value="dark">Dark</option>
                                <option value="light">Light</option>
                                <option value="system">System</option>
                            </select>
                        </div>
                    </div>
                )}

                {activeTab === 'models' && (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                        <div>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>AI Provider</label>
                            <select 
                                value={config.model_provider}
                                onChange={(e) => setConfig({...config, model_provider: e.target.value})}
                                style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                            >
                                {providers.map(p => (
                                    <option key={p.id} value={p.id}>{p.name}</option>
                                ))}
                            </select>
                        </div>
                        <div>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Model</label>
                            <select 
                                value={config.model_name}
                                onChange={(e) => setConfig({...config, model_name: e.target.value})}
                                style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                            >
                                {availableModels.map(m => (
                                    <option key={m.id} value={m.id}>{m.name} ({m.context_window/1000}k ctx)</option>
                                ))}
                                {!availableModels.find(m => m.id === config.model_name) && (
                                    <option value={config.model_name}>{config.model_name} (Custom)</option>
                                )}
                            </select>
                        </div>
                        {config.model_provider === 'ollama' && (
                            <div>
                                <label style={{ display: 'block', marginBottom: '0.5rem', color: '#9ca3af' }}>Ollama Endpoint</label>
                                <input 
                                    type="text"
                                    value={config.ollama_endpoint}
                                    onChange={(e) => setConfig({...config, ollama_endpoint: e.target.value})}
                                    style={{ padding: '0.5rem', width: '100%', background: '#222', color: '#fff', border: '1px solid #444', borderRadius: '4px' }}
                                />
                            </div>
                        )}
                    </div>
                )}

                {activeTab === 'telemetry' && (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                            <div>
                                <h3 style={{ margin: '0 0 0.5rem 0' }}>Enable Telemetry</h3>
                                <p style={{ margin: 0, fontSize: '0.9rem', color: '#9ca3af' }}>
                                    Help improve Pryx by sending anonymous usage data. PII is always redacted.
                                </p>
                            </div>
                            <input 
                                type="checkbox"
                                checked={!!config.telemetry_enabled}
                                style={{ width: '20px', height: '20px' }}
                            />
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}