# FutureBuild Product Vision
## Residential Construction Path Model (CPM-res1.0)

**Version:** 2.0.0
**Classification:** North-Star Vision & Architecture Overview
**Last Updated:** 2026-02-02

---

## 1. System Identity

### 1.1 Name
**FutureBuild**

### 1.2 Tagline
AI-powered construction project management — from permit to punchlist.

### 1.3 Purpose Statement
FutureBuild automates residential construction scheduling by combining probabilistic AI (document understanding, photo verification, context extraction) with deterministic algorithms (critical path calculation, duration estimation, weather adjustment) to manage construction projects from permit issuance through completion.

### 1.4 Unique Value Proposition
The system treats the PROJECT as the primary entity, not human users. Each project is modeled as a living dependency graph that computes its own state, identifies risks, and coordinates stakeholders through AI agents. Builders interact through conversation, not forms — the chat interface is the command center.

---

## 2. Core Philosophy

### 2.1 Project-Native Intelligence
The project is the first-class citizen. Every residential project runs as an independent process with its own task graph (~80 tasks across 11 phases), context variables, and schedule state. The project knows what needs to happen next, what's at risk, and who to notify.

### 2.2 Deterministic Core, Probabilistic Perception

| Layer Type | Purpose | Certainty |
|------------|---------|-----------|
| Probabilistic | Understand documents, photos, messages | Variable (0.0-1.0 confidence) |
| Deterministic | Calculate schedules, durations, costs | Exact (reproducible) |

AI is used for perception and communication. Scheduling, budgeting, and procurement calculations are deterministic and reproducible — no black-box scheduling.

### 2.3 Chat-First UX
Users interact through a conversational interface, not traditional form-heavy dashboards. Visual artifacts (Gantt charts, budgets, invoices) render inline alongside the conversation. The system reduces cognitive load by surfacing what matters through Daily Focus briefings and proactive agent notifications.

### 2.4 Multi-Tenant by Design
Every query, every handler, every data path is scoped by `org_id`. Tenant isolation is enforced at the data layer, not the application layer. There are no global queries.

---

## 3. System Boundaries

### 3.1 What FutureBuild Does (In Scope)
- Post-permit scheduling (WBS 5.2+ through completion)
- Task dependency management (~80 tasks, 11 phases)
- Duration calculation via DHSM (Duration Heuristic Scaling Model)
- Weather-aware scheduling via SWIM (pre-dry-in phase adjustment)
- Critical path computation and risk identification
- Invoice tracking with approval workflows (estimated, committed, actual)
- Procurement coordination with lead-time and volatility awareness
- Site photo verification via AI vision analysis
- Stakeholder communication (email, SMS, portal)
- Project completion lifecycle with audit-grade reporting
- Per-organization learning from historical project data

### 3.2 What FutureBuild Does NOT Do (Out of Scope)
- Pre-permit phases (design, zoning, financing)
- Architectural design or BIM modeling
- Payment processing (invoices are tracked, not settled)
- Payroll or HR management
- Equipment/fleet tracking

---

## 4. Project Lifecycle

### 4.1 Lifecycle Stages

```
Creation → Document Processing → Schedule Generation → Active Construction → Completion
```

1. **PROJECT CREATION**: Builder initiates a project via chat. System geocodes the address, creates the project record, and provisions the task graph from the CPM-res1.0 WBS template.

2. **DOCUMENT PROCESSING**: Builder uploads blueprints, permits, and specs. Gemini extracts project context variables (square footage, floors, bathrooms, garage bays, roof complexity, etc.) that feed duration calculations.

3. **SCHEDULE GENERATION**: The Physics Engine runs DHSM to compute task durations from context variables, applies SWIM weather adjustments to pre-dry-in phases, and solves the critical path via forward/backward pass. Procurement order dates are calculated from need dates minus lead times.

4. **ACTIVE CONSTRUCTION**: The project is live. Daily Focus briefings surface critical-path tasks, weather risks, and upcoming inspections. Subcontractors receive SMS notifications. Clients view progress through the portal. Invoices are submitted, reviewed, and approved. Site photos are verified by vision AI. The schedule recalculates as actuals diverge from plan.

5. **COMPLETION**: Builder marks the project complete. The system generates a CompletionReport (schedule summary, budget variance, weather impact, procurement totals), transitions project status atomically, archives the project, and notifies the project manager.

---

## 5. Architecture Overview

### 5.1 Layer Stack

| Layer | Name | Role |
|-------|------|------|
| L0 | Real World Inputs | Documents, photos, emails, SMS |
| L1 | Context Engine | Gemini AI extraction and embeddings |
| L2 | Data Spine | PostgreSQL + pgvector (source of truth) |
| L3 | Physics Engine | CPM, DHSM, SWIM, Procurement calculators |
| L4 | Action Engine | Autonomous agents and chat orchestrator |
| L5 | Learning Layer | Per-org duration biases from actuals |

### 5.2 Layer 0: Real World Inputs
Ingests blueprint PDFs, site photos, permit documents, invoices, emails (SendGrid inbound), and SMS (Twilio webhooks). All inputs are stored as documents with extracted text chunks and vector embeddings for RAG retrieval.

### 5.3 Layer 1: Context Engine
Google Vertex AI (Gemini 2.5 Flash for speed, Gemini 2.5 Pro for complex analysis). Extracts structured project context from unstructured documents. Generates text embeddings (text-embedding-004) for semantic search and RAG. Powers vision analysis for site photo verification.

### 5.4 Layer 2: Data Spine
PostgreSQL 15+ with pgvector extension. Multi-tenant schema with `org_id` scoping on every table. Key domains: organizations, users, projects, tasks, WBS templates, budgets, invoices, procurement, documents, communications, and learning signals. The database is the source of truth — agents are stateless calculators.

### 5.5 Layer 3: Physics Engine
All calculations are deterministic and reproducible:
- **DHSM**: Task duration = Base Duration x SUM(Variable x Weight). Variables are extracted from project context (sqft, floors, bathrooms, etc.).
- **SWIM**: Weather adjustment multipliers applied only to pre-dry-in phases. Post-dry-in interior work is stable.
- **CPM**: Forward/backward pass on a DAG (gonum/graph) to compute early start, late finish, float, and critical path.
- **Procurement**: Order date = Need date - Lead time - Buffer. Monitors volatility signals.
- **Inspection Checkpoints**: Hard gates that block dependent tasks until inspector sign-off.

### 5.6 Layer 4: Action Engine
Autonomous agents that run on schedules or triggers:
- **Daily Focus Agent**: Morning briefings with critical-path tasks, weather alerts, and inspection reminders.
- **Procurement Agent**: Calculates order windows, monitors lead-time volatility.
- **Subcontractor Liaison**: SMS confirmations, photo collection, progress check-ins.
- **Client Reporter**: Weekly progress summaries with verified photos.
- **Chat Orchestrator**: Intent classification, tool execution, RAG-augmented responses.
- **Inbound Processor**: Parses incoming emails/SMS into structured communication logs.

### 5.7 Layer 5: Learning Layer
Per-organization duration biases captured from actual vs. planned task durations. Multiplier overrides persist per org. Cross-org anonymized training data improves base model over time.

---

## 6. User Model

### 6.1 User Types and Access

| User Type | Auth Method | Access Level |
|-----------|-------------|-------------|
| **Admin** | Clerk (email/OAuth) | Full org control, user management, all project data |
| **Builder** | Clerk (email/OAuth) | Project CRUD, scheduling, invoices, completion, chat |
| **Viewer** | Clerk (email/OAuth) | Read-only project, task, budget, and document access |
| **Client** | Magic link (portal) | View milestones, progress, ask questions |
| **Subcontractor** | Magic link (portal) | View assigned tasks, report progress, upload photos |
| **Vendor** | Magic link (portal) | View procurement orders, confirm deliveries |

### 6.2 Dual Authentication Model
- **Internal users** (Admin, Builder, Viewer): Authenticated via Clerk. Clerk manages sign-up, sign-in, OAuth, and session tokens. The backend validates Clerk-issued JWTs and extracts `org_id` and `user_id` claims.
- **Portal contacts** (Client, Subcontractor, Vendor): Authenticated via magic-link emails. Portal JWTs are scoped to a specific project and contact record with restricted permissions.

### 6.3 RBAC
Permissions are scope-based, mapped statically to roles:
- Scopes: `project:read`, `project:create`, `project:complete`, `task:read`, `task:write`, `budget:read`, `finance:edit`, `document:read`, `document:write`, `chat:read`, `chat:write`, `settings:write`
- Admin has `ScopeAll`. Builder and Viewer have explicitly enumerated scope sets. Portal roles are constrained to read + task-update operations.

---

## 7. Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.24+, Chi Router, raw SQL (pgx), no ORM |
| Database | PostgreSQL 15+ (pgvector), Redis (Asynq job queues) |
| Frontend | Lit 3.0, TypeScript 5.0+ (strict), Vite, Signals (@lit-labs/preact-signals) |
| AI | Google Vertex AI (Gemini 2.5 Flash/Pro), text-embedding-004 |
| Auth | Clerk (internal), magic-link JWT (portal) |
| Email | SendGrid (outbound + inbound parse) |
| SMS | Twilio |
| Storage | DigitalOcean Spaces (S3-compatible) |
| Hosting | DigitalOcean App Platform |

### 7.1 Hard Constraints
- NO React, NO ORMs — raw SQL via pgx only
- NO Python logic — Go is the only backend language
- All TypeScript must compile with `noImplicitAny` enabled
- Database is the source of truth; agents are stateless
- Rosetta Stone type parity: Go `pkg/types/` and TS `frontend/src/types/` must stay in sync

---

## 8. Frontend Architecture

### 8.1 3-Panel Agent Command Center

| Panel | Width | Content |
|-------|-------|---------|
| Left | 280px | Project tree, thread navigation, Daily Focus, Complete Project action |
| Center | flex | Conversational message stream with inline action cards and artifacts |
| Right | 320px, collapsible | Gantt chart, budget overview, invoice detail |

### 8.2 Key Components
- `<fb-chat-interface>` — Central chat hub with message list, input bar, action cards
- `<fb-artifact-gantt>` — D3/SVG timeline with critical path highlighting and dependency arrows
- `<fb-artifact-budget>` — 3-column financial view (Estimated / Committed / Actual)
- `<fb-artifact-invoice>` — Interactive invoice with Approve/Reject workflows
- `<fb-view-completion-report>` — Schedule, budget, weather, and procurement summary cards
- `<fb-vision-badge>` — AI-verified photo indicators on messages
- `<fb-mobile-nav>` — Bottom tab bar for viewports under 768px

### 8.3 State Management
Signals-based reactive store using `@lit-labs/preact-signals`. Single `AppState` object with computed signals for derived data. Actions dispatch through a central store; components react to signal changes without manual subscription management.

---

## 9. Communication System

### 9.1 Channels
| Channel | Provider | Use Case |
|---------|----------|----------|
| Email | SendGrid | Invoices, completion notifications, daily digests |
| SMS | Twilio | Sub confirmations, inspection alerts, progress check-ins |
| Portal | Web app | Client milestones, sub task views, vendor orders |
| In-app chat | Native | Builder ↔ AI agent conversation |

### 9.2 Communication Logging
All outbound and inbound communications are logged in `communication_logs` with channel, direction, project association, and delivery status. Inbound emails and SMS are parsed by the Inbound Processor agent into structured records.

---

## 10. Deployment Model

### 10.1 Topology
- **Platform**: DigitalOcean App Platform
- **Region**: NYC3
- **Environments**: Staging (`staging` branch) and Production (`production` branch)
- **Database**: DO Managed PostgreSQL 15 (auto-managed backups)
- **Migrations**: Automated via `golang-migrate` on container startup (entrypoint.sh)
- **Deployments**: Branch-triggered (manual trigger; `deploy_on_push: false` for staging)

### 10.2 Environment Isolation
Staging mirrors production configuration with staging-specific values. Seed data (org, users, business config) is applied via idempotent migrations (`ON CONFLICT DO NOTHING`). Production uses real Clerk instances, SendGrid keys, and Twilio credentials.

---

## 11. System Constraints

- API response targets: <100ms for reads, <500ms for writes, <3s for AI chat responses
- Request body limit: 1MB
- Notes/text fields: 10,000 character cap
- Rate limiting on auth and chat endpoints
- 7-year data retention for regulatory compliance
- Completion reports are immutable once generated (one per project, unique constraint)

---

## 12. Integration Points

| Service | Purpose |
|---------|---------|
| Google Vertex AI | Document extraction, vision analysis, chat, embeddings |
| Clerk | Identity management, OAuth, session tokens |
| SendGrid | Outbound email, inbound email parsing |
| Twilio | SMS notifications and inbound message handling |
| DigitalOcean Spaces | Document and photo storage (S3-compatible) |
| Weather API | SWIM model data for pre-dry-in phase adjustments |
| Geocoding API | Address resolution for project creation |

---

## 13. What Differentiates FutureBuild

1. **Domain-specific scheduling model**: CPM-res1.0 is purpose-built for residential construction — not a generic project management tool adapted to construction.
2. **Deterministic physics, not AI scheduling**: Durations and critical paths are calculated from observable variables with weighted multipliers. Results are reproducible and auditable.
3. **Chat-first interface**: Builders talk to their project, not navigate dashboards. Visual artifacts render inline alongside conversation.
4. **Inspection checkpoints as hard gates**: Dependent tasks cannot proceed until inspectors sign off. This models real regulatory constraints.
5. **Weather-aware, phase-aware**: SWIM adjustments apply only where they matter (pre-dry-in exterior work), not globally.
6. **Vision-verified progress**: Site photos are analyzed by Gemini to verify reported task completion against visual evidence.
7. **Per-org learning**: The system gets better for each builder over time by calibrating duration multipliers from their actual project data.
8. **Audit-grade completion reports**: Every completed project produces an immutable report with schedule, budget, weather, and procurement summaries.
