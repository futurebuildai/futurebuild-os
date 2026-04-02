# Personas

**Document ID:** AG-02-PERSONAS
**Created:** 2026-04-02
**Status:** Stage 02 Complete

---

## Persona 1: Mike the General Contractor (Primary)

**Demographics:** Male, 42, owns a residential construction company with 8-15 employees in San Diego, CA. Annual revenue: $3-8M. Builds 4-8 custom homes per year.

**Current Tools:**
- XUI Projects for project management (feed cards, phases, WBS)
- GableERP for lumber/materials procurement
- LocalBlue for subcontractor bidding
- QuickBooks Online for accounting/invoicing
- Gmail for communication
- Spreadsheets for everything else

**Legacy Context:** Mike is the exact user the existing MaterialsFlow and LaborFlow in `reference-vault/FB-Brain/internal/orchestrator/` were designed for. He triggers materials quotes through XUI and reviews bids that come back through the feed card system.

**Jobs to Be Done:**
1. "When I start a roofing phase, I need to get materials quoted AND labor bids simultaneously so I can keep the schedule on track"
2. "When a bid comes in from a sub, I need to compare it against my budget and approve or reject without opening 3 different systems"
3. "When materials are ordered, I need QuickBooks to automatically create the PO so my bookkeeper doesn't have to re-enter everything"

**Pain Points:**
- Logs into 5+ systems daily; forgets which system has the information he needs
- Re-enters data across systems (quote in GableERP -> PO in QuickBooks manually)
- Cannot see the holistic status of a project across all systems in one place
- Misses subcontractor bid deadlines because notifications are scattered

**Maestro Interaction Model:**
- "Hey Maestro, order roofing materials for the Riverside project and get me labor bids"
- Maestro parses intent, looks up AccountLink, calls GableERP MCP tools, pushes RFQ to LocalBlue, creates review cards in XUI, and emits A2A webhook to FB-OS for scheduling impact

**Success Metric:** Time from "I need roofing materials" to "quote reviewed and approved" drops from 2 hours (manual) to 5 minutes (Maestro-assisted).

---

## Persona 2: Sarah the Office Manager / Bookkeeper (Secondary)

**Demographics:** Female, 35, works at Mike's company. Handles accounting, invoicing, payroll, and vendor relations.

**Current Tools:**
- QuickBooks Online (primary workspace)
- Gmail/Outlook for vendor communication
- 1Build for cost data lookups
- Spreadsheets for bid comparison

**Legacy Context:** Sarah interacts with the systems defined in `reference-vault/FB-Brain/internal/registry/quickbooks.go` and `gmail.go`. She is the one who manually creates invoices and POs in QuickBooks after Mike approves quotes.

**Jobs to Be Done:**
1. "When a materials order is approved, I need the PO to appear in QuickBooks automatically with correct line items"
2. "When a subcontractor is awarded a bid, I need to send the award letter and create the subcontract in QuickBooks"
3. "I need to reconcile what we ordered in GableERP with what shows up in QuickBooks without manual data entry"

**Pain Points:**
- Double-entry between construction software and accounting
- Cannot track committed costs until bills arrive (QuickBooks limitation)
- Vendor invoice doesn't match PO because someone changed the order and didn't tell her
- No single dashboard showing outstanding POs, unpaid subs, and cash flow impact

**Maestro Interaction Model:**
- "Show me all unpaid subcontractor invoices for the Riverside project"
- Maestro queries QuickBooks MCP tools for bills filtered by vendor + project, cross-references with LocalBlue bid data

**Success Metric:** Monthly close process reduces from 3 days to 1 day through automated reconciliation.

---

## Persona 3: Dave the Subcontractor (Tertiary)

**Demographics:** Male, 48, owns a roofing company with 6 employees. Works with 10-15 different builders per year.

**Current Tools:**
- LocalBlue for receiving RFQs and submitting bids
- QuickBooks Self-Employed for invoicing
- Phone/text for most communication

**Legacy Context:** Dave is the external party who triggers the `bid_submitted` webhook defined in `reference-vault/FB-Brain/internal/registry/localblue.go`. His bid data flows through the `LaborFlow.HandleBidSubmitted()` method in `reference-vault/FB-Brain/internal/orchestrator/labor_flow.go`.

**Jobs to Be Done:**
1. "When I get an RFQ from a builder, I need to respond quickly before another sub takes the job"
2. "When my bid is accepted, I need to know the project schedule so I can plan my crew"
3. "When I finish a phase, I need to submit my invoice and get paid within 30 days"

**Pain Points:**
- Receives RFQs through multiple channels (email, LocalBlue, phone calls)
- No visibility into when the builder will actually need him on site
- Payment delays because the builder's bookkeeper doesn't know the work is done

**Maestro Interaction Model:**
- Dave interacts indirectly -- his bids trigger A2A webhooks from Brain to FB-OS, which update the project schedule. He sees results in LocalBlue.

**Success Metric:** Time from bid acceptance to receiving project schedule drops from "whenever the builder calls" to instant (via A2A webhook -> FB-OS schedule update -> LocalBlue notification).

---

## Persona 4: Alex the Platform Admin (Internal)

**Demographics:** Male, 28, DevOps/platform engineer at FutureBuild. Manages Brain deployment, monitors integrations, onboards new integration partners.

**Current Tools:**
- Hub admin UI (React frontend described in `reference-vault/FB-Brain/internal/hub/hub.go`)
- PostgreSQL admin tools
- GitHub Actions for CI/CD

**Legacy Context:** Alex uses the Hub's platforms, connections, and integrations CRUD endpoints. He registers new systems, tests connectivity, and manages OAuth credentials for QuickBooks/Gmail/Outlook.

**Jobs to Be Done:**
1. "When a new integration partner wants to connect, I need to register their system in the registry with correct auth config"
2. "When an OAuth token expires, I need to know immediately and re-authorize"
3. "When an integration fails, I need the audit trail to debug which system call failed"

**Pain Points:**
- Registry changes require code deployment (compile-time registration)
- No visibility into MCP tool health or usage metrics
- OAuth token refresh failures are silent until a user reports a broken integration

**Maestro Interaction Model:**
- Admin UI for MCP server registry management (Lit Web Components)
- OIDC client management dashboard
- A2A Agent Card browser

**Success Metric:** New integration registration drops from "PR + deploy" (hours) to "admin UI form submission" (minutes).
