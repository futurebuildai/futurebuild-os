### Execution Roadmap
**Governing Strategy:** See `PRODUCTION_PLAN.md` for detailed Phase definitions and Validation Criteria.

#### Current Focus
*   [x] **Latest Spec Implemented:** Phase 5, Step 39 (Confidence Scoring & Review Flags) ✅
*   [x] **Health Status:** Green ✅ (Integration Tests Passing)
*   [x] **Current Goal:** Phase 5, Step 40 (Site Photo Verification Flow)

#### Development Queue
##### Phase 5: Context Engine - AI Integration [Status: ⏳ In Progress]
*   [x] **Vertex AI & Object Storage Setup:** Implemented S3 & Vertex Clients (Ref: `PRODUCTION_PLAN.md` Step 35) ✅
*   [x] **Vertex Client Refactor:** Multi-model & Multimodal support (Step 35.5) ✅
*   [x] **RAG Pipeline:** Implement document chunking and pgvector (Ref: `PRODUCTION_PLAN.md` Step 36) ✅
*   [x] **Invoice Processor:** PDF -> InvoiceExtraction JSON (Ref: `PRODUCTION_PLAN.md` Step 37) ✅
*   [x] **DirectoryService:** Map phases to contacts (Ref: `PRODUCTION_PLAN.md` Step 38) ✅
*   [x] **Review Flags:** Confidence scoring & human gates (Ref: `PRODUCTION_PLAN.md` Step 39) ✅

#### Development Queue
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
