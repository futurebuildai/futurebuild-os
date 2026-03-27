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
| invoice_date | DATE | Extracted invoice date |
| invoice_number | VARCHAR | Extracted invoice number |
| status | ENUM | Pending, Approved, Exported |
| confidence | FLOAT | AI Confidence Score (0.0 - 1.0) |
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

### 5.3 CHAT_MESSAGES
State for the internal Chat Orchestrator (Step 43). This is the "Action Engine" event loop.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| project_id | UUID (FK) | |
| user_id | UUID (FK) | Internal User (Superintendent/Builder) |
| role | ENUM | user, model, system, tool |
| content | TEXT | |
| tool_calls | JSONB | Gemini tool call payloads (if role=model) |
| created_at | TIMESTAMP | |

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

---

## 7. Domain 6: Corporate Financials (Phase 18)
Migration 000083. All monetary values as BIGINT (cents).

### 7.1 CORPORATE_BUDGETS
Org-wide budget rollups by fiscal year/quarter.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| fiscal_year | INT | |
| quarter | INT (1-4) | |
| total_estimated_cents | BIGINT | |
| total_committed_cents | BIGINT | |
| total_actual_cents | BIGINT | |
| project_count | INT | |
| last_rollup_at | TIMESTAMP | |
UNIQUE(org_id, fiscal_year, quarter)

### 7.2 GL_SYNC_LOGS
Audit trail for QuickBooks/Xero exports.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| sync_type | VARCHAR(50) | |
| status | VARCHAR(20) | |
| records_synced | INT | |
| error_message | TEXT | |
| synced_at | TIMESTAMP | |

### 7.3 AR_AGING_SNAPSHOTS
Cash flow aging buckets.
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| snapshot_date | DATE | |
| current_cents | BIGINT | 0-30 days |
| days_30_cents | BIGINT | 30-60 days |
| days_60_cents | BIGINT | 60-90 days |
| days_90_plus_cents | BIGINT | 90+ days |
| total_receivable_cents | BIGINT | |
UNIQUE(org_id, snapshot_date)

## 8. Domain 7: HR & Employees (Phase 18)
Migration 000084.

### 8.1 EMPLOYEES
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| employee_number | VARCHAR(50) | UNIQUE(org_id, employee_number) |
| status | VARCHAR(20) | active/on_leave/terminated |
| pay_rate_cents | BIGINT | |
| pay_type | VARCHAR(10) | hourly/salary |
| classification | VARCHAR(100) | Trade classification |

### 8.2 TIME_LOGS
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| employee_id | UUID (FK) | |
| project_id | UUID (FK) | |
| hours_worked | DECIMAL(5,2) | |
| overtime_hours | DECIMAL(5,2) | |
| approved | BOOLEAN | |

### 8.3 CERTIFICATIONS
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| employee_id | UUID (FK) | |
| cert_type | VARCHAR(100) | e.g., OSHA-30, Crane Operator |
| expiration_date | DATE | |
| status | VARCHAR(20) | valid/expiring_soon/expired |

### 8.4 PREVAILING_WAGE_RATES
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| region | VARCHAR(100) | |
| classification | VARCHAR(100) | |
| hourly_rate_cents | BIGINT | |
| fringe_benefit_cents | BIGINT | |
UNIQUE(org_id, region, classification, effective_date)

## 9. Domain 8: Fleet & Equipment (Phase 18)
Migration 000085. Requires `btree_gist` PostgreSQL extension.

### 9.1 FLEET_ASSETS
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| asset_number | VARCHAR(50) | UNIQUE(org_id, asset_number) |
| status | VARCHAR(20) | available/in_use/maintenance/retired |
| purchase_cost_cents | BIGINT | |
| current_value_cents | BIGINT | |

### 9.2 EQUIPMENT_ALLOCATIONS
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| asset_id | UUID (FK) | |
| project_id | UUID (FK) | |
| allocated_from | DATE | |
| allocated_to | DATE | |
| status | VARCHAR(20) | planned/active/completed/cancelled |
EXCLUDE USING GIST — prevents double-booking (same asset, overlapping date ranges, status IN ('planned','active'))

### 9.3 MAINTENANCE_LOGS
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| asset_id | UUID (FK) | |
| maintenance_type | VARCHAR(50) | |
| scheduled_date | DATE | |
| cost_cents | BIGINT | |

## 10. Domain 9: A2A Logging (Phase 18)
Migration 000086.

### 10.1 A2A_EXECUTION_LOGS
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| workflow_id | VARCHAR(100) | |
| source_system | VARCHAR(100) | |
| target_system | VARCHAR(100) | |
| payload | JSONB | |
| status | VARCHAR(20) | |
| duration_ms | INT | |

### 10.2 ACTIVE_AGENT_CONNECTIONS
| Field | Type | Description |
|---|---|---|
| id | UUID (PK) | |
| org_id | UUID (FK) | |
| agent_name | VARCHAR(100) | UNIQUE(org_id, agent_name) |
| status | VARCHAR(20) | active/paused/error |
| execution_count | INT | |
| error_count | INT | |
