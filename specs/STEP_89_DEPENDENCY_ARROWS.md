# L7 Spec: Step 89 - Dependency Arrows (Gantt)

**Context:** Phase 14, Step 89
**Goal:** Render Bezier curve arrows connecting dependent tasks in the Gantt chart.
**Prerequisites:** Step 88 (Gantt Visuals) is helpful.

---

## 1. Component Architecture
**File:** `frontend/src/components/artifacts/fb-artifact-gantt.ts`

### 1.1 SVG Layer
We cannot use standard HTML elements easily for diagonal lines. We must overlay an `<svg>` that matches the dimensions of the `.gantt-container`.

- **Coordinate System:**
    - X-axis: Time (mapped to pixels via `date/totalDuration`).
    - Y-axis: Row height (index * rowHeight + padding).

### 1.2 The Connector Logic (`_renderConnections`)
For each dependency `(Predecessor A -> Successor B)`:
1.  **Start Point (x1, y1):** `A.endX`, `A.y + A.height/2`.
2.  **End Point (x2, y2):** `B.startX`, `B.y + B.height/2`.
3.  **Control Points:**
    - `cp1`: `(x1 + 20, y1)`
    - `cp2`: `(x2 - 20, y2)`
4.  **Path:** `<path d="M x1 y1 C cp1 y1, cp2 y2, x2 y2" ... />`

### 1.3 Styling
- **Stroke:** `var(--fb-border)` (Grey #e5e7eb or #333 in dark mode).
- **Stroke-Width:** `1.5px`.
- **Marker-End:** Define an arrow marker in `<defs>`.
- **Hover:** If mouse is over Task A or B, set stroke color to `var(--fb-primary)` and `z-index: 10`.

---

## 2. Implementation Steps (Claude Code Instructions)

1.  **Modify `fb-artifact-gantt.ts`**:
    - Wrap the content in a `position: relative` container.
    - Add an `<svg>` overlay with `position: absolute; top: 0; left: 0; pointer-events: none;`.
    - Implement `_calculateCoordinates(task)` helper.
    - Implement `_renderDependencyLines()` method looping through `data.dependencies`.

---

## 3. Automated Verification Logic
**Tool:** `/chome` (Claude in Chrome)

**Instructions for the Agent:**
Execute the following verification script using the browser tool:

1.  **Navigate:** Go to `http://localhost:5173/gantt`.
2.  **Verify UI:**
    - Inspect the DOM to ensure an `<svg>` element exists within the Gantt component.
    - **Visual Check:** Verify curved lines connect tasks.
    - **Interactive Check:** Hover over a Task Bar -> Verify the connected lines turn Blue.
3.  **Screenshot:** Capture the chart with arrows visible.

> **Visual Test Command:**
> `/chome "Go to the Gantt view, verify that dependency arrows are drawn between tasks, hover over a task to see them highlight, and capture a screenshot." --auto-accept`
