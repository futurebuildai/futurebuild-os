/**
 * A2A (Agent-to-Agent) Types — Phase 18 ERP
 * See FRONTEND_SCOPE.md Section 15.1
 * Matches Go models in internal/models/a2a.go
 */

export type AgentConnectionStatus = 'active' | 'paused' | 'error';

export interface A2AExecutionLog {
    id: string;
    org_id: string;
    workflow_id?: string;
    source_system: string;
    target_system: string;
    action_type: string;
    payload?: Record<string, unknown>;
    status: string;
    error_message?: string;
    duration_ms?: number;
    executed_at: string;
    created_at: string;
}

export interface ActiveAgentConnection {
    id: string;
    org_id: string;
    agent_name: string;
    agent_type: string;
    brain_workflow_id?: string;
    status: AgentConnectionStatus;
    last_execution_at?: string;
    execution_count: number;
    error_count: number;
    created_at: string;
    updated_at: string;
}
