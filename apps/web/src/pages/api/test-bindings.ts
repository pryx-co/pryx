import type { APIRoute } from 'astro';

export const GET: APIRoute = async (ctx) => {
    const locals = (ctx as any).locals || {};
    const runtime = locals.runtime || {};
    const runtimeCtx = runtime.ctx || {};

    const bindings: any = {
        attemptDirectAccess: false,
        attemptWaitUntil: false,
    };

    // Try accessing bindings directly through runtime.env
    const directEnv = runtime.env || {};
    bindings.directAccess = {
        deviceCodes: !!directEnv.DEVICE_CODES,
        tokens: !!directEnv.TOKENS,
        sessions: !!directEnv.SESSIONS,
    };

    // Try through waitUntil context
    if (typeof runtimeCtx.waitUntil === 'function') {
        bindings.attemptWaitUntil = true;
        try {
            // This might give us access to the execution context with bindings
            const promise = runtimeCtx.waitUntil(Promise.resolve({ test: true }));
            bindings.waitUntilWorks = true;
        } catch (e) {
            bindings.waitUntilError = String(e);
        }
    }

    console.log('Final bindings check:', bindings);

    return new Response(JSON.stringify(bindings, null, 2), {
        headers: { 'Content-Type': 'application/json' }
    });
};
