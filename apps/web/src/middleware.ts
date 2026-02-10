import { defineMiddleware } from 'astro:middleware';

interface MiddlewareEnv {
    ADMIN_API_KEY?: string;
    LOCALHOST_ADMIN_KEY?: string;
    ENVIRONMENT?: string;
}

interface AuthContext {
    role: 'user' | 'superadmin' | 'localhost';
}

function isNonProduction(env: MiddlewareEnv): boolean {
    const normalized = (env.ENVIRONMENT ?? '').toLowerCase();
    return normalized === 'development' || normalized === 'staging' || normalized === 'test' || normalized === 'local';
}

function extractAuthContext(authToken: string | undefined, env: MiddlewareEnv): AuthContext | null {
    if (!authToken) return null;

    const token = authToken.trim();
    if (!token) return null;

    if (
        (env.ADMIN_API_KEY && token === env.ADMIN_API_KEY)
        || (env.ADMIN_API_KEY && token === `superadmin:${env.ADMIN_API_KEY}`)
    ) {
        return { role: 'superadmin' };
    }

    if ((env.LOCALHOST_ADMIN_KEY && token === env.LOCALHOST_ADMIN_KEY) || (token === 'localhost' && isNonProduction(env))) {
        return { role: 'localhost' };
    }

    if (token.startsWith('user:')) {
        return { role: 'user' };
    }

    return null;
}

export const onRequest = defineMiddleware(async (context, next) => {
    const pathname = context.url.pathname;
    const requiresAuth = pathname.startsWith('/dashboard') || pathname.startsWith('/superadmin');
    const platformEnv = ((context as any).platform?.env ?? {}) as MiddlewareEnv;

    if (requiresAuth) {
        const authToken = context.cookies.get('auth_token')?.value;
        const authContext = extractAuthContext(authToken, platformEnv);

        if (!authContext) {
            return context.redirect(`/auth?next=${encodeURIComponent(pathname)}`);
        }

        if (pathname.startsWith('/superadmin') && authContext.role !== 'superadmin' && authContext.role !== 'localhost') {
            return context.redirect('/auth?next=/superadmin&role=superadmin');
        }
    }

    // Ensure locals.runtime.env is available and contains bindings for Hono
    if (context.locals) {
        const mutableLocals = context.locals as any;
        const waitUntil = (context as any).waitUntil;

        // Set up locals.runtime.env with all bindings for Hono to access
        mutableLocals.runtime = {
            env: platformEnv,
            cf: (context as any).cf,
            caches: (context as any).caches,
            ctx: {
                waitUntil: (promise: Promise<any>) => waitUntil(promise),
                passThroughOnException: () => {},
            },
        };
    }
    return next();
});
