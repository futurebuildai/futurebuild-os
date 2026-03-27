# Sprint 1.1: The Global Store & Router State

> **Epic:** 1 — The Context Neural Network (UI Architecture)
> **Objective:** Implement the `ContextState` in the global store that drives the entire application scope (Global vs. Project).

---

## Sprint Tasks

### Task 1.1.1: Refactor `store.ts` — Implement `ContextState`

**Goal:** Add a strict `ContextState` interface (`{ scope: 'global' | 'project', projectId: string | null }`) to the store and make it the single source of truth for UI scope.

**Status:** ✅ Complete

#### Atomic Steps

**Step 1: Define `ContextState` Interface**
- File: `frontend/src/store/types.ts`
- Add new interface:
  ```ts
  export type ContextScope = 'global' | 'project';
  export interface ContextState {
    scope: ContextScope;
    projectId: string | null;
  }
  ```
- Add `context: ContextState` field to `AppState` interface.

**Step 2: Add `ContextState` Signals**
- File: `frontend/src/store/store.ts`
- Add internal writable signals:
  ```ts
  const _contextScope$ = signal<ContextScope>('global');
  const _contextProjectId$ = signal<string | null>(null);
  ```
- Add computed `_contextState$`:
  ```ts
  const _contextState$ = computed<ContextState>(() => ({
    scope: _contextScope$.value,
    projectId: _contextProjectId$.value,
  }));
  ```

**Step 3: Add Context Actions**
- File: `frontend/src/store/types.ts` → `StoreActions`
- Add to `StoreActions` interface:
  ```ts
  setContext(scope: ContextScope, projectId: string | null): void;
  clearContext(): void; // Resets to global scope
  ```
- File: `frontend/src/store/store.ts` → `actions` object
- Implement:
  ```ts
  setContext(scope, projectId) {
    _contextScope$.value = scope;
    _contextProjectId$.value = projectId;
    // Also sync the legacy activeProjectId for backward compat
    _activeProjectId$.value = projectId;
  },
  clearContext() {
    _contextScope$.value = 'global';
    _contextProjectId$.value = null;
    _activeProjectId$.value = null;
  },
  ```

**Step 4: Expose Context in Store Singleton**
- File: `frontend/src/store/store.ts` → `store` export
- Add readonly signal exposure:
  ```ts
  contextScope$: _contextScope$ as ReadonlySignal<ContextScope>,
  contextProjectId$: _contextProjectId$ as ReadonlySignal<string | null>,
  contextState$: _contextState$,
  ```

**Step 5: Wire `selectProject` to Context**
- Modify existing `selectProject` action to also set context:
  ```ts
  selectProject(id: string | null): void {
    // ...existing logic...
    _contextScope$.value = id ? 'project' : 'global';
    _contextProjectId$.value = id;
  }
  ```
- Modify `setActiveProject` similarly.

**Step 6: Update `resetSession` Action**
- Add context reset to `resetSession`:
  ```ts
  _contextScope$.value = 'global';
  _contextProjectId$.value = null;
  ```

---

### Task 1.1.2: Implement `fb-project-selector` (The "Pills")

**Goal:** Create a new component that acts as the primary state dispatcher. Clicking "All" clears `projectId`; clicking a project sets it.

**Status:** ⬜ Not Started (Sprint 1.2 scope — no new components this sprint)

**Key Requirements:**
- New file: `frontend/src/components/layout/fb-project-selector.ts`
- Accepts `ProjectPill[]` as prop (from `fb-app-shell`)
- Renders horizontal pill buttons: "All" + one per project
- Dispatches `store.actions.setContext()` on click
- Emits `fb-filter-change` event for router integration
- Active pill visually highlighted (accent color)

---

### Task 1.1.3: Bind URL Query Params to Store

**Goal:** Reloading `/schedule?project=123` must hydrate the store correctly.

**Status:** ✅ Complete

**Key Requirements:**
- Modify `fb-app-shell.ts` → `_syncRoute()` to parse `?project=` query param
- On route sync: if `?project=<id>` exists, call `store.actions.setContext('project', id)`
- On context change: update URL query params via `history.replaceState`
- Ensure back/forward navigation syncs correctly

---

## Codebase Analysis (Pre-Sprint)

### Current State of `store.ts`
- **Architecture:** Signals-based (`@preact/signals-core`), singleton pattern
- **Project state:** `_activeProjectId$` signal exists, used by `selectProject` and `setActiveProject`
- **No `ContextState` concept exists** — scope is implicitly derived from `_activeProjectId$` being null or not
- **894 lines**, well-organized with clear sections

### Current State of `fb-left-nav.ts`
- **Already context-aware:** `_navigate()` checks `this._projectId` to dispatch either global or project-scoped views
- **Listens to:** `store.activeProjectId$` via signal effect
- **Ready for refactor:** Switch from `activeProjectId$` to `contextState$`

### Current State of `fb-app-shell.ts`
- **Routing:** `matchRoute()` function parses URL path, `_syncRoute()` syncs to store
- **Project pills:** Managed via `_projects: ProjectPill[]`, passed to `fb-top-bar`
- **Filter handling:** `_handleFilterChange` navigates to `/project/:id` or `/`
- **Currently syncs** `store.activeProjectId$` from route in `_syncRoute()`

### Dependencies
- Sprint 1.1 is foundational — **no blockers**
- Sprint 1.2 (Adaptive Navigation) depends on Task 1.1.1 completion
- EPIC 2-6 are independent of EPIC 1 at the task level
