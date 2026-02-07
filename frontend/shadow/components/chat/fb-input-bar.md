# fb-input-bar

## Intent
*   **High Level:** User input component for chat messages with auto-resizing textarea
*   **Business Value:** Enables users to compose and send messages to AI agents with a polished UX

## Responsibility
*   Renders a textarea with placeholder text and send button
*   Auto-resizes textarea as user types, up to 35% of viewport height
*   Handles Enter key to send (Shift+Enter for newlines)
*   Clears input and resets height after sending
*   Provides file upload button that triggers store.handleFileDrop()

## Key Logic
*   **Auto-resize:** `_autoResize()` calculates `min(scrollHeight, 35vh)` and sets textarea height
*   **Send flow:** `_send()` captures value, clears both reactive state AND DOM element, emits 'send' event
*   **Viewport-relative max height:** Uses `window.innerHeight * 0.35` for responsive expansion
*   **File picker:** Hidden `<input type="file">` triggered by upload button click

## Dependencies
*   **Upstream:** Used by `fb-view-chat`, `fb-onboarding-chat`
*   **Downstream:** Emits `send` event with `{ content: string }`, calls `store.actions.handleFileDrop()`
