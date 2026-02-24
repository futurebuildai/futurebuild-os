/**
 * Artifact Types - View Layer Data Contracts
 * See FRONTEND_SCOPE.md Section 8.3
 *
 * This file defines the specific data structures required by the Artifact Panel components.
 * It adheres to the L7 "Dry" principle by importing core business models where possible
 * and extending them for view-specific needs.
 */

import { InvoiceExtraction, GanttData } from './models';
import { InvoiceStatus } from './enums';

// ============================================================================
// Invoice Artifact — Confidence Scoring (Sprint 3.1)
// ============================================================================

/**
 * Bounding box referencing a region in the source PDF.
 * Coordinates are normalized (0–1) relative to page dimensions.
 * Uses w/h naming to match VisionService output format.
 */
export interface InvoiceFieldBoundingBox {
    page: number;
    x: number;
    y: number;
    w: number;
    h: number;
}

/**
 * Per-field confidence metadata from AI extraction.
 * Enables "diff view" highlighting on invoice artifacts.
 * See Sprint 3.1 — Task 3.1.1.
 */
export interface InvoiceFieldConfidence {
    /** Dot-notation field path, e.g. "vendor", "line_items[0].description" */
    field: string;
    /** Confidence score 0.0–1.0 */
    score: number;
    /** Extraction source */
    source?: 'vision_extraction' | 'manual';
    /** PDF region where this value was extracted from */
    boundingBox?: InvoiceFieldBoundingBox;
}

// ============================================================================
// Invoice Artifact Data
// ============================================================================

/**
 * Invoice view data.
 * Extends core InvoiceExtraction with display-specific fields.
 * See PHASE_13_PRD.md Step 82: Interactive Invoice
 */
export interface InvoiceArtifactData extends InvoiceExtraction {
    /** Invoice ID for API operations (edit, approve, reject) */
    id?: string;
    /** Optional display address for the vendor (not always in extraction) */
    address?: string;
    /** Current invoice status — controls edit/approve affordances */
    status?: InvoiceStatus;
    /** Approval metadata (Step 83) */
    approved_by_id?: string;
    approved_at?: string;
    rejected_by_id?: string;
    rejected_at?: string;
    rejection_reason?: string;
    /** Per-field confidence scores from AI extraction (Sprint 3.1) */
    fieldConfidences?: InvoiceFieldConfidence[];
}

// ============================================================================
// Budget Artifact
// ============================================================================

export interface BudgetCategory {
    name: string;
    budget: number;
    spent: number;
}

/**
 * Budget view data.
 * No core model exists for this yet, so defined here.
 */
export interface BudgetArtifactData {
    totalBudget: number;
    totalSpent: number;
    categories: BudgetCategory[];
}

// ============================================================================
// Gantt Artifact
// ============================================================================

/**
 * Gantt view data.
 * Directly maps to core GanttData.
 */
export type GanttArtifactData = GanttData;

// ============================================================================
// Discriminated Union
// ============================================================================

export type ArtifactData =
    | InvoiceArtifactData
    | BudgetArtifactData
    | GanttArtifactData;
