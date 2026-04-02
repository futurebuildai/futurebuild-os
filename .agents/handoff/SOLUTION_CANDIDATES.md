# Solution Candidates

**Document ID:** AG-03-SOLUTIONS
**Created:** 2026-04-02
**Status:** Stage 03 Complete

---

## Concept 1: Evolve registry.go into Full MCP Server Registry

### Description
Transform the compile-time `Registry` struct (from `reference-vault/FB-Brain/internal/registry/registry.go`) into a runtime MCP Server Registry that supports dynamic registration, tool discovery via `tools/list`, and tool execution via `tools/call`.

### Architecture

```
Current:                          Revamp:
registry.New()                    MCPRegistry.RegisterServer(server)
  -> r.register(gableSystem())     -> stores MCP Server definition in PostgreSQL
  -> r.register(xuiSystem())       -> validates JSON Schema for tools
  -> ...7 hardcoded systems         -> exposes tools/list, tools/call endpoints

Registry.Get("gable")            MCPRegistry.DiscoverTools(filter)
  -> returns SystemDefinition       -> returns MCP Tool definitions
                                    -> filterable by server, capability, tag

Registry.GetAction("gable",      MCPRegistry.CallTool("gable", "create_quote", input)
  "create_quote")                   -> validates input against JSON Schema
  -> returns ActionDefinition       -> executes via HTTP client with stored auth
                                    -> returns MCP Tool result
```

### Schema Mapping

From `reference-vault/FB-Brain/internal/registry/types.go`:

```go
// Legacy FieldDef
type FieldDef struct {
    Name        string `json:"name"`
    Type        string `json:"type"`     // "string", "number", etc.
    Required    bool   `json:"required"`
    Description string `json:"description"`
}

// Evolves to MCP Tool inputSchema (JSON Schema 2020-12)
{
    "type": "object",
    "properties": {
        "customer_id": {"type": "string", "description": "GableERP customer ID"},
        "items": {"type": "array", "items": {"type": "object", ...}}
    },
    "required": ["customer_id", "items"]
}
```

### Implementation Plan
1. Create `internal/mcp/registry.go` with PostgreSQL-backed server storage
2. Write migration to convert legacy SystemDefinition to MCP Server JSON
3. Implement MCP transport handlers using `modelcontextprotocol/go-sdk`
4. Add admin API endpoints for dynamic server CRUD
5. Preserve backward-compatible `/api/hub/registry/systems` endpoint

### Verdict: **RECOMMENDED** -- This is the natural evolution of the existing pattern, adds dynamic registration without losing the structured definition approach.

---

## Concept 2: Evolve MaterialsFlow/LaborFlow into Maestro AI Orchestrator

### Description
Replace the hardcoded `MaterialsFlow` and `LaborFlow` (from `reference-vault/FB-Brain/internal/orchestrator/`) with a probabilistic AI orchestrator (Maestro) that classifies natural-language intent, selects MCP Tools, and emits signed A2A webhooks.

### Architecture

```
Current:                          Revamp:
MaterialsFlow.Start(ctx,         Maestro.ProcessIntent(ctx,
  orgID, projectID, cardID)         "order roofing materials for Riverside")
  -> hardcoded 9-step pipeline      -> Step 1: Claude classifies intent
                                    -> Step 2: Resolve MCP Tools
                                    -> Step 3: Execute Tool chain
                                    -> Step 4: Emit A2A webhook to FB-OS

LaborFlow.Start(ctx,             Same Maestro entry point:
  orgID, projectID, cardID)       Maestro.ProcessIntent(ctx,
  -> hardcoded 6-step pipeline      "find roofers for Riverside, $15k budget")
```

### Maestro Pipeline

```
User Input -> Regex Fast-Path (exact commands)
           -> Claude Tool-Use (probabilistic intent -> MCP tool selection)
           -> Confidence Gate (>0.85 = auto-execute, <0.85 = human confirmation)
           -> MCP Tool Execution (with auth from OIDC provider)
           -> A2A Webhook Emission (signed with JWS, sent to FB-OS)
           -> Result Assembly (response to user)
```

### A2A Webhook Format

```json
{
    "jsonrpc": "2.0",
    "method": "tasks/send",
    "params": {
        "id": "task-uuid",
        "message": {
            "role": "agent",
            "parts": [{
                "type": "data",
                "data": {
                    "flow_type": "materials",
                    "action": "quote_approved",
                    "rfq_id": "rfq-123",
                    "gable_order_id": "order-456",
                    "line_items": [...]
                }
            }]
        }
    }
}
```

### Implementation Plan
1. Create `internal/maestro/orchestrator.go` with three-tier classification
2. Create `internal/maestro/intent.go` for intent taxonomy (top-20 construction intents)
3. Create `internal/a2a/client.go` for signed webhook emission using `go-jose/v4`
4. Create `internal/a2a/agent_card.go` for Agent Card management
5. Preserve existing MaterialsFlow/LaborFlow as "deterministic fallback" during migration

### Verdict: **RECOMMENDED** -- The hybrid probabilistic/deterministic approach aligns with enterprise AI best practices. Legacy flows serve as fallback and test baseline.

---

## Concept 3: OIDC Provider Implementation Approach

### Option 3A: Custom OIDC Provider with zitadel/oidc Library

**Architecture:** Embed the `zitadel/oidc` v3 library directly into Brain's Go binary. Brain's existing `hub_users` table is extended with OIDC fields. Brain serves standard OIDC endpoints on the same Chi router.

```
Brain (single binary)
  ├── Chi Router
  │   ├── /.well-known/openid-configuration
  │   ├── /authorize
  │   ├── /token
  │   ├── /userinfo
  │   ├── /jwks
  │   ├── /api/hub/* (existing Hub routes)
  │   └── /api/mcp/* (new MCP routes)
  ├── zitadel/oidc OP (OpenID Provider)
  ├── go-jose/v4 (key management)
  └── PostgreSQL (users, clients, grants, keys)
```

**Pros:** Single binary deployment; full control; minimal new dependencies; reuses existing user management.
**Cons:** Must implement consent UI, key rotation, and session management manually.

### Option 3B: Ory Hydra as Sidecar

**Architecture:** Deploy Hydra as a separate Docker container. Brain implements the login/consent endpoints that Hydra redirects to. Hydra handles all OIDC protocol compliance.

```
Docker Compose:
  brain:8082 (Brain API) <--> hydra:4444 (public OIDC)
                              hydra:4445 (admin API)
                              postgres:5435 (shared DB)
```

**Pros:** OpenID Certified; battle-tested by OpenAI and others; handles protocol edge cases.
**Cons:** Two processes; operational complexity; Hydra requires Ory Kratos for user management (third process); 100MB+ memory overhead.

### Option 3C: Zitadel as Full Platform

**Architecture:** Deploy Zitadel as the identity management platform. Brain delegates all auth to Zitadel. Brain becomes a Zitadel client.

**Pros:** Most complete feature set; built-in multi-tenancy; certified.
**Cons:** Heavy deployment (CockroachDB/PG, 500MB+ RAM); Brain loses control of user management; architectural inversion (Brain depends on Zitadel, not the other way around).

### Verdict: **RECOMMEND 3A** (Custom with zitadel/oidc library)

Rationale: Brain already has user management, org management, and session infrastructure from the Hub. The `zitadel/oidc` library provides certified OP implementation as a Go library that can be embedded. This preserves single-binary deployment and full architectural control. Ory Hydra is the backup option if OIDC certification becomes a hard requirement before launch.

---

## Concept 4: Hub Admin UI Migration (React to Lit)

### Option 4A: Incremental Migration

**Approach:** Introduce Lit components alongside existing React components using a micro-frontend pattern. New pages are built in Lit; existing pages are migrated one at a time.

```
Phase 1: Add Lit entry point alongside React (both served by Vite)
Phase 2: New pages (MCP Registry Browser, OIDC Client Manager) in Lit
Phase 3: Migrate existing pages (Ecosystem, Marketplace, Login) to Lit
Phase 4: Remove React dependency
```

**Pros:** No big-bang rewrite; new features ship immediately; risk is bounded per page.
**Cons:** Temporary dual-framework overhead; some shared state complexity.

### Option 4B: Clean Break

**Approach:** Rebuild the entire Hub UI in Lit from scratch. Freeze React UI; build Lit UI in parallel; swap at a milestone.

**Pros:** Cleaner architecture; no framework mixing; faster long-term velocity.
**Cons:** Parallel development; feature parity gap during migration; higher short-term cost.

### Verdict: **RECOMMEND 4A** (Incremental Migration)

Rationale: The Hub UI is not the critical path. New features (MCP Registry Browser, OIDC Client Manager, A2A Agent Card Viewer) should be built in Lit from day one. Existing pages can be migrated opportunistically. This aligns with FB-OS's Lit architecture (from `reference-vault/futurebuild-os/CLAUDE.md`) for ecosystem consistency.

---

## Concept 5: Maestro as Primary Interface (Chat-First)

### Description
Make the Maestro AI co-pilot the **primary** user interface, with the admin dashboard as a secondary view. Users interact via natural language; the system translates intent into MCP tool calls and A2A webhooks.

### Architecture

```
User -> Maestro Chat Interface (Lit Web Component)
     -> Claude Tool-Use (MCP tools as function definitions)
     -> MCP Tool Execution (against registered servers)
     -> A2A Webhook Emission (to FB-OS for execution)
     -> Rich Card Response (in chat UI with action buttons)
```

### Key Design Decisions

1. **Chat is the router, not the destination** -- Maestro routes users to the right system, not replaces those systems. "Order materials" starts a flow; the review card appears in XUI, not in chat.
2. **Cards replace forms** -- Instead of form-based workflows, Maestro presents rich cards with pre-filled data and approve/reject buttons.
3. **Confidence transparency** -- Maestro shows its confidence level: "I'm 92% sure you want to order roofing materials for Riverside. Confirm?"
4. **Deterministic fallback** -- Legacy MaterialsFlow/LaborFlow remain as fallback for when AI classification fails or when users prefer direct actions.

### Verdict: **RECOMMENDED WITH CAVEAT** -- Chat-first is the differentiator, but it must not block users who prefer direct actions. The admin dashboard remains available for configuration, monitoring, and direct tool invocation. Maestro is the AI layer ON TOP of the MCP registry, not a replacement for it.
