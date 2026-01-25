# FutureShade Initialization Technical Specs (Step 64)

## 1. Overview
This specification details the technical barriers and scaffolding required to initialize "FutureShade," the isolated intelligence layer for FutureBuild. The goal is to establish a physical and logical separation between the core application ("The Tree") and the AI governance layer ("The Shadow"). This step focuses purely on initialization and directory structure, setting the stage for the "Tribunal" and "Shadow Protocol" in subsequent steps.

## 2. Architecture

### 2.1 Backend Architecture (`internal/futureshade`)
The backend component will be isolated in `internal/futureshade`.
*   **Dependency Rule**: `internal/futureshade` may import `internal/app` types (read-only interfaces preferred), but `internal/app` services SHOULD NOT import `internal/futureshade` logic components directly, except for the main `server` wiring or via defined interfaces.
*   **Entry Point**: `Service` struct in `service.go`.

**Structure:**
```text
internal/futureshade/
├── service.go         # Service lifecycle (Start, Stop, Health)
├── config.go          # Configuration struct (API Keys, Toggles)
├── types.go           # Shared types (ShadowDoc)
├── tribunal/          # Sub-package for Consensus Logic (Empty for now)
│   └── types.go       # Tribunal-specific types
└── shadow/            # Sub-package for Observer Logic (Empty for now)
    └── observer.go    # Observer interface/stub
```

### 2.2 Frontend Architecture (`frontend/src/futureshade`)
The frontend component mirrors the backend separation.
**Structure:**
```text
frontend/src/futureshade/
├── components/        # React components specific to FutureShade
├── services/          # API clients for FutureShade endpoints
└── types/             # TypeScript interfaces for ShadowDoc/Tribunal
```

### 2.3 Integration
*   `internal/server/server.go`: Initializes `futureshade.Service`.
*   **Health Check**: The main server health check will NOT fail if FutureShade is degraded (soft dependency), but FutureShade will expose its own status.

## 3. API Specification

### 3.1 Internal Interfaces (Go)

**`Service` Interface**
```go
type Service interface {
    Health() error
    // Observe will be added in later steps
}
```

### 3.2 HTTP Endpoints
*   **GET /api/v1/futureshade/health** (Admin Only)
    *   **Response**: `200 OK` if configured, `503 Service Unavailable` if keys missing.
    *   **Payload**: `{"status": "active", "tribunal_count": 3}`

## 4. Data Model

### 4.1 ShadowDoc (Stub)
Defined in `internal/futureshade/types.go`
```go
type ShadowDoc struct {
    ID          string                 `json:"id"`
    SourceType  string                 `json:"source_type"` // "PRD", "Spec", "Code"
    SourceID    string                 `json:"source_id"`   // File path or DB ID
    ContentHash string                 `json:"content_hash"`
    Analysis    map[string]interface{} `json:"analysis"`    // JSONB bucket for AI output
    CreatedAt   time.Time              `json:"created_at"`
}
```

## 5. Security Considerations

### 5.1 Authentication & Authorization
*   **Admin Access Only**: All HTTP endpoints under `/api/v1/futureshade` must be protected by `RequireAdmin` middleware (or equivalent strong auth).
*   **API Keys**: AI Provider keys (Gemini, etc.) must be loaded from `config.Config` and never hardcoded.

### 5.2 Data Privacy
*   **PII Scrubbing**: Future implementations of `Observer` must ensure PII is not sent to external LLMs. (Noted for implementation).

## 6. Testing Strategy

### 6.1 Unit Tests
*   `TestNewService`: Verify configuration loading.
*   `TestHealth`: Verify logic when keys are present vs absent.

### 6.2 Integration Tests
*   `TestServerStartup_WithFutureShade`: Ensure `server.Start()` succeeds with the new service wired up.

## 7. Implementation Notes
*   **Fail Open**: If FutureShade fails to initialize (e.g., missing API key), the main application MUST still start. Log the error and continue in "Shadowless" mode.
*   **NoOp Defaults**: Use the Null Object pattern if the service is disabled, so call sites don't need `if service != nil` checks everywhere.
