# Metrics Framework

**Document ID:** AG-04-METRICS
**Created:** 2026-04-02
**Status:** Stage 04 Complete -- PAUSED AT APPROVAL GATE 1

---

## 1. North Star Metric

**Cross-System Workflow Completion Rate** -- Percentage of Maestro-initiated workflows that complete end-to-end across all involved systems without manual intervention.

Target: >80% at MVP launch, >95% at 6 months.

---

## 2. Pillar Metrics

### Pillar 1: OIDC Identity Provider

| Metric | Definition | Target | Measurement |
|--------|-----------|--------|-------------|
| Token Issuance Latency | p99 time from `/authorize` redirect to token issuance | <500ms | Prometheus histogram on `/token` endpoint |
| Token Validation Latency | p99 time for FB-OS to validate Brain-issued JWT via JWKS | <10ms | Prometheus histogram on JWKS cache hit/miss |
| JWKS Availability | Uptime of `/.well-known/openid-configuration` and `/jwks` endpoints | 99.99% | Uptime monitor |
| Active OIDC Sessions | Number of concurrent valid sessions | Baseline + growth tracking | PostgreSQL count |
| Failed Auth Attempts | Rate of failed authentication attempts (brute force detection) | <1% of total attempts | Prometheus counter |
| Refresh Token Rotation Success | Percentage of refresh token rotations that succeed without re-auth | >99.5% | Prometheus counter |

### Pillar 2: MCP Server Registry

| Metric | Definition | Target | Measurement |
|--------|-----------|--------|-------------|
| Registered MCP Servers | Total servers in registry | 7 at MVP (matching legacy integrations) | PostgreSQL count |
| Tool Discovery Latency | p99 time for `tools/list` response | <50ms | Prometheus histogram |
| Tool Execution Latency | p99 time for `tools/call` end-to-end (including external API call) | <3s | Prometheus histogram, bucketed by server |
| Tool Execution Success Rate | Percentage of `tools/call` that return non-error responses | >95% | Prometheus counter (success/total) |
| Schema Validation Errors | Rate of tool calls rejected by JSON Schema validation | <2% | Prometheus counter |
| Server Health | Per-server health check pass rate | >99% per server | Cron health check + Prometheus gauge |

### Pillar 3: Maestro AI Co-pilot

| Metric | Definition | Target | Measurement |
|--------|-----------|--------|-------------|
| Intent Classification Accuracy | Percentage of intents correctly classified (measured against labeled evaluation set) | >90% top-1 accuracy | Weekly evaluation against 200-utterance test set |
| Intent Classification Latency | p99 time for Maestro to classify intent and select MCP tools | <2s | Prometheus histogram |
| Confidence Distribution | Distribution of classification confidence scores | Median >0.90 | Prometheus histogram |
| Human Confirmation Rate | Percentage of intents requiring human confirmation (confidence < threshold) | <15% | Prometheus counter |
| Workflow Completion Rate | Percentage of Maestro-initiated workflows that complete successfully | >80% | Prometheus counter (completed/initiated) |
| Fallback to Regex | Percentage of intents handled by regex fast-path vs. Claude | Track ratio | Prometheus counter per tier |
| Cost per Intent | Average Claude API cost per intent classification | <$0.03 | Anthropic usage tracking |

### Pillar 4: A2A Webhooks

| Metric | Definition | Target | Measurement |
|--------|-----------|--------|-------------|
| Webhook Delivery Latency | p99 time from webhook emission to acknowledgment by receiver | <500ms | Prometheus histogram |
| Webhook Delivery Rate | Percentage of webhooks successfully delivered (HTTP 2xx) | >99.9% | Prometheus counter |
| Webhook Retry Rate | Percentage of webhooks requiring retry | <1% | Prometheus counter |
| Dead Letter Rate | Percentage of webhooks that exhaust all retries | <0.01% | Prometheus counter + alerting |
| Signature Verification Success | Percentage of webhooks with valid JWS signature at receiver | 100% | FB-OS Prometheus counter |

### Pillar 5: Admin UI (Lit)

| Metric | Definition | Target | Measurement |
|--------|-----------|--------|-------------|
| Lighthouse Performance | Lighthouse performance score | >90 | Lighthouse CI |
| Lighthouse Accessibility | Lighthouse accessibility score | >95 | Lighthouse CI |
| Bundle Size | Total JS bundle size (gzipped) | <50KB | Vite build output |
| First Contentful Paint | Time to FCP | <1.5s | Lighthouse |
| Time to Interactive | TTI | <3s | Lighthouse |

---

## 3. Business Metrics

| Metric | Definition | Target | Measurement |
|--------|-----------|--------|-------------|
| Time to First Integration | Time from account creation to first successful MCP tool call | <30 minutes | Funnel tracking |
| Active Connected Systems | Average number of MCP servers connected per org | >3 at 90 days | PostgreSQL query |
| Weekly Active Maestro Users | Users who interact with Maestro at least once per week | >60% of active users | Usage analytics |
| Manual Data Entry Reduction | Estimated reduction in cross-system data re-entry (measured via survey) | >50% | Quarterly user survey |
| QuickBooks Auto-PO Adoption | Percentage of approved quotes that trigger automatic QuickBooks PO | >70% within 90 days | Prometheus counter |

---

## 4. Operational Metrics

| Metric | Definition | Target | Alerting Threshold |
|--------|-----------|--------|-------------------|
| Brain API Uptime | Percentage of time Brain responds to health checks | 99.9% | Alert at <99.5% over 5m |
| PostgreSQL Connection Pool | Active vs. max connections | <80% utilization | Alert at >85% |
| Memory Usage | Brain process RSS | <512MB | Alert at >400MB |
| CPU Usage | Brain process CPU | <50% sustained | Alert at >70% for 5m |
| Error Rate (5xx) | Percentage of HTTP 5xx responses | <0.1% | Alert at >0.5% |
| IntegrationEvent Volume | Events logged per hour | Baseline + anomaly detection | Alert at >3x baseline |

---

## 5. Success Criteria by Phase

### Phase 1 (Weeks 1-4): Foundation

| Criterion | Target | Pass/Fail |
|-----------|--------|-----------|
| FB-OS validates Brain OIDC token | JWT accepted by FB-OS middleware | Binary |
| OIDC conformance test | Pass basic profile tests | Binary |
| MCP tools/list responds | Returns registered servers with tools | Binary |
| Walking skeleton end-to-end | User authenticates -> calls MCP tool -> gets result | Binary |

### Phase 2 (Weeks 5-8): Integration Servers

| Criterion | Target | Pass/Fail |
|-----------|--------|-----------|
| 4 core MCP servers operational | GableERP, LocalBlue, XUI, QuickBooks all pass tool execution tests | Binary |
| QuickBooks CloudEvents | Webhook receiver handles CloudEvents format | Binary |
| A2A webhook delivery | FB-OS receives and processes webhook | Binary |
| Tool execution success rate | >90% across all servers | Threshold |

### Phase 3 (Weeks 9-12): Maestro AI

| Criterion | Target | Pass/Fail |
|-----------|--------|-----------|
| Intent classification accuracy | >90% on evaluation set | Threshold |
| Materials flow end-to-end | "Order roofing materials" -> quote created -> review card appears | Binary |
| Labor flow end-to-end | "Find roofers" -> RFQ sent -> bid review card appears | Binary |
| QuickBooks PO automation | Approved quote -> PO created in <60s | Threshold |
| Cross-system workflow completion | >80% of initiated workflows complete | Threshold |

---

## 6. Instrumentation Plan

All metrics are collected via:
- **Prometheus** -- Go `promhttp` middleware on Chi router; custom counters/histograms per pillar
- **OpenTelemetry** -- Distributed tracing for cross-system calls (Brain -> GableERP, Brain -> FB-OS)
- **slog (structured JSON)** -- Application logging for debugging and audit
- **PostgreSQL** -- Business metrics via queries against `integration_events`, `hub_users`, `mcp_servers` tables

Dashboard: Grafana with pre-built panels for each pillar metric group.
