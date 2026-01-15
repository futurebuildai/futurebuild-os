# Remediation Spec: Handler Test Coverage

**Status:** Proposed  
**Target Phase:** Phase 8 (Production Readiness) or Immediate (if blocking)  
**Objective:** Increase `internal/api/handlers` package coverage from ~23% to >80%.

---

## 1. Problem Analysis

The `internal/api/handlers` package fails coverage audits because the core handlers (`ProjectHandler`, `TaskHandler`) are tightly coupled to concrete Service structs.

**Current State:**
```go
// internal/api/handlers/task_handler.go
type TaskHandler struct {
    // Tightly coupled to concrete struct. Cannot mock without a DB.
    scheduleService *service.ScheduleService 
}
```

**Desired State:**
```go
// internal/api/handlers/task_handler.go
type TaskHandler struct {
    // Decoupled via interface. Easy to mock.
    scheduleService types.ScheduleServiceInterface 
}
```

---

## 2. Refactoring Plan (The "Interface Injection" Pattern)

To enable unit testing, we must refactor the handlers to accept interfaces defined in `pkg/types/interfaces.go` (or `internal/service/interfaces.go`).

### Step 2.1: Define Interfaces

Ensure the following interfaces exist and match the public methods of your services:

**File:** `pkg/types/interfaces.go`
```go
type ProjectService interface {
    CreateProject(ctx context.Context, p *models.Project) error
    GetProject(ctx context.Context, projectID, orgID uuid.UUID) (*models.Project, error)
}

type ScheduleService interface {
    GetTask(ctx context.Context, taskID, projectID, orgID uuid.UUID) (*models.ProjectTask, error)
    UpdateTaskDuration(ctx context.Context, taskID, projectID, orgID uuid.UUID, days float64, reason string) error
    CreateTaskProgress(ctx context.Context, projectID, taskID, userID uuid.UUID, percent int, notes string) error
    UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status types.TaskStatus) error
    RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*models.ScheduleResult, error)
    CreateInspectionRecord(ctx context.Context, projectID, taskID uuid.UUID, inspector, result, notes string, date time.Time) error
}

type VisionService interface {
    VerifyAndPersistTask(ctx context.Context, db QueryExec, taskID, projectID, orgID uuid.UUID, imageURL, desc string) (bool, float64, error)
}
```

### Step 2.2: Update Handlers

Refactor struct definitions and constructors in `internal/api/handlers/`.

**Example (`project_handler.go`):**
```go
type ProjectHandler struct {
    service types.ProjectService // Changed from *service.ProjectService
}

func NewProjectHandler(s types.ProjectService) *ProjectHandler {
    return &ProjectHandler{service: s}
}
```

---

## 3. Test Coverage Specification

Once refactored, implement the following test suites using `httptest` and a mock implementation of the interfaces.

### 3.1 ProjectHandler Tests (`project_handler_test.go`)
| Test Case | Inputs | Mock Behavior | Expected Status |
|-----------|--------|---------------|-----------------|
| CreateProject_Success | Valid JSON, valid Org header | CreateProject returns nil | 201 Created |
| CreateProject_NoHeader | Missing X-Org-ID | (None) | 400 Bad Request |
| CreateProject_OrgMismatch | Body OrgID != Header OrgID | (None) | 403 Forbidden |
| CreateProject_ServiceError | Valid inputs | CreateProject returns error | 500 Internal Server Error |
| GetProject_Success | Valid ID | GetProject returns Project | 200 OK |
| GetProject_NotFound | Valid ID | GetProject returns ErrNotFound | 404 Not Found |

### 3.2 TaskHandler Tests (`task_handler_test.go`)
| Test Case | Inputs | Mock Behavior | Expected Status |
|-----------|--------|---------------|-----------------|
| UpdateTask_Success | ManualOverride=5.0 | UpdateTaskDuration returns nil, Recalculate returns nil | 200 OK |
| UpdateTask_RecalcFail | ManualOverride=5.0 | UpdateTaskDuration returns nil, Recalculate returns error | 500 Internal Server Error |
| RecordProgress_Complete | Percent=100 | CreateTaskProgress returns nil, UpdateTaskStatus returns nil, Recalculate returns nil | 200 OK |
| RecordProgress_Partial | Percent=50 | CreateTaskProgress returns nil (Status/Recalc not called) | 200 OK |
| Inspection_Pass | Result="Passed" | CreateInspectionRecord returns nil, UpdateTaskStatus returns nil, Recalculate returns nil | 200 OK |
| Inspection_Fail | Result="Failed" | CreateInspectionRecord returns nil (Status/Recalc not called) | 200 OK |
| Inspection_InvalidEnum | Result="Maybe" | (None) | 400 Bad Request |

---

## 4. Implementation Strategy

*   **Phase 8, Step 58** is currently allocated for "Go Interface mock testing".
*   **Recommendation:** Do not block Phase 6 (Chat) on this. The chat handler is isolated and covered.
*   **Action Item:** Append this spec to `specs/SPEC_INDEX.md` and execute during Step 58.
