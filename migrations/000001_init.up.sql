-- Migration: Initial Schema
-- Includes Domain 1: Identity & Access (IAM) per DATA_SPINE_SPEC.md

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create Custom Enum Types
CREATE TYPE user_role_type AS ENUM ('Admin', 'Builder', 'Client', 'Subcontractor');
CREATE TYPE contact_role_type AS ENUM ('Client', 'Subcontractor');
CREATE TYPE contact_preference_type AS ENUM ('SMS', 'Email', 'Both');

-- Domain 2: Project Core (The Graph)
CREATE TYPE project_status_type AS ENUM ('Preconstruction', 'Active', 'Paused', 'Completed');
CREATE TYPE task_status_type AS ENUM ('Pending', 'Ready', 'In_Progress', 'Completed', 'Blocked', 'Delayed');
CREATE TYPE dependency_type_enum AS ENUM ('FS', 'SS', 'FF');

-- 2.1 ORGANIZATIONS
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    project_limit INT DEFAULT 5 NOT NULL
);

-- 2.2 USERS
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    role user_role_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2.3 CONTACTS
CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE NOT NULL,
    name VARCHAR(255) NOT NULL,
    company VARCHAR(255),
    phone VARCHAR(50),
    email VARCHAR(255),
    global_role contact_role_type NOT NULL,
    contact_preference contact_preference_type DEFAULT 'Both' NOT NULL
);

-- Standard Indexes
CREATE INDEX IF NOT EXISTS idx_users_org_id ON users(org_id);
CREATE INDEX IF NOT EXISTS idx_contacts_org_id ON contacts(org_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- 3.1 PROJECTS
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE NOT NULL,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    permit_issued_date DATE,
    target_end_date DATE,
    gsf FLOAT DEFAULT 0.0,
    status project_status_type DEFAULT 'Preconstruction' NOT NULL
);

-- 3.2 PROJECT_CONTEXT
CREATE TABLE IF NOT EXISTS project_context (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE NOT NULL,
    supply_chain_volatility INT DEFAULT 1,
    rough_inspection_latency INT DEFAULT 1,
    final_inspection_latency INT DEFAULT 5,
    zip_code VARCHAR(20),
    climate_zone VARCHAR(100)
);

-- WBS Library (Master Templates)
CREATE TABLE IF NOT EXISTS wbs_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50),
    is_default BOOLEAN DEFAULT FALSE,
    entry_point_wbs VARCHAR(20) DEFAULT '5.2',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wbs_phases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID REFERENCES wbs_templates(id) ON DELETE CASCADE NOT NULL,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_weather_sensitive BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS wbs_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phase_id UUID REFERENCES wbs_phases(id) ON DELETE CASCADE NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    base_duration_days FLOAT DEFAULT 0.0,
    responsible_party VARCHAR(255),
    deliverable TEXT,
    notes TEXT,
    is_inspection BOOLEAN DEFAULT FALSE,
    is_milestone BOOLEAN DEFAULT FALSE,
    is_long_lead BOOLEAN DEFAULT FALSE,
    lead_time_weeks_min INT DEFAULT 0,
    lead_time_weeks_max INT DEFAULT 0,
    predecessor_codes TEXT[], -- Array of codes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3.3 PROJECT_TASKS
CREATE TABLE IF NOT EXISTS project_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE NOT NULL,
    wbs_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    early_start DATE,
    early_finish DATE,
    calculated_duration FLOAT DEFAULT 0.0,
    weather_adjusted_duration FLOAT DEFAULT 0.0,
    manual_override_days FLOAT,
    status task_status_type DEFAULT 'Pending' NOT NULL,
    verified_by_vision BOOLEAN DEFAULT FALSE,
    verification_confidence FLOAT DEFAULT 0.0
);

-- 3.4 TASK_DEPENDENCIES
CREATE TABLE IF NOT EXISTS task_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE NOT NULL,
    predecessor_id UUID REFERENCES project_tasks(id) ON DELETE CASCADE NOT NULL,
    successor_id UUID REFERENCES project_tasks(id) ON DELETE CASCADE NOT NULL,
    dependency_type dependency_type_enum DEFAULT 'FS' NOT NULL,
    lag_days INT DEFAULT 0
);

-- 3.5 PROJECT_ASSIGNMENTS
CREATE TABLE IF NOT EXISTS project_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE NOT NULL,
    contact_id UUID REFERENCES contacts(id) ON DELETE CASCADE NOT NULL,
    wbs_phase_id VARCHAR(20) NOT NULL
);

-- Domain 3: Financials (Stubbed)
CREATE TABLE IF NOT EXISTS project_budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE NOT NULL,
    wbs_phase_id VARCHAR(20) NOT NULL,
    estimated_amount DECIMAL(15,2) DEFAULT 0.0,
    committed_amount DECIMAL(15,2) DEFAULT 0.0,
    actual_amount DECIMAL(15,2) DEFAULT 0.0
);

-- Extra Indexes
CREATE INDEX IF NOT EXISTS idx_projects_org_id ON projects(org_id);
CREATE INDEX IF NOT EXISTS idx_project_tasks_project_id ON project_tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_wbs_phases_template_id ON wbs_phases(template_id);
CREATE INDEX IF NOT EXISTS idx_wbs_tasks_phase_id ON wbs_tasks(phase_id);
