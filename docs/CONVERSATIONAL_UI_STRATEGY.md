# FutureBuild: Conversational UI Strategy

## Executive Summary
FutureBuild v1 will use a **Chat-First Desktop Interface** that positions the AI Agent Layer as the primary "Operating System" for construction project management. Instead of dashboards and forms, the user interacts with an intelligent concierge that translates complex physics calculations into actionable natural-language priorities.

---

## 1. The Core User & The Core Workflow

### Primary User
**PM/Superintendent Hybrid**
*   **Small Builders:** One person owns both on-site execution and project management.
*   **Larger Builders:** Dedicated PM (office/field) and Superintendent (on-site) use the app similarly.
*   **Profile:** Often distracted, on-site, time-constrained. Needs immediate clarity, zero cognitive load.

### The Sacred Workflow (The "Daily Heartbeat")
1.  **Login**
2.  **Agent generates top priority for the day** (synthesized from Physics Engine + stakeholder state)
3.  **User completes associated tasks** via conversational interaction (approve, reject, delegate, reschedule)
4.  **Goal:** Keep schedule on time, budget intact, stakeholders satisfied.

---

## 2. Architectural Philosophy: Agent-as-Sieve

### Traditional UI Problem
*   **User is the Filter:** They scan dashboards, Gantt charts, reports to find what matters.
*   **Cognitive Overload:** 1,000 data points presented at once.
*   **Learning Curve:** Requires training to interpret color codes, statuses, hierarchies.

### FutureBuild Solution
*   **Agent is the Filter:** Physics Engine processes 1,000 calculations. Agent extracts the ONE thing that needs human intervention.
*   **Zero Learning Curve:** User receives plain-English instructions.
*   **Concierge Model:** "Here is your briefing. Here are your options. Click to execute."

---

## 3. Conversational Interaction Patterns

### 3.1 Morning Briefing (The "Lobby Entrance")
**When:** User logs in or opens app.

**Agent Output Example:**
```
Good morning, Alex. Here's your focus for today:

🎯 TOP PRIORITY
Foundation inspection is scheduled for 10 AM, but the concrete pour still shows moisture levels at 6.2% (target: <5%). 

If we don't pass, we lose 3 days on the critical path and delay framing.

ACTIONS:
1. Check moisture again at 9 AM
2. Notify inspector if we need to reschedule
3. [View alternative schedule if delayed]

Would you like me to set a reminder for 8:45 AM?
```

**Design Notes:**
*   **Single Priority Card:** Not a list of 50 tasks.
*   **Context + Consequence:** User understands the "why."
*   **Actionable Buttons:** Pre-generated next steps.

### 3.2 Execution Loop (The "Action Conversation")
**Pattern:** User makes a statement or asks a question. Agent interprets intent and updates the database or requests clarification.

**Example Exchange:**
```
USER: "Lumber truck is stuck in traffic. ETA now 2 PM."

AGENT: "Got it. I've logged the delay. This pushes framing start from 1 PM to 2:30 PM. 
Should I notify the framing crew, or do you want to call them directly?"

USER: "You notify them."

AGENT: "Done. Crew notified via SMS. Updated schedule attached."
[Show mini-timeline snippet]
```

**Backend Logic:**
1.  Agent parses "lumber delay" intent.
2.  Calls `ScheduleService.UpdateTaskStart(...)`.
3.  Physics Engine recalculates downstream impacts.
4.  Agent generates notification payload for crew.
5.  Confirmation message sent to user.

### 3.3 Visual Fallback (Ephemeral UI)
**Philosophy:** Chat is the primary interface, but when precision is needed, inject a micro-visualization.

**When to Use:**
*   **Before/After Comparisons:** "Here's the old schedule vs. the new one."
*   **Financial Impact:** "This delay costs $1,200 in crew idle time."
*   **Approval Gates:** "Here's the updated budget. Approve to proceed."

**Format:**
*   Inline embeds in the chat (like Slack unfurls).
*   Collapsible cards.
*   Never full-screen takeovers.

---

## 4. Technical Architecture

### 4.1 Frontend Stack (Lit 3.0 + TypeScript)
**Component Hierarchy:**
```
<app-shell>
  <chat-interface>
    <message-list>
      <user-message />
      <agent-message />
      <ephemeral-card /> <!-- Schedule snippet, approval form, etc. -->
    </message-list>
    <input-bar>
      <text-input />
      <voice-button /> <!-- Future: v1 optional -->
    </input-bar>
  </chat-interface>
</app-shell>
```

### 4.2 Backend Integration
**Flow:**
1.  User sends message via WebSocket or SSE.
2.  Backend parses via Gemini/Vertex AI (Intent Classification).
3.  Agent calls appropriate service (`ScheduleService`, `DirectoryService`, etc.).
4.  Database updated.
5.  Response streamed back to client.

**Key API Endpoint:**
```
POST /api/v1/agent/message
{
  "message": "Lumber is delayed 2 hours",
  "context": {
    "project_id": "...",
    "user_role": "superintendent"
  }
}

Response:
{
  "reply": "Got it. I've updated...",
  "actions_taken": ["schedule_updated", "crew_notified"],
  "ephemeral_ui": { ... } // Optional visual card
}
```

---

## 5. Design Principles

### 5.1 Premium but Rugged
*   **Premium:** Clean typography (Inter/Roboto), smooth animations, polished micro-interactions.
*   **Rugged:** High contrast, large touch targets, works with construction gloves, readable in direct sunlight.

### 5.2 Conversational Guardrails
*   **No Dead Ends:** Agent always provides clear next steps.
*   **No Multi-Turn Traps:** If a decision tree gets >3 turns, offer a "Jump to Summary" option.
*   **Undo Safety:** "Oops, I didn't mean that" → "No problem, I've reverted the change."

### 5.3 Stakeholder Happiness (Implicit Metric)
*   We don't necessarily need a `stakeholder_sentiment` column (yet).
*   "Happiness" is measured by:
    *   Schedule adherence
    *   Budget variance
    *   Communication frequency (e.g., no "surprise" delays)

---

## 6. Open Questions for Next Phase

1.  **Voice Integration for Desktop v1:** Optional or mandatory? (User indicated optional for v1).
2.  **Offline Mode:** Does the PM need limited functionality if internet drops on-site?
3.  **Multi-Project Switching:** How does the Agent handle a PM managing 3 concurrent projects?

---

## 7. Next Steps

### Phase 1: Spec Alignment
*   Update `FRONTEND_SCOPE.md` to reflect Chat-First architecture.
*   Update `PRODUCTION_PLAN.md` to include conversational UI implementation steps.
*   Create `CONVERSATIONAL_AGENT_SPEC.md` to define Agent response patterns.

### Phase 2: Prototype
*   Build minimal chat interface with mock Agent responses.
*   Test with real PM/Superintendent users for flow validation.

### Phase 3: Full Integration
*   Wire Agent to Physics Engine and Database.
*   Implement ephemeral UI cards.
*   Deploy v1 Web App.

---

**Document Status:** Draft v1.0  
**Last Updated:** 2026-01-12  
**Consultation Mode:** /brain
