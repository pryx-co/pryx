import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SkillCard, { type SkillProps } from './SkillCard';

describe('SkillCard', () => {
    const mockSkill: SkillProps = {
        id: 'skill-1',
        name: 'filesystem',
        description: 'Read and write files on the local filesystem',
        emoji: '📁',
        enabled: true,
        source: 'bundled',
        path: '/skills/filesystem',
    };

    const mockOnToggle = vi.fn();

    beforeEach(() => {
        mockOnToggle.mockClear();
    });

    describe('Rendering', () => {
        it('should render skill name', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('filesystem')).toBeInTheDocument();
        });

        it('should render skill description', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('Read and write files on the local filesystem')).toBeInTheDocument();
        });

        it('should render skill emoji', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('📁')).toBeInTheDocument();
        });

        it('should render skill path', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('/skills/filesystem')).toBeInTheDocument();
        });

        it('should render source badge', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('bundled')).toBeInTheDocument();
        });
    });

    describe('Enabled State', () => {
        it('should show ENABLED button when skill is enabled', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('ENABLED')).toBeInTheDocument();
        });

        it('should show DISABLED button when skill is disabled', () => {
            const disabledSkill = { ...mockSkill, enabled: false };
            render(<SkillCard skill={disabledSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('DISABLED')).toBeInTheDocument();
        });

        it('should have full opacity when enabled', () => {
            const { container } = render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            const card = container.firstChild as HTMLElement;
            expect(card.style.opacity).toBe('1');
        });

        it('should have reduced opacity when disabled', () => {
            const disabledSkill = { ...mockSkill, enabled: false };
            const { container } = render(<SkillCard skill={disabledSkill} onToggle={mockOnToggle} />);
            const card = container.firstChild as HTMLElement;
            expect(card.style.opacity).toBe('0.6');
        });
    });

    describe('Toggle Interaction', () => {
        it('should call onToggle with skill id when button is clicked', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            const toggleButton = screen.getByText('ENABLED');
            fireEvent.click(toggleButton);
            expect(mockOnToggle).toHaveBeenCalledWith('skill-1');
        });

        it('should call onToggle when disabled skill button is clicked', () => {
            const disabledSkill = { ...mockSkill, enabled: false };
            render(<SkillCard skill={disabledSkill} onToggle={mockOnToggle} />);
            const toggleButton = screen.getByText('DISABLED');
            fireEvent.click(toggleButton);
            expect(mockOnToggle).toHaveBeenCalledWith('skill-1');
        });
    });

    describe('Source Types', () => {
        it('should render managed source badge correctly', () => {
            const managedSkill = { ...mockSkill, source: 'managed' as const };
            render(<SkillCard skill={managedSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('managed')).toBeInTheDocument();
        });

        it('should render workspace source badge correctly', () => {
            const workspaceSkill = { ...mockSkill, source: 'workspace' as const };
            render(<SkillCard skill={workspaceSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('workspace')).toBeInTheDocument();
        });

        it('should render bundled source badge with correct color', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            const badge = screen.getByText('bundled');
            expect(badge).toHaveStyle({ color: '#60a5fa' });
        });

        it('should render managed source badge with correct color', () => {
            const managedSkill = { ...mockSkill, source: 'managed' as const };
            render(<SkillCard skill={managedSkill} onToggle={mockOnToggle} />);
            const badge = screen.getByText('managed');
            expect(badge).toHaveStyle({ color: '#c084fc' });
        });
    });

    describe('Button Styling', () => {
        it('should have green styling when enabled', () => {
            render(<SkillCard skill={mockSkill} onToggle={mockOnToggle} />);
            const button = screen.getByText('ENABLED');
            expect(button).toHaveStyle({
                backgroundColor: 'rgba(16, 185, 129, 0.2)',
                color: '#10b981',
            });
        });

        it('should have gray styling when disabled', () => {
            const disabledSkill = { ...mockSkill, enabled: false };
            render(<SkillCard skill={disabledSkill} onToggle={mockOnToggle} />);
            const button = screen.getByText('DISABLED');
            expect(button).toHaveStyle({
                backgroundColor: '#374151',
                color: '#9ca3af',
            });
        });
    });

    describe('Edge Cases', () => {
        it('should handle long descriptions gracefully', () => {
            const longDescSkill = {
                ...mockSkill,
                description: 'This is a very long description that should still be displayed properly without breaking the layout. It contains lots of information about what this skill does.',
            };
            render(<SkillCard skill={longDescSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText(longDescSkill.description)).toBeInTheDocument();
        });

        it('should handle skills without emoji (fallback)', () => {
            const noEmojiSkill = { ...mockSkill, emoji: '' };
            render(<SkillCard skill={noEmojiSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('filesystem')).toBeInTheDocument();
        });

        it('should handle very long skill names', () => {
            const longNameSkill = { ...mockSkill, name: 'very-long-skill-name-that-exceeds-normal-length' };
            render(<SkillCard skill={longNameSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText(longNameSkill.name)).toBeInTheDocument();
        });

        it('should handle special characters in path', () => {
            const specialPathSkill = { ...mockSkill, path: '/skills/test-skill_v1.0' };
            render(<SkillCard skill={specialPathSkill} onToggle={mockOnToggle} />);
            expect(screen.getByText('/skills/test-skill_v1.0')).toBeInTheDocument();
        });
    });
});
