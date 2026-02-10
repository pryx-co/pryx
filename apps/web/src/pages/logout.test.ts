import { describe, it, expect } from 'vitest';
import { GET } from './logout';

describe('Logout Endpoint', () => {
    const createMockUrl = (searchParams: Record<string, string> = {}) => {
        const params = new URLSearchParams(searchParams);
        return {
            searchParams: params,
            href: `https://example.com/logout?${params.toString()}`,
        };
    };

    describe('Successful Logout', () => {
        it('should clear auth cookies and redirect to /auth by default', async () => {
            const url = createMockUrl();
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
            expect(response.headers.get('Location')).toBe('/auth');

            const setCookie = response.headers.getSetCookie();
            expect(setCookie).toHaveLength(2);
            expect(setCookie[0]).toMatch(/auth_token=;.*Max-Age=0/);
            expect(setCookie[1]).toMatch(/auth_role=;.*Max-Age=0/);
        });

        it('should redirect to custom next URL when provided', async () => {
            const url = createMockUrl({ next: '/dashboard' });
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
            expect(response.headers.get('Location')).toBe('/dashboard');
        });

        it('should handle URL-encoded next parameter', async () => {
            const url = createMockUrl({ next: '/dashboard%2Fsettings' });
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
            expect(response.headers.get('Location')).toBe('/dashboard%2Fsettings');
        });
    });

    describe('Cookie Clearing', () => {
        it('should set auth_token cookie with correct attributes', async () => {
            const url = createMockUrl();
            const response = await GET({ url } as any);

            const setCookie = response.headers.getSetCookie();
            const authTokenCookie = setCookie.find(c => c.startsWith('auth_token='));

            expect(authTokenCookie).toBeDefined();
            expect(authTokenCookie).toMatch(/Path=\//);
            expect(authTokenCookie).toMatch(/Max-Age=0/);
            expect(authTokenCookie).toMatch(/SameSite=Lax/);
        });

        it('should set auth_role cookie with correct attributes', async () => {
            const url = createMockUrl();
            const response = await GET({ url } as any);

            const setCookie = response.headers.getSetCookie();
            const authRoleCookie = setCookie.find(c => c.startsWith('auth_role='));

            expect(authRoleCookie).toBeDefined();
            expect(authRoleCookie).toMatch(/Path=\//);
            expect(authRoleCookie).toMatch(/Max-Age=0/);
            expect(authRoleCookie).toMatch(/SameSite=Lax/);
        });
    });

    describe('Edge Cases', () => {
        it('should handle empty next parameter', async () => {
            const url = createMockUrl({ next: '' });
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
            expect(response.headers.get('Location')).toBe('/auth');
        });

        it('should handle external URLs in next parameter', async () => {
            const url = createMockUrl({ next: 'https://evil.com/steal' });
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
            expect(response.headers.get('Location')).toBe('/auth');
        });

        it('should reject protocol-relative next parameter', async () => {
            const url = createMockUrl({ next: '//evil.com/phish' });
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
            expect(response.headers.get('Location')).toBe('/auth');
        });

        it('should handle deeply nested paths', async () => {
            const url = createMockUrl({ next: '/admin/users/123/settings' });
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
            expect(response.headers.get('Location')).toBe('/admin/users/123/settings');
        });

        it('should handle paths with query parameters', async () => {
            const url = createMockUrl({ next: '/dashboard?tab=settings' });
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
            expect(response.headers.get('Location')).toBe('/dashboard?tab=settings');
        });
    });

    describe('Security', () => {
        it('should have empty response body', async () => {
            const url = createMockUrl();
            const response = await GET({ url } as any);

            expect(response.body).toBeNull();
        });

        it('should use 302 status for temporary redirect', async () => {
            const url = createMockUrl();
            const response = await GET({ url } as any);

            expect(response.status).toBe(302);
        });
    });
});
