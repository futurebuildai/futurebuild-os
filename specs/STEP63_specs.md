# Technical Specification: Shadow Site & Protocol (Step 63)

| Metadata | Details |
| :--- | :--- |
| **Feature ID** | STAT-63 (Shadow Site) |
| **Author** | DevTeam (Architect, Backend, DevOps, Security) |
| **Status** | **COMPLETED** |
| **Input** | `docs/STEP63_PRD.md` |

---

## 1. Overview
This specification details the implementation of the **Shadow Site** logic. The goal is to enforce a "Dual-Write" protocol where every logic file in the source code has a corresponding markdown observation file in a parallel `shadow/` directory.

## 2. Architecture & Design ("The Architect Setup")

### 2.1 File System Topology
To support the repository structure, we will establish **two** distinct shadow roots. This separation ensures that frontend and backend concerns remain decoupled.

1.  **Frontend Shadow Root**: `frontend/shadow/`
    *   **Source Mirror**: `frontend/src/`
    *   **Mapping Rule**: `frontend/src/path/to/File.ts` -> `frontend/shadow/path/to/File.md`
    *   **Example**: `frontend/src/components/Button.ts` -> `frontend/shadow/components/Button.md`

2.  **Backend Shadow Root**: `backend/shadow/`
    *   **Source Mirrors**: `internal/`, `cmd/`, `pkg/`
    *   **Mapping Rule**: `[source_root]/path/to/File.go` -> `backend/shadow/[source_root]/path/to/File.md`
    *   **Example**: `internal/auth/handler.go` -> `backend/shadow/internal/auth/handler.md`
    *   **Note**: Using `backend/shadow` instead of `server/shadow` because `server` is a compiled binary file in the repo root.

### 2.2 The Shadow Object Model (L7 Template)
Each Shadow File MUST adhere to this exact markdown schema to ensure consistency for future AI agents (FutureShade).

```markdown
# [Filename]

## Intent
*   **High Level:** [Auto-filled: Pending documentation]
*   **Business Value:** [Auto-filled: Pending documentation]

## Responsibility
*   State what this component handles.

## Key Logic
*   Describe flows and state management.

## Dependencies
*   **Upstream:** [Incoming calls]
*   **Downstream:** [Outgoing calls]
```

## 3. Implementation Details ("The Backend Logic")

We will implement the tooling in **Node.js** (TypeScript/JavaScript). This avoids introducing a new runtime dependency for the build scripts.

### 3.1 Script 1: `scripts/shadow/scaffold.js`

**Purpose**: Automatically generate missing shadow files.

**Algorithm:**
1.  **Configuration**: Define a map of `Source Root` -> `Shadow Root`.
    *   `frontend/src` -> `frontend/shadow`
    *   `internal` -> `backend/shadow/internal`
    *   `cmd` -> `backend/shadow/cmd`
    *   `pkg` -> `backend/shadow/pkg`
2.  **Traversal**: Recursively walk each `Source Root`.
3.  **Filtering (Inclusion Criteria)**:
    *   **Include**: Files ending in `.ts`, `.tsx`, `.go`, `.js`, `.py`.
    *   **Exclude**:
        *   Test files: `*.test.ts`, `*.spec.ts`, `*_test.go`
        *   Styles/Assets: `*.css`, `*.scss`, `*.png`, `*.svg`
        *   Directories: `node_modules`, `.git`
4.  **Generation**:
    *   Determine target shadow path.
    *   **Check**: Does file exist?
    *   **Action**: If NO, create file using the L7 Template.
    *   **Logging**: Output "Created: [path]" for each file.

**Constraint**: Use `fs/promises` for async I/O to ensure performance.

### 3.2 Script 2: `scripts/shadow/check.js` (The Enforcer)

**Purpose**: Fail the build if shadow files are missing.

**Algorithm:**
1.  **Traversal**: Re-use the traversal and filtering logic from `scaffold.js`.
2.  **Verification**: For every valid source file found, check if the corresponding shadow file exists.
3.  **Collection**: Maintain a list of `missing_files`.
4.  **Reporting**:
    *   If `missing_files.length > 0`:
        *   Print "ERROR: Shadow Protocol Violation. The following files are missing shadow docs:".
        *   Print the list of missing files.
        *   Print "Run 'npm run shadow:scaffold' to fix this."
        *   **EXIT CODE 1**.
    *   Else:
        *   Print "Shadow Protocol Verified."
        *   **EXIT CODE 0**.

## 4. DevOps & Security ("The Guardians")

### 4.1 CI/CD Integration
The following scripts must be added to `package.json` to integrate with the standard build pipeline.

```json
{
  "scripts": {
    "shadow:scaffold": "node scripts/shadow/scaffold.js",
    "shadow:check": "node scripts/shadow/check.js"
  }
}
```

### 4.2 Security Constraints (Zero Trust)
1.  **No Content Copy**: The `scaffold.js` script must **NEVER** read the content of the source file. It only operates on file *paths*. This guarantees that no hardcoded secrets or PII are accidentally copied into the shadow directory.
2.  **Sanitization**: The generated template uses static placeholder text ("Pending documentation"). It does not attempt to "guess" business logic, preventing hallucinated security claims.

## 5. Verification Checklist

The Software Engineer executing this spec must verify:
- [x] `npm run shadow:scaffold` creates the `shadow/` folder structure.
- [x] `npm run shadow:scaffold` populates it with `.md` files mirroring source logic.
- [x] `npm run shadow:check` passes (Exit 0) after scaffolding.
- [x] `npm run shadow:check` fails (Exit 1) if a shadow file is manually deleted.
- [x] No `.test.ts` or `_test.go` files have shadow copies.
- [x] Scaffold does NOT overwrite existing shadow files.
- [x] CSS/styles and assets directories are excluded.

**Implementation Status**: COMPLETED (2026-01-25)
- Frontend shadow: 48 files in `frontend/shadow/`
- Backend shadow: 82 files in `backend/shadow/`
- Total: 130 shadow documentation files
