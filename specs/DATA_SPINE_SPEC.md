# Layer 2: Data Spine Specification (DATA_SPINE_SPEC1.0)
## Database Schema & Data Strategy

**Version:** 1.0.0  
**Status:** Approved Specification  
**Philosophy:** "The Database is the State." All Agents and Physics Engines are stateless calculators that read from and write to this Spine.

---

## 1. Strategy & Stack

*   **Database:** PostgreSQL 15+
*   **Extensions:** 
    *   `pgvector`: For RAG (Retrieval-Augmented Generation) and document embeddings.
    *   `postgis`: (Optional) For geospatial processing (lat/long geocoding).
*   **Concurrency Model:** Optimistic locking for task updates; ACID compliance for financial and state transitions.

---

## 2. Domain 1: Identity & Access (IAM)

### 2.1 ORGANIZATIONS
The multi-tenant container.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | Unique identifier |
| name | VARCHAR | Legal name of the building entity |
| slug | VARCHAR | URL-friendly identifier |
| settings | JSONB | Org-level configuration |
| created_at | TIMESTAMP | |
| project_limit | INT | Hard cap on active projects. Default: 5 |

### 2.2 USERS
Internal users managed via Magic Link auth.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | Unique identifier |
| org_id | UUID (FK) | Reference to ORGANIZATIONS |
| email | VARCHAR | Unique email (Primary Auth Identifier) |
| name | VARCHAR | Full name |
| role | ENUM | Admin, Builder, Client, Subcontractor (per [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md)) |
| created_at | TIMESTAMP | |

### 2.3 CONTACTS
The Global Address Book ("The Rolodex"). Stores trade partners, clients, and vendors.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | Unique identifier |
| org_id | UUID (FK) | Reference to ORGANIZATIONS |
| name | VARCHAR | Contact Name |
| company | VARCHAR | Company Name |
| phone | VARCHAR | Primary SMS target |
| email | VARCHAR | |
| role | ENUM | Client, Subcontractor (per [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md)) |
| contact_preference | ENUM | SMS, Email, Both |

---

## 3. Domain 2: Project Core (The Graph)

### 3.1 PROJECTS
High-level project container.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| name | VARCHAR | e.g., "Lot 42 - Skyline" |
| address | TEXT | |
| permit_issued_date | DATE | System activation date (WBS 5.2) |
| target_end_date | DATE | Initial target completion |
| gsf | FLOAT | Gross Square Footage (used for SAF calculation) |
| status | ENUM | Preconstruction, Active, Paused, Completed |

### 3.2 PROJECT_CONTEXT
Calibrated physics variables extracted from documents or manually set.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| project_id | UUID (FK) | |
| supply_chain_volatility | INT | Range 1-3 (Default 1) |
| rough_inspection_latency | INT | Days delay (Default 1) |
| final_inspection_latency | INT | Days delay (Default 5) |
| zip_code | VARCHAR | For weather service lookups |
| climate_zone | VARCHAR | For SWIM model selection |

### 3.3 PROJECT_TASKS
The specific instances of tasks for a project.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| project_id | UUID (FK) | |
| wbs_code | VARCHAR | The WBS identifier (e.g., "9.3") |
| name | VARCHAR | Task name |
| is_inspection | BOOLEAN | True if task is an inspection checkpoint |
| early_start | DATE | Calculated ES from CPM |
| early_finish | DATE | Calculated EF from CPM |
| late_start | DATE | Calculated LS from CPM |
| late_finish | DATE | Calculated LF from CPM |
| calculated_duration | FLOAT | DHSM output before weather/SAF |
| weather_adjusted_duration | FLOAT | Output after SWIM & SAF adjustments |
| manual_override_days | FLOAT | User-applied manual adjustment (Nullable) |
| total_float_days | FLOAT | Calculated Total Float from CPM |
| status | ENUM | pending, ready, in_progress, inspection_pending, completed, blocked, delayed |
| verified_by_vision | BOOLEAN | Result of Gemini Flash validation |
| verification_confidence | FLOAT | Confidence score (0.0 - 1.0) |
| is_human_review_required | BOOLEAN | Flag for human review when AI confidence is low |


### 3.4 TASK_DEPENDENCIES
The Directed Acyclic Graph (DAG) edges.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| project_id | UUID (FK) | |
| predecessor_id | UUID (FK) | Reference to PROJECT_TASKS |
| successor_id | UUID (FK) | Reference to PROJECT_TASKS |
| dependency_type | ENUM | FS (Finish-to-Start), SS (Start-to-Start), FF (Finish-to-Finish), SF (Start-to-Finish) |
| lag_days | INT | |

### 3.5 PROJECT_ASSIGNMENTS
The "Role Map" linking contacts to project phases.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| project_id | UUID (FK) | |
| contact_id | UUID (FK) | |
| wbs_phase_id | VARCHAR | The phase code (e.g., "9.x") |

---

## 4. Domain 3: Financials (The Wallet)

### 4.1 PROJECT_BUDGETS
Three-column tracking per WBS Phase.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| project_id | UUID (FK) | |
| wbs_phase_id | VARCHAR | |
| estimated_amount | DECIMAL | From Spec/SAF extraction |
| committed_amount | DECIMAL | From Contracts/POs |
| actual_amount | DECIMAL | From Paid Invoices |

### 4.2 INVOICES
Parsed artifacts from the Action Engine.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| project_id | UUID (FK) | |
| vendor_name | VARCHAR | Extracted vendor identity |
| amount | DECIMAL | Total invoice value |
| line_items | JSONB | Matches `InvoiceExtraction` schema |
| detected_wbs_code | VARCHAR | Predicted WBS mapping |
| status | ENUM | Pending, Approved, Exported |
| is_human_review_required | BOOLEAN | Flag for human review when AI confidence is low |
| source_document_id | UUID (FK) | Reference to source DOCUMENT (Nullable, SET NULL) |

---

## 5. Domain 4: Communication & History

### 5.1 COMMUNICATION_LOGS
History of Agent <-> User interaction.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| project_id | UUID (FK) | |
| contact_id | UUID (FK) | |
| direction | ENUM | Inbound, Outbound |
| content | TEXT | Message body |
| channel | ENUM | SMS, Chat, Email |
| timestamp | TIMESTAMP | |

### 5.2 NOTIFICATIONS
Queue for user alerts and daily focus briefings.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| user_id | UUID (FK) | |
| type | VARCHAR | e.g., "Schedule_Slip", "Invoice_Ready" |
| priority | INT | |
| status | ENUM | Unread, Read, Dismissed |

---

## 6. Domain 5: The Learning Layer

### 6.1 ORG_LEARNING
Stores the organizational bias for hyper-localization.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| bias_vector | JSONB | Map of {WBS_Code: Bias_Factor} |
| updated_at | TIMESTAMP | |

### 6.2 DOCUMENT_CHUNKS
Vector embeddings for RAG-based context retrieval.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| document_id | UUID (FK) | |
| chunk_content | TEXT | |
| embedding | VECTOR(768) | Vertex AI text-embedding-004 |
| metadata | JSONB | Page refs, document type, etc. |
