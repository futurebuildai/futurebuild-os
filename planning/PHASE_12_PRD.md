# Phase 12 PRD: Identity & Sovereignty (Auth Refactor)

**Version:** 1.0.0
**Status:** Draft
**Context:** ROADMAP.md (Step 78-81) | Replaces "Magic Link" system

---

## 1. Executive Summary

**Goal:** Replace the fragile, custom "Magic Link" authentication system with a robust, enterprise-grade Identity Provider (Clerk or Auth0). This phase introduces deep multi-tenancy support, allowing users to create organizations, invite members, and manage permissions with granular Role-Based Access Control (RBAC).

**Why:**
- **Security:** Offload sensitive auth logic (MFA, session management, password security) to a specialist provider.
- **Scalability:** Enable "Team" tiers with multi-user access to the same project data.
- **Experience:** Eliminate email delivery delays associated with magic links.

---

## 2. User Stories

### 2.1 The Agency Owner
> "As a builder with 3 project managers, I want to invite them to my 'FutureBuild Organization' so they can see all our active jobs, but I don't want to share my personal login credentials."

### 2.2 The Subcontractor
> "As a plumber working with multiple builders, I want a single login identity that lets me toggle between different General Contractors' workspaces without logging out and back in."

### 2.3 The Developer
> "As a backend engineer, I want to trust the JWT header for user identity and role claims so I stop maintaining custom session cookies and token tables."

---

## 3. Functional Requirements

### 3.1 Identity Provider Integration (Step 78)
- **Primary Auth:** Integrate Clerk (Recommended due to superior React/Next.js DX and Tenant support) or Auth0.
- **Login Flow:**
    - Replace `fb-view-login` with Provider's pre-built component or hosted login page.
    - Support Email/Password + Social Providers (Google, Microsoft).
    - **Remove** internal `auth_tokens` table generation logic.
- **Session Management:**
    - Handle token rotation and expiry automatically via SDK.

### 3.2 Organization Management (Step 80)
- **Tenant Creation:**
    - New users create a "Workspace" (Organization) during onboarding.
- **Invite System:**
    - UI to invite members by email.
    - Status tracking (Pending, Accepted).
- **Settings View:**
    - A new `fb-view-settings-team` page.
    - List members, revoke access, change roles.

### 3.3 Role-Based Access Control (Step 81)
- **Global Roles (Organization Level):**
    - `Owner`: Full access + Billing + Member Management.
    - `Admin`: Full access to Projects + Member Management.
    - `Member`: Create/Edit Projects.
    - `Viewer`: Read-only access to Projects.
- **Scope Mapping:**
    - Map Provider roles to Backend `PermissionMatrix` (Middleware enforcement).

---

## 4. Technical Architecture

### 4.1 Frontend (`/velocity-frontend`)
- **Package:** Install `@clerk/clerk-react` (or equivalent).
- **App Shell:** Wrap `App.tsx` in `<ClerkProvider>`.
- **Components:**
    - `<SignIn />` / `<SignUp />`: Hosted or embedded components.
    - `<OrganizationSwitcher />`: Allow toggling between contexts.
    - `<UserProfile />`: Avatar and account settings.
- **API Client:** Interceptor to inject `Authorization: Bearer <token>` on every request.

### 4.2 Backend (`/velocity-backend`)
- **Middleware Update (Step 79):**
    - **Previous:** Checked database for custom token validity.
    - **New:** `RequireAuth` middleware verifies OIDC/JWKS signature from Provider.
    - Extracts `org_id` and `permissions` from JWT claims.
- **Database Migrations:**
    - **`users` Table:** Add `external_id` (Provider User ID).
    - **`organizations` Table:** Add `external_id` (Provider Org ID).
    - **Cleanup:** Drop `auth_tokens` table after successful migration.

### 4.3 Database Schema Impacts

#### [MODIFY] `organizations`
```sql
ALTER TABLE organizations 
ADD COLUMN external_id VARCHAR(255) UNIQUE, -- Clerk Org ID
ADD COLUMN slug VARCHAR(100) UNIQUE;
```

#### [MODIFY] `users`
```sql
ALTER TABLE users 
ADD COLUMN external_id VARCHAR(255) UNIQUE, -- Clerk User ID
ADD COLUMN avatar_url TEXT; 
-- Remove password_hash if it exists (we were using magic links, so likely clean validation)
```

#### [DELETE] `auth_tokens`
- Table is no longer needed.

---

## 5. Migration Strategy

1.  **Dual Run (Optional but Recommended):**
    - Allow existing Magic Link users to "Claim Account" via email verification on the new Provider, linking their specialized email to the new Identity.
2.  **Hard Cutover (Simpler):**
    - Since we are in **Beta**, a hard cutover is acceptable.
    - Script to provision Organizations in Provider for existing `users` and email them an invite to "Reset Password/Claim Account".

---

## 6. Definition of Done

1.  **No Custom Auth Code:** The `cmd/server/auth.go` handlers for login/verify are removed.
2.  **JWT Validation:** Backend rejects requests without a valid, signed Provider JWT.
3.  **Org Context:** Every API write operation logs the `org_id` from the token, not a request body param.
4.  **Team View:** An Owner can successfully invite a second user who can then log in and see the same projects.

---

## 7. Success Metrics

- **Login Success Rate:** > 99.9% (Provider SLA).
- **Onboarding Friction:** Reduced drop-off during account creation (Social Login vs. waiting for Email Link).
- **Support Volume:** Zero tickets restricted to "I didn't get my login link."
