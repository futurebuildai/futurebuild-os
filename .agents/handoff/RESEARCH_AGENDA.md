# Research Agenda

**Document ID:** AG-00-RESEARCH-AGENDA
**Created:** 2026-04-02
**Status:** Stage 00 Complete

---

## Research Domains

### Domain 1: OIDC Provider Implementation in Go
- **Questions:** Build custom OIDC provider atop `zitadel/oidc` library vs. embed Ory Hydra vs. deploy Zitadel as sidecar?
- **Sources:** Zitadel OIDC Go library (certified by OpenID Foundation), Ory Hydra GitHub, `go-jose/v4` for JWT signing, `lestrrat-go/jwx` for JWK management, OIDC conformance test suite
- **Legacy context:** From `reference-vault/FB-Brain/internal/hub/auth_handlers.go` -- existing session cookie auth must be migrated to OIDC flows. From `reference-vault/futurebuild-os/CLAUDE.md` -- FB-OS currently uses Clerk for JWT validation via JWKS; Brain must replace Clerk as the IdP.

### Domain 2: MCP Server Registry Architecture
- **Questions:** How does MCP Tool schema map to legacy `ActionDefinition`? How does MCP Resource map to `EntitySchema`? How to implement dynamic server registration?
- **Sources:** MCP specification (2025-11-25 revision), official Go SDK (`modelcontextprotocol/go-sdk`), MCP Registry API (v0.1, September 2025), `mark3labs/mcp-go` community SDK
- **Legacy context:** From `reference-vault/FB-Brain/internal/registry/types.go` -- `ActionDefinition` has `InputSchema`/`OutputSchema` as `[]FieldDef` which maps to MCP Tool `inputSchema` (JSON Schema). `TriggerDefinition` maps to MCP notification/subscription patterns.

### Domain 3: A2A Protocol for Signed Webhooks
- **Questions:** How to implement Agent Cards for each integration? How to sign webhook payloads with JWS? How does A2A task lifecycle map to MaterialsFlow/LaborFlow state machine?
- **Sources:** A2A Protocol Specification (a2a-protocol.org, v0.3+), GitHub `a2aproject/A2A`, Red Hat A2A security guide
- **Legacy context:** From `reference-vault/FB-Brain/internal/orchestrator/materials_flow.go` -- the existing flow creates XUI feed cards and logs events; A2A webhooks replace this with standards-compliant async task updates.

### Domain 4: QuickBooks Online API Evolution
- **Questions:** Impact of refresh token 5-year policy on legacy `quickbooks.go`? CloudEvents webhook format migration by May 2026? Reconnect URL field requirements?
- **Sources:** Intuit Developer Portal, QuickBooks OAuth2 documentation, CloudEvents specification
- **Legacy context:** From `reference-vault/FB-Brain/internal/registry/quickbooks.go` -- existing definition uses standard OAuth2 with `com.intuit.quickbooks.accounting` scope. Must update for new token policies.

### Domain 5: Construction ERP Competitive Landscape
- **Questions:** What market gaps exist that FutureBuild can exploit? What are Procore/Buildertrend pricing models? What do users complain about?
- **Sources:** G2, Capterra, TrustRadius reviews; Reddit r/construction; market research reports
- **Legacy context:** From `reference-vault/FB-Brain/internal/registry/` -- the 7 existing integrations target a specific niche (custom residential builders using XUI + GableERP + LocalBlue + QuickBooks) that larger platforms ignore.

### Domain 6: AI Co-pilot Architecture
- **Questions:** How to implement probabilistic intent classification in Go? Hybrid probabilistic/deterministic architecture patterns? What are the top-20 construction intents?
- **Sources:** Enterprise AI architecture reports, intent classification literature, Claude tool-use patterns
- **Legacy context:** From `reference-vault/futurebuild-os/CLAUDE.md` -- FB-OS already has a dual chat orchestrator (regex fallback + Claude Opus). Brain's Maestro extends this pattern with A2A webhook emission.

### Domain 7: L7 Zero-Trust Implementation
- **Questions:** SPIFFE/SPIRE deployment for Go microservices? mTLS with Chi router? Per-request authorization middleware patterns?
- **Sources:** NIST SP 800-207, SPIFFE specification, BeyondCorp papers, Go SPIFFE libraries
- **Legacy context:** From `reference-vault/FB-Brain/internal/config/config.go` -- current config uses simple API keys (`IntegrationKey`). Must evolve to SPIFFE SVIDs and mTLS.

### Domain 8: Lit Web Components for Admin UI
- **Questions:** Migration path from React 19 to Lit 3.0? Signals-based state management? Dashboard component patterns?
- **Sources:** Lit documentation, FB-OS frontend patterns (already uses Lit), web components enterprise case studies
- **Legacy context:** From `reference-vault/futurebuild-os/CLAUDE.md` -- FB-OS frontend already uses Lit 3.0 + TypeScript + Signals. Brain's Hub UI must align for ecosystem consistency.

---

## Research Prioritization

| Priority | Domain | Rationale |
|----------|--------|-----------|
| P0 | OIDC Provider | Foundation -- everything else depends on identity |
| P0 | MCP Server Registry | Core value proposition -- tool discovery |
| P1 | A2A Protocol | Differentiator -- signed webhook orchestration |
| P1 | AI Co-pilot Architecture | Differentiator -- Maestro intent parsing |
| P1 | L7 Zero-Trust | Security requirement -- non-negotiable |
| P2 | QuickBooks API | Integration continuity -- must not break |
| P2 | Competitive Landscape | Market positioning |
| P3 | Lit Web Components | UI modernization -- can follow backend |
