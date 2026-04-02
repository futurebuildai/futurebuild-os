# Assumptions Register

**Document ID:** AG-00-ASSUMPTIONS
**Created:** 2026-04-02
**Status:** Stage 00 Complete

---

## Critical Assumptions

| ID | Assumption | Risk if Wrong | Validation Method | Status |
|----|-----------|---------------|-------------------|--------|
| A-01 | FB-OS will accept Brain-issued OIDC tokens in place of Clerk JWTs | FB-OS auth breaks entirely | Prototype OIDC flow -> validate FB-OS JWKS middleware accepts Brain tokens | UNVALIDATED |
| A-02 | The `zitadel/oidc` Go library provides sufficient OP (provider) functionality to build a custom OIDC provider without deploying a separate Ory Hydra/Zitadel instance | Architecture becomes more complex if we need sidecar | Build PoC OIDC provider with zitadel/oidc v3 | UNVALIDATED |
| A-03 | Legacy `SystemDefinition.ActionDefinition` maps cleanly to MCP Tool definitions | Significant refactoring if mapping is lossy | Map all 21 actions to MCP Tool JSON Schema manually | UNVALIDATED |
| A-04 | The official MCP Go SDK (`modelcontextprotocol/go-sdk`) is production-ready for server implementation | Must fall back to community SDK or custom implementation | Review SDK maturity, test coverage, open issues | UNVALIDATED |
| A-05 | A2A protocol v0.3+ webhook signing is compatible with our JWT/JWS infrastructure | May need separate signing key management | Implement Agent Card signing PoC with go-jose/v4 | UNVALIDATED |
| A-06 | Claude (Anthropic) can reliably classify top-20 construction intents with >90% accuracy | Maestro becomes unreliable; may need fine-tuned model | Create evaluation dataset of 200 utterances; benchmark | UNVALIDATED |
| A-07 | QuickBooks webhook migration to CloudEvents format (May 2026 deadline) can be handled by Brain without breaking existing flows | QuickBooks integration fails after deadline | Review CloudEvents spec; implement adapter | UNVALIDATED |
| A-08 | FB-OS will implement A2A webhook receiver endpoints for Brain-emitted tasks | Brain has no execution target for workflows | Coordinate with FB-OS team; define A2A contract | UNVALIDATED |
| A-09 | PostgreSQL 16+ is sufficient for OIDC token/session storage without a dedicated Redis cache for token lookups | Token validation latency exceeds acceptable threshold | Load test OIDC token validation at 1000 RPS | UNVALIDATED |
| A-10 | Migrating Hub frontend from React 19 to Lit 3.0 can be done incrementally (page by page) without a big-bang rewrite | Frontend migration blocks other work | Verify Vite can serve mixed React + Lit components during migration | UNVALIDATED |
| A-11 | Existing `hub_users` table can be extended with OIDC-required fields (password hash, PKCE, consent records) without breaking existing auth flows | Migration breaks existing admin users | Write reversible migration; test with existing data | UNVALIDATED |
| A-12 | SPIFFE/SPIRE can be deployed for local development using Docker Compose without excessive complexity | Dev experience degrades; developers skip zero-trust testing | Test SPIRE Docker Compose setup; measure cold-start time | UNVALIDATED |
| A-13 | The 7 existing integration API contracts (GableERP, XUI, LocalBlue, QuickBooks, 1Build, Gmail, Outlook) remain stable through the revamp | Integration tests fail; must chase API changes | Pin API versions in integration tests; monitor changelogs | ASSUMED STABLE |

---

## Architectural Assumptions

| ID | Assumption | Implication |
|----|-----------|-------------|
| AA-01 | Brain and FB-OS remain separate deployments (polyrepo) | A2A webhooks must be network-resilient; no shared-memory shortcuts |
| AA-02 | Brain is the ONLY identity provider in the ecosystem | No federation with external IdPs in MVP (can add later via OIDC federation) |
| AA-03 | Maestro AI runs inside Brain's Go process (not a sidecar LLM) | Claude API calls happen server-side; Brain needs ANTHROPIC_API_KEY |
| AA-04 | MCP transport is Streamable HTTP (not stdio) for production | Clients connect via HTTP; no requirement for persistent connections |
| AA-05 | The revamp preserves backward compatibility with existing API routes during migration | Existing integrations continue working while new OIDC/MCP/A2A endpoints are added |

---

## Business Assumptions

| ID | Assumption | Risk if Wrong |
|----|-----------|---------------|
| BA-01 | Residential custom builders (5-50 employees) are the primary user segment | Feature prioritization misaligned with actual users |
| BA-02 | QuickBooks Online is the dominant accounting system for target users | Accounting integration work wasted if users prefer Sage/Xero |
| BA-03 | Users prefer natural-language interaction over form-based workflows | Maestro AI investment yields low adoption |
| BA-04 | Ecosystem billing (plan-gated features) remains in Brain, not FB-OS | Revenue model architecture must support tiered access |
