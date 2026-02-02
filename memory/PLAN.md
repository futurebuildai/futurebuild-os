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

### [x] Step 82: Interactive Invoice
**Spec:** [STEP_82_INTERACTIVE_INVOICE.md](../specs/STEP_82_INTERACTIVE_INVOICE.md)

- [x] **Frontend**: Refactor `fb-artifact-invoice.ts` to support `isEditing` state.
- [x] **Frontend**: Implement input fields for Quantity, Unit Price, Description.
- [x] **Frontend**: Implement client-side recalculation of Totals.
- [x] **Backend**: `PUT /api/v1/invoices/:id` with full validation + `GET /api/v1/invoices/:id`.
- [x] **Verification**:
    - [x] **Manual**: User can edit a line item, see the total update, and save. Reload persists the change.
    - [x] **Guardrail**: Non-Draft invoices CANNOT be edited (UI disabled + Backend 403).
- [x] **L7 Audit**: 3 CRITICAL, 4 HIGH, 5 MEDIUM remediated and re-verified.
    - C1: TOCTOU race fixed with atomic UPDATE WHERE status='Draft'
    - C2: Float truncation fixed with math.Round
    - C3: Missing org_id guard added to UPDATE query
    - H1: MaxLineItems=100 limit added
    - H2: Description length limit (500) enforced server-side
    - H3: Dead `notes` parameter removed from interface
    - H4: Rate limiting deferred (covered by auth + body size limit)
    - M1: Sentinel error ErrInvoiceNotEditable with errors.Is
    - M3: Default status changed to Pending (non-editable) instead of Draft

### [x] Step 83: Approval Actions
**Spec:** [STEP_83_APPROVAL_ACTIONS.md](../specs/STEP_83_APPROVAL_ACTIONS.md)

- [x] **Frontend**: Created `<fb-artifact-actions>` component with Approve/Reject modals.
- [x] **Frontend**: Integrated into `fb-artifact-invoice.ts` with finality badges.
- [x] **Backend**: Implemented `POST /api/v1/invoices/:id/approve` and `/reject` with atomic transitions.
- [x] **Backend**: Approved state locks artifact — irreversible, recorded with approver ID + timestamp.
- [x] **Verification**:
    - [x] **Manual**: Approve/Reject buttons trigger confirmation modals, then status change.
    - [x] **L7 Audit**: Green/Red finality badges convey permanence. Edit button hidden after approval.
- [x] **L7 Audit**: 1 CRITICAL, 3 HIGH remediated.
    - C1: InvoiceResponse type now includes all 5 approval metadata fields (Rosetta Stone parity)
    - H1: Replaced brittle string comparison with pgx.ErrNoRows sentinel (3 occurrences)
    - H2: Rejection reason truncated to 100 chars for logging (log injection prevention)
    - H3: Fixed reject reason captured before clearing in event emission

### [x] Step 84: Field Feedback Loop
**Spec:** [STEP_84_FIELD_FEEDBACK.md](../specs/STEP_84_FIELD_FEEDBACK.md)

- [x] **Frontend**: Update `fb-photo-upload.ts` to implement the "Thinking" state machine.
- [x] **Frontend**: Implement exponential backoff polling for `api.vision.status`.
- [x] **Backend**: Implement `GET /api/vision/status/:id` with multi-tenancy guard.
- [x] **Backend**: Created `project_assets` table (migration 000064) + `ProjectAsset` model.
- [x] **Verification**:
    - [x] **Performance**: Upload -> "Verifying..." (Wait) -> Result. No awkward "flicker".
    - [x] **UX**: User is never left wondering "Did it upload?".
- [x] **L7 Audit**: 2 HIGH, 2 MEDIUM remediated and re-verified.
    - H1: Double-upload race condition fixed with state guard in upload()
    - H2: Polling failure limit (MAX_POLL_FAILURES=5) added alongside 30s timeout
    - M1: RBAC scope (budget:read) added to vision status route
    - M2: Portal auth gap documented for future upload handler integration

### [x] Step 85: Vision Badges
**Spec:** [STEP_85_VISION_BADGES.md](../specs/STEP_85_VISION_BADGES.md)

- [x] **Frontend**: Updated `fb-vision-badge` with spec-correct colors (yellow=verifying, green=verified, red=flagged).
- [x] **Frontend**: Added click interaction for flagged/verified badges with `fb-badge-click` event.
- [x] **Frontend**: Created `fb-photo-gallery` component with badge overlays on thumbnails.
- [x] **Backend**: Added `GET /api/v1/projects/:id/assets` endpoint with multi-tenancy guard.
- [x] **Frontend**: Added `api.assets.list()` method and `ProjectAssetResponse` type.
- [x] **Verification**:
    - [x] **Visual**: "Verifying" spinner animates. "Flagged" is distinctively red.
    - [x] **Integration**: Statuses match the `project_assets` database state.
- [x] **L7 Audit**: 1 HIGH, 1 MEDIUM remediated and re-verified.
    - H1: Added LIMIT 200 to ListProjectAssets query (DoS prevention)
    - M1: Centralized `assetReturnColumns` + `scanAsset()` helper (DRY compliance)
