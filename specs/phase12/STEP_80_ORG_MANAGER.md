# Step 80: Organization Manager

**Context:** Phase 12, Step 80  
**Goal:** Build "Team Settings" to manage Organizations via the Provider API.  
**Files:**
- [NEW] `frontend/src/components/views/fb-view-team.ts`
- [MODIFY] `internal/api/handlers/organizations.go`
- [MODIFY] `internal/database/models.go`

---

## 1. Backend Changes (Sync & Proxy)

### 1.1 Webhook Handler (`/webhooks/clerk`)
- **Event:** `organization.created`, `organization.updated`.
- **Action:** Sync to local `organizations` table.
  - `id` (Local UUID) <-> `external_id` (Clerk ID).
  - Update `name`, `slug`.

### 1.2 User Sync
- **Event:** `user.created`, `organizationMembership.created`.
- **Action:** Sync to `users` table.
  - Ensures local foreign keys (e.g., `project.owner_id`) point to valid records.

### 1.3 Team Management API (Proxy)
- While Clerk handles the "source of truth", we may perform actions via backend proxy if we want to log audit events, or use the Clerk Frontend SDK directly.
- **Decision:** Use Clerk Component / Frontend SDK for invitation flow to save dev time. Backend only needs to *receive* the result via Webhook.

---

## 2. Frontend Implementation

### 2.1 Team View (`fb-view-team.ts`)
- **Route:** `/settings/team`
- **Component:** Embed Clerk's `<OrganizationProfile />`.
  - Capabilities: Invite Members, Remove Members, Change Roles.
- **Custom Wrapper:**
  - Add "FutureBuild Specifics" if needed (e.g., assigning default projects to new members).

### 2.2 Org Switching
- Add `<OrganizationSwitcher />` to the Sidebar (`fb-sidebar.ts`).
- **Event:** On switch, trigger `api.client.resetStore()` to clear project data from the previous org.

---

## 3. Acceptance Criteria
1.  **Sync:** Creating an Org in the UI results in a row in the `organizations` table (via Webhook).
2.  **Invite:** Admin can invite `user@example.com`, and upon acceptance, `user` appears in `users` table.
3.  **Switching:** Changing Org in Sidebar updates the JWT `org_id` for subsequent requests.
