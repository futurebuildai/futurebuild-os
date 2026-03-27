/**
 * Fleet / Equipment Types — Phase 18 ERP
 * See BACKEND_SCOPE.md Section 20.3
 * Matches Go models in internal/models/fleet.go
 */

export type AssetStatus = 'available' | 'in_use' | 'maintenance' | 'retired';
export type AllocationStatus = 'planned' | 'active' | 'completed' | 'cancelled';

export interface FleetAsset {
    id: string;
    org_id: string;
    asset_number: string;
    asset_type: string;
    make?: string;
    model?: string;
    year?: number;
    vin?: string;
    license_plate?: string;
    purchase_date?: string;
    purchase_cost_cents?: number;
    current_value_cents?: number;
    status: AssetStatus;
    location?: string;
    notes?: string;
    created_at: string;
    updated_at: string;
}

export interface EquipmentAllocation {
    id: string;
    asset_id: string;
    project_id: string;
    task_id?: string;
    allocated_from: string;
    allocated_to: string;
    status: AllocationStatus;
    notes?: string;
    created_at: string;
    updated_at: string;
}

export interface MaintenanceLog {
    id: string;
    asset_id: string;
    maintenance_type: string;
    description?: string;
    scheduled_date?: string;
    completed_date?: string;
    cost_cents?: number;
    vendor_name?: string;
    notes?: string;
    created_at: string;
    updated_at: string;
}
