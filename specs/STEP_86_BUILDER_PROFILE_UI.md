# L7 Spec: Step 86 - Builder Profile UI

**Context:** Phase 14, Step 86
**Goal:** Implement the "Construction Physics" settings UI (Speed Slider & Work Days) in `fb-view-settings.ts`.
**Prerequisites:** None (Frontend only first, distinct from persistence).

---

## 1. Component Architecture
**File:** `frontend/src/components/views/fb-view-settings.ts`

### 1.1 State Management
- **New State Properties:**
    - `_speedMultiplier`: number (1.0 default).
    - `_workDays`: number[] (default [1,2,3,4,5]).
    - `_physicsDirty`: boolean (tracks changes to these specific fields).

### 1.2 UI Components
Add a new card `<div class="settings-card">` with title "Construction Physics".

#### Speed Slider
- **Label:** "My Pace (Schedule Padding)"
- **Element:** `<input type="range" min="0.5" max="1.5" step="0.1">`
- **Visual Feedback:**
    - Render a text badge next to the slider value:
        - `0.5-0.8`: "Aggressive (Fast Track)"
        - `0.9-1.1`: "Standard (Industry Avg)"
        - `1.2-1.5`: "Relaxed (Padding Added)"

#### Work Days
- **Label:** "Standard Work Week"
- **Element:** Flex container of toggle buttons.
- **Rendering:** Loop `['M', 'T', 'W', 'T', 'F', 'S', 'S']`.
- **Interaction:** Click toggles inclusion in `_workDays` array.
- **Style:**
    - Active: `bg-primary text-white`
    - Inactive: `bg-tertiary text-muted`

---

## 2. Implementation Steps (Claude Code Instructions)

1.  **Modify `fb-view-settings.ts`**:
    - Add the new state properties.
    - Implement `_renderPhysicsCard()` method.
    - Add the slider and day toggles.
    - wire `input` events to update local state.
    - **Note:** For this step, simply log the values to console on "Save". Persistence comes in Step 87.

2.  **Style Updates**:
    - Ensure the slider and toggles match the "Premium Construction" aesthetic (Dark mode, accessible touch targets).

---

## 3. Automated Verification Logic
**Tool:** `/chome` (Claude in Chrome)

**Instructions for the Agent:**
Execute the following verification script using the browser tool:

1.  **Navigate:** Go to `http://localhost:5173/settings`.
2.  **Verify UI:**
    - Assert "Construction Physics" card is visible.
    - Take a screenshot.
3.  **Interact:**
    - Drag the slider to "0.8" -> Verify text changes to "Aggressive".
    - Click "S" (Saturday) -> Verify it turns Green/Primary color.
    - Click "S" (Sunday) -> Verify it turns Green.
4.  **Console Check:**
    - Open console logs.
    - Click "Save Changes".
    - **Assert:** Console prints JSON with `speed_multiplier: 0.8` and `work_days` including 6 & 7.

> **Visual Test Command:**
> `/chome "Go to localhost:5173/settings, scroll to the Physics card, set speed to Aggressive, enable Saturday, and verify the UI updates correctly. Take a screenshot." --auto-accept`
