# Pryx Web - Production Ready MVP Checklist

## Current State Analysis

### ✅ Completed
- D1 database configured and schema applied
- 74 tests passing with good coverage
- Basic auth UI (sign in/out)
- Admin API with 3-layer auth (user, superadmin, localhost)
- Protected routes middleware

### ❌ Critical Issues for Production MVP

---

## Priority 1: Critical Production Blockers

### 1. Remove Hardcoded Localhost URLs
**Impact:** HIGH - App won't work in production

Current Issues:
- `Dashboard.tsx` → `ws://localhost:3000/ws` 
- `SkillList.tsx` → `http://localhost:8080/skills`

**Solution:**
```typescript
// Create src/lib/config.ts
export const API_BASE_URL = import.meta.env.PUBLIC_API_URL || 'https://api.pryx.dev';
export const WS_BASE_URL = import.meta.env.PUBLIC_WS_URL || 'wss://api.pryx.dev';
```

**Tasks:**
- [ ] Create environment-based config module
- [ ] Update Dashboard.tsx to use production WebSocket URL
- [ ] Update SkillList.tsx to use production API URL
- [ ] Add `.env.example` file
- [ ] Configure environment variables in Cloudflare dashboard

### 2. Proper Session Validation
**Impact:** HIGH - Security vulnerability

Current Issues:
- Auth middleware only checks cookie existence, not token validity
- No server-side token validation on protected routes

**Tasks:**
- [ ] Add `/api/auth/validate` endpoint to verify tokens
- [ ] Update middleware to validate tokens server-side
- [ ] Add token expiration handling
- [ ] Implement refresh token flow

### 3. Production Environment Configuration
**Impact:** HIGH - Deployment will fail

**Tasks:**
- [ ] Set up production secrets in Cloudflare:
  - `ADMIN_API_KEY`
  - `LOCALHOST_ADMIN_KEY` (only for non-prod)
- [ ] Configure D1 database for production environment
- [ ] Set up KV namespaces for production
- [ ] Add environment-specific CORS origins

---

## Priority 2: Core MVP Features

### 4. Real Dashboard Data (Issue: pryx-ql7.3)
**Impact:** MEDIUM - Current dashboard shows mock data

**Tasks:**
- [ ] Create `/api/dashboard/stats` endpoint
- [ ] Create `/api/dashboard/devices` endpoint
- [ ] Update Dashboard component to fetch from API
- [ ] Add loading states
- [ ] Add error handling with retry
- [ ] Remove WebSocket dependency or make it optional

### 5. User Registration/Onboarding
**Impact:** MEDIUM - Users can't sign up

**Tasks:**
- [ ] Create `/api/auth/register` endpoint
- [ ] Build registration UI page
- [ ] Add email validation
- [ ] Create user in D1 database on registration
- [ ] Send welcome email (optional for MVP)

### 6. Superadmin Dashboard Integration (Issue: pryx-ql7.4)
**Impact:** MEDIUM - Superadmin features incomplete

**Tasks:**
- [ ] Connect superadmin dashboard to real API endpoints
- [ ] Add user management (list, view, suspend)
- [ ] Add device management (list, sync, unpair)
- [ ] Add cost analytics display
- [ ] Add telemetry/logs viewer

---

## Priority 3: Infrastructure & DevOps

### 7. Error Handling & Monitoring
**Impact:** MEDIUM - Hard to debug production issues

**Tasks:**
- [ ] Add structured logging to API endpoints
- [ ] Set up Sentry or similar error tracking
- [ ] Add health check endpoint (`/api/health`)
- [ ] Create error boundary for React components
- [ ] Add user-friendly error messages

### 8. Performance Optimization
**Impact:** LOW-MEDIUM - Better UX

**Tasks:**
- [ ] Add caching headers for static assets
- [ ] Implement API response caching where appropriate
- [ ] Add loading skeletons for dashboard
- [ ] Optimize bundle size (check with `astro build`)

### 9. Security Hardening
**Impact:** MEDIUM - Security best practices

**Tasks:**
- [ ] Add rate limiting to auth endpoints
- [ ] Implement CSRF protection
- [ ] Add security headers (HSTS, CSP, etc.)
- [ ] Sanitize all user inputs
- [ ] Add SQL injection protection (use D1 prepared statements)

---

## Priority 4: Documentation & Polish

### 10. User Documentation (Issue: pryx-90r.*)
**Impact:** LOW - Helps users understand the product

**Tasks:**
- [ ] Create `/docs` route with Astro content collections
- [ ] Write quick start guide
- [ ] Document API endpoints
- [ ] Add FAQ page

### 11. SEO & Meta Tags
**Impact:** LOW - Better discoverability

**Tasks:**
- [ ] Add proper meta tags to all pages
- [ ] Create sitemap.xml
- [ ] Add robots.txt
- [ ] Set up Open Graph tags

### 12. Testing in Production
**Impact:** HIGH - Ensure everything works

**Tasks:**
- [ ] Deploy to staging environment
- [ ] Run E2E tests against staging
- [ ] Test auth flow end-to-end
- [ ] Verify all API endpoints work
- [ ] Check D1 database connectivity
- [ ] Validate WebSocket connections (if using)

---

## Immediate Next Steps (This Week)

1. **Fix localhost URLs** (2 hours)
   - Create config module
   - Update Dashboard and SkillList
   - Add environment variables

2. **Set up production secrets** (30 min)
   - Generate secure ADMIN_API_KEY
   - Add to Cloudflare dashboard
   - Test in staging

3. **Add session validation** (4 hours)
   - Create validate endpoint
   - Update middleware
   - Test auth flow

4. **Deploy to staging** (1 hour)
   - Run `wrangler deploy --env staging`
   - Test all critical paths
   - Fix any issues

---

## Files to Modify

### High Priority
1. `src/components/Dashboard.tsx` - Remove localhost WebSocket
2. `src/components/skills/SkillList.tsx` - Remove localhost API
3. `src/middleware.ts` - Add token validation
4. `src/pages/api/[...path].ts` - Add auth validate endpoint
5. `wrangler.toml` - Add production environment config

### Medium Priority
6. `src/pages/dashboard.astro` - Add real data fetching
7. `src/components/SuperadminDashboard.tsx` - Connect to real API
8. `src/pages/register.astro` - Create registration page (new)
9. `src/pages/api/` - Add registration endpoint

### Low Priority
10. `src/layouts/Layout.astro` - Add SEO meta tags
11. Create `src/pages/docs/` - Documentation pages
12. Add error tracking integration

---

## Success Criteria for Production MVP

- [ ] User can register and sign in
- [ ] Dashboard shows real data from API
- [ ] Auth tokens are validated server-side
- [ ] All localhost URLs removed
- [ ] Deployed to staging and tested
- [ ] All 74 tests passing
- [ ] No console errors in production
- [ ] Basic error handling in place

---

## Estimated Timeline

**Week 1:**
- Fix localhost URLs
- Set up production secrets
- Add session validation

**Week 2:**
- Real dashboard data
- User registration
- Deploy to staging

**Week 3:**
- Superadmin dashboard
- Error handling
- Production deploy

**Total: 3 weeks to production MVP**
