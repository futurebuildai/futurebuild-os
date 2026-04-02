# Vision Brief: FutureBuild Brain (System of Connection) Revamp

**Document ID:** AG-00-VISION
**Created:** 2026-04-02
**Status:** Stage 00 Complete
**Author:** Antigravity Orchestrator

---

## 1. Executive Summary

FutureBuild Brain is being revamped from a construction-industry integration orchestrator into the ecosystem's **central Connection Plane**. The revamp evolves -- not replaces -- the existing registry, orchestrator, and hub patterns into three new architectural pillars:

1. **OIDC Identity Provider** -- Brain becomes the single identity authority for the entire FutureBuild ecosystem, issuing JWTs consumed by FB-OS and all future services.
2. **MCP Server Registry** -- The legacy `registry.go` pattern (SystemDefinition/ActionDefinition/TriggerDefinition) evolves into a standards-compliant Model Context Protocol server registry with dynamic tool discovery.
3. **Maestro AI Co-pilot** -- A probabilistic intent parser that replaces the hardcoded MaterialsFlow/LaborFlow orchestration with AI-driven routing, emitting signed A2A webhooks to trigger deterministic execution in FB-OS.

All three pillars operate behind **L7 Zero-Trust boundaries** (SPIFFE workload identity, mTLS, per-request authorization).

---

## 2. Legacy Foundation

### 2.1 The Seven Existing Integrations

From `reference-vault/FB-Brain/internal/registry/registry.go`, the current registry hardcodes seven systems:

| # | Slug | Name | Auth Type | Actions | Triggers | Entities |
|---|------|------|-----------|---------|----------|----------|
| 1 | `gable` | GableERP | api_key | 4 (get_products, bulk_price, create_quote, accept_quote) | 2 (price_updated, order_status_changed) | 2 |
| 2 | `xui` | XUI Projects | api_key | 2 (create_feed_card, assign_contact_to_phase) | 3 (card_action_approve/reject/request_change) | 2 |
| 3 | `localblue` | LocalBlue | host_header | 2 (push_rfq, update_bid_status) | 1 (bid_submitted) | 2 |
| 4 | `quickbooks` | QuickBooks Online | oauth2 | 4 (create_invoice, get_customers, create_bill, create_purchase_order) | 2 (invoice_created, payment_received) | 4 |
| 5 | `onebuild` | 1Build API | api_key | 3 (get_estimate, search_cost_data, get_line_items) | 1 (estimate_completed) | 3 |
| 6 | `gmail` | Gmail | oauth2 | 3 (send_email, list_emails, get_email) | 1 (email_received) | 3 |
| 7 | `outlook` | Outlook 365 | oauth2 | 3 (send_email, list_emails, get_email) | 1 (email_received) | 2 |

**Total:** 21 actions, 11 triggers, 18 entity schemas across 7 systems.

### 2.2 MaterialsFlow/LaborFlow Orchestration Patterns

From `reference-vault/FB-Brain/internal/orchestrator/materials_flow.go` and `labor_flow.go`:

**MaterialsFlow** (9-step deterministic pipeline):
1. Look up AccountLink to resolve GableERP customer ID
2. Get products from GableERP by category ("Roofing")
3. Build product map by SKU
4. Build bulk price request from hardcoded roofing scope
5. Get prices from GableERP
6. Create quote in GableERP
7. Store RFQ record in PostgreSQL
8. Create review card in XUI Projects
9. Log IntegrationEvent

**LaborFlow** (6-step deterministic pipeline):
1. Look up AccountLink for LocalBlue site ID
2. Build scope items (hardcoded roofing)
3. Create placeholder RFQ
4. Push RFQ to LocalBlue
5. Update RFQ with LocalBlue reference
6. Log IntegrationEvent

**Post-Approval** (`post_approval.go`): Checks if both materials ordered AND labor approved, then creates delivery confirmation card in XUI.

**Critical observation:** These flows are hardcoded for a single demo scenario (roofing on "Riverside New Home"). The revamp must make them dynamic and AI-driven.

### 2.3 Hub Admin UI Architecture

From `reference-vault/FB-Brain/internal/hub/hub.go`:

The Hub is a Chi-based REST API with session cookie auth, organized into:
- **Auth:** Login, logout, magic link, SSO exchange
- **Organizations:** CRUD, member management, invites
- **Platforms:** CRUD with connectivity testing
- **Connections:** Direct connect between platforms
- **Templates & Integrations:** Plan-gated workflow templates
- **Engine:** Workflow validation, dry-run, execution (via `WorkflowExecutor`)
- **Registry:** HTTP handlers exposing SystemDefinition to the frontend
- **Chat:** AI assistant (Claude-powered, plan-gated)
- **OAuth:** Email integration flows (Gmail, Outlook)
- **SSE:** Real-time event streaming

Frontend: React 19 + TypeScript + Vite + TanStack Query + XY Flow canvas visualization + Radix UI + Tailwind CSS 4.

### 2.4 What Must Be Preserved vs. Evolved

| Aspect | Preserve | Evolve |
|--------|----------|--------|
| Registry data model | SystemDefinition concept (slug, auth, actions, triggers, entities) | Evolve into MCP Tool/Resource definitions with JSON Schema 2020-12 input/output schemas |
| 7 integrations | All integration definitions and their API contracts | Add dynamic registration; move from compile-time to runtime |
| AccountLink pattern | Cross-system identity mapping (XUI org -> Gable customer -> LocalBlue site) | Expand to universal identity mapping via OIDC subject claims |
| IntegrationEvent audit trail | Event logging with source/target/payload | Evolve to OpenTelemetry-instrumented, signed event chain |
| Hub admin routes | Core CRUD patterns for orgs, platforms, connections | Rebuild UI in Lit Web Components; add MCP registry browser |
| Flow orchestration | The concept of multi-system workflows | Replace hardcoded flows with Maestro AI probabilistic routing + A2A webhooks |
| Session auth | Session-based auth for admin UI | Replace with OIDC-issued JWT tokens; Brain becomes the IdP |
| Chi router | HTTP framework | Keep Chi; add OIDC endpoints and MCP transport handlers |
| PostgreSQL | Data store | Keep pgx; add OIDC tables (clients, grants, keys, sessions) |

---

## 3. Vision Statement

> FutureBuild Brain is the ecosystem's Connection Plane: the OIDC Identity Provider that authenticates every user and service, the MCP Server Registry that exposes every integration as a discoverable tool, and the Maestro AI Co-pilot that understands natural-language intent and orchestrates cross-system workflows via signed A2A webhooks -- all behind L7 Zero-Trust boundaries.

---

## 4. Architectural Pillars

### Pillar 1: OIDC Identity Provider
- Brain issues JWTs (access + refresh + ID tokens) via standard OIDC flows
- `/.well-known/openid-configuration` discovery endpoint
- `/authorize`, `/token`, `/userinfo`, `/jwks` endpoints
- FB-OS and all ecosystem services delegate auth to Brain
- Replaces Clerk dependency in FB-OS with Brain-issued tokens

### Pillar 2: MCP Server Registry
- Evolve SystemDefinition -> MCP Tool definitions with JSON Schema
- Dynamic registration via admin API (not just compile-time)
- `tools/list`, `tools/call` MCP transport handlers
- Resources and Prompts primitives for context injection
- Official Go SDK (`modelcontextprotocol/go-sdk`) for implementation

### Pillar 3: Maestro AI Co-pilot
- Probabilistic intent classification (replaces hardcoded flow routing)
- Natural-language interface for construction workflows
- Emits signed A2A webhooks to FB-OS for deterministic execution
- Agent Card per integration with JWS-signed payloads (RFC 8785 canonical JSON)
- Hybrid architecture: probabilistic intelligence inside deterministic guardrails

### Cross-cutting: L7 Zero-Trust
- SPIFFE/SPIRE workload identity for service-to-service auth
- mTLS on all internal connections
- Per-request authorization (no implicit trust from network location)
- NIST 800-207 aligned architecture

---

## 5. Non-Goals (Explicit Exclusions)

1. Brain does NOT execute business logic -- it routes to FB-OS/GableERP/LocalBlue for execution
2. Brain does NOT replace individual system UIs -- it provides the admin/connection plane
3. Brain does NOT store construction project data -- that lives in FB-OS
4. Brain does NOT implement its own user database from scratch -- it evolves the existing hub_users table into an OIDC-compliant identity store
5. Brain does NOT become a message queue -- A2A webhooks are fire-and-forget with retry

---

## 6. Success Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| OIDC compliance | OpenID Certified basic profile | OIDC conformance test suite |
| MCP server count | 7+ servers (matching legacy integrations) | Registry API count |
| Intent classification accuracy | >90% on top-20 construction intents | Evaluation set of 200 utterances |
| A2A webhook delivery | <500ms p99 latency, 99.9% delivery rate | Prometheus metrics |
| Zero-Trust coverage | 100% service-to-service mTLS | SPIRE attestation audit |
| Admin UI Lighthouse score | >90 across all categories | Lighthouse CI |

---

## 7. Key Stakeholders

| Role | Concern |
|------|---------|
| Construction PM (end user) | Must be able to say "order roofing materials for Project X" and have it work |
| FB-OS (downstream system) | Must be able to validate Brain-issued JWTs and receive A2A webhooks |
| GableERP/LocalBlue (integration targets) | Must continue working with existing API contracts |
| DevOps | Must be able to deploy Brain independently of FB-OS |
| Security | Must pass NIST 800-207 Zero-Trust audit |
