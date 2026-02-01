# Spec: Drag-to-Chat Wiring (Step 73)

| Metadata | Details |
| :--- | :--- |
| **Step ID** | 73 |
| **Component** | Global / `fb-view-chat.ts` |
| **Goal** | Enable drag-and-drop file ingestion to trigger chat uploads. |
| **Complexity** | Medium |

---

## 1. Context & Scope
We want users to be able to drag files (PDFs, Images) onto the application window to initiate an upload to the chat context. The `store` already contains logic for `handleFileDrop`, but the UI needs to listen for these events and provide visual feedback.

## 2. Technical Requirements

### 2.1 Global Listener (App Shell)
*   **Location:** `frontend/src/app-shell.ts` (or main layout component).
*   **Events:**
    *   `dragover`: Prevent default, set `store.actions.setDragging(true)`.
    *   `dragleave`: Set `store.actions.setDragging(false)`.
    *   `drop`: Prevent default, call `store.actions.handleFileDrop(e.dataTransfer.files)`.

### 2.2 Drop Overlay UI
*   **Component:** Create `<fb-drop-overlay>` (or inline in App Shell).
*   **Visibility:** visible when `store.isDragging$` is true.
*   **Style:** Fixed overlay, semi-transparent background, "Drop to Upload" message, centered icon.
*   **Z-Index:** Max (9999).

### 2.3 Chat Integration
*   **Feedback:** When `store.pendingFiles$` changes, the Chat View should show a "Uploading..." indicator or optimistic message (already handled by `store.handleFileDrop` creating a message, but visual verify needed).

---

## 3. Implementation Steps
1.  **Modify** `frontend/src/components/layout/fb-app-shell.ts` (or equivalent root).
2.  **Add** event listeners for `dragover`, `dragleave`, `drop` to `window` or `host`.
3.  **Implement** the Drop Overlay UI keying off `store.isDragging$`.
4.  **Verify** `store.handleFileDrop` correctly processes the file list.

## 4. Verification
*   **Test:** Drag a dummy PDF from desktop to browser.
*   **Expectation:**
    1.  Overlay appears on hover.
    2.  Overlay disappears on drop.
    3.  Console logs "File uploaded".
    4.  Chat shows new message "📎 Uploaded: filename.pdf".
