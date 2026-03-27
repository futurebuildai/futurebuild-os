/**
 * Corporate Financials Types — Phase 18 ERP
 * See BACKEND_SCOPE.md Section 20.1
 * Matches Go models in internal/models/corporate_financials.go
 */

export interface CorporateBudget {
    id: string;
    org_id: string;
    fiscal_year: number;
    quarter: number;
    total_estimated_cents: number;
    total_committed_cents: number;
    total_actual_cents: number;
    project_count: number;
    last_rollup_at: string;
    created_at: string;
    updated_at: string;
}

export type GLSyncStatus = 'pending' | 'completed' | 'failed';

export interface GLSyncLog {
    id: string;
    org_id: string;
    sync_type: string;
    status: GLSyncStatus;
    records_synced?: number;
    error_message?: string;
    synced_at?: string;
    created_at: string;
}

export interface ARAgingSnapshot {
    id: string;
    org_id: string;
    snapshot_date: string;
    current_cents: number;
    days_30_cents: number;
    days_60_cents: number;
    days_90_plus_cents: number;
    total_receivable_cents: number;
    created_at: string;
}
