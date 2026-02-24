---
description: Template prompt for a deep Epic-level gate audit after all sprints in an Epic are complete.
---

# Epic Gate Audit — Starter Prompt Template

Copy the prompt below into a **new Antigravity thread**, replacing all `{{PLACEHOLDER}}` values with your Epic-specific details.

---

```
## Deep Review: EPIC {{EPIC_NUMBER}} — {{EPIC_NAME}}

**Master Roadmap:** `planning/ROADMAP_SPEC.md` — our north star: *"Achieve 'Game Changer' status for live beta testers and pass the Google Production Readiness Audit."*

**Sprint Specs Completed in This Epic:**
{{LIST_EACH_SPRINT_SPEC_FILE, e.g.:}}
- `planning/sprints/sprint-X-Y-name.md`
- `planning/sprints/sprint-X-Z-name.md`

Read all sprint specs first to understand the full scope of work delivered.

### Your Mission
Perform an end-to-end audit of EPIC {{EPIC_NUMBER}} as a whole. You are the quality gate before we mark this Epic complete and move resources to the next Epic. Think like the Google Production Readiness auditor.

### Review Checklist

**1. Epic Objective — Did we achieve it?**
- The stated objective: *"{{EPIC_OBJECTIVE_FROM_ROADMAP}}"*
- Verify this works end-to-end by testing in the browser (`npm run dev` from `frontend/`) and/or backend (`go run` / `go test` from `backend/`)
- Test the full user journey relevant to this Epic
- Does this feel like a **"Game Changer"** or just a feature? Be honest.

**2. Cross-Sprint Integration — Do the sprints compose correctly?**
- Verify there are no seams between sprints — do the pieces work together seamlessly?
- Verify no duplicated logic — is each concern handled in ONE place and consumed everywhere?
- Check for orphaned code — did later sprints fully replace patterns introduced in earlier sprints?

**3. Code Quality Audit (Production Readiness)**
- Review every file modified or created across all sprints in this Epic
- Check: type safety, error handling, edge cases, memory leaks, race conditions
- Check: accessibility (aria labels, keyboard navigation, screen reader support)
- Check: performance (unnecessary re-renders, expensive computations, bundle size impact)

**4. Success Criteria Verification**
- Open `planning/ROADMAP_SPEC.md` and find the Success Criteria relevant to this Epic
- For each criterion: Is it fully met? Partially? What's missing?

**5. Backward Compatibility & Build Health**
- Frontend: `npm run build` from `frontend/` — clean build, zero errors?
- Backend: `go build ./...` from `backend/` — clean compilation?
- Are there any components or services NOT yet migrated that should be? (List as tech debt)

**6. Documentation Update**
- Verify all sprint spec files in this Epic have accurate task statuses
- Update `planning/ROADMAP_SPEC.md`:
  - If EPIC passes: Set all sprint statuses to "✅ Complete"
  - If issues found: Set to "🟠 Needs Fix" with notes
  - Check/uncheck relevant Success Criteria checkboxes
  - Add tech debt notes if applicable

### Rules
1. Be thorough and adversarial — you are the Google auditor
2. Categorize findings as **Blocker** (must fix before next Epic), **Warning** (fix within next sprint cycle), or **Tech Debt** (track for later)
3. Do NOT fix issues yourself — document them only
4. End with a clear **EPIC GO / EPIC NO-GO** verdict with justification
```
