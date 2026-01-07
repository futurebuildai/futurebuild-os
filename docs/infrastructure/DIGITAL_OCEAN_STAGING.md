# Staging Infrastructure Requirements: Digital Ocean

// See PRODUCTION_PLAN.md Step 5

## 1. App Platform Configuration (futurebuild-staging)

| Component | Type | Instance Size | Scale | Role |
|-----------|------|---------------|-------|------|
| **api** | Web Service | Basic (1 vCPU, 512 MB) | 1 | Chi Router + Frontend Static Assets |
| **worker** | Worker | Basic (1 vCPU, 512 MB) | 1 | Asynq Background Tasks (Invoice OCR, Notifications) |

## 2. Managed Database (Managed PostgreSQL)

- **Version:** PostgreSQL 15+
- **Extensions:** `pgvector` (Required for RAG)
- **Size:** Development (1 vCPU, 1 GB RAM, 10 GB Disk) - $15/mo
- **Trusted Sources:** Only allow connections from App Platform services and Architect's IP.

## 3. Managed Redis

- **Version:** 7.0+
- **Size:** Development (1 vCPU, 1 GB RAM) - $15/mo
- **Role:** Backend for Asynq task queue.

## 4. Object Storage (Digital Ocean Spaces)

- **Region:** SGP1 (or same as app)
- **CORS:** Allowed for staging domain.
- **Role:** PDF Storage (Blueprints, Invoices) and Site Photos.

## 5. Environment Variables (Identity & Access)

| Key | Description | Source |
|-----|-------------|--------|
| `DATABASE_URL` | Postgres connection string | DO Database Cluster |
| `REDIS_URL` | Redis connection string | DO Redis Cluster |
| `JWT_SECRET` | Secret for auth tokens | Manual Secret |
| `GEMINI_API_KEY` | Google Vertex AI / AI Studio Key | Manual |
| `SPACES_KEY` | DO Spaces Access Key | IAM |
| `SPACES_SECRET` | DO Spaces Secret Key | IAM |
| `SPACES_ENDPOINT` | S3-compatible endpoint URL | DO Spaces |
| `SPACES_BUCKET` | Storage bucket name | DO Spaces |
| `STRIPE_SECRET_KEY` | Stripe API secret key | Manual Secret |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret | Manual Secret |
| `APP_ENV` | Environment name (development/staging/production) | Manual |
| `APP_PORT` | HTTP server port | Manual |
| `APP_URL` | Public-facing application URL | Manual |

---
*Verified against BACKEND_SCOPE.md Section 1.*
