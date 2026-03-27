# FutureBuild Specification Index

This document serves as the master map for all technical and product specifications within the FutureBuild repository. Use this index to navigate the various layers of the system architecture and implementation roadmap.

---

## Section 1: The Foundation (Strategy & Planning)

### [PRODUCT_VISION.md](./PRODUCT_VISION.md)
**Purpose:** Defines the "Why" and "What" of FutureBuild, including the 6-Layer Architecture (0-5) and the core philosophy of "Project-Native Intelligence."
**Key Contents:** High-level project goals, architectural overviews, core philosophical tenets, and system-wide design principles.

### [PRODUCTION_PLAN.md](./PRODUCTION_PLAN.md)
**Purpose:** The Execution Roadmap. A step-by-step implementation guide (Phase 0 to Phase 8) defining the build order, dependencies, and critical testing checkpoints.
**Key Contents:** Milestones, phase breakdown, dependency mapping, and rigorous QA verification protocols.

---

## Section 2: The Logic Core (Backend & Data)

### [DATA_SPINE_SPEC.md](./DATA_SPINE_SPEC.md)
**Purpose:** Layer 2 (Database Schema). The definitive PostgreSQL schema reference.
**Key Contents:** Definitions for Multi-Tenancy (Orgs), Projects, Financials, and core domain schemas.

### [CPM_RES_MODEL_SPEC.md](./CPM_RES_MODEL_SPEC.md)
**Purpose:** Layer 3 (Physics Engine). The "Deterministic" math specification.
**Key Contents:** DHSM formulas, SWIM weather logic, and Critical Path Method (CPM) algorithms.

### [AGENT_BEHAVIOR_SPEC.md](./AGENT_BEHAVIOR_SPEC.md)
**Purpose:** Layer 4 (Action Engine). The "Probabilistic" logic specification.
**Key Contents:** Behavior, triggers, and logic flows for the 5 Autonomous Agents (Daily Focus, Procurement, Chat, etc.).

### [BACKEND_SCOPE.md](./BACKEND_SCOPE.md)
**Purpose:** Technical Stack & Business Rules. Defines the Go technology stack and SaaS "Gatekeeper" limits.
**Key Contents:** WBS phase definitions, inspection/business rules, and server-side architectural constraints.

---

## Section 3: The Interface Contracts (The "Rosetta Stone")

### [API_AND_TYPES_SPEC.md](./API_AND_TYPES_SPEC.md)
**Purpose:** Backend & Shared Contracts. The single source of truth for backend services and shared enums.
**Key Contents:** Shared Enums (TaskStatus, UserRole), Go Service Interfaces, and API JSON payload definitions.

### [FRONTEND_TYPES_SPEC.md](./FRONTEND_TYPES_SPEC.md)
**Purpose:** Frontend Data Contracts. Strict TypeScript interface definitions that mirror the Backend Types.
**Key Contents:** TypeScript interfaces for all shared entities ensuring type safety across the network boundary.

---

## Section 4: The Experience (Frontend & UI)

### [FRONTEND_SCOPE.md](./FRONTEND_SCOPE.md)
**Purpose:** Frontend Architecture. Defines the Lit 3.0 + Vite technology stack and UX paradigm.
**Key Contents:** Component hierarchy, routing strategy, and the "Chat + Dashboard" layout specifications.

### [MASTER_PRD.md](./MASTER_PRD.md)
**Purpose:** User Requirements & UI Stories. The detailed Product Requirement Document describing specific features and acceptance criteria.
**Key Contents:** Command Center, Gantt, Invoice Artifacts, User Stories, and Acceptance criteria for the Generative UI.

---

## Section 5: Remediation & Audits

### [REMEDIATION_HANDLERS_TEST.md](./REMEDIATION_HANDLERS_TEST.md)
**Purpose:** Phase 8 Remediation Plan.
**Key Contents:** Specification for refactoring `internal/api/handlers` to use Dependency Injection (Interfaces) to increase unit test coverage from 23% to 80%+.
