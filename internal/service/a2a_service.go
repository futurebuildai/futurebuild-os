package service

import (
	"context"
	"fmt"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// A2AService implements A2AServicer for agent-to-agent logging.
// See FRONTEND_SCOPE.md Section 15.1
type A2AService struct {
	db *pgxpool.Pool
}

func NewA2AService(db *pgxpool.Pool) *A2AService {
	return &A2AService{db: db}
}

func (s *A2AService) LogExecution(ctx context.Context, log *models.A2AExecutionLog) error {
	query := `
		INSERT INTO a2a_execution_logs (org_id, workflow_id, source_system, target_system, action_type, payload, status, error_message, duration_ms, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	return s.db.QueryRow(ctx, query,
		log.OrgID, log.WorkflowID, log.SourceSystem, log.TargetSystem,
		log.ActionType, log.Payload, log.Status, log.ErrorMessage,
		log.DurationMs, log.ExecutedAt,
	).Scan(&log.ID, &log.CreatedAt)
}

func (s *A2AService) GetExecutionLogs(ctx context.Context, orgID uuid.UUID, limit int) ([]models.A2AExecutionLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	query := `
		SELECT id, org_id, workflow_id, source_system, target_system, action_type,
			payload, status, error_message, duration_ms, executed_at, created_at
		FROM a2a_execution_logs
		WHERE org_id = $1
		ORDER BY executed_at DESC
		LIMIT $2`

	rows, err := s.db.Query(ctx, query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.A2AExecutionLog
	for rows.Next() {
		var l models.A2AExecutionLog
		if err := rows.Scan(
			&l.ID, &l.OrgID, &l.WorkflowID, &l.SourceSystem, &l.TargetSystem,
			&l.ActionType, &l.Payload, &l.Status, &l.ErrorMessage,
			&l.DurationMs, &l.ExecutedAt, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (s *A2AService) GetActiveAgents(ctx context.Context, orgID uuid.UUID) ([]models.ActiveAgentConnection, error) {
	query := `
		SELECT id, org_id, agent_name, agent_type, brain_workflow_id, status,
			last_execution_at, execution_count, error_count, created_at, updated_at
		FROM active_agent_connections
		WHERE org_id = $1
		ORDER BY agent_name`

	rows, err := s.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []models.ActiveAgentConnection
	for rows.Next() {
		var a models.ActiveAgentConnection
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.AgentName, &a.AgentType, &a.BrainWorkflowID,
			&a.Status, &a.LastExecutionAt, &a.ExecutionCount, &a.ErrorCount,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}

	return agents, nil
}

func (s *A2AService) PauseAgent(ctx context.Context, agentID, orgID uuid.UUID) error {
	query := `UPDATE active_agent_connections SET status = 'paused' WHERE id = $1 AND org_id = $2`
	tag, err := s.db.Exec(ctx, query, agentID, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("agent connection %s not found", agentID)
	}
	return nil
}

func (s *A2AService) ResumeAgent(ctx context.Context, agentID, orgID uuid.UUID) error {
	query := `UPDATE active_agent_connections SET status = 'active' WHERE id = $1 AND org_id = $2`
	tag, err := s.db.Exec(ctx, query, agentID, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("agent connection %s not found", agentID)
	}
	return nil
}
