# L7 Spec: Step 85 - Vision Badges

**Context:** Phase 13, Step 85
**Goal:** Visually communicate the state of "AI Verification" on project assets. Users need to know *at a glance* which photos have been checked by the AI and which have issues.

---

## 1. Component Architecture
**File:** `/velocity-frontend/src/components/shared/fb-badge.ts` (New or Extend)
**File:** `/velocity-frontend/src/components/project/fb-photo-gallery.ts`

### 1.1 UI/UX
- **Badge Variants:**
    - `VERIFYING`:
        - Color: Yellow (`--color-warning-500`)
        - Icon: Spinner (Animated)
        - Text: "Verifying..."
    - `VERIFIED`:
        - Color: Green (`--color-success-500`)
        - Icon: Check-circle (`heroicons:check-circle`)
        - Text: "Verified"
    - `FLAGGED`:
        - Color: Red (`--color-danger-500`)
        - Icon: Exclamation (`heroicons:exclamation-triangle`)
        - Text: "Flagged"
- **Placement:**
    - Absolute positioned overlay on the top-right of image thumbnails.
    - Inline next to the file name in list views.

### 1.2 Interaction
- **Click Actions:**
    - Clicking `FLAGGED` badge opens the "Issue Details" modal (showing AI reasoning).
    - Clicking `VERIFIED` shows "Compliance Passed".

---

## 2. API Contract
**Endpoint:** `GET /api/projects/:id/assets`

### 2.1 Schema Extension
The asset object now includes:
```json
{
  "id": "123",
  "url": "...",
  "analysis": {
    "status": "flagged",
    "summary": "Missing safety rail on 2nd floor balcony.",
    "confidence": 0.95
  }
}
```

---

## 3. Implementation Steps (Claude Code Instructions)

1.  **Create Components**:
    - `fb-status-badge`: A generic pill component with status mapping.
    - Styles: Use Tailwind utility classes for colors/rounding.

2.  **Integrate into Gallery**:
    - `fb-photo-gallery` iterates over assets.
    - Render `<fb-status-badge .status=${asset.analysis.status}>` on top of the `<img>`.

3.  **Mock Data**:
    - Ensure your local mock server returns varied states for testing.

---

## 4. Verification
- **Visual Regression:**
    - Verify the "Verifying" spinner is actually spinning.
    - Verify text contrast is accessible (white text on dark badge background).
- **Behavior:**
    - Hovering a "Flagged" badge should show a tooltip with the summary.
- **Responsiveness:**
    - Badges should scale down or disappear on very small thumbnails if they obscure too much.
