import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';
import Dashboard from './Dashboard';

vi.stubGlobal('WebSocket', vi.fn(() => ({
    onopen: null,
    onclose: null,
    onmessage: null,
    close: vi.fn(),
})));

describe('Dashboard', () => {
    it('renders without crashing', () => {
        const { container } = render(<Dashboard />);
        expect(container).toBeDefined();
    });

    it('shows disconnected state initially', () => {
        const { getByText } = render(<Dashboard />);
        expect(getByText('Offline')).toBeDefined();
    });

    it('displays header', () => {
        const { getByText } = render(<Dashboard />);
        expect(getByText('Pryx Local')).toBeDefined();
    });
});
