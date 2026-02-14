# Sprint 6.2: Observability

> **Epic:** 6 — Production Hardening (Google Audit)
> **Depends On:** Sprint 3.2 (audit_wal foundation)
> **Objective:** Every Agent decision is logged, and distributed tracing connects UI → API → Agent → DB.

---

## Sprint Tasks

### Task 6.2.1: Structured Agent Decision Logging

**Status:** ⬜ Not Started

**Current State:**
- [audit_wal.md](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/chat/audit_wal.md) — placeholder stub
- Sprint 3.2 introduces `audit/wal.go` for correction events — this sprint extends it

**Concept:** Every AI agent decision (extraction, classification, recommendation) must be logged as structured JSON. This creates an audit trail for compliance and debugging.

**Log Schema:**

```json
{
    "id": "uuid",
    "timestamp": "2026-02-13T15:00:00Z",
    "agent": "interrogator",
    "action": "extract_field",
    "input_summary": "User uploaded 3-page PDF: residential_plans.pdf",
    "decision": "Extracted 30 electrical tasks, confidence 0.87",
    "confidence": 0.87,
    "model": "gemini-2.0-flash",
    "latency_ms": 1234,
    "project_id": "proj_123",
    "user_id": "user_456",
    "trace_id": "trace_789"
}
```

**Atomic Steps:**

1. **Define `AgentDecisionEntry` struct** in `backend/internal/audit/`:
   ```go
   type AgentDecisionEntry struct {
       ID            string         `json:"id"`
       Timestamp     time.Time      `json:"timestamp"`
       Agent         string         `json:"agent"`
       Action        string         `json:"action"`
       InputSummary  string         `json:"input_summary"`
       Decision      string         `json:"decision"`
       Confidence    float64        `json:"confidence"`
       Model         string         `json:"model"`
       LatencyMS     int64          `json:"latency_ms"`
       ProjectID     string         `json:"project_id"`
       UserID        string         `json:"user_id"`
       TraceID       string         `json:"trace_id"`
       Metadata      map[string]any `json:"metadata,omitempty"`
   }
   ```

2. **Create `AgentLogger`** interface:
   ```go
   type AgentLogger interface {
       LogDecision(ctx context.Context, entry AgentDecisionEntry) error
       QueryDecisions(ctx context.Context, filter DecisionFilter) ([]AgentDecisionEntry, error)
   }
   ```

3. **Implement PostgreSQL backend:** `audit_decisions` table
4. **Instrument all agent services:** Interrogator, VisionService, ScheduleService
5. **Include `trace_id` from OpenTelemetry** (wired in Task 6.2.2)

---

### Task 6.2.2: Distributed Tracing (OpenTelemetry)

**Status:** ⬜ Not Started

**Concept:** Trace a request from UI → API → Agent → DB using OpenTelemetry. This is critical for debugging latency and understanding the full execution path.

**Architecture:**

```
Browser (Frontend)                              Backend
┌─────────────────┐    HTTP + traceparent    ┌──────────────────┐
│  fb-app-shell   │  ──────────────────────► │  API Handler     │
│  (JS SDK trace) │                          │  (OTel middleware)│
└─────────────────┘                          │    ├── Agent Svc  │
                                             │    │   ├── AI Call │
                                             │    │   └── DB     │
                                             │    └── Response   │
                                             └──────────────────┘
                                                       │
                                                       ▼
                                             ┌──────────────────┐
                                             │  Trace Exporter  │
                                             │  (OTLP → Jaeger/ │
                                             │   Cloud Trace)   │
                                             └──────────────────┘
```

**Atomic Steps:**

**Backend (Go):**

1. **Add OpenTelemetry dependencies:**
   ```
   go.opentelemetry.io/otel
   go.opentelemetry.io/otel/sdk
   go.opentelemetry.io/otel/exporters/otlp/otlptrace
   go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
   ```

2. **Initialize tracer provider** in `main.go`:
   ```go
   tp := sdktrace.NewTracerProvider(
       sdktrace.WithBatcher(exporter),
       sdktrace.WithResource(resource.NewWithAttributes(
           semconv.ServiceNameKey.String("futurebuild-api"),
       )),
   )
   otel.SetTracerProvider(tp)
   ```

3. **Wrap HTTP router** with OTel middleware: `otelhttp.NewHandler(router, "futurebuild")`

4. **Instrument agent calls:**
   ```go
   ctx, span := tracer.Start(ctx, "interrogator.ProcessMessage")
   defer span.End()
   span.SetAttributes(attribute.String("project.id", projectID))
   ```

5. **Instrument DB calls** with `otelsql` or manual spans

**Frontend (optional, nice-to-have):**

6. **Propagate `traceparent` header** from frontend HTTP calls:
   ```ts
   // services/http.ts - add W3C trace context header
   headers['traceparent'] = generateTraceParent();
   ```

---

## Codebase References

| File | Path | Status | Notes |
|------|------|--------|-------|
| audit_wal.md | `backend/shadow/internal/chat/audit_wal.md` | Stub | Extended by this sprint |
| wal.go | `backend/internal/audit/` | Sprint 3.2 | Foundation for agent logging |
| main.go | `backend/cmd/server/` | Existing? | OTel init goes here |
| http.ts | `frontend/src/services/http.ts` | Existing | Optional trace propagation |

## Verification Plan

- **Manual:** Trigger an onboarding flow → check backend logs for structured `AgentDecisionEntry` JSON
- **Manual:** Open Jaeger/Cloud Trace UI → verify request trace from API handler → agent → DB
- **Automated:** Integration test: Make API call → verify `trace_id` appears in agent decision log
- **Manual:** Verify trace spans show latency breakdown (API parsing, AI model call, DB write)
