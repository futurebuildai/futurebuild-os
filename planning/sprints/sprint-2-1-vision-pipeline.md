# Sprint 2.1: The Vision Pipeline & Extraction

> **Epic:** 2 — The Interrogator Gate (Onboarding Intelligence)
> **Depends On:** None (can run in parallel with EPIC 1)
> **Objective:** Ensure the VisionService correctly parses uploaded plans and the DocumentHandler returns confidence data.

---

## Sprint Tasks

### Task 2.1.1: Verify VisionService Parses Uploaded Plans

**Status:** ⬜ Not Started

**Current State:**
- Backend: [vision_service.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/service/vision_service.md) — placeholder stub, no Go implementation found
- Backend: [vision_handler.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/api/handlers/vision_handler.md) — placeholder stub
- Frontend: onboarding-store has `hasDocumentUploaded`, `isProcessing`, `applyAIExtraction()` ready
- The Interrogator service spec mentions `extractFromDocument/extractFromBytes` using Vision API

**Required Implementation (Backend Go):**

1. **Create `backend/internal/service/vision_service.go`:**
   - `VisionService` struct with Vertex AI / Gemini Vision client
   - `ParseDocument(ctx, fileBytes []byte, mimeType string) (*ExtractionResult, error)`
   - Returns structured data: WBS candidates, task names, material lists
   - Uses Gemini multimodal prompts to extract construction plan data

2. **Create `backend/internal/api/handlers/vision_handler.go`:**
   - `POST /api/v1/vision/extract` — multipart upload endpoint
   - Accepts PDF/image, calls VisionService, returns JSON
   - Must validate file size (max 20MB) and MIME type

3. **Wire to existing onboarding flow:**
   - The Interrogator's `ProcessMessage()` should call VisionService when document is uploaded
   - Frontend `fb-onboarding-dropzone.ts` already handles file selection

**Atomic Steps:**

1. Define `ExtractionResult` Go struct: `Tasks []TaskCandidate`, `Materials []Material`, `RawText string`, `Confidence float64`
2. Implement Gemini Vision API call with construction-specific system prompt
3. Parse Gemini response into `ExtractionResult`
4. Create HTTP handler with multipart form parsing
5. Register route in router
6. Test with sample construction plan PDF

---

### Task 2.1.2: Enhance DocumentHandler to Return `ConfidenceReport`

**Status:** ⬜ Not Started

**Current State:**
- [document_handler.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/api/handlers/document_handler.md) — placeholder stub
- Frontend `onboarding-store.ts` already has `onboardingConfidence` signal and `confidenceScores` parameter in `applyAIExtraction()`

**Required Implementation:**

1. **Define `ConfidenceReport` struct:**
   ```go
   type ConfidenceReport struct {
       OverallConfidence float64                    `json:"overall_confidence"`
       FieldConfidences  map[string]float64         `json:"field_confidences"`
       Warnings          []string                   `json:"warnings"`
       SuggestedQuestions []string                  `json:"suggested_questions"`
   }
   ```

2. **Update extraction response** to include `ConfidenceReport` alongside `extracted_values`

3. **Frontend wire-up:** The `applyAIExtraction()` function in `onboarding-store.ts` already accepts `confidenceScores: Record<string, number>` — just need to pass through the backend response

**Atomic Steps:**

1. Define `ConfidenceReport` in `backend/internal/models/`
2. Update VisionService to compute per-field confidence scores
3. Update document/onboarding handler response to include confidence report
4. Verify frontend `applyAIExtraction()` receives and stores confidence data
5. Verify `fieldsNeedingVerification` computed signal flags low-confidence fields

---

## Codebase References

| File | Path | Status | Notes |
|------|------|--------|-------|
| vision_service.md | `backend/shadow/internal/service/vision_service.md` | Stub | Needs Go implementation |
| vision_handler.md | `backend/shadow/internal/api/handlers/vision_handler.md` | Stub | Needs Go implementation |
| document_handler.md | `backend/shadow/internal/api/handlers/document_handler.md` | Stub | Needs Go implementation |
| interrogator_service.md | `backend/shadow/internal/service/interrogator_service.md` | Documented | Has ProcessMessage flow spec |
| onboarding-store.ts | `frontend/src/store/onboarding-store.ts` | Ready | Has confidence signals and extraction logic |
| fb-onboarding-dropzone.ts | `frontend/src/components/features/onboarding/fb-onboarding-dropzone.ts` | Exists | File upload UI exists |

## Verification Plan

- **Automated:** Upload a test PDF via `/api/v1/vision/extract`, verify JSON response contains `tasks`, `materials`, `confidence`
- **Manual:** Upload a construction plan in the UI, verify extraction card appears in onboarding chat
- **Manual:** Verify low-confidence fields show "Verify" badge in the UI
