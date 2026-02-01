- Phase: Phase 12: Identity & Sovereignty (Auth Refactor)

## Phase 12: Identity & Sovereignty (Auth Refactor)
**PRD Reference:** [PHASE_12_PRD.md](../planning/PHASE_12_PRD.md)

### 🛡️ Guardrails & Alignment
Before executing any step, the Engineer MUST verify alignment with:
1.  **Product Vision:** Does this change maintain the "Construction Professional" aesthetic and high-trust environment?
2.  **Master PRD:** Does this adhere to the Multi-Tenant Schema (Section 4.1) and Role Definitions?
3.  **L7 Spec:** Does the implementation match the granular requirement in the linked spec file?

**CRITICAL:** Perform an **L7 Self-Reflective Test** before marking any step as complete. Ask: "If I were an antagonistic auditor, would I find a security hole or UX regression here?"

---

- [ ] Step 78: Auth Provider Integration @Frontend
  - **Task:** Integrate Clerk/Auth0 to replace magic link system.
  - **Spec:** [STEP_78_AUTH_PROVIDER.md](../specs/phase12/STEP_78_AUTH_PROVIDER.md)
  - **Core Requirement:** Zero custom auth code in frontend; use Provider SDK. Validate strictly against the "Construction Professional" dark mode aesthetic.

- [ ] Step 79: Middleware Swap @Backend
  - **Task:** Update Go middleware to validate JWKS from Provider instead of DB tokens.
  - **Spec:** [STEP_79_MIDDLEWARE_SWAP.md](../specs/phase12/STEP_79_MIDDLEWARE_SWAP.md)
  - **Core Requirement:** Stateless verification. Hard cutover (legacy tokens invalid). Security Audit: Ensure no algorithmic confusion attacks possible on JWT.

- [ ] Step 80: Organization Manager @Frontend
  - **Task:** Implement "Team Settings" to invite/manage members.
  - **Spec:** [STEP_80_ORG_MANAGER.md](../specs/phase12/STEP_80_ORG_MANAGER.md)
  - **Core Requirement:** Org-switching in UI must trigger immediate data store reset to prevent data leaks between tenants.

- [ ] Step 81: Role Mapping @Backend
  - **Task:** Map Provider roles (Admin/Member) to internal `PermissionMatrix`.
  - **Spec:** [STEP_81_ROLE_MAPPING.md](../specs/phase12/STEP_81_ROLE_MAPPING.md)
  - **Core Requirement:** Strict enforcement of "Viewer" vs "Builder" roles at the API level (RBAC).

---
## 🧠 Memory Logs
- **Product Orchestrator:** Phase 12 Identity plan initialized.
- **L7 Gatekeeper:** Phase 11 Audit Passed. Phase 12 Specs generated with strict security focus.
- **Reference:** PRD available at [PHASE_12_PRD.md](../planning/PHASE_12_PRD.md).
