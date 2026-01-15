```
### Execution Roadmap
**Governing Strategy:** See `PRODUCTION_PLAN.md` for detailed Phase definitions and Validation Criteria.

*   [x] **Mock Ingestion:** Create deterministic test fixtures (Ref: `PRODUCTION_PLAN.md` Step 42) ✅

#### Current Focus
*   [x] **Verification:** Verify endpoint with mock Auth Token and DB check (Ref: `PRODUCTION_PLAN.md` Step 43.6) ✅
*   [x] **Artifact Mapping:** Implement tool-to-artifact mapping and Rich UI response (Ref: `PRODUCTION_PLAN.md` Step 44) ✅

#### Current Focus
*   [x] **Latest Completed:** Phase 6, Step 47 (Sub Liaison Agent) ✅
*   [x] **Operation Ironclad:** Technical Debt Remediation ✅
*   [x] **Health Status:** Green ✅
*   [ ] **Current Goal:** Phase 6, Step 48 (Inbound Message Processing & State Machine) ⏳



##### Phase 5: Context Engine - AI Integration [Status: ✅ Completed]
*   [x] **Vertex AI Setup:** Client and PDF upload pipeline (Ref: `PRODUCTION_PLAN.md` Step 35)
*   [x] **RAG Pipeline:** Implement pgvector embeddings (Ref: `PRODUCTION_PLAN.md` Step 36)
*   [x] **Invoice Processor:** PDF -> InvoiceExtraction JSON (Ref: `PRODUCTION_PLAN.md` Step 37)
*   [x] **Directory Service:** Project Phase -> Contact lookup (Ref: `PRODUCTION_PLAN.md` Step 38)
*   [x] **Review Flags:** Confidence scoring and human review trigger (Ref: `PRODUCTION_PLAN.md` Step 39)
*   [x] **Site Verification:** Build site photo verification flow (Ref: `PRODUCTION_PLAN.md` Step 40)
*   [x] **SDK Upgrade:** Upgrade Vertex AI SDK (Ref: `PRODUCTION_PLAN.md` Step 40b)
*   [x] **Audit Trail:** Document re-processing and audit system (Ref: `PRODUCTION_PLAN.md` Step 41)
*   [x] **Mock Ingestion:** Mock test fixture for pipeline (Ref: `PRODUCTION_PLAN.md` Step 42)
##### Phase 0: Foundation & CI/CD [Status: ✅ Completed]
*   [x] **Git Setup:** Initialize .gitignore (Ref: `PRODUCTION_PLAN.md` Step 1)
*   [x] **CI/CD:** Configure GitHub Actions (Ref: `PRODUCTION_PLAN.md` Step 2)
*   [x] **Testing:** Set up automated testing pipeline (Ref: `PRODUCTION_PLAN.md` Step 3)
*   [x] **Quality:** Configure developer Makefile (Ref: `PRODUCTION_PLAN.md` Step 4)
*   [x] **Staging:** Document infrastructure & Docker prep (Ref: `PRODUCTION_PLAN.md` Step 5 & 8)
*   [x] **Env Vars:** Configure environment variable management (Ref: `PRODUCTION_PLAN.md` Step 6)
*   [x] **Migrations:** Set up database migrations framework (Ref: `PRODUCTION_PLAN.md` Step 7)

##### Phase 1: Database & Core Models [Status: ✅ Completed]
*   [x] **Database:** Provision PostgreSQL & pgvector (Ref: `PRODUCTION_PLAN.md` Step 9) <!-- id: 9 -->
*   [x] **Identity:** Create Org, User, Project models & DB Schema (Ref: `PRODUCTION_PLAN.md` Step 10)
*   [x] **WBS Library:** Create Template, Phase, Task models & DB Schema (Ref: `PRODUCTION_PLAN.md` Step 11)
*   [x] **Financials:** Create Domain 3 tables (Ref: `PRODUCTION_PLAN.md` Step 12)
*   [x] **Communication:** Create Domain 4 tables (Ref: `PRODUCTION_PLAN.md` Step 13)
*   [x] **Project Tasks:** Create Task & Dependency models (Ref: `PRODUCTION_PLAN.md` Step 14)
*   [x] **Seeding:** Seed WBS Tasks & Ghost Predecessors (Ref: `PRODUCTION_PLAN.md` Step 15)
*   [x] **Index/Health:** DB indexing and health check endpoint (Ref: `PRODUCTION_PLAN.md` Step 16.1)
*   [x] **Project CRUD:** Implement POST/GET /api/v1/projects & Multi-Tenancy gates (Ref: `PRODUCTION_PLAN.md` Step 16.2)
*   [x] **Type System:** Create pkg/types (Ref: `PRODUCTION_PLAN.md` Step 17) <!-- id: 17 -->
*   [x] **Frontend Types:** Create frontend/src/types (Ref: `PRODUCTION_PLAN.md` Step 18)
*   [x] **Service Interfaces:** Define Go Service Interfaces (Ref: `PRODUCTION_PLAN.md` Step 19)
*   [x] **Contract Validation:** Write Go-to-TS parity test suite (Ref: `PRODUCTION_PLAN.md` Step 20)
*   [x] **Magic Link Auth:** Implement stateful passwordless login (Ref: `PRODUCTION_PLAN.md` Step 21)
*   [x] **JWT Generation:** Issue signed tokens with multi-tenant claims (Ref: `PRODUCTION_PLAN.md` Step 22)
*   [x] **RBAC Middleware:** Role-based access control and Context Safety (Ref: `PRODUCTION_PLAN.md` Step 23)
*   [x] **Portal Access Tokens:** Enable external `CONTACTS` via Magic Link (Ref: `PRODUCTION_PLAN.md` Step 24)
*   [x] **Rate Limiting:** Protect auth endpoints with IP-based throttling (Ref: `PRODUCTION_PLAN.md` Step 25)

##### Phase 4: Physics Engine - Core Scheduling [Status: ✅ Completed]
*   [x] **DHSM Calculator:** Implement multiplier logic (Ref: `PRODUCTION_PLAN.md` Step 26)
*   [x] **Dependency Graph:** Build gonum/graph DAG (Ref: `PRODUCTION_PLAN.md` Step 27)
*   [x] **CPM Forward Pass:** Calculate ES, EF (Ref: `PRODUCTION_PLAN.md` Step 28)
*   [x] **CPM Backward Pass:** Calculate LS, LF, float, critical path (Ref: `PRODUCTION_PLAN.md` Step 29)
*   [x] **Event Duration Locking:** Fixed durations for Inspections bypass SAF (Ref: `PRODUCTION_PLAN.md` Step 30)
*   [x] **SWIM Weather Overlay:** Pre-dry-in duration adjustments (Ref: `PRODUCTION_PLAN.md` Step 31)
*   [x] **Schedule Recalculation Trigger:** Auto-recalc on task changes (Ref: `PRODUCTION_PLAN.md` Step 32) ✅
*   [x] **Golden Master Physics Test:** Validated DHSM & CPM against 50+ scenarios (Ref: `PRODUCTION_PLAN.md` Step 33) ✅
*   [x] **Cycle Detection Test:** Verified CPM solver rejects circular dependencies (Ref: `PRODUCTION_PLAN.md` Step 34) ✅


#### Future Ideas / Backlog
*(See `VISION.md` Section 3.2 for Out of Scope items)*
| Feature | Priority | Related Spec |
| :--- | :--- | :--- |
| [FUTURE_FEATURE_1] | [High/Mid/Low] | [e.g., New Domain in DATA_SPINE] |
