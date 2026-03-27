# Agent Instruction: Generate Sprint Plan

**Context:** You are the Lead Architect for FutureBuild. We are executing the `ROADMAP.md` to reach Public Beta.
**Current State:** The codebase is robust but requires specific feature implementations as outlined in the Roadmap.

**Your Mission:** Generate a comprehensive, step-by-step `PLAN.md` file for **[INSERT PHASE NUMBER AND NAME HERE]** (e.g., "Phase 11: The Conversational Hook").

---

### 🛑 Strict Rules for PLAN.md Generation

**1. Step 0.1: The "Spec Check" (Mandatory)**
* The FIRST task in the plan must be to **Generate or Update a Specification File** (e.g., `specs/committed/PHASE_11_SPEC.md`).
* This spec must explicitly define:
    * **User Stories:** "As a user, I want to..."
    * **Data Contracts:** Exact JSON request/response shapes.
    * **Component Interface:** Props, Events, and State for any UI components.
* *Why:* We do not write code without a blueprint.

**2. Step 0.2: The "Code Audit" (Mandatory)**
* The SECOND task must be to **Read and Audit** the existing relevant files.
* Explicitly list the files you need to read (e.g., "Read `frontend/src/components/views/fb-view-projects.ts` and `backend/internal/api/handlers/project_handler.go`").
* *Why:* Verify assumptions before modifying the codebase.

**3. Granularity & Separation of Concerns**
* Break down tasks by layer (Frontend vs Backend). **DO NOT** combine them.
* Example of GOOD structure:
    * `[ ] 1.1: [Backend] Define struct in 'internal/models'`
    * `[ ] 1.2: [Backend] Implement handler logic`
    * `[ ] 1.3: [Frontend] Create API client method`
    * `[ ] 1.4: [Frontend] Build UI Component`

**4. Verification Steps**
* Every major logical block must end with a **Verification Step**.
* Examples: "Run `go test ./...`", "Verify endpoint via `curl`", "Check browser console for errors".

---

### Output Format
Please output the full content of the `PLAN.md` file inside a code block. Do not execute the plan yet; just generate the plan file.
