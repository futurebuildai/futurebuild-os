# FutureBuild: Beta Launch Roadmap (Steps 70-95)

**Version:** 2.1.0 (Beta Remediation + Identity Overhaul)
**Previous Status:** Completed Phase 9 (FutureShade / Step 69)
**Current Focus:** Bridging the gap between "Static UI" and "AI-Driven Action" for Public Beta.

---

## 🚦 Phase 10: The Brain Connection (Dashboard & Chat)
**Goal:** Connect the "Daily Focus" and "Chat" views to the live Data Spine so users feel the AI's presence immediately.

| Status | Step | Task | Gap Addressed | Est. Days |
|--------|------|------|---------------|-----------|
| [x] | 70 | **Dashboard Data Wiring**: Refactor `fb-view-dashboard.ts` to subscribe to `store.focusTasks$` instead of hardcoded strings. | Gap 2 (Dashboard Brain) | 1 |
| [x] | 71 | **Focus Card Component**: Create `<fb-status-card>` to render specific Agent 1 outputs (Critical Path, Weather). | Gap 2 (Dashboard Brain) | 1 |
| [x] | 72 | **Chat View Implementation**: Build the actual `<fb-message-list>` and `<fb-input-bar>` inside `fb-view-chat.ts`. | Gap 7 (Chat Interface) | 1 |
| [x] | 73 | **Drag-to-Chat Wiring**: Connect the global `fb-file-drop` event to trigger `api.chat.send` with attachments. | Gap 7 (Chat Interface) | 1 |

## 🏗️ Phase 11: The Conversational Hook (Smart Onboarding)
**Goal:** Replace the manual form with a "Chat + Upload" wizard where Agent 2 interviews the user to build the project spec.

| Status | Step | Task | Gap Addressed | Est. Days |
|--------|------|------|---------------|-----------|
| [x] | 74 | **Split-Screen Wizard**: Create `fb-view-onboarding` with Chat (Left) and Live Form (Right). | Gap 1 (Creation Workflow) | 1 |
| [x] | 75 | **"The Interrogator" Agent**: Implement backend logic for Agent 2 to ask clarifying questions (e.g., "I see 3 baths, is one a master en-suite?"). | User Request (Interactive) | 2 |
| [x] | 76 | **Real-Time Form Filling**: Wire the chat stream to auto-update the `ProjectDetail` form state as the AI extracts data. | User Request (Structured Data) | 1 |
| [x] | 77 | **Magic Upload Trigger**: Ensure dragging a Blueprint PDF triggers the analysis workflow immediately. | Gap 1 (Creation Workflow) | 0.5 |

## 🆔 Phase 12: Identity & Sovereignty (Auth Refactor)
**Goal:** Replace fragile magic links with a robust Auth Provider (Clerk/Auth0) and implement deep Tenant/Org management.

| Status | Step | Task | Gap Addressed | Est. Days |
|--------|------|------|---------------|-----------|
| [ ] | 78 | **Auth Provider Integration**: Replace `fb-view-login` and backend handlers with Clerk/Auth0 SDKs. | User Request (Auth) | 1 |
| [ ] | 79 | **Middleware Swap**: Update Go middleware to validate 3rd-party JWTs instead of internal tokens. | User Request (Auth) | 1 |
| [ ] | 80 | **Organization Manager**: Build a "Team Settings" view to invite/remove members via the Provider API. | Gap (Org Mgmt) | 2 |
| [ ] | 81 | **Role Mapping**: Map Provider roles (Admin/Member) to FutureBuild's `PermissionMatrix`. | Gap (Org Mgmt) | 0.5 |

## 🔄 Phase 13: The Action Loop (Invoice & Field)
**Goal:** Make Artifacts interactive so users can validate and approve AI decisions.

| Status | Step | Task | Gap Addressed | Est. Days |
|--------|------|------|---------------|-----------|
| [ ] | 82 | **Interactive Invoice**: Rewrite `fb-artifact-invoice.ts` to use `<input>` fields for values (Edit Mode). | Gap 3 (Invoice Loop) | 2 |
| [ ] | 83 | **Approval Actions**: Add "Approve" & "Reject" buttons to artifacts that call `api.finance.approve`. | Gap 3 (Invoice Loop) | 1 |
| [ ] | 84 | **Field Feedback Loop**: Update `fb-photo-upload.ts` to poll `api.vision.status` after upload. | Gap 6 (Field Portal) | 1 |
| [ ] | 85 | **Vision Badges**: Implement "Verifying..." vs "Verified ✅" UI states in the portal. | Gap 6 (Field Portal) | 0.5 |

## ⚙️ Phase 14: Physics Calibration (Settings & Gantt)
**Goal:** Allow users to tune the engine and see the results visually.

| Status | Step | Task | Gap Addressed | Est. Days |
|--------|------|------|---------------|-----------|
| [ ] | 86 | **Builder Profile UI**: Add "Speed" (Slider) and "Work Days" (Checkbox) to `fb-view-settings.ts`. | Gap 4 (Physics Tuning) | 1 |
| [ ] | 87 | **Config Persistence**: Wire settings to update `business_config` table in the backend. | Gap 4 (Physics Tuning) | 1 |
| [ ] | 88 | **Critical Path Visuals**: Add `.critical-path` CSS styling to `fb-artifact-gantt` tasks where `is_critical === true`. | Gap 5 (Gantt) | 1 |
| [ ] | 89 | **Dependency Arrows**: Render SVG connectors between dependent tasks in the Gantt view. | Gap 5 (Gantt) | 2 |

## 💅 Phase 15: Polish & Launch
**Goal:** Final UX implementations to ensure a professional feel.

| Status | Step | Task | Gap Addressed | Est. Days |
|--------|------|------|---------------|-----------|
| [ ] | 90 | **Mobile Navigation**: Implement the bottom tab bar `<fb-mobile-nav>` for screens < 768px. | Gap 8 (Navigation) | 1 |
| [ ] | 91 | **Notification UI**: Build the Notification Bell/Stream in the App Shell. | Gap 10 (Notifications) | 1 |
| [ ] | 92 | **Risk Indicators**: Add the "Red Dot" logic to `<fb-project-card>` based on blocking issues. | Gap 9 (Command Center) | 0.5 |
| [ ] | 93 | **Beta Release Tag**: Final regression test and deployment to Production. | Release | 1 |

---

## Definition of Done (Beta)
1. **No "Lorem Ipsum"**: All dashboards display real database values.
2. **Conversation First**: Project creation is a chat, not a form.
3. **Secure**: Auth is handled by a dedicated provider (Clerk/Auth0).
4. **Loop Closure**: Users receive feedback (Success/Fail) after every action.
