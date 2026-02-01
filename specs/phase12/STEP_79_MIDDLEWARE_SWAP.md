# Step 79: Middleware Swap (JWT Validation)

**Context:** Phase 12, Step 79  
**Goal:** Update Go middleware to validate 3rd-party JWTs (Clerk) instead of internal tokens.  
**Files:**
- [MODIFY] `internal/middleware/auth_middleware.go`
- [MODIFY] `pkg/types/auth.go`
- [NEW] `internal/auth/jwks.go`

---

## 1. JWKS Validation Logic

### 1.1 Key Set Retrieval
- Implement a JWKS (JSON Web Key Set) cache.
- **Library:** Use `github.com/MicahParks/keyfunc/v2` or similar.
- **Config:** `CLERK_ISSUER_URL` (e.g., `https://clerk.futurebuild.ai`).

### 1.2 Claims Extraction Update
- **Old:** Looked up `token_hash` in `auth_tokens` table.
- **New:** Stateless verification.
  - Parse JWT from `Authorization: Bearer` header.
  - Verify Signature using JWKS.
  - Verify Issuer (`iss`), Audience (`aud`), and Expiry (`exp`).

### 1.3 Custom Claims Mapping
- Map Clerk claims to `types.Claims`.
- **Clerk Template:** We must configure Clerk to inject:
  ```json
  {
    "org_id": "{{org_id}}",
    "role": "{{org_role}}",
    "subject_type": "user"
  }
  ```
- **Fallback:** If `org_id` is missing (user has Personal Workspace only), resolve/create a default organization or handle as 403 `ErrNoOrg`.

---

## 2. Middleware Implementation

### 2.1 `RequireAuth`
- **Input:** Request Context.
- **Process:** 
    1. Extract Token.
    2. Verify (JWKS).
    3. `claims := ExtractClaims(token)`.
    4. `ctx = context.WithValue(ctx, types.UserContextKey, claims)`.
    5. `next.ServeHTTP`.

### 2.2 Handling Legacy Tokens
- **Strategy:** Hard Cutover (per PRD).
- **Action:** Remove code paths checking `auth_tokens` DB.

---

## 3. Acceptance Criteria
1.  **Valid Token:** A request with a valid Clerk JWT passes middleware and populates `request.Context`.
2.  **Invalid Token:** Expired or wrong-issuer tokens return 401 Unauthorized.
3.  **No Org:** A user without an `org_id` in claims is flagged (or handled gracefully).
