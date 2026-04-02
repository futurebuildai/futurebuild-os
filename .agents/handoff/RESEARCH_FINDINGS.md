# Research Findings

**Document ID:** AG-01-RESEARCH
**Created:** 2026-04-02
**Status:** Stage 01 Complete
**Research Queries Executed:** 20+ web searches

---

## Legacy Vault Analysis

### Registry Architecture (Source: `reference-vault/FB-Brain/internal/registry/`)

The legacy registry is a **compile-time, in-memory map** of 7 `SystemDefinition` structs. Key observations:

1. **Static registration** -- `registry.New()` calls `r.register()` for each system at startup. No runtime registration capability.
2. **Flat type system** -- `FieldDef` uses string-typed `Type` field ("string", "number", "boolean", "object", "array") without JSON Schema validation. This is semantically identical to MCP Tool `inputSchema` but lacks the expressiveness of JSON Schema 2020-12.
3. **Auth model heterogeneity** -- Three distinct auth types: `api_key` (Gable, XUI, 1Build), `oauth2` (QuickBooks, Gmail, Outlook), and `host_header` (LocalBlue). The MCP equivalent would be per-server auth configuration in the MCP Server manifest.
4. **Action-Trigger symmetry** -- Every system has both actions (outbound operations) and triggers (inbound events). This maps directly to MCP Tools (actions) and MCP notifications/subscriptions (triggers).

### Orchestration Patterns (Source: `reference-vault/FB-Brain/internal/orchestrator/`)

1. **Hardcoded domain logic** -- `roofingScope` is a compile-time constant with specific SKUs. The revamp must make scope definition dynamic (AI-inferred or user-specified).
2. **Cross-system identity resolution** -- `AccountLink` maps `XUIOrgID` to `GableCustomerID` to `LocalBlueSiteID`. This is a manual identity federation pattern that OIDC replaces with subject-claim-based identity.
3. **Event sourcing (partial)** -- `IntegrationEvent` with `source_system`/`target_system`/`payload` provides audit trail but is not a full event store. A2A task lifecycle provides richer state tracking.
4. **Post-approval convergence** -- `CheckPostApproval()` waits for both MaterialsFlow AND LaborFlow to complete before creating the delivery card. This is a simple join pattern that Maestro can express as a probabilistic workflow.

### Hub Admin UI (Source: `reference-vault/FB-Brain/internal/hub/hub.go`)

1. **Session-based auth** -- Cookie sessions via `hub_sessions` table. Must migrate to OIDC token-based auth.
2. **Plan-gated features** -- `RequirePlan("all_integrations")` middleware. This billing model must be preserved in the OIDC claims/scopes.
3. **Engine handlers** -- `WorkflowExecutor` with validate/dry-run/execute pipeline. This pattern is preserved but execution moves to A2A webhooks to FB-OS.
4. **EventBus (SSE)** -- Real-time events via Server-Sent Events. Can be preserved alongside MCP's notification mechanism.

### FB-OS Downstream System (Source: `reference-vault/futurebuild-os/CLAUDE.md`)

1. **Current auth:** Clerk JWT via JWKS + Magic Link for portal contacts + API Key for Brain integration
2. **Frontend:** Lit 3.0 + TypeScript + Signals -- Brain's UI must align
3. **Dual chat orchestrator:** Regex fallback + Claude Opus -- Maestro extends this pattern
4. **A2A package exists:** `pkg/a2a/` already present in FB-OS, indicating readiness for A2A integration
5. **Worker architecture:** Asynq (Redis) for async jobs -- Brain's A2A webhooks will trigger these

---

## OIDC Provider Research

### Comparison: Build Custom vs. Ory Hydra vs. Zitadel

| Criterion | Custom (zitadel/oidc lib) | Ory Hydra (sidecar) | Zitadel (full platform) |
|-----------|--------------------------|--------------------|-----------------------|
| Deployment complexity | Low (embedded in Brain binary) | Medium (separate process, ports 4444/4445) | High (separate service + CockroachDB/PG) |
| Go integration | Native (library) | SDK client calls | SDK client calls |
| Customization | Full control over consent/login UI | Consent/login via HTTP redirect to your app | Custom through Zitadel actions (JS) |
| OIDC certification | Zitadel oidc lib is certified | Hydra is certified | Zitadel is certified |
| User management | Must build (evolve hub_users) | Pairs with Ory Kratos (another sidecar) | Built-in |
| Multi-tenancy | Must build | Via OAuth2 clients per tenant | Built-in |
| Resource footprint | Minimal (library) | ~100MB RAM per instance | ~500MB+ RAM |

**Recommendation:** Build custom OIDC provider using the `zitadel/oidc` v3 library. Rationale:
- Brain already has a user database (`hub_users`), session management, and org management
- Embedding the OIDC provider keeps the deployment as a single binary
- The `zitadel/oidc` library is OpenID Certified and provides both RP and OP implementations
- Ory Hydra adds operational complexity (separate process, admin API) for a capability we can build directly
- Full Zitadel is overkill -- we don't need its UI, RBAC engine, or CockroachDB dependency

Sources: [Zitadel OIDC Library](https://github.com/zitadel/oidc), [Ory Hydra](https://github.com/ory/hydra), [Zitadel vs Casdoor vs Authentik](https://www.pkgpulse.com/blog/zitadel-vs-casdoor-vs-authentik-open-source-iam-2026)

### Go JWT/JWKS Libraries

| Library | Purpose | Recommendation |
|---------|---------|----------------|
| `go-jose/go-jose/v4` | JWS/JWE/JWK primitives | Use for OIDC token signing and JWKS endpoint |
| `lestrrat-go/jwx/v3` | Complete JWx suite with auto-refresh JWKS | Use for A2A webhook signature verification |
| `coreos/go-oidc/v3` | OIDC client (RP) only | Use in FB-OS to validate Brain tokens |

Sources: [go-jose v4](https://pkg.go.dev/github.com/go-jose/go-jose/v4/jwt), [lestrrat-go/jwx](https://github.com/lestrrat-go/jwx)

---

## MCP Server Registry Research

### Spec Mapping: Legacy -> MCP

From the MCP specification (2025-11-25 revision):

| Legacy Concept | MCP Equivalent | Mapping Notes |
|---------------|----------------|---------------|
| `SystemDefinition` | MCP Server | Each integration becomes an MCP server with its own capabilities |
| `ActionDefinition` | MCP Tool | `InputSchema`/`OutputSchema` -> JSON Schema `inputSchema`; `HTTPMethod`+`PathTemplate` become tool implementation details |
| `TriggerDefinition` | MCP Notification / Subscription | `EventType` (webhook/polling/manual) maps to MCP notification channels |
| `EntitySchema` | MCP Resource | Entities become addressable resources with URI templates |
| `FieldDef` | JSON Schema property | `Type` string -> JSON Schema `type`; `Required` -> `required` array; `Description` preserved |
| `Registry.Get(slug)` | `tools/list` filtered by server | Server-scoped tool discovery |
| `Registry.GetAction(sys, action)` | `tools/call` | Direct tool invocation |

### Official MCP Go SDK

The `modelcontextprotocol/go-sdk` (maintained with Google) provides:
- `mcp.Server` for creating MCP servers
- `mcp.AddTool()` for registering tools with JSON Schema
- Transport support for stdio and Streamable HTTP
- The SDK is in active development; suitable for production use

**Recommendation:** Use the official Go SDK for MCP server implementation. Each legacy SystemDefinition becomes an MCP server. The Brain hosts a registry of MCP servers, each exposing their actions as MCP Tools.

Sources: [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk), [MCP Spec](https://modelcontextprotocol.io/specification/2025-11-25), [MCP Registry](https://github.com/modelcontextprotocol/registry)

---

## A2A Protocol Research

### Specification Analysis (v0.3+)

Key A2A concepts that map to Brain's revamp:

1. **Agent Card** -- JSON document describing an agent's capabilities, skills, and auth requirements. Each Brain integration gets an Agent Card. Canonicalized per RFC 8785 and signed with JWS.
2. **Task lifecycle** -- `submitted` -> `working` -> `input-required` -> `completed`/`failed`. Maps to the existing RFQ status flow: `rfq_sent` -> `bid_received` -> `approved`/`ordered`.
3. **Push Notifications** -- Async updates via HTTP POST to webhook URLs. This replaces Brain's direct API calls to XUI/GableERP with standards-compliant webhooks.
4. **Authentication** -- A2A supports OAuth2 bearer tokens and API keys. Brain's OIDC provider issues tokens that A2A webhook receivers validate.

**Recommendation:** Implement A2A v0.3 for Brain -> FB-OS communication. Brain emits signed A2A webhooks; FB-OS receives and executes deterministically. This decouples probabilistic intent parsing (Brain) from deterministic execution (OS).

Sources: [A2A Specification](https://a2a-protocol.org/latest/specification/), [A2A GitHub](https://github.com/a2aproject/A2A), [Red Hat A2A Security](https://developers.redhat.com/articles/2025/08/19/how-enhance-agent2agent-security)

---

## QuickBooks API Research

### Breaking Changes Timeline

| Date | Change | Impact on Brain |
|------|--------|----------------|
| Oct 2028 | First refresh tokens expire (5-year policy) | Must implement reconnect flow |
| Jan 2026 | Reconnect URL field mandatory in dev portal | Configuration requirement |
| May 15, 2026 | CloudEvents format required for webhooks | Must update webhook receiver to parse CloudEvents |

**Critical:** The CloudEvents webhook migration deadline (May 2026) is imminent. The legacy `quickbooks.go` trigger definitions (`invoice_created`, `payment_received`) must be updated to expect CloudEvents envelope format.

**Recommendation:** Implement CloudEvents adapter in the MCP Server for QuickBooks before May 2026. Use `cloudevents/sdk-go` for Go-native CloudEvents support.

Sources: [Intuit Refresh Token Policy](https://blogs.intuit.com/2025/11/12/important-changes-to-refresh-token-policy/), [QuickBooks OAuth2](https://developer.intuit.com/app/developer/qbo/docs/develop/authentication-and-authorization/oauth-2.0)

---

## AI Co-pilot Architecture Research

### Intent Classification Approaches

| Approach | Pros | Cons | Recommendation |
|----------|------|------|----------------|
| Claude tool-use (function calling) | Native tool selection; no training needed; probabilistic | API latency (~500ms); cost per call | **Primary** -- Maestro uses Claude with MCP tools as function definitions |
| Fine-tuned classifier | Fast inference; low cost per call | Training data required; model management | **Future** -- once intent corpus is large enough |
| Regex fallback | Zero latency; zero cost; deterministic | Brittle; limited coverage | **Fallback** -- mirrors FB-OS's dual orchestrator pattern |

**Recommended architecture:** Maestro uses a **three-tier classification pipeline**:
1. **Regex fast-path** -- catches exact-match commands ("approve quote RFQ-123")
2. **Claude tool-use** -- probabilistic intent -> MCP tool selection with confidence scores
3. **Human-in-the-loop** -- if confidence < threshold, present options to user for confirmation

This mirrors the hybrid probabilistic/deterministic pattern observed in enterprise AI deployments.

Sources: [Enterprise AI Architecture](https://blog.scottlogic.com/2025/06/03/navigating-enterprise-ai-architecture.html), [Intent Classification](https://medium.com/aimonks/intent-classification-generative-ai-based-application-architecture-3-79d2927537b4), [State of GenAI in Enterprise](https://menlovc.com/perspective/2025-the-state-of-generative-ai-in-the-enterprise/)

---

## L7 Zero-Trust Research

### Implementation Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Workload Identity | SPIFFE/SPIRE | Issue SVIDs (X.509 or JWT) to Brain and FB-OS workloads |
| Transport Security | mTLS | Encrypt and authenticate all service-to-service traffic |
| Authorization | OPA (Open Policy Agent) or Cedar | Per-request policy evaluation |
| API Gateway | None (Chi middleware) | L7 policy enforcement at the application layer |
| Key Management | SPIRE + go-jose/v4 | Automatic key rotation; short-lived certificates |

**Recommendation:** Implement SPIFFE/SPIRE for workload identity with mTLS between Brain and FB-OS. Use Chi middleware for per-request authorization based on OIDC claims and SPIFFE SVIDs. Avoid adding an API gateway -- keep the architecture simple with in-process middleware.

**Phasing:** Start with mTLS between Brain and FB-OS in Phase 1. Add SPIRE attestation in Phase 2. Add OPA policies in Phase 3.

Sources: [NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final), [Machine Identity mTLS + SPIFFE](https://petronellatech.com/blog/machine-identity-is-the-new-perimeter-mtls-spiffe-for-zero-trust/), [Zero Trust 2026](https://dev.to/walid_azrour_0813f6b60398/zero-trust-architecture-the-security-model-every-developer-needs-to-understand-in-2026-4c03)

---

## Lit Web Components Research

### Migration from React to Lit

| Aspect | React 19 (Current) | Lit 3.0 (Target) |
|--------|-------------------|------------------|
| State management | TanStack Query + Context | Signals (@lit-labs/preact-signals) |
| Component model | JSX + hooks | Template literals + decorators |
| Styling | Tailwind CSS 4 | CSS custom properties (design tokens) |
| Bundle size | ~45KB (React) + ~15KB (ReactDOM) | ~5KB (Lit) |
| Framework coupling | React-only | Framework-agnostic Web Components |
| FB-OS compatibility | Different stack | Same stack (Lit 3.0 + Signals) |

**Recommendation:** Migrate Hub admin UI from React to Lit 3.0 for ecosystem consistency with FB-OS. Use the same component architecture: `FBElement` base class, Signals for reactive state, CSS custom properties for theming. The migration can be incremental -- Lit components render as standard Custom Elements that can coexist with React during transition.

Sources: [Lit 3.0 Enterprise](https://markaicode.com/web-components-2025-lit-stencil-enterprise/), [FB-OS Frontend Architecture](reference-vault/futurebuild-os/CLAUDE.md)
