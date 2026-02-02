# FutureBuild Phase 14 PRD: Physics Calibration & Gantt

**Version:** 1.0.0
**Status:** DRAFT
**Context:** Phase 14 (Steps 86-89)
**Parent:** `MASTER_PRD.md`

---

## 1. Executive Summary
**Goal:** Empower users to fine-tune the "physics" of the construction engine (timeline speed, work days) and visualize the direct impact of those settings on the Critical Path.

This phase bridges the gap between **Agent Configuration** (Inputs) and **Schedule Visualization** (Outputs).

---

## 2. Feature Set A: Builder Profile (Physics Tuning)
**Context:** Step 86 & 87
**Goal:** Replace complex algorithmic parameters with intuitive "Builder Personality" settings.

### 2.1 UI Requirement: Settings View Update
**File:** `frontend/src/components/views/fb-view-settings.ts`

Add a new section **"Construction Physics"** below the Profile card.

#### a. Speed Slider ("My Pace")
- **Visual:** A range slider with 3 snapped positions.
- **Labels:**
    - **Relaxed:** 1.2x duration padding. (For cautious builders).
    - **Standard:** 1.0x baseline. (Industry average).
    - **Aggressive:** 0.8x duration compression. (For fast-trackers).
- **Binding:** Updates `speed_multiplier` in `business_config`.

#### b. Work Days (Calendar Config)
- **Visual:** A set of checkbox pills.
- **Options:** `Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat`, `Sun`.
- **Default:** Mon-Fri.
- **Logic:** Unchecked days are treated as non-working days in the CPM Solver (Task duration expands to cover them).

### 2.2 Data Requirement: `BusinessConfig`
**Table:** `business_config`
**Scope:** One config per **Organization** (Tenant).

| Field | Type | Description |
| :--- | :--- | :--- |
| `id` | `UUID` | Primary Key |
| `org_id` | `UUID` | Foreign Key to `organizations`. Unique constraint. |
| `speed_multiplier` | `DECIMAL` | Default `1.0`. Range `0.5` - `2.0`. |
| `work_days` | `JSONB` | Array of integers `[1,2,3,4,5]` (Mon-Fri). |
| `created_at` | `TIMESTAMPTZ` | |
| `updated_at` | `TIMESTAMPTZ` | |

### 2.3 API Contract
**Endpoint:** `PUT /api/org/settings/physics`

**Request:**
```json
{
  "speed_multiplier": 0.8,
  "work_days": [1, 2, 3, 4, 5, 6] // Mon-Sat
}
```

**Response:**
- Returns the updated config object.
- **Side Effect:** Triggers a `ScheduleRecalculation` event for all Active projects in the Org (Async/Job).

---

## 3. Feature Set B: Gantt Visualization Upgrade
**Context:** Step 88 & 89
**Goal:** Transform the list-based Gantt into a true dependency graph.

### 3.1 UI Requirement: Critical Path Highlighting
**File:** `frontend/src/components/artifacts/fb-artifact-gantt.ts`

**Logic:**
- The CPM Solver already returns an `is_critical` boolean for each task.
- If `is_critical === true`:
    - **Bar Color:** Change from Primary Blue (`#667eea`) to **Critical Red** (`#c62828`).
    - **Row Background:** Subtle red tint (`rgba(198, 40, 40, 0.05)`).
    - **Label:** Append "🔥" icon or "Critical" badge.

### 3.2 UI Requirement: Dependency Arrows
**File:** `frontend/src/components/artifacts/fb-artifact-gantt.ts`

**Visual:**
- Render an SVG layer `<svg class="dependency-layer">` *on top* of the task bars.
- Draw cubic bezier curves connecting:
    - **Source:** Right edge of Predecessor Bar.
    - **Target:** Left edge of Successor Bar.
- **Styling:**
    - Stroke: `var(--fb-border)` (Grey).
    - Stroke Width: `2px`.
    - Marker End: Arrowhead.

**Interaction:**
- Hovering a task highlights its incoming and outgoing arrows in **Primary Blue**.

---

## 4. Definition of Done
1.  **Persistence:** Changing the "Speed Slider" and refreshing the page retains the value.
2.  **Visual Feedback:** Setting "Aggressive" speed visibly shortens task bars in the Gantt view (simulated or real).
3.  **Criticality:** At least one path of tasks in the Gantt chart is highlighted Red.
4.  **Connectivity:** Dependencies are visually represented by SVG lines, not just implied by dates.
