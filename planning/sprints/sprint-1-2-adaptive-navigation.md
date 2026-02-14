# Sprint 1.2: Adaptive Navigation (The Left Nav)

> **Epic:** 1 — The Context Neural Network (UI Architecture)
> **Depends On:** Sprint 1.1 (ContextState in store)
> **Objective:** The Left Nav dynamically adapts its link behavior based on whether the user is in Global or Project context.

---

## Sprint Tasks

### Task 1.2.1: Refactor `fb-left-nav.ts` to Listen to `ContextState`

**Status:** ⬜ Not Started

**Current State:**
- [fb-left-nav.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/layout/fb-left-nav.ts) (507 lines)
- Already has context-aware navigation via `_projectId` check in `_navigate()`
- Listens to `store.activeProjectId$` via signal effect
- Needs to switch to `store.contextState$` (from Sprint 1.1)

**Atomic Steps:**

1. **Replace signal subscription:** In `connectedCallback()`, change `store.activeProjectId$` → `store.contextState$`. Derive both `_projectId` and a new `_scope` state property.
2. **Add `_scope` state property:** `@state() private _scope: 'global' | 'project' = 'global';`
3. **Update `_navigate()` dispatch logic:** Use `this._scope` instead of checking `this._projectId` truthiness. This is semantically clearer and aligns with the Context Spine.
4. **Add context indicator:** Show a subtle "Global" or project-name label below the logo to visually reinforce context.

---

### Task 1.2.2: Update "Daily Focus" Link Behavior

**Status:** ⬜ Not Started

**Behavior Matrix:**

| Context | "Daily Focus" Routes To | View |
|---------|------------------------|------|
| Global | Aggregated Risk Feed | `fb-home-feed` (no project filter) |
| Project | Project-Specific Critical Path Feed | `fb-home-feed?project=<id>` (filtered) |

**Atomic Steps:**

1. In `_navigate('feed')`, when `_scope === 'global'` → emit `{ view: 'home' }`.
2. When `_scope === 'project'` → emit `{ view: 'project', projectId: this._projectId }`.
3. Update aria-label dynamically: "Daily Focus" vs "Project Focus".

> [!NOTE]
> This behavior already exists in the current implementation. The refactor makes it explicit via scope rather than implicit via null-check.

---

### Task 1.2.3: Update "Path" (Schedule) Link Behavior

**Status:** ⬜ Not Started

**Behavior Matrix:**

| Context | "Path" Routes To | View |
|---------|-----------------|------|
| Global | Master Gantt (high-level timeline of all jobs) | `/schedule` |
| Project | Detailed CPM Gantt | `/project/<id>/schedule` |

**Atomic Steps:**

1. In `_navigate('schedule')`, when `_scope === 'global'` → emit `{ view: 'schedule' }`.
2. When `_scope === 'project'` → emit `{ view: 'project-schedule', projectId: this._projectId }`.
3. Update tooltip/aria-label: "All Schedules" vs "Project Schedule".

---

### Task 1.2.4: Update "Financials" Link Behavior

**Status:** ⬜ Not Started

**Behavior Matrix:**

| Context | "Financials" Routes To | View |
|---------|----------------------|------|
| Global | Company Cash Flow Forecast | `/budget` |
| Project | Job Cost Budget vs. Actual | `/project/<id>/budget` (needs new route) |

**Atomic Steps:**

1. In `_navigate('budget')`, when `_scope === 'global'` → emit `{ view: 'budget' }`.
2. When `_scope === 'project'` → emit `{ view: 'project-budget', projectId: this._projectId }`.
3. **Router gap:** `fb-app-shell.ts` `matchRoute()` needs a new pattern: `/project/:id/budget`.
4. Add route case in `_renderContent()` for `project-budget` view.

---

## Codebase References

| File | Path | Lines | Relevance |
|------|------|-------|-----------|
| fb-left-nav.ts | `frontend/src/components/layout/fb-left-nav.ts` | 507 | Primary modification target |
| fb-app-shell.ts | `frontend/src/components/layout/fb-app-shell.ts` | 783 | Routing changes for new `/project/:id/budget` |
| store.ts | `frontend/src/store/store.ts` | 894 | New `contextState$` signal (from Sprint 1.1) |

## Verification Plan

- **Manual:** Click "All" pill → verify left nav links route to global views
- **Manual:** Click a project pill → verify left nav links route to project-scoped views
- **Manual:** Verify nav item tooltips update based on context
- **Manual:** Verify budget route `/project/<id>/budget` renders correctly
