# FutureBuild Deployment Architecture

This document serves as the single source of truth for the FutureBuild infrastructure and deployment strategy.

## 🏗️ Environment Map

| Environment | Branch | Purpose | DigitalOcean App ID |
| :--- | :--- | :--- | :--- |
| **Demo** | `demo` | Client previews & testing | `da91737b-2680-4fed-9acd-e06bc3a6cd0f` |
| **Staging** | `staging` | Internal QA & pre-release | `97d84b55-b347-412e-a2b5-ca551461cba9` |
| **Production** | `production` | Live customer traffic | `e74de7f9-b255-4e03-ab55-0b043d4a64b2` |

## 🛠️ Infrastructure Stack

- **Cloud Platform**: DigitalOcean App Platform
- **Build Strategy**: Dockerfile (Multi-stage: Node.js frontend + Go binary)
- **Database**: Managed PostgreSQL (Dev Database for Staging/Demo)
- **Intelligence Layer**: Vertex AI (GCP)
- **CI/CD**: GitHub Actions

## 🔑 Secret Management (GitHub)

The following secrets must be present in GitHub for the pipelines to function:

- `DIGITALOCEAN_ACCESS_TOKEN`: Personal Access Token with Write access.
- `DO_APP_ID_DEMO` / `STAGING` / `PROD`: The unique IDs for the App Platform projects.
- `GCP_SA_JSON`: Google Service Account JSON for Vertex AI access.

## 📡 Branching & Workflow Logic

1. **Development**: Feature work happens in `build` or feature branches.
2. **Promotion Flow**: `build` -> `staging` -> `demo` -> `production`.
3. **Trigger Logic**:
   - `demo` / `staging`: Automatic deployment on every push/merge.
   - `production`: Manual trigger (via `Run workflow` button) or `v*` tag push.

## 🧠 Deployment Narrative
When deploying, the assistant should always consider the context:
- Is this a quick fix for a client? -> **Demo**
- Is this ready for team review? -> **Staging**
- Is this a stable release for all users? -> **Production**
