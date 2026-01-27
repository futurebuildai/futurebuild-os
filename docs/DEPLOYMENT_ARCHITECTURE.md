# FutureBuild Deployment Architecture

This document serves as the single source of truth for the FutureBuild infrastructure and deployment strategy.

## 🏗️ Environment Map

| Environment | Branch | Purpose | DigitalOcean App ID | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Demo** | `demo` | Client previews & testing | `da91737b-2680-4fed-9acd-e06bc3a6cd0f` | 🔴 Not Configured |
| **Staging** | `staging` | Internal QA & pre-release | `97d84b55-b347-412e-a2b5-ca551461cba9` | 🟢 Configured |
| **Production** | `production` | Live customer traffic | `e74de7f9-b255-4e03-ab55-0b043d4a64b2` | 🔴 Not Configured |

## 🛠️ Infrastructure Stack

- **Cloud Platform**: DigitalOcean App Platform
- **Build Strategy**: Dockerfile (Multi-stage: Node.js frontend + Go binary)
- **Database**: Dev Database (PostgreSQL) for Staging/Demo; Managed Database for Production
- **Intelligence Layer**: Vertex AI (GCP) via Service Account
- **CI/CD**: GitHub Actions (`deploy-staging.yml`, `deploy-demo.yml`, `deploy-prod.yml`)

## 📁 App Spec Structure (`deployment/<env>/app.yaml`)

The `app.yaml` file defines the App Platform configuration. Key sections:

```yaml
name: futurebuild-staging
region: nyc3

# Dev Database (MUST be in 'databases:' block, NOT 'services:')
databases:
- name: db
  engine: PG
  production: false

services:
- name: api
  github:
    repo: futurebuildai/futurebuild-repo
    branch: staging
    deploy_on_push: false  # We control deploys via GitHub Actions
  dockerfile_path: Dockerfile
  source_dir: .
  http_port: 8080
  instance_count: 1
  instance_size_slug: basic-xxs  # 512MB RAM, $5/mo
  envs:
  - key: APP_ENV
    value: staging
    scope: RUN_TIME
  - key: DATABASE_URL
    value: ${db.DATABASE_URL}  # References the 'db' component
    scope: RUN_AND_BUILD_TIME
    type: SECRET
  - key: GCP_SA_JSON_CONTENT
    scope: RUN_TIME
    type: SECRET
  # NOTE: JWT_SECRET and REDIS_URL are managed in Dashboard ONLY
  # Do NOT add them here to prevent automation conflicts
  health_check:
    http_path: /health
```

## 🔑 Environment Variables

### Managed in `app.yaml` (Safe for automation):
| Key | Value | Notes |
|-----|-------|-------|
| `APP_ENV` | `staging` / `demo` / `production` | Hardcoded per environment |
| `DATABASE_URL` | `${db.DATABASE_URL}` | Auto-resolved from DB component |
| `GCP_SA_JSON_CONTENT` | (empty in yaml) | Set value in Dashboard |

### Managed in Dashboard ONLY (Do NOT add to `app.yaml`):
| Key | Example Value | Notes |
|-----|---------------|-------|
| `JWT_SECRET` | `my-super-secret-key-12345` | Required, Encrypted |
| `REDIS_URL` | `redis://localhost:6379` | Optional for MVP |
| `VERTEX_PROJECT_ID` | `futurebuild-ai` | Required for Vertex AI |
| `VERTEX_LOCATION` | `us-central1` | Required for Vertex AI |
| `AUDIT_WAL_PATH` | `/tmp/audit.wal` | Required for staging/production |

> [!CAUTION]
> **Critical Lesson Learned:** If you add a secret to `app.yaml` without a value (e.g., `type: SECRET` only), `doctl apps update` may interpret this as "set to empty", effectively deleting your manually configured secrets. Keep sensitive secrets in the Dashboard ONLY.

## 🔑 Secret Management (GitHub)

The following secrets must be present in GitHub for the pipelines to function:

- `DIGITALOCEAN_ACCESS_TOKEN`: Personal Access Token with Write access.
- `DO_APP_ID_DEMO` / `DO_APP_ID_STAGING` / `DO_APP_ID_PROD`: The unique App IDs.

## 📡 Branching & Workflow Logic

1. **Development**: Feature work happens in `build` or feature branches.
2. **Promotion Flow**: `build` -> `staging` -> `demo` -> `production`.
3. **Trigger Logic**:
   - `demo` / `staging`: Automatic deployment on push via GitHub Actions.
   - `production`: Manual trigger (via `Run workflow` button) or `v*` tag push.

## ⚠️ Critical Lessons Learned

### 1. Dev Database Syntax
Dev Databases use the `databases:` block, NOT `services:`:
```yaml
# ✅ Correct
databases:
- name: db
  engine: PG
  production: false

# ❌ Wrong (causes "unknown field" errors)
services:
- name: db
  type: DATABASE
  engine: PG
```

### 2. `doctl apps update` is Destructive
When you run `doctl apps update --spec app.yaml`, it applies the spec **declaratively**. If a component (like a database) is missing from the yaml, it will be **DELETED** from the app. Always ensure your `app.yaml` is complete.

### 3. Manual Database Creation
If the database doesn't appear after `doctl apps update`, manually add it via:
1. DigitalOcean Dashboard -> App -> "Add components" -> "Database" -> "Dev Database"
2. Name it `db` (must match the reference name in `${db.DATABASE_URL}`)

### 4. GCP Service Account Injection
The Go application reads `GCP_SA_JSON_CONTENT` from the environment, writes it to `/tmp/service-account.json`, and sets `GOOGLE_APPLICATION_CREDENTIALS` automatically. See `internal/config/config.go`.

### 5. GitHub Actions: Use `create-deployment`, NOT `apps update --spec`
The GitHub Actions workflows use `doctl apps create-deployment` to trigger deployments. This preserves all environment variables configured in the Dashboard.

**DO NOT** use `doctl apps update --spec` in CI/CD pipelines—it will delete any env vars not in the yaml file (like JWT_SECRET, VERTEX_PROJECT_ID, etc.).

## 🧠 Deployment Narrative

When deploying, consider the context:
- Is this a quick fix for a client? -> **Demo**
- Is this ready for team review? -> **Staging**
- Is this a stable release for all users? -> **Production**

## 📋 New Environment Setup Checklist

When setting up a new environment (e.g., Demo or Production):

1. [ ] Create DigitalOcean App via Dashboard or `doctl`
2. [ ] Add Dev Database component named `db`
3. [ ] Set environment variables in Dashboard:
   - [ ] `JWT_SECRET` (Encrypted)
   - [ ] `GCP_SA_JSON_CONTENT` (Paste full JSON, Encrypted)
   - [ ] `VERTEX_PROJECT_ID` (e.g., `futurebuild-ai`)
   - [ ] `VERTEX_LOCATION` (e.g., `us-central1`)
   - [ ] `AUDIT_WAL_PATH` (e.g., `/tmp/audit.wal`)
   - [ ] `REDIS_URL` (Optional)
4. [ ] Push to the corresponding branch to trigger GitHub Action
5. [ ] Verify `/health` endpoint returns 200
6. [ ] Verify `/` returns index.html (frontend loads)
