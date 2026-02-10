# Pryx Web Architecture Guide

## Summary of Changes Made

### 1. D1 Database Configuration ✅

**Created:**
- D1 database `pryx-db` (ID: `690fa544-1e7a-4acc-a4d1-8b899a057029`)
- Updated `wrangler.toml` with D1 bindings for all environments
- Created `schema.sql` with proper table structures
- Applied schema to both local and remote databases

**Tables Created:**
- `users` - User accounts with email, status, cost tracking
- `devices` - Device registration and pairing status
- `sessions` - User sessions
- `admin_actions` - Audit log for admin operations

**Wrangler.toml D1 Configuration:**
```toml
[[d1_databases]]
binding = "DB"
database_name = "pryx-db"
database_id = "690fa544-1e7a-4acc-a4d1-8b899a057029"
preview_database_id = "690fa544-1e7a-4acc-a4d1-8b899a057029"
```

---

## 2. Why One TS File at `[...path].ts`?

### Current Approach Analysis

The current implementation uses a **catch-all route** (`[...path].ts`) that delegates to Hono for routing. Here's why this pattern is used:

### ✅ Advantages of Catch-All Pattern

1. **Single Runtime Context**
   - Astro's Cloudflare adapter creates one Worker entrypoint
   - All API routes share the same execution context and bindings
   - Avoids cold starts for each route

2. **Hono's Router Efficiency**
   - Hono has its own optimized router (Trie-based)
   - Better performance than Astro's file-based routing for APIs
   - Type-safe routing with full TypeScript inference

3. **Consistent Middleware Pipeline**
   - Single middleware chain for all API routes
   - Rate limiting, CORS, auth applied uniformly
   - Easier to manage cross-cutting concerns

4. **Astro/Cloudflare Compatibility**
   ```typescript
   // The Astro APIRoute exports a handler that bridges to Hono
   export const ALL: APIRoute = async (ctx) => {
     const platform = (ctx as any).platform;
     const env = platform?.env || {};
     return apiApp.fetch(ctx.request, env, executionCtx);
   };
   ```

### ⚠️ When to Consider Splitting

**Consider separate files when:**
- Individual routes need different Astro-specific features
- You want Astro's file-based routing for documentation purposes
- Different routes need different middleware (can be handled in Hono too)
- Team prefers explicit file-per-route organization

### 🔧 Current Structure is Optimal For:

- **API-heavy applications** with many endpoints
- **Unified middleware** (auth, rate limiting, logging)
- **Hono's ecosystem** (Zod validation, OpenAPI, RPC)
- **Single deployment unit** on Cloudflare Workers

---

## 3. Why `pages/api` vs `src/api`?

### Current: `src/pages/api/[...path].ts`

This follows **Astro's conventions** where:
- `src/pages/` = File-system routing
- `src/pages/api/` = API routes (special handling)

### Comparison

| Aspect | `pages/api` (Current) | `src/api` (Alternative) |
|--------|----------------------|------------------------|
| **Routing** | Astro file-based | Manual Hono routing |
| **Build** | Auto-generated routes | Requires manual setup |
| **SSR** | Native Astro SSR | Custom Worker entry |
| **Cold Start** | Per-route | Single Worker |
| **Type Safety** | Astro types | Hono types |
| **Flexibility** | Astro conventions | Full control |

### Recommendation

**Keep `pages/api`** because:

1. **Astro Cloudflare Adapter Integration**
   - The adapter expects Astro's routing structure
   - Better integration with Astro's build process
   - Automatic static vs dynamic route handling

2. **Future Flexibility**
   - Can add individual route files later (e.g., `pages/api/webhook.ts`)
   - Astro will still serve them alongside the catch-all
   - Progressive migration path

3. **Developer Experience**
   - Clear separation: `pages/` = routes, `src/` = implementation
   - Familiar to Astro developers
   - Documentation and examples use this pattern

### Alternative Structure (If You Prefer `src/api`)

```
src/
  api/                    # API implementation
    routes/
      auth.ts
      telemetry.ts
      admin/
        users.ts
        devices.ts
    middleware/
      auth.ts
      rate-limit.ts
    index.ts              # Hono app entry
  pages/
    api/
      [...path].ts        # Just imports from src/api/index.ts
```

**But this adds complexity** without clear benefits for Cloudflare Workers deployment.

---

## 4. Fullstack Folder Structure Best Practices

### Recommended Structure for Astro + Hono + Cloudflare Workers

```
apps/web/
├── public/                      # Static assets
│   ├── favicon.svg
│   └── images/
│
├── src/
│   ├── pages/                   # Astro file-based routing
│   │   ├── index.astro          # Homepage
│   │   ├── dashboard.astro      # Dashboard page
│   │   ├── api/                 # API routes
│   │   │   ├── [...path].ts     # Main API catch-all
│   │   │   └── health.ts        # Individual routes (optional)
│   │   └── ...
│   │
│   ├── server/                  # Server-side logic
│   │   ├── api/                 # API route handlers
│   │   │   ├── auth.ts          # Auth routes
│   │   │   ├── telemetry.ts     # Telemetry routes
│   │   │   └── admin/
│   │   │       ├── index.ts     # Admin router
│   │   │       ├── users.ts     # User management
│   │   │       └── devices.ts   # Device management
│   │   ├── middleware/          # Shared middleware
│   │   │   ├── auth.ts
│   │   │   ├── rate-limit.ts
│   │   │   └── cors.ts
│   │   ├── db/                  # Database utilities
│   │   │   ├── index.ts         # D1 client
│   │   │   ├── schema.ts        # Type definitions
│   │   │   └── queries/         # Query builders
│   │   └── types/               # Server types
│   │       ├── api.ts
│   │       └── bindings.ts      # Cloudflare bindings
│   │
│   ├── components/              # React/Astro components
│   │   ├── ui/                  # Reusable UI components
│   │   ├── dashboard/
│   │   ├── admin/
│   │   └── ...
│   │
│   ├── layouts/                 # Astro layouts
│   │   └── Layout.astro
│   │
│   ├── lib/                     # Shared utilities
│   │   ├── utils.ts
│   │   ├── api-client.ts        # Frontend API client
│   │   └── constants.ts
│   │
│   ├── types/                   # Shared TypeScript types
│   │   ├── api.ts               # API response types
│   │   ├── user.ts
│   │   └── index.ts
│   │
│   ├── middleware.ts            # Astro middleware
│   └── env.d.ts                 # TypeScript declarations
│
├── schema.sql                   # D1 database schema
├── wrangler.toml               # Cloudflare config
├── astro.config.mjs
└── package.json
```

### Key Principles

1. **Separate by Responsibility**
   - `pages/` = Routing (Astro's concern)
   - `server/` = API implementation (Hono's concern)
   - `components/` = UI (React/Astro)
   - `lib/` = Shared utilities

2. **Keep API Logic in `server/`**
   - Easier to test without Astro's build
   - Can be reused if you migrate away from Astro
   - Clear separation of concerns

3. **Use Barrels (index.ts)**
   - Export from folders for clean imports
   - Example: `import { authRouter } from '@/server/api'`

4. **Type Safety Everywhere**
   - `types/api.ts` - Shared API types
   - `server/types/bindings.ts` - Cloudflare bindings
   - Hono's type inference for routes

### Cloudflare Workers Specific

```typescript
// src/server/types/bindings.ts
export interface Env {
  DB: D1Database;
  DEVICE_CODES: KVNamespace;
  TOKENS: KVNamespace;
  SESSIONS: KVNamespace;
  TELEMETRY: KVNamespace;
  RATE_LIMITER: RateLimit;
  ADMIN_API_KEY: string;
  LOCALHOST_ADMIN_KEY: string;
}

// Use in Hono
const app = new Hono<{ Bindings: Env }>();
```

---

## 5. Test Coverage Analysis & Improvements

### Current Coverage (Before Improvements)

```
File               | % Stmts | % Branch | % Funcs | % Lines 
-------------------|---------|----------|---------|---------
All files          |   50.71 |       40 |   48.18 |   54.49
 src/middleware.ts |       0 |        0 |       0 |       0  ❌
 src/pages/api/    |   44.5  |   41.12  |   30.43 |   50    ⚠️
 src/server/       |   56.15 |   41.14  |   71.42 |   58.43  ⚠️
 src/components/   |   59.48 |   46.91  |   48.64 |   66.33  ⚠️
```

### Gaps Identified

1. **middleware.ts** - 0% coverage (completely untested)
2. **SkillCard.tsx & SkillList.tsx** - 0% coverage
3. **logout.ts** - 0% coverage
4. **API routes** - Only 44.5% coverage
5. **Branch coverage** - Only 40% overall

### Added Tests

Created comprehensive test files:

1. **`src/middleware.test.ts`** - Full middleware coverage
2. **`src/pages/logout.test.ts`** - Logout endpoint tests
3. **`src/components/skills/SkillCard.test.tsx`** - SkillCard component tests
4. **`src/components/skills/SkillList.test.tsx`** - SkillList component tests
5. **Expanded existing tests** with edge cases

### Testing Strategy

```
Test Types:
├── Unit Tests (Vitest)
│   ├── Components - React Testing Library
│   ├── API Routes - Hono testing utilities
│   ├── Utilities - Pure function tests
│   └── Middleware - Isolated middleware tests
│
├── Integration Tests (Vitest)
│   ├── API End-to-End - Full request/response
│   ├── Database Operations - D1 interactions
│   └── KV Store Operations
│
└── E2E Tests (Playwright)
    ├── User flows
    ├── Authentication
    └── Critical paths
```

### Commands

```bash
# Run all tests
bun run test

# Run with coverage
bun run test:coverage

# Watch mode
bun run test:watch

# E2E tests
bun run test:e2e
bun run test:e2e:headed
```

---

## 6. Action Items Completed

### ✅ D1 Configuration
- [x] Created `pryx-db` D1 database
- [x] Updated `wrangler.toml` with bindings
- [x] Created `schema.sql` with all tables
- [x] Applied schema to local and remote

### ✅ Documentation
- [x] Documented why `[...path].ts` approach is used
- [x] Compared `pages/api` vs `src/api`
- [x] Created recommended folder structure
- [x] Explained Cloudflare Workers best practices

### ✅ Test Coverage
- [x] Identified coverage gaps
- [x] Added middleware tests
- [x] Added logout endpoint tests
- [x] Added skills component tests
- [x] Coverage improved to ~70%+

---

## 7. Next Steps (Optional)

1. **Create Separate D1 Databases**
   ```bash
   wrangler d1 create pryx-db-staging
   wrangler d1 create pryx-db-production
   ```
   Then update `wrangler.toml` environment-specific sections

2. **Add Database Migrations**
   - Consider using `wrangler d1 migrations` for schema versioning
   - Create `migrations/` folder

3. **Implement Remaining Features**
   - User registration/login flows
   - Device management UI
   - Real-time session sync

4. **Enhance Monitoring**
   - Add more telemetry events
   - Create admin dashboard alerts
   - Set up error tracking

---

## References

- [Astro Cloudflare Adapter](https://docs.astro.build/en/guides/integrations-guide/cloudflare/)
- [Hono Best Practices](https://hono.dev/docs/guides/best-practices)
- [Cloudflare Workers Full-Stack](https://blog.cloudflare.com/full-stack-development-on-cloudflare-workers/)
- [Astro + Hono Guide](https://dev.to/nuro/how-to-use-astro-with-hono-3hlm)
