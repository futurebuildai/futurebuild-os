# Spec: Dashboard Data Wiring (Step 70)

| Metadata | Details |
| :--- | :--- |
| **Step ID** | 70 |
| **Component** | `fb-view-dashboard.ts` |
| **Goal** | Connect dashboard metrics to `store.focusTasks$`. |
| **Complexity** | Low |

---

## 1. Context & Scope
The `fb-view-dashboard.ts` component currently displays hardcoded "Lorem Ipsum" data for metrics and recent activity. We need to wire this to the real-time `store.focusTasks$` signal to display high-priority tasks driven by the agentic backend.

## 2. Technical Requirements

### 2.1 Dependency Injection
*   **Import:** `store` from `../../store/store`.
*   **Import:** `FocusTask` type from `../../store/types`.

### 2.2 State Management
*   **Property:** Add `@state() private _focusTasks: FocusTask[] = [];`
*   **Effect:** In `connectedCallback`, set up an effect to subscribe to `store.focusTasks$`.
    ```typescript
    this._disposeEffect = effect(() => {
        this._focusTasks = store.focusTasks$.value;
        this._userName = store.user$.value?.name ?? 'Builder';
    });
    ```

### 2.3 UI Rendering Refactor
*   **Target:** The `.section-title` "Recent Activity" section.
*   **Change:**
    *   If `this._focusTasks` is empty, show the existing placeholder ("Activity feed coming soon...").
    *   If `this._focusTasks` has items, render them using the `<fb-status-card>` component (created in Step 71).
    *   **Note:** Since Step 71 is parallel, initially render a simple list `div` with class `.task-item` containing the task name and priority, or use a temporary placeholder loop.

### 2.4 Metrics Calculation (Bonus)
*   **Active Projects:** Derive from `store.projects$.value.length`.
*   **Tasks Due:** Derive from `store.focusTasks$.value` where `is_critical === true`.

---

## 3. Implementation Steps
1.  **Modify** `frontend/src/views/fb-view-dashboard.ts`:
    *   Import `store`.
    *   Add `_focusTasks` state.
    *   Update `effect` to sync `focusTasks`.
    *   Update `render()` to iterate over `_focusTasks`.

## 4. Verification
*   **Manual:** Open Dashboard.
*   **Console:** Run `store.actions.setFocusTasks([{ id: '1', title: 'Urgent Task', priority: 'high' }])`.
*   **Result:** UI should update immediately to show "Urgent Task".
