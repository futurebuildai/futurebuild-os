# API and Types Specification (The Rosetta Stone)

**Version:** 1.0.0
**Status:** Unified Single Source of Truth

---

## 1. Shared Enums (Universal Casing)

All implementations (Go, TypeScript, SQL) must use these exact strings/casing.

### 1.1 TaskStatus
Defines the lifecycle of a `ProjectTask`.
- `Pending`
- `Ready`
- `In_Progress`
- `Inspection_Pending`
- `Completed`
- `Blocked`
- `Delayed`

### 1.2 UserRole
Defines the permissions and identity of a `PortalUser` or `User`.
- `Admin`
- `Builder`
- `Client`
- `Subcontractor`

### 1.3 ArtifactType
Defines the visual components displayed in the Chat Orchestrator.
- `Invoice`
- `Budget_View`
- `Gantt_View`

---

---

## 2. Go Interfaces (Layer 4 Tools)

These define the strict service contracts for the Action Engine's tools.

### 2.1 WeatherService
Integration for the SWIM Model.
```go
type Forecast struct {
    Date                     string  `json:"date"`
    HighTempC                float64 `json:"high_temp_c"`
    LowTempC                 float64 `json:"low_temp_c"`
    PrecipitationMM          float64 `json:"precipitation_mm"`
    PrecipitationProbability float64 `json:"precipitation_probability"`
    Conditions               string  `json:"conditions"`
}

type WeatherService interface {
    GetForecast(lat, long float64) (Forecast, error)
}
```

### 2.2 VisionService
Validation Protocol service.
```go
type VisionService interface {
    // VerifyTask returns (is_verified, confidence_score, error)
    // Context is required for AI inference timeout/cancellation control.
    VerifyTask(ctx context.Context, imageURL string, taskDescription string) (bool, float64, error)
}
```

### 2.3 NotificationService
Outbound communication service.
```go
type NotificationService interface {
    SendSMS(contactID string, message string) error
}
```

### 2.4 DirectoryService
Contact and assignment lookups.

**Interface Contract (Layer 4 Agents):**
```go
type DirectoryService interface {
    GetContactForPhase(phaseID string) (Contact, error)
}
```
> [!NOTE]
> Layer 4 agents resolve `projectID` and `orgID` from the current execution context before calling the service layer.

**Service Layer Signature (Internal):**
```go
func (s *DirectoryService) GetContactForPhase(ctx context.Context, projectID, orgID uuid.UUID, phaseCode string) (*types.Contact, error)
```

### 2.5 VertexClient
AI Infrastructure client.
```go
type Client interface {
    GenerateContent(ctx context.Context, modelType ModelType, parts ...genai.Part) (string, error)
    GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
    Close() error
}
```

---

## 3. API Contracts (Layer 2 ↔ Layer 1)

### 3.1 Document Analysis
`POST /api/v1/documents/analyze`

**Input:** Multi-part Form Data (File)

**Output (InvoiceExtraction):**
```json
{
  "vendor": "String",
  "date": "ISO-8601 Date",
  "invoice_number": "String",
  "total_amount": 0.00,
  "line_items": [
    {
      "description": "String",
      "quantity": 0.0,
      "unit_price": 0.0,
      "total": 0.0
    }
  ],
  "suggested_wbs_code": "String",
  "confidence": 0.95
}
```

### 3.2 Authentication (Magic Link)
`POST /api/v1/auth/login`

**Input (AuthRequest):**
```json
{
  "email": "String"
}
```

**Output (AuthResponse):**
```json
{
  "message": "String"
}
```

`GET /api/v1/auth/verify?token={token}`

**Output:** `User` JSON (See `DATA_SPINE_SPEC.md` Section 2.2)

### 3.3 Project Schedule
`GET /api/v1/projects/{id}/schedule`

**Output (Gantt Data):**
```json
{
  "project_id": "UUID",
  "calculated_at": "ISO-8601 Timestamp",
  "projected_end_date": "ISO-8601 Date",
  "critical_path": ["WBS_CODE_1", "WBS_CODE_2"],
  "tasks": [
    {
      "wbs_code": "String",
      "name": "String",
      "status": "TaskStatus",
      "early_start": "ISO-8601 Date",
      "early_finish": "ISO-8601 Date",
      "duration_days": 0.0,
      "is_critical": true
    }
  ]
}
```

---

## 4. Shared Structs

### 4.1 Contact
```go
type Contact struct {
    ID                string            `json:"id"`
    Name              string            `json:"name"`
    Company           string            `json:"company"`
    Phone             string            `json:"phone"`
    Email             string            `json:"email"`
    Role              UserRole          `json:"role"`
    ContactPreference ContactPreference `json:"contact_preference,omitempty"` // See DATA_SPINE_SPEC.md Section 2.3
}
```

### 4.2 Shared Auth Types
```go
type AuthRequest struct {
    Email string `json:"email"`
}

type AuthResponse struct {
    Message string `json:"message"`
}
```
