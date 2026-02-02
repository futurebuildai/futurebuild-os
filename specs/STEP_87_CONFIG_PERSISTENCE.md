# L7 Spec: Step 87 - Config Persistence

**Context:** Phase 14, Step 87
**Goal:** Persist the physics settings to the `business_config` table and update the frontend to fetch/save real data.
**Prerequisites:** Step 86 (UI) must be implemented.

---

## 1. Database Schema
**Migration:** `backend/migrations/0000XX_create_business_config.up.sql`

```sql
CREATE TABLE business_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    speed_multiplier DECIMAL(3, 2) DEFAULT 1.00 NOT NULL,
    work_days JSONB DEFAULT '[1, 2, 3, 4, 5]'::jsonb NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_business_config_org_id UNIQUE (org_id)
);
```

---

## 2. API Implementation
**File:** `backend/internal/api/handlers/settings.go` (and related services)

### 2.1 Backend
1.  **Model:** Define `BusinessConfig` struct in `internal/models`.
2.  **Repo:** Add `GetConfig(orgID)` and `UpdateConfig(orgID, config)` to `SettingsRepository`.
3.  **Handler:** 
    - `GET /api/org/settings/physics`
    - `PUT /api/org/settings/physics`
4.  **Middleware:** Ensure `RequireOrgAccess` protects these routes.

### 2.2 Frontend Integration
1.  **Service:** Add `getPhysics()` and `updatePhysics()` to `api.settings` in `frontend/src/services/api.ts`.
2.  **View:** Update `fb-view-settings.ts` to:
    - Call `getPhysics()` on mount.
    - Populate `_speedMultiplier` and `_workDays`.
    - Call `updatePhysics()` on "Save Changes".

---

## 3. Automated Verification Logic
**Tool:** `/chome` (Claude in Chrome)

**Instructions for the Agent:**
Execute the following verification script using the browser tool:

1.  **Navigate:** Go to `http://localhost:5173/settings`.
2.  **Interact:** 
    - Change "Speed" to `1.2` (Relaxed).
    - Toggle "Sunday" (Day 0/7) to ON.
    - Click "Save Changes".
    - Wait for "Success" toast.
3.  **Reload:** Refresh the page (Command + R).
4.  **Verify:**
    - Assert Slider is still at `1.2`.
    - Assert "Sunday" toggle is still Green/Active.

> **Visual Test Command:**
> `/chome "Go to localhost:5173/settings, set speed to 1.2, enable Sunday, save, reload the page, and verify the settings persisted." --auto-accept`
