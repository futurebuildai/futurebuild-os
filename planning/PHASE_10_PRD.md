# Product Requirements Document: The Brain Connection (Phase 10)

| Metadata | Details |
| :--- | :--- |
| **Phase** | Phase 10: The Brain Connection (Dashboard & Chat) |
| **Goal** | Connect "Daily Focus" and "Chat" views to the live Data Spine. |
| **Status** | **PLANNING** |
| **Owner** | Product Orchestrator |
| **Authors** | Antigravity, User |
| **Related Roadmap Items** | Steps 70-73 |

---

## 1. Executive Summary

Phase 10 serves as the pivotal "Brain Connection" step in the FutureBuild Beta roadmap. Its primary purpose is to bridge the gap between static user interface elements and the dynamic "Data Spine" driven by the project's state. By wiring the Dashboard to real-time agent outputs and activating the Chat interface with file-drop capabilities, we transform the application from a passive viewer into an active, intelligent workspace.

This phase eliminates "Gap 2 (Dashboard Brain)" and "Gap 7 (Chat Interface)" by implementing the specific components required to visualize critical path items and facilitate direct user-to-agent communication via chat and file ingestion.

---

## 2. Problem Statement

### 2.1 The "Static Dead End"
Currently, the Dashboard and Chat views are largely cosmetic.
*   **The Dashboard** displays hardcoded or placeholder data, failing to convey the urgency or status of the actual project.
*   **The Chat** is a non-functional shell. Users cannot actually message agents or upload files, breaking the promise of an "AI-First" construction management platform.

### 2.2 The Disconnected Workflow
Users expect to interact with their project data dynamically. Without a working Chat interface that accepts files (e.g., invoices, blueprints), the core value proposition of "drag-and-drop ingestion" is missing.

---

## 3. Goals & Success Metrics

### 3.1 Primary Goals
1.  **Live Dashboard:** The Dashboard must accurately reflect the `store.focusTasks$` state, showing real-time prioritized tasks.
2.  **Functional Chat UI:** Users must be able to send messages and see a history of interactions.
3.  **Seamless Ingestion:** Dragging a file onto the application must trigger the appropriate chat workflow.

### 3.2 Success Metrics
| Metric | Target | Measurement Strategy |
| :--- | :--- | :--- |
| **Dashboard Latency** | < 200ms | Time from state change to UI update. |
| **Chat Message Success** | 100% | Messages sent appear in the list and trigger API. |
| **File Drop Recognition** | 100% | Valid file types drop triggers upload event. |

---

## 4. Functional Requirements

### 4.1 Dashboard Data Wiring (Step 70)
**Objective:** Replace hardcoded strings in `fb-view-dashboard.ts` with subscriptions to the global state store.

*   **Requirement:** The dashboard must subscribe to `store.focusTasks$`.
*   **Requirement:** Changes to the critical path in the backend must push updates to the dashboard via this subscription.
*   **Component:** `frontend/src/views/fb-view-dashboard.ts`

### 4.2 Focus Card Component (Step 71)
**Objective:** Create a dedicated UI component to visualize high-priority Agent 1 outputs.

*   **Component Name:** `<fb-status-card>`
*   **Content:**
    *   **Critical Path:** Displays top 3 urgent tasks.
    *   **Weather:** Displays generic site weather readiness (e.g., "Pour Window: Open").
*   **Styling:** Must use distinct "Crisis" or "Success" visual states (e.g., Red/Green accents) to denote status.

### 4.3 Chat View Implementation (Step 72)
**Objective:** Build the functional chat interface components.

*   **Component Name:** `fb-view-chat`
*   **Sub-Components:**
    *   `<fb-message-list>`: Renders a scrollable list of `Message` type objects (User vs. System differentiation).
    *   `<fb-input-bar>`: Text input field with "Send" button and specialized actions (e.g., "attach").
*   **Behavior:**
    *   Optimistic UI updates (message appears immediately).
    *   Auto-scroll to bottom on new message.

### 4.4 Drag-to-Chat Wiring (Step 73)
**Objective:** Enable global file drag-and-drop to contextually trigger chat upload.

*   **Event:** `fb-file-drop` (Global window event or specific drop zone).
*   **Action:**
    1.  Detects file drop.
    2.  Prevent default browser behavior (opening file).
    3.  Triggers `api.chat.send` with the file attachment.
    4.  Shows upload progress state in the Chat View.

---

## 5. User Stories

### 5.1 The Morning Check-in
> "As a Project Manager, I want to open the Dashboard and immediately see the 3 tasks that will delay the project if not done today, so I can focus my crew."

### 5.2 The Invoice Drop
> "As a Builder, I want to drag a PDF invoice directly onto the chat window so the AI can ingest it without me typing in the details."

### 5.3 Talking to the Project
> "As a Superintendent, I want to ask 'What's the weather impact on the slab pour?' in the chat and get a stored answer based on the project's weather data."

---

## 6. Implementation Plan (Summary)

### 6.1 Frontend Development
*   **Day 1:** Refactor `fb-view-dashboard.ts` and build `<fb-status-card>`.
*   **Day 2:** Implement `fb-view-chat.ts`, `<fb-message-list>`, and `<fb-input-bar>`. Focus on layout and scrolling.
*   **Day 3:** Implement global drag-and-drop event listener and wire to `api.chat.send`.

### 6.2 Backend Dependencies (Assumed Existing or Mocked)
*   `store.focusTasks$` (Frontend Store)
*   `api.chat.send` (API Endpoint)
*   Back-end WebSocket or Polling for "Live" updates (if applicable for Phase 10, otherwise polling).

---

## 7. Acceptance Criteria

- [ ] **Dashboard:** Displays at least 3 tasks from `store.focusTasks$`.
- [ ] **Component:** `<fb-status-card>` renders correctly in "Critical" and "Normal" states.
- [ ] **Chat:** User can type a message and press Enter to see it appended to the list.
- [ ] **Chat:** "System" messages are visually distinct from "User" messages.
- [ ] **Drag & Drop:** Dragging a `.pdf` or `.png` into the view logs the file object to the console (interim) or triggers the API call.
