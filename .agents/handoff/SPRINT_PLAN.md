# Sprint Plan

**System:** FutureBuild Brain (System of Connection)
**Pipeline Stage:** 08 - Implementation Plan
**Date:** 2026-04-02
**Status:** COMPLETE
**Sprint Duration:** 2 weeks
**Total Sprints:** 6 (S0–S5)
**Estimated Timeline:** 12 weeks

---

## Cross-System Dependency Map

```
FB-Brain Sprint 0 ──► FB-OS Sprint 0
  (OIDC issuer)        (JWT middleware — BLOCKED until Brain OIDC is live)

FB-Brain Sprint 1 ──► FB-Brain Sprint 2
  (MCP Registry)        (Maestro needs registry to call tools)

FB-Brain Sprint 2 ──► FB-Brain Sprint 3
  (Maestro flows)        (A2A emits events from Maestro flow results)

FB-Brain Sprint 3 ──► FB-OS Sprint 5
  (A2A Client)           (OS webhook receiver — BLOCKED until Brain emits)

FB-Brain Sprint 4 ──► FB-Brain Sprint 5
  (All MCP servers)      (UI needs live data from integrations)

All Brain backend ──► FB-Brain Sprint 5
  (Sprints 0-4)          (Lit Admin UI requires backend endpoints)
```

---

## Sprint 0: Walking Skeleton (Weeks 1–2)

**Goal:** FB-Brain issues valid JWTs that FB-OS can validate. End-to-end OIDC flow works.

### Deliverables

| # | Task | Package | Priority |
|---|------|---------|----------|
| 1 | PostgreSQL schema: `organizations`, `hub_users`, `org_memberships` | `migrations/001_initial_schema.sql` | P0 |
| 2 | PostgreSQL schema: `oidc_clients`, `oidc_auth_requests`, `oidc_refresh_tokens`, `oidc_signing_keys` | `migrations/002_oidc_tables.sql` | P0 |
| 3 | RSA key pair generation and storage | `internal/oidc/keys.go` | P0 |
| 4 | OIDC Provider: `/.well-known/openid-configuration` | `internal/oidc/provider.go` | P0 |
| 5 | OIDC Provider: `/authorize` (authorization code flow) | `internal/oidc/provider.go` | P0 |
| 6 | OIDC Provider: `/token` (code exchange + PKCE verification) | `internal/oidc/provider.go` | P0 |
| 7 | OIDC Provider: `/jwks` (public key endpoint) | `internal/oidc/provider.go` | P0 |
| 8 | OIDC Provider: `/userinfo` | `internal/oidc/provider.go` | P0 |
| 9 | Custom claims: `org_id`, `role`, `plan_tier` in access + ID tokens | `internal/oidc/claims.go` | P0 |
| 10 | OIDC Storage adapter (PostgreSQL-backed) | `internal/oidc/storage.go` | P0 |
| 11 | Magic link authentication (send email, verify token) | `internal/oidc/provider.go` | P0 |
| 12 | Chi router + middleware stack (CORS, telemetry, rate limiting) | `internal/api/router.go`, `internal/api/middleware/` | P0 |
| 13 | Docker Compose: PostgreSQL 16 | `docker-compose.yml` | P0 |
| 14 | CI pipeline: go test, golangci-lint, SQL migration linter | `.github/workflows/ci.yml`, `scripts/lint-migrations.sh` | P0 |
| 15 | Seed data: register `fb-os` as OIDC client | `migrations/` or seed script | P0 |

### Exit Criteria

- `GET /.well-known/openid-configuration` returns valid discovery document
- `GET /jwks` returns RSA public key
- Full authorization code flow with PKCE completes: `/authorize` → magic link → `/token` → JWT
- JWT contains `sub`, `org_id`, `role`, `plan_tier` claims
- FB-OS can validate JWT using `/jwks` endpoint (tested via curl)
- CI passes: SQL linter, go test, golangci-lint

### Blocking

- **BLOCKS FB-OS Sprint 0:** OS JWT middleware cannot be tested until Brain issues tokens
- **Mitigation:** Brain deploys OIDC endpoints by end of Week 1; OS connects in Week 2

---

## Sprint 1: MCP Server Registry + OIDC Completion (Weeks 3–4)

**Goal:** MCP Registry is operational. OIDC has full token lifecycle management.

### Deliverables

| # | Task | Package | Priority |
|---|------|---------|----------|
| 1 | OIDC: `/revoke` endpoint (RFC 7009) | `internal/oidc/provider.go` | P0 |
| 2 | OIDC: `/consent` screen (GET/POST) | `internal/oidc/consent.go` | P0 |
| 3 | OIDC: Refresh token rotation | `internal/oidc/storage.go` | P0 |
| 4 | PostgreSQL schema: `mcp_servers`, `mcp_tools`, `mcp_triggers`, `integration_credentials` | `migrations/003_mcp_registry.sql` | P0 |
| 5 | MCPRegistry service: RegisterServer, ListServers, GetServer, ListTools, CallTool, HealthCheck | `internal/mcp/registry.go` | P0 |
| 6 | MCP HTTP handlers: `POST /mcp/tools/list`, `POST /mcp/tools/call` | `internal/mcp/handler.go` | P0 |
| 7 | JSON Schema 2020-12 input validation | `internal/mcp/schema.go` | P1 |
| 8 | Per-server health check runner | `internal/mcp/health.go` | P1 |
| 9 | GableERP MCP server implementation (4 tools, 2 triggers) | `internal/clients/gable.go` | P0 |
| 10 | Hub API: `GET/POST/PUT/DELETE /api/hub/registry/servers` | `internal/hub/registry_handlers.go` | P1 |
| 11 | Hub API: `GET/POST/PUT/DELETE /api/hub/oidc/clients` | `internal/hub/oidc_handlers.go` | P1 |

### Exit Criteria

- Token revocation works; revoked tokens rejected by `/userinfo`
- Consent screen renders, approves, and issues auth code
- GableERP `get_products_by_category` tool call returns data via MCP protocol
- JSON Schema validation rejects invalid tool inputs
- Health check reports GableERP status

### Blocking

- **BLOCKS Sprint 2:** Maestro Orchestrator calls `MCPRegistry.CallTool()`

---

## Sprint 2: Maestro Orchestrator (Weeks 5–6)

**Goal:** Natural language input produces cross-system tool calls and structured responses.

### Deliverables

| # | Task | Package | Priority |
|---|------|---------|----------|
| 1 | PostgreSQL schema: `maestro_sessions`, `maestro_messages` | `migrations/004_maestro.sql` | P0 |
| 2 | Maestro cost tracking: `cost_cents` + `cost_currency_code` (Composite Currency Pattern) | `internal/maestro/orchestrator.go` | P0 |
| 3 | Three-tier intent classifier: regex fast-path | `internal/maestro/classifier.go` | P0 |
| 4 | Three-tier intent classifier: Claude tool-use (Anthropic API) | `internal/maestro/classifier.go` | P0 |
| 5 | Three-tier intent classifier: human confirmation fallback | `internal/maestro/classifier.go` | P1 |
| 6 | Intent definitions (top-10 MVP intents) | `internal/maestro/intents.go` | P0 |
| 7 | MaterialsFlow coordinator (scope → products → pricing → quote → RFQ → A2A emit) | `internal/maestro/flows/materials.go` | P0 |
| 8 | LaborFlow coordinator (scope → RFQ → bid → approval) | `internal/maestro/flows/labor.go` | P0 |
| 9 | PostApproval convergence check | `internal/maestro/flows/post_approval.go` | P1 |
| 10 | Chat session management | `internal/maestro/chat.go` | P1 |
| 11 | Hub API: `POST /api/hub/chat/message`, `GET /api/hub/chat/sessions` | `internal/hub/` | P0 |

### Exit Criteria

- "Order lumber for project X" classified as `materials_procurement` with >0.90 confidence
- MaterialsFlow calls GableERP tools via MCP and returns quote data
- Chat session persists across multiple messages
- Cost per intent tracked in `maestro_messages.cost_cents`
- Regex fast-path handles top-3 intents in <1ms

### Blocking

- **BLOCKS Sprint 3:** A2A webhook emission is triggered by Maestro flow results

---

## Sprint 3: A2A Webhooks + RFQ State Machine (Weeks 7–8)

**Goal:** Brain emits JWS-signed webhooks to FB-OS. Full RFQ lifecycle tracked.

### Deliverables

| # | Task | Package | Priority |
|---|------|---------|----------|
| 1 | PostgreSQL schema: `rfqs` (with `total_cents` + `currency_code`), `integration_events`, `a2a_webhook_log` | `migrations/005_a2a.sql` | P0 |
| 2 | A2AClient: JWS RS256 signing via go-jose/v4 | `internal/a2a/signer.go` | P0 |
| 3 | A2AClient: Emit method with detached signature + idempotency key | `internal/a2a/client.go` | P0 |
| 4 | A2A retry strategy: exponential backoff (30s→1hr, 7 attempts, dead letter) | `internal/a2a/retry.go` | P0 |
| 5 | Webhook event schemas: `review_material_quote` (with `currency_code`) | `internal/a2a/types.go` | P0 |
| 6 | Webhook event schemas: `review_labor_bid` (with `currency_code`) | `internal/a2a/types.go` | P0 |
| 7 | Webhook event schemas: `update_schedule`, `delivery_confirmation`, `create_feed_card` | `internal/a2a/types.go` | P0 |
| 8 | RFQ state machine (materials: pending→quote_received→approved→ordered→delivered) | `internal/models/rfq.go`, `internal/store/rfq.go` | P0 |
| 9 | RFQ state machine (labor: pending→rfq_sent→bid_received→approved→assigned) | `internal/models/rfq.go`, `internal/store/rfq.go` | P0 |
| 10 | Integration event audit trail logging | `internal/store/integration_event.go` | P1 |
| 11 | Connect MaterialsFlow/LaborFlow → A2A emission on quote/bid completion | `internal/maestro/flows/` | P0 |

### Exit Criteria

- `review_material_quote` webhook delivered to FB-OS with valid JWS signature and `currency_code: "USD"`
- FB-OS verifies signature using Brain's `/jwks` public key
- Failed delivery retries with exponential backoff; after 7 failures moves to dead letter
- RFQ status transitions are atomic and logged in `integration_events`
- `a2a_webhook_log` records delivery status, latency, retry count

### Blocking

- **BLOCKS FB-OS Sprint 5:** OS A2A receiver endpoint depends on Brain emitting valid webhooks
- **Mitigation:** Brain A2A client tested with mock OS endpoint in Sprint 3; OS implements receiver in Sprint 5

---

## Sprint 4: Remaining MCP Servers + Integration Connections (Weeks 9–10)

**Goal:** All 7 MCP servers operational. External OAuth2 flows for QuickBooks/Gmail/Outlook.

### Deliverables

| # | Task | Package | Priority |
|---|------|---------|----------|
| 1 | LocalBlue MCP server (2 tools, 1 trigger) | `internal/clients/localblue.go` | P0 |
| 2 | XUI Projects MCP server (2 tools, 3 triggers) | `internal/clients/xui.go` | P0 |
| 3 | QuickBooks MCP server (4 tools, 2 triggers + CloudEvents) | `internal/clients/quickbooks.go` | P0 |
| 4 | 1Build MCP server (3 tools) | `internal/clients/onebuild.go` | P1 |
| 5 | Gmail MCP server (3 tools, 1 trigger) | `internal/clients/gmail.go` | P1 |
| 6 | Outlook MCP server (3 tools, 1 trigger) | `internal/clients/outlook.go` | P1 |
| 7 | OAuth2 external flows: `/api/hub/oauth/{provider}/authorize` + `/callback` | `internal/hub/oauth_handlers.go` | P0 |
| 8 | Integration credential management (AES-256-GCM encryption) | `internal/store/` | P0 |
| 9 | Hub API: `GET/POST/DELETE /api/hub/connections` | `internal/hub/integration_handlers.go` | P0 |
| 10 | SSE event streaming: `GET /api/hub/events/stream` | `internal/hub/event_bus.go` | P1 |
| 11 | Hub dashboard endpoint: `GET /api/hub/dashboard` | `internal/hub/` | P1 |
| 12 | MCP trigger subscription: `POST /mcp/triggers/subscribe` | `internal/mcp/handler.go` | P2 |

### Exit Criteria

- All 7 MCP servers respond to `tools/call` via registry
- QuickBooks OAuth2 flow completes: authorize → callback → encrypted token stored
- SSE stream emits `integration_event`, `webhook_delivered`, `health_check` events
- Dashboard returns connected integrations count and recent events

### Blocking

- **BLOCKS Sprint 5:** Lit Admin UI needs all backend endpoints to render live data

---

## Sprint 5: Lit Admin UI (Weeks 11–12)

**Goal:** Full admin UI for Brain. All pages functional with live backend data.

### Deliverables

| # | Task | Package | Priority |
|---|------|---------|----------|
| 1 | FBBaseElement setup (glassmorphism, glow, skeleton utilities) | `frontend/src/components/base/fb-element.ts` | P0 |
| 2 | GableLBM design tokens (CSS custom properties) | `frontend/src/styles/variables.css` | P0 |
| 3 | Shared atom components (fb-button, fb-icon, fb-badge, fb-input, fb-spinner, fb-avatar) | `frontend/src/components/atoms/` | P0 |
| 4 | fb-login-page (magic link email input, SSO redirect) | `frontend/src/components/pages/fb-login-page.ts` | P0 |
| 5 | fb-home-page (split: activity ticker + maestro chat) | `frontend/src/components/pages/fb-home-page.ts` | P0 |
| 6 | fb-maestro-chat (chat bubbles, execution rows) | `frontend/src/components/organisms/fb-maestro-chat.ts` | P0 |
| 7 | fb-maestro-drawer (global FAB, slide-out panel) | `frontend/src/components/organisms/fb-maestro-drawer.ts` | P0 |
| 8 | fb-ecosystem-page (D3 force-directed graph) | `frontend/src/components/pages/fb-ecosystem-page.ts` | P1 |
| 9 | fb-admin-registry (MCP server table + tool cards) | `frontend/src/components/pages/fb-admin-registry.ts` | P0 |
| 10 | fb-marketplace-page | `frontend/src/components/pages/fb-marketplace-page.ts` | P2 |
| 11 | fb-settings-page (OIDC manager, org settings) | `frontend/src/components/pages/fb-settings-page.ts` | P1 |
| 12 | Client-side router | `frontend/src/router.ts` | P0 |
| 13 | Signals-based state management | `frontend/src/state/` | P0 |
| 14 | Vite build + Lighthouse CI (>90 perf, >95 a11y, <50KB gzip) | `frontend/vite.config.ts` | P0 |

### Exit Criteria

- Magic link login flow completes end-to-end in browser
- Maestro chat sends message, receives AI response, shows MCP tool call results
- MCP registry page lists servers with health indicators
- Ecosystem canvas renders Brain + OS + integrations as force-directed graph
- Lighthouse: Performance >90, Accessibility >95, bundle <50KB gzip
- All numerical data rendered in JetBrains Mono

---

## Sprint Summary

| Sprint | Weeks | Focus | Deliverables | Blocks |
|--------|-------|-------|-------------|--------|
| S0 | 1–2 | Walking Skeleton | OIDC Provider (5 endpoints), PostgreSQL schema, Chi router, CI | FB-OS Sprint 0 |
| S1 | 3–4 | MCP Registry + OIDC | MCP tables, registry service, GableERP client, consent/revoke | Sprint 2 |
| S2 | 5–6 | Maestro Orchestrator | 3-tier classifier, MaterialsFlow, LaborFlow, chat API | Sprint 3 |
| S3 | 7–8 | A2A Webhooks + RFQ | JWS signing, 5 event types, retry, RFQ state machine | FB-OS Sprint 5 |
| S4 | 9–10 | MCP Servers + Integrations | 6 remaining servers, OAuth2 flows, SSE, dashboard | Sprint 5 |
| S5 | 11–12 | Lit Admin UI | 6 pages, Maestro drawer, ecosystem canvas, login | — |

### Velocity Assumptions

- 2 backend engineers + 1 frontend engineer
- Sprint 0–4: backend-heavy (2 engineers full-time)
- Sprint 5: frontend-heavy (1 FE + 1 BE for API polish)
- Each sprint includes: code, unit tests, integration tests, documentation
