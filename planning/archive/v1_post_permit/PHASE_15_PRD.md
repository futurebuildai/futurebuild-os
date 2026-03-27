# Product Requirements Document: Polish & Launch (Phase 15)

| Metadata | Details |
| :--- | :--- |
| **Phase** | Phase 15: Polish & Launch |
| **Goal** | Final UX implementations to ensure a professional, cohesive feel and prepare for Beta release. |
| **Status** | **PLANNING** |
| **Owner** | Product Orchestrator |
| **Authors** | Antigravity, User |
| **Related Roadmap Items** | Steps 90-93 |
| **Estimated Duration** | 3.5 Days |

---

## 1. Executive Summary

Phase 15 is the final sprint before the Beta Release (Version 2.1.0). While previous phases focused on deep functionality (Physics Engine, AI Agents, Auth), Phase 15 addresses the **"Fit and Finish"** gaps that separate a prototype from a product.

The primary focus is on **Accessibility** (Mobile Navigation), **Responsiveness** (System Notifications), and **At-a-Glance Intelligence** (Risk Indicators). This phase concludes with a final regression test and deployment tag.

### Key Outcomes
- **Mobile-First Experience**: A dedicated mobile navigation bar ensures usability on small screens.
- **System Awareness**: Users are notified of critical events via a centralized notification stream.
- **Proactive Risk Management**: Users can identify "at risk" projects from the dashboard without digging into details.
- **Production Readiness**: A stable, tagged release ready for beta users.

---

## 2. Problem Statement

### 2.1 The "Desktop-Only" Gap
Currently, navigation relies on a sidebar that is optimized for desktop. On mobile devices, this sidebar is either hidden or consumes too much screen real estate, making the app difficult to navigate on the go—a critical issue for Superintendents and Builders in the field.

### 2.2 The "Silent Application" Problem
When asynchronous events occur (e.g., an AI agent finishes an analysis, a file upload completes, or a bill is approved), the user has no way of knowing unless they are staring at the specific screen. The app lacks a central nervous system for alerts.

### 2.3 The "Hidden Fire" Issue
To see if a project is off-track, a user currently has to open the project, view the Gantt chart, and analyze the critical path. There is no high-level indicator on the dashboard that screams "requires attention," forcing users to "click-hunt" for problems.

---

## 3. Goals & Success Metrics

### 3.1 Primary Goals

| Goal | Description |
| :--- | :--- |
| **Mobile Accessibility** | Implement a native-app-like bottom navigation bar for screens < 768px. |
| **Centralized Notifications** | Create a Notification Center (Bell Icon) to aggregate alerts, mentions, and system events. |
| **Visual Risk Signaling** | Expose critical project health issues directly on the main dashboard cards. |
| **Beta Stability** | Ensure all regression tests pass and the application is deployed to Production with a version tag. |

### 3.2 Success Metrics

| Metric | Target | Measurement Strategy |
| :--- | :--- | :--- |
| **Mobile Usage** | increase > 20% | Track session counts on mobile breakpoints post-launch |
| **Time to Act** | < 5 seconds | Time from notification click to relevant action/view |
| **Risk Visibility** | 100% of delayed projects | Visual audit: do all delayed projects have a red dot? |
| **Bug Bash** | 0 Critical/High issues | Final regression report findings |

---

## 4. User Stories

### 4.1 Mobile Navigation (Step 90)
> "As a Superintendent on site, I want a thumb-friendly navigation bar so I can quickly switch between Chat and the Schedule without opening a menu."

**Acceptance Criteria:**
- [ ] Bottom tab bar visible ONLY on mobile (< 768px).
- [ ] Sidebar hidden on mobile when tab bar is active.
- [ ] Quick access to: Dashboard, Projects, Chat, Settings.
- [ ] Active state styling for the current tab.

### 4.2 Notification Center (Step 91)
> "As a Project Manager, I want to know immediately when Agent 1 detects a schedule slip so I can intervene."

**Acceptance Criteria:**
- [ ] Bell icon in the top App Shell (with unread badge).
- [ ] Dropdown/Drawer showing list of recent notifications.
- [ ] "Mark all as read" functionality.
- [ ] Clicking a notification deep-links to the relevant context (e.g., specific project or task).

### 4.3 Risk Indicators (Step 92)
> "As an Executive, I want to scan my project list and instantly see which projects are 'bleeding' without opening them."

**Acceptance Criteria:**
- [ ] "Red Dot" indicator on `<fb-project-card>` if project has critical blocking issues.
- [ ] Tooltip/Summary on hover explaining the risk (e.g., "3 Critical Path Delays").
- [ ] Rules engine: Risk = Missed deadlines OR Open Blocking Issues OR Budget Overrun > 10%.

---

## 5. Functional Requirements

### 5.1 Mobile Navigation Component

**Component Name:** `<fb-mobile-nav>`

**Design:**
- **Position:** Fixed bottom (z-index: 1000).
- **Height:** 64px (safe area compliant).
- **Items:** Icons + Labels (10px text).
- **Behavior:** Hides on scroll down (optional), shows on scroll up.

**Items:**
1. **Home** (Dashboard)
2. **Projects** (List View)
3. **Chat** (Global Context)
4. **Menu** (Drawer trigger for less frequent items)

**CSS Strategy:**
```css
@media (min-width: 768px) {
  fb-mobile-nav { display: none; }
}
@media (max-width: 767px) {
  fb-sidebar { display: none; } /* Or hidden behind hamburger */
  fb-mobile-nav { display: flex; }
}
```

### 5.2 Notification System

**Components:**
- `<fb-notification-bell>`: Trigger icon with `badge-count`.
- `<fb-notification-list>`: The scrollable list view.
- `<fb-notification-item>`: Individual card (Icon, Title, Time, Snippet).

**Data Structure (Notification):**
```typescript
interface Notification {
  id: string;
  type: 'system' | 'mention' | 'alert' | 'success';
  title: string;
  message: string;
  link?: string;
  is_read: boolean;
  created_at: string; // ISO
  metadata?: Record<string, any>;
}
```

**Store Logic:**
- Polling (or WebSocket) for new notifications.
- Local optimistic update for "Mark Read".

### 5.3 Risk Indicator Logic

**Component Update:** `<fb-project-card>`

**Logic:**
The risk status is a computed property derived from the Project's "Health Score" or specific flags.

**Risk Rules:**
1. **Critical Delay:** `project.finish_date > project.baseline_finish_date` (threshold: > 2 days).
2. **Blockers:** `active_issues.some(i => i.severity === 'blocker')`.
3. **Budget:** `cost.actual > cost.budget` (threshold: > 10%).

**Visuals:**
- **High Risk:** Pulsing Red Dot + Red Border.
- **Medium Risk:** Orange Dot.
- **Healthy:** No Dot or Green Check.

---

## 6. Technical Design

### 6.1 Frontend Architecture

- **Mobile Nav:** Pure UI component. Uses existing Router links.
- **Notifications:**
    - **Service:** `NotificationService` (fetches `GET /api/v1/notifications`).
    - **Store:** `NotificationStore` (Signal-based).
- **Risk Logic:**
    - Computed in `ProjectCard` or derived in the `ProjectService` via a `calculateRisk(project)` utility.

### 6.2 Backend Architecture

- **Notification Table:** `notifications` (User, Type, Body, ReadStatus).
- **Risk Calculation:** Ideally computed on `GET /projects` summary or pre-calculated via a background job / Trigger if heavy. For Beta, on-the-fly calculation during list fetching is acceptable if dataset is small (< 100 projects).

---

## 7. Implementation Plan

### Step 90: Mobile Navigation
1. Create `components/layout/fb-mobile-nav.ts`.
2. Update `fb-app.ts` to include `<fb-mobile-nav>` conditional on media query.
3. Add icons (Home, Folder, Chat, Menu).
4. Verify routing works on mobile emulation.

### Step 91: Notification UI
1. Create `services/notification-service.ts` (Mocked for now or wired to real endpoint).
2. Create `components/notifications/fb-notification-bell.ts`.
3. Implement dropdown/popover logic.
4. Style `<fb-notification-item>` for different types (Alert vs Info).

### Step 92: Risk Indicators
1. Extend `Project` interface to include `risk_factors` or compute client-side.
2. Update `components/dashboard/fb-project-card.ts`.
3. Add CSS for `.status-dot.risk` and `.status-dot.warning`.
4. Implement tooltip logic for "Why is this red?".

### Step 93: Beta Release
1. Run full E2E test suite.
2. Check for functionality regressions in Chat, Settings, and Gantt.
3. Tag release `v2.1.0-beta`.
4. Deploy to Production environment.

---

## 8. Definition of Done (Phase 15)

- [ ] App is fully navigable on an iPhone SE / Pixel (mobile viewport).
- [ ] Notifications appear and can be dismissed/read.
- [ ] At least one "Risk" scenario can be demonstrated on the Dashboard.
- [ ] `v2.1.0` tag exists in git.
- [ ] Production URL is live and stable.
