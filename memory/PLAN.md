# FutureBuild Phase 13: The Action Loop (Invoice & Field)

**PRD Reference:** [PHASE_13_PRD.md](../planning/PHASE_13_PRD.md)
**Objective:** Transform static artifacts into interactive decision points, enabling the "Action Loop" where users validate and approve AI-generated data.

---

## 🛡️ CORE GUARDRAILS & PRODUCT ALIGNMENT
**Alignment Check:** Every step below must be executed with strict adherence to the **FutureBuild Product Vision**:
1.  **"AI as a Partner, not a Tool":** The AI proposes (Draft), the User disposes (Approve/Reject). The UI must explicitly model this relationship.
2.  **"Zero Latency Perception":** Feedback loops (like "Verifying...") must be immediate, even if the backend is processing.
3.  **"Data Sovereignty":** Approvals are cryptographic commitments. We are building a system of record.
4.  **"Master PRD Compliance":** All flows must respect the multi-tenant architecture and permission matrix defined in Phase 12.

**Verification Standard:**
- **L7 Self-Reflection:** Before marking *any* task complete, you must ask: *"Does this code fundamentally improve the user's trust in the system, or is it just a feature?"*
- **No Regression:** Ensure existing "View Only" modes still work for users without "Edit" permissions.

---

## 📝 TASKS

### [ ] Step 82: Interactive Invoice
**Spec:** [STEP_82_INTERACTIVE_INVOICE.md](../specs/STEP_82_INTERACTIVE_INVOICE.md)

- [ ] **Frontend**: Refactor `fb-artifact-invoice.ts` to support `isEditing` state.
- [ ] **Frontend**: Implement input fields for Quantity, Unit Price, Description.
- [ ] **Frontend**: Implement client-side recalculation of Totals.
- [ ] **Backend**: Verify `PUT /api/finance/invoice/:id` accepts and validates updates.
- [ ] **Verification**:
    - [ ] **Manual**: User can edit a line item, see the total update, and save. Reload persists the change.
    - [ ] **Guardrail**: Non-Draft invoices CANNOT be edited (UI disabled + Backend 403).

### [ ] Step 83: Approval Actions
**Spec:** [STEP_83_APPROVAL_ACTIONS.md](../specs/STEP_83_APPROVAL_ACTIONS.md)

- [ ] **Frontend**: Create `<fb-artifact-actions>` component.
- [ ] **Frontend**: Integrate into Invoice and Estimate artifacts.
- [ ] **Backend**: Implement `POST /approve` and `POST /reject` endpoints with transition logic.
- [ ] **Backend**: Ensure "Approve" locks the artifact (status -> APPROVED).
- [ ] **Verification**:
    - [ ] **Manual**: "Approve" button triggers a modal, then a status change.
    - [ ] **L7 Audit**: Does the "Approved" state clearly convey *finality* to the user? (e.g., Green Badge, disabled inputs).

### [ ] Step 84: Field Feedback Loop
**Spec:** [STEP_84_FIELD_FEEDBACK.md](../specs/STEP_84_FIELD_FEEDBACK.md)

- [ ] **Frontend**: Update `fb-photo-upload.ts` to implement the "Thinking" state machine.
- [ ] **Frontend**: Implement exponential backoff polling for `api.vision.status`.
- [ ] **Backend**: Implement `GET /api/vision/status/:id` (mocked if necessary for now).
- [ ] **Verification**:
    - [ ] **Performance**: Upload -> "Verifying..." (Wait) -> Result. No awkward "flicker".
    - [ ] **UX**: User is never left wondering "Did it upload?".

### [ ] Step 85: Vision Badges
**Spec:** [STEP_85_VISION_BADGES.md](../specs/STEP_85_VISION_BADGES.md)

- [ ] **Frontend**: Create `fb-status-badge` component (Verifying, Verified, Flagged).
- [ ] **Frontend**: Overlay badges on `fb-photo-gallery` thumbnails.
- [ ] **Verification**:
    - [ ] **Visual**: "Verifying" spinner animates. "Flagged" is distinctively red.
    - [ ] **Integration**: Statuses match the `project_assets` database state.
