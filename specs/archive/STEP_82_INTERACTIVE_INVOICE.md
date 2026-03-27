# L7 Spec: Step 82 - Interactive Invoice

**Context:** Phase 13, Step 82
**Goal:** Convert the static `fb-artifact-invoice` component into an editable form that allows users to correct AI assumptions before approval.

---

## 1. Component Architecture
**File:** `/velocity-frontend/src/components/artifacts/fb-artifact-invoice.ts`

### 1.1 State Management
- **Internal State:**
    - `isEditing`: boolean (toggles view vs. edit mode).
    - `draftItems`: Copy of `invoice.items` for mutation.
    - `dirty`: boolean (true if changes made).

### 1.2 UX Flow
1.  **Initial Render:** Read-only view (Status: `DRAFT`).
2.  **Edit Trigger:** "Edit Invoice" button (only visible if status IS `DRAFT`).
3.  **Edit Mode:**
    - Line items become `<input>` fields.
    - `qty` (number), `unit_price` (currency), `description` (text).
    - "Add Item" button to append row.
    - "Delete" icon on hover per row.
4.  **Recalculation:**
    - `subtotal = sum(qty * unit_price)`
    - `total = subtotal + tax`
    - Recalculate immediately on `input` change.
5.  **Save:**
    - `api.finance.updateInvoice(id, draftItems)`
    - On success: Toggle `isEditing` -> false, update parent state.

---

## 2. API Contract
**Endpoint:** `PUT /api/finance/invoices/:id`

### 2.1 Request Body
```json
{
  "items": [
    {
      "description": "Drywall Installation",
      "quantity": 100,
      "unit_price": 45.00,
      "unit": "sqft"
    }
  ],
  "notes": "Updated labor rate per discussion."
}
```

### 2.2 Validation (Backend)
- **Middleware:** `RequireAuth`, `RequireOrgAccess`.
- **Logic:**
    - Invoice MUST be in `draft` status.
    - `quantity` > 0.
    - `unit_price` >= 0.

---

## 3. Implementation Steps (Claude Code Instructions)

1.  **Modify `fb-artifact-invoice.ts`**:
    - Import `LitElement`, `state` from `lit`.
    - Add `isEditing` state property.
    - Create `renderEditRow(item, index)` helper.
    - Bind inputs to `this.updateItem(index, field, value)`.

2.  **Add Services**:
    - Ensure `InvoiceService` has `update(id, payload)`.

3.  **Security Check**:
    - Ensure only users with `EDIT_FINANCE` permission can see the "Edit" button.

---

## 4. Verification
- **Test:** Open a DRAFT invoice. Click Edit. Change price. See Total update. Click Save. Reload page. Verify persistence.
- **Guardrail:** Attempt to edit an `APPROVED` invoice -> Should fail (UI disabled + Backend 403).
