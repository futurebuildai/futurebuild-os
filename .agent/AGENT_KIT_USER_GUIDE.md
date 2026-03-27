# FutureBuild Agent Kit: User Guide

Welcome to the **FutureBuild Agent Kit**. This guide outlines the operational protocol for the dual-agent architecture powering the transition to the Greenfield ERP/OS.

## 1. The Philosophy: Brain vs. Hands

We operate under a strict **separation of concerns** between two specialized AI engines:

*   **Antigravity (The Brain):** Responsible for high-level reasoning, planning, technical specifications, QA, security audits, and browser-based verification. Antigravity *thinks* so Claude doesn't have to.
*   **Claude Code (The Hands):** Optimized for high-speed file editing, code execution, compilation, and local unit testing. Claude *acts* based on the specs provided by the Brain.

## 2. Available Personas (Skills)

### Antigravity (Brain) - `.agent/skills/`
| Persona | Task | When to Invoke |
| :--- | :--- | :--- |
| **Product Owner** | Roadmaps & PRDs | Define requirements and project direction. |
| **System Architect** | Specs & Data Schemas | Convert requirements into technical blueprints. |
| **Software Tester** | E2E & Browser Tests | Verify UI/UX and complex user flows. |
| **L7 Gatekeeper** | Security Audits | Final zero-trust check before merge. |

### Claude Code (Hands) - `.claude/skills/`
| Persona | Task | When to Invoke |
| :--- | :--- | :--- |
| **Backend Developer** | Go / SQL / Logic | Implementing APIs and database migrations. |
| **Frontend Developer** | Lit / TS / CSS | Building UI components and layouts. |
| **AI Engineer** | AI Logic / CPM / Lit | Implementing document parsing and scheduling logic. |
| **System Integration Engineer** | Connectors / APIs | Building FB-Brain and 3P ERP integrations. |
| **DevOps Engineer** | Infra / CI / CD | Managing deployment pipelines and environments. |

## 3. The Prism Execution Loop

To move a feature from idea to production, follow these 5 steps:

1.  **Plan & Spec (Antigravity):** Invoke `/product` to generate a PRD and technical specs in `specs/`.
2.  **Prompt Generation (Antigravity):** Antigravity outputs a **Terminal Prompt** containing spec paths and constraints.
3.  **Execution (Claude Code):** Copy the prompt and run `claude -p "[Prompt]"` in your terminal.
4.  **Audit & Verification (Antigravity):** Once Claude finishes, run `make test`, `/chome`, and `/L7 Gate` in Antigravity to verify the work.
5.  **Finalize:** If verification passes, Antigravity authorizes the merge and updates the `ROADMAP.md`.

## 4. Commands & Triggers

*   `/product`: Start a new planning phase.
*   `/L7 Gate`: Trigger a security audit.
*   `/software_tester`: Run browser-based E2E tests.
*   `/devteam`: (Legacy/Alias) for technical planning.

## 5. Handoff Protocols

Context is passed between the Brain and the Hands primarily through:
1.  **Specification Files:** Stored in `specs/`, these are the source of truth for Claude Code.
2.  **Terminal Prompts:** Antigravity generates dense instructions that "prime" Claude Code with all necessary context, including absolute file paths to the specs.
3.  **Remediation Prompts:** If an audit fails, Antigravity generates a specific "fix" prompt to be fed back into Claude Code.

---
**Protocol Version:** 1.0.0 (Directive 16)
**Status:** ACTIVE
