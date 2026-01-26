---
description: Invoke the Dev Team to execute engineering work, build features, and write code.
---

1. **Validate Input**:
   - Verify `[TASKNAME]` is provided in the invocation.
   - Verify `docs/[TASKNAME]_PRD.md` exists.
   - If validation fails: "Cannot proceed. Please complete `/product` first."
2. **Invoke Skill**: Use the `view_file` tool to read `skills/devteam/SKILL.md`.
3. **Execute**: Assume the role of the **DevTeam Orchestrator** and follow the instructions in the SKILL file.
   - **You must** create an `implementation_plan.md` and `task.md` that explicitly lists which sub-skills (Architect, Backend, Security) will be used for architectural components and API design.
   - If no PRD exists for a large feature, **REJECT** and tell user to use `/product`.
   - If valid, create `specs/[TASKNAME]_specs.md` and start the sprint.
4. **Handoff**: Provide inter-thread transition instruction:
   > "Specs complete in `specs/[TASKNAME]_specs.md`. Please review the technical design."
   > "When ready, invoke `/plan_review [TASKNAME]` to generate the implementation context."
