# User Journeys

**Document ID:** AG-02-JOURNEYS
**Created:** 2026-04-02
**Status:** Stage 02 Complete

---

## Journey 1: Materials Procurement via Maestro (Evolves MaterialsFlow)

**Persona:** Mike the General Contractor
**Legacy flow:** `reference-vault/FB-Brain/internal/orchestrator/materials_flow.go` -- 9-step hardcoded roofing flow

### Steps

| Step | User Action | System Response | Legacy vs. Revamp |
|------|------------|-----------------|---------------------|
| 1 | Mike opens Brain Hub and types: "I need a quote for framing lumber on the Elm Street project" | Maestro AI parses intent: system=gable, action=bulk_calculate_price, project=elm-street | **Legacy:** Hardcoded `roofingScope` with fixed SKUs. **Revamp:** AI infers scope from project phase + 1Build cost data |
| 2 | (Automatic) | Brain resolves Mike's org -> GableERP customer ID via OIDC subject claims (evolved AccountLink) | **Legacy:** `repo.GetSupplierLink(ctx, orgID)`. **Revamp:** OIDC token contains org claims; MCP server resolves customer mapping |
| 3 | (Automatic) | Brain calls GableERP MCP Server: `tools/call` -> `get_products_by_category("Framing")` | **Legacy:** `f.gable.GetProductsByCategory(ctx, "Roofing")`. **Revamp:** MCP Tool call with JSON Schema validation |
| 4 | (Automatic) | Brain calls GableERP MCP Server: `tools/call` -> `bulk_calculate_price(customer_id, items)` | **Legacy:** `f.gable.BulkCalculatePrice(ctx, customerID, bulkItems)`. **Revamp:** Same API, MCP transport |
| 5 | (Automatic) | Brain calls GableERP MCP Server: `tools/call` -> `create_quote(customer_id, items)` | **Legacy:** `f.gable.CreateQuote(ctx, customerID, quoteLines)`. **Revamp:** Same API, MCP transport |
| 6 | (Automatic) | Brain emits A2A webhook to FB-OS: task=review_material_quote, payload={quote_data} | **Legacy:** `f.xui.CreateFeedCard(ctx, ...)` direct API call. **Revamp:** A2A signed webhook; FB-OS creates the feed card |
| 7 | Mike sees review card in XUI Projects feed with line items and total | Mike reviews the $12,450 framing lumber quote | Same UX, different delivery mechanism |
| 8 | Mike clicks "Approve Quote" | Brain receives approval, calls GableERP `accept_and_convert_quote`, emits A2A webhook to FB-OS, triggers QuickBooks PO creation | **Legacy:** `f.gable.AcceptAndConvertQuote(ctx, ...)`. **Revamp:** Adds QuickBooks MCP Tool call for automatic PO creation |
| 9 | (Automatic) | Brain emits A2A webhook to FB-OS: task=update_schedule, payload={materials_ordered, expected_delivery} | **New** -- FB-OS updates CPM schedule with delivery constraint |

### Given/When/Then

```
Given Mike is authenticated via Brain OIDC and has GableERP connected
When Mike says "I need a quote for framing lumber on the Elm Street project"
Then Maestro:
  - Classifies intent as materials_procurement with >90% confidence
  - Resolves Elm Street project to XUI project ID and GableERP customer ID
  - Calls GableERP MCP tools to get products, calculate pricing, create quote
  - Emits A2A webhook to FB-OS with review task
  - Mike sees a review card in XUI within 30 seconds
```

---

## Journey 2: Subcontractor Bidding via Maestro (Evolves LaborFlow)

**Persona:** Mike the General Contractor + Dave the Subcontractor
**Legacy flow:** `reference-vault/FB-Brain/internal/orchestrator/labor_flow.go` -- 6-step hardcoded labor flow

### Steps

| Step | User Action | System Response | Legacy vs. Revamp |
|------|------------|-----------------|---------------------|
| 1 | Mike: "Find me roofers for the Riverside project, budget around $15k" | Maestro parses intent: system=localblue, action=push_rfq, budget_hint=$15000 | **Legacy:** Hardcoded scope items. **Revamp:** AI constructs scope from project phase data |
| 2 | (Automatic) | Brain resolves org -> LocalBlue site ID, constructs RFQ with AI-generated scope items | **Legacy:** Fixed 5 scope items. **Revamp:** Maestro queries 1Build MCP Tools for labor rates, generates appropriate scope |
| 3 | (Automatic) | Brain calls LocalBlue MCP Server: `tools/call` -> `push_rfq(site_id, rfq_data)` | **Legacy:** `f.localblue.PushRFQ(ctx, siteID, lbRFQ)`. **Revamp:** MCP Tool call |
| 4 | Dave receives RFQ in LocalBlue, submits bid for $14,200 / 5 days | LocalBlue fires `bid_submitted` webhook to Brain | Same trigger mechanism |
| 5 | (Automatic) | Brain receives bid, calls Maestro to contextualize: "Bid is 5% under budget, timeline fits schedule" | **Legacy:** Raw bid data posted to XUI card. **Revamp:** AI enrichment with budget/schedule analysis |
| 6 | (Automatic) | Brain emits A2A webhook to FB-OS: task=review_labor_bid, payload={bid_data, ai_analysis} | **Legacy:** `f.xui.CreateFeedCard(ctx, ...)`. **Revamp:** A2A webhook with AI analysis |
| 7 | Mike approves bid | Brain calls LocalBlue `update_bid_status(accepted)`, assigns contact in XUI, creates QuickBooks subcontract | **Legacy:** `f.localblue.UpdateBidStatus(ctx, ...)` + `f.xui.AssignContactToPhase(ctx, ...)`. **Revamp:** Adds QuickBooks MCP call |
| 8 | (Automatic) | Brain checks post-approval: materials ordered + labor approved -> emits delivery confirmation A2A webhook | **Legacy:** `CheckPostApproval(ctx, ...)`. **Revamp:** A2A webhook replaces direct XUI feed card creation |

---

## Journey 3: First-Time Setup (OIDC + Integration Registration)

**Persona:** Alex the Platform Admin + Mike the General Contractor

### Steps

| Step | User Action | System Response |
|------|------------|-----------------|
| 1 | Alex deploys Brain with OIDC provider enabled | Brain serves `/.well-known/openid-configuration` with issuer, JWKS, and endpoint URLs |
| 2 | Alex configures FB-OS to use Brain as OIDC provider | FB-OS replaces Clerk JWKS URL with Brain's `/jwks` endpoint |
| 3 | Alex registers GableERP as MCP server via Hub admin UI | Brain stores MCP server definition with 4 tools, 2 triggers, auth config |
| 4 | Alex registers LocalBlue, QuickBooks, 1Build, Gmail, Outlook | 7 MCP servers registered with their respective auth configurations |
| 5 | Mike visits FB-OS login page | FB-OS redirects to Brain's `/authorize` endpoint (OIDC authorization code flow with PKCE) |
| 6 | Mike authenticates with Brain (magic link email) | Brain issues authorization code, FB-OS exchanges for access + refresh + ID tokens |
| 7 | Mike connects GableERP in Hub | Brain initiates API key exchange; stores encrypted credential linked to Mike's OIDC subject |
| 8 | Mike connects QuickBooks in Hub | Brain initiates OAuth2 flow with Intuit; stores encrypted tokens linked to Mike's OIDC subject |
| 9 | Mike types "check roofing material prices" | Maestro discovers GableERP MCP tools, validates Mike has active connection, executes `get_products_by_category` |

---

## Journey 4: QuickBooks Automated PO Creation (New Flow)

**Persona:** Sarah the Office Manager

### Steps

| Step | User Action | System Response |
|------|------------|-----------------|
| 1 | (Triggered by Mike approving materials quote in Journey 1) | Brain receives approval event |
| 2 | (Automatic) | Brain calls QuickBooks MCP Server: `tools/call` -> `create_purchase_order(company_id, vendor_id, line_items)` |
| 3 | (Automatic) | QuickBooks PO created with matching line items from GableERP quote |
| 4 | Sarah opens QuickBooks | She sees the PO already created, matching the approved quote exactly |
| 5 | Sarah reconciles | Zero manual data entry required; PO matches quote line-for-line |

### Given/When/Then

```
Given a materials quote has been approved in the MaterialsFlow
  And Mike's org has QuickBooks connected via OAuth2
  And vendor mapping exists between GableERP supplier and QuickBooks vendor
When the approval A2A webhook fires
Then Brain:
  - Maps GableERP quote lines to QuickBooks PO line items
  - Calls QuickBooks MCP Tool to create Purchase Order
  - Logs IntegrationEvent with source=gable, target=quickbooks
  - Sarah sees the PO in QuickBooks within 60 seconds
```

---

## Journey 5: AI-Powered Bid Comparison (New Flow)

**Persona:** Mike the General Contractor

### Steps

| Step | User Action | System Response |
|------|------------|-----------------|
| 1 | Mike: "Compare all bids for the plumbing phase on Riverside" | Maestro parses intent: compare bids, phase=plumbing, project=riverside |
| 2 | (Automatic) | Brain queries LocalBlue MCP Server for all bids on the RFQ |
| 3 | (Automatic) | Brain queries 1Build MCP Server for regional plumbing labor rates |
| 4 | (Automatic) | Maestro generates comparison: "3 bids received. AquaPipe Co at $22K (8% below market), PlumbRight at $24K (at market), DrainMasters at $28K (17% above market)" |
| 5 | Mike: "Award it to AquaPipe" | Brain calls LocalBlue `update_bid_status(accepted)`, creates QuickBooks subcontract, emits A2A webhook to FB-OS |
