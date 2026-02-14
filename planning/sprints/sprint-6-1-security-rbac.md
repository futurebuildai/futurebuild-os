# Sprint 6.1: Security & RBAC

> **Epic:** 6 — Production Hardening (Google Audit)
> **Depends On:** None (can run in parallel with earlier EPICs)
> **Objective:** Ensure every route is protected and role-based access controls are enforced.

---

## Sprint Tasks

### Task 6.1.1: Audit `auth_middleware.go`

**Status:** ⬜ Not Started

**Current State:**
- [auth_middleware.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/middleware/auth_middleware.md) — placeholder stub
- [auth_handler.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/api/handlers/auth_handler.md) — placeholder stub
- [auth_service.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/service/auth_service.md) — placeholder stub
- Frontend uses Clerk for auth: [clerk.ts service](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/services/clerk) manages tokens
- [fb-app-shell.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/layout/fb-app-shell.ts) has admin route protection (line 703-712)

**Audit Checklist:**

1. **Route inventory:** List every API route and its auth status
   ```
   Route                              Auth?   Role Required?
   ─────────────────────────────────────────────────────────
   GET  /api/v1/portfolio/feed         ☐       Any
   POST /api/v1/agent/onboard          ☐       PM+
   GET  /api/v1/projects/:id           ☐       Project Member
   POST /api/v1/corrections            ☐       Any
   GET  /api/v1/financials/summary     ☐       PM+
   GET  /admin/*                       ☐       Platform Admin
   ```

2. **Verify middleware chain:** Every route group must pass through auth middleware
3. **Token validation:** Clerk JWT verification with proper audience and issuer checks
4. **Session management:** Token expiry, refresh token flow, session invalidation

**Atomic Steps:**

1. Create route inventory spreadsheet (or markdown table)
2. For each route, verify auth middleware is applied
3. Identify unprotected routes → add middleware
4. Add integration test: Unauthenticated request → verify 401 response
5. Add integration test: Expired token → verify 401 response
6. Add integration test: Valid token → verify 200 response

---

### Task 6.1.2: Verify Role Mapping & Access Control

**Status:** ⬜ Not Started

**Roles:**

| Role | Access Level | Can See Budget? | Can Approve Invoices? | Can Access Admin? |
|------|-------------|-----------------|----------------------|-------------------|
| Admin | Full | ✅ | ✅ | ✅ |
| PM (Project Manager) | Project-scoped + Financials | ✅ | ✅ | ❌ |
| Subcontractor | Own tasks only | ❌ | ❌ | ❌ |
| Viewer | Read-only | ❌ | ❌ | ❌ |

**Current Frontend State:**
- [types/enums.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/types/enums) likely has `UserRole` type
- `store.user$.value?.role` used in some places
- `isPlatformAdmin()` check exists in `fb-app-shell.ts`

**Atomic Steps:**

1. **Backend: Implement role-checking middleware:**
   ```go
   func RequireRole(roles ...string) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               user := getUserFromContext(r.Context())
               if !slices.Contains(roles, user.Role) {
                   http.Error(w, "Forbidden", http.StatusForbidden)
                   return
               }
               next.ServeHTTP(w, r)
           })
       }
   }
   ```

2. **Apply role middleware to sensitive routes:**
   - Budget/Financial endpoints → `RequireRole("Admin", "PM")`
   - Invoice approval → `RequireRole("Admin", "PM")`
   - Admin routes → `RequireRole("Admin")`
   - Project creation → `RequireRole("Admin", "PM")`

3. **Frontend: Hide UI elements based on role:**
   - Hide "Budget" nav item for Subcontractor/Viewer
   - Hide "Approve" button on invoices for Subcontractor/Viewer
   - Disable "Admin" route for non-admins (already partially done)

4. **Test: URL manipulation prevention:**
   - Login as Subcontractor → navigate to `/budget` via URL bar → verify redirect or 403
   - Login as Viewer → try to approve invoice via API → verify 403

---

## Codebase References

| File | Path | Status | Notes |
|------|------|--------|-------|
| auth_middleware.md | `backend/shadow/internal/middleware/auth_middleware.md` | Stub | Needs Go implementation |
| auth_handler.md | `backend/shadow/internal/api/handlers/auth_handler.md` | Stub | Needs Go implementation |
| auth_service.md | `backend/shadow/internal/service/auth_service.md` | Stub | Needs Go implementation |
| auth.md | `backend/shadow/pkg/types/auth.md` | Stub | Role types |
| fb-app-shell.ts | `frontend/src/components/layout/fb-app-shell.ts` | 783 | Admin check exists (L703) |
| fb-left-nav.ts | `frontend/src/components/layout/fb-left-nav.ts` | 507 | Needs role-based item visibility |

## Verification Plan

- **Automated:** For each route, send request without auth token → verify 401
- **Automated:** For each role-restricted route, send request with wrong role → verify 403
- **Manual:** Login as Subcontractor → verify Budget nav item is hidden
- **Manual:** Login as Subcontractor → type `/budget` in URL bar → verify access denied
- **Manual:** Login as Admin → verify all nav items and actions are available
