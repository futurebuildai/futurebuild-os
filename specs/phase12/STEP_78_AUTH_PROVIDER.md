# Step 78: Auth Provider Integration (Clerk)

**Context:** Phase 12, Step 78  
**Goal:** Replace custom "Magic Link" auth with Clerk Identity Provider.  
**Files:**
- [MODIFY] `frontend/package.json`
- [MODIFY] `frontend/src/App.ts`
- [MODIFY] `frontend/src/components/views/fb-view-login.ts`
- [MODIFY] `frontend/src/services/api/client.ts`
- [DELETE] `internal/api/handlers/auth.go` (Partial/Full)

---

## 1. Frontend Integration

### 1.1 Dependencies
- Install `@clerk/clerk-react` (or `@clerk/lit` wrapper if available, otherwise adapt vanilla JS SDK).
- **Note:** Since we use Lit, we will likely need NOT use the React wrapper. Use `@clerk/clerk-js` directly or a lightweight Lit wrapper.
- **Decision:** Use `@clerk/clerk-js` for framework-agnostic implementation.

### 1.2 App Shell Update (`App.ts`)
- Initialize Clerk in `connectedCallback`.
- Wait for `clerk.load()` before rendering the router.
- Replace `auth` state management with Clerk's `session` state.

### 1.3 View Replacement (`fb-view-login.ts`)
- **Remove:** All magic link form logic (`_email`, `_handleLogin`, `api.auth.requestMagicLink`).
- **Implement:** Clerk's mounted UI.
  ```typescript
  // Inside render/firstUpdated
  const signInDiv = this.shadowRoot.getElementById('clerk-sign-in');
  clerk.mountSignIn(signInDiv);
  ```
- **Styling:** Ensure container matches "Construction Professional" dark mode.

### 1.4 API Client (`client.ts`)
- Update `RequestInterceptor`.
- **Old:** Read token from `localStorage.getItem('auth_token')`.
- **New:** Call `await clerk.session.getToken({ template: 'futurebuild-backend' })`.
- **Header:** `Authorization: Bearer <token>`.

---

## 2. Backend Cleanup

### 2.1 Remove Legacy Routes
- Delete/Deprecate:
    - `POST /api/v1/auth/magic-link`
    - `GET /api/v1/auth/verify/{token}`
    - `POST /api/v1/auth/logout`

### 2.2 Schema Cleanup
- Drop `auth_tokens` table is deferred to Step 79/81 (keep for rollback safety during Step 78).

---

## 3. Acceptance Criteria
1.  **Login:** User can sign in using Google or Email via Clerk UI.
2.  **Token:** `api.client` successfully sends a Clerk-minted JWT to the backend.
3.  **Persistence:** Refresh gives a valid session without re-login.
