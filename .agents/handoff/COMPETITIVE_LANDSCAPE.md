# Competitive Landscape

**Document ID:** AG-01-COMPETITIVE
**Created:** 2026-04-02
**Status:** Stage 01 Complete

---

## 1. Market Overview

The construction software market reached USD 10.19 billion in 2025 and is projected to hit USD 21.04 billion by 2032. The construction management software sub-segment is projected at USD 73.95 billion by 2035 with a 9.71% CAGR.

Sources: [Construction Software Market](https://www.researchandmarkets.com/report/construction-management-software), [Market Research Future](https://www.marketresearchfuture.com/reports/construction-management-software-market-29878)

---

## 2. Direct Competitors

### Procore Technologies (Enterprise / Commercial)

| Attribute | Details |
|-----------|---------|
| Target | Commercial GCs, $50M+ annual construction volume |
| Pricing | ACV-based (Annual Construction Volume), $10K-$80K/year |
| G2 Rating | 4.6/5 |
| Key Strength | Depth of modules, enterprise scalability |
| Key Weakness | Prohibitively expensive for SMBs; rigid customization; integration issues with accounting (XERO) |
| FutureBuild Advantage | Brain's MCP registry enables any integration without Procore's lock-in; 10x cheaper for residential builders |

### Buildertrend (SMB / Residential)

| Attribute | Details |
|-----------|---------|
| Target | Small-to-medium residential builders and remodelers |
| Pricing | $299/mo (Core), $499/mo (Pro), $699/mo (Premium) |
| G2 Rating | 4.2/5 |
| Key Strength | Project tracking, field communication, scalability |
| Key Weakness | "Mile wide, inch thin" -- many features but shallow depth; 50-65% price increases post-2022; no bulk data export |
| FutureBuild Advantage | Brain's Maestro AI co-pilot replaces form-based workflows with natural language; deeper integration than Buildertrend's native connectors |

### CoConstruct (Custom Home Builders)

| Attribute | Details |
|-----------|---------|
| Target | Custom home builders and remodelers |
| Pricing | $399/mo (Essential), $699/mo (Advanced), $999/mo (Complete) |
| G2 Rating | 4.3/5 (historically higher than Buildertrend) |
| Key Strength | Estimating, bidding, client communication |
| Key Weakness | Acquired by Buildertrend; no longer being actively updated; users fear sunsetting; prices increasing |
| FutureBuild Advantage | Brain provides the connection layer CoConstruct users are losing -- integrating their existing tools rather than replacing them |

Sources: [Procore G2 Reviews](https://www.g2.com/products/procore/reviews), [Buildertrend G2](https://www.g2.com/products/buildertrend/reviews), [CoConstruct G2](https://www.g2.com/products/co-construct-coconstruct/reviews), [Buildertrend Pricing](https://projul.com/blog/buildertrend-pricing-analysis-2026/), [Procore Pricing](https://www.itqlick.com/procore/pricing)

---

## 3. Integration/Connection Competitors

| Platform | What It Does | How Brain Differs |
|----------|-------------|-------------------|
| Zapier | Generic integration platform (triggers + actions) | Brain is construction-domain-specific with AI intent parsing; MCP-native rather than proprietary |
| Make (Integromat) | Visual workflow automation | Brain understands construction semantics ("order framing lumber") not just data mapping |
| Ramp / Workato | Enterprise iPaaS | Brain is purpose-built for construction; embeds OIDC identity rather than delegating to external IdP |
| Merge.dev | Unified API for categories (accounting, CRM, etc.) | Brain provides MCP Tools with construction-specific schemas, not generic unified APIs |

---

## 4. AI-Native Construction Competitors

| Platform | AI Capability | How Maestro Differs |
|----------|--------------|---------------------|
| Alice Technologies | AI scheduling optimization | Maestro orchestrates across systems, not just scheduling |
| ALICE by Procore | Document AI extraction | Maestro is a co-pilot, not a single-purpose extraction tool |
| Togal.AI | AI takeoff/estimating | Maestro routes to 1Build for estimating via MCP Tool, not competing |
| OpenSpace | 360 reality capture + AI | Different domain (site documentation vs. workflow orchestration) |

**Key insight:** No competitor combines OIDC identity + MCP tool registry + AI co-pilot + A2A webhooks in a single construction-focused platform. This is FutureBuild's whitespace.

---

## 5. Positioning Matrix

```
                    AI-Native
                        |
                   Maestro+Brain
                        |
            Procore AI  |
                 \      |
     Integration  \     |      Connection
     Depth         \    |        Breadth
     <--------------+---+------------>
                    /    |
          CoConstruct    |
                /        |
          Buildertrend   |
                        |
                    Form-Based
```

**FutureBuild Brain occupies the upper-right quadrant:** AI-native with broad connection breadth via MCP. No competitor currently occupies this position.

---

## 6. Market Gaps Exploitable by FutureBuild

1. **QuickBooks Integration Gap** -- Builders universally use QuickBooks but struggle with bid management, committed cost tracking, and sub management within QBO alone. Brain's MCP QuickBooks server bridges this gap.

2. **CoConstruct Refugee Gap** -- CoConstruct users (custom home builders) are losing their preferred platform post-Buildertrend acquisition. Brain can capture this segment by being the connection layer for their existing tool stack.

3. **AI Copilot Gap** -- No construction platform offers natural-language workflow orchestration. Procore's AI is document-focused; Buildertrend has no AI. Maestro fills this gap.

4. **Identity Fragmentation Gap** -- Builders use 5-10 different tools with separate logins. Brain's OIDC provider offers SSO across the construction ecosystem.

5. **Integration Lock-in Gap** -- Procore and Buildertrend integrate with QuickBooks but lock users into their ecosystem. Brain's MCP standard prevents lock-in via open protocol.

Sources: [QuickBooks Construction Pain Points](https://buildern.com/resources/blog/quickbooks-for-contractors/), [CoConstruct Alternatives](https://projul.com/blog/best-coconstruct-alternatives/), [Construction PM Reddit](https://ones.com/blog/top-project-management-software-construction-reddit-guide-2/)
