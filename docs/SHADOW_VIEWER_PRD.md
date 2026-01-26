# SHADOW_VIEWER PRD

**Status**: Draft
**Owner**: Product Orchestrator
**Reviewers**: Principal Engineer, UX Engineer, Security Engineer

## 1. Executive Summary
The "Shadow Viewer" is a specialized, high-privilege interface within the FutureBuild Command Center. It serves as the "System 2" dashboard, providing transparency into the "Tribunal's" decision-making process (Multi-Model Consensus logs) and hosting "ShadowDocs" (internal system documentation). It transforms the AI from a black box into an auditable, understandable teammate.

## 2. Problem Statement & Goals
**Problem**:
- **Opaque Logic**: When the "Tribunal" rejects an agent's code or action, developers have no UI to see *why* (votes, reasoning, confidence scores).
- **Hidden Knowledge**: System documentation exists in the repo (`docs/`, `specs/`) but is not easily accessible *within* the application flow.

**Goals**:
- **Radical Transparency**: Expose the raw "thought process" of the AI consensus engine.
- **Unified Knowledge Base**: Render markdown documentation directly in the app for seamless reference.
- **Auditability**: Provide a permanent, searchable log of every high-stakes decision made by the system.

## 3. User Stories
1. **As a Developer**, I want to browse a list of recent Tribunal decisions so I can spot patterns in false negatives/positives.
2. **As a Developer**, I want to click into a specific Decision to see what "Claude" voted vs. what "Gemini" voted, so I can debug model disagreement.
3. **As a Product Owner**, I want to read the system specs (ShadowDocs) inside the app without switching to GitHub.
4. **As a Security Engineer**, I want to ensure only authorized users can access the Shadow Viewer to protect sensitive internal logs.

## 4. Functional Requirements

### 4.1 Tribunal Log Feed
- **List View**: Display a chronological list of `TribunalDecision` records.
- **Columns**:
    - Status (Approved/Rejected/Conflict) - Visual badge (Green/Red/Amber).
    - Case ID (Short hash).
    - Context (e.g., "Refactor utils.go").
    - Models Consulted (Icons for participating models).
    - Timestamp.
- **Filtering**: By Decision Status, Time Range, or Specific Model involvement.

### 4.2 Case Detail View
- **Drill-down**: Clicking a row in the feed opens the Case Detail.
- **Consensus Header**: Big visual indicator of the final outcome and confidence score.
- **Juror Breakdown**: Card for each model's `ModelVerdict`.
    - Model Name.
    - Vote (Yea/Nay).
    - Reasoning (Text explanation).
    - Latency/Cost.
- **Diff View** (Future Scope): If the decision involved code, show the proposed diff.

### 4.3 ShadowDocs Viewer
- **Navigation**: Sidebar listing all markdown files in `docs/` and `specs/`.
- **Renderer**: High-fidelity Markdown rendering (GitHub style).
- **Deep Linking**: Ability to link from a Tribunal Decision to a specific policy doc (e.g., "Rejected per [Security Policy v2]").

## 5. UX/UI Flows

### 5.1 Entry Point
- **Shadow Toggle**: A dedicated icon/switch in the main `fb-panel-left` (Agents/System section).
- **Access**: Clicking it switches the `fb-panel-center` from "Chat" mode to "Shadow" mode.

### 5.2 Layout (Shadow Mode)
- **Left Panel**: Navigation (Log Feed | ShadowDocs File Tree).
- **Center Panel**: The Content (The Log Table or The Doc).
- **Right Panel**: Context/Metadata (Detail view for the selected log item).

## 6. Technical Considerations

### 6.1 Data Source
- **Logs**: Fetched from `TribunalService` (likely a specialized `ListDecisions` endpoint).
- **Docs**: Fetched via a "Content Service" that reads backend markdown files (or served statically).

### 6.2 Security
- **RBAC**: Strictly limited to `ADMIN` or `DEVELOPER` roles. Hiding the UI is not enough; API endpoints must enforce checks.

## 7. Success Metrics
- **Debug Velocity**: Time to diagnose a "why did it do that?" question (Target: < 1 min).
- **Documentation Usage**: Frequency of ShadowDocs access vs. external repo browsing.
