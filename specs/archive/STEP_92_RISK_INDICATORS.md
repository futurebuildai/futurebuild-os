# Technical Specification: Risk Indicators (Step 92)

| Metadata | Details |
| :--- | :--- |
| **Step** | 92 |
| **Feature** | Visual Risk Signaling |
| **Goal** | Expose critical project health issues directly on the dashboard cards. |
| **Related** | Phase 15, PRD Section 4.3 |

---

## 1. Feature Description

Users need to identify "at risk" projects without opening them. We will add a "Red Dot" indicator and border styling to `<fb-project-card>` when specific risk conditions are met.

### 1.1 Risk Logic
A project is "High Risk" if:
1. **Delay**: `finish_date` > `baseline_finish_date` (by > 2 days).
2. **Blockers**: Has active issues with `severity: 'blocker'`.
3. **Cost**: `actual_cost` > `budget` (by > 10%).

---

## 2. Components

### 2.1 `fb-project-card`
**Path**: `frontend/src/components/dashboard/fb-project-card.ts`

- **Visual**:
    - **Healthy**: Standard border.
    - **Risk**: Red left-border (`border-left: 4px solid var(--color-error)`).
    - **Indicator**: A pulsing red dot in the top-right corner.
- **Tooltip**: Hovering the red dot shows the specific reason (e.g., "Critical Path Delay: +5 days").

### 2.2 Model Update
**Path**: `frontend/src/models/Project.ts`

- Add `risk_level?: 'low' | 'medium' | 'high'` to the `Project` interface.
- Alternatively, compute it client-side in the `DashboardService`.

---

## 3. Implementation Steps

### Step 3.1: Update Mock Data/Service
- In `DashboardService` or `MockData`, modify one project (e.g., "Project Beta") to have a `finish_date` significantly later than `baseline`.
- Ensure this triggers the risk logic.

### Step 3.2: Update `fb-project-card.ts`
- Add CSS classes for `.card.risk-high`.
- Implement the red dot element:
    ```css
    .risk-dot {
      width: 10px; height: 10px;
      background: red;
      border-radius: 50%;
      position: absolute; top: 10px; right: 10px;
      animation: pulse 2s infinite;
    }
    ```
- Render logic: `html\`${this.project.risk === 'high' ? html`<div class="risk-dot" title="${this.project.riskReason}"></div>` : ''}\``

---

## 4. Verification Plan

### 4.1 Automated Browser Testing (Claude in Chrome)

**CRITICAL INSTRUCTION**: You must use the `/chome` extension (or equivalent Browser Tool) to verify this feature.

**Workflow**:
1. **Launch Browser**: Open `http://localhost:8080`.
2. **Identify Target**: Look for "Project Beta" (or whichever project has the risk flag).
3. **Verify Visuals**:
    - **Red Dot**: Confirm the red dot element exists on the card.
    - **Styling**: Confirm the project card has the red left-border.
4. **Hover Test (Optional)**:
    - Hover over the red dot.
    - Verify the tooltip text (e.g., "Delay detected") appears.
5. **Control Test**:
    - Look at "Project Alpha" (Healthy).
    - Confirm NO red dot and NO red border.

**Auto-Accept**:
- If using `/chome`, assume **Auto-Accept** permissions for localhost testing.
