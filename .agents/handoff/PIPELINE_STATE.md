# Pipeline State

**Document ID:** AG-PIPELINE
**Created:** 2026-04-02
**System:** FutureBuild Brain (System of Connection)
**Orchestrator:** Antigravity

---

## Stage Status

| Stage | Name | Status | Artifacts | Gate |
|-------|------|--------|-----------|------|
| 00 | Vision Intake | COMPLETE | VISION_BRIEF.md, RESEARCH_AGENDA.md, ASSUMPTIONS_REGISTER.md | Passed |
| 01 | Deep Research | COMPLETE | RESEARCH_FINDINGS.md, COMPETITIVE_LANDSCAPE.md | Passed |
| 02 | User Research | COMPLETE | PERSONAS.md, JTBD_ANALYSIS.md, USER_JOURNEYS.md | Passed |
| 03 | Solution Design | COMPLETE | SOLUTION_CANDIDATES.md, FEASIBILITY_ASSESSMENT.md, BUILD_VS_BUY.md | Passed |
| 04 | Scope & Prioritization | COMPLETE | SCOPE_DEFINITION.md, METRICS_FRAMEWORK.md | **PAUSED AT APPROVAL GATE 1** |
| 05 | Design System | COMPLETE | DESIGN_SYSTEM.md, INFORMATION_ARCHITECTURE.md | GableLBM Industrial Dark tokens, FBBaseElement, Tailwind CSS 4, 3 surfaces (Hub Admin, Maestro AI, Integration Registry) |
| 06 | Product Specification | COMPLETE | PRD.md | 16 user stories across 5 journeys, Given/When/Then acceptance criteria, 8 NFR categories, full traceability matrix |
| 07 | Architecture Spec | COMPLETE | ARCHITECTURE.md | Go package structure, PostgreSQL schema (OIDC + MCP + workflow tables), OIDC provider endpoints, MCP registry with 7 servers, Maestro 3-tier classifier, A2A JWS client, Lit admin UI, CI/CD pipeline |

---

## Approval Gate 1: Scope & Architecture Review

### Status: AWAITING APPROVAL

### Items Requiring Sign-off

1. **OIDC Provider Approach:** Build custom with `zitadel/oidc` library (not Ory Hydra sidecar, not Zitadel platform)
2. **MCP SDK Selection:** Official `modelcontextprotocol/go-sdk` (not community `mcp-go`)
3. **Maestro Architecture:** Three-tier classification (regex -> Claude tool-use -> human confirmation)
4. **A2A Implementation:** Custom client with `go-jose/v4` JWS signing (not external webhook service)
5. **Admin UI Framework:** Lit 3.0 via incremental migration from React 19
6. **Zero-Trust Phasing:** OIDC JWT bearer tokens first, SPIFFE/SPIRE second
7. **MVP Scope:** 12-week roadmap covering OIDC + 4 MCP servers + Maestro AI + A2A webhooks
8. **Walking Skeleton:** 2-week proof of end-to-end flow (auth -> tool discovery -> tool call -> webhook)

### Approval Criteria

- [ ] Stakeholder agrees with OIDC provider build approach
- [ ] Stakeholder agrees with MVP scope and timeline
- [ ] Stakeholder agrees with Build vs. Buy decisions
- [ ] Stakeholder confirms FB-OS team readiness for A2A webhook receiver
- [ ] Stakeholder confirms QuickBooks CloudEvents priority (May 2026 deadline)

---

## Key Decisions Made

| Decision | Choice | Rationale | Document |
|----------|--------|-----------|----------|
| OIDC approach | Custom with zitadel/oidc | Single binary; certified library; evolves existing auth | BUILD_VS_BUY.md |
| MCP SDK | Official Go SDK | Google-backed; full spec | BUILD_VS_BUY.md |
| AI classification | Claude tool-use + regex fallback | Immediate accuracy; no training data needed | BUILD_VS_BUY.md |
| A2A client | Custom build | Simple spec; standard Go libraries | BUILD_VS_BUY.md |
| Frontend | Lit 3.0 incremental | FB-OS alignment; TECH_STACK.md mandate | BUILD_VS_BUY.md |
| Zero-Trust | Phased (JWT -> SPIFFE) | Pragmatic; OIDC first | BUILD_VS_BUY.md |

---

## Legacy Vault Files Consulted

| File | Key Insight |
|------|-------------|
| `reference-vault/FB-Brain/CLAUDE.md` | Full architecture context; Chi + pgx + React stack |
| `reference-vault/FB-Brain/internal/registry/registry.go` | 7 hardcoded system registrations; slug-based lookup |
| `reference-vault/FB-Brain/internal/registry/types.go` | SystemDefinition/ActionDefinition/TriggerDefinition/FieldDef type system |
| `reference-vault/FB-Brain/internal/registry/quickbooks.go` | OAuth2 auth with Intuit endpoints; 4 actions, 2 triggers, 4 entities |
| `reference-vault/FB-Brain/internal/registry/gable.go` | API key auth; product/pricing/quote/order actions |
| `reference-vault/FB-Brain/internal/registry/xui.go` | Feed card and phase assignment actions; card action triggers |
| `reference-vault/FB-Brain/internal/registry/localblue.go` | Host header auth; RFQ push and bid status actions |
| `reference-vault/FB-Brain/internal/registry/onebuild.go` | Authorization header auth; estimate and cost data actions |
| `reference-vault/FB-Brain/internal/registry/gmail.go` | Google OAuth2; send/list/get email actions |
| `reference-vault/FB-Brain/internal/registry/outlook.go` | Microsoft OAuth2; send/list/get email actions |
| `reference-vault/FB-Brain/internal/models/models.go` | AccountLink, RFQ, IntegrationEvent, flow request/response types |
| `reference-vault/FB-Brain/internal/config/config.go` | All env vars including Anthropic, QuickBooks, 1Build, Gmail, Outlook |
| `reference-vault/FB-Brain/internal/orchestrator/materials_flow.go` | 9-step hardcoded roofing materials flow |
| `reference-vault/FB-Brain/internal/orchestrator/labor_flow.go` | 6-step hardcoded labor bidding flow |
| `reference-vault/FB-Brain/internal/orchestrator/post_approval.go` | Materials + labor convergence check |
| `reference-vault/FB-Brain/internal/hub/hub.go` | Hub router with all admin endpoints; plan-gating; SSE events |
| `reference-vault/FB-Brain/internal/hub/engine_handlers.go` | WorkflowExecutor validate/dry-run/execute pipeline |
| `reference-vault/FB-Brain/internal/hub/registry_handlers.go` | HTTP handlers for listing/getting system definitions |
| `reference-vault/futurebuild-os/CLAUDE.md` | Clerk auth (to be replaced); Lit 3.0 frontend; A2A package exists |
| `futurebuild-brain/.agents/TECH_STACK.md` | Go 1.25, Chi, pgx, Lit 3.0, OIDC provider role, MCP + REST |

---

## Web Research Sources Consulted

| Topic | Key Sources |
|-------|-------------|
| OIDC providers | Zitadel OIDC library, Ory Hydra, Zitadel platform, Casdoor, Dex |
| MCP specification | MCP 2025-11-25 spec, official Go SDK, MCP Registry API |
| A2A protocol | A2A spec v0.3, GitHub a2aproject/A2A, Red Hat security guide |
| QuickBooks API | Intuit refresh token policy, CloudEvents deadline, OAuth2 docs |
| Construction market | G2 reviews, Capterra, Reddit r/construction, market research |
| AI architecture | Enterprise AI patterns, intent classification, hybrid architectures |
| Zero-Trust | NIST 800-207, SPIFFE/SPIRE, BeyondCorp, mTLS guides |
| Go libraries | go-jose/v4, lestrrat-go/jwx, coreos/go-oidc, go-spiffe |
| Lit components | Lit 3.0 enterprise patterns, web components benchmarks |

---

## Next Steps (After Approval Gate 1)

1. **Stage 05: Architecture Spec** -- Detailed technical architecture document with Go package structure, database schema, API contracts
2. **Stage 06: Implementation Plan** -- Sprint-level task breakdown with story points
3. **Stage 07: Walking Skeleton** -- 2-week implementation of end-to-end proof of concept
4. **Stage 08: MVP Phase 1** -- OIDC Provider + MCP Registry foundation (weeks 3-4)

---

## Artifact Inventory

| File | Stage | Size | Purpose |
|------|-------|------|---------|
| `VISION_BRIEF.md` | 00 | Vision | System vision, legacy foundation, architectural pillars |
| `RESEARCH_AGENDA.md` | 00 | Planning | Research domains, priorities, questions |
| `ASSUMPTIONS_REGISTER.md` | 00 | Risk | Critical, architectural, and business assumptions |
| `RESEARCH_FINDINGS.md` | 01 | Research | Legacy vault analysis + web research findings |
| `COMPETITIVE_LANDSCAPE.md` | 01 | Research | Market analysis, competitor comparison, positioning |
| `PERSONAS.md` | 02 | UX | 4 user personas grounded in legacy system users |
| `JTBD_ANALYSIS.md` | 02 | UX | 5 core jobs + 5 supporting jobs |
| `USER_JOURNEYS.md` | 02 | UX | 5 end-to-end user journeys with Given/When/Then |
| `SOLUTION_CANDIDATES.md` | 03 | Design | 5 solution concepts with verdicts |
| `FEASIBILITY_ASSESSMENT.md` | 03 | Design | Technical feasibility per pillar; risk register |
| `BUILD_VS_BUY.md` | 03 | Design | 6 build-vs-buy decisions with scoring |
| `SCOPE_DEFINITION.md` | 04 | Scope | Walking skeleton, MVP features, phased roadmap |
| `METRICS_FRAMEWORK.md` | 04 | Scope | North star, pillar, business, and operational metrics |
| `PIPELINE_STATE.md` | Meta | State | This file -- pipeline status and audit trail |
