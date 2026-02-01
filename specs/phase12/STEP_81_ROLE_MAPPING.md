# Step 81: Role Mapping (RBAC)

**Context:** Phase 12, Step 81  
**Goal:** Map Provider roles (Admin/Member) to FutureBuild's `PermissionMatrix` and enforce in Go.  
**Files:**
- [MODIFY] `internal/middleware/auth_middleware.go`
- [MODIFY] `pkg/types/auth.go`
- [NEW] `internal/auth/rbac.go`

---

## 1. Role Definitions

### 1.1 Provider Roles (Clerk)
- `admin`
- `basic_member`
- `guest` (if applicable)

### 1.2 Internal Roles (`pkg/types/auth.go`)
- Update `UserRole` enum:
  ```go
  const (
      RoleOwner   UserRole = "org:admin"   // Map from Clerk 'admin'
      RoleBuilder UserRole = "org:member"  // Map from Clerk 'basic_member'
      RoleViewer  UserRole = "org:viewer"
  )
  ```

---

## 2. Permission Matrix (`internal/auth/rbac.go`)

### 2.1 Scope Definition
Define granular permissions (Scopes):
- `project:create`
- `project:delete`
- `budget:approve`
- `settings:write`

### 2.2 Logic
```go
var RolePermissions = map[UserRole][]Scope{
    RoleOwner:   {ScopeAll},
    RoleBuilder: {ScopeProjectRead, ScopeProjectWrite, ScopeBudgetRead},
    RoleViewer:  {ScopeProjectRead},
}

func Can(role UserRole, scope Scope) bool {
    // ... check logic
}
```

---

## 3. Enforcement

### 3.1 Middleware "Guard"
- Create simpler guards for routes:
  ```go
  r.With(RequirePermission("project:delete")).Delete("/{id}", DeleteProject)
  ```

### 3.2 Endpoint Logic
- Inside handlers, check specifically if operation allows it (e.g., approving a budget > $10k might need `RoleOwner`).

---

## 4. Acceptance Criteria
1.  **Builder Role:** Can create projects but cannot delete Organization or change Billing.
2.  **Viewer Role:** Can `GET /projects` but `POST /projects` returns 403 Forbidden.
3.  **Mapping:** A user with Clerk role `admin` is correctly recognized as `RoleOwner` in the backend.
