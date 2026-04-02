# Build vs. Buy Analysis

**Document ID:** AG-03-BUILD-VS-BUY
**Created:** 2026-04-02
**Status:** Stage 03 Complete

---

## Decision 1: OIDC Identity Provider

| Option | Build | Buy / Embed |
|--------|-------|-------------|
| **A: Custom with zitadel/oidc lib** | Build OP logic; Build consent UI; Build key rotation | Buy: zitadel/oidc library (certified OP); Buy: go-jose/v4 (JWK/JWS) |
| **B: Ory Hydra sidecar** | Build login/consent redirect handler; Build admin API calls | Buy: Ory Hydra (full OIDC server); Optionally buy: Ory Kratos (user mgmt) |
| **C: Zitadel platform** | Build integration adapter | Buy: Zitadel (full IdP platform + CockroachDB/PG) |
| **D: Auth0 / Clerk SaaS** | Build nothing (delegate entirely) | Buy: SaaS subscription ($0.05-0.23/MAU) |

### Evaluation

| Criterion | Weight | A (zitadel/oidc) | B (Hydra) | C (Zitadel) | D (SaaS) |
|-----------|--------|-------------------|-----------|-------------|----------|
| Deployment simplicity | 25% | 9 (single binary) | 5 (sidecar) | 3 (platform) | 10 (hosted) |
| Architectural control | 25% | 10 (full control) | 7 (consent hook) | 4 (limited) | 2 (none) |
| OIDC compliance | 20% | 8 (certified lib) | 10 (certified) | 10 (certified) | 10 (certified) |
| Cost at scale | 15% | 10 (free) | 10 (free) | 10 (free) | 3 ($$ at scale) |
| Development effort | 15% | 5 (moderate) | 7 (less code) | 8 (minimal) | 9 (none) |
| **Weighted Score** | | **8.45** | **7.55** | **6.60** | **6.35** |

### Decision: **BUILD with zitadel/oidc library (Option A)**

**Rationale:**
1. Brain already has user management (`hub_users`), org management, session infrastructure, and magic link auth. Building OIDC on top is incremental, not greenfield.
2. Single binary deployment is a hard architectural constraint (from TECH_STACK.md).
3. The `zitadel/oidc` library is OpenID Certified, so we get compliance without operating a separate service.
4. SaaS options (Auth0, Clerk) create vendor lock-in and are expensive at scale. FB-OS already has a Clerk dependency that this revamp is specifically designed to eliminate.
5. Ory Hydra is the backup plan if custom OP development stalls or certification requirements tighten.

---

## Decision 2: MCP Server Implementation

| Option | Build | Buy / Use |
|--------|-------|-----------|
| **A: Official Go SDK** | Build MCP servers using official SDK | Use: modelcontextprotocol/go-sdk |
| **B: Community SDK (mcp-go)** | Build MCP servers using community SDK | Use: mark3labs/mcp-go |
| **C: Custom from spec** | Build MCP JSON-RPC handling from scratch | Use: nothing (raw implementation) |

### Decision: **USE Official Go SDK (Option A)**

**Rationale:**
1. Maintained in collaboration with Google; highest probability of long-term support.
2. Implements the full MCP spec including Streamable HTTP transport.
3. Community SDK (mcp-go) is a viable fallback if the official SDK has gaps.
4. Building from scratch is not justified when a certified SDK exists.

---

## Decision 3: AI Intent Classification

| Option | Build | Buy / Use |
|--------|-------|-----------|
| **A: Claude tool-use** | Build Maestro orchestrator; define tool schemas | Use: Anthropic Claude API (usage-based) |
| **B: Fine-tuned model** | Build training pipeline; collect/label data; train classifier | Use: Vertex AI / Anthropic fine-tuning API |
| **C: Rule-based / Regex** | Build intent regex patterns (like FB-OS's fallback orchestrator) | Use: nothing |

### Decision: **BUILD hybrid A+C (Claude primary, regex fallback)**

**Rationale:**
1. Claude tool-use provides immediate, high-accuracy intent classification without training data.
2. Regex fallback (already proven in FB-OS per `reference-vault/futurebuild-os/CLAUDE.md`) provides zero-latency exact-match for common commands.
3. Fine-tuned model (Option B) is a Phase 3 optimization once production intent data is available.
4. Cost is manageable: ~$0.01-0.05 per classification at Claude Sonnet pricing.

---

## Decision 4: A2A Webhook Infrastructure

| Option | Build | Buy / Use |
|--------|-------|-----------|
| **A: Custom A2A client** | Build webhook emitter, JWS signer, retry logic | Use: go-jose/v4, retryablehttp |
| **B: A2A SDK (when available)** | Build adapter to SDK | Use: Official A2A SDK (not yet available for Go) |
| **C: Generic webhook service** | Integrate external webhook delivery service | Buy: Svix, Hookdeck |

### Decision: **BUILD custom A2A client (Option A)**

**Rationale:**
1. A2A spec is straightforward HTTP + JSON-RPC; no need for an SDK.
2. `go-jose/v4` handles JWS signing; `hashicorp/go-retryablehttp` handles retry.
3. External webhook services (Svix, Hookdeck) add unnecessary dependency and cost.
4. When an official Go A2A SDK becomes available, migration from custom client is low-effort.

---

## Decision 5: Admin Dashboard Framework

| Option | Build | Buy / Use |
|--------|-------|-----------|
| **A: Lit 3.0 (ecosystem alignment)** | Build component library; build dashboard pages | Use: Lit 3.0, @lit-labs/preact-signals |
| **B: Keep React 19** | Continue building in current stack | Use: React 19, TanStack Query, Radix UI |
| **C: Hybrid (Lit for new, React for existing)** | Build new pages in Lit; migrate old pages later | Use: Both frameworks during transition |

### Decision: **BUILD with Lit 3.0 via incremental migration (Option C then A)**

**Rationale:**
1. FB-OS uses Lit 3.0 (from `reference-vault/futurebuild-os/CLAUDE.md`). Ecosystem consistency requires Brain to align.
2. TECH_STACK.md explicitly specifies "Vite + Lit (vanilla Web Components)" for Brain's frontend.
3. Incremental migration (Option C) avoids blocking product delivery with a full rewrite.
4. New pages (MCP Registry Browser, OIDC Client Manager, A2A Dashboard) start in Lit from day one.

---

## Decision 6: Zero-Trust Implementation

| Option | Build | Buy / Use |
|--------|-------|-----------|
| **A: SPIFFE/SPIRE** | Deploy SPIRE server/agent; integrate go-spiffe | Use: SPIFFE/SPIRE (open source) |
| **B: Istio service mesh** | Deploy Istio sidecar proxies | Use: Istio (heavy) |
| **C: OIDC-only (JWT bearer)** | Use Brain-issued JWTs for service-to-service auth | Use: Brain's own OIDC provider |
| **D: Phased (C then A)** | Start with JWT bearer; add SPIRE later | Use: Both, sequentially |

### Decision: **BUILD phased approach (Option D)**

**Rationale:**
1. Phase 1: OIDC JWT bearer tokens for Brain -> FB-OS auth (immediate, no new infrastructure).
2. Phase 2: SPIFFE/SPIRE for workload identity (after OIDC is stable).
3. Istio is overkill for a two-service architecture.
4. This aligns with the dependency graph: OIDC must be built first anyway.

---

## Summary of Decisions

| Component | Decision | Rationale |
|-----------|----------|-----------|
| OIDC Provider | BUILD with zitadel/oidc library | Single binary; certified library; evolves existing Hub auth |
| MCP Registry | USE official Go SDK | Google-backed; full spec coverage |
| AI Intent | BUILD hybrid Claude + regex | Immediate accuracy; zero training data needed |
| A2A Webhooks | BUILD custom client | Simple spec; standard Go libraries |
| Admin UI | BUILD in Lit 3.0 (incremental) | Ecosystem alignment with FB-OS |
| Zero-Trust | PHASED: OIDC JWT -> SPIFFE/SPIRE | Pragmatic; no premature infrastructure |
