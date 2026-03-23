/**
 * Schedule Preview Types
 * Types for the instant schedule preview pipeline (onboarding + what-if scenarios).
 * Matches Go types in internal/service/schedule_preview_service.go
 */

import type { GanttData } from './models';

// ============================================================================
// Request Types
// ============================================================================

export interface LongLeadItem {
    name: string;
    brand?: string;
    model?: string;
    category: string; // windows, doors, hvac, appliances, millwork, finishes
    estimated_lead_weeks: number;
    wbs_code?: string;
    notes?: string;
}

export interface CompletedPhaseInput {
    wbs_code: string;   // e.g. "8.0" or "8.x" for whole phase
    actual_end: string;  // YYYY-MM-DD
    status: 'completed' | 'in_progress';
}

export interface SchedulePreviewRequest {
    square_footage: number;
    foundation_type: string;
    start_date: string;    // YYYY-MM-DD
    stories: number;
    address?: string;
    latitude?: number;
    longitude?: number;
    topography?: string;
    soil_conditions?: string;
    bedrooms?: number;
    bathrooms?: number;
    long_lead_items?: LongLeadItem[];
    is_in_progress?: boolean;
    completed_phases?: CompletedPhaseInput[];
    current_date?: string;
}

export interface ScenarioComparisonRequest {
    base: SchedulePreviewRequest;
    alternatives: SchedulePreviewRequest[];
}

// ============================================================================
// Response Types
// ============================================================================

export interface PhasePreview {
    phase_name: string;
    wbs_code: string;
    start_date: string;
    end_date: string;
    duration_days: number;
    is_critical: boolean;
    status: string; // pending, in_progress, completed
}

export interface ProcurementDate {
    item_name: string;
    brand?: string;
    wbs_code: string;
    lead_weeks: number;
    order_by_date: string;
    install_date: string;
    status: string; // overdue, urgent, upcoming, ok
}

export interface WeatherPhaseImpact {
    phase_name: string;
    wbs_code: string;
    extra_days: number;
    reason: string;
}

export interface WeatherImpact {
    affected_phases: WeatherPhaseImpact[];
    total_extra_days: number;
    risk_months: string[];
    summary: string;
}

export interface TradeGap {
    phase_name: string;
    wbs_code: string;
    required_trade: string;
    start_date: string;
    has_contact: boolean;
    contact_name?: string;
}

export interface ScopeChange {
    rule_applied: string;
    tasks_added?: string[];
    tasks_removed?: string[];
    duration_adjustments?: Record<string, number>;
}

export interface SchedulePreviewResponse {
    projected_end: string;
    total_working_days: number;
    remaining_days: number;
    critical_path: string[];
    phase_timeline: PhasePreview[];
    procurement_dates?: ProcurementDate[];
    weather_impact?: WeatherImpact;
    scope_changes: ScopeChange[];
    trade_gaps?: TradeGap[];
    gantt_preview?: GanttData;
    completion_percent: number;
}

export interface ScenarioResult {
    label: string;
    preview: SchedulePreviewResponse;
    delta_days: number;
    delta_cost_cents: number;
    critical_path_diff: string[];
}

export interface ScenarioComparisonResponse {
    scenarios: ScenarioResult[];
}
