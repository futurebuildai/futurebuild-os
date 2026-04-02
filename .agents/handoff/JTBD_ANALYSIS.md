# Jobs to Be Done Analysis

**Document ID:** AG-02-JTBD
**Created:** 2026-04-02
**Status:** Stage 02 Complete

---

## Core Jobs

### Job 1: Cross-System Workflow Execution

**Job Statement:** When I need to procure materials and labor for a construction phase, I want to trigger a single request that orchestrates across my ERP, project management, and subcontractor systems, so that I don't waste hours switching between applications and re-entering data.

**Legacy Foundation:** From `reference-vault/FB-Brain/internal/orchestrator/materials_flow.go`, the 9-step MaterialsFlow already implements this for roofing materials with GableERP -> XUI Projects. The revamp generalizes this from hardcoded roofing scope to AI-inferred scope for any construction phase.

**Functional Requirements:**
1. Accept natural-language intent ("order framing lumber for Phase 3")
2. Resolve cross-system identities via AccountLink (evolved to OIDC subject claims)
3. Execute MCP Tool calls against registered servers (GableERP, QuickBooks, etc.)
4. Emit A2A webhook to FB-OS with execution results
5. Create review cards for human approval when confidence is below threshold

**Success Criteria:**
- Given a user says "get me a quote for roofing materials on Riverside"
- When Maestro processes the intent with >90% confidence
- Then GableERP MCP tools are called, a quote is created, and a review card appears in XUI within 30 seconds

---

### Job 2: Single Sign-On Across Construction Tools

**Job Statement:** When I start my workday, I want to log in once and access all my construction tools (project management, ERP, subcontractor portal, accounting) without re-authenticating, so that I spend my time building houses, not managing passwords.

**Legacy Foundation:** From `reference-vault/FB-Brain/internal/hub/auth_handlers.go`, the Hub already has login/logout/magic-link/SSO-exchange endpoints. From `reference-vault/futurebuild-os/CLAUDE.md`, FB-OS uses Clerk for JWT validation. The revamp replaces Clerk with Brain-issued OIDC tokens.

**Functional Requirements:**
1. OIDC-compliant authorization code flow with PKCE
2. JWT access tokens with construction-domain claims (org_id, role, plan_tier)
3. JWKS endpoint for downstream token validation
4. Refresh token rotation with configurable TTL
5. Session management with revocation capability

**Success Criteria:**
- Given a user authenticates with Brain's OIDC provider
- When they navigate to FB-OS
- Then FB-OS validates the Brain-issued JWT via JWKS without requiring separate login

---

### Job 3: Integration Discovery and Management

**Job Statement:** When my company adopts a new construction tool, I want to connect it to my ecosystem through a standard protocol, so that I don't need custom development for each integration.

**Legacy Foundation:** From `reference-vault/FB-Brain/internal/registry/types.go`, `SystemDefinition` provides a structured way to describe external systems. From `reference-vault/FB-Brain/internal/hub/registry_handlers.go`, the Hub exposes registry data via REST API.

**Functional Requirements:**
1. MCP Server Registry with `tools/list` discovery endpoint
2. Dynamic server registration via admin API (no code deployment required)
3. Per-server auth configuration (API key, OAuth2, mTLS)
4. Tool schema validation against JSON Schema 2020-12
5. Health monitoring and usage metrics per MCP server

**Success Criteria:**
- Given an admin registers a new MCP server for "municipal-permits-api"
- When a user asks Maestro "check my permit status"
- Then Maestro discovers the new server's tools and invokes them without code changes

---

### Job 4: Automated Financial Reconciliation

**Job Statement:** When materials are ordered or labor bids are approved, I want the corresponding financial records (POs, bills, invoices) to be created in QuickBooks automatically, so that my bookkeeper doesn't have to re-enter data.

**Legacy Foundation:** From `reference-vault/FB-Brain/internal/registry/quickbooks.go`, QuickBooks actions include `create_invoice`, `create_bill`, and `create_purchase_order`. Currently, these are defined but not wired into the MaterialsFlow/LaborFlow -- the financial step is manual.

**Functional Requirements:**
1. Post-approval A2A webhook triggers QuickBooks MCP Tool to create PO
2. Line items from GableERP quote map to QuickBooks PO line items
3. Vendor mapping between GableERP supplier and QuickBooks vendor
4. CloudEvents webhook format for QuickBooks notifications (May 2026 deadline)
5. Reconciliation dashboard showing GableERP orders vs. QuickBooks POs

**Success Criteria:**
- Given a materials quote is approved in the MaterialsFlow
- When the approval A2A webhook fires
- Then a Purchase Order is created in QuickBooks with matching line items within 60 seconds

---

### Job 5: Real-Time Project Intelligence

**Job Statement:** When something changes in any of my connected systems (price update, bid received, order shipped), I want to be notified in context with recommended actions, so that I can make decisions proactively instead of reactively.

**Legacy Foundation:** From `reference-vault/FB-Brain/internal/hub/eventbus.go` and `sse_handlers.go`, the Hub already has an SSE-based EventBus for real-time notifications. From `reference-vault/FB-Brain/internal/models/models.go`, `IntegrationEvent` captures cross-system events.

**Functional Requirements:**
1. MCP notification subscriptions for registered server events
2. Maestro AI contextualizes events with recommended actions
3. A2A push notifications to connected agents (FB-OS, mobile)
4. Event correlation across systems (e.g., GableERP price change -> QuickBooks budget impact)
5. Notification preferences per user (email, SSE, webhook)

**Success Criteria:**
- Given GableERP fires a `price_updated` trigger for roofing shingles
- When Brain receives the webhook
- Then Maestro calculates budget impact, notifies the PM via SSE, and suggests updating pending quotes

---

## Supporting Jobs

| Job | Statement | Priority |
|-----|-----------|----------|
| J6: Audit Trail | "I need to see exactly what happened across all systems for compliance and dispute resolution" | P1 |
| J7: Cost Estimation | "I need accurate cost data from 1Build before I commit to a bid" | P2 |
| J8: Email Integration | "I need project-related emails (vendor quotes, sub communications) linked to projects automatically" | P2 |
| J9: Multi-Org Management | "I manage 3 builder entities and need to switch between them without re-logging" | P2 |
| J10: Plan-Gated Features | "I want to try basic integrations for free, then upgrade for AI and advanced workflows" | P3 |
