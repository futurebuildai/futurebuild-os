# FutureShade: The Intelligence Plane

**Motto:** *"One generation plants the trees, and another gets the shade."*
**Focus:** Multi-Model Consensus, Architectural Legacy, Automated Governance.

---

## 1. The Philosophy
FutureBuild (the product) is the "Tree." FutureShade (this system) is the "Gardener."
It is a shadow application that monitors the health, growth, and architectural integrity of the main codebase. It does not just "log" events; it *judges* them.

## 2. The "Council of Models" Architecture

### Phase 1: The Intake (The Seed)
* **Input:** User Query ("Fix this bug") OR GitHub Event (PR Open).
* **Processor:** Gemini Flash (Router).
* **Output:** `StandardizedIntent` JSON.

### Phase 2: The Tribunal (The Roots)
Three distinct LLMs process the `StandardizedIntent` in isolation:
1.  **The Architect (Claude):** Focuses on pattern correctness and security.
2.  **The Historian (Gemini Pro):** Uses full repo context to ensure consistency with `DATA_SPINE_SPEC.md`.
3.  **The Logician (Open Source/DeepSeek):** optimizing for algorithmic efficiency and raw logic.

### Phase 3: The Cross-Pollination (The Shade)
* **Reflexion Loop:** The outputs of Phase 2 are swapped.
* *Prompt to Gemini:* "Claude suggests X. What is the risk?"
* *Prompt to Claude:* "DeepSeek suggests Y. Is this secure?"
* **Synthesis:** The "Best of Breed" solution is selected.

### Phase 4: Execution (The Fruit)
* **Output:** A `RemediationPlan` (Spec).
* **Agent:** Antigravity (FutureBuild's Agent System) executes the plan.

## 3. Features

### The Review Log (Sprint Management)
* Every Tribunal session is logged as a "Ticket."
* Status: `Proposed` -> `Debating` -> `Consensus` -> `Implemented`.
* This replaces traditional Jira/Linear workflows with an AI-native "Decision Stream."

### The Shadow Viewer
* A visual interface (Lit + TypeScript) to view the "Digital Twin" of the codebase.
* Visualizes the `shadowdocs/` alongside the live code.