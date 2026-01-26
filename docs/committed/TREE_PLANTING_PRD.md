# Product Requirement Document: The "Tree Planting" Ceremony (Step 69)

## 1. Problem Statement & Goals

### The "Why"
FutureBuild has evolved into a complex ecosystem with multiple intelligent agents (Procurement, Setup, Liaison) and a central intelligence layer (FutureShade). While we have unit and integration tests for individual components, we lack a definitive proof that the system can **self-heal** and **maintain itself** without human intervention.

### The "Tree Planting" Metaphor
In step 69, we "plant a tree" — a self-sustaining organism. This ceremony represents the moment the code becomes autonomous enough to prune its own dead branches and water itself.

### Goals
1.  **Prove Autonomy**: Demonstrate FutureShade can diagnose a system fault and propose a valid fix.
2.  **Safety First**: Ensure the autonomous agent cannot destroy the system (Compliance/Security).
3.  **Audibility**: Every decision made by the agent must be logged and explainable (The Tribunal).

## 2. User Stories

1.  **As a Developer**, I want to run a `go test` that simulates a broken production environment so that I can see if the agent detects it.
2.  **As a Compliance Officer**, I want the agent's fix to be proposed as a "Pull Request" (or equivalent simulation) so that it doesn't bypass review gates.
3.  **As a System Architect**, I want the failure scenario to be deterministic so the test is not flaky.

## 3. The Test Scenario (The Script)

We will implement a **Go Integration Test** `test/integration/tree_planting_test.go` that follows this act structure:

### Act I: The Sabotage (Setup)
-   **State**: System is healthy.
-   **Action**: The test injects a "Broken Logic" into the `ProjectService`.
    -   *Specific Fault*: A "Chaos Monkey" flag forces `CreateProject` to return a `500 Internal Server Error` with a specific coherent error message (e.g., "DB Connection Pool Exhausted" or "Invalid Region").
    -   *Constraint*: Do NOT actually break the DB; act at the Service layer.

### Act II: The Awakening (Detection)
-   **Action**: The `Tribunal` (FutureShade) runs its scheduled "Health Check" or receives the Error Event.
-   **Observed Behavior**: FutureShade analyzes the error log.
-   **Tribunal Logic (AI)**:
    -   Reads the error: "Invalid Region: 'Mars' is not a supported region."
    -   Consults `docs/` or `codebase` (Knowledge Retrieval).
    -   Determines: "Region configuration is missing 'Mars', but 'Mars' was requested." OR "The code is hardcoded to 'Earth' only."

### Act III: The Diagnosis (Analysis)
-   **Action**: FutureShade formulates a *Root Cause Analysis (RCA)*.
-   **Output**: A JSON object (Tribunal Decision) logged to the Shadow Viewer.
    -   `{ "fault": "ConfigurationDrift", "confidence": 0.95, "proposed_fix": "Update valid_regions list" }`

### Act IV: The Restoration (Remediation)
-   **Action**: FutureShade generates a **Patch**.
    -   *Simulation*: Instead of rewriting disk files (risky for a test), it generates a `git diff` string or a specific "Configuration Override" object.
-   **Verification**: The test applies this override to the in-memory config.
-   **Retry**: The test invokes `CreateProject` again.
-   **Result**: Success (200 OK).

## 4. Technical Requirements

### 4.1. Interfaces
-   **Chaos Injector**: A minimal interface to force specific failures in services.
-   **Tribunal Client**: The existing FutureShade client must be capable of receiving "System Health" prompts.

### 4.2. Autonomy Logic (AI/ML)
-   **Model**: Gemini 1.5 Flash (Fast, reasoning-capable).
-   **Prompt Engineering**:
    -   *System Prompt*: "You are a Site Reliability Engineer. Analyze this error trace and propose a configuration fix."
    -   *Safety*: "Do not propose dropping tables. Do not propose deleting users."

### 4.3. Security & Compliance
-   **Sandboxing**: The test runs in a containerized or isolated environment.
-   **Read-Only Code**: The agent is NOT allowed to modify `.go` files on disk during this test. It modifies *State/Configuration* only. This satisfies the "Least Privilege" security requirement.

## 5. Acceptance Criteria

-   [ ] **Test Passes**: `go test -v ./test/integration/tree_planting_test.go` completes successfully.
-   [ ] **Diagnosis Accuracy**: The Tribunal correctly identifies the injected fault (Chaos Type matches Diagnosis).
-   [ ] **Fix Validity**: The proposed fix, when applied, actually resolves the error.
-   [ ] **Latency**: The entire diagnose-fix loop takes < 30 seconds.
-   [ ] **Logs**: The "Shadow Viewer" contains a full transcript of the Tribunal's reasoning.

## 6. L7 Risk Assessment (Pre-Mortem)

-   **Risk**: The LLM hallucinates a fix that doesn't work.
    -   *Mitigation*: The test asserts that the fix works. If it fails, the test fails. This is acceptable for a "Ceremony" test.
-   **Risk**: Flakiness due to LLM non-determinism.
    -   *Mitigation*: Set `temperature=0` for the Tribunal during this critical mode. Use a standard "Golden Set" of known faults.
-   **Risk**: Infinite Loop (Break -> Fix -> Break).
    -   *Mitigation*: Hard limit of 1 remediation attempt per test run.

## 7. Estimated Complexity
**Small to Medium**. The infrastructure for Tribunal exists (Step 64-68). This is primarily "wiring" the components into a cohesive narrative test.
