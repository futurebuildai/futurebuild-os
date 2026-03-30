# Production Gap Patch Specification

## 1. Overview
This specification details the remediation tasks required to secure and productionize the FutureBuild OS and FB-Brain repositories prior to Phase 18 deployment. The focus is on container security, CI/CD parity, and cross-boundary distributed tracing.

## 2. Components & Requirements

### 2.1 FB-Brain Dockerfile (Security)
**Target:** `FB-Brain/Dockerfile`
**Problem:** The Runtime stage (Stage 3) executes as the default `root` user.
**Requirements:**
- Add an explicit user/group creation in the `alpine` stage: `RUN addgroup -S appgroup && adduser -S appuser -G appgroup`.
- Set ownership of the working directory: `RUN chown -R appuser:appgroup /app`.
- Enforce the non-root execution policy via `USER appuser` immediately before the `EXPOSE` and `CMD` directives.

### 2.2 FB-Brain GitHub Actions (CI/CD Parity)
**Targets:** `FB-Brain/.github/workflows/deploy-prod.yaml` & `deploy-demo.yaml`
**Problem:** FB-Brain lacks automated deployment triggers for semantic versions and branch pushes to staging/production.
**Requirements:**
- Copy `deploy-prod.yaml` and `deploy-demo.yaml` from `futurebuild-repo/.github/workflows/` to `FB-Brain/.github/workflows/`.
- Update repository/image targets in the copied files from `futurebuild-repo` to `fb-brain`.
- Ensure the ECR repository variables explicitly point to the `fb-brain` registry.

### 2.3 Cross-Boundary Observability (Tracing)
**Targets:** 
- `futurebuild-repo/internal/api/handlers/integration_client.go`
- `FB-Brain/internal/api/engine_webhook_handler.go`
**Problem:** A2A webhooks lack a unified tracing ID, obscuring distributed logs.
**Requirements:**
- **FutureBuild OS:** In `integration_client.go`, generate a W3C-compliant `traceparent` (or a simple UUIDv4 `X-Trace-ID` if OTel is unavailable) and attach it to the outbound `http.Request` headers.
- **FB-Brain:** In `engine_webhook_handler.go`, extract the `X-Trace-ID` header. Include this ID context within the `slog` fields (e.g., `slog.Info(..., "trace_id", traceID)`). Ensure the `X-Trace-ID` is safely stored in the `integration_events` database schema payload if possible, or minimally logged uniformly.

### 2.4 End-to-End Orchestration Stub
**Target:** `dev_ecosystem/e2e/`
**Problem:** No cross-repository integration test harness exists.
**Requirements:**
- Initialize a basic Playwright testing directory at the root of `dev_ecosystem` (`mkdir e2e`).
- Initialize an empty `package.json` with Playwright dependencies via `npm init playwright@latest` (non-interactive).
- This establishes the foundation for the upcoming QA Automation sprint.
