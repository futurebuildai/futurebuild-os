# FutureBuild Phase 14: Physics Calibration (Settings & Gantt)

**PRD Reference:** [PHASE_14_PRD.md](../specs/PHASE_14_PRD.md)
**Objective:** Allow users to tune the construction physics engine and visualize the results in the Gantt chart with critical path highlighting and dependency arrows.

---

## CORE GUARDRAILS & PRODUCT ALIGNMENT
**Alignment Check:** Every step below must be executed with strict adherence to the **FutureBuild Product Vision**:
1.  **"Calibration as Trust":** Users must see how their settings affect the schedule. Opaque algorithms erode trust.
2.  **"Zero Latency Perception":** Slider changes must reflect immediately in the UI badge. Save must feel instant.
3.  **"Data Sovereignty":** Physics config is per-org, not per-user. One org = one physics engine configuration.
4.  **"Master PRD Compliance":** All flows must respect the multi-tenant architecture and permission matrix defined in Phase 12.

**Verification Standard:**
- **L7 Self-Reflection:** Before marking *any* task complete, you must ask: *"Does this code fundamentally improve the user's trust in the system, or is it just a feature?"*
- **No Regression:** Existing settings (Profile, Account Info) must remain fully functional.

---

## TASKS

### [x] Step 86: Builder Profile UI
**Spec:** [STEP_86_BUILDER_PROFILE_UI.md](../specs/STEP_86_BUILDER_PROFILE_UI.md)

- [x] **Frontend**: Added "Construction Physics" card to `fb-view-settings.ts`.
- [x] **Frontend**: Implemented speed slider (0.5-1.5, step 0.1) with contextual pace badges.
- [x] **Frontend**: Implemented work days toggle buttons (M-S) with active/inactive states.
- [x] **Frontend**: Console.log on save (persistence deferred to Step 87).
- [x] **Frontend**: Minimum 1 work day enforced (cannot deselect all).
- [x] **Verification**:
    - [x] **Build**: TypeScript compiles with `noImplicitAny`, Vite build passes.
    - [x] **UX**: Slider updates badge in real-time. Day toggles are tactile.
- [x] **L7 Audit**: 2 HIGH, 3 MEDIUM remediated and re-verified.
    - H1: Float precision fixed with Math.round(val * 10) / 10 in _handleSpeedChange
    - H2: CSS `input` selector scoped to `input:not([type="range"])` to prevent slider style bleed
    - M1: Dedicated `_physicsSuccess` state prevents cross-card success message leaking
    - M2: `aria-pressed` added to day toggle buttons for screen reader support
    - M3: `aria-label` on slider + `role="group"` on workdays container for accessibility

### [x] Step 87: Config Persistence
**Spec:** [STEP_87_CONFIG_PERSISTENCE.md](../specs/STEP_87_CONFIG_PERSISTENCE.md)

- [x] **Backend**: Created migration `000065_create_business_config.up.sql` + `.down.sql`.
- [x] **Backend**: Defined `BusinessConfig` model in `internal/models/business_config.go`.
- [x] **Backend**: Implemented `ConfigService` with `GetConfig()` (lazy defaults) and `UpdateConfig()` (atomic UPSERT).
- [x] **Backend**: Created `GET /api/v1/org/settings/physics` and `PUT /api/v1/org/settings/physics` in `config_handler.go`.
- [x] **Backend**: Added `ConfigServicer` interface to `internal/service/interfaces.go`.
- [x] **Backend**: Wired handler in `server.go` with RBAC: GET requires `ScopeProjectRead`, PUT requires `ScopeSettingsWrite`.
- [x] **Backend**: Added `ScopeSettingsWrite` to Builder role in RBAC.
- [x] **Frontend**: Added `api.settings.getPhysics()` and `api.settings.updatePhysics()` + types.
- [x] **Frontend**: Wired `fb-view-settings.ts` to fetch on mount and save via API.
- [x] **Verification**:
    - [x] **Build**: Both Go and TypeScript compile cleanly.
    - [x] **Multi-tenancy**: SQL queries scoped to `org_id` from JWT claims.
- [x] **L7 Audit**: 2 CRITICAL, 5 HIGH, 3 MEDIUM remediated and re-verified.
    - C-1: GET /physics gated with ScopeProjectRead (defense-in-depth)
    - C-2: Backend range aligned to 0.5-1.5 (matching frontend slider + DB CHECK)
    - H-1: NaN/Infinity guard added before range validation
    - H-4: Redundant index removed (UNIQUE constraint already creates index)
    - H-5: Duplicate log removed from handler (service already logs)
    - M-3: Dedicated `_physicsError` state prevents cross-card error leaking
    - M-5: Auth errors differentiated from network errors in _loadPhysics
    - L-6: `errors.Is(err, pgx.ErrNoRows)` for robust wrapped-error comparison

### [x] Step 88: Critical Path Visuals
**Spec:** [STEP_88_CRITICAL_PATH_VISUALS.md](../specs/STEP_88_CRITICAL_PATH_VISUALS.md)

- [x] **Frontend**: Refactored `fb-artifact-gantt.ts` — extracted `_renderTaskRow()` and `_getBarWidth()` methods.
- [x] **Frontend**: Applied `.task-row.critical` class with `rgba(198, 40, 40, 0.05)` background.
- [x] **Frontend**: Applied `.task-bar.critical` class with `#c62828` red color.
- [x] **Frontend**: Added fire emoji (&#128293;) icon for critical tasks with `title="Critical Path"`.
- [x] **Frontend**: Added hover tooltip "Critical Path" using CSS-only technique.
- [x] **Verification**:
    - [x] **Build**: TypeScript compiles cleanly.
    - [x] **Visual**: Red bars + flame icon create strong visual differentiation.
- [x] **L7 Audit**: No CRITICAL/HIGH findings. 2 MEDIUM (accepted: tooltip reflow minimal, emoji spec-compliant).

### [x] Step 89: Dependency Arrows
**Spec:** [STEP_89_DEPENDENCY_ARROWS.md](../specs/STEP_89_DEPENDENCY_ARROWS.md)

- [x] **Frontend**: Added SVG overlay layer (`.dependency-layer`) with `position: absolute` + `pointer-events: none`.
- [x] **Frontend**: Implemented `_renderDependencyLines()` with cubic bezier connectors.
- [x] **Frontend**: Defined arrowhead markers (`#arrowhead`, `#arrowhead-highlight`) in SVG `<defs>`.
- [x] **Frontend**: Implemented hover highlighting — `_hoveredTask` state triggers blue stroke + thicker width.
- [x] **Frontend**: Added `_resolveDependencies()` with explicit deps or critical_path fallback inference.
- [x] **Frontend**: Extended `GanttData` type (Go + TS) with `dependencies: GanttDependency[]`.
- [x] **Frontend**: Updated contract test with Dependencies sample data.
- [x] **Verification**:
    - [x] **Build**: Go + TypeScript compile cleanly.
    - [x] **Visual**: Bezier curves connect predecessor to successor via SVG overlay.
    - [x] **Interactive**: Hovering a task highlights its connected arrows (blue, thicker).
- [x] **L7 Audit**: No CRITICAL/HIGH findings. 2 MEDIUM remediated.
    - M1: Self-referencing dependency guard (`dep.from === dep.to` skip)
    - M2: `aria-hidden="true"` on SVG layer for screen reader compliance

---

## DEEP DIVE AUDIT: Production Readiness (Post-Phase 14)

### [x] Remediation 1: Schedule API Endpoint (CRITICAL)
- [x] **Backend**: Created `schedule_handler.go` with `GetSchedule()` and `RecalculateSchedule()` handlers.
- [x] **Backend**: Added `GetGanttData()` to `ScheduleService` — fetches tasks, dependencies (joined via WBS codes), and builds full `types.GanttData`.
- [x] **Backend**: Added `GetGanttData` to `ScheduleServicer` interface and `MockScheduleService` mock.
- [x] **Backend**: Registered routes in `server.go`:
    - `GET /projects/{id}/schedule` (ScopeProjectRead)
    - `POST /projects/{id}/schedule/recalculate` (ScopeTaskWrite)
- [x] **Backend**: Nil-slice guards for Tasks, CriticalPath, and Dependencies (JSON `[]` not `null`).

### [x] Remediation 2: Contract Test Sample (LOW)
- [x] **Backend**: Added 2nd task ("Framing", WBS 1.2) to `GanttData` sample in `contract_test.go` for consistency with `critical_path` and `dependencies` arrays.

### [x] Remediation 3: Settings Physics Error Visibility (MEDIUM)
- [x] **Frontend**: Added `_physicsUsingDefaults` state flag in `fb-view-settings.ts`.
- [x] **Frontend**: Set flag on network error in `_loadPhysics()`, cleared on successful save.
- [x] **Frontend**: Added `.info-message` CSS class and "Using default settings" banner in physics card.

### Verification
- [x] **Build**: Go + TypeScript compile cleanly after all remediations.
- [x] **Re-audit**: All CRITICAL/HIGH issues resolved. Zero remaining blockers for Phase 15.
