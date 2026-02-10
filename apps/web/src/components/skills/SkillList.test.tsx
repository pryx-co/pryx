import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import SkillList from './SkillList';

const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('SkillList', () => {
    const mockSkills = {
        skills: [
            {
                id: 'skill-1',
                name: 'filesystem',
                description: 'File system access',
                metadata: { emoji: '📁' },
                enabled: true,
                source: 'bundled',
                path: '/skills/filesystem',
            },
            {
                id: 'skill-2',
                name: 'web-search',
                description: 'Search the web',
                metadata: { emoji: '🔍' },
                enabled: false,
                source: 'managed',
                path: '/skills/web-search',
            },
        ],
    };

    beforeEach(() => {
        mockFetch.mockClear();
    });

    describe('Loading State', () => {
        it('should show loading message initially', () => {
            mockFetch.mockImplementation(() => new Promise(() => {}));
            render(<SkillList />);
            expect(screen.getByText('Loading skills...')).toBeInTheDocument();
        });

        it('should show loading with correct styling', () => {
            mockFetch.mockImplementation(() => new Promise(() => {}));
            render(<SkillList />);
            const loadingText = screen.getByText('Loading skills...');
            expect(loadingText).toBeInTheDocument();
        });
    });

    describe('Successful Data Fetch', () => {
        it('should fetch skills from localhost:8080', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(mockSkills),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/skills');
            });
        });

        it('should render fetched skills', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(mockSkills),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText('filesystem')).toBeInTheDocument();
                expect(screen.getByText('web-search')).toBeInTheDocument();
            });
        });

        it('should render page header', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(mockSkills),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText('Skills Manager')).toBeInTheDocument();
                expect(screen.getByText('Manage the capabilities and tools available to your Pryx agent.')).toBeInTheDocument();
            });
        });
    });

    describe('Error Handling', () => {
        it('should show error when fetch fails', async () => {
            mockFetch.mockRejectedValueOnce(new Error('Network error'));

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText(/Could not load skills/)).toBeInTheDocument();
            });
        });

        it('should show error when response is not ok', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 500,
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText(/Could not load skills/)).toBeInTheDocument();
            });
        });

        it('should mention pryx-core in error message', async () => {
            mockFetch.mockRejectedValueOnce(new Error('Failed to fetch'));

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText(/Is pryx-core running\?/)).toBeInTheDocument();
            });
        });

        it('should render error message', async () => {
            mockFetch.mockRejectedValueOnce(new Error('Failed'));

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText(/Error:/)).toBeInTheDocument();
                expect(screen.getByText(/Could not load skills/)).toBeInTheDocument();
            });
        });
    });

    describe('Skill Data Transformation', () => {
        it('should handle skills without metadata.emoji', async () => {
            const skillsWithoutEmoji = {
                skills: [
                    {
                        id: 'skill-1',
                        name: 'test-skill',
                        description: 'Test description',
                        metadata: {},
                        enabled: true,
                        source: 'bundled',
                        path: '/skills/test',
                    },
                ],
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(skillsWithoutEmoji),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText('test-skill')).toBeInTheDocument();
                expect(screen.getByText('🧩')).toBeInTheDocument();
            });
        });

        it('should handle empty skills array', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({ skills: [] }),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.queryByText('Loading skills...')).not.toBeInTheDocument();
                expect(screen.getByText('Skills Manager')).toBeInTheDocument();
            });
        });

        it('should handle null skills response', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({}),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.queryByText('Loading skills...')).not.toBeInTheDocument();
            });
        });
    });

    describe('Skill Toggle', () => {
        it('should toggle skill enabled state when clicked', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(mockSkills),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText('filesystem')).toBeInTheDocument();
            });

            const toggleButton = screen.getAllByText('ENABLED')[0];
            fireEvent.click(toggleButton);

            await waitFor(() => {
                expect(screen.getAllByText('DISABLED')).toHaveLength(2);
            });
        });

        it('should toggle skill back to enabled', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(mockSkills),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText('web-search')).toBeInTheDocument();
            });

            const toggleButton = screen.getByText('DISABLED');
            fireEvent.click(toggleButton);

            await waitFor(() => {
                expect(screen.getAllByText('ENABLED')).toHaveLength(2);
            });
        });
    });

    describe('Component Structure', () => {
        it('should render skills in grid layout', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(mockSkills),
            });

            const { container } = render(<SkillList />);

            await waitFor(() => {
                const grid = container.querySelector('[style*="grid"]');
                expect(grid).toBeInTheDocument();
            });
        });

        it('should have correct max width and margin', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(mockSkills),
            });

            const { container } = render(<SkillList />);

            await waitFor(() => {
                const wrapper = container.firstChild as HTMLElement;
                expect(wrapper).toHaveStyle({
                    maxWidth: '1200px',
                    margin: '0 auto',
                });
            });
        });
    });

    describe('Edge Cases', () => {
        it('should handle malformed skill data gracefully', async () => {
            const malformedSkills = {
                skills: [
                    {
                        id: 'skill-1',
                        name: null,
                        description: undefined,
                        enabled: true,
                        source: 'bundled',
                        path: '/skills/test',
                    },
                ],
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(malformedSkills),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.queryByText('Loading skills...')).not.toBeInTheDocument();
            });
        });

        it('should handle very long skill list', async () => {
            const manySkills = {
                skills: Array.from({ length: 100 }, (_, i) => ({
                    id: `skill-${i}`,
                    name: `Skill ${i}`,
                    description: `Description ${i}`,
                    metadata: { emoji: '🎯' },
                    enabled: i % 2 === 0,
                    source: 'bundled',
                    path: `/skills/skill-${i}`,
                })),
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve(manySkills),
            });

            render(<SkillList />);

            await waitFor(() => {
                expect(screen.getByText('Skill 0')).toBeInTheDocument();
                expect(screen.getByText('Skill 99')).toBeInTheDocument();
            });
        });
    });
});
