-- Migration: Add Communication Models (Revised)
-- Domain 4: Communication & History per DATA_SPINE_SPEC.md

-- Create Custom Enum Types
CREATE TYPE communication_direction_type AS ENUM ('Inbound', 'Outbound');
CREATE TYPE communication_channel_type AS ENUM ('SMS', 'Chat', 'Email');
CREATE TYPE notification_status_type AS ENUM ('Unread', 'Read', 'Dismissed');
CREATE TYPE notification_type_enum AS ENUM ('Schedule_Slip', 'Invoice_Ready', 'Assignment_New', 'Daily_Briefing');

-- 5.1 COMMUNICATION_LOGS
-- Logical Fix: Added user_id for internal User <-> Agent interaction logging
-- Persistence Fix: contact_id and user_id use ON DELETE SET NULL to preserve audit history
CREATE TABLE IF NOT EXISTS communication_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE NOT NULL,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    direction communication_direction_type NOT NULL,
    content TEXT NOT NULL,
    channel communication_channel_type NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 5.2 NOTIFICATIONS
-- Spec Parity: Removed created_at/updated_at as they are not in DATA_SPINE_SPEC.md 5.2
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    type notification_type_enum NOT NULL,
    priority INT DEFAULT 0 NOT NULL,
    status notification_status_type DEFAULT 'Unread' NOT NULL
);

-- Indexes for Domain 4
CREATE INDEX IF NOT EXISTS idx_communication_logs_project_id ON communication_logs(project_id);
CREATE INDEX IF NOT EXISTS idx_communication_logs_contact_id ON communication_logs(contact_id);
CREATE INDEX IF NOT EXISTS idx_communication_logs_user_id ON communication_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
