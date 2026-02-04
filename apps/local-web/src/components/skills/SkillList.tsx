import { useState, useEffect } from 'react';
import SkillCard, { type SkillProps } from './SkillCard';

export default function SkillList() {
    const [skills, setSkills] = useState<SkillProps[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchSkills = async () => {
            try {
                const res = await fetch('http://localhost:42424/api/skills');
                if (!res.ok) throw new Error('Failed to fetch skills');
                const data = await res.json();

                const mapped: SkillProps[] = (data.skills || []).map((s: any) => ({
                    id: s.id,
                    name: s.name,
                    description: s.description,
                    emoji: s.metadata?.emoji || '🧩',
                    enabled: s.enabled,
                    source: s.source,
                    path: s.path,
                }));
                setSkills(mapped);
            } catch (err) {
                console.error(err);
                setError('Could not load skills from Host API. Is pryx-host running?');
            } finally {
                setLoading(false);
            }
        };

        fetchSkills();
    }, []);

    if (loading) return <div style={{ padding: '2rem', color: '#6b7280' }}>Loading skills...</div>;
    if (error) return <div style={{ padding: '2rem', color: '#ef4444' }}>Error: {error}</div>;

    const handleToggle = (id: string) => {
        setSkills(prev => prev.map(s =>
            s.id === id ? { ...s, enabled: !s.enabled } : s
        ));
    };

    return (
        <div style={{ padding: '1rem', fontFamily: 'system-ui', maxWidth: '1200px', margin: '0 auto' }}>
            <header style={{ marginBottom: '2rem' }}>
                <h1 style={{ margin: 0, fontSize: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <span>🛠️</span> Skills Manager
                </h1>
                <p style={{ color: '#9ca3af', marginTop: '0.5rem' }}>
                    Manage the capabilities and tools available to your Pryx agent.
                </p>
                <nav style={{ marginTop: '1rem', display: 'flex', gap: '1rem', fontSize: '0.9rem' }}>
                    <a href="/" style={{ color: '#9ca3af', textDecoration: 'none' }}>Dashboard</a>
                    <span style={{ color: '#fff', fontWeight: 'bold' }}>Skills</span>
                </nav>
            </header>

            <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
                gap: '1.5rem'
            }}>
                {skills.map(skill => (
                    <SkillCard key={skill.id} skill={skill} onToggle={handleToggle} />
                ))}
            </div>
        </div>
    );
}
