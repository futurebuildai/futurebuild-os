# Product Requirements Document: Shadow Site & Protocol (step63_PRD)

| Metadata | Details |
| :--- | :--- |
| **Feature ID** | STAT-63 (Shadow Site) |
| **Status** | **APPROVED** (Target: Phase 8 - Production Readiness) |
| **Owner** | Product Orchestrator |
| **Authors** | Antigravity, User |
| **Last Updated** | 2026-01-25 |
| **Dependents** | FutureShade (The Tribunal), Compliance Team |

---

## 1. Executive Summary
The **Shadow Site** is a parallel documentation architecture that mirrors the production codebase file-for-file. Unlike traditional documentation (which rots in a wiki) or code comments (which are too granular), the Shadow Site provides a "Zoomed Out" natural language explanation of the system's intent, responsibility, and dependencies.

This PRD establishes the **Shadow Protocol**, a strict "Dual-Write" enforcement mechanism that requires every meaningful code artifact to have a corresponding "Shadow" explanation. This lays the foundation for **FutureShade** (Step 64+), where AI agents will use this shadow layer to reason about, govern, and evolve the system without needing to ingest raw code.

---

## 2. Problem Statement
### 2.1 The Code-Context Gap
As the FutureBuild codebase scales (currently 5 Domains, 20+ Core Services), knowledge becomes siloed.
*   **The Problem:** "To understand how the *Invoice Hydration* works, you must read 400 lines of `InvoiceService.ts`."
*   **The Consequence:**
    1.  **AI Blindness:** Agents consume massive context windows reading implementation details when they only need architectural understanding.
    2.  **Onboarding Friction:** New engineers (and non-engineers) cannot verify business logic without parsing syntax.
    3.  **Drift:** Business requirements ("Invoices must verify WBS codes") verify against the code only during initial implementation, then vanish into implicit knowledge.

### 2.2 The "Rotting Docs" Fallacy
External documentation (Notion, Confluence) inevitably drifts from the code. By the time a feature ships, the wiki is obsolete.
*   **The Solution:** Documentation must live *adjacent* to the code, in the repository, enforced by the same CI pipeline that enforces compilation.

---

## 3. Goals & Success Metrics

### 3.1 Primary Goals
1.  **Bit-Level Parity**: Establish a directory structure `shadow/` that mirrors `src/` 1:1 for all structural components.
2.  **Automated Governance**: Implement a CI barrier (The "Shadow Check") that fails the build if code exists without its shadow.
3.  **AI Readiness**: Create a noise-free "Context Layer" for FutureShade agents.

### 3.2 Success Metrics
| Metric | Target | Measurement Strategy |
| :--- | :--- | :--- |
| **Shadow Coverage** | 100% | `scripts/shadow/check.ts` finds 0 missing files. |
| **Scaffold Speed** | < 2s | Scaffolding script runtime on full repo. |
| **CI Latency** | < 5s | Impact on total build time. |
| **Drift Detection** | N/A | (Future Scope: Step 68 Tribunal). |

---

## 4. Functional Requirements

### 4.1 The Shadow Protocol (Dual-Write)
The core requirement is a simplified implementation of a "Mirror World".

**Rule 1: structural Equivalence**
For every file $F$ in `src/` that meets the **inclusion criteria**, there MUST exist a file $S$ in `shadow/` with the same relative path, replacing the extension with `.md`.

*   **Source:** `frontend/src/components/billing/InvoiceCard.ts`
*   **Shadow:** `shadow/components/billing/InvoiceCard.md`

**Rule 2: The Inclusion Criteria**
The Shadow Protocol applies to:
*   Logic Files: `.ts`, `.js`, `.go`, `.py`
*   UI Components: `.tsx`, `.vue` (if applicable), `.ts` (Lit elements)

It explicitly **EXCLUDES**:
*   Tests: `.test.ts`, `.spec.ts`, `_test.go`
*   Styles: `.css`, `.scss`
*   Assets: `.png`, `.svg`
*   Fixtures/Mocks: `fixtures/`, `mocks/` patterns (unless they contain critical business logic logic).
*   Index Barrels: `index.ts` (unless they contain logic).

### 4.2 Content Standards (The L7 Template)
Every Shadow File must adhere to the following schema to ensure consistency.

```markdown
# [Component Name]

## Intent
*   **High Level:** (1-2 sentences). Why does this exist? "Displays the invoice line items."
*   **Business Value:** "Allows users to verify costs."

## Responsibility
*   What is this component allowed to do? (e.g., "Fetches data", "Renders UI")
*   What is it strictly FORBIDDEN from doing? (e.g., "Mutating DB state").

## Key Logic
*   **Flows:** Natural language description of complex methods.
*   **State:** What local state does it manage?

## Dependencies
*   **Upstream:** What calls this?
*   **Downstream:** What does this call? (e.g., `InvoiceService`).
```

### 4.3 Tooling: Scaffolding (The Catalyst)
To prevent developer revolt, we must provide a tool to generate these files automatically.
*   **Command:** `npm run shadow:scaffold`
*   **Behavior:**
    1.  Walks the `src/` tree.
    2.  Identifies missing shadow files.
    3.  Creates the file with the template above.
    4.  Pre-fills "Intent" with "Pending documentation." to allow the build to pass (initially), flagging it for review.

### 4.4 Tooling: Enforcement (The Gatekeeper)
*   **Command:** `npm run shadow:check`
*   **Behavior:**
    1.  Walks the `src/` tree.
    2.  Checks for existence of shadow equivalent.
    3.  **EXIT 1** if missing.
    4.  **EXIT 0** if all present.
*   **CI Integration:** Must run on every Pull Request.

---

## 5. Non-Functional Requirements

### 5.1 Performance
The `check` script must be extremely fast (Node.js stream or Go implementation). It should not add perceptible time to the local commit hook or CI pipeline.

### 5.2 Security & Privacy
*   Shadow files are conceptually "Public Safe". They describe *how* the system works, not *what* data it holds.
*   **Constraint:** Shadow files MUST NOT contain PII, Secrets, or hardcoded credentials (inheriting the same rules as Key Code, but stricter).

### 5.3 Scalability
The structure must support multiple roots (monorepo support):
*   `frontend/shadow/`
*   `backend/shadow/`

---

## 6. Pre-Mortem (Risk Analysis)

**Risk:** "The Blank File Problem."
*   *Scenario:* Developers run `scaffold`, commit the "Pending documentation" stubs, and never write the actual content. The coverage metric reads 100%, but utility is 0%.
*   *Mitigation (Step 63):* Acceptable risk for plumbing.
*   *Mitigation (Step 68 - Tribunal):* The "Tribunal" AI will eventually audit shadow files and reject PRs where the shadow file is a stub but the code is complex.

**Risk:** "Refactor Rot."
*   *Scenario:* A developer moves `Invoice.ts` to `billing/Invoice.ts`.
*   *Effect:* The check script fails (missing shadow in new location).
*   *Friction:* Developer must manually move the shadow file.
*   *Mitigation:* Update `scaffold` tool to handle moves or provide a `shadow:mv` helper? *Decision:* Keep it simple. Fail the build. Developer moves the file manually.

---

## 7. Migration & Rollout Plan

1.  **Phase 1: Plumbing (Step 63)**
    *   Create directory structure.
    *   Write `scaffold.ts` and `check.ts`.
    *   Add content templates.
2.  **Phase 2: Initial Hydration**
    *   Run `scaffold` on the entire repo.
    *   Commit huge batch of stub files.
3.  **Phase 3: Enforcement**
    *   Enable CI check.
    *   From this point forward, no new code can enter without a shadow.
4.  **Phase 4: Backfill (Lazy)**
    *   As developers touch files, they should update the shadow docs (Boy Scout Rule).

---
