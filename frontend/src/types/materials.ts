/**
 * Material & Budget Types — Rosetta Stone parity with Go models.
 * See internal/models/project_material.go
 */

/**
 * Source of a material or budget entry.
 * Matches Go models.MaterialSource.
 */
export type MaterialSource = 'ai' | 'user' | 'default';

/**
 * Full project material record from the database.
 * Matches Go models.ProjectMaterial JSON output.
 */
export interface ProjectMaterial {
    id: string;
    project_id: string;
    wbs_phase_code: string;
    name: string;
    category: string;
    quantity: number;
    unit: string;
    unit_cost_cents: number;
    total_cost_cents: number;
    source: MaterialSource;
    confidence: number;
    brand: string;
    model: string;
    sku: string;
    notes: string;
    created_at: string;
    updated_at: string;
}

/**
 * Lightweight material estimate returned during onboarding.
 * Matches Go models.MaterialEstimate.
 */
export interface MaterialEstimate {
    name: string;
    category: string;
    wbs_phase_code: string;
    quantity: number;
    unit: string;
    unit_cost_cents: number;
    total_cost_cents: number;
    confidence: number;
    source: MaterialSource;
}

/**
 * Per-phase budget estimate breakdown.
 * Matches Go models.PhaseBudgetEstimate.
 */
export interface PhaseBudgetEstimate {
    wbs_phase_code: string;
    phase_name: string;
    estimated_cents: number;
    materials_cents: number;
    labor_cents: number;
    confidence: number;
}

/**
 * Aggregate budget estimate for a project.
 * Matches Go models.BudgetEstimate.
 */
export interface BudgetEstimate {
    total_estimated_cents: number;
    phase_breakdown: PhaseBudgetEstimate[];
    regional_multiplier: number;
    confidence_overall: number;
}

/**
 * Request body for creating a material manually.
 * Matches Go models.CreateMaterialRequest.
 */
export interface CreateMaterialRequest {
    name: string;
    category: string;
    wbs_phase_code: string;
    quantity: number;
    unit: string;
    unit_cost_cents: number;
}

/**
 * Partial update request for a material.
 * Matches Go models.MaterialUpdateRequest.
 */
export interface MaterialUpdateRequest {
    name?: string;
    category?: string;
    wbs_phase_code?: string;
    quantity?: number;
    unit?: string;
    unit_cost_cents?: number;
    brand?: string;
    model?: string;
    sku?: string;
    notes?: string;
}

/**
 * Request body for seeding a project budget from materials.
 * Matches Go models.BudgetSeedRequest.
 */
export interface BudgetSeedRequest {
    materials: MaterialEstimate[];
    regional_multiplier?: number;
}

/**
 * Project budget row (per-phase).
 * Matches Go models.ProjectBudget JSON output.
 */
export interface ProjectBudget {
    id: string;
    project_id: string;
    wbs_phase_id: string;
    estimated_amount_cents: number;
    committed_amount_cents: number;
    actual_amount_cents: number;
    source: MaterialSource;
    confidence: number;
    is_locked: boolean;
    created_at: string;
    updated_at: string;
}
