# Scope Definition

**Document ID:** AG-04-SCOPE
**Created:** 2026-04-02
**Status:** Stage 04 Complete -- PAUSED AT APPROVAL GATE 1

---

## 1. Walking Skeleton (Week 1-2)

The walking skeleton proves the architectural concept end-to-end with minimal functionality:

```
User authenticates with Brain OIDC
  -> Brain issues JWT
  -> User calls Maestro: "list products in GableERP"
  -> Maestro discovers GableERP MCP tool: get_products_by_category
  -> Maestro calls tool, gets products
  -> Maestro returns result to user
  -> Brain emits A2A webhook to FB-OS with event log
```

### Walking Skeleton Components

| Component | Scope | Done When |
|-----------|-------|-----------|
| OIDC Provider (minimal) | `/.well-known/openid-configuration`, `/authorize`, `/token`, `/jwks` with a single test client | FB-OS can validate a Brain-issued JWT |
| MCP Registry (minimal) | Single MCP server (GableERP) with `get_products_by_category` tool registered in PostgreSQL | `tools/list` returns GableERP tools; `tools/call` executes against GableERP API |
| Maestro (minimal) | Regex-only classification for "list products" -> GableERP tool call | No Claude dependency for skeleton; deterministic path |
| A2A Client (minimal) | Emit unsigned HTTP POST to FB-OS webhook endpoint | FB-OS receives webhook payload (signing added in MVP) |
| Admin UI (minimal) | Static page listing registered MCP servers (Lit) | Page renders at `/admin/registry` |

---

## 2. MVP Definition (Week 3-12)

### MVP Scope: "Brain as Connection Plane for Roofing Workflow"

The MVP recreates the existing MaterialsFlow + LaborFlow functionality using the new architecture (OIDC + MCP + Maestro + A2A), proving that the revamp is functionally equivalent to the legacy system and then exceeds it with AI-driven interaction and QuickBooks automation.

### MVP Features

#### P0: Must Have (Ship-blocking)

| Feature | Description | Legacy Reference |
|---------|-------------|-----------------|
| OIDC Provider | Full authorization code flow with PKCE; JWKS endpoint; refresh token rotation; consent screen | Evolves `hub/auth_handlers.go` session auth |
| OIDC Client Management | Admin CRUD for OIDC clients (FB-OS, future services) | New capability |
| MCP GableERP Server | All 4 actions + 2 triggers as MCP tools: get_products, bulk_price, create_quote, accept_quote, price_updated, order_status_changed | Evolves `registry/gable.go` |
| MCP LocalBlue Server | All 2 actions + 1 trigger: push_rfq, update_bid_status, bid_submitted | Evolves `registry/localblue.go` |
| MCP XUI Server | All 2 actions + 3 triggers: create_feed_card, assign_contact, card_action_* | Evolves `registry/xui.go` |
| MCP QuickBooks Server | All 4 actions + 2 triggers with CloudEvents webhook support | Evolves `registry/quickbooks.go`; addresses May 2026 CloudEvents deadline |
| Maestro Intent Classification | Top-10 construction intents via Claude tool-use with regex fallback | Evolves `orchestrator/materials_flow.go` + `labor_flow.go` |
| Maestro Materials Flow | AI-driven materials procurement (replaces hardcoded 9-step pipeline) | Evolves `orchestrator/materials_flow.go` |
| Maestro Labor Flow | AI-driven subcontractor bidding (replaces hardcoded 6-step pipeline) | Evolves `orchestrator/labor_flow.go` |
| A2A Signed Webhooks | JWS-signed webhook emission to FB-OS with retry | Replaces direct XUI API calls |
| Post-Approval Logic | Convergence check (materials + labor -> delivery confirmation) via A2A | Evolves `orchestrator/post_approval.go` |
| QuickBooks PO Automation | Automatic PO creation in QuickBooks when materials quote is approved | New capability (Job 4 from JTBD) |

#### P1: Should Have (Week 10-12)

| Feature | Description | Legacy Reference |
|---------|-------------|-----------------|
| MCP 1Build Server | 3 actions: get_estimate, search_cost_data, get_line_items | Evolves `registry/onebuild.go` |
| MCP Gmail Server | 3 actions + 1 trigger: send_email, list_emails, get_email, email_received | Evolves `registry/gmail.go` |
| MCP Outlook Server | 3 actions + 1 trigger: send_email, list_emails, get_email, email_received | Evolves `registry/outlook.go` |
| Dynamic MCP Server Registration | Admin API for registering new MCP servers without code deployment | New capability |
| Lit MCP Registry Browser | Admin dashboard page showing registered servers, tools, health status | Replaces React registry_handlers UI |
| Lit OIDC Client Manager | Admin dashboard page for managing OIDC clients | New capability |
| Maestro Top-20 Intents | Expand from 10 to 20 construction intents | New capability |

#### P2: Could Have (Post-MVP)

| Feature | Description |
|---------|-------------|
| A2A Agent Cards | Full A2A Agent Card with JWS signatures and RFC 8785 canonical JSON |
| SPIFFE/SPIRE Integration | Workload identity for Brain <-> FB-OS mTLS |
| Maestro Bid Comparison | AI-powered bid analysis with market rate comparison via 1Build |
| Lit Migration (existing pages) | Migrate Ecosystem, Marketplace, Login pages from React to Lit |
| Multi-org OIDC | Support switching between orgs with different OIDC scopes/claims |
| Webhook Dashboard | Real-time view of A2A webhook delivery status, retries, failures |
| MCP Prompts | MCP Prompt primitives for context-aware AI responses |

#### P3: Won't Have (This Release)

| Feature | Rationale |
|---------|-----------|
| Full Zitadel/Hydra deployment | Custom OIDC with zitadel/oidc lib is sufficient |
| Fine-tuned intent classifier | Not enough production data yet |
| Mobile app | Web-first; mobile is a separate product decision |
| Municipal portal MCP server | Requires partner API access; deferred to Phase 2 |
| Istio service mesh | Overkill for two-service architecture |

---

## 3. MVP Phased Roadmap

### Phase 1: Foundation (Weeks 1-4)

**Theme:** Identity and Registry

| Week | Deliverable |
|------|-------------|
| 1 | Walking skeleton: minimal OIDC + minimal MCP + minimal Maestro |
| 2 | Full OIDC provider: authorization code flow, PKCE, JWKS, refresh tokens, consent screen |
| 3 | OIDC client management; FB-OS integration testing (replace Clerk with Brain tokens) |
| 4 | MCP Server Registry: PostgreSQL schema, admin CRUD API, tools/list endpoint |

**Gate 1 Checkpoint:** FB-OS validates Brain-issued JWTs; MCP tools/list returns registered servers.

### Phase 2: Integration Servers (Weeks 5-8)

**Theme:** MCP Servers for Core Integrations

| Week | Deliverable |
|------|-------------|
| 5 | MCP GableERP Server: 4 tools + 2 triggers with schema validation |
| 6 | MCP LocalBlue Server: 2 tools + 1 trigger; MCP XUI Server: 2 tools + 3 triggers |
| 7 | MCP QuickBooks Server: 4 tools + 2 triggers with CloudEvents webhook support |
| 8 | A2A webhook client: JWS signing, retry logic, FB-OS receiver integration |

**Gate 2 Checkpoint:** All 4 core MCP servers pass tool execution tests; A2A webhooks delivered to FB-OS.

### Phase 3: Maestro AI (Weeks 9-12)

**Theme:** AI-Driven Orchestration

| Week | Deliverable |
|------|-------------|
| 9 | Maestro orchestrator: three-tier classification (regex -> Claude -> human confirmation) |
| 10 | Maestro Materials Flow: AI-driven (replaces hardcoded pipeline); post-approval convergence |
| 11 | Maestro Labor Flow: AI-driven; QuickBooks PO automation on approval |
| 12 | Remaining MCP servers (1Build, Gmail, Outlook); Lit admin UI (MCP browser, OIDC manager) |

**Gate 3 Checkpoint:** End-to-end roofing workflow via Maestro natural language; >90% intent accuracy.

---

## 4. Out-of-Scope Boundaries

### What Brain Does NOT Do (Boundary with FB-OS)

| Concern | Brain's Role | FB-OS's Role |
|---------|-------------|-------------|
| Project data | Routes queries to XUI MCP tools | Stores projects, WBS, phases, tasks |
| CPM scheduling | Emits A2A webhook with schedule-affecting events | Runs CPM forward/backward pass |
| Daily briefings | Not involved | Worker cron jobs |
| Document AI | Not involved | Vertex AI extraction |
| Portal contacts | Issues OIDC tokens for portal users | Renders portal UI |
| Feed cards | Emits A2A webhook to create cards | Stores and renders feed cards in XUI |

### What Brain Does NOT Do (Boundary with Integrations)

| Concern | Brain's Role | Integration's Role |
|---------|-------------|-------------------|
| Product catalog | Queries GableERP via MCP tool | GableERP stores products |
| Bid management | Routes RFQs via MCP tool | LocalBlue manages bids |
| Accounting | Creates POs/invoices via MCP tool | QuickBooks stores financial records |
| Cost data | Queries 1Build via MCP tool | 1Build stores cost database |
| Email delivery | Sends via Gmail/Outlook MCP tool | Gmail/Outlook delivers email |

---

## 5. Acceptance Criteria for MVP

```
Given the Brain revamp is deployed
  And OIDC provider is serving tokens
  And 4 core MCP servers are registered (GableERP, LocalBlue, XUI, QuickBooks)
  And Maestro AI is operational with Claude tool-use
  And A2A webhooks are delivering to FB-OS

When a user authenticated via Brain OIDC says:
  "Order roofing materials for the Riverside project"

Then:
  1. Maestro classifies intent as materials_procurement (>85% confidence)
  2. GableERP MCP tools are called to get products, calculate pricing, create quote
  3. A2A webhook is sent to FB-OS with review task
  4. User sees review card in XUI within 30 seconds
  5. On approval, GableERP converts quote to order via MCP tool
  6. QuickBooks PO is created automatically via MCP tool
  7. A2A webhook notifies FB-OS of order placement
  8. All events are logged in IntegrationEvent audit trail
  9. Total time from intent to review card: <30 seconds
  10. Total time from approval to QuickBooks PO: <60 seconds
```
