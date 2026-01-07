# Environment Variable Management Strategy

// See PRODUCTION_PLAN.md Step 6

## Overview

This document defines how secrets and configuration are managed across all FutureBuild environments: **local development**, **staging**, and **production**.

---

## 1. Variable Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | ✅ | PostgreSQL connection string |
| `REDIS_URL` | ✅ | Redis connection string (Asynq) |
| `JWT_SECRET` | ✅ | HMAC signing key for auth tokens (min 32 chars) |
| `GEMINI_API_KEY` | ✅ | Google AI API key for Vision/RAG |
| `SPACES_KEY` | ✅ | DO Spaces / S3 access key |
| `SPACES_SECRET` | ✅ | DO Spaces / S3 secret key |
| `SPACES_ENDPOINT` | ✅ | S3-compatible endpoint URL |
| `SPACES_BUCKET` | ✅ | Storage bucket name |
| `STRIPE_SECRET_KEY` | ✅ | Stripe API secret key |
| `STRIPE_WEBHOOK_SECRET` | ✅ | Stripe webhook signing secret |
| `APP_ENV` | ✅ | Environment name (development/staging/production) |
| `APP_PORT` | ✅ | HTTP server port |
| `APP_URL` | ✅ | Public-facing application URL |

---

## 2. Injection Strategy by Environment

### 2.1 Local Development (Docker Compose)

**Method:** `.env` file in repository root.

```bash
# Copy the template
cp .env.example .env

# Edit with real values
nano .env
```

Docker Compose automatically loads `.env` via the `env_file` directive:

```yaml
# docker-compose.yml
services:
  api:
    env_file:
      - .env
```

> ⚠️ **Important:** `.env` is gitignored. Never commit real secrets.

---

### 2.2 Staging (Digital Ocean App Platform)

**Method:** App Platform Environment Variables UI or `app.yaml` spec.

1. Navigate to **App Platform → futurebuild-staging → Settings → App-Level Environment Variables**.
2. Add each variable from Section 1.
3. Mark sensitive values (JWT_SECRET, *_KEY, *_SECRET) as **Encrypted**.

Alternatively, define in `app.yaml`:

```yaml
envs:
  - key: DATABASE_URL
    scope: RUN_TIME
    value: ${db.DATABASE_URL}  # Injected from managed DB
  - key: JWT_SECRET
    scope: RUN_TIME
    type: SECRET
    value: EV[1:encrypted-value]
```

---

### 2.3 Production

**Method:** Same as staging, with dedicated production App Platform instance.

| Difference from Staging | Notes |
|------------------------|-------|
| `APP_ENV=production` | Enables production logging/security |
| `APP_URL` | Points to production domain |
| Managed DB/Redis | Production-tier clusters |
| Stripe keys | Live keys (`sk_live_*`) |

---

## 3. Security Guidelines

1. **Never commit secrets** — All `.env*` files (except `.env.example`) are gitignored.
2. **Rotate secrets regularly** — Especially JWT_SECRET and Stripe keys.
3. **Use encrypted secrets** — In DO App Platform, always mark sensitive vars as "Encrypted".
4. **Principle of least privilege** — Use scoped API keys where possible.
5. **Audit access** — Restrict who can view/edit environment variables in DO dashboard.

---

## 4. Verification Checklist

Before deploying to any environment, confirm:

- [ ] All required variables from Section 1 are set
- [ ] `APP_ENV` matches the target environment
- [ ] Database/Redis URLs point to correct clusters
- [ ] Stripe keys match environment (test vs live)
- [ ] Object storage bucket exists and is accessible

---

*Verified against DIGITAL_OCEAN_STAGING.md and BACKEND_SCOPE.md.*
