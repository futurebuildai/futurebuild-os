# CPM-res1.0 Model Specification
## Deterministic Logic & AI Specification

**Version:** 1.1.0  
**Status:** Refined Specification (Integrated DHSM SAF Corrections)

---

## 4. Domain Model (→ Refined Section 4)

### 4.1 Entity Hierarchy

```
PROJECT
├── PROJECT_CONTEXT (extracted variables)
├── PROJECT_TASK[] (instantiated from WBS_TASK)
│   ├── TASK_DEPENDENCY[]
│   ├── TASK_PROGRESS[]
│   └── INSPECTION_RECORD (if is_inspection=true)
├── PROCUREMENT_ITEM[]
├── DOCUMENT[]
│   └── DOCUMENT_CHUNK[] (for RAG)
├── BUDGET[]
├── COST_ACTUAL[]
├── COMMUNICATION[]
└── PORTAL_USER[] (external: client, subcontractor, vendor)
```

### 4.2 Core Entities

#### 4.2.3 PROJECT
| Attribute | Type | Description |
|-----------|------|-------------|
| id | UUID | Primary key |
| status | enum | preconstruction, active, paused, completed |
| permit_issued_date | date | System activation date |
| target_end_date | date | Projected completion |
| square_footage | float | From Project Context (GSF) |

#### 4.2.4 PROJECT_TASK
| Attribute | Type | Description |
|-----------|------|-------------|
| calculated_duration_days | float | Base duration after DHSM |
| weather_adjusted_duration_days | float | After SWIM adjustment |
| manual_override_days | float | Builder override (optional) |
| is_on_critical_path | boolean | True if zero float |

---

## 5. Work Breakdown Structure (→ Refined Section 5)

### 5.1 WBS Phase Structure

| Phase Code | Phase Name | Task Count | In Scope |
|------------|------------|------------|----------|
| 6.x | Procurement | 10 | Yes |
| 7.x | Site Prep | 5 | Yes |
| 8.x | Foundation | 12 | Yes |
| 9.x | Framing | 8 | Yes |
| 10.x | Rough-Ins | 6 | Yes |
| 11.x | Insulation/Drywall | 6 | Yes |
| 12.x | Interior Finishes | 15 | Yes |
| 13.x | Exterior | 7 | Yes |
| 14.x | Commissioning/Closeout | 8 | Yes |

### 5.5 Long-Lead Procurement Items (Ghost Predecessors)
These tasks represent materials that must be ordered and delivered before physical work can begin. They are "hard-linked" as Finish-to-Start (FS) predecessors to their respective installation tasks.

| WBS | Item | Hard Dependency Mapping (FS Predecessor To) |
|-----|------|---------------------------------------------|
| 6.0 | Trusses | WBS 9.3 (Roof Framing) |
| 6.1 | Windows | WBS 9.5 (Window Install) |
| 6.2 | HVAC | WBS 10.1 (Mech Rough) |
| 6.5 | Cabinets | WBS 12.1 (Cabinet Install) |
| 6.8 | Siding/Cladding | WBS 13.0 (Siding Install) |

### 5.4 Inspection Checkpoints

| WBS | Inspection Name | Blocks Tasks |
|-----|-----------------|--------------|
| 8.2 | Footings | 8.3+ |
| 8.5 | Foundation walls & steel | 8.6+ |
| 9.7 | Framing/Sheathing | 10.0+ |
| 10.4 | Rough-in (P/M/E) | 11.0+ |

---

## 10. Layer 2: Data Spine (→ Refined Section 10)

### 10.1 Vector Store Configuration

```sql
CREATE TABLE document_chunks (
    id UUID PRIMARY KEY,
    document_id UUID REFERENCES documents(id),
    content TEXT NOT NULL,
    embedding vector(768),
    metadata JSONB
);
```

---

## 11. Layer 3: Physics Engine (→ Refined Section 11)

### 11.1 Purpose
Deterministic calculations for scheduling. No AI uncertainty in this layer.

### 11.2 DHSM Calculator (Duration-Hours-Scope-Multiplier)

#### 11.2.1 Size Adjustment Factor (SAF)
Construction duration scales non-linearly with size. The SAF is calculated relative to a 2,250 GSF baseline.

**Formula:**
`SAF = (New_GSF / 2250) ^ 0.75`

**Event Duration Locking (Inspection Exception):**
While SAF applies to standard construction tasks (WBS 8.0–13.6), it **EXCLUDES** any task where `Type == Inspection`.
- **Constraint:** Inspection tasks (WBS 8.2, 8.5, 8.9, 9.7, 10.4, 11.1, 14.1) must have a fixed `SAF = 1.0` regardless of project size.
- **Note:** Final Inspection (WBS 14.1) may optionally allow a manual duration override (e.g., 2 days for estate homes), but the default calculation must not scale automatically.

#### 11.2.2 Global Volatility Variables
These variables adjust the safety buffers and lead times based on external market conditions.

| Variable | Baseline | Range | Description |
|----------|----------|-------|-------------|
| `supply_chain_volatility` | 1 | 1 (Stable) to 3 (Volatile) | Risk multiplier for material lead times |

#### 11.2.3 Structural Adder Logic
Structural adders are physical volume delays and must be scaled by the SAF.

**Corrected Calculation:**
`Duration = Base_Duration + (Adder_Base_Days * SAF)`

#### 11.2.3 Parameter Table: Structural Adders (Subject to SAF)
| Variable | Base Adder (Days) | Application Area |
|----------|-------------------|------------------|
| Foundation: Crawlspace | 8 | WBS 8.0–8.11 |
| Foundation: Basement | 25 | WBS 8.0–8.11 |
| Topography: Moderate Slope | 10 | WBS 7.3, 8.0 |
| Topography: Steep Slope | 30 | WBS 7.3, 8.0 |
| Soil Conditions: Poor/Rock | 15 | WBS 8.0 |

| Variable | Adjustment | Notes |
|----------|------------|-------|
| Regulatory Hurdles | +30 to 180 days | Fixed Adder (Static) |
| `rough_inspection_latency` | +1 day base | Applied to WBS 8.2, 8.5, 9.7, 10.4 |
| `final_inspection_latency` | +5 days base | Applied to WBS 14.1 (Finals) |

#### 11.2.5 Final DHSM Formula
The final duration for any task incorporates the volume adjustments, market volatility, and organization-specific historical performance, followed by a quantization step for site management.

**Formula:**
1. `Raw_Duration = (D_base * SAF * Multipliers) + Org_Bias[Phase]`
2. `Duration_final = Ceiling(Raw_Duration * 2) / 2`

**Explanation:**
The `Ceiling(x * 2) / 2` function forces all durations to snap to the nearest 0.5-day increment (e.g., 4.2 → 4.5). This aligns the schedule with standard construction shift blocks (Morning/Afternoon).

Where:
- `D_base`: Baseline task duration.
- `SAF`: Size Adjustment Factor (Locked to 1.0 for Inspections).
- `Multipliers`: Cumulative product of SWIM, Supply Chain, and other environmental factors.
- `Org_Bias[Phase]`: The organization's historical deviation for that specific phase.

### 11.5 Procurement Calculator
Calculates Order Dates for materials based on lead times and supply chain volatility.

#### 11.5.1 Procurement Buffer Days
Decouples Supply Chain risk from Labor Market risk.
**Formula:**
`Buffer_Days = Baseline_Buffer (5 days) * supply_chain_volatility`

#### 11.5.2 Order Date Calculation
**Formula:**
`Order_Date = Need_Date - Lead_Time - Buffer_Days`

### 11.4 CPM Solver
Critical Path Method implementation using gonum/graph to solve for Early Start (ES), Early Finish (EF), Late Start (LS), and Late Finish (LF).

---

## 13. Layer 5: Learning Layer (→ Refined Section 13)

### 13.1 Feedback Loops (Multiplier Adjustment Logic)
The system captures the delta between `Calculated_Duration` and `Actual_End` to refine future `Adder_Base_Days`, `Multipliers`, and Organization-specific biases.

### 13.2 Org Bias Vectors (Personalization)
The system tracks how a specific Organization deviates from the Global Baseline to provide personalized scheduling.

- **Variable:** `Org_Bias[Phase_ID]`
- **Update Trigger:** When a `PROJECT_TASK` status changes to `Completed`.
- **Logic:** Define the error delta between reality and the baseline model.
  - `Delta = Actual_Duration - (Baseline_Duration * Global_Multipliers)`
- **Formula (Weighted Moving Average):**
  - `Org_Bias_new = (Org_Bias_old * 0.9) + (Delta * 0.1)`
  - *Note: The 0.9 weighting prevents a single outlier project from significantly distorting the organization's profile.*

### 13.3 Global Weight Calibration (Optimization)
Refines core Physics constants (e.g., SAF, Weather Impact) using anonymized aggregate data across all organizations.

#### 13.3.1 Cluster Variance Check
- **Rule:** If `Count(Projects) > 100` within `Cluster X` (e.g., Climate Zone 5, Size > 4000sf) exhibit a `Mean_Error > 15%` on a specific phase, the system flags the `Base_Multiplier` for recalibration.
- **Goal:** Identifying systemic inaccuracies in the physics engine that require global adjustment.

#### 13.3.2 Outlier Filtering (Data Poisoning Protection)
To ensure the integrity of the Global Model, anomalous data must be excluded from training sets.
- **Rule:** Discard "Anomalies" where `Duration > 3σ` (three standard deviations) from the cluster mean.
- **Purpose:** Prevents data poisoning from extreme edge cases or reporting errors.

---

## 19. Business Rules (→ Refined Section 19)

### 19.1 Inspection Gate Rule
Successor tasks are blocked (`status = 'blocked'`) until the predecessor inspection task is marked `passed`.

### 19.2 Weather Sensitivity Rule (SWIM)
Applied to `WBS < 10.0` OR `WBS 13.x`. Interior work (`WBS >= 10.0` AND `WBS != 13.x`) bypasses weather adjustments. Exterior Finishes (Phase 13) are explicitly weather-dependent and must be adjusted by SWIM logic.

---

## 20. State Machines (→ Refined Section 20)

### 20.2 Task Status Transitions
- `pending` → `ready`: `all_predecessors_complete` AND `NOT blocked_by_inspection`
- `ready` → `in_progress`: `work_started`
- `in_progress` → `completed`: `work_complete`

---

## 21. Decision Trees (→ Refined Section 21)

### 21.1 Chat Intent Classification
Decision matrix for routing user queries to specific system tools (Schedule, Budget, Tasks, etc.).
