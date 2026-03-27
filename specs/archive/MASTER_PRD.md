# FutureBuild Master PRD: Generative UI & Action Engine

**Version:** 1.0.0  
**Status:** Baseline for Generative UI (Google Stitch) Development  
**Context:** Integrated Specification Suite (CPM-res1.0)

---

## 1. Global UX Definitions

### 1.1 Navigation & Responsive Behavior
[Reference FRONTEND_SCOPE.md Section 3.2]

The application employs a unified, responsive navigation system that adapts to the user's device while maintaining functional parity.

| Breakpoint | Navigation Pattern | Layout Behavior |
| :--- | :--- | :--- |
| **Desktop (≥1200px)** | Permanent Left Sidebar (`<fb-sidebar>`) | Sidebar + Dashboard (`<fb-dashboard-view>`) + Chat (`<fb-chat-panel>`) rendered in a side-by-side multi-pane view. **Home button defaults to Command Center.** |
| **Tablet (768-1199px)** | Collapsible Sidebar (Hamburger) | Stacked layout: Dashboard on top, Chat below (or via toggle). |
| **Mobile (<768px)** | Bottom Tab Bar (`<fb-mobile-nav>`) | Single-view focus: Bottom nav for "Home", "Chat", "Tasks", "Schedule". Swipe gestures to transition between views. |

### 1.2 Navigation Shortcuts
- **Home:** Always routes to the `<fb-project-gallery>` (Section 7).
- **Daily Focus Shortcut:** One-click access from the Gallery/Home to the Daily Focus View of the highest-priority project (determined by closest `Early_Finish` or `Blocked` status).

### 1.2 Theme: "Construction Professional"
The aesthetic is designed for high-stakes environments where clarity and immediate status recognition are paramount.

- **Primary Motif:** Dark Mode by default.
- **Colors:** 
    - **Backgrounds:** Jet Black (`#000000`), Deep Grey (`#0a0a0a`), Card Surface (`#111111`).
    - **Typography:** High-contrast pure white (`#ffffff`) for readability; muted grey (`#aaaaaa`) for metadata.
    - **Status Colors (Functional):**
        - **Green (`#2e7d32`):** Completed / Passed / Healthy.
        - **Yellow (`#e65100`):** Warning / Near-term / Action Required.
        - **Red (`#c62828`):** Critical / Delayed / Blocked / Over Budget.
- **Typography:** 'Poppins' or 'Inter' (San-serif), comfortable density for data-rich tables (DHSM metrics).

---

## 2. Feature Set A: The Intelligent Dashboard (Home)

### 2.1 User Story
> "As a Superintendent, I want to see my Daily Focus items immediately so I know what to inspect and which crews to push."

### 2.2 UI Requirement: Daily Focus Card (`<fb-status-card type="focus">`)
[Reference AGENT_BEHAVIOR_SPEC.md Section 2.3]

Visualizes the output of **Agent 1 (Daily Focus Agent)**.

- **Layout:** High-priority card at the top of the dashboard.
- **Fields:**
    - **Headline:** Natural language summary (e.g., "3 Critical Tasks for Today").
    - **Priority List:** Top 3 tasks where `Critical_Path == TRUE`.
    - **Constraint Badges:** 
        - `BLOCKED BY QC`: If predecessor inspection is not `Passed`.
        - `WEATHER RISK`: If `Rain_Probability > 40%` and task is `Weather_Sensitive`.
    - **Action:** "Go to Chat" or "Acknowledge" (updates `COMMUNICATION_LOGS`).

### 2.3 UI Requirement: Notifications Stream
[Reference DATA_SPINE_SPEC.md Section 5.2]

Visualizes the `NOTIFICATIONS` table.

- **Layout:** Scrolling vertical list (`<fb-notification-center>`) in the header or sidebar.
- **Visuals:** 
    - **Iconography:** Categorized by `type` (e.g., Bell for "Schedule_Slip", Receipt for "Invoice_Ready").
    - **Interaction:** Click to open deep-link (e.g., open `<fb-artifact-invoice>` in chat).
    - **State:** Unread indicators (small blue dot) cleared upon view.

### 2.4 Acceptance Criteria
- [ ] Dashboard renders at least one "Daily Focus" item if critical tasks exist.
- [ ] "Weather Risk" flag appears only when rain probability exceeds 40% and WBS is < 10.0 or 13.x.
- [ ] Notifications update in real-time as Agent 2 (Procurement) or Agent 4 (Liaison) trigger alerts.

---

## 3. Feature Set B: The Conversational Workspace

### 3.1 User Story
> "As a Builder, I want to drag an invoice into the chat and see it parsed instantly so I can code it to the budget without manual entry."

### 3.2 Interaction Flow: Invoice Ingestion
[Reference FRONTEND_SCOPE.md Section 8.2]

1.  **Idle State:** Standard chat input field (`<fb-chat-input>`).
2.  **Drag State:** A full-screen or chat-restricted **Dropzone Overlay** with the label "Drop Invoice to Parse".
3.  **Processing State:** The input area is replaced by a **Skeleton Loader** with animated pulses; text reads "Extracting line items...".
4.  **Review State:** Render the `<fb-artifact-invoice>` artifact inline within the conversation thread.

### 3.3 Artifact Definition: `<fb-artifact-invoice>`
[Reference API_AND_TYPES_SPEC.md Section 3.1]

A split-screen interface (Desktop) or stacked modal (Mobile) for validation.

- **Left Side:** High-fidelity document viewer (Zoom/Rotate support).
- **Right Side:** Data validation form.
- **Fields:**
    - **Vendor:** (`String`) Autocomplete from `CONTACTS` table.
    - **Invoice #:** (`String`).
    - **Date:** (`ISO-8601`).
    - **Total Amount:** (`DECIMAL`).
    - **WBS Code:** (`String`) Predicted via `suggested_wbs_code`.
    - **Line Items:** Table showing `description`, `quantity`, `unit_price`, `total`.
- **Primary Action:** "Approve & Post to Budget" (Commits record to `INVOICES` and `PROJECT_BUDGETS` tables).

### 3.4 Acceptance Criteria
- [ ] Dropzone triggers on file hover.
- [ ] `<fb-artifact-invoice>` correctly maps all fields from the `InvoiceExtraction` JSON payload.
- [ ] Clicking "Approve" triggers a "Budget Updated" system message in the chat.

---

## 4. Feature Set C: Project Scheduling (Gantt)

### 4.1 User Story
> "As a Manager, I want to toggle 'Critical Path' to see what drives the completion date and identify which supply chain delays are fatal."

### 4.2 UI Requirement: `<fb-artifact-gantt>`
[Reference FRONTEND_SCOPE.md Section 9]

A high-performance interactive timeline visualizing the `PROJECT_TASKS` and `TASK_DEPENDENCIES`.

- **Visual Requirements:**
    - **Baseline:** Early Start (`early_start`) vs. Early Finish (`early_finish`).
    - **Dependencies:** SVGs arrows connecting predecessors to successors.
    - **Critical Path:** Bars on the critical path are highlighted in **Bright Red** (`#c62828`); non-critical tasks are **FutureBuild Primary Blue** (`#667eea`).
    - **Ghost Predecessors:** Specialized markers for WBS 6.x (Procurement). These should appear as **Diamond Icons** on the timeline, linked to their physical successor (e.g., WBS 6.1 Windows linked to WBS 9.6 Exterior Windows).
- **Interactions:**
    - **Toggle:** "Show Critical Path Only" (dims/hides non-critical activities).
    - **Zoom:** Day/Week/Month view levels.
    - **Detail:** Click a bar to show a popover with the `calculated_duration` and `weather_adjusted_duration` breakdown.

### 4.3 Acceptance Criteria
- [ ] Gantt renders the specific `early_start` and `early_finish` dates from the CPM Solver.
- [ ] Critical path tasks are visually distinct from float-available tasks.
- [ ] Procurement tasks (WBS 6.x) are rendered as milestones/diamonds.

---

## 5. Feature Set D: The Field Portal (Mobile Subcontractor)

### 5.1 User Story
> "As a Subcontractor, I want to reply to a text link and upload a photo of my work so I get paid faster without logging into a complex app."

### 5.2 UI Requirement: "Task Focus View"
[Reference AGENT_BEHAVIOR_SPEC.md Section 5.4]

A specialized, minimal-UI view targeted for one-handed mobil use.

- **Layout:** Single-column, large text, high-contrast buttons.
- **Components:**
    - **Task Identity:** Large heading (e.g., "WBS 9.3 Roof Framing").
    - **Large Status Toggle:** A giant pill-shape toggle (`Pending` -> `In_Progress` -> `Complete`).
    - **Rich Media Button:** Persistent "Camera/Upload" button with prominent icon.
    - **Vision Feedback:** If a photo is uploaded, show a "Verifying..." spinner, then a "Verified ✅" green checkmark once `VisionService` returns true.
- **Minimal Text:** Focus on touch targets (min 44x44px); text should be limited to task instructions and status titles.

### 5.3 Acceptance Criteria
- [ ] Mobile view passes WCAG 2.1 AA contrast for status colors.
- [ ] Photo upload triggers `VisionService` verification workflow automatically.
- [ ] Status toggle updates the `PROJECT_TASKS` table in real-time.

---

## 7. Feature Set E: The Command Center (Project Gallery)

### 7.1 User Story
> "As a Builder running 8 projects, I want to see a visual board of my active jobs so I can instantly spot which one needs my attention today."

### 7.2 UI Requirement: Project Gallery (`<fb-project-gallery>`)
A card-based visual hub for multi-project oversight, replacing complex tables with intuitive status cards.

- **Layout:** Responsive grid of large cards.
- **Card Elements:**
    - **Project Name:** Large, bold typography.
    - **Hero Metric:** Primary status indicator (e.g., "On Schedule", "Delayed 2 Days", "Inspections Pending").
    - **"The Red Dot":** A pulsating red indicator (`#c62828`) that appears if:
        - Any task is `Blocked`.
        - An invoice is `Pending` approval.
        - A critical weather risk is detected.
    - **Progress Ring:** A lightweight circular SVG showing % Completion based on WBS phase progress.
- **Usage Indicator Widget:** Visual counter (e.g., "Active: 3/5"). 
    - **Logic:** Green (0-3 projects), Yellow (4), Red (5). 
    - **Constraint:** If at limit (Red), disable the "Primary Action" in the Quick-Add Wizard.
- **Interactions:** Click card to enter the specific Project Dashboard/Chat.

### 7.3 UI Requirement: Quick-Add Wizard
- **Philosophy:** Low administrative barrier for new projects.
- **Flow:**
    1. **Identity:** Name and Address input.
    2. **Upload Plans:** Drag-and-drop area for architectural blueprints/specs.
    3. **Go:** Trigger Agent 2 to initialize the WBS and physics model. (Remove granular config steps).

---

## 8. Feature Set F: Company Settings

### 8.1 User Story
> "As a Player-Coach, I want to set my company's defaults once so the AI knows how I build without me tweaking math formulas."

### 8.2 UI Requirement: Team & Crew
- **Flat Hierarchy:** Simplified user management favoring the Boutique Builder model (Owner + Site Managers).
- **Preferred Vendors:** replaces the "Global Rolodex."
    - **Default Assignment:** "Who is your default plumber?" / "Who is your default framer?".
    - **Behavior:** The AI assumes these vendors for all new projects unless overridden at the phase level.

### 8.3 UI Requirement: Builder Profile (The "Persona" Tuner)
Replaces complex physics tuning with plain-English business preferences.

- **Fields:**
    - **"My Speed":** Segmented control [Relaxed / Standard / Aggressive]. (Internally maps to Physics Multipliers).
    - **"Work Days":** Selection for the standard construction week [M-F / M-Sat / 7 Days].
    - **"Self-Perform":** A checklist of WBS phases the builder handles in-house (e.g., [x] Framing, [ ] Siding). 
        - **Logic:** These phases bypass "Preferred Vendor" logic and assign to the internal crew.

---

## 9. Definition of Done (Generative UI)
The PRD is considered satisfied when:
1.  **State Synchronization:** UI reflects the Data Spine state within <500ms of a backend update.
2.  **Artifact Integrity:** All 3 core ArtifactTypes (`Invoice`, `Budget_View`, `Gantt_View`) are rendered as high-fidelity interactive elements, not static text.
3.  **Role Sensitivity:** The same `/chat` endpoint serves different artifacts/views based on the `UserRole` in the JWT.
