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
    total_amount: string; // P1 Fix: String for 64-bit precision
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
    unit_price: string; // P1 Fix: String for 64-bit precision
    total: string; // P1 Fix: String for 64-bit precision
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
