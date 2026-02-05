/**
 * TaskStatus defines the lifecycle of a ProjectTask.
 * See API_AND_TYPES_SPEC.md Section 1.1
 */
export enum TaskStatus {
    Pending = "Pending",
    Ready = "Ready",
    InProgress = "In_Progress",
    InspectionPending = "Inspection_Pending",
    Completed = "Completed",
    Blocked = "Blocked",
    Delayed = "Delayed",
}

/**
 * UserRole defines the permissions and identity of a PortalUser or User.
 * See API_AND_TYPES_SPEC.md Section 1.2
 */
export enum UserRole {
    Admin = "Admin",
    Builder = "Builder",
    PM = "PM",
    Viewer = "Viewer",
    Client = "Client",
    Subcontractor = "Subcontractor",
}

/**
 * InvoiceStatus defines the lifecycle of an Invoice.
 * See PHASE_13_PRD.md Section 3.1 and models/financial.go
 */
export enum InvoiceStatus {
    Draft = "Draft",
    Pending = "Pending",
    Approved = "Approved",
    Rejected = "Rejected",
    Exported = "Exported",
}

/**
 * ArtifactType defines the visual components displayed in the Chat Orchestrator.
 * See API_AND_TYPES_SPEC.md Section 1.3
 */
export enum ArtifactType {
    Invoice = "Invoice",
    BudgetView = "Budget_View",
    GanttView = "Gantt_View",
}

// ============================================================================
// Routing Types
// ============================================================================

/**
 * ViewId defines the strict set of valid view identifiers.
 * See PRODUCTION_PLAN.md Step 51.4 - L7 Amendment: Type Safety
 *
 * Used by:
 * - store.ui.activeView$ signal
 * - FBAppShell router switch
 * - Navigation rail click handlers
 */
export type ViewId =
    | 'dashboard'
    | 'projects'
    | 'chat'
    | 'schedule'
    | 'directory'
    | 'login';

/**
 * Array of valid ViewIds for runtime validation (e.g., URL parsing).
 */
export const VIEW_IDS: readonly ViewId[] = [
    'dashboard',
    'projects',
    'chat',
    'schedule',
    'directory',
    'login',
] as const;

/**
 * Type guard to check if a string is a valid ViewId.
 * Used for URL parsing and defensive routing.
 */
export function isViewId(value: string): value is ViewId {
    return VIEW_IDS.includes(value as ViewId);
}
