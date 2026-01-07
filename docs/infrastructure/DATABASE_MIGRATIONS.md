# Database Migrations Guide

// See PRODUCTION_PLAN.md Step 7

## Overview

FutureBuild uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management. All schema changes are tracked as versioned migration files.

---

## 1. Directory Structure

```
migrations/
├── 000001_init.up.sql           # Initial extensions
├── 000001_init.down.sql         # Rollback for init
├── 000002_identity.up.sql       # (Phase 1) Organizations, Users, Contacts
├── 000002_identity.down.sql
└── ...
```

**Naming Convention:** `{version}_{description}.{up|down}.sql`

---

## 2. Install golang-migrate CLI

```bash
# macOS
brew install golang-migrate

# Linux (Debian/Ubuntu)
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Verify installation
migrate -version
```

---

## 3. Running Migrations

### Local Development (Docker Compose)

```bash
# Apply all pending migrations
migrate -database "${DATABASE_URL}" -path migrations up

# Rollback last migration
migrate -database "${DATABASE_URL}" -path migrations down 1

# Go to specific version
migrate -database "${DATABASE_URL}" -path migrations goto 2

# Check current version
migrate -database "${DATABASE_URL}" -path migrations version
```

### Using Makefile (Recommended)

```bash
# Apply migrations
make migrate-up

# Rollback
make migrate-down

# Create new migration
make migrate-create name=add_projects_table
```

---

## 4. Creating New Migrations

```bash
# Create a new migration pair
migrate create -ext sql -dir migrations -seq add_projects_table
```

This generates:
- `migrations/000002_add_projects_table.up.sql`
- `migrations/000002_add_projects_table.down.sql`

### Rules for Migration Files

1. **Idempotent where possible** — Use `IF NOT EXISTS`, `IF EXISTS`
2. **Always pair up/down** — Every `.up.sql` needs a `.down.sql`
3. **One concern per migration** — Single table or related changes
4. **No data migrations mixed with schema** — Use separate migrations

---

## 5. Environment-Specific Execution

| Environment | Method | Trigger |
|-------------|--------|---------|
| **Local** | `make migrate-up` | Manual |
| **Staging** | App Platform Job | On deploy (via Dockerfile) |
| **Production** | App Platform Job | Manual approval + deploy |

### Staging/Production (DO App Platform)

Migrations run automatically via the Dockerfile entrypoint:

```dockerfile
# In production Dockerfile
CMD ["sh", "-c", "migrate -database $DATABASE_URL -path /app/migrations up && ./api"]
```

---

## 6. Troubleshooting

### Dirty Database State

If a migration fails mid-execution:

```bash
# Force set version (use with caution)
migrate -database "${DATABASE_URL}" -path migrations force 1
```

### Check Migration Status

```bash
migrate -database "${DATABASE_URL}" -path migrations version
```

---

## 7. Spec References

| Domain | Spec Section | Migration |
|--------|--------------|-----------|
| Identity & Access | `DATA_SPINE_SPEC.md` Domain 1 | `000002_identity` |
| Project Core | `DATA_SPINE_SPEC.md` Domain 2 | `000003_projects` |
| Financials | `DATA_SPINE_SPEC.md` Domain 3 | `000004_financials` |
| Communication | `DATA_SPINE_SPEC.md` Domain 4 | `000005_communication` |
| Learning | `DATA_SPINE_SPEC.md` Domain 5 | `000006_learning` |

---

*Verified against PRODUCTION_PLAN.md Step 7 and DATA_SPINE_SPEC.md.*
