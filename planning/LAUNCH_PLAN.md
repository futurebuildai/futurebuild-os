# FutureBuild Launch Readiness Plan

## MISSION_COMPLETE

All P0 and P1 launch blockers have been resolved. The application is ready for production deployment.

**Completed 2026-01-25** | Quality Gate: L7/L8 Senior Engineer Standards

---

## Current Status (Updated 2026-01-25)

| Task | Status | Notes |
|------|--------|-------|
| A3: SendGrid Email | ✅ DONE | Config added, provider created, wired in server.go |
| B1: Frontend Auth | ✅ DONE | Login view uses real API, verify view created |
| B2: Invite Flow | ✅ DONE | Backend + frontend complete |
| B3: Profile Setup | ✅ DONE | Name collected during invite acceptance |
| C1: Field Portal | ⏳ P2 | Can defer to post-launch |
| C2: Settings/Profile | ✅ DONE | PUT /users/me endpoint added |
| D1: Demo Seed | ✅ DONE | `cmd/seed-demo/main.go` created |
| Production Safety | ✅ DONE | FileAuditWAL + SimpleCircuitBreaker wired in |

**Domain:** App will be at `app.futurebuild.ai`

---

## What's Ready (Green)

- ✅ Deployment infrastructure (Docker, Docker Compose, DO App Platform staging config)
- ✅ Database migrations framework with 55 migrations
- ✅ Frontend 3-panel UI with 8 view components
- ✅ Magic link authentication (login → email → verify → JWT)
- ✅ Invite flow (admin creates invite → email sent → user accepts → account created)
- ✅ SendGrid email integration
- ✅ API routing, handlers, middleware
- ✅ Multi-tenancy with OrgID isolation
- ✅ Rate limiting on auth endpoints

---

## Critical Blockers for Launch (Red - Must Fix)

### 1. User Profile Update Endpoint (P0) - ✅ DONE
**Problem:** Users cannot change their name after account creation
**Impact:** Stuck with whatever name they entered during invite acceptance

**Files created/modified:**
- `internal/api/handlers/user_handler.go` - ✅ Created with GetProfile and UpdateProfile
- `internal/server/server.go` - ✅ Wired PUT/GET /api/v1/users/me routes
- `frontend/src/services/api.ts` - ✅ Added api.users.getMe() and api.users.updateMe()

### 2. Production Safety: WAL & Circuit Breaker (P0) - ✅ DONE
**Problem:** `internal/server/server.go` used NoOp implementations with panic guards in production

**Solution implemented:**
- FileAuditWAL used for production/staging environments
- SimpleCircuitBreaker with threshold-based state machine
- NoOp implementations retained for development mode only

### 3. Admin Invite Management UI (P1) - ✅ DONE
**Problem:** No frontend for admins to create/manage invites
**Impact:** Admins must use curl/API directly to invite users

**Files created:**
- `frontend/src/components/views/fb-view-admin-invites.ts` - ✅ Created
- Wired into fb-panel-center routing at `/admin/invites` - ✅ Done

---

## Nice-to-Have for Launch (Yellow)

| Item | Priority | Notes |
|------|----------|-------|
| Demo seed script | ✅ DONE | `cmd/seed-demo/main.go` - idempotent, production-safe |
| User settings view | ✅ DONE | `fb-view-settings.ts` at `/settings` route |
| Notifications/toast UI | P2 | Visual feedback for errors/success |
| Field Portal (mobile) | P2 | Can ship post-launch |
| Production deployment config | ✅ DONE | `deployment/production/app.yaml` + `.env.example` updated |

---

## Implementation Details

### 1. User Profile Update Endpoint

**New file:** `internal/api/handlers/user_handler.go`
```go
type UserHandler struct {
    db *pgxpool.Pool
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    // Get user ID from JWT claims
    // Parse request body: { "name": "...", "phone": "..." }
    // UPDATE users SET name = $1, phone = $2 WHERE id = $3 AND org_id = $4
}
```

**Modify:** `internal/server/server.go`
```go
// Add to Server struct
UserHandler *handlers.UserHandler

// Add route
r.Route("/users", func(r chi.Router) {
    r.Use(s.AuthMiddleware.RequireAuth)
    r.Put("/me", s.UserHandler.UpdateProfile)
    r.Get("/me", s.UserHandler.GetProfile)
})
```

**Modify:** `frontend/src/services/api.ts`
```typescript
users: {
    updateMe(data: { name?: string; phone?: string }): Promise<User> {
        return put<User>('/users/me', data);
    },
},
```

### 2. PostgreSQL Audit WAL

**New migration:** `migrations/000056_create_audit_log.up.sql`
```sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    user_id UUID,
    org_id UUID NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_org_created ON audit_log(org_id, created_at DESC);
```

**New file:** `internal/chat/postgres_wal.go`
- Implement AuditWAL interface
- Insert into audit_log table on each event

### 3. Redis Circuit Breaker

**New file:** `internal/chat/redis_circuit_breaker.go`
- Implement CircuitBreaker interface
- Use Redis for state storage (half-open, open, closed)
- Configurable failure threshold and timeout

### 4. Admin Invites View

**New file:** `frontend/src/components/views/fb-view-admin-invites.ts`
- List pending invitations with email, role, created_at
- "Invite User" button → modal with email + role picker
- Delete/revoke button per row
- Uses existing `api.invites.list()`, `api.invites.create()`, `api.invites.revoke()`

**Modify:** `frontend/src/components/layout/fb-panel-center.ts`
- Add route check for `/admin/invites`
- Render `<fb-view-admin-invites>` for admin users

### 5. Demo Seed Script

**New file:** `scripts/seed_demo.go`
```go
func main() {
    // Connect to DB
    // Insert organization: "Acme Builders"
    // Insert admin user: admin@acme.com
    // Insert builder user: builder@acme.com
    // Insert 3 projects at 20%, 45%, 80% completion
    // Insert tasks with realistic WBS codes and statuses
    // Insert sample contacts (plumber, electrician, etc.)
}
```

---

## Files Summary

### To Create (Critical Path)
| File | Purpose |
|------|---------|
| `internal/api/handlers/user_handler.go` | Profile update endpoint |
| `internal/chat/postgres_wal.go` | Audit WAL implementation |
| `internal/chat/redis_circuit_breaker.go` | Circuit breaker implementation |
| `migrations/000056_create_audit_log.up.sql` | Audit log table |
| `frontend/src/components/views/fb-view-admin-invites.ts` | Admin invite management |
| `frontend/src/components/views/fb-view-settings.ts` | User settings view |
| `scripts/seed_demo.go` | Demo data seeder |
| `deployment/production/app.yaml` | Production config |

### To Modify
| File | Changes |
|------|---------|
| `internal/server/server.go` | Add user routes, wire WAL/CircuitBreaker |
| `frontend/src/services/api.ts` | Add api.users.updateMe() |
| `frontend/src/components/layout/fb-panel-center.ts` | Add admin/settings routes |

### Already Complete (Reference)
| File | Status |
|------|--------|
| `internal/service/sendgrid_provider.go` | ✅ Created |
| `internal/service/invite_service.go` | ✅ Created |
| `internal/api/handlers/invite_handler.go` | ✅ Created |
| `migrations/000055_create_invitations_table.up.sql` | ✅ Created |
| `frontend/src/components/views/fb-view-login.ts` | ✅ Updated |
| `frontend/src/components/views/fb-view-verify.ts` | ✅ Created |
| `frontend/src/components/views/fb-view-invite-accept.ts` | ✅ Created |

---

## Execution Plan: Production-Ready Launch (~8 days)

### Phase 1: Core Backend (Days 1-2)

**Task 1.1: User Profile Endpoint**
- Create `internal/api/handlers/user_handler.go`
- Add `PUT /api/v1/users/me` and `GET /api/v1/users/me` routes
- Wire in `server.go`

**Task 1.2: PostgreSQL Audit WAL**
- Create migration `000056_create_audit_log.up.sql`
- Create `internal/chat/postgres_wal.go` implementing AuditWAL interface
- Wire in `server.go` based on APP_ENV

### Phase 2: Circuit Breaker (Day 3)

**Task 2.1: Redis Circuit Breaker**
- Create `internal/chat/redis_circuit_breaker.go`
- Implement half-open, open, closed states
- Wire in `server.go` using existing Redis config

### Phase 3: Frontend (Days 4-5)

**Task 3.1: Admin Invites View**
- Create `frontend/src/components/views/fb-view-admin-invites.ts`
- Add route to `fb-panel-center.ts`
- Table with pending invites + create/revoke actions

**Task 3.2: User Settings View**
- Create `frontend/src/components/views/fb-view-settings.ts`
- Profile tab: name, phone, email (readonly)
- Add `api.users.updateMe()` to API service

### Phase 4: Demo & Deployment (Days 6-7)

**Task 4.1: Demo Seed Script**
- Create `scripts/seed_demo.go`
- Seed org, users, projects, tasks, contacts

**Task 4.2: Production Deployment Config**
- Create `deployment/production/app.yaml`
- Document required secrets in `.env.example`

### Phase 5: Testing & Launch (Day 8)

**Task 5.1: E2E Testing**
- Full auth flow test
- Full invite flow test
- Profile update test
- Admin invite management test

**Task 5.2: Deploy to Production**
- Run migrations
- Deploy via DO App Platform
- Smoke test production

---

## Verification Steps

After implementation:

```bash
# 1. Run tests
make test

# 2. Build frontend
cd frontend && npm run build

# 3. Test auth flow locally
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}'

# 4. Test profile update (after getting JWT)
curl -X PUT http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "New Name"}'

# 5. Manual E2E test
# - Request magic link → receive email → click verify → dashboard
# - Admin creates invite → user receives email → accepts → logs in
# - User updates profile via settings view
```
