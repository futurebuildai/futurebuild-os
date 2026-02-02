# FutureBuild Phase 15: Polish & Launch

**PRD Reference:** [PHASE_15_PRD.md](../planning/PHASE_15_PRD.md)
**Objective:** Finalize the "Fit and Finish" of the application, ensuring mobile accessibility, system-wide awareness via notifications, and proactive risk visibility, culminating in a stable Beta Release.

---

## CORE GUARDRAILS & PRODUCT ALIGNMENT
**Alignment Check:** Every step below must be executed with strict adherence to the **FutureBuild Product Vision**:
1.  **"Mobile First Field Work":** The mobile experience (`<fb-mobile-nav>`) must feel native and robust, not like a shrunken desktop site.
2.  **"The Nervous System":** Notifications must be immediate and actionable. The user should never feel "out of the loop."
3.  **"At-a-Glance Intelligence":** Risk indicators must be visually distinct and instantly understandable. No "click-hunting" for bad news.
4.  **"Visual Excellence":** All new components must match the premium aesthetic (glassmorphism, smooth animations, Lucide icons).

**Verification Standard:**
- **L7 Self-Reflection:** Before marking *any* task complete, you must ask: *"Does this feel like a finished, premium product, or a prototype?"*
- **Browser Automation:** YOU MUST use the `/chome` extension (or equivalent) to visually verify every UI change. Code compilation is not enough.
- **Regression Check:** Ensure desktop navigation remains fully functional when mobile nav is implemented.

---

## TASKS

### [ ] Step 90: Mobile Navigation
**Spec:** [STEP_90_MOBILE_NAV.md](../specs/STEP_90_MOBILE_NAV.md)

- [ ] **Frontend**: Create `fb-mobile-nav` component with Home, Projects, Chat, Settings tabs.
- [ ] **Frontend**: Implement CSS media queries in `fb-app` to toggle between Sidebar (Desktop) and Bottom Nav (Mobile).
- [ ] **Frontend**: Ensure `z-index` layering places nav above content but below modals.
- [ ] **Verification**:
    - [ ] **Visual**: Use `/chome` at 375px width to verify visibility and layout.
    - [ ] **Functional**: Click all tabs, verify routing works without full page reload.
    - [ ] **Regression**: Verify Sidebar restores at 1024px.
- [ ] **L7 Audit**: Verify touch targets are >44px and safe-area insets are respected.

### [ ] Step 91: Notification UI
**Spec:** [STEP_91_NOTIFICATION_UI.md](../specs/STEP_91_NOTIFICATION_UI.md)

- [ ] **Frontend**: Implement `NotificationService` (mock data acceptable for Step 91).
- [ ] **Frontend**: Create `fb-notification-bell` with unread badge logic.
- [ ] **Frontend**: Create `fb-notification-list` popover/drawer.
- [ ] **Verification**:
    - [ ] **Visual**: Use `/chome` to click bell and inspect popover styling.
    - [ ] **State**: Verify badge count matches mock data.
- [ ] **L7 Audit**: Ensure accessible focus management (Esc closes popover).

### [ ] Step 92: Risk Indicators
**Spec:** [STEP_92_RISK_INDICATORS.md](../specs/STEP_92_RISK_INDICATORS.md)

- [ ] **Frontend**: Update `fb-project-card` to accept `risk` prop or compute it.
- [ ] **Frontend**: Implement "Red Dot" pulse animation and `.risk-high` border styling.
- [ ] **Frontend**: Add tooltip explaining the risk factor (Delay, Cost, Issues).
- [ ] **Verification**:
    - [ ] **Visual**: Use `/chome` to verify "Project Beta" (risky) vs "Project Alpha" (healthy).
    - [ ] **Logic**: Confirm delay > 2 days triggers high risk.
- [ ] **L7 Audit**: Ensure red color is color-blind friendly (accompanied by label/symbol if possible, or robust tooltip).

### [ ] Step 93: Beta Release
**Spec:** [STEP_93_BETA_RELEASE.md](../specs/STEP_93_BETA_RELEASE.md)

- [ ] **Meta**: Bump version to `2.1.0-beta` in `package.json`.
- [ ] **Meta**: Create git tag `v2.1.0-beta`.
- [ ] **Verification**:
    - [ ] **Smoke Test**: Full critical path walkthrough via `/chome` (Login -> Dashboard -> Project -> Chat).
    - [ ] **Sanity**: Verify version number in UI footer.
- [ ] **L7 Audit**: No console errors during smoke test. Clean network log.

---

## INTER-THREAD PROTOCOL
When executing these steps, you are the **Executor**.
1.  **Read the Spec**: `view_file specs/STEP_XX.md`.
2.  **Execute**: Write code.
3.  **Verify**: Run `/chome` tests.
4.  **Audit**: Self-reflect on quality.
5.  **Commit**: `feat(phase15): ...`
