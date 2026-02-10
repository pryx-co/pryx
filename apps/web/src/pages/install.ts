import type { APIRoute } from 'astro';
import installScript from '../../../../install.sh?raw';

const INSTALL_CONTENT_TYPE = 'text/x-shellscript; charset=utf-8';

export const GET: APIRoute = async () => {
    return new Response(installScript, {
        headers: {
            'content-type': INSTALL_CONTENT_TYPE,
            'cache-control': 'public, max-age=300',
            'x-content-type-options': 'nosniff',
        },
    });
};
