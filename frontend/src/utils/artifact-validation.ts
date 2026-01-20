/**
 * Artifact Data Validation - Type Guards & Validation Utilities
 * See FRONTEND_SCOPE.md Section 8.3
 * Step 58.5: Fortress Hardening - Flag 1 (White Screen of Death)
 *
 * Pre-render validation prevents component crashes from corrupt/missing data.
 */

import { normalizeArtifactType } from './artifact-helpers';
import type { InvoiceArtifactData, BudgetArtifactData, GanttArtifactData } from '../types/artifacts';

// ============================================================================
// Validation Result
// ============================================================================

export interface ValidationResult {
    valid: boolean;
    error?: string;
}

const VALID: ValidationResult = { valid: true };

function invalid(error: string): ValidationResult {
    return { valid: false, error };
}

// ============================================================================
// Type Guards
// ============================================================================

function isObject(val: unknown): val is Record<string, unknown> {
    return typeof val === 'object' && val !== null && !Array.isArray(val);
}

function isArray(val: unknown): val is unknown[] {
    return Array.isArray(val);
}

function isString(val: unknown): val is string {
    return typeof val === 'string';
}

function isNumber(val: unknown): val is number {
    return typeof val === 'number' && !Number.isNaN(val);
}

// ============================================================================
// Invoice Validation
// ============================================================================

function validateInvoice(data: unknown): ValidationResult {
    if (!isObject(data)) return invalid('Invoice data must be an object');

    const invoice = data as Partial<InvoiceArtifactData>;

    if (!isString(invoice.invoice_number)) return invalid('Missing invoice_number');
    if (!isString(invoice.vendor)) return invalid('Missing vendor');
    if (!isString(invoice.date)) return invalid('Missing date');
    // total_amount_cents is a string for 64-bit precision
    if (!isString(invoice.total_amount_cents)) return invalid('Missing or invalid total_amount_cents');
    if (!isArray(invoice.line_items)) return invalid('Missing line_items array');

    // Validate each line item
    for (let i = 0; i < invoice.line_items.length; i++) {
        const idx = String(i);
        const item = invoice.line_items[i];
        if (!isObject(item)) return invalid(`line_items[${idx}] must be an object`);
        if (!isString((item as Record<string, unknown>).description)) return invalid(`line_items[${idx}] missing description`);
        if (!isNumber((item as Record<string, unknown>).quantity)) return invalid(`line_items[${idx}] missing quantity`);
        // unit_price_cents and total_cents are strings for 64-bit precision
        if (!isString((item as Record<string, unknown>).unit_price_cents)) return invalid(`line_items[${idx}] missing unit_price_cents`);
        if (!isString((item as Record<string, unknown>).total_cents)) return invalid(`line_items[${idx}] missing total_cents`);
    }

    return VALID;
}

// ============================================================================
// Budget Validation
// ============================================================================

function validateBudget(data: unknown): ValidationResult {
    if (!isObject(data)) return invalid('Budget data must be an object');

    const budget = data as Partial<BudgetArtifactData>;

    if (!isNumber(budget.totalBudget)) return invalid('Missing or invalid totalBudget');
    if (!isNumber(budget.totalSpent)) return invalid('Missing or invalid totalSpent');
    if (!isArray(budget.categories)) return invalid('Missing categories array');

    // Validate each category
    for (let i = 0; i < budget.categories.length; i++) {
        const idx = String(i);
        const cat = budget.categories[i];
        if (!isObject(cat)) return invalid(`categories[${idx}] must be an object`);
        if (!isString((cat as Record<string, unknown>).name)) return invalid(`categories[${idx}] missing name`);
        if (!isNumber((cat as Record<string, unknown>).budget)) return invalid(`categories[${idx}] missing budget`);
        if (!isNumber((cat as Record<string, unknown>).spent)) return invalid(`categories[${idx}] missing spent`);
    }

    return VALID;
}

// ============================================================================
// Gantt Validation
// ============================================================================

function validateGantt(data: unknown): ValidationResult {
    if (!isObject(data)) return invalid('Gantt data must be an object');

    const gantt = data as Partial<GanttArtifactData>;

    if (!isString(gantt.project_id)) return invalid('Missing project_id');
    if (!isString(gantt.calculated_at)) return invalid('Missing calculated_at');
    if (!isString(gantt.projected_end_date)) return invalid('Missing projected_end_date');
    if (!isArray(gantt.tasks)) return invalid('Missing tasks array');

    // Validate each task
    for (let i = 0; i < gantt.tasks.length; i++) {
        const idx = String(i);
        const task = gantt.tasks[i];
        if (!isObject(task)) return invalid(`tasks[${idx}] must be an object`);
        if (!isString((task as Record<string, unknown>).name)) return invalid(`tasks[${idx}] missing name`);
        if (!isString((task as Record<string, unknown>).early_start)) return invalid(`tasks[${idx}] missing early_start`);
        if (!isString((task as Record<string, unknown>).early_finish)) return invalid(`tasks[${idx}] missing early_finish`);
        if (!isNumber((task as Record<string, unknown>).duration_days)) return invalid(`tasks[${idx}] missing duration_days`);
    }

    return VALID;
}

// ============================================================================
// Unified Validator
// ============================================================================

/**
 * Validates artifact data before rendering.
 * Returns validation result with error message on failure.
 *
 * @param type - The artifact type (will be normalized)
 * @param data - The data to validate
 * @returns ValidationResult with valid flag and optional error
 */
export function validateArtifactData(type: string, data: unknown): ValidationResult {
    if (data === null || data === undefined) {
        return invalid('No data provided');
    }

    const normalizedType = normalizeArtifactType(type);

    switch (normalizedType) {
        case 'invoice':
            return validateInvoice(data);
        case 'budget':
            return validateBudget(data);
        case 'gantt':
            return validateGantt(data);
        default:
            return invalid(`Unknown artifact type: ${type}`);
    }
}
