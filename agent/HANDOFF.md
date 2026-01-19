# Handoff

**Current Phase:** Phase 7: Frontend - Lit + TypeScript
**Completed Step:** 51.3: App Shell & Layout Structure (Command Center)
**Next Step:** 51.4: View Routing & Guards

---

## ✅ Recent Achievements

### 1. App Shell & Command Center (Step 51.3)
- **Pivot to Command Center**: Implemented VS Code-style layout with Navigation Rail and Context Header.
- **Components Created**:
  - `fb-nav-rail.ts`: 64px vertical rail with icons and active state indication.
  - `fb-header.ts`: 56px context header with breadcrumbs and status.
  - `fb-app-shell.ts`: Root Grid container with Viewport Lock (`100vh`, `overflow: hidden`).
- **Responsive Behavior**: Rail transforms to bottom nav on mobile; Grid areas re-flow.
- **Verification**: Browser-verified layout, theme toggling, and nav selection.

### 2. Reactive State Engine (Step 51.2)
- **Global Store**: Signals-based state (`@preact/signals-core`) for Auth, UI, and Data.
- **Network Layer**: Typed `http.ts` service with automatic auth injection.

---

## 📋 Next Step: View Routing & Guards (Step 51.4)

**Objective**: Implement the internal "State Router" that swaps views within the App Shell's Main Stage.

### 🛑 Critical Architectural Constraints
1.  **Viewport Lock**: The Window must NEVER scroll.
2.  **Panels, Not Pages**: Views are panels injected into `slot="stage"`.
3.  **Internal Scrolling**: Views must handle their own `overflow-y: auto`.
4.  **State Router**: Route based on `store.ui.activeView$`, NOT the URL.

### Checklist
1.  **View Registry**: Create a map of `ViewId` -> `TemplateResult` or Component Class.
2.  **The Router**:
    - In `fb-app-shell.ts` (or a new `fb-router.ts`), listen to `store.ui.activeView$`.
    - Dynamically render the correct view into the "stage" slot.
3.  **View Containers**:
    - Create placeholder views for: `Dashboard`, `Projects`, `Chat`, `Schedule`, `Directory`.
    - Ensure each view has `height: 100%; overflow-y: auto;`.
4.  **Route Guards**:
    - Intercept navigation if `!store.auth.isAuthenticated$`.
    - Redirect to `Login` view (to be created).

### Context
- The `FBAppShell` is already the root. The router logic likely belongs inside it or as a direct child.
- We are simulating a "Single Page App" where the URL might not even change, or only changes hash.

---

## 📦 System State
- **Frontend Path**: `frontend/`
- **Dev Server**: `npm run dev` (Port 3001)
- **Build Status**: Passing
- **Lint Status**: Passing