/**
 * Employee / HR Types — Phase 18 ERP
 * See BACKEND_SCOPE.md Section 20.2
 * Matches Go models in internal/models/employee.go
 */

export type EmployeeStatus = 'active' | 'on_leave' | 'terminated';
export type PayType = 'hourly' | 'salary';
export type CertStatus = 'valid' | 'expiring_soon' | 'expired';

export interface Employee {
    id: string;
    org_id: string;
    contact_id?: string;
    first_name: string;
    last_name: string;
    employee_number?: string;
    email?: string;
    phone?: string;
    hire_date?: string;
    termination_date?: string;
    status: EmployeeStatus;
    pay_rate_cents?: number;
    pay_type?: PayType;
    classification?: string;
    created_at: string;
    updated_at: string;
}

export interface TimeLog {
    id: string;
    employee_id: string;
    project_id?: string;
    task_id?: string;
    log_date: string;
    hours_worked: number;
    overtime_hours: number;
    notes?: string;
    approved: boolean;
    approved_by?: string;
    approved_at?: string;
    created_at: string;
    updated_at: string;
}

export interface Certification {
    id: string;
    employee_id: string;
    cert_type: string;
    cert_number?: string;
    issue_date?: string;
    expiration_date: string;
    issuing_authority?: string;
    document_url?: string;
    status: CertStatus;
    created_at: string;
    updated_at: string;
}

export interface PrevailingWageRate {
    id: string;
    org_id: string;
    region: string;
    classification: string;
    effective_date: string;
    hourly_rate_cents: number;
    fringe_benefit_cents: number;
    source?: string;
    created_at: string;
    updated_at: string;
}
