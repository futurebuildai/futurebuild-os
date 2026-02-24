# Sprint 3.1: The Invoice "Diff" View

> **Epic:** 3 — Intelligent Artifacts (Interactive "Diffs")
> **Depends On:** Sprint 2.1 (ConfidenceReport from VisionService)
> **Objective:** Transform invoices from static forms into AI proposals with confidence scores, highlighting, and hover explanations.

---

## Sprint Tasks

### Task 3.1.1: Update `fb-artifact-invoice.ts` to Accept ConfidenceScores

**Status:** ✅ Complete

**Current State:**
- [fb-artifact-invoice.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/artifacts/fb-artifact-invoice.ts) (666 lines)
- Has `DraftLineItem`, status management, edit/save workflow, currency formatting
- Currently accepts `InvoiceArtifactData` (from `types/artifacts.ts`)
- No confidence score support exists

**Atomic Steps:**

1. **Extend `InvoiceArtifactData` type** (in `types/artifacts.ts`):
   ```ts
   export interface InvoiceFieldConfidence {
       field: string;        // e.g., "line_items[0].description"
       score: number;        // 0.0 - 1.0
       source?: string;      // "vision_extraction" | "manual"
       boundingBox?: {       // PDF region reference  
           page: number;
           x: number; y: number; w: number; h: number;
       };
   }
   
   export interface InvoiceArtifactData {
       // ...existing fields...
       fieldConfidences?: InvoiceFieldConfidence[];
   }
   ```

2. **Add `_confidenceMap` computed property** to `FBArtifactInvoice`:
   - Build a `Map<string, InvoiceFieldConfidence>` from the array for O(1) lookup

3. **Pass confidence to render methods:**
   - `_renderViewRow()` receives per-field confidence
   - `_renderEditRow()` shows confidence badge next to editable fields

---

### Task 3.1.2: Implement "Yellow Highlighting" for Low-Confidence Fields

**Status:** ✅ Complete

**Threshold:** Fields with confidence < 85% (0.85) get highlighted.

**Atomic Steps:**

1. **Add CSS classes:**
   ```css
   .field--low-confidence {
       background: rgba(245, 158, 11, 0.08);
       border-left: 3px solid #f59e0b;
       position: relative;
   }
   .confidence-badge {
       position: absolute;
       top: -8px; right: -8px;
       background: #f59e0b;
       color: white;
       font-size: 10px;
       padding: 2px 6px;
       border-radius: 8px;
   }
   ```

2. **Apply in `_renderViewRow()`:**
   - Check `_confidenceMap.get(fieldKey)?.score < 0.85`
   - If low confidence, add `field--low-confidence` class
   - Show confidence percentage badge

3. **Apply in `_renderEditRow()`:**
   - Same highlighting
   - Add small "⚠ AI extracted — verify" label below low-confidence fields

---

### Task 3.1.3: Add "Hover Explanations" (Source Bounding Boxes)

**Status:** ✅ Complete

**Concept:** Hovering over a field shows a tooltip explaining where in the PDF the value came from, using bounding boxes from VisionService.

**Atomic Steps:**

1. **Create `fb-field-provenance` tooltip component** [NEW]:
   ```ts
   // Shows PDF thumbnail with highlighted bounding box
   @customElement('fb-field-provenance')
   export class FBFieldProvenance extends FBElement {
       @property() pdfUrl = '';
       @property({ type: Object }) boundingBox: BoundingBox;
       @property({ type: Number }) confidence = 1.0;
       // Renders a mini PDF preview with the region highlighted
   }
   ```

2. **Wire hover events** in `fb-artifact-invoice.ts`:
   - On `mouseenter` of a field cell → show `fb-field-provenance` popover
   - On `mouseleave` → hide popover
   - Position popover above or to the side of the field

3. **Fallback:** If no `boundingBox` data exists, show text-only tooltip: "Extracted by AI (85% confident)"

4. **Accessibility:** Ensure popover is keyboard-accessible (`focus`/`blur` events)

---

## Codebase References

| File | Path | Lines | Notes |
|------|------|-------|-------|
| fb-artifact-invoice.ts | `frontend/src/components/artifacts/fb-artifact-invoice.ts` | 666 | Primary modification target |
| InvoiceExtraction.schema.json | `frontend/src/types/schemas/InvoiceExtraction.schema.json` | Existing | May need schema update |
| invoice.ts | `frontend/src/fixtures/invoice.ts` | Existing | Test data fixtures |
| fb-field-provenance.ts | `frontend/src/components/artifacts/` | [NEW] | Hover tooltip component |

## Verification Plan

- **Manual:** View an AI-extracted invoice → verify yellow highlighting on fields with < 85% confidence
- **Manual:** Hover a highlighted field → verify provenance tooltip shows PDF region
- **Manual:** Edit a highlighted field → verify confidence resets and highlighting clears
- **Manual:** View a manually-created invoice → verify no highlighting (all 100% confidence)
