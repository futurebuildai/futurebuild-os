import { TaskStatus, UserRole } from "./enums";

/**
 * Forecast represents weather integration data.
 * See API_AND_TYPES_SPEC.md Section 2.1
 */
export interface Forecast {
    date: string;
    high_temp_c: number;
    low_temp_c: number;
    precipitation_mm: number;
    precipitation_probability: number;
    conditions: string;
}

/**
 * Contact represents a shared contact model.
 * See API_AND_TYPES_SPEC.md Section 4.1
 */
export interface Contact {
    id: string;
    name: string;
    company: string;
    phone: string;
    email: string;
    role: UserRole;
}

/**
 * InvoiceExtraction represents the output of document analysis.
 * See API_AND_TYPES_SPEC.md Section 3.1
 */
export interface InvoiceExtraction {
    vendor: string;
    date: string; // ISO-8601 Date
    invoice_number: string;
    total_amount_cents: string; // P1 Fix: String for 64-bit precision, aligned with Go JSON tag
    line_items: InvoiceExtractionItem[];
    suggested_wbs_code: string;
    confidence: number;
}

/**
 * InvoiceExtractionItem represents a single line item in an invoice.
 */
export interface InvoiceExtractionItem {
    description: string;
    quantity: number;
    unit_price_cents: string; // P1 Fix: String for 64-bit precision, aligned with Go JSON tag
    total_cents: string; // P1 Fix: String for 64-bit precision, aligned with Go JSON tag
}

/**
 * GanttData represents the project schedule for the Gantt view.
 * See API_AND_TYPES_SPEC.md Section 3.2
 */
export interface GanttData {
    project_id: string; // UUID
    calculated_at: string; // ISO-8601 Timestamp
    projected_end_date: string; // ISO-8601 Date
    critical_path: string[];
    tasks: GanttTask[];
    dependencies?: GanttDependency[]; // Step 89: Dependency edges for SVG arrows
}

/**
 * GanttTask represents an individual task in the Gantt data.
 */
export interface GanttTask {
    wbs_code: string;
    name: string;
    status: TaskStatus;
    early_start: string; // ISO-8601 Date
    early_finish: string; // ISO-8601 Date
    duration_days: number;
    is_critical: boolean;
}

/**
 * GanttDependency represents a directed edge between two tasks.
 * See STEP_89_DEPENDENCY_ARROWS.md Section 1.2
 */
export interface GanttDependency {
    from: string; // Predecessor WBS code
    to: string;   // Successor WBS code
}

/**
 * CompletionReport represents the project completion report.
 * Rosetta Stone parity with Go pkg/types.CompletionReport.
 */
export interface CompletionReport {
    id: string;
    project_id: string;
    generated_by?: string;
    schedule_summary: ScheduleSummary;
    budget_summary: BudgetSummary;
    weather_impact_summary?: WeatherImpactSummary;
    procurement_summary?: ProcurementSummary;
    notes?: string;
    created_at: string;
}

/**
 * ScheduleSummary aggregates schedule metrics.
 * Rosetta Stone parity with Go pkg/types.ScheduleSummary.
 */
export interface ScheduleSummary {
    total_tasks: number;
    completed_tasks: number;
    on_time_percent: number;
    total_duration_days: number;
    actual_duration_days: number;
}

/**
 * BudgetSummary aggregates financial metrics.
 * All monetary values in int64 cents.
 * Rosetta Stone parity with Go pkg/types.BudgetSummary.
 */
export interface BudgetSummary {
    estimated_cents: number;
    committed_cents: number;
    actual_cents: number;
    variance_cents: number;
}

/**
 * WeatherImpactSummary aggregates weather delay data.
 * Rosetta Stone parity with Go pkg/types.WeatherImpactSummary.
 */
export interface WeatherImpactSummary {
    total_delay_days: number;
    phases_affected: number;
}

/**
 * ProcurementSummary aggregates procurement metrics.
 * Rosetta Stone parity with Go pkg/types.ProcurementSummary.
 */
export interface ProcurementSummary {
    total_items: number;
    total_spend_cents: number;
    vendor_count: number;
}

/**
 * Thread represents a conversation thread within a project.
 * Rosetta Stone parity with Go pkg/types.Thread.
 */
export interface Thread {
    id: string;
    project_id: string;
    title: string;
    is_general: boolean;
    archived_at?: string;  // ISO-8601 Timestamp, nullable
    created_by?: string;   // UUID string, nullable for system threads
    created_at: string;    // ISO-8601 Timestamp
    updated_at: string;    // ISO-8601 Timestamp
}
