-- Migration: Normalize chat tool calls into relational table
-- See Phase 2 Remediation: Task 1 - Database Normalization
-- This enables high-performance analytics (e.g., "Tool Usage Frequency")

-- Create normalized table for tool usage tracking
CREATE TABLE IF NOT EXISTS chat_tool_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    tool_name TEXT NOT NULL,
    input_payload JSONB,
    output_payload JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for analytics queries (e.g., "Most used tools", "Tool usage trends")
CREATE INDEX IF NOT EXISTS idx_chat_tool_usage_tool_name ON chat_tool_usage(tool_name);

-- Index for message lookups (for retrieving all tools used in a message)
CREATE INDEX IF NOT EXISTS idx_chat_tool_usage_message_id ON chat_tool_usage(message_id);

COMMENT ON TABLE chat_tool_usage IS 'Normalized table for AI tool invocations, extracted from chat_messages.tool_calls JSONB';
COMMENT ON COLUMN chat_tool_usage.tool_name IS 'Function name invoked by the AI model';
COMMENT ON COLUMN chat_tool_usage.input_payload IS 'Arguments passed to the tool (matches ToolCall.Args)';
COMMENT ON COLUMN chat_tool_usage.output_payload IS 'Tool execution response (matches ToolCall.Response)';
