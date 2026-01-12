# Conversational Agent Specification (Layer 4)

**Version:** 1.0  
**Date:** 2026-01-12  
**Status:** Planning

---

## 1. Overview

FutureBuild v1 uses a **Chat-First Web Application** architecture. The AI Agent Layer (Layer 4) serves as the primary "Operating System" for the construction process, filtering the complexity of the Physics Engine (Layer 3) into simple, actionable, natural-language priorities.

### 1.1 Core Philosophy: Agent-as-Sieve
The Physics Engine generates 1,000 data points (floats, dates, constraints). The Agent filters this down to the **ONE** thing that requires human attention.

**The Sacred Workflow:**
1.  **Login** → 
2.  **Morning Briefing** (Top Priority Card) → 
3.  **Execution Loop** (User resolves priority via chat) → 
4.  **Task Completion**

---

## 2. Agent Response Patterns

### 2.1 Morning Briefing (The "Lobby")
**When:** First interaction of the day or after login.
**Goal:** Direct user focus to the single most critical risk.

**Pattern:**
- **Greeting**: "Good morning, [Name]."
- **Priority Card**: One single `<fb-priority-card>`.
- **Reasoning**: Plain English explanation of *why* this matters (critical path impact).
- **Suggested Actions**: 2-3 buttons for immediate resolution.

**Example Artifact:**
```json
{
  "card_type": "priority",
  "title": "Concrete Pour at Risk",
  "risk_impact": "+3 days to Critical Path",
  "reason": "Moisture levels at 6.2% (Target: <5%)",
  "actions": [
    { "label": "Reschedule Pour", "intent": "RESCHEDULE_TASK" },
    { "label": "Call Inspector", "intent": "CONTACT_INSPECTOR" }
  ]
}
```

### 2.2 Execution Loop (The "Action")
**When:** User responds to a briefing or initiates a command.
**Goal:** Execute complex backend logic via natural language.

**Flow:**
1.  **User**: "Reschedule the pour for Tuesday."
2.  **System**: Parses `UPDATE_TASK_SCHEDULE` intent.
3.  **Agent**: Updates Physics Engine (`ScheduleService.UpdateTask`).
4.  **System**: Recalculates Critical Path.
5.  **Agent**: Responds with confirmation + Impact Analysis.

**Response:**
"Done. Foundation Pour moved to Tuesday, Jan 15th. This pushes Framing start to Thursday (+2 days). I've notified the concrete sub."

### 2.3 Visual Fallback (Ephemeral UI)
**When:** Text is insufficient for data density.
**Rule:** Never replace the chat. Embed the data AS A CARD in the stream.

**Triggers:**
- **Comparisons**: "Show me the schedule change." → `<fb-artifact-gantt>` (diff view).
- **Financials**: "What's the budget impact?" → `<fb-artifact-budget>` (row view).
- **Lists**: "Who is on site?" → `<fb-contact-card>` (list view).

### 2.4 Dynamic Micro-Flows (A2UI)
**When:** The Agent needs to gather structured input that doesn't warrant a full bespoke Artifact (e.g., surveys, quick choices, simple logs).
**Mechanism:** The Agent constructs an ephemeral UI using `ArtifactType: Dynamic_UI`.

**Example: Subcontractor Selection**
User: "Find an electrician for Phase 2."
Agent Response:
```json
{
  "title": "Select Electrician",
  "root": {
    "type": "box",
    "props": { "direction": "col", "gap": "sm" },
    "children": [
      { "type": "text", "props": { "variant": "h2", "text": "Recommended Electricians" } },
      { "type": "select", "props": { 
          "label": "Choose Vendor", 
          "options": ["Sparky Co ($50/hr)", "Volt Masters ($55/hr)"] 
        } 
      },
      { "type": "button", "props": { "label": "Assign", "action_id": "assign_sub" } }
    ]
  }
}
```

---

## 3. Intent Classification

The `ChatOrchestrator` uses Vertex AI (Gemini) to classify user messages into **Intents**.

| Intent ID | Trigger Examples | Service / Action |
|-----------|------------------|------------------|
| `PROCESS_INVOICE` | User uploads PDF, "Here is an invoice" | `InvoiceService.Process` |
| `QUERY_BUDGET` | "Are we over budget?", "Show me costs" | `BudgetService.GetSummary` → `<fb-artifact-budget>` |
| `GET_SCHEDULE` | "When is framing?", "Show timeline" | `ScheduleService.GetGantt` → `<fb-artifact-gantt>` |
| `UPDATE_TASK_STATUS` | "Framing is done", "Mark #203 complete" | `ScheduleService.UpdateProgress` |
| `REPORT_DELAY` | "Lumber is late", "Rain delay today" | `ScheduleService.UpdateStart` + `NotificationService` |
| `MANAGE_CONTACTS` | "Who is the electrician?", "Call Bob" | `DirectoryService.GetContact` → `<fb-contact-card>` |
| `EXPLAIN_DELAY` | "Why is this late?", "What is blocking?" | `Physics.GetCriticalPath` → `<fb-card>` |

---

## 4. API Specification

### 4.1 Message Endpoint
**POST** `/api/v1/agent/message`

**Request:**
```json
{
  "project_id": "uuid",
  "message": "Reschedule framing to next Monday",
  "context": {
    "user_role": "superintendent",
    "client_time": "2026-01-12T09:00:00Z"
  }
}
```

**Response (Streaming or JSON):**
```json
{
  "reply": "I've handled that.",
  "intents_triggered": ["UPDATE_TASK_SCHEDULE"],
  "artifacts": [
    {
      "type": "gantt_diff",
      "data": { ... }
    }
  ],
  "suggested_actions": ["Notify Crew"]
}
```

### 4.2 Real-Time Updates (WebSocket)
**URL:** `wss://api.futurebuild.app/ws/agent?project_id=...`

**Events:**
- `agent.typing`: System is processing.
- `agent.message`: Text chunk (streaming).
- `agent.artifact`: Renders a component.
- `agent.notification`: System alert (non-blocking).

---

## 5. Constraint Integration

### 5.1 Physics vs. Agent Layer
- **Layer 3 (Physics)**: Deterministic. Calculates dates, float, costs. NEVER hallucinates.
- **Layer 4 (Agent)**: Probabilistic. Interprets intent, synthesizes explanations.
- **Rule**: The Agent **reads** from Physics results. The Agent **writes** to Service methods. The Agent **never** performs schedule math itself.

### 5.2 Speech-to-Text (STT)
**Requirement**: The frontend `<fb-input-bar>` must include a microphone button.
- **Implementation**: Browser Native Speech API or specialized STT service.
- **Flow**: Audio → Text → Chat Input → Agent. (Voice output is optional).
