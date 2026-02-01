# Spec: Focus Card Component (Step 71)

| Metadata | Details |
| :--- | :--- |
| **Step ID** | 71 |
| **Component** | `fb-status-card.ts` |
| **Goal** | Create a reusable component for displaying Agent 1 status updates. |
| **Complexity** | Low |

---

## 1. Context & Scope
The dashboard needs to display high-priority tasks and status updates (Agent 1 outputs). We need a reusable `<fb-status-card>` component that can handle different types of content (Critical Path tasks, Weather alerts) with appropriate visual styling.

## 2. Technical Requirements

### 2.1 File Location
*   `frontend/src/components/widgets/fb-status-card.ts`

### 2.2 Inputs (Properties)
*   `@property({ type: String }) type`: 'critical' | 'info' | 'success' | 'warning' (Default: 'info')
*   `@property({ type: String }) title`: The main headline (e.g., "Pour Window Open").
*   `@property({ type: String }) subtitle`: Secondary text (e.g., "3 days remaining").
*   `@property({ type: String }) icon`: Material symbol name (optional).

### 2.3 Styling
*   **Base:** Card style (`.card`).
*   **Variants:**
    *   `type="critical"`: Red border/background tint.
    *   `type="success"`: Green border/background tint.
*   **Layout:** Flexbox, icon left, title/subtitle stacked right.

### 2.4 Usage Example
```html
<fb-status-card
    type="critical"
    title="Foundation Inspection"
    subtitle="Due today at 2:00 PM"
    icon="priority_high"
></fb-status-card>
```

---

## 3. Implementation Steps
1.  **Create** `frontend/src/components/widgets/fb-status-card.ts`.
2.  **Define** class `FBStatusCard` extending `LitElement`.
3.  **Implement** styles and properties.
4.  **Register** custom element `fb-status-card`.

## 4. Verification
*   **Test:** Create a test harness or add to `fb-view-dashboard` temporarily.
*   **Check:** Verify all 4 types render with correct colors.
