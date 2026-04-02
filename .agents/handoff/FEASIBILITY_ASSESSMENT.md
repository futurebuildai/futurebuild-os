# Feasibility Assessment

**Document ID:** AG-03-FEASIBILITY
**Created:** 2026-04-02
**Status:** Stage 03 Complete

---

## Technical Feasibility Matrix

### Pillar 1: OIDC Identity Provider

| Criterion | Assessment | Risk Level |
|-----------|-----------|------------|
| Library maturity | `zitadel/oidc` v3 is OpenID Certified by the OpenID Foundation; actively maintained | LOW |
| Go integration | Native Go library; embeds directly into Chi router via `op.Handler()` | LOW |
| Database requirements | OIDC tables (clients, grants, auth_codes, keys, sessions) can share existing PostgreSQL instance | LOW |
| Key management | `go-jose/v4` provides RSA/ECDSA key generation, rotation, and JWKS serialization | LOW |
| PKCE support | `zitadel/oidc` supports PKCE (required for public clients like SPAs) | LOW |
| FB-OS integration | FB-OS already has JWKS-based JWT validation middleware; swap Clerk URL for Brain URL | MEDIUM |
| Consent UI | Must build consent screen; no existing pattern in Hub UI | MEDIUM |
| Token storage | Must decide: opaque tokens (DB lookup) vs. self-contained JWTs (no DB lookup) | LOW |
| Certification | Can pursue OpenID Certification if using certified library components | LOW |

**Overall: FEASIBLE** -- The `zitadel/oidc` library handles OIDC protocol compliance; Brain team builds user management and consent UI on top. Estimated effort: 4-6 weeks for core OIDC provider.

### Pillar 2: MCP Server Registry

| Criterion | Assessment | Risk Level |
|-----------|-----------|------------|
| Go SDK availability | Official `modelcontextprotocol/go-sdk` maintained with Google; production-ready | LOW |
| Schema mapping | Legacy `FieldDef` maps cleanly to JSON Schema properties (validated manually for all 21 actions) | LOW |
| Dynamic registration | PostgreSQL JSONB column stores MCP Server definition; admin API for CRUD | LOW |
| Tool execution | HTTP client with per-server auth config replaces existing `clients/` package | MEDIUM |
| Transport | Streamable HTTP transport for production; stdio for local dev/testing | LOW |
| Backward compatibility | Legacy `/api/hub/registry/systems` endpoint can be preserved as a compatibility layer | LOW |
| Performance | Tool discovery is a PostgreSQL query; tool execution adds ~10ms overhead for schema validation | LOW |

**Overall: FEASIBLE** -- Direct evolution of existing registry pattern. The mapping from SystemDefinition to MCP Server is straightforward. Estimated effort: 3-4 weeks for registry + transport.

### Pillar 3: Maestro AI Co-pilot

| Criterion | Assessment | Risk Level |
|-----------|-----------|------------|
| Claude tool-use | Claude supports function calling with JSON Schema tool definitions; proven pattern in FB-OS | LOW |
| Intent accuracy | Top-20 construction intents are well-defined domain tasks; Claude excels at domain-specific classification | MEDIUM |
| Latency | Claude API calls add ~500-1500ms; acceptable for workflow initiation but not for real-time queries | MEDIUM |
| Cost | ~$0.01-0.05 per intent classification (Claude Sonnet); manageable at expected scale | LOW |
| Deterministic fallback | Legacy MaterialsFlow/LaborFlow provide exact-match fallback | LOW |
| A2A webhook signing | `go-jose/v4` supports JWS signing with RFC 8785 canonical JSON | LOW |
| A2A webhook delivery | HTTP POST with retry (exponential backoff) is standard; `go-retryablehttp` library available | LOW |
| FB-OS receiver readiness | FB-OS has `pkg/a2a/` package already present; A2A endpoint implementation is the dependency | MEDIUM |

**Overall: FEASIBLE** -- The three-tier classification pipeline (regex -> Claude -> human confirmation) is a proven enterprise pattern. A2A webhook signing uses standard JWS. Estimated effort: 5-7 weeks for Maestro + A2A.

### Pillar 4: L7 Zero-Trust

| Criterion | Assessment | Risk Level |
|-----------|-----------|------------|
| SPIFFE/SPIRE | Production-grade; used by Netflix, Uber, Google; Go libraries available (`spiffe/go-spiffe`) | LOW |
| mTLS in Go | Standard `crypto/tls` config with SPIRE-issued X.509 SVIDs | LOW |
| Dev experience | SPIRE Docker Compose setup is well-documented; adds ~30s to cold start | MEDIUM |
| Per-request authz | Chi middleware pattern for checking OIDC claims + SPIFFE SVIDs | LOW |
| Operational complexity | SPIRE server + agent deployment; requires infrastructure planning | MEDIUM |

**Overall: FEASIBLE but DEFER** -- mTLS between Brain and FB-OS is straightforward. Full SPIFFE/SPIRE deployment is Phase 2. Start with OIDC-based service-to-service auth (JWT bearer tokens) in Phase 1.

### Pillar 5: Lit Web Components Migration

| Criterion | Assessment | Risk Level |
|-----------|-----------|------------|
| Vite support | Vite natively supports Lit; same build tooling as FB-OS | LOW |
| Component library | No existing Lit component library for admin dashboards; must build from scratch | MEDIUM |
| State management | `@lit-labs/preact-signals` is proven in FB-OS; same pattern | LOW |
| Migration path | Lit Custom Elements can coexist with React in the same page (web standards) | LOW |
| CSS design system | Must define design tokens; can reuse FB-OS's GableLBM design system | LOW |

**Overall: FEASIBLE** -- Incremental migration is low-risk. New pages in Lit; old pages migrated opportunistically. Estimated effort: 2-3 weeks per major page.

---

## Dependency Analysis

```
OIDC Provider ─────────────────────────────────────────── Foundation
      │
      ├── MCP Server Registry (needs OIDC for per-server auth tokens)
      │         │
      │         ├── Maestro AI (needs MCP tools to call)
      │         │         │
      │         │         └── A2A Webhooks (needs Maestro to emit tasks)
      │         │
      │         └── Lit Admin UI (needs registry API to browse MCP servers)
      │
      └── FB-OS Integration (needs Brain OIDC tokens)
```

**Critical path:** OIDC Provider -> MCP Registry -> Maestro -> A2A Webhooks

---

## Resource Estimates

| Component | Effort (weeks) | Go LOC Estimate | Dependencies |
|-----------|---------------|-----------------|--------------|
| OIDC Provider | 4-6 | 3,000-4,000 | zitadel/oidc, go-jose/v4, PostgreSQL migrations |
| MCP Server Registry | 3-4 | 2,000-2,500 | modelcontextprotocol/go-sdk, PostgreSQL JSONB |
| Maestro AI Orchestrator | 5-7 | 2,500-3,500 | Anthropic SDK, MCP registry, intent taxonomy |
| A2A Webhook Client | 2-3 | 1,000-1,500 | go-jose/v4, retryablehttp |
| Lit Admin UI (new pages) | 3-4 | 2,000 TS LOC | Lit 3.0, Signals |
| L7 Zero-Trust (Phase 1) | 2-3 | 500-800 | go-spiffe, mTLS config |
| Migration + backward compat | 2-3 | 1,000-1,500 | Database migrations, API adapters |
| **Total** | **21-30 weeks** | **12,000-15,300** | |

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| OIDC provider doesn't pass certification | Medium | High | Use certified `zitadel/oidc` library; run conformance tests early |
| Claude intent classification accuracy <90% | Low | Medium | Three-tier fallback; grow training corpus from production data |
| FB-OS delays A2A webhook receiver | Medium | High | Brain can fall back to direct API calls (legacy pattern) during transition |
| QuickBooks CloudEvents deadline (May 2026) | High | Medium | Prioritize QuickBooks MCP server adapter; deadline is 6 weeks away |
| MCP spec changes breaking Go SDK | Low | Medium | Pin SDK version; monitor spec changelog |
| Lit migration slower than expected | Medium | Low | React UI remains functional; migration is non-blocking |
