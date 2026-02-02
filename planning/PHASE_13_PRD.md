# Phase 13 PRD: The Action Loop (Invoice & Field)

**Version:** 1.0.0
**Status:** Draft
**Context:** ROADMAP.md (Step 82-85) | Bridging "Static Artifacts" to "Interactive Decisions"

---

## 1. Executive Summary

**Goal:** Transform static, read-only artifacts (Invoices, Estimates, Site Photos) into interactive decision points. Users must be able to correct AI assumptions (edit values) and definitively "Approve" or "Reject" work. This closes the loop between "AI Insight" and "Human Action," giving the system the confirmation it needs to proceed.

**Why:**
- **Trust:** Users won't trust an AI that generates a $50k invoice if they can't fix a $500 mistake.
- **Workflow:** Moving from "Observation" to "Project Management" requires state changes (Draft -> Approved).
- **Latency:** Immediate feedback on field photos (e.g., "Verifying...") prevents users from wondering if their upload worked.

---

## 2. User Stories

### 2.1 The General Contractor
> "When the AI estimates the plumbing cost, I want to be able to click on the line item and adjust the labor rate before I hit 'Send', because I know my plumber just raised his prices."

### 2.2 The Project Manager
> "I need to explicitly 'Approve' a sub's invoice so the system knows it's okay to pay. Just seeing it isn't enough; I need a button that records my decision."

### 2.3 The Field Tech
> "After I upload a photo of the framing, I want to know instantly that the system is checking it. Seeing a 'Verifying...' badge makes me feel confident the AI is working, and seeing 'Verified' let's me leave the site knowing I'm done."

---

## 3. Functional Requirements

### 3.1 Interactive Invoice (Step 82)
- **Edit Mode:**
    - `fb-artifact-invoice` must support a toggle (or always-on) edit state.
    - Click-to-edit for: Quantity, Unit Price, Description.
    - Automatic total recalculation on client-side when inputs change.
- **Validation:**
    - Ensure inputs are numbers where appropriate.
    - Prevent negative values if business logic forbids it.

### 3.2 Approval Actions (Step 83)
- **Action Bar:**
    - Sticky footer or header on artifacts (Invoice, Estimate).
    - Primary Action: "Approve" (Green) -> Calls `api.finance.approve`.
    - Secondary Action: "Reject" (Red) -> Opens a feedback modal (optional for now, or just updates state).
- **State Reflection:**
    - Once approved, the artifact locks (read-only).
    - UI updates status badge from "Draft" to "Approved".

### 3.3 Field Feedback Loop (Step 84)
- **Immediate Feedback:**
    - `fb-photo-upload` must show specific state during processing, not just a generic spinner.
- **Polling/Sockets:**
    - Poll `api.vision.status` (or similar endpoint) to track analysis progress.
    - Timeout handling if analysis takes > 30s.

### 3.4 Vision Badges (Step 85)
- **Visual Status:**
    - **Verifying...** (Yellow/Pulsing): AI is analyzing.
    - **Verified ✅** (Green): Compliance check passed.
    - **Flagged ⚠️** (Red): Issue detected (e.g., "Missing safety rail").
- **Overlay:**
    - Badges should overlay the image thumbnail in the gallery or list view.

---

## 4. Technical Architecture

### 4.1 Frontend (`/velocity-frontend`)
- **Components:**
    - Update `fb-artifact-invoice.ts`: Replace `<span>` text displays with `<input>` or custom generic `<fb-inline-edit>` components.
    - Update `fb-photo-upload.ts`: Implement polling logic using `setInterval` or RxJS `timer`.
- **State Management:**
    - Local state for draft edits (don't persist every keystroke, persist on blur or "Save").
    - Optimistic UI updates for "Approve" actions.

### 4.2 Backend (`/velocity-backend`)
- **API Endpoints:**
    - `POST /api/finance/invoice/:id/approve`: Transition state, log actor.
    - `PUT /api/finance/invoice/:id`: Update line items (bulk update).
    - `GET /api/vision/status/:id`: Lightweight status check for uploaded assets.
- **Data Model:**
    - Ensure `invoices` table supports `status` enum (`draft`, `approved`, `rejected`, `paid`).
    - Ensure `project_assets` / `photos` supports `analysis_status` and `analysis_result`.

---

## 5. Definition of Done

1.  **Edits Works:** I can change a line item price on an invoice, and the total updates. I can save this change.
2.  **Approval Works:** Clicking "Approve" locks the invoice and updates its status in the DB.
3.  **Feedback Loop:** Uploading a photo shows "Verifying...", which flips to "Verified" or "Flagged" once the mock backend returns a result.
4.  **No Dead Clicks:** Every interactive element provides visual feedback (hover, active, disabled states).

---

## 6. Success Metrics

- **Edit Frequency:** % of AI-generated invoices modified by users (measure of AI accuracy vs. user control).
- **Time to Approve:** Avg. time from generation to approval action.
- **Upload Confidence:** qualitative user trust (measured via reduced "did it work?" re-uploads).
