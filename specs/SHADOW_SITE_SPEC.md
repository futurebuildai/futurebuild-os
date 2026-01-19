# Spec: The Shadow Site (Architectural Digital Twin)

**Version:** 1.0.0
**Status:** DRAFT (Target: Phase 8)
**Focus:** Architectural Observability, Knowledge Preservation, AI Context

---

## 1. The Vision
The "Shadow Site" is an internal documentation platform that mirrors the production codebase file-for-file. It provides a "Zoomed Out" natural language explanation of the system, allowing non-engineers and AI agents to navigate, audit, and understand the product without reading a single line of code.

**The Metaphor:**
* **The Codebase** is the "Engine Room" (Noisy, complex, dangerous).
* **The Shadow Site** is the "Control Room" (Clean, schematic, labeled).

---

## 2. The Shadow Protocol (The Rules)

To maintain the integrity of the Shadow Site, we enforce the **"Dual-Write" Protocol**:

1.  **Parity Rule:** For every structural element in `src/` (Directory or Key Component), there MUST be a corresponding entry in `shadow/`.
    * `src/components/billing/InvoiceCard.ts` -> `shadow/components/billing/InvoiceCard.md`
2.  **The "Why" Filter:** Shadow files must NEVER contain code snippets. They must strictly contain:
    * **Intent:** Why does this exist?
    * **Responsibility:** What business problem does it solve?
    * **Dependencies:** Who does it talk to? (e.g., "Fetches data from the Invoice Service").
3.  **CI Enforcement:** The build pipeline in Phase 8 will reject any Pull Request that modifies a `src` file without a corresponding update timestamp or hash change in its `shadow` counterpart.

---

## 3. Directory Structure (The Mirror)

The `shadow/` directory lives at the root of the frontend (and eventually backend) and mirrors the source exactly. For example:

```text
frontend/
├── src/
│   ├── components/
│   │   ├── layout/
│   │   │   └── Sidebar.ts  <-- The Code (400 lines of Lit/TS)
│   │   └── billing/
│   │       └── Invoice.ts
│   └── store/
│       └── auth.ts
└── shadow/
    ├── components/
    │   ├── layout/
    │   │   └── Sidebar.md  <-- The Shadow (The "Story")
    │   └── billing/
    │       └── Invoice.md
    └── store/
        └── auth.md