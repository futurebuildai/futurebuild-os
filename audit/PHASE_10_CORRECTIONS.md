# Phase 10 Remediation Instructions

**Target Agent:** Claude Code
**Context:** L7 Audit Failure
**Priority:** P0 (Critical)

---

## 🚨 Critical Architecture Correction Required

The implementation of Phase 10 (`fb-view-chat.ts`) deviated from the approved specification by using WebSockets (`realtimeService`) for message transport instead of the mandated REST API (`api.chat.send`). This introduced significant reliability risks (silent message loss).

### 1. Revert to REST Transport
**File:** `frontend/src/components/views/fb-view-chat.ts`
**Goal:** Replace fire-and-forget WebSocket calls with robust `await api.chat.send()`.

**Instructions:**
1.  **Remove** `realtimeService.send(...)` from `_handleSend`.
2.  **Add** `await api.chat.send(this._projectId, content)` inside `_handleSend`.
3.  **Handle Errors:** Wrap the API call in a `try/catch` block.
    *   On Error: Remove the optimistically added message from the store (rollback) AND set `store.chatError$` to notify the user.
4.  **Loading State:** Ensure `store.actions.setChatLoading(true)` is called before sending (optional, given optimistic UI) or at least ensure the UI remains responsive.

### 2. Fix File Upload Implementation (Step 73)
**File:** `frontend/src/store/store.ts` (Action: `handleFileDrop`)
**Goal:** Ensure file uploads also use a reliable transport or verified acknowledgment.

**Instructions:**
1.  Check `handleFileDrop`. If it uses `realtimeService.send` for the file, **CHANGE IT** to use an `api.chat.upload` endpoint (if available) or `api.chat.send` with attachment logic.
    *   *Note:* If the backend only supports file upload via specific endpoint, ensure that is used. If the definition was to use `api.chat.send`, strictly follow `STEP_73_DRAG_DROP.md`.

### 3. Verify Spec Compliance
*   **Reference:** `specs/phase10/STEP_72_CHAT_VIEW.md`
*   **Requirement:** "Action: On send: 1. Call `api.chat.send(projectId, text)`." -> **Verify this line exists exactly.**

---

## Execution Checklist

- [ ] Modify `fb-view-chat.ts` to use `api.chat.send`.
- [ ] Implement `try/catch` error handling in `_handleSend`.
- [ ] Implement rollback logic for failed optimistic messages.
- [ ] Verify `handleFileDrop` in `store.ts` logic aligned with Spec 73.
