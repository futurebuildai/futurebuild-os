# Project Onboarding Specification (Step 62.5)

**Status:** Draft
**Owner:** DevTeam (Architect / Frontend Lead)
**Context:** Filling the gap between User Auth and Project Dashboard (The "Missing Link").

---

## 1. Overview
The **Project Gallery** is the user's home base. It allows them to:
1.  See a visual summary of their active projects (`<fb-project-card>`).
2.  Create a new project via a simple wizard (`<fb-project-wizard>`).
3.  Enter a project context (Dashboard/Chat) by clicking a card.

**Current State:** `<fb-view-projects>` is a placeholder.
**Target State:** A responsive grid of cards + a creation modal.

---

## 2. Architecture & Hierarchy (Architect)

### 2.1 Component Tree
We will introduce a new feature directory `src/components/features/project/` to house these domain components.

```mermaid
graph TD
    View[fb-view-projects] --> Header[Header: 'Projects' + 'New Project' Button]
    View --> Grid[Project Grid Container]
    Grid --> Card[fb-project-card (Repeated)]
    View --> Wizard[fb-project-dialog (Modal Wrapper)]
    Wizard --> Form[fb-project-form]
```

### 2.2 New Components

#### `fb-project-card.ts`
*   **Type:** Dumb / Presentational.
*   **Props:** `project: ProjectSummary`.
*   **Visuals:**
    *   Project Name (H3).
    *   Address (Subtitle).
    *   Status Badge (Pill).
    *   "Enter" Action (Implicit click).
*   **Events:** `@click` -> Dispatches `project-selected` event.

#### `fb-project-dialog.ts`
*   **Type:** Smart / Container.
*   **State:** Open/Closed.
*   **Content:** Wraps the form. Handles overlay/backdrop.

#### `fb-project-form.ts`
*   **Type:** Smart / Form.
*   **Inputs:** Name, Address, GSF (Gross Square Footage).
*   **State:** Form validation, Loading state.
*   **Actions:** Calls `api.projects.create()`.

---

## 3. Implementation Plan (Frontend Developer)

### 3.1 Data Flow
The `Store` and `API` are already ready (verified in `api.ts` and `store.ts`).

1.  **Mount:** `fb-view-projects.connectedCallback()`
    *   Call `api.projects.list()`.
    *   Update `store.actions.setProjects()`.
2.  **Create:** User submits form.
    *   Call `api.projects.create(payload)`.
    *   On Success:
        *   Add new project to `store.projects$`.
        *   Close dialog.
        *   Navigate to `/project/:id/dashboard` (Router).

### 3.2 Styling Strategy (Lit + CSS Variables)
*   **Grid:** `display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: var(--fb-spacing-lg);`
*   **Card:**
    *   `background: var(--fb-bg-card);`
    *   `border: 1px solid var(--fb-border);`
    *   `border-radius: var(--fb-radius-lg);`
    *   Hover effects: `transform: translateY(-2px); box-shadow: var(--fb-shadow-md);`

### 3.3 TypeScript Interfaces
Confirming alignment with `src/types/models.ts` (via `api.ts`).

```typescript
// Payload for creation
interface CreateProjectRequest {
    name: string;
    address: string;
    square_footage: number; // Required by backend
    // Defaults to provide if missing in UI:
    bedrooms: 0,
    bathrooms: 0,
    start_date: new Date().toISOString()
}
```

---

## 4. Functional Requirements (Product)

### 4.1 The "Project Gallery" View
*   **Empty State:** If `projects.length === 0`, show a "Zero Data" hero component.
    *   *Copy:* "Welcome to FutureBuild. Let's start your first project."
    *   *Action:* Big "Create Project" button.
*   **Loading State:** Show 3-4 skeleton cards while fetching.

### 4.2 The "Create Project" Flow (MVP)
*   **Modal Title:** "New Project".
*   **Step 1:** Basic Info.
    *   Name (Required)
    *   Address (Required)
    *   Square Footage (Default to 2500 if skipped, or make required).
*   **Submit:**
    *   Show spinner on button.
    *   Disable inputs.
    *   Handle errors (e.g., API failure) with a toast/alert.

---

## 5. Verification Plan
1.  **Manual:**
    *   Open App -> See Empty State.
    *   Click "Create" -> Fill Form -> Submit.
    *   Verify Card appears in Grid.
    *   Click Card -> Verify navigation to Dashboard.
2.  **Automated:**
    *   Unit test `fb-project-form` validation logic.
    *   (Future) E2E test for the full creation flow.
