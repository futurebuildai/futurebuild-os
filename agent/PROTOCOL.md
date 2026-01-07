# Prism Transition Protocol

This document formalizes the session-to-session transition for Specification-Driven Development.

## The Atomic Step Loop
1. **Initiate**: User pastes the `SYSTEM_PROMPT.md` (which includes the current Task).
2. **Execute**: Agent completes exactly **one step** from `PRODUCTION_PLAN.md`.
3. **Audit**: User triggers `/CTO`. Agent performs Triple Review.
4. **Finalize**: User triggers `/NEXT`.
    - Agent updates `agent/ROADMAP.md` (Check off step).
    - Agent updates `agent/HANDOFF.md` (Session summary).
    - Agent updates `agent/SYSTEM_PROMPT.md` (Update "Task Prompt" to next step).
    - Agent notifies user with the **FULL PROMPT** for the next context.

## State Maintenance
- **Immutable Constraints**: Never modified by the loop.
- **Dynamic State**: `ROADMAP.md`, `HANDOFF.md`, and the "Task" section in `SYSTEM_PROMPT.md`.
