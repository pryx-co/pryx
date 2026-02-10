# Pryx Web (Astro + Cloudflare)

## API Routing Strategy

- Canonical API entrypoint: `src/pages/api/[...path].ts`
- Admin API router module: `src/server/admin-api.ts` (mounted under `/api/admin/*`)
- Deploy entrypoint: `@astrojs/cloudflare/entrypoints/server` in `wrangler.toml`

This avoids split runtime behavior between a custom Worker and Astro routes. In both local Astro development and Cloudflare deployment, API traffic is served through the Astro Cloudflare adapter path.

## Docs Architecture

- ADR: `apps/web/docs/adr/0001-docs-architecture.md`
- Decision: ship docs in Astro under `/docs` now, keep migration path open for Docusaurus later.

## Telemetry Pipeline

- Ingest endpoint: `/api/telemetry/ingest`
- Query endpoint: `/api/telemetry/query` (supports `level`, `category`, `device_id`, `session_id`, `start`, `end`, `limit`)
- Admin query endpoint: `/api/admin/telemetry` (same core filters)
- Retention: events stored in KV with 7-day TTL (`604800` seconds)
- PII redaction: email/API key/credit card/phone patterns are redacted from string fields before persistence

## Deployment Notes

- Build before deploy: `bun run build`
- Wrangler serves the built worker from Astro entrypoint and static assets from `./dist`.
- Install endpoint route: `/install` serves the repository canonical installer `install.sh` as raw shell script.
- Deploy smoke test: `bun run test:smoke:install -- https://pryx.dev/install`
- Full deployment runbook: `apps/web/docs/deployment.md`
- Post-deploy checks: `bun run test:smoke:deploy -- https://pryx.dev`

```sh
bun create astro@latest -- --template basics
```

> 🧑‍🚀 **Seasoned astronaut?** Delete this file. Have fun!

## 🚀 Project Structure

Inside of your Astro project, you'll see the following folders and files:

```text
/
├── public/
│   └── favicon.svg
├── src
│   ├── assets
│   │   └── astro.svg
│   ├── components
│   │   └── Welcome.astro
│   ├── layouts
│   │   └── Layout.astro
│   └── pages
│       └── index.astro
└── package.json
```

To learn more about the folder structure of an Astro project, refer to [our guide on project structure](https://docs.astro.build/en/basics/project-structure/).

## 🧞 Commands

All commands are run from the root of the project, from a terminal:

| Command                   | Action                                           |
| :------------------------ | :----------------------------------------------- |
| `bun install`             | Installs dependencies                            |
| `bun dev`             | Starts local dev server at `localhost:4321`      |
| `bun build`           | Build your production site to `./dist/`          |
| `bun preview`         | Preview your build locally, before deploying     |
| `bun astro ...`       | Run CLI commands like `astro add`, `astro check` |
| `bun astro -- --help` | Get help using the Astro CLI                     |

## 👀 Want to learn more?

Feel free to check [our documentation](https://docs.astro.build) or jump into our [Discord server](https://astro.build/chat).
