# FRONTEND V2 SPEC — Feed-First Engine-Forward UI

> **Branch:** `claude/explain-codebase-mlfs25zjisii6eei-Zalu0`
> **Status:** Design — Pre-Implementation
> **Principle:** Don't ask users to manage projects. Show them the engine thinking about their projects.

---

## 1. Design Philosophy

### 1.1 Core Shift

**V1:** Project-first. User picks a project → opens chat → asks questions → gets artifacts.
**V2:** Engine-first. User opens app → sees what the engine found → takes action → drills into chat when needed.

The backend has 5 autonomous agents, a CPM physics engine with weather adjustments, procurement MRP calculations, and inspection gate logic. V1 hides all of this behind a chat prompt. V2 makes the engine's work the primary interface.

### 1.2 Interaction Hierarchy

```
1. FEED (awareness)     — What needs attention across all projects, right now?
2. CARDS (context)      — Why does this matter? What's the consequence?
3. ACTIONS (execution)  — One-tap resolution. Approve, confirm, order, snooze.
4. CHAT (exploration)   — Drill deeper. Ask questions. Get explanations.
5. ARTIFACTS (evidence) — Gantt, budget, invoice. Visual proof of the engine's work.
```

Chat moves from Layer 1 to Layer 4. Most interactions never need it.

### 1.3 Consequences, Not Statuses

Every card shows the downstream impact, not just the current state. Instead of:
- "Roof trusses: Critical" → **"Order trusses by Feb 7 or framing slips 2 weeks → completion moves to June 14"**
- "Electrical rough-in: Completed" → **"Inspection passed. Insulation + drywall are unblocked. Schedule recalculated: still on track for May 1."**

The CPM engine calculates this. The Procurement agent calculates this. We just need to include the consequence in the card payload.

---

## 2. Views & Navigation

### 2.1 View Map

| Route | View | When Shown | Layout |
|-------|------|------------|--------|
| `/` | **Home Feed** | Default authenticated state | Full-width feed, optional right panel |
| `/onboard` | **Onboarding** | First project creation (new user or "+ New") | Full-screen, no chrome |
| `/project/:id` | **Project Detail** | Tapped project card or pill filter | Feed filtered to one project + right panel |
| `/project/:id/chat` | **Project Chat** | "Tell me more" or direct nav | 2-panel: chat + artifacts |
| `/project/:id/schedule` | **Schedule View** | Tapped Gantt card or nav | Full-width Gantt renderer |
| `/settings` | **Settings** | User settings | Standard settings page |
| `/settings/team` | **Team** | Team management | Standard settings page |
| `/admin` | **Admin Shell** | Platform admin | Separate layout |
| `/portal/*` | **Portal Views** | External contacts | Separate layout (keep V1) |
| `/login` | **Login** | Unauthenticated | Full-screen Clerk |

### 2.2 Navigation Model

**Project Pills (horizontal filter bar, top of feed):**
```
[All Projects] [123 Main St] [456 Oak Ave] [789 Pine Dr] [+ New]
```
- `All Projects` = home feed, all projects interleaved by urgency
- Tap a project pill = filter feed to that project only (route: `/project/:id`)
- `+ New` = navigate to `/onboard`
- Pills render from `store.projects$`
- Active pill highlighted with primary color

**Persistent Elements:**
- Top bar: Logo, project pills, notification bell, user avatar/menu
- No left sidebar in default view
- Right panel: appears when artifact is active (same as V1 behavior)
- Mobile: project pills become horizontal scroll; bottom nav for Home/Chat/Settings

### 2.3 View Details

#### A. Home Feed (`/`)

The primary view. Renders a vertical stream of FeedCards grouped by time horizon.

**Sections:**
```
[Greeting banner — "Good morning, Marcus. 5 active projects."]

[── TODAY ──────────────────────────────────────────────]
  FeedCard (procurement_critical)
  FeedCard (weather_window)
  FeedCard (task_completed + schedule_recalc)
  FeedCard (sub_confirmation)

[── THIS WEEK ──────────────────────────────────────────]
  FeedCard (inspection_upcoming)
  FeedCard (client_report_due)
  FeedCard (task_starting)

[── HORIZON (2-4 weeks) ───────────────────────────────]
  FeedCard (procurement_warning)
  FeedCard (permit_renewal)
  FeedCard (sub_unconfirmed)
```

**Empty states:**
- No projects: Full-screen onboarding CTA (see §2.3.B)
- Projects but no feed items: "All clear. Your agents are monitoring 3 projects."

**Data source:** New endpoint `GET /api/v1/portfolio/feed` (see §5.1)

#### B. Onboarding (`/onboard`)

Full-screen, no chrome. Two entry points:

**Path 1: File-first (primary)**
```
"What are you building?"
"Drop your plans, a permit, a contract — anything about your project."
[Large drop zone]
"or just tell me about it:"
[Text input]
```

**Path 2: Text-first (fallback)**
User types a description. Engine parses it, creates project skeleton, asks clarifying questions.

**After file drop or text submit → Extraction Experience:**
Engine narrates its work in real-time (SSE stream):
```
✓ Found: 123 Main St, Austin TX — 2,500 sqft, slab foundation
✓ 47 tasks generated from WBS template
✓ Dependencies mapped (62 edges)
✓ Durations calculated (DHSM for 2,500 sqft)
◎ Applying weather for Austin, TX...
  → Rain risk Feb 15-17, adjusted exterior tasks +2 days
✓ Critical path: 127 days → Projected completion: June 18, 2026
⚠ 3 long-lead items detected
```

**Correction loop:** Input at bottom for natural language corrections.
**Activate:** "Looks good — activate project" → `POST /api/v1/projects` → redirect to `/` with new project in feed.

**Backend:** Uses existing `POST /api/v1/agent/onboard` (multipart) + `POST /api/v1/projects`. The extraction narration is a new SSE response mode from the onboard endpoint (see §5.3).

#### C. Project Detail (`/project/:id`)

Same feed layout, but filtered to one project. Additional context:

**Project header bar:**
```
← Back | 123 Main St | Active | 67% complete | June 18 completion
[Schedule] [Budget] [Chat] [Contacts]
```

Quick-nav buttons open respective views or the right panel with the artifact.

#### D. Project Chat (`/project/:id/chat`)

Reuses V1 chat components (fb-message-list, fb-input-bar, fb-action-card, fb-typing-indicator). But now the chat is pre-seeded with context from the feed card that triggered it.

When user taps "Tell me more" on a feed card, the chat opens with:
1. The card's context as a system message (visible to user as a context banner)
2. The agent's detailed explanation as the first assistant message

**Wired to real backend:** `POST /api/v1/chat` with project context. No more mock service.

#### E. Schedule View (`/project/:id/schedule`)

Full-width Gantt with timeline axis. This is the V2 Gantt — not the V1 task list.

**Requirements:**
- Horizontal timeline (date axis)
- Task bars positioned by EarlyStart → EarlyFinish
- Critical path as connected red chain
- Dependency arrows (SVG)
- Float visualization (ghost bars from EarlyFinish → LateFinish)
- Click task → right panel shows task detail card
- "What-if" mode: drag a task → shows projected impact on end date

**Data:** `GET /api/v1/projects/:id/schedule` → `GanttData`

---

## 3. Feed Card System

### 3.1 Card Types

Each card type maps to a backend agent output or engine event.

| Card Type | Agent/Source | Priority | Actions |
|-----------|-------------|----------|---------|
| `procurement_critical` | ProcurementAgent (status=critical) | P0 | [Order Now] [Snooze 1 Day] [Tell me more] |
| `procurement_warning` | ProcurementAgent (status=warning) | P1 | [Order Now] [Remind me on {date}] [Tell me more] |
| `weather_window` | SWIM model + WeatherService | P0 | [Confirm Start] [Push to next window] |
| `weather_risk` | SWIM model + WeatherService | P1 | [Acknowledge] [Adjust schedule] |
| `task_completed` | ScheduleService (status change) | P2 | [View schedule impact] |
| `schedule_recalc` | CPM engine (after recalculation) | P1 | [View diff] [Tell me more] |
| `inspection_upcoming` | Task with is_inspection + EarlyStart in range | P1 | [Confirm ready] [Request reschedule] |
| `inspection_passed` | InspectionRecord (result=Passed) | P2 | [View unlocked tasks] |
| `inspection_failed` | InspectionRecord (result=Failed) | P0 | [Schedule re-inspection] [Tell me more] |
| `sub_confirmation` | SubLiaison agent (inbound reply) | P2 | [Acknowledge] |
| `sub_unconfirmed` | SubLiaison (no reply within 24h) | P1 | [Send reminder] [Find alternate] |
| `invoice_ready` | InvoiceService (new extraction) | P1 | [Review & Approve] [Tell me more] |
| `daily_briefing` | DailyFocusAgent | P1 | [Expand] |
| `client_report_due` | Weekly cycle (Friday) | P2 | [Review draft] [Approve & Send] |
| `task_starting` | Task with EarlyStart in range | P3 | [Confirm] |
| `blocker_detected` | Task with status=Blocked | P0 | [View blocker] [Tell me more] |

### 3.2 FeedCard Data Structure

```typescript
interface FeedCard {
  id: string;
  type: FeedCardType; // Union of card types above
  priority: 0 | 1 | 2 | 3; // 0 = highest

  // Context
  projectId: string;
  projectName: string;
  taskId?: string;        // If card relates to a specific task
  taskName?: string;
  wbsCode?: string;

  // Display
  headline: string;       // "Order trusses by Feb 7"
  body: string;           // "6-week lead. If you don't order today, framing slips..."
  consequence?: string;   // "→ Completion moves from May 1 to May 15"

  // Time
  timestamp: string;      // When the event occurred / was detected
  horizon: 'today' | 'this_week' | 'horizon'; // Which section
  deadline?: string;      // Action deadline (for urgency display)

  // Actions
  actions: FeedCardAction[];

  // Source metadata
  agentSource?: string;   // Which agent generated this
  engineData?: Record<string, unknown>; // Raw data for drill-down
}

interface FeedCardAction {
  id: string;
  label: string;           // "Order Now"
  style: 'primary' | 'secondary' | 'danger' | 'ghost';

  // One of these:
  endpoint?: string;       // API call: POST /api/v1/...
  payload?: Record<string, unknown>;
  navigate?: string;       // Client-side nav: /project/:id/chat
  dismiss?: boolean;       // Just dismiss the card
}

type FeedCardType =
  | 'procurement_critical'
  | 'procurement_warning'
  | 'weather_window'
  | 'weather_risk'
  | 'task_completed'
  | 'schedule_recalc'
  | 'inspection_upcoming'
  | 'inspection_passed'
  | 'inspection_failed'
  | 'sub_confirmation'
  | 'sub_unconfirmed'
  | 'invoice_ready'
  | 'daily_briefing'
  | 'client_report_due'
  | 'task_starting'
  | 'blocker_detected';
```

### 3.3 Card Visual Design

```
┌──────────────────────────────────────────────────────┐
│ 🔴 ORDER BY TODAY              123 Main St           │ ← Priority dot + deadline + project
├──────────────────────────────────────────────────────┤
│                                                      │
│ Roof trusses (engineered)                            │ ← Headline (bold)
│                                                      │
│ 6-week lead time. If you don't order today,          │ ← Body (regular)
│ framing slips to March 28.                           │
│                                                      │
│ → Completion moves from May 1 to May 15              │ ← Consequence (accent color)
│                                                      │
│ ┌──────────┐  ┌───────────────┐  ┌──────────────┐   │
│ │ Order Now │  │ Snooze 1 Day  │  │ Tell me more │   │ ← Actions
│ └──────────┘  └───────────────┘  └──────────────┘   │
└──────────────────────────────────────────────────────┘
```

**Priority indicators:**
- P0: Red dot + red left border
- P1: Orange dot + orange left border
- P2: Blue dot + subtle border
- P3: Gray dot + no border

**Consequence line:** Rendered in accent color (red for negative impact, green for positive).

### 3.4 Card Grouping & Sorting

Within each time horizon section (TODAY / THIS WEEK / HORIZON):
1. Sort by priority (P0 first)
2. Within same priority, sort by deadline (soonest first)
3. Within same deadline, sort by project (group same-project cards)

---

## 4. Component Architecture

### 4.1 Component Inventory (New + Kept + Rebuilt)

#### New Components

| Component | Tag | Purpose |
|-----------|-----|---------|
| `fb-home-feed` | `fb-home-feed` | Main feed view. Renders greeting + time horizon sections + FeedCards |
| `fb-feed-card` | `fb-feed-card` | Individual card. Receives `FeedCard` data. Renders headline/body/consequence/actions |
| `fb-feed-section` | `fb-feed-section` | Time horizon group header ("TODAY", "THIS WEEK", "HORIZON") |
| `fb-project-pills` | `fb-project-pills` | Horizontal scrollable project filter bar |
| `fb-top-bar` | `fb-top-bar` | Persistent top navigation: logo, pills, notifications, user menu |
| `fb-onboard-flow` | `fb-onboard-flow` | Full-screen onboarding: drop zone + extraction narration + correction loop |
| `fb-extraction-stream` | `fb-extraction-stream` | Real-time extraction display (SSE-driven line items) |
| `fb-project-header` | `fb-project-header` | Project detail header bar (name, status, completion, quick nav) |
| `fb-schedule-view` | `fb-schedule-view` | Full-width timeline Gantt with date axis, dependency arrows, float bars |
| `fb-schedule-task-bar` | `fb-schedule-task-bar` | Individual task bar on Gantt timeline |
| `fb-schedule-diff` | `fb-schedule-diff` | Before/after schedule comparison overlay |
| `fb-greeting-banner` | `fb-greeting-banner` | "Good morning, Marcus" with portfolio summary metrics |
| `fb-empty-home` | `fb-empty-home` | No-projects CTA: leads to onboarding |
| `fb-user-menu` | `fb-user-menu` | Dropdown: settings, team, theme toggle, sign out |

#### Kept (No Changes)

| Component | Tag | Reason |
|-----------|-----|--------|
| `FBElement` | N/A (base) | Base class for all components |
| `FBViewElement` | N/A (base) | Base for page views |
| `fb-error-boundary` | `fb-error-boundary` | Error fallback |
| `fb-artifact-modal` | `fb-artifact-modal` | Full-screen artifact popout |
| `fb-artifact-invoice` | `fb-artifact-invoice` | Editable invoice (most interactive artifact) |
| `fb-artifact-budget` | `fb-artifact-budget` | Budget display |
| `fb-artifact-actions` | `fb-artifact-actions` | Approve/reject bar for invoices |
| `fb-action-card` | `fb-action-card` | Inline chat action card |
| `fb-typing-indicator` | `fb-typing-indicator` | Chat typing dots |
| `fb-toast-container` | `fb-toast-container` | Global toast stack |
| `fb-toast` | `fb-toast` | Individual toast |
| `fb-file-drop` | `fb-file-drop` | Global drag overlay |
| `fb-status-card` | `fb-status-card` | Reusable status card |

#### Rebuilt (Same concept, new implementation)

| Component | V1 Tag | V2 Changes |
|-----------|--------|------------|
| `fb-app-shell` | `fb-app-shell` | Remove 3-panel grid default. Top bar + content area + optional right panel. No left sidebar. |
| `fb-panel-right` | `fb-panel-right` | Keep as artifact panel. Triggered by feed card actions or chat artifacts. |
| `fb-message-list` | `fb-message-list` | Keep virtualizer. Add context banner for feed-card-originated chats. |
| `fb-input-bar` | `fb-input-bar` | Keep. Wire to real `POST /api/v1/chat` instead of mock service. |
| `fb-mobile-nav` | `fb-mobile-nav` | Simplify: [Home] [Chat] [Settings]. Remove Projects tab (pills replace it). |
| `fb-artifact-gantt` | `fb-artifact-gantt` | Complete rewrite as timeline-based Gantt (see §2.3.E). Keep as `fb-artifact-gantt` for inline chat use; `fb-schedule-view` for full-page. |
| `fb-agent-activity` | `fb-agent-activity` | Merge into feed. Agent actions become FeedCards, not a sidebar widget. |
| `fb-notification-bell` | `fb-notification-bell` | Move to top bar. |
| `fb-notification-list` | `fb-notification-list` | Keep dropdown behavior. |

#### Scrapped

| Component | Reason |
|-----------|--------|
| `fb-panel-left` | No left sidebar in V2. Project navigation via pills. |
| `fb-panel-center` | Replaced by direct view routing in app-shell. |
| `fb-resize-handle` | No left panel to resize against. Right panel keeps V1 behavior. |
| `fb-view-chat` (as default) | Chat is now per-project, not the default view. |
| `fb-view-dashboard` | Never wired. Replaced by fb-home-feed. |
| `fb-project-card` | Replaced by fb-feed-card system. |
| `fb-project-dialog` | Project creation via onboarding flow, not modal form. |
| `fb-project-form` | Fields extracted by AI, not entered by user. |
| `demo-button` | Temporary, delete. |

### 4.2 App Shell Layout (V2)

```
┌─ fb-app-shell ──────────────────────────────────────────────────────┐
│ ┌─ fb-top-bar ────────────────────────────────────────────────────┐ │
│ │ [Logo]  [All] [Proj1] [Proj2] [+New]     [🔔] [Avatar ▾]     │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                     │
│ ┌─ content-area ──────────────────────┬─ fb-panel-right (opt) ───┐ │
│ │                                     │                          │ │
│ │  <router-outlet>                    │  Artifact panel          │ │
│ │    fb-home-feed                     │  (appears on demand)     │ │
│ │    fb-onboard-flow                  │                          │ │
│ │    fb-project-chat                  │                          │ │
│ │    fb-schedule-view                 │                          │ │
│ │    fb-view-settings                 │                          │ │
│ │  </router-outlet>                   │                          │ │
│ │                                     │                          │ │
│ └─────────────────────────────────────┴──────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

**CSS Grid:**
```css
/* Default: no right panel */
grid-template-columns: 1fr;
grid-template-rows: 56px 1fr;

/* With artifact panel */
grid-template-columns: 1fr var(--fb-right-panel-width, 380px);
```

**Mobile (<768px):**
```css
grid-template-columns: 1fr;
grid-template-rows: 56px 1fr 64px; /* top-bar, content, mobile-nav */
```

### 4.3 Routing

Replace `pushState` + `popstate` with a minimal router in the app shell.

```typescript
type Route =
  | { view: 'home' }
  | { view: 'onboard' }
  | { view: 'project'; projectId: string }
  | { view: 'project-chat'; projectId: string; threadId?: string }
  | { view: 'project-schedule'; projectId: string }
  | { view: 'settings' }
  | { view: 'team' }
  | { view: 'admin' }
  | { view: 'login' }
  | { view: 'portal'; subpath: string };

function matchRoute(path: string): Route { ... }
```

The app shell listens for `popstate` and calls `matchRoute`. Each route renders the corresponding view component. No SPA router library needed — keep the lightweight approach, just centralize the logic.

### 4.4 State Management Changes

**New signals:**

```typescript
// Portfolio feed
const feedCards$ = signal<FeedCard[]>([]);
const feedLoading$ = signal<boolean>(false);
const feedFilter$ = signal<string | null>(null); // null = all projects

// Active route
const currentRoute$ = signal<Route>({ view: 'home' });

// Portfolio summary (for greeting banner)
const portfolioSummary$ = signal<PortfolioSummary | null>(null);
```

**Modified signals:**
```typescript
// activeProjectId$ — now derived from route, not sidebar selection
// Computed from currentRoute$ when view is 'project' | 'project-chat' | 'project-schedule'
```

**Removed signals:**
```typescript
// leftPanelOpen$ — no left panel
// shadowModeEnabled$ — admin-only, keep but deprioritize
// focusTasks$ — merged into feedCards$
```

**New actions:**
```typescript
actions.loadFeed(): Promise<void>          // GET /api/v1/portfolio/feed
actions.executeFeedAction(cardId: string, actionId: string): Promise<void>
actions.dismissCard(cardId: string): void
actions.setFeedFilter(projectId: string | null): void
```

---

## 5. Backend Requirements

### 5.1 New Endpoint: Portfolio Feed

```
GET /api/v1/portfolio/feed
Authorization: Bearer <token>
Query params:
  project_id (optional) — filter to one project

Response: {
  greeting: string,              // "Good morning, Marcus"
  summary: PortfolioSummary,
  cards: FeedCard[]              // Sorted by horizon, then priority, then deadline
}
```

**PortfolioSummary:**
```go
type PortfolioSummary struct {
    ActiveProjectCount   int       `json:"active_project_count"`
    TotalTasks           int       `json:"total_tasks"`
    CriticalAlerts       int       `json:"critical_alerts"`
    ProjectedCompletions []ProjectCompletionSummary `json:"projected_completions"`
}

type ProjectCompletionSummary struct {
    ProjectID   uuid.UUID `json:"project_id"`
    ProjectName string    `json:"project_name"`
    EndDate     string    `json:"end_date"`
    OnTrack     bool      `json:"on_track"`
    SlipDays    int       `json:"slip_days"`
}
```

**Implementation approach:**
This endpoint aggregates data from existing services. No new tables needed.

```go
func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
    orgID := middleware.GetClaims(r.Context()).OrgID
    projectFilter := r.URL.Query().Get("project_id")

    // 1. Load all active projects for org
    projects := h.projectService.ListActiveProjects(ctx, orgID)

    // 2. For each project, gather:
    //    a. Procurement items (warning/critical) → procurement_* cards
    //    b. Tasks starting within 7 days → task_starting cards
    //    c. Tasks recently completed → task_completed cards
    //    d. Inspections upcoming → inspection_upcoming cards
    //    e. Weather forecast for project zip → weather_* cards
    //    f. Recent agent activity (communication_logs) → agent cards

    // 3. Build FeedCard[] with consequence text
    // 4. Sort and return
}
```

**Key: Consequence calculation.** For procurement cards, run a hypothetical:
"If this item slips N days, what happens to the critical path?" This uses:
```go
// Existing: ScheduleService.RecalculateSchedule()
// New: ScheduleService.SimulateSlip(ctx, projectID, taskID, slipDays) → SlipImpact
```

### 5.2 New Endpoint: Feed Action Execution

```
POST /api/v1/portfolio/feed/action
Authorization: Bearer <token>
Body: {
  card_id: string,
  action_id: string,
  payload?: Record<string, unknown>
}

Response: {
  success: boolean,
  message: string,
  updated_cards?: FeedCard[]  // Cards that changed as a result
}
```

This is a command dispatcher. Based on `action_id`, it routes to existing service methods:
- `"order_now"` → Update procurement item status
- `"confirm_start"` → Update task status to In_Progress
- `"approve_invoice"` → InvoiceService.ApproveInvoice()
- `"send_reminder"` → NotificationService.SendSMS()
- `"snooze"` → Mark card as snoozed (new field in communication_logs or dedicated table)
- `"dismiss"` → Client-side only (remove from feed)

### 5.3 Enhanced Onboarding Endpoint (SSE)

Modify `POST /api/v1/agent/onboard` to support SSE streaming for extraction narration:

```
POST /api/v1/agent/onboard
Accept: text/event-stream
Content-Type: multipart/form-data

SSE events:
  event: extraction
  data: {"step": "address", "value": "123 Main St, Austin TX", "confidence": 0.95}

  event: scheduling
  data: {"step": "tasks_generated", "count": 47}

  event: scheduling
  data: {"step": "dependencies_mapped", "count": 62}

  event: scheduling
  data: {"step": "durations_calculated", "model": "DHSM", "sqft": 2500}

  event: weather
  data: {"step": "forecast_applied", "location": "Austin, TX", "adjustments": [{"task": "Exterior Paint", "delta_days": 2, "reason": "Rain risk Feb 15-17"}]}

  event: scheduling
  data: {"step": "critical_path", "days": 127, "end_date": "2026-06-18"}

  event: procurement
  data: {"step": "long_lead_detected", "items": [{"name": "Roof Trusses", "lead_weeks": 6, "order_by": "2026-02-14"}]}

  event: complete
  data: {"session_id": "...", "ready_to_create": true, "extracted_values": {...}}
```

This requires refactoring the onboarding handler to emit SSE events at each stage of processing instead of returning a single JSON response. The existing `InterrogatorService` logic stays the same — we just stream intermediate results.

### 5.4 New Endpoint: Schedule Simulation

```
GET /api/v1/projects/:id/schedule/simulate
Query params:
  task_id: uuid
  slip_days: int

Response: {
  original_end_date: string,
  simulated_end_date: string,
  slip_days: int,
  affected_tasks: [{
    wbs_code: string,
    name: string,
    original_start: string,
    simulated_start: string,
    delta_days: int
  }]
}
```

**Implementation:** Clone the task graph in memory, apply the slip, run CPM forward pass, diff against original. The physics engine already supports this — `RecalculateSchedule` does a full pass. We just need a read-only variant that doesn't persist.

### 5.5 New Service: FeedService

```go
// internal/service/feed_service.go

type FeedServicer interface {
    GetFeed(ctx context.Context, orgID uuid.UUID, projectFilter *uuid.UUID) (*FeedResponse, error)
    ExecuteAction(ctx context.Context, orgID uuid.UUID, cardID string, actionID string, payload map[string]interface{}) (*ActionResult, error)
}

type FeedService struct {
    projects    ProjectServicer
    schedule    ScheduleServicer
    procurement ProjectServicer // for ListProcurementItems
    weather     WeatherServicer
    directory   DirectoryServicer
    config      ConfigServicer
    geocoding   GeocodingServicer
    clock       clock.Clock
}
```

### 5.6 Existing Endpoints Used (No Changes)

| Endpoint | Used By |
|----------|---------|
| `GET /api/v1/projects/:id/schedule` | Schedule view, Gantt artifact |
| `POST /api/v1/projects/:id/schedule/recalculate` | After task updates |
| `GET /api/v1/projects/:id/procurement` | Feed card data |
| `POST /api/v1/chat` | Project chat view |
| `GET /api/v1/projects/:id/threads` | Chat thread list |
| `GET /api/v1/projects/:id/threads/:tid/messages` | Chat history |
| `POST /api/v1/projects` | Project creation (after onboarding) |
| `GET /api/v1/invoices/:id` | Invoice artifact |
| `PUT /api/v1/invoices/:id` | Invoice edit |
| `POST /api/v1/invoices/:id/approve` | Invoice approval |
| `POST /api/v1/invoices/:id/reject` | Invoice rejection |
| `GET /api/v1/auth/me` | Auth state |
| `GET /api/v1/users/me` | User profile |
| `GET /api/v1/org/members` | Team management |

---

## 6. Engine Improvements

Opportunities to strengthen the backend as we surface more of it.

### 6.1 Consequence Text Generation

**Problem:** Feed cards need human-readable consequence strings. The CPM engine calculates numeric impacts but doesn't produce natural language.

**Solution:** Add a `ConsequenceFormatter` utility:
```go
// internal/physics/consequence.go

func FormatSlipConsequence(original, simulated time.Time, taskName string) string {
    delta := simulated.Sub(original).Hours() / 24
    if delta == 0 {
        return fmt.Sprintf("No impact on completion date (absorbed by float)")
    }
    return fmt.Sprintf("→ Completion moves from %s to %s (+%d days)",
        original.Format("Jan 2"), simulated.Format("Jan 2"), int(delta))
}
```

This keeps consequence generation deterministic (no AI needed for schedule math).

### 6.2 Feed Card Materialization

**Problem:** Building the feed on every request by querying 5+ tables per project is expensive for orgs with many projects.

**Solution:** Materialized feed table, updated by agents:
```sql
CREATE TABLE feed_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    project_id UUID NOT NULL REFERENCES projects(id),
    card_type TEXT NOT NULL,
    priority INT NOT NULL DEFAULT 2,
    headline TEXT NOT NULL,
    body TEXT NOT NULL,
    consequence TEXT,
    horizon TEXT NOT NULL, -- 'today', 'this_week', 'horizon'
    deadline TIMESTAMPTZ,
    actions JSONB NOT NULL DEFAULT '[]',
    engine_data JSONB,
    agent_source TEXT,
    task_id UUID REFERENCES project_tasks(id),
    dismissed_at TIMESTAMPTZ,
    snoozed_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ -- auto-cleanup stale cards
);

CREATE INDEX idx_feed_cards_org_active ON feed_cards(org_id)
    WHERE dismissed_at IS NULL AND (snoozed_until IS NULL OR snoozed_until < NOW());
```

**Writers:** Each agent writes cards as a side effect of its processing:
- ProcurementAgent → writes `procurement_warning` / `procurement_critical` cards
- DailyFocusAgent → writes `daily_briefing` card
- SubLiaison → writes `sub_confirmation` / `sub_unconfirmed` cards
- ScheduleService (on recalc) → writes `schedule_recalc` / `task_completed` cards
- WeatherService (on forecast fetch) → writes `weather_window` / `weather_risk` cards

**Reader:** `GET /api/v1/portfolio/feed` reads from this table with filters.

**Expiry:** Cron job cleans cards older than `expires_at`. Task completion cards expire after 24h. Procurement cards update in-place.

### 6.3 Schedule Diff on Recalculation

**Problem:** When CPM recalculates, the old schedule is overwritten. No diff is available.

**Solution:** Before recalculation, snapshot the current end date and critical path:
```go
func (s *ScheduleService) RecalculateSchedule(ctx, projectID, orgID) (*CPMResult, error) {
    // 1. Snapshot current state
    oldGantt, _ := s.GetGanttData(ctx, projectID, orgID)
    oldEndDate := oldGantt.ProjectedEndDate

    // 2. Run CPM
    result, err := s.physics.Calculate(ctx, projectID)

    // 3. Persist new schedule (existing logic)
    s.persistSchedule(ctx, result)

    // 4. Compute diff
    newGantt, _ := s.GetGanttData(ctx, projectID, orgID)
    diff := computeScheduleDiff(oldGantt, newGantt)

    // 5. Write schedule_recalc feed card with diff
    if diff.EndDateDelta != 0 || len(diff.ChangedTasks) > 0 {
        s.feedWriter.WriteCard(ctx, orgID, projectID, diff.ToFeedCard())
    }

    return result, nil
}
```

### 6.4 Weather Window Detection

**Problem:** The SWIM model adjusts durations for weather, but doesn't proactively detect good windows.

**Solution:** When fetching forecasts, look for consecutive clear days that align with pending weather-sensitive tasks:
```go
func DetectWeatherWindows(forecasts []Forecast, tasks []ProjectTask) []WeatherWindow {
    // Find runs of clear days (precip < 30%, no "Rain"/"Storm")
    // Match against pending weather-sensitive tasks (pre-dry-in phases)
    // Return windows with task assignments and close dates
}
```

This feeds `weather_window` cards into the feed.

### 6.5 Real-Time Feed Updates (SSE)

**Problem:** Feed cards should update live (e.g., when a sub confirms via SMS).

**Solution:** SSE endpoint for feed subscriptions:
```
GET /api/v1/portfolio/feed/stream
Accept: text/event-stream

Events:
  event: card_added
  data: { card: FeedCard }

  event: card_updated
  data: { card_id: string, changes: Partial<FeedCard> }

  event: card_removed
  data: { card_id: string }
```

**Implementation:** Use PostgreSQL LISTEN/NOTIFY on `feed_cards` table changes. The API server subscribes and fans out to connected SSE clients.

---

## 7. Implementation Phases

### Phase 1: Foundation (Backend + Shell)
1. Create `feed_cards` migration and model
2. Implement `FeedService` with `GetFeed` (read from feed_cards table)
3. Implement `GET /api/v1/portfolio/feed` handler
4. Rebuild `fb-app-shell` with top bar + content area layout
5. Build `fb-top-bar` with project pills
6. Build `fb-home-feed` rendering FeedCards from new endpoint
7. Build `fb-feed-card` component
8. Update routing to default to home feed

### Phase 2: Card Population (Agent Integration)
9. Modify ProcurementAgent to write feed_cards on status change
10. Modify ScheduleService to write feed_cards on recalculation (with diff)
11. Add weather window detection, write feed_cards
12. Modify SubLiaison to write feed_cards on outbound/inbound
13. Modify DailyFocusAgent to write feed_card instead of (or in addition to) email
14. Implement card expiry cron job

### Phase 3: Actions & Interactivity
15. Implement `POST /api/v1/portfolio/feed/action` dispatcher
16. Wire feed card action buttons to API
17. Add consequence calculation (SimulateSlip endpoint)
18. Add consequence text to procurement and schedule cards

### Phase 4: Onboarding Redesign
19. Build `fb-onboard-flow` full-screen component
20. Implement SSE streaming mode for onboarding endpoint
21. Build `fb-extraction-stream` component
22. Wire correction loop to existing onboard chat endpoint

### Phase 5: Chat Integration
23. Wire `fb-input-bar` to real `POST /api/v1/chat` (kill mock service)
24. Implement "Tell me more" flow: feed card context → chat thread
25. Add context banner to chat for card-originated conversations

### Phase 6: Schedule View
26. Build `fb-schedule-view` with timeline Gantt
27. Build `fb-schedule-task-bar` positioned by date
28. Add dependency arrows on timeline
29. Add float visualization (ghost bars)
30. Add schedule diff overlay (`fb-schedule-diff`)

### Phase 7: Real-Time & Polish
31. Implement SSE feed stream endpoint
32. Wire store to SSE for live card updates
33. Mobile optimization pass
34. Accessibility audit (WCAG 2.1 AA)

---

## 8. Files to Delete

These V1 files are replaced by the new architecture and should be removed:

```
frontend/src/components/layout/fb-panel-left.ts        → No left sidebar
frontend/src/components/layout/fb-panel-center.ts       → Routing moves to app-shell
frontend/src/components/layout/fb-resize-handle.ts      → No left/center resize
frontend/src/components/base/demo-button.ts             → Temporary, delete
frontend/src/components/views/fb-view-dashboard.ts      → Never wired, replaced by fb-home-feed
frontend/src/components/features/project/fb-project-card.ts    → Replaced by feed cards
frontend/src/components/features/project/fb-project-dialog.ts  → Replaced by onboarding flow
frontend/src/components/features/project/fb-project-form.ts    → Replaced by AI extraction
frontend/src/services/realtime/mock-service.ts          → Replace with real API calls
```

---

## 9. Files to Keep Unchanged

These V1 files work as-is and should not be modified:

```
frontend/src/components/base/FBElement.ts
frontend/src/components/base/FBViewElement.ts
frontend/src/components/base/fb-error-boundary.ts
frontend/src/components/artifacts/fb-artifact-modal.ts
frontend/src/components/artifacts/fb-artifact-invoice.ts
frontend/src/components/artifacts/fb-artifact-budget.ts
frontend/src/components/artifacts/fb-artifact-actions.ts
frontend/src/components/chat/fb-action-card.ts
frontend/src/components/chat/fb-typing-indicator.ts
frontend/src/components/feedback/fb-toast-container.ts
frontend/src/components/feedback/fb-toast.ts
frontend/src/components/features/fb-file-drop.ts
frontend/src/components/widgets/fb-status-card.ts
frontend/src/services/api.ts
frontend/src/services/http.ts
frontend/src/services/clerk.ts
frontend/src/services/realtime/index.ts
frontend/src/services/realtime/interfaces.ts
frontend/src/services/realtime/types.ts
frontend/src/styles/variables.css
frontend/src/styles/main.css
```

---

## 10. Open Questions

1. **Snooze persistence:** Where to store snoozed card state? Dedicated column in `feed_cards` (proposed) vs. user-scoped table?
2. **Card deduplication:** If ProcurementAgent runs twice, how to prevent duplicate cards? Upsert on `(project_id, card_type, task_id)` composite key?
3. **Feed pagination:** For orgs with 50+ projects, paginate or load all? Recommendation: load all for "today" section, paginate "this_week" and "horizon".
4. **Offline/PWA:** Should the feed work offline? Would require service worker + IndexedDB cache of feed_cards.
5. **Multi-user feed:** Should the feed show cards for all org members, or scope to the logged-in user's projects? Recommendation: scope by user's assigned projects (via RBAC).
