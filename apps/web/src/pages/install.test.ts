import { describe, expect, it } from 'vitest';
import { GET } from './install';

describe('/install endpoint', () => {
    it('returns shell installer content with script content type', async () => {
        const response = await GET({} as any);

        expect(response.status).toBe(200);
        expect(response.headers.get('content-type')).toContain('text/x-shellscript');
        expect(response.headers.get('cache-control')).toContain('max-age=300');

        const body = await response.text();
        expect(body.startsWith('#!/usr/bin/env bash')).toBe(true);
        expect(body.toLowerCase()).not.toContain('<html');
    });
});
