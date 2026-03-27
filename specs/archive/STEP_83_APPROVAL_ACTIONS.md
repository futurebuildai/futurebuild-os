# L7 Spec: Step 83 - Approval Actions

**Context:** Phase 13, Step 83
**Goal:** Implement the "Approve/Reject" logic for critical artifacts (Invoice, Estimate). This action must be cryptographically signed (via JWT context) and irreversible for the "Approve" case without Admin intervention.

---

## 1. Component Architecture
**Files:**
- `/velocity-frontend/src/components/artifacts/fb-artifact-invoice.ts`
- `/velocity-frontend/src/components/artifacts/fb-artifact-estimate.ts`

### 1.1 UI/UX
- **Action Bar:**
    - A new footer component: `<fb-artifact-actions .status=${this.status} .permissions=${this.permissions}>`
    - **Draft:** Shows "Approve" (Green) | "Reject" (Red)
    - **Approved:** Shows "Approved by [User]" (Green Badge) - Actions Hidden.
    - **Rejected:** Shows "Rejected by [User]" (Red Badge) - Actions Hidden.

### 1.2 Frontend Logic
- **Approve Click:**
    - Show confirmation modal: "Are you sure? This will finalize the document."
    - Call `api.finance.approve(id)`.
    - Optimistic update: Set local status to `approving...` -> `approved`.
- **Reject Click:**
    - Show reason modal: "Reason for rejection?" (Textarea).
    - Call `api.finance.reject(id, reason)`.

---

## 2. API Contract

### 2.1 Endpoints
- `POST /api/finance/invoices/:id/approve`
- `POST /api/finance/invoices/:id/reject`

### 2.2 Request Body (Reject)
```json
{
  "reason": "Labor costs are too high compared to estimate."
}
```

### 2.3 Backend Logic (`/velocity-backend`)
- **Handlers:** `HandleApproveInvoice`, `HandleRejectInvoice`.
- **Validation:**
    - User must have `APPROVE_FINANCE` permission.
    - Artifact must be in a specific state (e.g., cannot reject an already paid invoice).
- **Side Effects:**
    - Log transition in `audit_log` table.
    - Notify creator via `NotificationService` (Phase 15, but stub logic now).

---

## 3. Implementation Steps (Claude Code Instructions)

1.  **Backend Extensions**:
    - Update `Invoice` struct/model to include `ApproverID` and `ApprovedAt`.
    - Implement `HandleApprove` which updates the status and records metadata.

2.  **Frontend Components**:
    - Create `fb-artifact-actions.ts`: A shared LitElement for the buttons.
    - Integrate into Invoice and Estimate artifacts at the bottom slot.

3.  **Permissions**:
    - Ensure the "Approve" button is `disabled` or `hidden` if `user.permissions.includes('APPROVE_FINANCE')` is false.

---

## 4. Verification
- **Test:** Login as Owner. View Draft Invoice. Click Approve. Check Network tab for 200 OK. Refresh page. Verify status is "Approved" and buttons are gone.
- **Audit:** Check DB `audit_logs` for an entry with `action: "INVOICE_APPROVE"`.
