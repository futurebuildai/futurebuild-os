# L7 Spec: Step 88 - Critical Path Visuals (Gantt)

**Context:** Phase 14, Step 88
**Goal:** Visually highlight tasks on the Critical Path in the Gantt chart.
**Prerequisites:** Step 86/87 (Persistence) is NOT strictly required, but CPM logic must be calculating `is_critical`.

---

## 1. Component Architecture
**File:** `frontend/src/components/artifacts/fb-artifact-gantt.ts`

### 1.1 Visual Requirements
The backend CPM Solver already computes `is_critical: true` for tasks that have zero float (are on the longest path).

#### Task Bar (`.task-bar`)
- **Default:** `var(--fb-primary)` (Blue/Green).
- **Critical:** `var(--fb-error)` (Red #c62828).
    - Requires CSS class `.is-critical`.

#### Task Row (`.task-row`)
- **Critical:** Add a subtle background highlight: `background: rgba(198, 40, 40, 0.05);`.
- **Icon:** Add a "🔥" (fire emoji) or a small SVG flame icon next to the task name if critical.

### 1.2 Interaction
- **Tooltip:** Hovering the bar should show "Critical Path: Yes" in the metadata.

---

## 2. Implementation Steps (Claude Code Instructions)

1.  **Modify `fb-artifact-gantt.ts`**:
    - Update `_renderTaskRow` to check `task.is_critical`.
    - Apply conditional classes.
    - Add styles for `.task-bar.critical` and `.task-row.critical`.
    - Inject the icon into the task name.

---

## 3. Automated Verification Logic
**Tool:** `/chome` (Claude in Chrome)

**Instructions for the Agent:**
Execute the following verification script using the browser tool:

1.  **Navigate:** Go to `http://localhost:5173/gantt` (or the dashboard view containing the gantt).
2.  **Verify UI:**
    - Locate at least one Task Bar that is Red (`#c62828`).
    - Locate at least one Task Bar that is Blue/Default.
    - **Assert:** The Red task row has a background color.
3.  **Screenshot:** Capture the full Gantt chart to verify the path continuity.

> **Visual Test Command:**
> `/chome "Go to the Gantt view, verify that critical tasks are highlighted in red with a flame icon, and capture a screenshot." --auto-accept`
