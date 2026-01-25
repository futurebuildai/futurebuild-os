# FutureShade Service Initialization (Step 64) - PRD

**Status**: Draft
**Step**: 64
**Feature**: FutureShade Service & "The Tribunal" Foundation

## 1. Executive Summary
"FutureShade" represents the separate intelligence layer of the FutureBuild system. Unlike the standard application logic, FutureShade operates as a "Shadow" service—observing, analyzing, and occasionally intervening via "The Tribunal". This PRD covers the initialization of this service, establishing the directory structure, basic service scaffolding, and the boundaries required for it to operate without polluting the core business logic.

## 2. Problem Statement & Goals
**Problem**: The current system mixes core business logic (CRUD, State) with emerging AI intelligence logic. As we introduce more complex AI features (Consensus protocols, Auto-correction), we risk creating a monolithic "God Object" or circular dependencies if this logic isn't strictly isolated.

**Goal**: Establish a clean physical and logical separation for the AI Intelligence layer.
- **Physical**: `internal/futureshade` (Backend) and `frontend/src/futureshade` (Frontend).
- **Logical**: FutureShade can import `app` types, but `app` services should largely treat FutureShade as an external observer or sink, minimizing reverse dependencies.

## 3. Architecture Definition

### Backend Structure
New directory: `internal/futureshade`
```text
internal/futureshade/
├── service.go         # Main entry point (Initialize, Start)
├── config.go          # AI-specific configuration
├── tribunal/          # "The Tribunal" sub-package (Consensus logic)
│   └── types.go       # Tribunal-specific types
└── shadow/            # "Shadow" observer sub-package
    └── observer.go    # Logic for observing system events
```

### Frontend Structure
New directory: `frontend/src/futureshade`
```text
frontend/src/futureshade/
├── components/        # FutureShade-specific UI (Tribunal logs, Shadow view)
├── services/          # dedicated frontend services for FutureShade APIs
└── types/             # Frontend types for Tribunal/Shadow entities
```

### Integration Points
- **Initialization**: `internal/app/app.go` (or `main.go`) will initialize `futureshade.Service` alongside other core services.
- **Event Bus**: FutureShade will eventually subscribe to the domain event bus (if/when implemented) or wrap specific service method calls via decorators (to be defined in later steps).

## 4. Functional Requirements

### FR1: Service Initialization
- The system MUST be able to initialize the FutureShade service on startup.
- The initialization MUST verify that required API Keys (e.g., Gemini) are present if FutureShade is enabled.

### FR2: Health Check
- FutureShade MUST expose a `Health()` method or endpoint.
- Returns `OK` if the underlying AI clients are configured correctly.

### FR3: Shadow Protocol (Proprietary)
- Define the `ShadowDoc` interface (stub) which will act as the "Record of Truth" for AI analysis.
- *Note*: This is a foundational step; actual implementation of the protocol comes in Step 65.

## 5. Data Models

### ShadowDoc (Stub)
```go
// internal/futureshade/types.go

type ShadowDoc struct {
    ID          string                 `json:"id"`
    SourceType  string                 `json:"source_type"` // e.g., "PRD", "CodeFile"
    SourceID    string                 `json:"source_id"`
    ContentHash string                 `json:"content_hash"`
    Analysis    map[string]interface{} `json:"analysis"`
    CreatedAt   time.Time              `json:"created_at"`
}
```

## 6. Security & Compliance
- **Internal Only**: FutureShade endpoints (if any) MUST be protected by Admin-only middleware.
- **Data Privacy**: FutureShade analysis logs should be scrubbed of PII unless explicitly required for the "Tribunal" cases.

## 7. Success Metrics
- `internal/futureshade` package compiles.
- `frontend/src/futureshade` directory exists.
- Application starts up successfully with FutureShade service initialized (even if doing nothing).
- Go tests pass for the new package.
