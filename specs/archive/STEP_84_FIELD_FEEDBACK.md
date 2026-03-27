# L7 Spec: Step 84 - Field Feedback Loop

**Context:** Phase 13, Step 84
**Goal:** Implement a robust feedback loop for "Shadow Site" uploads. When a field user drops a photo, they must see immediate confirmation that the AI is "Thinking" (analyzing), followed by a definitive result.

---

## 1. Component Architecture
**File:** `/velocity-frontend/src/components/shared/fb-photo-upload.ts`

### 1.1 State Machine
- **States:**
    - `IDLE`: Waiting for input.
    - `UPLOADING`: Bytes transferring to bucket (0-100%).
    - `ANALYZING`: Upload complete, polling backend for Vision API result.
    - `COMPLETE`: Analysis done, results displayed.
    - `ERROR`: Upload failed or analysis timed out.

### 1.2 Polling Logic
- **Method:** Exponential Backoff Polling.
- **Intervals:** 1s, 2s, 4s, 5s... (Max 30s).
- **Target:** `GET /api/vision/status/:asset_id`
- **Response:** `{ status: "processing" | "completed" | "failed" }`

---

## 2. API Contract

### 2.1 Endpoint
`GET /api/vision/status/:id`

### 2.2 Response Schema
```json
{
  "id": "uuid",
  "status": "processing", // or "completed"
  "analysis": null // populated when status="completed"
}
```

---

## 3. Implementation Steps (Claude Code Instructions)

1.  **Frontend**:
    - Modify `fb-photo-upload`.
    - `handleUpload()`:
        - POST to `/api/upload` -> returns `{ asset_id }`.
        - Set state `ANALYZING`.
        - Start `pollStatus(asset_id)`.
    - `pollStatus(id)`:
        - Fetch status.
        - If `completed`, emit `analysis-complete` event.
        - If `processing`, wait and recurse.

2.  **Backend**:
    - Ensure the "Vision Worker" (or mock) updates the `project_assets` table column `analysis_status`.
    - Implement the lightweight status endpoint.

---

## 4. Verification
- **Network Throttle:** Test with "Slow 3G" to ensure Uploading state is visible.
- **Latency Simulation:** Hardcode backend to sleep 5s before marking "completed" to verify "Analyzing" UI state.
- **Success:** User sees: Upload Bar -> "Verifying..." (Spinner) -> "Verified" (Green Check).
