package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/colton/futurebuild/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentConfigService manages agent configuration per organization.
type AgentConfigService struct {
	db *pgxpool.Pool
}

// NewAgentConfigService creates an AgentConfigService.
func NewAgentConfigService(db *pgxpool.Pool) *AgentConfigService {
	return &AgentConfigService{db: db}
}

// GetAgentConfig returns the agent settings for an org, or defaults if none exist.
func (s *AgentConfigService) GetAgentConfig(ctx context.Context, orgID uuid.UUID) (*config.AgentSettings, error) {
	query := `SELECT config FROM agent_configs WHERE org_id = $1`
	var raw json.RawMessage
	err := s.db.QueryRow(ctx, query, orgID).Scan(&raw)
	if err == pgx.ErrNoRows {
		defaults := config.DefaultAgentSettings()
		return &defaults, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get agent config: %w", err)
	}

	var settings config.AgentSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal agent config: %w", err)
	}
	return &settings, nil
}

// UpdateAgentConfig upserts the agent settings for an org.
func (s *AgentConfigService) UpdateAgentConfig(ctx context.Context, orgID, userID uuid.UUID, settings *config.AgentSettings) error {
	raw, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal agent config: %w", err)
	}

	query := `
		INSERT INTO agent_configs (org_id, config, updated_by, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (org_id) DO UPDATE SET
			config = EXCLUDED.config,
			updated_by = EXCLUDED.updated_by,
			updated_at = now()
	`
	_, err = s.db.Exec(ctx, query, orgID, raw, userID)
	if err != nil {
		return fmt.Errorf("upsert agent config: %w", err)
	}
	return nil
}
