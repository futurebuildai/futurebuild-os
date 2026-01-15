-- Rollback: Drop normalized chat tool usage table
-- See Phase 2 Remediation: Task 1 - Database Normalization

DROP INDEX IF EXISTS idx_chat_tool_usage_message_id;
DROP INDEX IF EXISTS idx_chat_tool_usage_tool_name;
DROP TABLE IF EXISTS chat_tool_usage;
