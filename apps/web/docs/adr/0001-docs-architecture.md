# ADR 0001: Web Documentation Architecture

- Status: Accepted
- Date: 2026-02-09
- Owner: Pryx Web + Docs

## Context

`apps/web` currently has no first-party docs surface. We need an approach that ships quickly, keeps docs close to product changes, and still allows future Docusaurus adoption without breaking public URLs.

## Options Considered

### Option A: Astro-native docs in `apps/web`

Use Astro content collections + markdown/mdx routes in the same deployment unit as web app and API.

Pros:
- Fastest time-to-live because no new app/repo required
- Shared components, style system, and deployment pipeline
- Versioned with product code changes in one PR

Cons:
- Fewer built-in docs features than Docusaurus (versioning/search plugins ecosystem)
- Additional custom work for advanced docs UX

### Option B: Dedicated Docusaurus docs app now

Create separate docs site immediately (likely under `docs.pryx.dev`).

Pros:
- Mature docs-focused tooling out of the box
- Strong plugin ecosystem for docs-only workflows

Cons:
- Slower initial launch (new app, infra, ownership boundaries)
- Cross-repo/app drift risk with product changes
- Additional deployment and auth/navigation integration work

## Decision

Adopt **Option A now**, with a migration-safe architecture for **Option B later** if needed.

1. Launch docs from `apps/web` under `/docs`.
2. Keep URL contract stable around `/docs/*` regardless of future implementation.
3. If Docusaurus is needed later, move rendering backend while preserving URL semantics through redirects/proxying.

## URL Strategy

- Canonical public docs URL now: `https://pryx.dev/docs/*`
- If Docusaurus is introduced later:
  - Preferred: keep canonical `pryx.dev/docs/*` via reverse proxy/worker routing.
  - Optional secondary host: `docs.pryx.dev` (non-canonical mirror or authoring environment).

## Migration Plan (Astro -> Docusaurus if required)

1. Build docs IA/content model in Astro first (`/docs`, `/docs/getting-started`, `/docs/install`, etc.).
2. Add content front matter conventions compatible with future migration.
3. Introduce Docusaurus app behind internal or staging domain.
4. Cut over by routing `pryx.dev/docs/*` to Docusaurus while preserving path structure and adding 301 redirects for any changed slugs.

## Ownership Model

- Docs platform ownership: Web team
- Content ownership: Feature owners per area (install/auth/dashboard/api)
- Release gate: PRs that change install/auth/API behavior must update corresponding docs pages before merge.

## Follow-up Implementation Tasks

1. Create `/docs` route and docs layout in `apps/web`.
2. Publish initial docs pages: install, quickstart, auth/device flow, troubleshooting.
3. Add docs navigation + search baseline.
4. Add docs contribution guide and page ownership metadata.
5. Add URL/redirect policy tests for docs route integrity.
