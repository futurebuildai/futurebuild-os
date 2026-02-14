# Sprint 2.2: The Interrogator Interface

> **Epic:** 2 — The Interrogator Gate (Onboarding Intelligence)
> **Depends On:** Sprint 2.1 (Vision Pipeline)
> **Objective:** Implement the "Split Screen" wizard that pairs PDF viewing with AI interrogation chat.

---

## Sprint Tasks

### Task 2.2.1: Implement the "Split Screen" Wizard

**Status:** ⬜ Not Started

**Target File:** `frontend/src/components/features/onboarding/fb-engine-calibration.ts` [NEW]

**Current State:**
- Existing onboarding components:
  - [fb-onboard-flow.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/features/onboarding/fb-onboard-flow.ts) — current flow container
  - [fb-onboarding-chat.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/features/onboarding/fb-onboarding-chat.ts) — existing chat panel
  - [fb-onboarding-dropzone.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/features/onboarding/fb-onboarding-dropzone.ts) — file upload
  - [fb-onboarding-steps.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/features/onboarding/fb-onboarding-steps.ts) — progress steps
- [onboarding-store.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/store/onboarding-store.ts) — state management with stages: `upload → extract → details → review`

**Design:**

```
┌──────────────────────────────────────────────────────────────┐
│                    Progress Bar (Steps)                       │
├─────────────────────────────┬────────────────────────────────┤
│                             │                                │
│     PDF Viewer              │     AI Chat                    │
│     (Highlighting           │     "I found 30 electrical     │
│      extraction zones)      │      tasks, but no sub-panel   │
│                             │      installation. Is that     │
│     ┌─────────────────┐     │      existing?"               │
│     │  Bounding Box   │     │                                │
│     │  for extracted   │     │     [User Reply Input]         │
│     │  item            │     │                                │
│     └─────────────────┘     │     [Extraction Card]          │
│                             │                                │
├─────────────────────────────┴────────────────────────────────┤
│                  Action Bar: [Generate Schedule]              │
└──────────────────────────────────────────────────────────────┘
```

**Atomic Steps:**

1. **Create `fb-engine-calibration.ts`** as a new Lit component extending `FBElement`
2. **Layout:** CSS Grid with two columns — left (PDF viewer), right (AI chat)
3. **Left Panel — PDF Viewer:**
   - Embed PDF.js viewer (`<canvas>` or `<iframe>` with pdf.js)
   - Accept `pdfUrl` prop from parent
   - Accept `highlights: BoundingBox[]` prop to overlay extraction zones
   - Use SVG overlays on top of the PDF canvas for bounding boxes
4. **Right Panel — AI Chat:**
   - Reuse `fb-onboarding-chat` or create a simplified version
   - Connect to `onboardingMessages` signal
   - Show extraction cards inline (confidence badges per field)
5. **Responsive:** On mobile, stack vertically (PDF on top, chat below)
6. **Register** the component and import it in `fb-view-onboarding.ts`

---

### Task 2.2.2: Wire "Answer" Input to InterrogatorService

**Status:** ⬜ Not Started

**Current State:**
- Backend [interrogator_service.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/service/interrogator_service.md) documents the flow:
  - `POST /api/v1/agent/onboard` — accepts text/document messages
  - Returns `extracted_values`, `confidence_scores`, `reply`, `ready_to_create`
- Frontend `fb-onboarding-chat.ts` already sends messages

**Atomic Steps:**

1. **Verify API contract:** Ensure frontend sends `{ message: string, project_context: {} }` to `/api/v1/agent/onboard`
2. **On user reply:** Call `applyAIExtraction()` with response's `extracted_values` and `confidence_scores`
3. **Update PDF highlights:** When AI references specific document regions, update bounding box overlays
4. **Handle `ready_to_create`:** When backend returns `ready_to_create: true`, enable the "Generate Schedule" button
5. **Error handling:** If API fails, show friendly error in chat, don't crash the wizard

---

## Codebase References

| File | Path | Lines | Notes |
|------|------|-------|-------|
| fb-engine-calibration.ts | `frontend/src/components/features/onboarding/` | [NEW] | Split-screen wizard |
| fb-onboard-flow.ts | `frontend/src/components/features/onboarding/fb-onboard-flow.ts` | Existing | Current flow container |
| fb-onboarding-chat.ts | `frontend/src/components/features/onboarding/fb-onboarding-chat.ts` | Existing | Chat panel (can extend) |
| onboarding-store.ts | `frontend/src/store/onboarding-store.ts` | 247 | State: stages, confidence, extraction |
| interrogator_service.md | `backend/shadow/internal/service/interrogator_service.md` | 32 | API contract documented |

## Verification Plan

- **Manual:** Upload a PDF → verify split-screen renders (PDF left, chat right)
- **Manual:** Click on an extracted item in chat → verify PDF scrolls/highlights the region
- **Manual:** Type a response → verify extraction values update in real-time
- **Manual:** Mobile viewport → verify vertical stack layout
