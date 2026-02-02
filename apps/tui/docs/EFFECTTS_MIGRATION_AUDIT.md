# TUI EffectTS Migration Audit Report

**Project:** Pryx TUI  
**Date:** 2026-01-30  
**Commit:** d686a6dfff4eca54f64908f2ecb63fded89b9888

---

## Executive Summary

The TUI codebase has **partial EffectTS adoption**. Out of 31 TypeScript source files:

- **8 files** (26%) fully use EffectTS
- **6 files** (19%) use regular async/await (need conversion)
- **1 file** (3%) has mixed patterns (needs cleanup)
- **16 files** (52%) are pure (no async/IO operations)

**Migration Priority:** Medium - Core services use EffectTS, but several components still use raw async/await.

---

## 📊 File Categorization

### ✅ Category 1: ALREADY USING EFFECTTS (8 files)

These files have been successfully migrated to EffectTS:

| File                              | EffectTS Patterns Used                         | Status               |
| --------------------------------- | ---------------------------------------------- | -------------------- |
| `components/Chat.tsx`             | Effect.runFork, Stream.runForEach, Effect.sync | ✅ Complete          |
| `components/OnboardingWizard.tsx` | Effect.gen, Effect.tryPromise                  | ✅ Complete          |
| `components/SessionExplorer.tsx`  | Effect.gen, Effect.tryPromise                  | ✅ Complete          |
| `lib/hooks.ts`                    | Effect.gen, Runtime, ManagedRuntime            | ✅ Complete          |
| `services/config.ts`              | Effect.gen, Effect.tryPromise                  | ✅ Complete          |
| `services/ws.ts`                  | Effect.gen, Effect.async, Layer                | ✅ Complete          |
| `services/skills-api.ts`          | Effect.gen, Effect.tryPromise                  | ⚠️ Mixed (see below) |
| `test-ws.ts`                      | Effect.gen, Effect.runPromise, Effect.provide  | ✅ Complete          |

**Key Patterns Found:**

- `Effect.gen(function* () { ... })` for generator-style effects
- `Effect.tryPromise({ try: () => fetch(...), catch: (e) => ... })` for async operations
- `Effect.runFork()` for running effects without blocking
- `Stream.runForEach()` for handling streams

---

### ❌ Category 2: NOT USING EFFECTTS - NEED CONVERSION (6 files)

These files use raw async/await and Promises, violating the EffectTS architecture:

| File                             | Current Pattern                         | Impact                     | Priority  |
| -------------------------------- | --------------------------------------- | -------------------------- | --------- |
| `components/App.tsx`             | `await fetch()` for health checks       | High - Core app logic      | 🔴 High   |
| `components/ProviderManager.tsx` | `await fetch()` for providers           | High - Provider management | 🔴 High   |
| `components/SetupRequired.tsx`   | `await fetch()` for provider config     | High - Setup flow          | 🔴 High   |
| `components/Skills.tsx`          | `await fetch()` for skills API          | Medium - Skills UI         | 🟡 Medium |
| `components/Channels.tsx`        | `await handleSave()`                    | Medium - Channel settings  | 🟡 Medium |
| `hooks/useMouse.ts`              | `await navigator.clipboard.writeText()` | Low - Clipboard utility    | 🟢 Low    |

**Specific Issues Found:**

```typescript
// App.tsx - Line 58
checkStatus: async () => {
  const res = await fetch(`${apiUrl}/health`, { method: "GET" });
  const data = await res.json(); // ❌ Should use Effect.tryPromise
};

// ProviderManager.tsx - Lines 50-60
const fetchProviders = async () => {
  const response = await fetch(`${API_BASE}/api/v1/providers`);
  const data = await response.json(); // ❌ Should use Effect.tryPromise
};

// SetupRequired.tsx - Lines 35-39
const response = await fetch(`${API_BASE}/api/v1/providers`);
const data = await response.json(); // ❌ Should use Effect.tryPromise

// Skills.tsx - Lines 24-28
const res = await fetch(`${apiUrl}/skills`);
const data = await res.json(); // ❌ Should use Effect.tryPromise

// hooks/useMouse.ts - Lines 70-80
await navigator.clipboard.writeText(text); // ❌ Should use Effect.tryPromise
await proc.exited; // ❌ Should use Effect.async
```

---

### ⚠️ Category 3: MIXED PATTERNS - NEEDS CLEANUP (1 file)

These files use BOTH EffectTS and raw async/await inconsistently:

| File                     | EffectTS Usage                | Raw Async Usage                   | Issue                                       |
| ------------------------ | ----------------------------- | --------------------------------- | ------------------------------------------- |
| `services/skills-api.ts` | Effect.gen, Effect.tryPromise | `await fetch()` inside Effect.gen | Inconsistent - uses await inside generators |

**Code Smell Example:**

```typescript
// services/skills-api.ts - Lines 50-56
const fetchSkills = Effect.gen(function* () {
  const res = await fetch(`${getApiUrl()}/skills`); // ❌ Using await inside Effect.gen!
  const data = (await res.json()) as SkillsResponse; // ❌ Should use yield* Effect.tryPromise
  return data;
});
```

**Correct Pattern:**

```typescript
const fetchSkills = Effect.gen(function* () {
  const data = yield* Effect.tryPromise({
    try: () => fetch(`${getApiUrl()}/skills`).then(r => r.json()),
    catch: e => new SkillsApiError(String(e)),
  });
  return data as SkillsResponse;
});
```

---

### ✅ Category 4: PURE FILES - NO CONVERSION NEEDED (16 files)

These files have no IO operations and don't need EffectTS:

| File                                      | Description                                          |
| ----------------------------------------- | ---------------------------------------------------- |
| `components/AppHeader.tsx`                | UI component - no IO                                 |
| `components/CommandPalette.tsx`           | UI component - no IO                                 |
| `components/KeyboardShortcuts.tsx`        | UI component - no IO                                 |
| `components/Message.tsx`                  | UI component - no IO                                 |
| `components/Notifications.tsx`            | UI component - no IO                                 |
| `components/SearchableCommandPalette.tsx` | UI component - no IO                                 |
| `components/Settings.tsx`                 | UI component - uses loadConfig/saveConfig (services) |
| `hooks/useKeybind.ts`                     | Pure keyboard logic                                  |
| `lib/keybindings.ts`                      | Key binding constants                                |
| `opentui.d.ts`                            | Type definitions                                     |
| `services/ws.test.ts`                     | Test file                                            |
| `theme.ts`                                | Theme constants                                      |

---

## 🎯 Migration Roadmap

### Phase 1: High Priority (Core Components)

**Files:** `App.tsx`, `ProviderManager.tsx`, `SetupRequired.tsx`

These files handle critical app initialization and provider configuration. They should be converted to use the existing EffectTS services instead of raw fetch.

**Migration Strategy:**

1. Use existing `services/config.ts` (already EffectTS)
2. Create new EffectTS service for provider API calls
3. Replace `await fetch()` with `yield* Effect.tryPromise()`

**Example Migration:**

```typescript
// BEFORE (App.tsx)
createEffect(() => {
  const checkStatus = async () => {
    const res = await fetch(`${apiUrl}/health`);
    const data = await res.json();
    setConnectionStatus(data.status);
  };
  checkStatus();
});

// AFTER (App.tsx)
createEffect(() => {
  const checkStatus = Effect.gen(function* () {
    const data = yield* HealthService.check(apiUrl);
    yield* Effect.sync(() => setConnectionStatus(data.status));
  });
  Effect.runFork(checkStatus);
});
```

---

### Phase 2: Medium Priority (Feature Components)

**Files:** `Skills.tsx`, `Channels.tsx`

These files should use the existing `skills-api.ts` service (after fixing the mixed patterns).

**Migration Strategy:**

1. Fix `skills-api.ts` to remove raw async/await
2. Update `Skills.tsx` to use the EffectTS service
3. Update `Channels.tsx` similarly

---

### Phase 3: Low Priority (Utilities)

**Files:** `hooks/useMouse.ts`

The clipboard operation is low priority but should still be converted for consistency.

**Migration Strategy:**

1. Create a ClipboardService using EffectTS
2. Replace `await navigator.clipboard` with Effect.tryPromise

---

## 🔍 Migration Blockers

### Blocker 1: Inconsistent Error Handling

**Issue:** Some files catch errors with try/catch, others use Effect.catch
**Impact:** Error handling patterns are inconsistent across the codebase
**Resolution:** Standardize on Effect.catch and Effect.tryPromise error handling

### Blocker 2: SolidJS Integration

**Issue:** SolidJS effects (createEffect) don't naturally work with EffectTS
**Current Pattern:**

```typescript
createEffect(() => {
  // How to properly integrate EffectTS here?
});
```

**Solution:** Use `Effect.runFork()` inside createEffect, as done in `Chat.tsx`:

```typescript
createEffect(() => {
  const fiber = Effect.runFork(effect);
  return () => Effect.runSync(Fiber.interrupt(fiber));
});
```

### Blocker 3: Service Dependencies

**Issue:** Components directly call fetch instead of using services
**Resolution:** Ensure all API calls go through EffectTS services (config.ts, ws.ts, skills-api.ts)

---

## 📈 Statistics

| Metric                       | Count | Percentage |
| ---------------------------- | ----- | ---------- |
| Total TS/TSX Files           | 31    | 100%       |
| Fully EffectTS               | 7     | 23%        |
| Mixed (needs cleanup)        | 1     | 3%         |
| Raw Async (needs conversion) | 6     | 19%        |
| Pure (no IO)                 | 16    | 52%        |
| JS/JSX Compiled Files        | 20    | -          |

**Adoption Rate:** 26% of IO-related files use EffectTS (7 out of 27)

---

## ✅ Files Verified (Not Compiled Output)

**Source Files (.ts, .tsx):**

- All `.ts` and `.tsx` files in the list are source files
- `.js` and `.jsx` files are compiled output (generated by Bun build process)
- The compiled files show EffectTS patterns as `effect_1.Effect.gen` etc.

**Key Source Files:**

```
components/*.tsx (16 files)
services/*.ts (5 files)
hooks/*.ts (2 files)
lib/*.ts (2 files)
theme.ts
test-ws.ts
opentui.d.ts
```

---

## 🎓 Migration Examples

### Converting Fetch to EffectTS

**Before:**

```typescript
const fetchData = async () => {
  try {
    const res = await fetch("/api/data");
    if (!res.ok) throw new Error("Failed");
    const data = await res.json();
    return data;
  } catch (e) {
    console.error(e);
    return null;
  }
};
```

**After:**

```typescript
import { Effect } from "effect";

class ApiError {
  readonly _tag = "ApiError";
  constructor(readonly message: string) {}
}

const fetchData = Effect.gen(function* () {
  const res = yield* Effect.tryPromise({
    try: () => fetch("/api/data"),
    catch: e => new ApiError(String(e)),
  });

  if (!res.ok) {
    yield* Effect.fail(new ApiError(`HTTP ${res.status}`));
  }

  const data = yield* Effect.tryPromise({
    try: () => res.json(),
    catch: e => new ApiError(String(e)),
  });

  return data;
});
```

### Converting Component Effects

**Before:**

```typescript
createEffect(() => {
  const load = async () => {
    const data = await fetchData();
    setData(data);
  };
  load();
});
```

**After:**

```typescript
createEffect(() => {
  const fiber = Effect.runFork(
    fetchData.pipe(
      Effect.tap(data => Effect.sync(() => setData(data))),
      Effect.catchAll(error => Effect.sync(() => console.error(error)))
    )
  );

  onCleanup(() => {
    Effect.runSync(Fiber.interrupt(fiber));
  });
});
```

---

## 📝 Action Items

### Immediate (This Week)

1. [ ] Fix `services/skills-api.ts` - Remove `await` from inside Effect.gen
2. [ ] Migrate `components/Skills.tsx` - Use EffectTS skills service
3. [ ] Test all EffectTS services still work after fixes

### Short Term (Next 2 Weeks)

1. [ ] Migrate `components/App.tsx` health check to EffectTS
2. [ ] Migrate `components/ProviderManager.tsx` to use EffectTS services
3. [ ] Migrate `components/SetupRequired.tsx` to use EffectTS services

### Medium Term (Next Month)

1. [ ] Migrate `components/Channels.tsx`
2. [ ] Create ClipboardService and migrate `hooks/useMouse.ts`
3. [ ] Document EffectTS patterns for future developers

### Long Term

1. [ ] Add EffectTS linting rules to enforce consistent usage
2. [ ] Create testing utilities for EffectTS in SolidJS components
3. [ ] Consider Effect.Layers for dependency injection across all services

---

## 🔗 References

- EffectTS Documentation: https://effect.website/
- EffectTS GitHub: https://github.com/Effect-TS/effect
- SolidJS + EffectTS patterns in codebase:
  - `Chat.tsx` - Best example of Effect.runFork with Solid
  - `services/ws.ts` - Best example of Effect service architecture
  - `lib/hooks.ts` - Best example of useEffectService pattern

---

**Report Generated By:** Comprehensive TypeScript/EffectTS Audit  
**Status:** Ready for Migration Sprint Planning
