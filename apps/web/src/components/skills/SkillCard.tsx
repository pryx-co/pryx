export interface SkillProps {
    id: string;
    name: string;
    description: string;
    emoji: string;
    enabled: boolean;
    source: 'bundled' | 'managed' | 'workspace';
    path: string;
}

export default function SkillCard({ skill, onToggle }: { skill: SkillProps; onToggle: (id: string) => void }) {
    return (
        <div style={{
            border: '1px solid #333',
            borderRadius: 8,
            padding: '1.25rem',
            backgroundColor: '#111',
            display: 'flex',
            flexDirection: 'column',
            gap: '0.75rem',
            transition: 'border-color 0.2s',
            opacity: skill.enabled ? 1 : 0.6
        }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                <span style={{ fontSize: '2rem' }}>{skill.emoji}</span>
                <button
                    onClick={() => onToggle(skill.id)}
                    style={{
                        padding: '4px 12px',
                        borderRadius: '999px',
                        border: 'none',
                        fontSize: '0.75rem',
                        fontWeight: 'bold',
                        cursor: 'pointer',
                        backgroundColor: skill.enabled ? 'rgba(16, 185, 129, 0.2)' : '#374151',
                        color: skill.enabled ? '#10b981' : '#9ca3af',
                    }}
                >
                    {skill.enabled ? 'ENABLED' : 'DISABLED'}
                </button>
            </div>
            <div>
                <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '1.1rem', fontFamily: 'monospace' }}>{skill.name}</h3>
                <p style={{ margin: 0, fontSize: '0.875rem', color: '#9ca3af', lineHeight: '1.4' }}>
                    {skill.description}
                </p>
            </div>

            <div style={{ marginTop: 'auto', paddingTop: '0.75rem', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                    <span style={{
                        fontSize: '0.7rem',
                        textTransform: 'uppercase',
                        color: skill.source === 'bundled' ? '#60a5fa' : '#c084fc',
                        border: '1px solid #333',
                        padding: '2px 6px',
                        borderRadius: 4,
                        fontWeight: 'bold'
                    }}>
                        {skill.source}
                    </span>
                </div>

                <code style={{
                    fontSize: '0.7rem',
                    color: '#6b7280',
                    background: '#000',
                    padding: '4px',
                    borderRadius: 4,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap'
                }}>
                    {skill.path}
                </code>
            </div>
        </div>
    );
}
