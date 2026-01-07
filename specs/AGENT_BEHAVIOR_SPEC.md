# Agent Behavior Specification (Layer 4: Action Engine)

**Version:** 1.0.0  
**Status:** Initial Specification  
**Context:** CPM-res1.0 Ecosystem (Layers 0-5)

---

## 1. Architecture & Interaction Model

### 1.1 The "Driver" Role
In the FutureBuild ecosystem, Agents in Layer 4 (Action Engine) act as the "Driver" for the Layer 3 (Physics Engine). While Layer 3 handles the deterministic math, Layer 4 agents observe the Data Spine (Layer 2), evaluate complex logical conditions, and trigger calculations or state updates.

### 1.2 State Awareness & Hierarchy
Agents must respect and traverse the standard Entity Hierarchy defined in the Model Spec:
- **Project:** The top-level container for all context and state.
- **Phase:** Logical groupings of work (e.g., Procurement, Foundation).
- **Task:** Individual units of execution with specific WBS codes.

Agents maintain awareness of cross-project constraints but operate primarily at the Project/Task granularity.

---

## 2. Agent 1: The Daily Focus Agent (Superintendent)

### 2.1 Overview
The Daily Focus Agent ensures the on-site team knows exactly what matters most each day, filtering the noise of a 100+ task schedule into actionable priorities.

### 2.2 Execution Logic
- **Trigger:** Cron schedule at `06:00` Local Project Time.
- **Input:** 
  - `Project_Graph`: The current Live Directed Acyclic Graph (DAG) from the CPM Solver.
  - `Weather_Forecast`: 3-day forecast from integrated weather service.

### 2.3 Logic Flow
1.  **Critical Filter:** 
    - Identify tasks where:
      - `Status` IS `In_Progress` OR `Ready`.
      - `Critical_Path` IS `TRUE`.
    - **Limit:** Select Top 3 prioritized by `Early_Start`.
2.  **Constraint Check (Gatekeeper):**
    - **Inspection Gate:** Reference Section 19.1. If a task is `Ready` but its predecessor is an `Inspection` task NOT marked as `Passed`, flag task status as `BLOCKED BY QC`.
    - **Weather Gate:** Reference Section 19.2. 
      - If `Rain_Probability > 40%` AND `Task` is `Weather_Sensitive` (`WBS < 10.0` OR `WBS 13.x`).
      - Flag task as `WEATHER RISK`.
3.  **Output:** 
    - Generate a natural language briefing delivered via SMS/Portal.
    - Example: *"Good morning. We are focused on WBS 9.3 Roof Framing today. Note: WBS 6.0 Trusses are on-site, but we are monitoring a 50% rain risk for this afternoon."*

---

## 3. Agent 2: The Procurement Agent (Supply Chain Guard)

### 3.1 Overview
The Procurement Agent manages "Ghost Predecessor" logic to ensure material availability does not delay physical construction.

### 3.2 Execution Logic
- **Trigger:** Daily System Check OR any `Task.Status` change.
- **Input:** 
  - WBS 6.x Tasks (defined in Spec Section 5.5).
  - External Supply Chain Volatility data.

### 3.3 Logic Flow
1.  **Volatility Check:** Fetch `supply_chain_volatility` (SCV) from Project Variables (Layer 2).
2.  **Buffer Calculation:**
    - `Buffer_Days = 5 * SCV` (As per Spec Section 11.5).
3.  **Order Window Math:**
    - `Latest_Order_Date = Need_Date - Lead_Time - Buffer_Days`.
4.  **State Evaluation:**
    - **Green:** `Today < Latest_Order_Date`.
    - **Yellow:** `Today >= Latest_Order_Date - 7 days`. (Action: Set priority to High, notify user to "Order Soon").
    - **Red (Critical):** `Today > Latest_Order_Date`. (Action: Alert User "Schedule Slip Imminent").
5.  **Hard Link Enforcement:** 
    - If `WBS 6.x Status != Completed` AND `Today > Need_Date`.
    - The agent MUST trigger a CPM Recalculation. This pushes the `Early_Start` of the successor installation task, effectively sliding the entire construction schedule.

---

## 4. Agent 3: The Chat Orchestrator (Interface)

### 4.1 Overview
The Chat Orchestrator serves as the primary interface between human users and the system's deterministic logic.

### 4.2 Execution Logic
- **Trigger:** Inbound message via Webhook (SMS, WhatsApp, Portal Chat).
- **Classification:** Use Intent Classification Matrix (Spec Section 21.1) to route requests.

### 4.3 Tool Capabilities
- **UpdateTaskStatus(id, status):** 
  - Updates the Data Spine.
  - Automatically triggers the Layer 5 Learning Loop to update `Org_Bias`.
- **GetSchedule(view_depth):** 
  - Returns a formatted list of tasks based on the requested depth (Project, Phase, or Task).
- **ExplainDelay(task_id):** 
  - Backtraces the Critical Path within the CPM Solver.
  - Identifies the root cause (e.g., *"The project is delayed by 4 days due to WBS 6.1 Windows having an uncompleted procurement status."*)

### 4.4 Intent: PROCESS_INVOICE
- **Trigger:** User uploads image/PDF with keyword "Invoice" or "Receipt".
- **Workflow:**
    1. **Extraction:** Send file to Layer 1 **Invoice Processor** via `POST /api/v1/documents/analyze`.
    2. **Payload:** Receive `InvoiceExtraction` JSON (as defined in [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md)).
    3. **WBS Prediction:** 
        - Compare `InvoiceExtraction.date` to `Project_Graph` to find active phases.
        - Match `InvoiceExtraction.vendor` Name to `CONTACTS` list.
        - Suggest most likely `WBS_Code`.
    4. **Response:** Return structured summary: *"I read an invoice from [Vendor] for $[Amount]. I've coded this to [Phase Name]. Confirm?"*

### 4.5 Safety Rails
- **Manual Overwrite Protection:** The Agent cannot manually modify `calculated_duration`. 
- **Override Protocol:** Any deviation from calculated values must be explicitly stored as `manual_override` in the `PROJECT_TASK` entity, flagging it for review in the Layer 5 audit trail.

---

## 5. Agent 4: The Subcontractor Liaison (Outbound SMS)

### 5.1 Overview
The Subcontractor Liaison performs proactive coordination with trade partners to confirm mobilization and progress, reducing manual data entry for the Superintendent.

### 5.2 Execution Logic
- **Tools:** Twilio (SMS), Vision Analyzer (Layer 1).
- **Context:** Operates on the `PROJECT_TASK` and `PORTAL_USER` (Subcontractor) entities.

### 5.3 Workflow A: Start Confirmation
- **Trigger:** Cron schedule at `07:00` AM Local Time on `Task.Early_Start_Date`.
- **Target:** All tasks with `Status == Ready` or `Pending`.
- **Action:** 
    1. Call `DirectoryService.GetContactForPhase(Task.Phase_ID)` to find the Contact.
    2. If Contact is **NULL**:
        - Alert Agent 1 (Daily Focus Agent) with "No Contact Assigned" for the Superintendent's briefing.
    3. If Contact is **FOUND**:
        - Call `NotificationService.SendSMS(Contact.ID, message)`: *"Reminder: You are scheduled to start [Task.Name] today. Are you on-site? Reply YES or NO."*
- **Logic:**
    - If reply is **NEGATIVE**:
        - Flag `Task.Status` as `Delayed`.
        - Alert Agent 1 (Daily Focus Agent) to include this in the Superintendent's briefing.

### 5.4 Workflow B: Progress Check (The "Virtual PM")
- **Trigger:** Cron schedule at `16:00` PM Local Time for all tasks with `Status == In_Progress`.
- **Action:** Call `NotificationService.SendSMS(Contact.ID, message)`: *"Job done? Reply YES + Photo or NO + Est Days."*
- **Logging:** Log all outbound SMS and inbound replies to `COMMUNICATION_LOG` for Chat Orchestrator history visibility.
- **Logic:**
    - If reply contains **YES**:
        - Trigger **Validation Protocol** (Section 6).
    - If reply contains **NO**:
        - Parse "Est Days" from response.
        - Update `Task.Remaining_Duration` based on reply.

---

## 6. Validation Protocol

### 6.1 Overview
The Validation Protocol acts as a quality gate for automated status updates, preventing "garbage data" from entering the Data Spine.

### 6.2 Gate 1: Visual Verification
- **Input:** Subcontractor-provided photo from SMS reply.
- **Process:** Call `VisionService.VerifyTask(imageURL, Task.Description)`.
- **Logic:** Gemini Flash verifies if the image content matches the task description (e.g., if task is "Roof Framing", does the photo show completed roof trusses?).
- **Result:**
    - `Verified`: Logical match between photo and task.
    - `Unverified`: Mismatch or low-quality image.

### 6.3 Gate 2: Logic & Inspection Gates
- **Process:** Reference **Section 19.1 (Inspection Gates)**.
- **Logic:** A task cannot be marked `Completed` if a required **Predecessor Inspection** is `Pending` or `Failed`.
- **Reference Table:** Uses the `task_dependencies` table where `is_inspection_gate = TRUE`.

### 6.4 Outcomes
- **Pass (Gate 1 AND Gate 2):**
    - Auto-update `Task.Status` to `Completed`.
    - Record `verified_by_vision = TRUE` and `verification_confidence` in `task_progress`.
- **Fail (Either Gate):**
    - Flag project task as `Requires_Review` in the Superintendent Dashboard.
    - Route to Agent 1 (Daily Focus Agent) for high-priority notification.
    
---

## 7. Agent 5: The Client Reporter (Weekly Drafter)

### 7.1 Overview
The Client Reporter automates the generation of stakeholder progress reports, ensuring consistent communication with clients while maintaining human-level oversight.

### 7.2 Execution Logic
- **Trigger:** Cron schedule at `14:00` Local Time every Friday.
- **Action:** Generate a "Weekly Progress Report" Draft.

### 7.3 Content Definition
- **Accomplished:** SQL query for all `Project_Tasks` where `Status == Completed` AND `Actual_End` within the last 7 days.
- **Up Next:** SQL query for all `Project_Tasks` where `Early_Start` is within the next 7 days.
- **Visuals:** Retrieve the Top 3 photos from **Site Reports (Layer 0)** with the highest `verification_confidence` recorded during the current week.

### 7.4 Human-in-the-Loop Workflow
- **Destination:** Send the generated draft to the **Internal Project Admin** via the Portal Dashboard.
- **Action Required:** The Admin must select either `Approve_Send` (triggers outbound email/portal notification to Client) or `Edit` (opens markdown editor).
- **Audit Trail:** Log approval status and any edits in Layer 5 for future template optimization.
