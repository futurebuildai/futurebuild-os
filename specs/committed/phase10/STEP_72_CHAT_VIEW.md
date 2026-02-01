# Spec: Chat View Implementation (Step 72)

| Metadata | Details |
| :--- | :--- |
| **Step ID** | 72 |
| **Component** | `fb-view-chat.ts` |
| **Goal** | Build the functional chat interface with message list and input. |
| **Complexity** | Medium |

---

## 1. Context & Scope
`fb-view-chat.ts` is currently a placeholder. We need to implement a full chat interface that displays a history of messages and allows the user to send new ones. This view drives the primary "Agentic" interaction.

## 2. Technical Requirements

### 2.1 State Management
*   **Effect:** Subscribe to `store.messages$` and `store.chatLoading$`.
*   **Action:** Call `api.chat.history(projectId)` on `onViewActive()`.

### 2.2 Sub-Components
#### A. `<fb-message-list>` (Internal or separate file)
*   **Props:** `messages: ChatMessage[]`.
*   **Rendering:**
    *   Scrollable container (`flex: 1`, `overflow-y: auto`).
    *   Map `messages` to message bubbles.
    *   **User Bubble:** Align right, primary color bg.
    *   **Assistant Bubble:** Align left, gray bg.
    *   **System Bubble:** Center, small text, gray.
    *   **Artifacts:** If `message.artifactRef` exists, render a "View Artifact" button or embedded card.

#### B. `<fb-input-bar>` (Internal or separate file)
*   **State:** Local `inputText`.
*   **Events:** `keypress` (Enter to send).
*   **Action:** On send:
    1.  Call `api.chat.send(projectId, text)`.
    2.  Clear input.
    3.  Optimistically add user message to store via `actions.addMessage`.

### 2.3 Layout structure
```html
<div class="chat-container">
    <div class="header">...</div>
    <fb-message-list .messages=${this.messages}></fb-message-list>
    <fb-input-bar ?disabled=${this.isLoading}></fb-input-bar>
</div>
```

---

## 3. Implementation Steps
1.  **Refactor** `frontend/src/components/views/fb-view-chat.ts`.
2.  **Import** `store` and `api`.
3.  **Implement** `render()` with the 3-part layout (Header, List, Input).
4.  **Wire** `api.chat.send` to the input field.

## 4. Verification
*   **Manual:** Type "Hello" and hit Enter.
*   **Result:** Message appears immediately (optimistic). Spinner shows while waiting for response.
*   **Error Handling:** Simulate network failure; verify error toast/state.
