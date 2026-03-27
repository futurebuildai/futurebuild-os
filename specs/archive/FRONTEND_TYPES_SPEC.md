# Frontend Type Definitions Specification (FRONTEND_TYPES_SPEC.md)

This file serves as the "Interface Contract" for frontend development. All TypeScript implementations must adhere to these definitions to ensure parity with the backend agents and the Data Spine.

## 1. Shared Enums (The Contract)
*Source: API_AND_TYPES_SPEC.md, DATA_SPINE_SPEC.md*

```typescript
/**
 * Defines the lifecycle of a ProjectTask.
 * @source API_AND_TYPES_SPEC.md Section 1.1
 */
export enum TaskStatus {
  Pending = 'Pending',
  Ready = 'Ready',
  In_Progress = 'In_Progress',
  Completed = 'Completed',
  Blocked = 'Blocked',
  Delayed = 'Delayed'
}

/**
 * Defines the permissions and identity of a user.
 * @source API_AND_TYPES_SPEC.md Section 1.2
 */
export enum UserRole {
  Admin = 'Admin',
  Builder = 'Builder',
  Client = 'Client',
  Subcontractor = 'Subcontractor'
}

/**
 * Defines the visual components displayed in the Chat Orchestrator.
 * @source API_AND_TYPES_SPEC.md Section 1.3
 */
export enum ArtifactType {
  Invoice = 'Invoice',
  Budget_View = 'Budget_View',
  Gantt_View = 'Gantt_View'
}

/**
 * Defines the type of dependency between tasks in the schedule DAG.
 * @source DATA_SPINE_SPEC.md Section 3.4
 */
export enum DependencyType {
  FS = 'FS', // Finish-to-Start
  SS = 'SS', // Start-to-Start
  FF = 'FF', // Finish-to-Finish
  SF = 'SF'  // Start-to-Finish
}
```

---

## 2. Core Domain Entities (Database Mirrors)
*Source: DATA_SPINE_SPEC.md*

```typescript
/**
 * Represents a high-level project container.
 * @source DATA_SPINE_SPEC.md Section 3.1
 */
export interface Project {
  id: string; // UUID
  org_id: string; // UUID
  name: string;
  address: string;
  permit_issued_date: string; // ISO-8601 Date
  target_end_date: string; // ISO-8601 Date
  gsf: number; // Gross Square Footage
  status: 'Preconstruction' | 'Active' | 'Paused' | 'Completed';
  
  /** 
   * Primary status indicator calculated for the Project Gallery.
   * e.g., "On Schedule", "Delayed 2 Days", "Inspections Pending"
   * @source MASTER_PRD.md Section 7.2
   */
  hero_metric: string;
}

/**
 * Represents a specific task instance within a project.
 * @source DATA_SPINE_SPEC.md Section 3.3
 */
export interface ProjectTask {
  id: string; // UUID
  project_id: string; // UUID
  wbs_code: string; // e.g., "9.3"
  name: string;
  early_start: string; // ISO-8601 Date
  early_finish: string; // ISO-8601 Date
  calculated_duration: number; // DHSM output
  weather_adjusted_duration: number; // Output after SWIM/SAF
  manual_override_days: number | null;
  status: TaskStatus;
  
  /** Result of Gemini Flash validation */
  verified_by_vision: boolean;
  verification_confidence: number; // 0.0 - 1.0

  /** 
   * Calculated: True if Rain_Probability > 40% and task is weather sensitive.
   * @source MASTER_PRD.md Section 2.2
   */
  weather_risk: boolean;

  /** 
   * Calculated: True if a predecessor inspection is not "Passed".
   * @source MASTER_PRD.md Section 2.2
   */
  blocked_by_qc: boolean;
}

/**
 * Individual line item within an invoice.
 */
export interface InvoiceLineItem {
  description: string;
  quantity: number;
  unit_price: number;
  total: number;
}

/**
 * Represents a parsed invoice artifact.
 * @source DATA_SPINE_SPEC.md Section 4.2
 */
export interface Invoice {
  id: string; // UUID
  project_id: string; // UUID
  vendor_name: string;
  amount: number;
  line_items: InvoiceLineItem[];
  detected_wbs_code: string;
  status: 'Pending' | 'Approved' | 'Exported';
}

/**
 * Represents a contact in the global address book.
 * @source DATA_SPINE_SPEC.md Section 2.3
 */
export interface Contact {
  id: string; // UUID
  org_id: string; // UUID
  name: string;
  company: string;
  phone: string;
  email: string;
  global_role: UserRole;
  preferred_contact_method: 'SMS' | 'Email' | 'Both';
}
```

---

## 3. API Response Payloads
*Source: API_AND_TYPES_SPEC.md, AGENT_BEHAVIOR_SPEC.md*

```typescript
/**
 * Payload returned by the Invoice Processor Agent.
 * @source API_AND_TYPES_SPEC.md Section 3.1
 */
export interface InvoiceExtraction {
  vendor: string;
  date: string; // ISO-8601 Date
  invoice_number: string;
  total_amount: number;
  line_items: InvoiceLineItem[];
  suggested_wbs_code: string;
  confidence: number;
}

/**
 * Output from Agent 1 (Daily Focus Agent).
 * @source MASTER_PRD.md Section 2.2
 */
export interface DailyFocusPayload {
  headline: string; // Natural language summary
  priority_tasks: ProjectTask[]; // Top tasks on critical path
  weather_alert: boolean; // Pulsing alert indicator
}

/**
 * Graph structure for the scheduling view.
 * @source API_AND_TYPES_SPEC.md Section 3.2
 */
export interface GanttData {
  project_id: string; // UUID
  calculated_at: string; // ISO-8601 Timestamp
  projected_end_date: string; // ISO-8601 Date
  critical_path: string[]; // Array of WBS codes
  tasks: Array<{
    wbs_code: string;
    name: string;
    status: TaskStatus;
    early_start: string; // ISO-8601 Date
    early_finish: string; // ISO-8601 Date
    duration_days: number;
    is_critical: boolean;
  }>;
}
```

---

## 4. UI State Interfaces (Boutique Builder Profile)
*Source: MASTER_PRD.md, DATA_SPINE_SPEC.md*

```typescript
/**
 * Organizational preferences for the Boutique Builder model.
 * @source MASTER_PRD.md Section 8.3
 */
export interface BuilderProfile {
  speed_preference: 'Relaxed' | 'Standard' | 'Aggressive';
  work_days: 'M-F' | 'M-Sat' | '7 Days';
  self_perform_phases: string[]; // Array of WBS codes (e.g., ["9.1", "9.2"])
}

/**
 * Represents a user alert or daily focus briefing.
 * @source DATA_SPINE_SPEC.md Section 5.2
 */
export interface Notification {
  id: string; // UUID
  user_id: string; // UUID
  type: 'Schedule_Slip' | 'Invoice_Ready' | string;
  priority: number;
  status: 'Unread' | 'Read' | 'Dismissed';
  
  /** 
   * Deep-link action for the notification click.
   * e.g., "open_artifact:Invoice:uuid"
   * @source MASTER_PRD.md Section 2.3
   */
  link_action: string;
}
```
