# Technical Specifications: The "Tree Planting" Ceremony

## 1. Overview
This specification details the implementation of the "Tree Planting" integration test. This test validates the "Autonomy" capability of the FutureShade intelligence layer by injecting a controlled failure, observing the system's reaction (Tribunal), and verifying that a proposed fix restores functionality.

## 2. Architecture

### 2.1. System Components
1.  **ProjectService (Target)**: The service where we will inject failure.
2.  **ChaosInjector (New Component)**: A mechanism to intercept service calls and simulation errors.
3.  **Tribunal (FutureShade)**: The existing intelligence client that will be triggered to diagnose the issue.
4.  **TreePlantingTest (Harness)**: The integration test driver.

### 2.2. Interface Design
We will introduce a `ChaosInjector` interface that wraps the standard service method.

```go
// internal/chaos/injector.go

type FaultType string

const (
    FaultNone          FaultType = "None"
    FaultServiceError  FaultType = "ServiceError" // 500
    FaultConfigDrift   FaultType = "ConfigDrift"  // "Invalid Region"
)

type ChaosConfig struct {
    TargetMethod string    // e.g. "CreateProject"
    ActiveFault  FaultType
    ErrorMessage string
}

type Injector interface {
    ShouldFail(method string) (bool, error)
    RegisterFault(config ChaosConfig)
    ClearFaults()
}
```

### 2.3. Tribunal Interaction Flow
1.  **Fail**: `ProjectService.CreateProject` calls `injector.ShouldFail()`. It returns `true` and `error("Region 'Mars' not allowed")`.
2.  **Detect**: Test harness captures this error.
3.  **Prompt**: Test harness calls `Tribunal.Diagnose(errorString, context)`.
4.  **Analyze**: Tribunal (AI) returns a `TribunalDecision`.
5.  **Act**: Test harness parses `TribunalDecision.ProposedFix` and applies it to `ChaosConfig` (e.g., updates a "ValidRegions" map in memory).

## 3. Data Model

### 3.1. Tribunal Decision Schema
The AI must return strictly structured JSON.

```json
{
  "fault_diagnosis": "ConfigurationDrift",
  "confidence_score": 0.95,
  "reasoning": "The error indicates 'Mars' is invalid, but the request was for 'Mars'. Codebase analysis shows valid regions are [Earth, Moon].",
  "proposed_action": {
    "type": "UpdateConfiguration",
    "key": "ValidRegions",
    "value": ["Earth", "Moon", "Mars"]
  }
}
```

### 3.2. Go Structs
```go
type TribunalDecision struct {
    FaultDiagnosis string          `json:"fault_diagnosis"`
    Confidence     float64         `json:"confidence_score"`
    Reasoning      string          `json:"reasoning"`
    ProposedAction ReformAction    `json:"proposed_action"`
}

type ReformAction struct {
    Type  string      `json:"type"` // e.g., "UpdateConfiguration"
    Key   string      `json:"key"`
    Value interface{} `json:"value"`
}
```

## 4. Security Considerations

### 4.1. Sandboxing
-   The `ReformAction` is **not** applied to the actual `config.yaml` on disk.
-   It is applied to an **In-Memory Configuration Override** layer within the test process.
-   This ensures that even if the AI proposes "Delete All Users", the test harness interprets that action as invalid or simply doesn't hook it up to a deletion logic.

### 4.2. Least Privilege
-   The Tribunal Client used in this test will have a restricted scope (only access to read logs and suggest config changes).

## 5. Testing Strategy

### 5.1. Integration Test (`test/integration/tree_planting_test.go`)
This is the core deliverable.

**Test Steps:**
1.  **Setup**: Initialize `ProjectService` with a `MemoryChaosInjector`.
2.  **Baseline**: Call `CreateProject("Earth")` -> Expect Success.
3.  **Inject**: `injector.RegisterFault(FaultConfigDrift, "Mars")`.
4.  **Verify Failure**: Call `CreateProject("Mars")` -> Expect Error "Region 'Mars' Invalid".
5.  **Tribunal Loop**:
    -   Send Error to Tribunal.
    -   Assert `Decision.FaultDiagnosis == "ConfigurationDrift"`.
    -   Assert `Decision.ProposedAction` adds "Mars".
6.  **Remediate**: Apply `ProposedAction` to `ProjectService.ValidRegions`.
7.  **Verify Fix**: Call `CreateProject("Mars")` -> Expect Success.

### 5.2. Negative Testing
-   Verify that if Tribunal returns low confidence (< 0.5), no action is taken.
-   Verify that if Tribunal proposes an unknown action type, the test fails gracefully.

## 6. Implementation Notes

-   **Files**:
    -   `internal/chaos/`: New package for injection logic.
    -   `test/integration/tree_planting_test.go`: The main test file.
    -   `pkg/types/tribunal.go`: Update with new Decision logic if needed.
-   **AI Model**: Use `Gemini 1.5 Pro` or `Flash` via the existing `pkg/ai` client. Ensure `Temperature` is `0.0` for this call to ensure determinism.

## 7. L7 Pre-Mortem Check
-   **Risk**: The prompt to the AI is too vague.
    -   *Fix*: The `SystemPrompt` must be explicitly included in the code: *"You are an auto-remediation agent. You MUST output JSON matching this schema..."*
-   **Risk**: Test dependencies (external AI).
    -   *Fix*: If the AI API is down, skipping the test is acceptable (flaky infra), but failing is better for a "Ceremony". We will treat it as a hard dependency.
