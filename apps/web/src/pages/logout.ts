import type { APIRoute } from 'astro';

function sanitizeRedirectPath(rawNext: string | null): string {
    const candidate = rawNext && rawNext.trim() ? rawNext.trim() : '/auth';
    if (!candidate.startsWith('/') || candidate.startsWith('//')) {
        return '/auth';
    }

    try {
        const parsed = new URL(candidate, 'https://pryx.dev');
        if (parsed.origin !== 'https://pryx.dev') {
            return '/auth';
        }
        return `${parsed.pathname}${parsed.search}${parsed.hash}`;
    } catch {
        return '/auth';
    }
}

export const GET: APIRoute = async ({ url }) => {
    const next = sanitizeRedirectPath(url.searchParams.get('next'));

    const headers = new Headers({
        Location: next,
    });

    headers.append('Set-Cookie', 'auth_token=; Path=/; Max-Age=0; SameSite=Lax');
    headers.append('Set-Cookie', 'auth_role=; Path=/; Max-Age=0; SameSite=Lax');

    return new Response(null, {
        status: 302,
        headers,
    });
};
