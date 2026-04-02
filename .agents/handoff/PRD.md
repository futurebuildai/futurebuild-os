# Product Requirements Document

**System:** FutureBuild Brain (System of Connection)
**Pipeline Stage:** 06 - Product Specification
**Date:** 2026-04-02
**Status:** COMPLETE

---

## 1. Product Overview

FutureBuild Brain is the identity, integration, and orchestration layer for the FutureBuild ecosystem. It serves as the OIDC identity provider, hosts the MCP Server Registry for construction tool integrations, provides the Maestro AI co-pilot for natural-language workflow execution, and emits A2A signed webhooks to coordinate with FutureBuild OS.

### Vision Statement

Transform fragmented construction tool workflows into a unified, AI-orchestrated ecosystem where a general contractor can say "order materials for Phase 3" and the system executes across ERP, subcontractor, and accounting systems in under 30 seconds.

### Target Users

| Persona | Role | Primary Surface | Key Jobs |
|---------|------|-----------------|----------|
| Mike | General Contractor | Maestro AI Co-pilot | Materials procurement, labor bidding, bid comparison |
| Sarah | Office Manager | Maestro AI Co-pilot | QuickBooks PO automation, reconciliation |
| Alex | Platform Admin | Hub Admin Dashboard + Integration Registry | OIDC setup, MCP server registration, monitoring |
| Dave | Subcontractor | (External — receives RFQs via LocalBlue) | Responds to bid requests |

---

## 2. User Stories by Journey

### Journey 1: Materials Procurement via Maestro

**JTBD Reference:** J1 - Cross-System Workflow Execution
**Scope Reference:** P0 — Maestro Materials Flow, MCP GableERP Server

#### US-1.1: Natural Language Materials Request

**As** Mike (General Contractor),
**I want** to request a materials quote using natural language,
**So that** I don't need to navigate multiple systems or re-enter data.

**Acceptance Criteria:**

```
Given Mike is authenticated via Brain OIDC and has GableERP connected
  And Mike's org has an active AccountLink to a GableERP customer
When Mike types "I need a quote for framing lumber on the Elm Street project"
Then Maestro classifies the intent as materials_procurement with >90% confidence
  And Maestro resolves "Elm Street" to the XUI project ID via MCP tool call
  And Maestro resolves Mike's org to the GableERP customer ID via OIDC subject claims
  And intent classification completes in <2 seconds (p99)
  And the classification tier used (regex, Claude, or human) is logged for metrics
```

#### US-1.2: Automated MCP Tool Chain Execution

**As** the system,
**I want** to execute the materials procurement workflow through MCP tool calls,
**So that** the user gets a review-ready quote without manual steps.

**Acceptance Criteria:**

```
Given Maestro has classified an intent as materials_procurement
  And the target MCP server (GableERP) is registered and healthy
When Maestro executes the tool chain
Then the following MCP tool calls execute in sequence:
  1. GableERP: get_products_by_category(category="Framing") — returns product list
  2. GableERP: bulk_calculate_price(customer_id, items) — returns pricing
  3. GableERP: create_quote(customer_id, quote_lines) — returns quote_id
And each tool call validates input against JSON Schema 2020-12
  And each tool call completes within the 3-second p99 target
  And tool execution success rate is >95%
  And all tool calls are logged as IntegrationEvent records with source, target, action, status
```

#### US-1.3: A2A Webhook Emission for Review Task

**As** Mike,
**I want** to see a review card in my XUI Projects feed after a quote is created,
**So that** I can approve or modify the order.

**Acceptance Criteria:**

```
Given a materials quote has been created via MCP GableERP tools
When Brain emits an A2A webhook to FB-OS
Then the webhook payload includes: task_type="review_material_quote", quote_data (line items, totals, vendor)
  And the webhook is signed with JWS using go-jose/v4 (RS256 algorithm)
  And FB-OS verifies the JWS signature against Brain's public key
  And webhook delivery latency is <500ms (p99)
  And Mike sees the review card in XUI within 30 seconds of the original request
  And the review card displays line items with costs formatted from BIGINT cents
```

#### US-1.4: Quote Approval and Post-Approval Flow

**As** Mike,
**I want** to approve a materials quote and have the system handle the rest,
**So that** ordering and accounting happen without manual intervention.

**Acceptance Criteria:**

```
Given Mike clicks "Approve Quote" on the review card in FB-OS
When Brain receives the approval event (via A2A callback or polling)
Then Brain calls GableERP MCP tool: accept_and_convert_quote(quote_id)
  And Brain calls QuickBooks MCP tool: create_purchase_order(company_id, vendor_id, line_items)
  And Brain emits A2A webhook to FB-OS: task_type="update_schedule" with materials_ordered and expected_delivery
  And all three actions complete within 60 seconds of approval
  And the QuickBooks PO line items match the GableERP quote line-for-line
  And an IntegrationEvent audit record captures: source=gable, target=quickbooks, action=create_po
```

---

### Journey 2: Subcontractor Bidding via Maestro

**JTBD Reference:** J1 - Cross-System Workflow Execution
**Scope Reference:** P0 — Maestro Labor Flow, MCP LocalBlue Server

#### US-2.1: Natural Language Bid Request

**As** Mike,
**I want** to request subcontractor bids using natural language,
**So that** I can start the bidding process in seconds.

**Acceptance Criteria:**

```
Given Mike is authenticated and has LocalBlue connected
When Mike types "Find me roofers for the Riverside project, budget around $15k"
Then Maestro classifies intent as labor_bidding with >90% confidence
  And Maestro resolves "Riverside" to the project and extracts budget_hint=$15000
  And Maestro queries 1Build MCP Server for regional labor rates (if connected)
  And Maestro constructs RFQ scope items using AI-inferred project phase data
  And LocalBlue MCP tool push_rfq(site_id, rfq_data) is called
  And an A2A webhook notifies FB-OS that an RFQ was submitted
```

#### US-2.2: Bid Receipt and AI Enrichment

**As** Mike,
**I want** received bids to be enriched with budget and schedule context,
**So that** I can make informed decisions without manual research.

**Acceptance Criteria:**

```
Given a subcontractor submits a bid via LocalBlue
When LocalBlue fires the bid_submitted trigger to Brain
Then Maestro contextualizes the bid:
  - Compares bid amount to budget_hint (percentage above/below)
  - Compares bid timeline to CPM schedule constraints (if available via A2A)
  - Queries 1Build for market rate comparison (if connected)
And an A2A webhook is emitted to FB-OS: task_type="review_labor_bid" with bid_data and ai_analysis
  And the review card shows: bidder name, bid amount (BIGINT cents formatted), timeline, AI analysis summary
  And the AI analysis is generated by Claude with <$0.03 cost per enrichment
```

#### US-2.3: Bid Approval and Post-Approval Convergence

**As** Mike,
**I want** the system to handle all follow-up actions when I approve a bid,
**So that** the subcontractor is onboarded and accounting is updated automatically.

**Acceptance Criteria:**

```
Given Mike approves a labor bid in FB-OS
When Brain receives the approval event
Then Brain calls LocalBlue MCP tool: update_bid_status(bid_id, "accepted")
  And Brain calls XUI MCP tool: assign_contact_to_phase(project_id, phase_code, contact)
  And Brain calls QuickBooks MCP tool: create_subcontract(vendor_id, amount, project)
  And Brain checks post-approval convergence: if materials_ordered AND labor_approved, emit delivery confirmation A2A webhook
  And all post-approval actions complete within 60 seconds
  And all actions are logged as IntegrationEvent records
```

---

### Journey 3: First-Time Setup (OIDC + Integration Registration)

**JTBD Reference:** J2 - Single Sign-On, J3 - Integration Discovery and Management
**Scope Reference:** P0 — OIDC Provider, MCP Registry, OIDC Client Management

#### US-3.1: OIDC Provider Deployment

**As** Alex (Platform Admin),
**I want** Brain to serve as an OpenID Connect provider,
**So that** all ecosystem services authenticate through a single identity authority.

**Acceptance Criteria:**

```
Given Brain is deployed with OIDC provider enabled
When any client requests /.well-known/openid-configuration
Then Brain responds with a conformant OpenID Configuration document including:
  - issuer: Brain's base URL
  - authorization_endpoint: /authorize
  - token_endpoint: /token
  - jwks_uri: /jwks
  - userinfo_endpoint: /userinfo
  - revocation_endpoint: /revoke
  - supported scopes: openid, profile, email, org
  - supported grant types: authorization_code
  - supported response types: code
  - code_challenge_methods_supported: S256 (PKCE)
And JWKS endpoint returns RSA public keys for JWT verification
  And JWKS availability is 99.99%
```

#### US-3.2: OIDC Client Registration

**As** Alex,
**I want** to register FB-OS as an OIDC client,
**So that** FB-OS users authenticate via Brain.

**Acceptance Criteria:**

```
Given Alex accesses the OIDC Client Manager in the Admin UI
When Alex creates a new OIDC client with:
  - client_name: "FutureBuild OS"
  - redirect_uris: ["https://os.futurebuild.io/auth/callback"]
  - grant_types: ["authorization_code"]
  - token_endpoint_auth_method: "none" (public client with PKCE)
Then Brain stores the client configuration in PostgreSQL
  And generates a client_id (no client_secret for public clients with PKCE)
  And the client appears in the OIDC Client Manager list
  And FB-OS can initiate authorization code flow with PKCE against Brain
```

#### US-3.3: Cross-System JWT Authentication

**As** Mike,
**I want** to log into FB-OS using my Brain credentials,
**So that** I don't need separate accounts.

**Acceptance Criteria:**

```
Given FB-OS is configured with Brain's OIDC issuer URL
When Mike navigates to FB-OS login
Then FB-OS redirects to Brain's /authorize endpoint with PKCE challenge
  And Mike authenticates via Brain (magic link email)
  And Brain issues: access_token (JWT), refresh_token, id_token
  And access_token contains claims: sub, org_id, role, plan_tier, iat, exp
  And FB-OS validates the JWT via Brain's /jwks endpoint
  And token issuance latency is <500ms (p99)
  And FB-OS JWT validation latency is <10ms (with JWKS cache)
  And refresh token rotation succeeds >99.5% without re-authentication
```

#### US-3.4: MCP Server Registration

**As** Alex,
**I want** to register integration servers via the Admin UI,
**So that** Maestro can discover and use their tools.

**Acceptance Criteria:**

```
Given Alex accesses the MCP Registry Browser in the Admin UI
When Alex registers a new MCP server with:
  - name: "GableERP"
  - base_url: "https://api.gable-erp.com"
  - auth_type: "api_key"
  - tools: [{name: "get_products_by_category", input_schema: {...}, description: "..."}]
  - triggers: [{name: "price_updated", description: "..."}]
Then the server is stored in PostgreSQL with a unique server_id
  And tools/list returns the server and its tools
  And tool discovery latency is <50ms (p99)
  And the server appears in the Admin UI with health status indicator
  And health checks run on a configurable interval (default: 60s)
```

#### US-3.5: Integration Connection (User)

**As** Mike,
**I want** to connect my GableERP and QuickBooks accounts to Brain,
**So that** Maestro can execute actions on my behalf.

**Acceptance Criteria:**

```
Given Mike is authenticated via Brain OIDC
When Mike connects GableERP in the Hub
Then Brain initiates the appropriate auth flow:
  - API Key: Mike enters API key, Brain encrypts and stores linked to OIDC subject
  - OAuth2 (QuickBooks): Brain initiates OAuth2 flow with Intuit, stores encrypted tokens
And credentials are encrypted at rest with AES-256-GCM
  And credential refresh (for OAuth2) handles token rotation automatically
  And QuickBooks refresh token follows Intuit's 5-year rotation policy
  And connected integrations appear in Mike's Hub dashboard
```

---

### Journey 4: QuickBooks Automated PO Creation

**JTBD Reference:** J4 - Automated Financial Reconciliation
**Scope Reference:** P0 — QuickBooks PO Automation, MCP QuickBooks Server

#### US-4.1: Automatic PO on Materials Approval

**As** Sarah (Office Manager),
**I want** a Purchase Order to be automatically created in QuickBooks when a materials quote is approved,
**So that** I don't need to manually re-enter order data.

**Acceptance Criteria:**

```
Given a materials quote has been approved in the MaterialsFlow
  And Mike's org has QuickBooks connected via OAuth2
  And a vendor mapping exists between GableERP supplier and QuickBooks vendor
When the approval triggers QuickBooks MCP tool: create_purchase_order
Then a PO is created in QuickBooks with:
  - Vendor: mapped from GableERP supplier
  - Line items: matching GableERP quote lines (item name, quantity, unit price in cents)
  - Total: matching GableERP quote total
And Sarah sees the PO in QuickBooks within 60 seconds
  And an IntegrationEvent is logged with source=gable, target=quickbooks, action=create_po
  And zero manual data entry is required by Sarah
```

#### US-4.2: CloudEvents Webhook Support

**As** the system,
**I want** to receive QuickBooks notifications in CloudEvents format,
**So that** the integration is compliant with Intuit's May 2026 migration deadline.

**Acceptance Criteria:**

```
Given QuickBooks is configured to send webhooks to Brain
When QuickBooks emits a notification event
Then Brain's webhook receiver accepts CloudEvents-formatted payloads
  And the receiver validates the CloudEvents spec headers (ce-type, ce-source, ce-id, ce-time)
  And the event is stored as an IntegrationEvent for audit
  And relevant events trigger Maestro notifications (e.g., invoice paid -> feed card)
  And the implementation is complete before May 2026 Intuit deadline
```

#### US-4.3: Automatic Subcontract on Bid Approval

**As** Sarah,
**I want** a subcontract record to be created in QuickBooks when a labor bid is approved,
**So that** accounting tracks all committed costs.

**Acceptance Criteria:**

```
Given a labor bid has been approved in the LaborFlow
  And the subcontractor has a mapped QuickBooks vendor record
When Brain processes the bid approval
Then QuickBooks MCP tool creates a subcontract/bill with:
  - Vendor: subcontractor from LocalBlue
  - Amount: bid amount in BIGINT cents
  - Project: linked to the construction project
And Sarah sees the subcontract in QuickBooks within 60 seconds
  And the bid amount matches the QuickBooks entry exactly (integer cents, no rounding)
```

---

### Journey 5: AI-Powered Bid Comparison

**JTBD Reference:** J5 - Real-Time Project Intelligence
**Scope Reference:** P2 — Maestro Bid Comparison (post-MVP)

#### US-5.1: Multi-Bid Comparison Request

**As** Mike,
**I want** to compare all bids for a specific phase using natural language,
**So that** I can make data-driven award decisions.

**Acceptance Criteria:**

```
Given multiple bids have been received for a project phase via LocalBlue
When Mike types "Compare all bids for the plumbing phase on Riverside"
Then Maestro:
  1. Queries LocalBlue MCP Server for all bids on the relevant RFQ
  2. Queries 1Build MCP Server for regional labor rates (market baseline)
  3. Generates a comparison table: bidder name, bid amount, vs. market rate (%), timeline, rating
  4. Highlights the best-value bid based on price, timeline, and market alignment
And the comparison renders in the Maestro chat with JetBrains Mono for numerical data
  And all bid amounts are displayed from BIGINT cents (no floating-point)
  And Mike can say "Award it to [bidder]" to trigger the approval flow
```

---

## 3. Non-Functional Requirements

### NFR-1: Performance

| Requirement | Target | Measurement | Reference |
|-------------|--------|-------------|-----------|
| NFR-1.1: Token Issuance Latency (p99) | <500ms | Prometheus histogram on /token | OIDC Pillar |
| NFR-1.2: JWT Validation Latency (p99) | <10ms | Prometheus histogram (JWKS cached) | OIDC Pillar |
| NFR-1.3: Tool Discovery Latency (p99) | <50ms | Prometheus histogram on tools/list | MCP Pillar |
| NFR-1.4: Tool Execution Latency (p99) | <3s | Prometheus histogram on tools/call | MCP Pillar |
| NFR-1.5: Intent Classification Latency (p99) | <2s | Prometheus histogram | Maestro Pillar |
| NFR-1.6: Webhook Delivery Latency (p99) | <500ms | Prometheus histogram | A2A Pillar |
| NFR-1.7: Admin UI First Contentful Paint | <1.5s | Lighthouse CI | UI Pillar |
| NFR-1.8: Admin UI Time to Interactive | <3s | Lighthouse CI | UI Pillar |
| NFR-1.9: End-to-End Workflow (intent to review card) | <30s | Distributed tracing | Cross-System |
| NFR-1.10: Approval to QuickBooks PO | <60s | IntegrationEvent timestamps | Cross-System |

### NFR-2: Reliability

| Requirement | Target | Measurement |
|-------------|--------|-------------|
| NFR-2.1: Brain API Uptime | 99.9% | Health check monitoring |
| NFR-2.2: JWKS Availability | 99.99% | Uptime monitor on /.well-known/openid-configuration + /jwks |
| NFR-2.3: Webhook Delivery Rate | >99.9% | Prometheus counter (2xx / total) |
| NFR-2.4: Webhook Dead Letter Rate | <0.01% | Prometheus counter + alert |
| NFR-2.5: MCP Tool Execution Success | >95% | Prometheus counter per server |
| NFR-2.6: Maestro Workflow Completion | >80% at MVP, >95% at 6 months | Prometheus counter |
| NFR-2.7: Refresh Token Rotation | >99.5% success without re-auth | Prometheus counter |
| NFR-2.8: Error Rate (5xx) | <0.1% | Prometheus counter |

### NFR-3: Security

| Requirement | Specification |
|-------------|---------------|
| NFR-3.1: OIDC Compliance | OpenID Connect Core 1.0 conformant with PKCE (RFC 7636) |
| NFR-3.2: Token Signing | RS256 JWT signatures with key rotation support |
| NFR-3.3: A2A Webhook Signing | JWS with RS256 via go-jose/v4; RFC 8785 canonical JSON |
| NFR-3.4: Credential Encryption | All integration credentials encrypted at rest with AES-256-GCM |
| NFR-3.5: Brute Force Protection | Failed auth attempts rate-limited; <1% of total attempts threshold |
| NFR-3.6: SQL Injection Prevention | All queries via pgx parameterized queries |
| NFR-3.7: CORS | Strict origin allowlist for OIDC endpoints |
| NFR-3.8: Token Claims | Access tokens carry: sub, org_id, role, plan_tier, iat, exp, iss, aud |
| NFR-3.9: Zero-Trust Phase 1 | OIDC JWT bearer tokens for all service-to-service auth |
| NFR-3.10: Zero-Trust Phase 2 | SPIFFE/SPIRE workload identity for mTLS (post-MVP) |

### NFR-4: Data Integrity

| Requirement | Specification | Enforcement |
|-------------|---------------|-------------|
| NFR-4.1: Composite Currency Pattern | All monetary values in tool payloads and events use the Composite Currency Pattern: `amount_cents` (int64) + `currency_code` (string, "USD" or "CAD"). Cross-currency arithmetic forbidden. | SQL Migration Linter: (a) forbidden types = hard CI fail; (b) `amount_cents` without `currency_code` = hard CI fail |
| NFR-4.2: Audit Trail | Every cross-system action logged as IntegrationEvent with: source, target, action, status, timestamp, org_id | PostgreSQL table with retention policy |
| NFR-4.3: Idempotency | All webhook deliveries carry idempotency keys; receivers reject duplicates | UUID v7 per webhook |
| NFR-4.4: Webhook Signature Verification | 100% of inbound and outbound webhooks carry valid JWS signatures | go-jose/v4 RS256 |

### NFR-5: AI Cost & Quality

| Requirement | Target | Measurement |
|-------------|--------|-------------|
| NFR-5.1: Intent Classification Accuracy | >90% top-1 accuracy | Weekly eval against 200-utterance test set |
| NFR-5.2: Human Confirmation Rate | <15% of intents | Prometheus counter |
| NFR-5.3: Cost per Intent | <$0.03 average | Anthropic usage tracking |
| NFR-5.4: Confidence Distribution | Median >0.90 | Prometheus histogram |
| NFR-5.5: Regex Fast-Path Utilization | Track ratio vs Claude | Prometheus counter per tier |

### NFR-6: Admin UI Quality

| Requirement | Target | Measurement |
|-------------|--------|-------------|
| NFR-6.1: Lighthouse Performance | >90 | Lighthouse CI |
| NFR-6.2: Lighthouse Accessibility | >95 | Lighthouse CI |
| NFR-6.3: Bundle Size (gzipped) | <50KB | Vite build output |
| NFR-6.4: WCAG 2.1 AA | All admin surfaces | Automated + manual audit |
| NFR-6.5: Dark-Only | Industrial Dark theme only — no light mode | Design constraint |

### NFR-7: Observability

| Requirement | Specification |
|-------------|---------------|
| NFR-7.1: Metrics | Prometheus via promhttp middleware on Chi router |
| NFR-7.2: Tracing | OpenTelemetry distributed tracing for Brain -> external API calls |
| NFR-7.3: Logging | slog structured JSON with correlation IDs |
| NFR-7.4: Dashboards | Grafana with panels per pillar metric group |
| NFR-7.5: Alerting | Brain API <99.5% uptime, >0.5% 5xx, >85% PostgreSQL pool, webhook dead letter >0.01% |

### NFR-8: Operational

| Requirement | Target |
|-------------|--------|
| NFR-8.1: PostgreSQL Connection Pool | <80% utilization, alert at >85% |
| NFR-8.2: Memory Usage (RSS) | <512MB, alert at >400MB |
| NFR-8.3: CPU Usage | <50% sustained, alert at >70% for 5m |
| NFR-8.4: IntegrationEvent Volume | Baseline + anomaly detection, alert at >3x |

---

## 4. Traceability Matrix

Every P0 "Must Have" capability from SCOPE_DEFINITION.md is traced to user stories and NFRs.

### Walking Skeleton Traceability

| Component | User Stories | NFRs |
|-----------|-------------|------|
| OIDC Provider (minimal) | US-3.1, US-3.3 | NFR-1.1, NFR-1.2, NFR-2.2, NFR-3.1, NFR-3.2 |
| MCP Registry (minimal) | US-3.4 | NFR-1.3, NFR-2.5 |
| Maestro (regex only) | US-1.1 | NFR-1.5, NFR-5.5 |
| A2A Client (unsigned) | US-1.3 | NFR-1.6, NFR-2.3 |
| Admin UI (static list) | US-3.4 | NFR-6.1, NFR-6.2, NFR-6.3 |

### MVP P0 Feature Traceability

| Feature | Scope Ref | User Stories | NFRs | JTBD |
|---------|-----------|-------------|------|------|
| OIDC Provider (full) | P0 | US-3.1, US-3.2, US-3.3 | NFR-1.1, NFR-1.2, NFR-2.2, NFR-2.7, NFR-3.1, NFR-3.2, NFR-3.5, NFR-3.7, NFR-3.8 | J2 |
| OIDC Client Management | P0 | US-3.2 | NFR-3.1 | J2 |
| MCP GableERP Server | P0 | US-1.1, US-1.2, US-1.4 | NFR-1.3, NFR-1.4, NFR-2.5 | J1 |
| MCP LocalBlue Server | P0 | US-2.1, US-2.2, US-2.3 | NFR-1.3, NFR-1.4, NFR-2.5 | J1 |
| MCP XUI Server | P0 | US-1.3, US-2.3 | NFR-1.3, NFR-1.4, NFR-2.5 | J1 |
| MCP QuickBooks Server | P0 | US-4.1, US-4.2, US-4.3 | NFR-1.3, NFR-1.4, NFR-2.5 | J4 |
| Maestro Intent Classification | P0 | US-1.1, US-2.1 | NFR-1.5, NFR-5.1, NFR-5.2, NFR-5.3, NFR-5.4 | J1 |
| Maestro Materials Flow | P0 | US-1.1, US-1.2, US-1.3, US-1.4 | NFR-1.9 | J1 |
| Maestro Labor Flow | P0 | US-2.1, US-2.2, US-2.3 | NFR-1.9 | J1 |
| A2A Signed Webhooks | P0 | US-1.3, US-2.2, US-2.3 | NFR-1.6, NFR-2.3, NFR-2.4, NFR-3.3, NFR-4.3 | J1, J5 |
| Post-Approval Logic | P0 | US-1.4, US-2.3 | NFR-1.10 | J1 |
| QuickBooks PO Automation | P0 | US-4.1, US-4.2, US-4.3 | NFR-1.10, NFR-4.1 | J4 |

### MVP P1 Feature Traceability

| Feature | Scope Ref | Target Week | JTBD |
|---------|-----------|-------------|------|
| MCP 1Build Server | P1 | Week 12 | J7 |
| MCP Gmail Server | P1 | Week 12 | J8 |
| MCP Outlook Server | P1 | Week 12 | J8 |
| Dynamic MCP Registration | P1 | Week 12 | J3 |
| Lit MCP Registry Browser | P1 | Week 12 | J3 |
| Lit OIDC Client Manager | P1 | Week 12 | J2 |
| Maestro Top-20 Intents | P1 | Week 12 | J1 |

### Post-MVP Feature Traceability (P2+)

| Feature | Target Phase | JTBD |
|---------|-------------|------|
| A2A Agent Cards (full spec) | Post-MVP | J5 |
| SPIFFE/SPIRE Integration | Post-MVP | J2 |
| Maestro Bid Comparison | Post-MVP | J5 |
| Lit Migration (remaining pages) | Post-MVP | — |
| Multi-org OIDC | Post-MVP | J9 |
| Webhook Dashboard | Post-MVP | J6 |

---

## 5. Cross-System Contract

### FB-Brain -> FB-OS Interface

| Interface | Protocol | Authentication | Contract |
|-----------|----------|---------------|----------|
| JWT Issuance | OIDC | — | Brain issues; OS validates via JWKS |
| A2A Webhooks | HTTPS POST | JWS-signed (RS256) | Brain emits; OS verifies signature and processes |
| JWKS Endpoint | HTTPS GET | Public | Brain serves; OS caches and refreshes |

### A2A Webhook Event Types

| Event Type | Payload | Trigger |
|-----------|---------|---------|
| review_material_quote | quote_id, line_items, total_cents, currency_code, vendor | Materials quote created |
| review_labor_bid | bid_id, bidder, amount_cents, currency_code, timeline, ai_analysis | Bid received from LocalBlue |
| update_schedule | event_type, delivery_date, constraints | Materials ordered / sub confirmed |
| delivery_confirmation | materials_ordered, labor_approved, convergence_status | Post-approval convergence check |
| create_feed_card | card_type, title, body, actions, priority | Integration event notification |

### Blocking Dependencies

| Dependency | Blocks | Owner | Status |
|-----------|--------|-------|--------|
| FB-Brain JWKS endpoint | FB-OS walking skeleton auth | Brain team | NEEDED |
| FB-OS A2A webhook receiver | Brain walking skeleton webhook delivery | OS team | NEEDED |
| QuickBooks CloudEvents format | QB MCP Server webhook handler | Brain team | NEEDED before May 2026 |

---

## 6. Business Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Time to First Integration | <30 minutes | Funnel: account creation -> first MCP tool call |
| Active Connected Systems per Org | >3 at 90 days | PostgreSQL query |
| Weekly Active Maestro Users | >60% of active users | Usage analytics |
| Manual Data Entry Reduction | >50% | Quarterly survey |
| QuickBooks Auto-PO Adoption | >70% within 90 days | Approved quotes -> auto-PO count |
| Cross-System Workflow Completion | >80% at MVP, >95% at 6 months | Initiated vs completed workflows |

---

## 7. Success Criteria by Phase

### Phase 1: Foundation (Weeks 1-4)

```
Given Brain is deployed with OIDC and MCP Registry
When FB-OS validates a Brain-issued JWT
  And tools/list returns registered servers
  And walking skeleton end-to-end flow completes (auth -> tool call -> result)
Then Phase 1 is PASSED
```

### Phase 2: Integration Servers (Weeks 5-8)

```
Given 4 core MCP servers are operational (GableERP, LocalBlue, XUI, QuickBooks)
When all servers pass tool execution tests with >90% success rate
  And QuickBooks webhook receiver handles CloudEvents format
  And A2A webhooks are received and processed by FB-OS
Then Phase 2 is PASSED
```

### Phase 3: Maestro AI (Weeks 9-12)

```
Given Maestro is operational with Claude tool-use
When intent classification accuracy exceeds 90% on evaluation set
  And materials flow end-to-end completes ("Order roofing materials" -> review card)
  And labor flow end-to-end completes ("Find roofers" -> bid review card)
  And QuickBooks PO is created within 60s of quote approval
  And cross-system workflow completion exceeds 80%
Then Phase 3 is PASSED and MVP is COMPLETE
```

---

## 8. Out of Scope

| Item | Reason |
|------|--------|
| Full Zitadel/Hydra deployment | Custom OIDC with zitadel/oidc library is sufficient |
| Fine-tuned intent classifier | Insufficient production data; Claude tool-use is accurate enough |
| Mobile app | Web-first; Brain is an admin/integration tool |
| Municipal portal MCP server | Requires partner API access; deferred |
| Istio service mesh | Overkill for two-service architecture |
| React frontend (new) | Hard constraint: Lit 3.0 only (incremental migration from existing React) |
| Project data storage | FB-OS owns project data; Brain routes queries |
| CPM scheduling | FB-OS owns physics engine; Brain emits events |
| Daily briefings | FB-OS owns DailyFocusAgent |
| Light mode UI | Dark-only: Industrial Dark is brand identity |
