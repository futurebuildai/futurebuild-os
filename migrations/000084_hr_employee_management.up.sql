-- Migration: HR & Employee Management
-- See BACKEND_SCOPE.md Section 20.2 — Workforce tracking, labor burden, certifications

-- Employees (extends contacts for internal workforce management)
CREATE TABLE employees (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id            UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id        UUID REFERENCES contacts(id) ON DELETE SET NULL,
    first_name        VARCHAR(100) NOT NULL,
    last_name         VARCHAR(100) NOT NULL,
    employee_number   VARCHAR(50),
    email             VARCHAR(255),
    phone             VARCHAR(20),
    hire_date         DATE,
    termination_date  DATE,
    status            VARCHAR(20) NOT NULL DEFAULT 'active',
    pay_rate_cents    INT,
    pay_type          VARCHAR(20),
    classification    VARCHAR(50),
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, employee_number)
);

CREATE INDEX idx_employees_org ON employees(org_id);
CREATE INDEX idx_employees_contact ON employees(contact_id);
CREATE INDEX idx_employees_status ON employees(status);

CREATE TRIGGER update_employees_modtime
    BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Time logs for labor burden calculation
CREATE TABLE time_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    task_id         UUID REFERENCES project_tasks(id) ON DELETE SET NULL,
    log_date        DATE NOT NULL,
    hours_worked    DECIMAL(5,2) NOT NULL,
    overtime_hours  DECIMAL(5,2) DEFAULT 0,
    notes           TEXT,
    approved        BOOLEAN DEFAULT false,
    approved_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_time_logs_employee ON time_logs(employee_id);
CREATE INDEX idx_time_logs_project ON time_logs(project_id);
CREATE INDEX idx_time_logs_date ON time_logs(log_date);
CREATE INDEX idx_time_logs_task ON time_logs(task_id);

CREATE TRIGGER update_time_logs_modtime
    BEFORE UPDATE ON time_logs
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Certifications for compliance tracking (e.g., OSHA-30)
CREATE TABLE certifications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    cert_type           VARCHAR(100) NOT NULL,
    cert_number         VARCHAR(100),
    issue_date          DATE,
    expiration_date     DATE NOT NULL,
    issuing_authority   VARCHAR(255),
    document_url        TEXT,
    status              VARCHAR(20) DEFAULT 'valid',
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_certifications_employee ON certifications(employee_id);
CREATE INDEX idx_certifications_expiration ON certifications(expiration_date);
CREATE INDEX idx_certifications_status ON certifications(status);

CREATE TRIGGER update_certifications_modtime
    BEFORE UPDATE ON certifications
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Prevailing wage rates by region/classification
CREATE TABLE prevailing_wage_rates (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    region                VARCHAR(100) NOT NULL,
    classification        VARCHAR(100) NOT NULL,
    effective_date        DATE NOT NULL,
    hourly_rate_cents     INT NOT NULL,
    fringe_benefit_cents  INT DEFAULT 0,
    source                VARCHAR(255),
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, region, classification, effective_date)
);

CREATE INDEX idx_prevailing_wage_rates_org ON prevailing_wage_rates(org_id);
CREATE INDEX idx_prevailing_wage_rates_region ON prevailing_wage_rates(region);
