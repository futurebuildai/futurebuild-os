-- Rollback: Communication Models (Revised)

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS communication_logs;

DROP TYPE IF EXISTS notification_type_enum;
DROP TYPE IF EXISTS notification_status_type;
DROP TYPE IF EXISTS communication_channel_type;
DROP TYPE IF EXISTS communication_direction_type;
