/**
 * TaskStatus defines the lifecycle of a ProjectTask.
 * See API_AND_TYPES_SPEC.md Section 1.1
 */
export enum TaskStatus {
    Pending = "Pending",
    Ready = "Ready",
    InProgress = "In_Progress",
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
    Client = "Client",
    Subcontractor = "Subcontractor",
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
