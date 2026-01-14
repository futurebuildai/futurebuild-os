### Architecture & Design Decisions
This log tracks significant decisions to prevent re-litigating settled issues.
**Governing Constraints:** See `BACKEND_SCOPE.md` and `VISION.md` for hard constraints.

#### [DEC-001]: [DECISION_TITLE]
**Date:** [YYYY-MM-DD]
**Status:** [Proposed/Decided/Deprecated]

##### Context
[Describe the problem. Example: "We need to choose a database for the user profiles."]

##### Decision
[The specific choice made. Example: "Use PostgreSQL with JSONB columns."]

##### Rationale
*   **Constraint:** (e.g., "Aligned with `BACKEND_SCOPE.md` Section 1.1 which mandates a relational core.")
*   **Trade-off:** [Describe what we lost by making this choice vs. the alternative]
*   **Impact:** [How this affects future development]

--------------------------------------------------------------------------------

#### [DEC-002]: Intent Priority Logic (Safety over Specificity)
**Date:** 2026-01-13
**Status:** Decided

##### Context
In the V1 Chat Orchestrator (Step 43.2), the command "Update the schedule" is ambiguous. It contains both "update" (Action) and "schedule" (Noun).

##### Decision
We prioritize **Nouns/Get** operations over **Verbs/Update** operations when both are present.
*   "Update the schedule" -> `IntentGetSchedule`
*   "Schedule slip" -> `IntentExplainDelay` (Specific Crisis Match)

##### Rationale
*   **Safety:** In a high-stakes construction environment, defaulting to a Read-Only view (`GET_SCHEDULE`) is safer than triggering a state change flow (`UPDATE_TASK`) when the user's intent is ambiguous.
*   **Trade-off:** Users may need to be more explicit to trigger updates (e.g., "Update status"), but they won't accidentally mutate state when just asking for information.
*   **Impact:** Documented behavior for future NLP model tuning.

--------------------------------------------------------------------------------

#### Open Questions / Pending Decisions
*   **[Q1]:** [Question?] (Ref: `FRONTEND_SCOPE.md` Section 2)
