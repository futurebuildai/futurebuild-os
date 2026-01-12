# Google A2UI Protocol Reference

**Official Documentation & Resources:**
*   **Website:** [a2ui.org](https://a2ui.org)
*   **GitHub Repository:** [google/A2UI](https://github.com/google/A2UI)

## Overview
A2UI (Agent-to-User Interface) is a protocol that enables AI agents to generate rich, interactive user interfaces securely and declaratively. Instead of generating executable code (which poses security risks), the agent sends a JSON description of the UI, which the client renders using a "Trusted Catalog" of native components.

## Core Concepts

### 1. Server-Driven UI (SDUI)
The Agent is the "Server" in this context. It dictates *what* to show, but not *how* to draw it pixel-by-pixel. The Client handles the rendering implementation (Web, Mobile, etc.).

### 2. The Trusted Catalog
The Client Application maintains a list of allowed components (e.g., `Text`, `Button`, `Input`, `Column`). The Agent can only request components from this catalog. If the Agent requests a component that doesn't exist, the Client falls back gracefully (or ignores it).

### 3. Declarative JSON Structure
The UI is described as a tree of nodes.
```json
{
  "type": "column",
  "children": [
    { "type": "text", "text": "Hello, World!" },
    { "type": "button", "label": "Click Me" }
  ]
}
```

### 4. Adjacency List Model (Streaming)
For complex UIs, A2UI supports a flat list interaction model where components are defined by ID and parent reference. This is optimized for LLM streaming, allowing the UI to "paint" as the agent "thinks".
*(Note: FutureBuild currently uses a recursive tree model for simplicity, but may adopt streaming adjacency lists in Phase 8).*

## FutureBuild Implementation
We correspond to the A2UI philosophy but use a simplified schema customized for our domain.
See `specs/API_AND_TYPES_SPEC.md` Section 4.3 for our specific `DynamicComponent` schema.

**Authorized Component Catalog:**
*   Authorized components are defined in `specs/FRONTEND_SCOPE.md` Section 4.4.
*   **Do not hallucinate new components.** Stick to the catalog.

## Security
*   **No Code Execution:** Never attempt to send JavaScript or HTML strings.
*   **State Separation:** The Agent does not manage the state of the UI (e.g., "is the dropdown open?"). It only cares about the *Submission* (the data the user sends back).
