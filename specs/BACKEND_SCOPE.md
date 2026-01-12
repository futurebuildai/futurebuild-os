# FutureBuild Backend Scope
## AI-Powered Construction Scheduling Platform (CPM-res1.0)

**Version:** 1.0  
**Date:** December 30, 2025  
**Status:** Planning

---

## 1. Executive Summary

FutureBuild is a SaaS platform that automates residential construction scheduling using the Residential Construction Path Model (CPM-res1.0). The backend provides AI-powered document processing, deterministic scheduling algorithms, autonomous agents, and multi-tenant project management.

**Scope Entry Point:** The system activates after **Permit Issued (WBS 5.2)**. Pre-permit phases (Pre-Design, Schematic, Design Development, Construction Documents, and Permitting) are handled externally.

### Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22+ |
| Framework | Chi Router with net/http |
| Database | PostgreSQL 15+ with pgvector extension |
| AI/LLM | Google Vertex AI (Gemini 2.5 Flash/Pro) |
| Vector Store | pgvector (embedded in PostgreSQL) |
| Task Queue | Asynq (Redis-backed) |
| Email | SendGrid |
| SMS | Twilio |
| File Storage | Digital Ocean Spaces (S3-compatible) |
| Deployment | Digital Ocean App Platform |
| Auth | Magic link email authentication |

---

## 2. WBS Scope Definition

### 2.1 In-Scope Phases (Post-Permit)

The FutureBuild system manages approximately **80 tasks** starting from Permit Issued:

| Phase | WBS Range | Task Count | Description |
|-------|-----------|------------|-------------|
| Preconstruction | 5.3-5.6 | 4 | Bidding, scheduling, logistics |
| Procurement | 6.0-6.9 | 10 | Long-lead item ordering |
| Site Prep | 7.0-7.4 | 5 | Temp services, clearing, layout |
| Foundation | 8.0-8.11 | 12 | Excavation through backfill |
| Framing | 9.0-9.7 | 8 | Floor, walls, roof, dry-in |
| Rough-Ins | 10.0-10.5 | 6 | Plumbing, HVAC, electrical |
| Insulation/Drywall | 11.0-11.5 | 6 | Insulation, board, finish |
| Interior Finishes | 12.0-12.14 | 15 | Trim, cabinets, flooring, fixtures |
| Exterior | 13.0-13.6 | 7 | Cladding, landscaping |
| Commissioning | 14.0-14.7 | 8 | Testing, inspections, closeout |
| Warranty | 15.0-15.4 | 5 | Post-occupancy service |

### 2.2 Out-of-Scope Phases (Pre-Permit)

These phases occur before FutureBuild activates:

| Phase | WBS Range | Handled By |
|-------|-----------|------------|
| Pre-Design & Feasibility | 1.0-1.9 | Designer/Owner externally |
| Schematic Design | 2.0-2.6 | Designer externally |
| Design Development | 3.0-3.8 | Designer/Engineer externally |
| Construction Documents | 4.0-4.6 | Designer/Engineer externally |
| Permitting | 5.0-5.2 | AHJ (permit review) |

### 2.3 Inspection Checkpoints

Inspections act as **hard gates** that block all dependent tasks until passed:

| WBS | Inspection | Blocks |
|-----|------------|--------|
| 8.2 | Footings | 8.3 Pour footings |
| 8.5 | Foundation walls & steel | 8.6 Pour foundation |
| 8.9 | Under-slab/groundwork | 8.10 Slab pour |
| 9.7 | Framing/Sheathing | 10.0-10.3 All rough-ins |
| 10.4 | Rough-in (P/M/E) | 11.0 Insulation |
| 11.1 | Insulation/Air barrier | 11.3 Drywall hang |
| 14.1 | Finals (building/P/M/E) | 14.3 Punch walk-through |

**Checkpoint Logic:**
- Inspection tasks cannot be marked "passed" until inspector sign-off is recorded
- All successor tasks remain in "blocked" status until checkpoint clears
- System prevents scheduling dependent work before inspection date

### 2.4 Weather-Sensitive Phases (SWIM Model)

Weather adjustments apply **only to pre-dry-in phases** and are driven by the `WeatherService` interface (defined in [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md)):

| Phase | WBS Range | Weather Factors | WeatherService Mapping |
|-------|-----------|-----------------|------------------------|
| Site Prep | 7.2-7.4 | Rain delays clearing/grading | `GetForecast().PrecipitationProbability` |
| Excavation | 8.0 | Rain, ground conditions | `GetForecast().PrecipitationMM` |
| Foundation | 8.1-8.10 | Concrete pour temp limits, rain | `GetForecast().HighTempC / LowTempC` |
| Framing | 9.0-9.4 | Rain delays outdoor work | `GetForecast().Conditions` |
| Dry-In | 9.4-9.6 | Must complete before weather impacts interior | `GetForecast().PrecipitationProbability` |

**Post-dry-in phases (10.0+) are interior work and not weather-adjusted.**

### 2.5 Long-Lead Procurement Items

Items ordered during 6.x with calculated order dates:

| WBS | Item | Typical Lead Time | Order Date Calculation |
|-----|------|-------------------|------------------------|
| 6.0 | Roof trusses | 4-8 weeks | Needed by 9.3, order after 3.8 |
| 6.1 | Windows & exterior doors | 8-20 weeks | Needed by 9.5, order after 4.1 |
| 6.2 | HVAC equipment | 6-12 weeks | Needed by 10.1, order after 4.4 |
| 6.3 | Plumbing fixtures | 6-12 weeks | Needed by 10.0/12.7, order staged |
| 6.4 | Electrical fixtures & panels | 6-12 weeks | Needed by 10.2/12.8, order staged |
| 6.5 | Cabinetry & millwork | 10-16 weeks | Needed by 12.1, order after 4.1 |
| 6.6 | Appliances | 8-16 weeks | Needed by 12.13, order after 3.5 |
| 6.7 | Garage doors | 6-12 weeks | Needed by exterior phase |
| 6.8 | Exterior cladding & roofing | 4-12 weeks | Needed by 9.4/13.0 |
| 6.9 | Interior doors, trim, hardware | 6-12 weeks | Needed by 12.0 |

**Procurement Agent calculates:** `Order Date = Need Date - Lead Time - Buffer Days`

---

## 3. Architecture Layers

### 3.1 Layer 0: Real World Inputs

Entry points for external data ingestion (post-permit):

| Input Source | Data Types | Ingestion Method |
|--------------|------------|------------------|
| Builder/GC | Project specs, overrides, budgets | REST API, Web Forms |
| Documents | Blueprints (PDF), permits, contracts | File upload → Object Storage |
| Site Reports | Daily logs, photos, videos | Mobile upload API |
| External Comms | Email replies, SMS responses | Webhook receivers |

### 3.2 Layer 1: Context Engine (Probabilistic)

AI-powered understanding and extraction:

| Service | Model | Purpose |
|---------|-------|---------|
| Document Processor | Gemini 2.5 Flash | Extract structured data from PDFs |
| Invoice Processor | Gemini 2.5 Flash | Extract {Vendor, Date, Invoice_Number, Total_Amount, Line_Items[]} and predict WBS_Code |
| Context Extractor | Gemini 2.5 Pro | Complex RAG queries, multi-doc reasoning |
| Vision Analyzer | Gemini 2.5 Flash | Site photo/video progress verification |
| Embedding Generator | text-embedding-004 | Vector embeddings for RAG |


### 3.3 Layer 2: Data Spine (PostgreSQL + pgvector)

Persistent storage and vector search:

| Store | Purpose | Key Tables |
|-------|---------|------------|
| Project Context | Extracted variables per project | `project_context`, `project_variables` |
| WBS Library | Master templates (CPM-res1.0) | `wbs_templates`, `wbs_phases`, `wbs_tasks` |
| Physics Params | Calibrated multipliers with weights | `duration_multipliers`, `weather_impacts` |
| Live Project Graph | Real-time DAG of tasks | `project_tasks`, `task_dependencies` |
| Budget/Costs | Allocations and actuals | `budgets`, `cost_actuals`, `invoices` |
| Vector Store | Document embeddings | `document_chunks` (pgvector) |

### 3.4 Layer 3: Physics Engine (Deterministic)

Pure Go calculations with no AI uncertainty:

| Calculator | Algorithm | Output |
|------------|-----------|--------|
| DHSM | Duration = Base × Σ(Variable × Weight) | Task durations |
| SWIM | Weather adjustment (pre-dry-in only) | Duration modifiers |
| gonum/graph CPM | Critical Path Method on DAG | Schedule, critical path, float |
| Procurement Calculator | Need Date - Lead Time - Buffer | Optimal order dates |
| Checkpoint Validator | Block successors until inspection passes | Task status gates |

### 3.5 Layer 4: Action Engine (Agents)

Autonomous agents for orchestration and communication:

| Agent | Purpose | Primary Actions |
|-------|---------|-----------------|
| **Daily Focus Agent** (Superintendent) | Daily prioritized site management | Generate daily briefings, monitor critical path, check weather/inspection gates |
| **Procurement Agent** (Supply Chain) | material availability guard | Calculate order windows, monitor volatility, enforce "Ghost Predecessor" logic |
| **Chat Orchestrator** (Interface) | Primary user communication layer | Intent classification, RAG-based query response, tool/API execution |
| **Subcontractor Liaison** (Outbound SMS) | Trade partner coordination | Confirmation of site arrival, virtual PM status checks, photo collection |
| **Client Reporter** (Weekly Drafter) | Stakeholder automated reporting | Draft weekly progress summaries, curating verified project photos |

### 3.6 Layer 5: Learning Layer

Continuous improvement from feedback:

| Store | Scope | Purpose |
|-------|-------|---------|
| Org Training | Per-organization | Capture overrides, corrections, actual durations |
| Global Training | Cross-org (anonymized) | Improve base model multipliers |

---

## 4. Database Schema

### 4.1 Multi-Tenancy Model

```
Organization (tenant)
├── Users (builders, admins)
├── Projects
│   ├── Project Context (extracted variables)
│   ├── Tasks (instantiated from WBS)
│   ├── Documents
│   ├── Budget/Costs
│   ├── Communications
│   └── Portal Users (clients, subs)
└── Org Settings & Training Data
```

### 4.2 Core Tables

#### Organizations & Users

```sql
organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    slug VARCHAR(100) UNIQUE,
    settings JSONB,
    created_at TIMESTAMP
)

users (
    id UUID PRIMARY KEY,
    org_id UUID REFERENCES organizations,
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    role ENUM('Admin', 'Builder', 'Client', 'Subcontractor'),
    created_at TIMESTAMP
)

auth_tokens (
    token_hash VARCHAR(64) PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    used BOOLEAN DEFAULT FALSE
)

portal_users (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    email VARCHAR(255),
    name VARCHAR(255),
    role ENUM('client', 'subcontractor', 'vendor'),
    contact_preference ENUM('portal', 'email', 'sms', 'both'),
    phone VARCHAR(20),
    created_at TIMESTAMP
)
```

#### Projects & Context

```sql
projects (
    id UUID PRIMARY KEY,
    org_id UUID REFERENCES organizations,
    name VARCHAR(255),
    address TEXT,
    permit_issued_date DATE,
    target_end_date DATE,
    gsf FLOAT,
    status ENUM('Preconstruction', 'Active', 'Paused', 'Completed'),
    created_at TIMESTAMP
)

project_context (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    variable_key VARCHAR(100),
    variable_value TEXT,
    variable_type ENUM('numeric', 'categorical', 'boolean'),
    source ENUM('extracted', 'manual', 'calculated'),
    confidence FLOAT,
    extracted_from UUID REFERENCES documents,
    -- Specific Variables Required by Physics Engine:
    -- supply_chain_volatility (INT, Default 1, Range 1-3)
    -- rough_inspection_latency (INT, Default 1)
    -- final_inspection_latency (INT, Default 5)
    updated_at TIMESTAMP
)
```

#### WBS Library (Master Templates)

```sql
wbs_templates (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    version VARCHAR(50),
    is_default BOOLEAN,
    entry_point_wbs VARCHAR(20) DEFAULT '5.2',
    created_at TIMESTAMP
)

wbs_phases (
    id UUID PRIMARY KEY,
    template_id UUID REFERENCES wbs_templates,
    code VARCHAR(20),
    name VARCHAR(255),
    is_weather_sensitive BOOLEAN DEFAULT FALSE,
    sort_order INTEGER
)

wbs_tasks (
    id UUID PRIMARY KEY,
    phase_id UUID REFERENCES wbs_phases,
    code VARCHAR(50),
    name VARCHAR(255),
    base_duration_days FLOAT,
    responsible_party VARCHAR(100),
    deliverable VARCHAR(255),
    notes TEXT,
    is_inspection BOOLEAN DEFAULT FALSE,
    is_milestone BOOLEAN DEFAULT FALSE,
    is_long_lead BOOLEAN DEFAULT FALSE,
    lead_time_weeks_min INTEGER,
    lead_time_weeks_max INTEGER,
    predecessor_codes TEXT[],
    created_at TIMESTAMP
)
```

#### Duration Multipliers (Weighted Variables)

```sql
duration_multipliers (
    id UUID PRIMARY KEY,
    org_id UUID,
    wbs_task_code VARCHAR(50),
    variable_key VARCHAR(100),
    weight FLOAT,
    multiplier_formula TEXT,
    min_value FLOAT,
    max_value FLOAT,
    source ENUM('default', 'org_trained', 'global_trained'),
    confidence FLOAT,
    updated_at TIMESTAMP
)
```

#### Live Project Graph

```sql
project_tasks (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    wbs_task_id UUID REFERENCES wbs_tasks,
    wbs_code VARCHAR(50),
    assigned_to UUID REFERENCES portal_users,
    status ENUM('pending', 'ready', 'blocked', 'in_progress', 'inspection_pending', 'completed'),
    planned_start DATE,
    planned_end DATE,
    actual_start DATE,
    actual_end DATE,
    calculated_duration_days FLOAT,
    weather_adjusted_duration_days FLOAT, -- Storage for SWIM output
    manual_override_days FLOAT, -- Storage for User Overrides (Nullable)
    override_reason TEXT,
    is_on_critical_path BOOLEAN DEFAULT FALSE,
    total_float_days FLOAT,
    verified_by_vision BOOLEAN DEFAULT FALSE, -- From Validation Protocol
    verification_confidence FLOAT, -- From Gemini Flash (0.0-1.0)
    is_inspection BOOLEAN, -- Critical for "Event Duration Locking" rule
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)

task_dependencies (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    predecessor_id UUID REFERENCES project_tasks,
    successor_id UUID REFERENCES project_tasks,
    dependency_type ENUM('FS', 'SS', 'FF', 'SF') DEFAULT 'FS',
    lag_days INTEGER DEFAULT 0,
    is_inspection_gate BOOLEAN DEFAULT FALSE
)

task_progress (
    id UUID PRIMARY KEY,
    task_id UUID REFERENCES project_tasks,
    reported_by UUID,
    reported_at TIMESTAMP,
    percent_complete INTEGER,
    notes TEXT,
    verified_by_vision BOOLEAN DEFAULT FALSE,
    verification_confidence FLOAT
)

inspection_records (
    id UUID PRIMARY KEY,
    task_id UUID REFERENCES project_tasks,
    inspector_name VARCHAR(255),
    inspection_date DATE,
    result ENUM('pending', 'passed', 'failed', 'conditional'),
    notes TEXT,
    recorded_by UUID REFERENCES users,
    recorded_at TIMESTAMP
)
```

#### Procurement

```sql
procurement_items (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    wbs_task_id UUID REFERENCES wbs_tasks,
    item_name VARCHAR(255),
    vendor_id UUID REFERENCES portal_users,
    lead_time_weeks INTEGER,
    need_by_date DATE,
    calculated_order_date DATE,
    actual_order_date DATE,
    expected_delivery_date DATE,
    actual_delivery_date DATE,
    status ENUM('not_ordered', 'ordered', 'in_transit', 'delivered', 'installed'),
    po_number VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP
)
```

#### Documents & RAG

```sql
documents (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    type ENUM('blueprint', 'permit', 'contract', 'work_order', 'design_selection', 'invoice', 'site_photo', 'site_video', 'other'),
    filename VARCHAR(255),
    storage_path TEXT,
    mime_type VARCHAR(100),
    file_size_bytes BIGINT,
    processing_status ENUM('pending', 'processing', 'completed', 'failed'),
    extracted_text TEXT,
    metadata JSONB,
    uploaded_by UUID REFERENCES users,
    uploaded_at TIMESTAMP
)

document_chunks (
    id UUID PRIMARY KEY,
    document_id UUID REFERENCES documents,
    chunk_index INTEGER,
    content TEXT,
    embedding VECTOR(768),
    metadata JSONB,
    created_at TIMESTAMP
)
```

#### Budget & Costs

```sql
budgets (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    phase_code VARCHAR(20),
    vendor_id UUID REFERENCES portal_users,
    allocated_amount DECIMAL(12,2),
    notes TEXT,
    created_at TIMESTAMP
)

project_budget (
    project_id UUID REFERENCES projects,
    wbs_phase_id UUID REFERENCES wbs_phases,
    estimated_cost FLOAT,
    committed_cost FLOAT,
    actual_cost FLOAT,
    PRIMARY KEY (project_id, wbs_phase_id)
)

cost_actuals (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    phase_code VARCHAR(20),
    vendor_id UUID REFERENCES portal_users,
    amount DECIMAL(12,2),
    description TEXT,
    invoice_document_id UUID REFERENCES documents,
    recorded_at TIMESTAMP
)

invoice_log (
    id UUID PRIMARY KEY,
    file_url TEXT,
    parsed_data JSONB,
    status ENUM('pending', 'approved') DEFAULT 'pending',
    assigned_wbs_id UUID REFERENCES wbs_tasks,
    created_at TIMESTAMP
)
```

#### Contacts & Assignments

```sql
contacts (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    company VARCHAR(255),
    role VARCHAR(100), -- e.g., 'Plumber', 'Electrician'
    phone VARCHAR(20),
    email VARCHAR(255),
    created_at TIMESTAMP
)

project_assignments (
    project_id UUID REFERENCES projects,
    contact_id UUID REFERENCES contacts,
    assigned_phase_id UUID REFERENCES wbs_phases, -- Links a person to a specific WBS Phase
    PRIMARY KEY (project_id, contact_id, assigned_phase_id)
)
```

#### Communications

```sql
communications (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    direction ENUM('inbound', 'outbound'),
    channel ENUM('email', 'sms', 'portal'),
    recipient_type ENUM('client', 'subcontractor', 'vendor'),
    recipient_id UUID REFERENCES portal_users,
    subject TEXT,
    body TEXT,
    status ENUM('pending', 'sent', 'delivered', 'failed', 'received'),
    external_id VARCHAR(255),
    sent_at TIMESTAMP,
    created_at TIMESTAMP
)

communication_log (
    id UUID PRIMARY KEY,
    contact_id UUID REFERENCES contacts,
    direction ENUM('inbound', 'outbound'),
    content TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)

communication_templates (
    id UUID PRIMARY KEY,
    org_id UUID REFERENCES organizations,
    name VARCHAR(255),
    trigger_event VARCHAR(100),
    channel ENUM('email', 'sms'),
    subject_template TEXT,
    body_template TEXT,
    is_active BOOLEAN DEFAULT TRUE
)
```

#### Weather Data

```sql
weather_impacts (
    id UUID PRIMARY KEY,
    region VARCHAR(100),
    month INTEGER,
    trade_category VARCHAR(100),
    impact_multiplier FLOAT,
    precipitation_threshold_mm FLOAT,
    temperature_min_c FLOAT,
    temperature_max_c FLOAT
)

weather_forecasts (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    forecast_date DATE,
    high_temp_c FLOAT,
    low_temp_c FLOAT,
    precipitation_mm FLOAT,
    precipitation_probability FLOAT,
    conditions VARCHAR(100),
    fetched_at TIMESTAMP
)
```

#### Learning Layer & Org Personalization

```sql
org_learning (
    id UUID PRIMARY KEY,
    org_id UUID REFERENCES organizations,
    bias_vector JSONB, -- Stores {Phase_ID: Bias_Factor} map (Layer 5)
    updated_at TIMESTAMP
)

org_training_events (
    id UUID PRIMARY KEY,
    org_id UUID REFERENCES organizations,
    project_id UUID REFERENCES projects,
    event_type ENUM('duration_override', 'actual_vs_planned', 'feedback'),
    wbs_task_code VARCHAR(50),
    original_value JSONB,
    actual_value JSONB,
    delta_percent FLOAT,
    context_variables JSONB,
    created_at TIMESTAMP
)
```

---

## 5. API Design & Service Interfaces

### 5.1 Service Interfaces (Required)
The Go application MUST implement the following internal service interfaces as defined in [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md):
- **VertexClient**:
    ```go
    type Client interface {
        GenerateContent(ctx context.Context, modelType ModelType, parts ...genai.Part) (string, error)
        GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
        Close() error
    }
    ```
- **WeatherService**: Logic for SWIM model adjustments.
- **VisionService**: Image-based verification for the Validation Protocol.
- **NotificationService**: SMS and system notification delivery.
- **DirectoryService**: Unified lookup for project contacts and assignments.

### 5.2 API Structure
All API contracts, including request/response JSON schemas, are documented in the [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md).

```
/api/v1/
├── /auth
│   ├── POST /magic-link
│   ├── GET  /verify/{token}
│   └── POST /logout
│
├── /organizations
│   ├── GET  /
│   ├── POST /
│   └── /{org_id}/settings
│
├── /projects
│   ├── GET  /
│   ├── POST /
│   ├── GET  /{id}
│   ├── PUT  /{id}
│   ├── GET  /{id}/schedule
│   ├── POST /{id}/recalculate
│   │
│   ├── /{id}/context
│   │   ├── GET  /
│   │   └── PUT  /{key}
│   │
│   ├── /{id}/tasks
│   │   ├── GET  /
│   │   ├── GET  /{task_id}
│   │   ├── PUT  /{task_id}
│   │   ├── POST /{task_id}/progress
│   │   └── POST /{task_id}/inspection
│   │
│   ├── /{id}/procurement
│   │   ├── GET  /
│   │   ├── POST /{item_id}/order
│   │   └── POST /{item_id}/delivery
│   │
│   ├── /{id}/documents
│   │   ├── GET  /
│   │   ├── POST /
│   │   ├── GET  /{doc_id}
│   │   └── POST /{doc_id}/reprocess
│   │
│   ├── /{id}/budget
│   │   ├── GET  /
│   │   ├── POST /allocations
│   │   └── POST /actuals
│   │
│   ├── /{id}/portal-users
│   │   ├── GET  /
│   │   ├── POST /
│   │   └── DELETE /{user_id}
│   │
│   └── /{id}/communications
│       ├── GET  /
│       └── POST /
│
├── /chat
│   └── POST /
│
├── /wbs
│   ├── GET  /templates
│   └── GET  /templates/{id}
│
├── /webhooks
│   ├── POST /twilio
│   └── POST /sendgrid
│
└── /portal
    ├── POST /auth/magic-link
    ├── GET  /project
    ├── GET  /tasks
    ├── POST /tasks/{id}/update
    └── POST /documents
```

### 5.2 Key Endpoints

#### Create Project (Instantiates WBS)

```json
POST /api/v1/projects
{
    "name": "Smith Residence",
    "address": "123 Oak Street, Austin, TX 78701",
    "permit_issued_date": "2025-02-01",
    "wbs_template_id": "uuid-of-cpm-res1"
}

Response:
{
    "id": "project-uuid",
    "name": "Smith Residence",
    "status": "preconstruction",
    "tasks_generated": 80,
    "inspections_count": 7,
    "procurement_items": 10,
    "next_step": "Upload blueprints to extract project context"
}
```

#### Record Inspection Result

```json
POST /api/v1/projects/{id}/tasks/{task_id}/inspection
{
    "result": "passed",
    "inspector_name": "John Smith",
    "inspection_date": "2025-03-15",
    "notes": "Footings approved as designed"
}

Response:
{
    "task_id": "task-uuid",
    "inspection_result": "passed",
    "unblocked_tasks": ["8.3", "8.4"],
    "schedule_updated": true
}
```

#### Get Schedule with Critical Path

```json
GET /api/v1/projects/{id}/schedule

Response:
{
    "project_id": "uuid",
    "calculated_at": "2025-01-15T10:30:00Z",
    "permit_issued_date": "2025-02-01",
    "projected_end_date": "2025-07-15",
    "total_duration_days": 165,
    "critical_path": ["7.4", "8.0", "8.1", "8.2", "8.3", ...],
    "blocked_tasks": [
        {"wbs": "8.3", "blocked_by": "8.2 (inspection pending)"}
    ],
    "phases": [...]
}
```

---

## 6. Physics Engine

### 6.1 DHSM Calculator (Duration-Hours-Scope-Multiplier)

Duration calculation using weighted project variables:

```go
func CalculateTaskDuration(task WBSTask, context map[string]interface{}, multipliers []Multiplier) float64 {
    baseDuration := task.BaseDurationDays
    
    for _, mult := range multipliers {
        if mult.WBSTaskCode == task.Code || mult.WBSTaskCode == "*" {
            variableValue, exists := context[mult.VariableKey]
            if exists {
                adjustment := applyMultiplierFormula(
                    variableValue,
                    mult.Weight,
                    mult.MultiplierFormula,
                )
                baseDuration *= adjustment
            }
        }
    }
    
    return math.Round(baseDuration*10) / 10
}
```

**Example Multipliers:**

| Task | Variable | Weight | Formula |
|------|----------|--------|---------|
| 9.1 Floor framing | square_footage | 0.15 | `1 + (value - 2000) / 10000 * weight` |
| 9.2 Wall framing | floors | 0.3 | `1 + (value - 1) * weight` |
| 11.3 Drywall | square_footage | 0.2 | `1 + (value - 2000) / 10000 * weight` |
| 12.4 Tile | bathroom_count | 0.25 | `1 + (value - 2) * weight` |

### 6.2 SWIM Model (Weather - Pre-Dry-In Only)

```go
func ApplyWeatherAdjustment(
    task ProjectTask,
    project Project,
    forecast WeatherForecast,
) float64 {
    if !task.Phase.IsWeatherSensitive {
        return task.CalculatedDurationDays
    }
    
    if task.WBSCode >= "10.0" {
        return task.CalculatedDurationDays
    }
    
    multiplier := 1.0
    
    if forecast.PrecipitationMM > 10 {
        multiplier *= 1.15
    }
    
    if forecast.LowTempC < 0 {
        multiplier *= 1.25 // Frozen ground delays
    }
    
    if forecast.HighTempC > 35 {
        multiplier *= 1.1 // Heat restrictions
    }
    
    return task.CalculatedDurationDays * multiplier
}
```

### 6.3 CPM Calculator (Critical Path Method)

Uses gonum/graph for DAG operations:

```go
import (
    "gonum.org/v1/gonum/graph"
    "gonum.org/v1/gonum/graph/simple"
    "gonum.org/v1/gonum/graph/topo"
)

type CPMResult struct {
    Tasks        []TaskSchedule
    CriticalPath []string
    ProjectEnd   time.Time
}

type TaskSchedule struct {
    TaskID      string
    EarlyStart  time.Time
    EarlyFinish time.Time
    LateStart   time.Time
    LateFinish  time.Time
    TotalFloat  float64
    IsCritical  bool
}

func CalculateCPM(tasks []ProjectTask, deps []TaskDependency, startDate time.Time) CPMResult {
    g := simple.NewDirectedGraph()
    
    // Build graph from tasks and dependencies
    nodeMap := make(map[string]graph.Node)
    for _, task := range tasks {
        node := g.NewNode()
        g.AddNode(node)
        nodeMap[task.ID] = node
    }
    
    for _, dep := range deps {
        from := nodeMap[dep.PredecessorID]
        to := nodeMap[dep.SuccessorID]
        g.SetEdge(g.NewEdge(from, to))
    }
    
    // Topological sort for processing order
    sorted, err := topo.Sort(g)
    if err != nil {
        // Handle cycle error
    }
    
    // Forward pass - calculate early start/finish
    earlyDates := forwardPass(sorted, tasks, deps, startDate)
    
    // Backward pass - calculate late start/finish
    lateDates := backwardPass(sorted, tasks, deps, earlyDates)
    
    // Identify critical path (zero float)
    criticalPath := identifyCriticalPath(tasks, earlyDates, lateDates)
    
    return CPMResult{
        Tasks:        buildSchedule(tasks, earlyDates, lateDates),
        CriticalPath: criticalPath,
        ProjectEnd:   findProjectEnd(earlyDates),
    }
}
```

### 6.4 Procurement Calculator

```go
func CalculateOrderDate(item ProcurementItem, needByDate time.Time, bufferDays int) time.Time {
    leadTimeDays := item.LeadTimeWeeks * 7
    totalLeadTime := leadTimeDays + bufferDays
    
    orderDate := needByDate.AddDate(0, 0, -totalLeadTime)
    
    // Adjust for weekends
    for orderDate.Weekday() == time.Saturday || orderDate.Weekday() == time.Sunday {
        orderDate = orderDate.AddDate(0, 0, -1)
    }
    
    return orderDate
}
```

---

## 7. Agent Architecture

### 7.1 Chat Orchestrator

```go
type ChatOrchestrator struct {
    llmClient    *vertexai.Client
    tools        []Tool
    projectRepo  ProjectRepository
    taskRepo     TaskRepository
}

type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

func (co *ChatOrchestrator) ProcessMessage(ctx context.Context, projectID, userMessage string) (*ChatResponse, error) {
    // 1. Get project context
    project, err := co.projectRepo.GetByID(ctx, projectID)
    if err != nil {
        return nil, err
    }
    
    // 2. Build system prompt with project context
    systemPrompt := co.buildSystemPrompt(project)
    
    // 3. Call LLM with tools
    response, err := co.llmClient.Chat(ctx, vertexai.ChatRequest{
        SystemPrompt: systemPrompt,
        UserMessage:  userMessage,
        Tools:        co.tools,
    })
    if err != nil {
        return nil, err
    }
    
    // 4. Execute any tool calls
    if len(response.ToolCalls) > 0 {
        toolResults := co.executeTools(ctx, response.ToolCalls)
        response = co.llmClient.ContinueWithToolResults(ctx, toolResults)
    }
    
    // 5. Generate artifacts if needed
    artifacts := co.generateArtifacts(response)
    
    return &ChatResponse{
        Message:   response.Content,
        Artifacts: artifacts,
    }, nil
}
```

### 7.2 Daily Focus Agent

```go
type DailyFocusAgent struct {
    scheduler    *asynq.Scheduler
    projectRepo  ProjectRepository
    taskRepo     TaskRepository
    commsService *CommunicationsService
    llmClient    *vertexai.Client
}

func (agent *DailyFocusAgent) GenerateBriefing(ctx context.Context, projectID string) error {
    project, _ := agent.projectRepo.GetByID(ctx, projectID)
    
    // Get today's focus tasks
    todayTasks, _ := agent.taskRepo.GetTasksForDate(ctx, projectID, time.Now())
    
    // Get upcoming inspections
    inspections, _ := agent.taskRepo.GetUpcomingInspections(ctx, projectID, 7)
    
    // Get procurement alerts
    procurementAlerts, _ := agent.getProcurementAlerts(ctx, projectID)
    
    // Get weather forecast
    weather, _ := agent.getWeatherForecast(ctx, project.Latitude, project.Longitude)
    
    // Generate briefing with LLM
    briefing, _ := agent.llmClient.Generate(ctx, vertexai.GenerateRequest{
        Prompt: agent.buildBriefingPrompt(todayTasks, inspections, procurementAlerts, weather),
    })
    
    // Send to superintendent
    return agent.commsService.SendBriefing(ctx, project.SuperintendentEmail, briefing)
}
```

### 7.3 Asynq Task Queue Setup

```go
import (
    "github.com/hibiken/asynq"
)

func SetupTaskQueue(redisAddr string) *asynq.Server {
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: redisAddr},
        asynq.Config{
            Concurrency: 10,
            Queues: map[string]int{
                "critical": 6,
                "default":  3,
                "low":      1,
            },
        },
    )
    
    return srv
}

func SetupScheduler(redisAddr string) *asynq.Scheduler {
    scheduler := asynq.NewScheduler(
        asynq.RedisClientOpt{Addr: redisAddr},
        &asynq.SchedulerOpts{},
    )
    
    // Daily briefing at 6 AM for each active project
    scheduler.Register("0 6 * * *", asynq.NewTask("daily_briefing", nil))
    
    // Procurement check every 4 hours
    scheduler.Register("0 */4 * * *", asynq.NewTask("procurement_check", nil))
    
    // Weather forecast update every 6 hours
    scheduler.Register("0 */6 * * *", asynq.NewTask("weather_update", nil))
    
    return scheduler
}
```

---

## 8. Communication System

### 8.1 Email Service (SendGrid)

```go
import (
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService struct {
    client *sendgrid.Client
    from   *mail.Email
}

func (s *EmailService) SendMagicLink(ctx context.Context, email, token string) error {
    to := mail.NewEmail("", email)
    subject := "Your FutureBuild Login Link"
    
    htmlContent := fmt.Sprintf(`
        <p>Click the link below to log in to FutureBuild:</p>
        <a href="https://app.futurebuild.com/auth/verify/%s">Log In</a>
        <p>This link expires in 15 minutes.</p>
    `, token)
    
    message := mail.NewSingleEmail(s.from, subject, to, "", htmlContent)
    _, err := s.client.Send(message)
    return err
}

func (s *EmailService) SendDailyBriefing(ctx context.Context, email, briefing string, project Project) error {
    to := mail.NewEmail("", email)
    subject := fmt.Sprintf("Daily Focus: %s - %s", project.Name, time.Now().Format("Jan 2"))
    
    message := mail.NewSingleEmail(s.from, subject, to, briefing, briefing)
    _, err := s.client.Send(message)
    return err
}
```

### 8.2 SMS Service (Twilio)

```go
import (
    "github.com/twilio/twilio-go"
    openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type SMSService struct {
    client *twilio.RestClient
    from   string
}

func (s *SMSService) SendTaskReminder(ctx context.Context, phone, taskName, dueDate string) error {
    params := &openapi.CreateMessageParams{}
    params.SetTo(phone)
    params.SetFrom(s.from)
    params.SetBody(fmt.Sprintf(
        "FutureBuild Reminder: %s is due on %s. Reply DONE when complete.",
        taskName, dueDate,
    ))
    
    _, err := s.client.Api.CreateMessage(params)
    return err
}
```

---

## 9. Project Structure

```
futurebuild/
├── cmd/
│   ├── api/
│   │   └── main.go           # API server entrypoint
│   └── worker/
│       └── main.go           # Asynq worker entrypoint
├── internal/
│   ├── api/
│   │   ├── router.go         # Chi router setup
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   └── logging.go
│   │   └── handlers/
│   │       ├── auth.go
│   │       ├── projects.go
│   │       ├── tasks.go
│   │       ├── chat.go
│   │       └── webhooks.go
│   ├── domain/
│   │   ├── models/
│   │   │   ├── organization.go
│   │   │   ├── user.go
│   │   │   ├── project.go
│   │   │   ├── task.go
│   │   │   └── document.go
│   │   └── services/
│   │       ├── project_service.go
│   │       ├── scheduling_service.go
│   │       ├── document_service.go
│   │       └── communication_service.go
│   ├── physics/
│   │   ├── dhsm.go           # Duration calculator
│   │   ├── swim.go           # Weather model
│   │   ├── cpm.go            # Critical path
│   │   └── procurement.go    # Order date calculator
│   ├── agents/
│   │   ├── orchestrator.go   # Chat orchestrator
│   │   ├── daily_focus.go    # Daily briefing agent
│   │   ├── procurement.go    # Procurement agent
│   │   └── tools/
│   │       ├── schedule.go
│   │       ├── tasks.go
│   │       └── documents.go
│   ├── repository/
│   │   ├── postgres/
│   │   │   ├── organization.go
│   │   │   ├── project.go
│   │   │   ├── task.go
│   │   │   └── document.go
│   │   └── interfaces.go
│   ├── ai/
│   │   ├── vertexai/
│   │   │   ├── client.go
│   │   │   ├── embeddings.go
│   │   │   └── chat.go
│   │   └── rag/
│   │       ├── pipeline.go
│   │       └── retriever.go
│   └── config/
│       └── config.go
├── pkg/
│   ├── storage/
│   │   └── s3.go             # S3-compatible storage
│   ├── email/
│   │   └── sendgrid.go
│   └── sms/
│       └── twilio.go
├── migrations/
│   ├── 001_initial_schema.up.sql
│   ├── 001_initial_schema.down.sql
│   └── ...
├── scripts/
│   └── seed_wbs.go           # WBS template seeder
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── Makefile
```

---

## 10. Configuration

### 10.1 Environment Variables

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/futurebuild?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# AI
GOOGLE_CLOUD_PROJECT=futurebuild-prod
VERTEX_AI_LOCATION=us-central1

# Storage
S3_ENDPOINT=https://nyc3.digitaloceanspaces.com
S3_BUCKET=futurebuild-docs
S3_ACCESS_KEY=xxx
S3_SECRET_KEY=xxx

# Email
SENDGRID_API_KEY=xxx
EMAIL_FROM=noreply@futurebuild.com

# SMS
TWILIO_ACCOUNT_SID=xxx
TWILIO_AUTH_TOKEN=xxx
TWILIO_PHONE_NUMBER=+1234567890

# Auth
JWT_SECRET=xxx
MAGIC_LINK_EXPIRY=15m
```

### 10.2 Config Struct

```go
type Config struct {
    Port        int    `env:"PORT" envDefault:"8080"`
    Environment string `env:"ENV" envDefault:"development"`
    
    DatabaseURL string `env:"DATABASE_URL,required"`
    RedisURL    string `env:"REDIS_URL,required"`
    
    GoogleCloudProject string `env:"GOOGLE_CLOUD_PROJECT,required"`
    VertexAILocation   string `env:"VERTEX_AI_LOCATION" envDefault:"us-central1"`
    
    S3Endpoint  string `env:"S3_ENDPOINT,required"`
    S3Bucket    string `env:"S3_BUCKET,required"`
    S3AccessKey string `env:"S3_ACCESS_KEY,required"`
    S3SecretKey string `env:"S3_SECRET_KEY,required"`
    
    SendGridAPIKey string `env:"SENDGRID_API_KEY,required"`
    EmailFrom      string `env:"EMAIL_FROM,required"`
    
    TwilioAccountSID  string `env:"TWILIO_ACCOUNT_SID,required"`
    TwilioAuthToken   string `env:"TWILIO_AUTH_TOKEN,required"`
    TwilioPhoneNumber string `env:"TWILIO_PHONE_NUMBER,required"`
    
    JWTSecret       string        `env:"JWT_SECRET,required"`
    MagicLinkExpiry time.Duration `env:"MAGIC_LINK_EXPIRY" envDefault:"15m"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

---

## 11. Dependencies

### 11.1 go.mod

```go
module github.com/futurebuild/futurebuild

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.12
    github.com/go-chi/cors v1.2.1
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/google/uuid v1.6.0
    github.com/hibiken/asynq v0.24.1
    github.com/jackc/pgx/v5 v5.5.3
    github.com/pgvector/pgvector-go v0.1.1
    github.com/sendgrid/sendgrid-go v3.14.0
    github.com/twilio/twilio-go v1.19.0
    github.com/caarlos0/env/v10 v10.0.0
    cloud.google.com/go/vertexai v0.7.0
    gonum.org/v1/gonum v0.14.0
    golang.org/x/sync v0.6.0
)
```

---

*Document Version: 1.0.0*

---

## 19. Business Rules

### 19.3 SaaS Limits (The Gatekeeper)
Define the logic: "Before POST /projects: Query COUNT(*) active_projects. If count >= org.project_limit, return 403 Forbidden LIMIT_REACHED."
