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
| `/settings/profile` | **Profile** | User profile | Standard settings page |
| `/settings/org` | **Org Settings** | Physics config, org info (Admin/Builder) | Standard settings page |
| `/settings/team` | **Team** | Team management (Admin) | Standard settings page |
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

**User Menu (avatar dropdown, top-right):**
```
[Avatar ▾]
├── My Profile        → /settings/profile
├── Organization      → /settings/org    (Admin/Builder only)
├── Team & Invites    → /settings/team   (Admin only)
├── ──────────
├── Theme: Dark/Light (toggle inline)
└── Sign Out
```

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

**Engine Calibration (first project only):**

After extraction, before activation, show a lightweight calibration step. This sets org-level physics config (`PUT /api/v1/org/settings/physics`) and project-level context (stored in `ProjectContext` via `POST /api/v1/projects`). Only shown on the user's first project — subsequent projects inherit org settings silently.

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│  Before I activate your schedule, two quick ones:    │
│                                                      │
│  Which days does your crew work?                     │
│  [M ✓] [T ✓] [W ✓] [Th ✓] [F ✓] [Sa ○] [Su ○]    │
│                                                      │
│  How long do inspections typically take              │
│  in your area?                                       │
│  Rough: [3 days ▾]  Final: [5 days ▾]               │
│                                                      │
│  [ Use defaults and skip ]                           │
│  [ Apply → Activate project ]                        │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Design decisions:**
- **No speed multiplier slider here.** Speed calibration is too abstract at project creation time. The engine starts at 1.0x (industry standard) and the system observes actual performance over time (see §11.2 for passive recalibration).
- **Work days are concrete and knowable** — every builder knows their crew's schedule.
- **Inspection latency is local knowledge** — builders know their jurisdiction.
- **"Use defaults and skip"** always available. No gates.
- Supply chain volatility is not exposed — the Procurement agent handles this internally via weather + lead time math.

**Activate:** "Apply → Activate project" → `PUT /api/v1/org/settings/physics` (work days) + `POST /api/v1/projects` (with inspection latency in ProjectContext) → redirect to `/` with new project in feed.

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
| `setup_team` | System (post-first-project) | P2 | [Add contacts] [Skip for now] |
| `setup_contacts` | SubLiaison (phase has no contact, task starting soon) | P1 | [Assign contact] [Skip notification] |
| `calibration_drift` | Engine (see §11.2 — only on sustained major deviation) | P2 | [Adjust] [Keep current] |

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
  | 'blocker_detected'
  | 'setup_team'
  | 'setup_contacts'
  | 'calibration_drift';
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
| `fb-user-menu` | `fb-user-menu` | Dropdown: profile, org settings, team, theme toggle, sign out |
| `fb-engine-calibration` | `fb-engine-calibration` | First-project-only: work days + inspection latency before activation |
| `fb-settings-profile` | `fb-settings-profile` | User profile page: name (editable), email, role (read-only) |
| `fb-settings-org` | `fb-settings-org` | Org settings: physics config (speed slider + work days), org info. Admin/Builder only |
| `fb-settings-team` | `fb-settings-team` | Team management: members list, pending invites, invite modal. Admin only |
| `fb-contact-quick-add` | `fb-contact-quick-add` | Inline contact add form (rendered inside feed card for setup_contacts) |

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
| `fb-view-settings` | `fb-view-settings` | Rebuild as router for profile/org/team sub-views. |

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
  | { view: 'settings-profile' }
  | { view: 'settings-org' }
  | { view: 'settings-team' }
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
| `GET /api/v1/users/me` | Profile settings |
| `PUT /api/v1/users/me` | Profile edit (name) |
| `GET /api/v1/org/settings/physics` | Org settings page (speed multiplier, work days) |
| `PUT /api/v1/org/settings/physics` | Org settings edit + onboarding calibration |
| `GET /api/v1/org/members` | Team settings page |
| `POST /api/v1/admin/invites` | Team invite creation |
| `GET /api/v1/admin/invites` | Pending invites list |
| `DELETE /api/v1/admin/invites/:id` | Revoke invite |

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

### Phase 4: Onboarding + Settings + Contacts
19. Build `fb-onboard-flow` full-screen component
20. Implement SSE streaming mode for onboarding endpoint
21. Build `fb-extraction-stream` component
22. Wire correction loop to existing onboard chat endpoint
23. Build `fb-engine-calibration` (first-project work days + inspection latency step)
24. Build `fb-settings-profile`, `fb-settings-org`, `fb-settings-team` pages
25. Build `fb-user-menu` dropdown with role-gated links
26. Implement contact CRUD endpoints (`POST /api/v1/contacts`, `POST /api/v1/contacts/bulk`, `GET /api/v1/contacts`)
27. Implement assignment endpoints (`POST/GET /api/v1/projects/:id/assignments`, bulk variant)
28. Build `fb-contact-phase-grid` with inline add and autocomplete
29. Build `fb-contact-bulk-input` with parse/review/save flow
30. Build `fb-contact-inline-add` for use inside `setup_contacts` feed cards
31. Wire `setup_team` and `setup_contacts` feed card types

### Phase 5: Chat Integration
32. Wire `fb-input-bar` to real `POST /api/v1/chat` (kill mock service)
33. Implement "Tell me more" flow: feed card context → chat thread
34. Add context banner to chat for card-originated conversations

### Phase 6: Schedule View
35. Build `fb-schedule-view` with timeline Gantt
36. Build `fb-schedule-task-bar` positioned by date
37. Add dependency arrows on timeline
38. Add float visualization (ghost bars)
39. Add schedule diff overlay (`fb-schedule-diff`)

### Phase 7: Real-Time & Polish
40. Implement SSE feed stream endpoint
41. Wire store to SSE for live card updates
42. Add passive drift detection (background calibration tracking — §11.2)
43. Mobile optimization pass
44. Accessibility audit (WCAG 2.1 AA)

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

## 10. Settings Architecture

### 10.1 Design Principle: Settings in Context

Settings are not a destination — they surface at the moment they matter. The settings pages exist as a reference, but the primary discovery path is through the work itself.

**Three access patterns:**
1. **Onboarding calibration** — First project only. Work days + inspection latency.
2. **Feed cards** — Contextual nudges when the engine detects missing config (no contact assigned, team not set up).
3. **Settings pages** — Traditional pages via user menu dropdown. Always accessible.

### 10.2 Settings Pages

Accessible from the user menu (avatar dropdown in `fb-top-bar`).

#### A. My Profile (`/settings/profile`)

Available to all authenticated users.

```
┌──────────────────────────────────────────────────────┐
│  My Profile                                          │
│                                                      │
│  Display Name  ┌──────────────────────────┐          │
│                │ Marcus Johnson           │  [Save]  │
│                └──────────────────────────┘          │
│                                                      │
│  Email         marcus@buildco.com (managed by Clerk) │
│  Role          Builder                               │
│  Member Since  January 12, 2026                      │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Backend:** `GET /api/v1/users/me`, `PUT /api/v1/users/me` (name only).
**Role:** Read-only. Tooltip: "Contact your administrator to change roles."

#### B. Organization Settings (`/settings/org`)

Admin and Builder roles only. Viewer/PM see read-only.

```
┌──────────────────────────────────────────────────────┐
│  Organization Settings                               │
│                                                      │
│  ── Construction Physics ──────────────────────────  │
│                                                      │
│  Crew Speed                                          │
│  ◄──────────── ● ──────────────────────►            │
│  Aggressive     Standard      Relaxed                │
│  (0.5x)         (1.0x)        (1.5x)                │
│                                                      │
│  Currently: 1.0x — Industry standard pace.           │
│  Changing this recalculates all active schedules.    │
│                                                      │
│  Work Days                                           │
│  [M ✓] [T ✓] [W ✓] [Th ✓] [F ✓] [Sa ○] [Su ○]    │
│                                                      │
│  [Save Changes]                                      │
│                                                      │
│  ── Organization Info ─────────────────────────────  │
│                                                      │
│  Name          BuildCo Construction                  │
│  Members       4                                     │
│  Projects      5 active                              │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Backend:** `GET /api/v1/org/settings/physics`, `PUT /api/v1/org/settings/physics`.
**Confirmation modal:** When speed or work days change, show: "This will recalculate schedules for all active projects. Continue?" with scope choice (apply to existing vs. future only).

**Live preview (stretch goal):** When the slider moves, show: "This would change 123 Main St completion from June 18 → June 4" using `GET /api/v1/projects/:id/schedule/simulate`.

#### C. Team & Invites (`/settings/team`)

Admin role only. Others see 403 or redirect to profile.

```
┌──────────────────────────────────────────────────────┐
│  Team & Invites                          [+ Invite]  │
│                                                      │
│  ── Members (4) ───────────────────────────────────  │
│  Marcus Johnson    marcus@buildco.com    Admin       │
│  Sarah Chen        sarah@buildco.com     Builder     │
│  Jake Williams     jake@buildco.com      PM          │
│  Lisa Park         lisa@buildco.com      Viewer      │
│                                                      │
│  ── Pending Invites (1) ───────────────────────────  │
│  tom@subelectric.com    Builder    Expires 2/10      │
│                                          [Revoke]    │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Invite modal:**
```
Email:  [                    ]
Role:   [Builder ▾]
        [Send Invite]
```

**Backend:** `GET /api/v1/org/members`, `POST /api/v1/admin/invites`, `GET /api/v1/admin/invites`, `DELETE /api/v1/admin/invites/:id`.

### 10.3 Contact & Directory Management

Contacts (subcontractors, vendors, clients) are the link between the engine and the real world. The SubLiaison agent sends SMS/email confirmations to contacts assigned to project phases. Without contacts, the agent is silent.

**Data model recap:**
- `Contact`: Name (required), Phone, Email, Company, Role (`Subcontractor`/`Client`), ContactPreference (`SMS`/`Email`/`Both`)
- `ProjectAssignment`: Links a Contact to a Project + WBS Phase (e.g., phase "9.0" = Rough-Ins)
- Constraint: one contact per (project_id, wbs_phase_id) pair

#### A. Quick-Add Flow (Primary Input Method)

Accessible from the `setup_team` feed card, the project detail contacts button, or `/project/:id/contacts`. Designed for speed — a builder should add their 6 core subs in under 60 seconds.

**The Phase Grid:**

Shows all WBS phases for the project with empty slots for contact assignment. The builder sees trade names, not WBS codes.

```
┌──────────────────────────────────────────────────────┐
│  Contacts — 123 Main St                              │
│                                                      │
│  Assign your subs to each trade. I'll handle         │
│  confirmations, reminders, and status checks.        │
│                                                      │
│  ┌────────────┬──────────────────────────────────┐   │
│  │ Foundation │  + Add contact                    │   │
│  ├────────────┼──────────────────────────────────┤   │
│  │ Framing    │  Rodriguez Framing  📱 555-0101  │   │
│  ├────────────┼──────────────────────────────────┤   │
│  │ Roofing    │  + Add contact                    │   │
│  ├────────────┼──────────────────────────────────┤   │
│  │ Electrical │  + Add contact                    │   │
│  ├────────────┼──────────────────────────────────┤   │
│  │ Plumbing   │  + Add contact                    │   │
│  ├────────────┼──────────────────────────────────┤   │
│  │ HVAC       │  + Add contact                    │   │
│  ├────────────┼──────────────────────────────────┤   │
│  │ Insulation │  + Add contact                    │   │
│  ├────────────┼──────────────────────────────────┤   │
│  │ Drywall    │  + Add contact                    │   │
│  ├────────────┼──────────────────────────────────┤   │
│  │ Finishes   │  + Add contact                    │   │
│  └────────────┴──────────────────────────────────┘   │
│                                                      │
│  [ Done ]                                            │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Clicking "+ Add contact" on a phase expands an inline row:**

```
│ Electrical │  Name: [Jake's Electric     ]         │
│            │  Phone: [555-0199           ]         │
│            │  Contact via: (●) SMS (○) Email (○) Both │
│            │  [Save]  [Cancel]                      │
```

**Minimum viable input:** Name + Phone OR Email. That's it. Company is optional. Role is auto-set to `Subcontractor`. ContactPreference defaults to `SMS` if phone provided, `Email` if only email.

**If the contact already exists in the org directory** (matched by phone or email), show a suggestion: "Jake Williams (555-0199) already in your directory. [Assign to Electrical]"

**Reuse across projects:** Once a contact exists in the org, they appear as suggestions when adding to other projects. Type "Ja..." → autocomplete shows "Jake's Electric — 555-0199".

#### B. Bulk Paste Input (Power User Shortcut)

For builders who have their sub list in a spreadsheet or notes app. A textarea that accepts freeform text:

```
┌──────────────────────────────────────────────────────┐
│  Bulk Add Contacts                                   │
│                                                      │
│  Paste your contact list. One per line.              │
│  Format: Name, Phone, Trade (optional)               │
│                                                      │
│  ┌──────────────────────────────────────────────┐    │
│  │ Jake Williams, 555-0199, Electrical          │    │
│  │ Mike's Plumbing, 555-0201                    │    │
│  │ Rodriguez Framing, 555-0101, Framing         │    │
│  │ ABC HVAC, 555-0155, HVAC                     │    │
│  │ Tom's Roofing, 555-0177                      │    │
│  └──────────────────────────────────────────────┘    │
│                                                      │
│  [Parse & Review]                                    │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**After "Parse & Review"** → shows parsed results with auto-detected trade matching:

```
│  ✓ Jake Williams     555-0199   → Electrical          │
│  ✓ Mike's Plumbing   555-0201   → Plumbing (matched)  │
│  ✓ Rodriguez Framing 555-0101   → Framing (matched)   │
│  ✓ ABC HVAC          555-0155   → HVAC                 │
│  ? Tom's Roofing     555-0177   → [Assign trade ▾]    │
│                                                        │
│  [Save All]                                            │
```

**Trade matching logic (client-side):** Simple keyword match on contact name or explicit trade field:
- "Electric" / "Electrical" → phase 9.0
- "Plumb" → phase 9.0 (or separate plumbing phase if exists)
- "Fram" → phase 7.0
- "HVAC" / "Mechanical" → phase 9.0
- "Roof" → phase 8.0
- "Drywall" → phase 10.0
- Unmatched → show dropdown to manually assign

**Parsing is lenient:** Accepts comma-separated, tab-separated, or natural text. The parser looks for phone number patterns (10+ digits, optional dashes/parens) and treats everything before as name, everything after as trade hint.

#### C. Contextual Feed Cards for Contacts

##### `setup_team` — After first project activation

```
┌──────────────────────────────────────────────────────┐
│ 👥 ADD YOUR SUBS                     123 Main St     │
│                                                      │
│ I've scheduled 9 trade phases. Adding your subs      │
│ lets me send them start confirmations, progress       │
│ checks, and delay alerts automatically.               │
│                                                      │
│ Foundation · Framing · Roofing · Electrical ·         │
│ Plumbing · HVAC · Insulation · Drywall · Finishes    │
│                                                      │
│ [ Add contacts ]  [ Paste a list ]  [ Later ]        │
└──────────────────────────────────────────────────────┘
```

"Add contacts" opens the phase grid. "Paste a list" opens the bulk input. "Later" dismisses for 7 days, then re-shows once. After second dismiss, gone permanently.

##### `setup_contacts` — Specific phase missing contact, task approaching

```
┌──────────────────────────────────────────────────────┐
│ ⚠️ NO ELECTRICIAN ASSIGNED           456 Oak Ave     │
│                                                      │
│ Electrical rough-in starts Monday. I need a          │
│ contact to send a start confirmation.                │
│                                                      │
│ Name:  [                    ]                        │
│ Phone: [                    ]                        │
│                                                      │
│ [ Save & Assign ]  [ Skip this phase ]               │
└──────────────────────────────────────────────────────┘
```

**This card has the input form inline.** No navigation needed. The builder types name + phone right in the card, taps "Save & Assign", and the SubLiaison agent can immediately send a confirmation. Two fields, one tap.

If the org already has contacts, show them as suggestions above the form:
```
│ From your directory:                                 │
│ [Jake's Electric — 555-0199]  [Mike's Plumb — 0201] │
```

#### D. New Backend Endpoints Required

```
POST /api/v1/contacts
Authorization: Bearer <token>
Scope: project:create (Admin/Builder)
Body: {
  name: string,           // required
  phone?: string,
  email?: string,
  company?: string,
  role: "Subcontractor" | "Client",  // default: Subcontractor
  contact_preference?: "SMS" | "Email" | "Both"  // inferred if omitted
}
Response: Contact (201)

POST /api/v1/contacts/bulk
Authorization: Bearer <token>
Scope: project:create (Admin/Builder)
Body: {
  contacts: [{
    name: string,
    phone?: string,
    email?: string,
    company?: string
  }]
}
Response: { created: Contact[], duplicates: Contact[] }

GET /api/v1/contacts
Authorization: Bearer <token>
Query: search (optional, matches name/phone/email)
Response: Contact[]

POST /api/v1/projects/:id/assignments
Authorization: Bearer <token>
Scope: project:create
Body: {
  contact_id: string,
  wbs_phase_id: string    // e.g., "9.0"
}
Response: ProjectAssignment (201)

POST /api/v1/projects/:id/assignments/bulk
Authorization: Bearer <token>
Scope: project:create
Body: {
  assignments: [{
    contact_id: string,
    wbs_phase_id: string
  }]
}
Response: { created: ProjectAssignment[] }

GET /api/v1/projects/:id/assignments
Authorization: Bearer <token>
Scope: project:read
Response: [{
  phase_code: string,
  phase_name: string,      // "Electrical", "Framing", etc.
  contact: Contact | null  // null = unassigned
}]
```

**Deduplication:** On `POST /api/v1/contacts`, check `(org_id, phone)` and `(org_id, email)` for existing match. If found, return existing contact instead of creating duplicate. Response includes a `matched: boolean` field so the frontend can show "Already in your directory."

#### E. New Components

| Component | Tag | Purpose |
|-----------|-----|---------|
| `fb-contact-phase-grid` | `fb-contact-phase-grid` | Phase-by-phase contact assignment grid with inline add |
| `fb-contact-bulk-input` | `fb-contact-bulk-input` | Textarea → parsed → review → save all |
| `fb-contact-inline-add` | `fb-contact-inline-add` | Minimal name+phone form, used inside feed cards and phase grid |
| `fb-contact-autocomplete` | `fb-contact-autocomplete` | Type-ahead search against org directory (`GET /api/v1/contacts?search=`) |

---

## 11. Engine Calibration Philosophy

### 11.1 Principle: The Engine Earns Trust by Being Right, Not by Asking Questions

Speed calibration is **not a user-facing setting in normal operation.** The speed multiplier starts at 1.0x (industry standard) and the system observes. Users should not need to think about abstract multipliers — the engine should just produce schedules that match reality.

### 11.2 Passive Drift Detection (Background, Minimal Surfacing)

The engine silently tracks actual task durations vs. predicted durations. This is a background calculation, not a user-facing feature.

```go
// internal/physics/calibration.go

type DriftObservation struct {
    TaskID          uuid.UUID
    PredictedDays   float64
    ActualDays      float64
    Ratio           float64 // actual / predicted
}

// Accumulates over completed tasks. Only triggers a card when:
// 1. At least 8 tasks have been completed (statistical minimum)
// 2. The rolling average ratio deviates > 25% from 1.0 consistently
// 3. The deviation has persisted across the last 5+ completed tasks (not a one-off)
```

**Threshold: sustained, major deviation only.** A crew finishing one task 10% early does not trigger anything. A crew consistently finishing 30% faster across 8+ tasks does.

### 11.3 Calibration Drift Card (Rare)

Only shown when all three conditions in §11.2 are met. This card should be infrequent — most users may never see it.

```
┌──────────────────────────────────────────────────────┐
│ 🔧 SCHEDULE ACCURACY                                │
│                                                      │
│ Your crew has completed the last 10 tasks an         │
│ average of 28% faster than predicted. Your           │
│ schedules may be overly conservative.                │
│                                                      │
│ Adjusting crew speed to 0.8x would bring             │
│ predictions closer to actual performance.             │
│                                                      │
│ [ Adjust to 0.8x ]  [ Keep current ]                │
└──────────────────────────────────────────────────────┘
```

**Design decisions:**
- No urgency indicator (P2 — blue dot, not red).
- "Keep current" is a valid permanent answer. If dismissed, do not re-show for 90 days.
- The card explains the **observation** (crew is faster), not the setting (speed multiplier). Users think in terms of their crew's pace, not abstract multipliers.
- Clicking "Adjust" calls `PUT /api/v1/org/settings/physics` with the suggested value, shows confirmation modal with schedule impact.

### 11.4 What Is NOT Exposed

- **Supply chain volatility:** Internal to Procurement agent. Calculated from weather + lead times. Not a user knob.
- **DHSM duration formulas:** Deterministic math based on sqft and task type. Not tunable per-org.
- **SWIM weather model parameters:** Calibrated globally. Not per-org.
- **Notification provider config:** Deployment-time environment variables, not user settings.
- **Organization JSONB settings:** Reserved for future use, no UI.

---

## 12. Open Questions

1. **Snooze persistence:** Where to store snoozed card state? Dedicated column in `feed_cards` (proposed) vs. user-scoped table?
2. **Card deduplication:** If ProcurementAgent runs twice, how to prevent duplicate cards? Upsert on `(project_id, card_type, task_id)` composite key?
3. **Feed pagination:** For orgs with 50+ projects, paginate or load all? Recommendation: load all for "today" section, paginate "this_week" and "horizon".
4. **Offline/PWA:** Should the feed work offline? Would require service worker + IndexedDB cache of feed_cards.
5. **Multi-user feed:** Should the feed show cards for all org members, or scope to the logged-in user's projects? Recommendation: scope by user's assigned projects (via RBAC).
6. **Contact management CRUD:** RESOLVED — New endpoints defined in §10.3.D. `POST /api/v1/contacts` (single + bulk), `GET /api/v1/contacts` (search), `POST /api/v1/projects/:id/assignments` (single + bulk), `GET /api/v1/projects/:id/assignments`. Deduplication by phone/email within org.
7. **Per-project physics overrides:** Currently speed multiplier and work days are org-level. Should individual projects be able to override (e.g., a project with a different crew)? Recommendation: defer until needed — org-level covers most single-crew builders.

---

## 13. Contact Model Expansion — Access Tiers & CRM Fields

### 13.1 The Problem

The current Contact model is a thin address book entry: Name, Phone, Email, Company, Role (`Client`/`Subcontractor`), ContactPreference. This creates three gaps:

1. **No trade knowledge.** The agent knows a contact is a "Subcontractor" but not that they're an electrician. When a new project needs an electrician, the agent can't suggest one from the existing directory. Trade affinity is only captured implicitly at the ProjectAssignment level (which WBS phase they're linked to).

2. **No access differentiation.** Every contact is treated the same, but in reality:
   - Most field subs just reply to SMS confirmations ("YES" / "NO")
   - Clients and key subs need to view schedules, documents, invoices via the portal
   - The portal auth system exists but there's no flag distinguishing who should receive magic-link access vs. who only gets texts

3. **No history.** No record of how reliable a contact is, when they were last used, or how responsive they are. The builder carries this knowledge in their head. The agent should learn it over time.

### 13.2 Design Principles

- **Fast input stays fast.** The Quick-Add flow (§10.3.A) still requires only Name + Phone. Expanded fields are optional and can be filled in later — by the user, by the agent (inferring trades from assignments), or never.
- **Agent-writable fields.** Some fields are computed by the system, not entered by the user. The agent observes behavior and enriches the contact record over time.
- **Portal access is opt-in.** Defaults to off. Enabling it requires an email address. The builder explicitly grants portal access — the system never auto-enables it.

### 13.3 Contact Access Tiers

Two tiers, controlled by a single `portal_enabled` boolean:

| | **Passive (default)** | **Portal-Enabled** |
|---|---|---|
| **Flag** | `portal_enabled = false` | `portal_enabled = true` |
| **Requirement** | Phone OR Email | Email (required) + Phone (optional) |
| **Receives** | SMS/email from SubLiaison agent | Same + magic-link portal invites |
| **Can access** | Nothing — responds to messages only | Portal dashboard: tasks, schedule, documents, invoices, messaging |
| **Agent behavior** | "Reply YES to confirm Monday start" | "View your upcoming tasks and upload documents: [link]" |
| **Typical persona** | Day-to-day field sub, material vendor | Homeowner client, GC partner, key sub with document needs |

**Validation rules:**
- `portal_enabled = true` requires non-null `email`
- Setting `portal_enabled = true` does NOT auto-send a magic link. The builder (or SubLiaison agent on first task notification) triggers the initial invite.
- Revoking portal access (`portal_enabled = false`) invalidates any active portal tokens for that contact.

**SubLiaison agent adaptation:**
```go
// When notifying a contact about an upcoming task:
if contact.PortalEnabled && contact.Email != nil {
    // Send magic-link email with task context + portal dashboard link
    // Email includes: task name, scheduled date, project address, portal link
    sendPortalNotification(ctx, contact, task, project)
} else if contact.Phone != nil {
    // Send SMS: "Hi {name}, {task} at {address} is scheduled for {date}. Reply YES to confirm."
    sendSMSConfirmation(ctx, contact, task)
} else if contact.Email != nil {
    // Send plain email (no portal link): "Hi {name}, {task} is scheduled for {date}. Reply to confirm."
    sendEmailConfirmation(ctx, contact, task)
}
```

### 13.4 Expanded Contact Fields

#### A. User-Entered Fields (New)

| Field | Type | Required | Purpose |
|-------|------|----------|---------|
| `trades` | `text[]` | No | Trade specialties: `["Electrical", "Fire Alarm", "Low Voltage"]`. Key for agent attribution — enables cross-project contact suggestions. |
| `license_number` | `varchar(100)` | No | Contractor license. Useful for compliance, inspection coordination. |
| `address_city` | `varchar(100)` | No | Business city. Enables service area matching. |
| `address_state` | `varchar(50)` | No | Business state. |
| `address_zip` | `varchar(20)` | No | Business zip. Enables proximity matching to project sites. |
| `website` | `varchar(500)` | No | Business website. |
| `notes` | `text` | No | Freeform builder notes: "prefers morning starts", "always needs 2-day lead time". Agent-readable context. |
| `portal_enabled` | `boolean` | No | Default `false`. See §13.3. |
| `source` | `varchar(50)` | No | How the contact was created: `manual`, `bulk_import`, `agent_inferred`. Default `manual`. |

#### B. Agent-Computed Fields (System-Written, Read-Only to User)

These fields are updated by background processes, never directly edited by users. They power agent intelligence and surface in the contact directory UI as read-only badges/stats.

| Field | Type | Updated By | Purpose |
|-------|------|-----------|---------|
| `last_contacted_at` | `timestamptz` | SubLiaison agent | When the system last sent this contact a message. Prevents over-contacting. |
| `total_projects` | Computed (not stored) | JOIN on `project_assignments` | Number of projects this contact has been assigned to. |
| `avg_response_time_hours` | `numeric(6,1)` | InboundProcessor | Average hours between outbound message and inbound response. Computed from `communication_logs`. |
| `on_time_rate` | `numeric(4,2)` | DailyFocusAgent | Percentage of assigned tasks where the contact confirmed and work started within the scheduled window. |
| `updated_at` | `timestamptz` | Trigger | Last modification timestamp. |

**Agent attribution logic using these fields:**

When the agent needs to suggest a contact for a trade phase on a new project:
```go
// internal/service/directory_service.go

func (s *DirectoryService) SuggestContactsForPhase(ctx context.Context, orgID uuid.UUID, tradeHint string, projectZip string) ([]ContactSuggestion, error) {
    // 1. Find contacts with matching trade in their trades[] array
    // 2. If no explicit trade match, fall back to contacts who've been assigned
    //    to similar WBS phases on past projects (inferred trade)
    // 3. Rank by: on_time_rate (desc), avg_response_time (asc), total_projects (desc)
    // 4. Optionally filter by proximity (contact zip vs. project zip)
    // 5. Return top 3 suggestions with confidence reason
}
```

#### C. Trade Inference (Agent-Driven Enrichment)

Most builders won't manually tag trades on their contacts — they'll just assign "Rodriguez" to the Framing phase. The system should learn from this:

```go
// After a ProjectAssignment is created:
func (s *DirectoryService) InferTradesFromAssignment(ctx context.Context, contactID uuid.UUID, wbsPhaseID string) {
    tradeName := MapWBSPhaseToTrade(wbsPhaseID) // "7.0" → "Framing"

    // Add to contact's trades[] if not already present
    // UPDATE contacts SET trades = array_append(trades, $1)
    // WHERE id = $2 AND NOT ($1 = ANY(trades))

    // Mark source as agent_inferred if trades were previously empty
}
```

Phase-to-trade mapping:
| WBS Phase | Trade Name |
|-----------|-----------|
| 5.2 | Sitework |
| 6.0 | Foundation |
| 7.0 | Framing |
| 8.0 | Roofing |
| 9.0 | Rough-Ins (Electrical/Plumbing/HVAC) |
| 10.0 | Insulation & Drywall |
| 11.0 | Interior Finishes |
| 12.0 | Exterior Finishes |
| 13.0 | Final / Punch |

For phase 9.0 (Rough-Ins), the system should assign the specific sub-trade based on the task name within the phase, not just "Rough-Ins."

### 13.5 Database Migration

```sql
-- Migration: expand_contacts_for_crm

-- New user-editable fields
ALTER TABLE contacts ADD COLUMN trades TEXT[] DEFAULT '{}';
ALTER TABLE contacts ADD COLUMN license_number VARCHAR(100);
ALTER TABLE contacts ADD COLUMN address_city VARCHAR(100);
ALTER TABLE contacts ADD COLUMN address_state VARCHAR(50);
ALTER TABLE contacts ADD COLUMN address_zip VARCHAR(20);
ALTER TABLE contacts ADD COLUMN website VARCHAR(500);
ALTER TABLE contacts ADD COLUMN notes TEXT;
ALTER TABLE contacts ADD COLUMN portal_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE contacts ADD COLUMN source VARCHAR(50) DEFAULT 'manual';

-- Agent-computed fields
ALTER TABLE contacts ADD COLUMN last_contacted_at TIMESTAMPTZ;
ALTER TABLE contacts ADD COLUMN avg_response_time_hours NUMERIC(6,1);
ALTER TABLE contacts ADD COLUMN on_time_rate NUMERIC(4,2);
ALTER TABLE contacts ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Index for trade-based lookups
CREATE INDEX idx_contacts_trades ON contacts USING GIN(trades);

-- Index for portal-enabled contacts (agent query: "who needs portal notifications?")
CREATE INDEX idx_contacts_portal ON contacts(org_id) WHERE portal_enabled = TRUE;

-- Index for geographic proximity matching
CREATE INDEX idx_contacts_zip ON contacts(org_id, address_zip) WHERE address_zip IS NOT NULL;

-- Trigger: auto-update updated_at
CREATE OR REPLACE FUNCTION update_contacts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER contacts_updated_at
    BEFORE UPDATE ON contacts
    FOR EACH ROW
    EXECUTE FUNCTION update_contacts_updated_at();
```

### 13.6 Updated Contact CRUD Endpoints

The endpoints defined in §10.3.D are expanded to support the new fields:

**`POST /api/v1/contacts` — Create contact (updated body)**
```json
{
  "name": "Jake Williams",
  "phone": "555-0199",
  "email": "jake@jakeselectric.com",
  "company": "Jake's Electric LLC",
  "role": "Subcontractor",
  "contact_preference": "SMS",
  "trades": ["Electrical", "Fire Alarm"],
  "license_number": "EC-2024-1234",
  "address_city": "Austin",
  "address_state": "TX",
  "address_zip": "78701",
  "website": "https://jakeselectric.com",
  "notes": "Prefers morning starts. 2-day lead time needed.",
  "portal_enabled": false
}
```

All new fields are optional. Minimum input remains: `name` + (`phone` OR `email`).

**`GET /api/v1/contacts` — List/search contacts (updated query params)**
```
GET /api/v1/contacts?search=jake              // name/phone/email search (existing)
GET /api/v1/contacts?trade=Electrical          // filter by trade
GET /api/v1/contacts?portal_enabled=true       // filter portal contacts
GET /api/v1/contacts?include_stats=true        // include computed fields (total_projects, etc.)
```

**`PUT /api/v1/contacts/:id` — Update contact**
```
PUT /api/v1/contacts/:id
Authorization: Bearer <token>
Body: { partial Contact fields }
Response: Contact (200)
```

Supports partial update. Specifically for toggling `portal_enabled`:
- Setting `portal_enabled: true` validates that `email` is non-null (400 if missing)
- Setting `portal_enabled: false` invalidates active portal tokens: `UPDATE portal_tokens SET used = true WHERE contact_id = $1 AND used = false`

**`GET /api/v1/contacts/:id/history` — Contact interaction history (new)**
```
GET /api/v1/contacts/:id/history
Authorization: Bearer <token>
Response: {
  projects: [{
    project_id: string,
    project_name: string,
    phases_assigned: string[],
    status: "active" | "completed"
  }],
  recent_messages: [{
    direction: "inbound" | "outbound",
    channel: "sms" | "email",
    summary: string,
    timestamp: string
  }],
  stats: {
    total_projects: number,
    avg_response_time_hours: number | null,
    on_time_rate: number | null
  }
}
```

### 13.7 UI Changes for Contact Tiers & Expanded Fields

#### A. Phase Grid Update (§10.3.A)

The inline add form gets one new visible element — a portal toggle — and the trade is auto-inferred from the phase:

```
│ Electrical │  Name:  [Jake Williams          ]         │
│            │  Phone: [555-0199               ]         │
│            │  Email: [jake@jakeselectric.com ] (opt)   │
│            │  Contact via: (●) SMS (○) Email (○) Both  │
│            │  ☐ Grant portal access                     │
│            │  [Save]  [Cancel]                          │
```

- Trade is auto-set to the phase name (e.g., "Electrical") — no user input needed
- "Grant portal access" checkbox only appears when email is entered
- Checking it sets `portal_enabled = true`
- Helper text below checkbox: "Portal contacts can view schedules, upload documents, and message you through FutureBuild."

#### B. Contact Detail Card (New Component: `fb-contact-detail`)

When tapping a contact in the phase grid or directory, a slide-over panel shows full details:

```
┌──────────────────────────────────────────────────────┐
│  Jake Williams                              [Edit]   │
│  Jake's Electric LLC                                 │
│  ────────────────────────────────────────────────── │
│                                                      │
│  📱 555-0199          ✉️ jake@jakeselectric.com      │
│  📍 Austin, TX 78701   🌐 jakeselectric.com          │
│  License: EC-2024-1234                               │
│                                                      │
│  Trades: [Electrical] [Fire Alarm]                   │
│  Contact via: SMS                                    │
│  Portal: ● Enabled                                   │
│                                                      │
│  ── Notes ──────────────────────────────────────── │
│  Prefers morning starts. 2-day lead time needed.     │
│                                                      │
│  ── Performance ────────────────────────────────── │
│  Projects: 3          On-time: 92%                   │
│  Avg response: 1.2 hrs    Last contacted: 2 days ago │
│                                                      │
│  ── Project History ────────────────────────────── │
│  123 Main St        Electrical    ● Active           │
│  456 Oak Ave        Electrical    ✓ Completed        │
│  789 Pine Dr        Fire Alarm    ✓ Completed        │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Performance section** only appears after 2+ project assignments. Stats are read-only, computed by agents. This is where the CRM value lives — a builder can see at a glance that Jake responds fast and is on-time 92% of the time.

#### C. Contact Directory View (`/settings/contacts` or `/contacts`)

A searchable, filterable directory of all org contacts. Not project-scoped — this is the org-wide address book.

```
┌──────────────────────────────────────────────────────┐
│  Directory                    [+ Add]  [Bulk Import] │
│                                                      │
│  Search: [                    ]  Trade: [All ▾]      │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │ Jake Williams    Jake's Electric   Electrical │   │
│  │ 📱 555-0199   ● Portal   3 projects   92% OT │   │
│  ├──────────────────────────────────────────────┤   │
│  │ Rodriguez Framing                    Framing  │   │
│  │ 📱 555-0101   ○ SMS only  2 projects  87% OT │   │
│  ├──────────────────────────────────────────────┤   │
│  │ Mike's Plumbing                     Plumbing  │   │
│  │ 📱 555-0201   ○ SMS only  1 project   — OT   │   │
│  ├──────────────────────────────────────────────┤   │
│  │ Sarah Thompson (Client)                       │   │
│  │ ✉️ sarah@gmail.com  ● Portal  1 project       │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
└──────────────────────────────────────────────────────┘
```

Badge system:
- `● Portal` (green) — portal-enabled contact
- `○ SMS only` (gray) — passive contact, SMS/email only
- `92% OT` — on-time rate (only shown after 2+ projects)
- Trade tags as pills

#### D. Bulk Paste Update (§10.3.B)

The bulk paste parser is extended to accept optional fields. The format stays lenient:

```
Jake Williams, 555-0199, jake@jakeselectric.com, Electrical, EC-2024-1234
Mike's Plumbing, 555-0201
Rodriguez Framing, 555-0101, Framing
Sarah Thompson, sarah@gmail.com, Client, portal
```

Parser rules for new fields:
- Email detected by `@` character
- Trade matched against known trade names (same keyword list as §10.3.B)
- License detected by pattern: 2-3 letters + dash + digits (e.g., `EC-2024-1234`)
- `"portal"` keyword anywhere in the line sets `portal_enabled = true`
- `"client"` keyword sets `role = Client` instead of default `Subcontractor`

### 13.8 New Components

| Component | Tag | Purpose |
|-----------|-----|---------|
| `fb-contact-detail` | `fb-contact-detail` | Slide-over panel with full contact info, stats, and history |
| `fb-contact-directory` | `fb-contact-directory` | Org-wide searchable/filterable contact list |
| `fb-contact-portal-toggle` | `fb-contact-portal-toggle` | Inline toggle for portal access with validation + confirmation |
| `fb-contact-stats` | `fb-contact-stats` | Read-only performance badges (on-time rate, response time, project count) |

### 13.9 Implementation Notes

**Phase 4 additions** (append to existing Phase 4 in §7):
- Step 26.1: Create `expand_contacts_for_crm` migration
- Step 26.2: Update Contact model in `internal/models/iam.go` with new fields
- Step 26.3: Update contact CRUD handlers for new fields + portal toggle validation
- Step 26.4: Implement `SuggestContactsForPhase` in DirectoryService
- Step 26.5: Implement trade inference on assignment creation
- Step 28.1: Add portal toggle to `fb-contact-inline-add`
- Step 30.1: Build `fb-contact-detail` slide-over
- Step 30.2: Build `fb-contact-directory` page
- Step 30.3: Add `/contacts` route to app shell

**Phase 7 additions** (agent enrichment, after real-time):
- Step 42.1: Implement `avg_response_time_hours` computation in InboundProcessor
- Step 42.2: Implement `on_time_rate` computation in DailyFocusAgent
- Step 42.3: Implement `last_contacted_at` update in SubLiaison
- Step 42.4: Implement trade inference trigger on ProjectAssignment creation
