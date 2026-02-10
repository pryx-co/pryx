import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import Dashboard from './Dashboard';

class MockWebSocket {
    onopen: (() => void) | null = null;
    onclose: (() => void) | null = null;
    onmessage: ((event: { data: string }) => void) | null = null;
    close = vi.fn();
}

vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);

describe('Dashboard', () => {
    it('renders without crashing', () => {
        const { container } = render(<Dashboard />);
        expect(container).toBeDefined();
    });

    it('shows offline state initially', () => {
        render(<Dashboard />);
        expect(screen.getByText('Offline')).toBeInTheDocument();
    });

    it('displays header title', () => {
        render(<Dashboard />);
        expect(screen.getByText('Pryx Cloud')).toBeInTheDocument();
    });
});
