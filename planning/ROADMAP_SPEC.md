# FutureBuild Production Readiness Roadmap (v1.0)

> **Goal:** Achieve "Game Changer" status for live beta testers and pass the Google Production Readiness Audit.
> **Focus:** Wiring the "Brain" (Backend Agents) to the "Body" (Frontend UI), implementing the Context Spine, and hardening infrastructure.

---

## High-Level Strategy: The "Context Spine"

The central architectural shift: moving from page-based to **Context-Based** navigation.

| Scope | Question It Answers |
|-------|-------------------|
| **Global Context** | "What is the health of my entire company?" (Cross-project aggregated feeds) |
| **Project Context** | "What is blocking this specific job?" (Drill-down into specific artifacts and schedule physics) |

---

## Progress Tracker

| Epic | Sprint | Spec File | Status | Notes |
|------|--------|-----------|--------|-------|
| 1: Context Neural Network | 1.1: Global Store & Router State | [sprint-1-1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-1-1-context-spine.md) | ✅ Complete | Done |
| 1: Context Neural Network | 1.2: Adaptive Navigation | [sprint-1-2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-1-2-adaptive-navigation.md) | ✅ Complete | Done |
| 2: Interrogator Gate | 2.1: Vision Pipeline | [sprint-2-1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-2-1-vision-pipeline.md) | ✅ Complete | Can parallel with EPIC 1 |
| 2: Interrogator Gate | 2.2: Interrogator Interface | [sprint-2-2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-2-2-interrogator-interface.md) | ✅ Complete | Split-screen wizard done |
| 2: Interrogator Gate | 2.3: Physics Trigger | [sprint-2-3](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-2-3-physics-trigger.md) | ✅ Complete | Gate + CPM engine |
| 3: Intelligent Artifacts | 3.1: Invoice Diff View | [sprint-3-1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-3-1-invoice-diff-view.md) | ✅ Complete | Confidence highlighting, hover tooltips |
| 3: Intelligent Artifacts | 3.2: Interactive Learning | [sprint-3-2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-3-2-interactive-learning.md) | ✅ Complete | CorrectionEvent capture, audit WAL spec |
| 4: Real-Time Financials | 4.1: Service Connection | [sprint-4-1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-4-1-service-connection.md) | ✅ Complete | Mock deleted, API wired, shadow specs done |
| 5: Reactive Command Center | 5.1: Agent Feed Aggregation | [sprint-5-1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-5-1-agent-feed-aggregation.md) | ✅ Complete | Shadow specs + frontend SSE wiring done |
| 5: Reactive Command Center | 5.2: Daily Focus Algorithm | [sprint-5-2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-5-2-daily-focus-algorithm.md) | 🟡 Spec Ready | Depends on 5.1 |
| 6: Production Hardening | 6.1: Security & RBAC | [sprint-6-1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-6-1-security-rbac.md) | 🟡 Spec Ready | Can parallel |
| 6: Production Hardening | 6.2: Observability | [sprint-6-2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-6-2-observability.md) | 🟡 Spec Ready | Depends on 3.2 |
| 6: Production Hardening | 6.3: Error Handling | [sprint-6-3](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-6-3-error-handling.md) | 🟡 Spec Ready | Can parallel |

---

## Success Criteria

- [ ] **Context:** Clicking a Project Pill instantly re-filters the Navigation and Dashboard.
- [ ] **Onboarding:** Uploading plans triggers a conversation with the Interrogator, not just a loading spinner.
- [ ] **Physics:** Changing a task duration in the Schedule automatically updates the Critical Path visuals.
- [ ] **Financials:** The "Total Cost" number is real, derived from summed invoices in the DB.
- [ ] **Security:** A standard user cannot access admin routes via URL manipulation.

---

## Epic Details

### EPIC 1: The Context Neural Network (UI Architecture)
**Objective:** Project Pills and Left Nav drive the entire application state. UI reacts instantly to scope changes.
- [Sprint 1.1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-1-1-context-spine.md): `ContextState` in store, `setContext`/`clearContext` actions, URL binding
- [Sprint 1.2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-1-2-adaptive-navigation.md): Left nav adapts links dynamically based on context scope

### EPIC 2: The Interrogator Gate (Onboarding Intelligence)
**Objective:** Schedule is never generated until the AI (Interrogator) has "approved" the extraction.
- [Sprint 2.1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-2-1-vision-pipeline.md): VisionService Go implementation, ConfidenceReport
- [Sprint 2.2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-2-2-interrogator-interface.md): Split-screen wizard (PDF viewer + AI chat)
- [Sprint 2.3](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-2-3-physics-trigger.md): Gate logic, CPM engine connection

### EPIC 3: Intelligent Artifacts (Interactive "Diffs")
**Objective:** Artifacts are "AI Proposals" that the user confirms, not static forms.
- [Sprint 3.1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-3-1-invoice-diff-view.md): Confidence highlighting, hover provenance tooltips
- [Sprint 3.2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-3-2-interactive-learning.md): CorrectionEvent capture, audit WAL logging

### EPIC 4: Real-Time Financials (Destubbing)
**Objective:** Replace all mock data with the live backend engine.
- [Sprint 4.1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-4-1-service-connection.md): Delete mock service, create Go financial handler, live Budget vs. Actual

### EPIC 5: The Reactive Command Center (Dashboard)
**Objective:** Dashboard is a "Feed" aggregating insights from all Agents, not a "Menu".
- [Sprint 5.1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-5-1-agent-feed-aggregation.md): Backend FeedAggregator, SSE push wiring
- [Sprint 5.2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-5-2-daily-focus-algorithm.md): Priority scoring, traffic-light styling, filter controls

### EPIC 6: Production Hardening (Google Audit)
**Objective:** Security, Stability, Scalability — non-functional requirements for production.
- [Sprint 6.1](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-6-1-security-rbac.md): Route audit, role middleware, access control tests
- [Sprint 6.2](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-6-2-observability.md): Structured agent logging, OpenTelemetry tracing
- [Sprint 6.3](file:///home/colton/Desktop/FutureBuild_HQ/XUI/planning/sprints/sprint-6-3-error-handling.md): Enhanced error boundary, AI graceful degradation
