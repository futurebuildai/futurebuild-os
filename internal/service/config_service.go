package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/models"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// configReturnColumns defines the SELECT columns for business_config queries.
// Centralized to prevent schema drift between GET and UPSERT.
const configReturnColumns = `id, org_id, speed_multiplier, work_days, created_at, updated_at`

// scanConfig scans a business_config row into a BusinessConfig struct.
func scanConfig(scanner interface{ Scan(...any) error }) (models.BusinessConfig, error) {
	var c models.BusinessConfig
	var workDaysJSON []byte
	err := scanner.Scan(
		&c.ID,
		&c.OrgID,
		&c.SpeedMultiplier,
		&workDaysJSON,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(workDaysJSON, &c.WorkDays); err != nil {
		return c, fmt.Errorf("unmarshal work_days: %w", err)
	}
	return c, nil
}

// ConfigService implements ConfigServicer with pgx.
// See STEP_87_CONFIG_PERSISTENCE.md Section 2
type ConfigService struct {
	db *pgxpool.Pool
}

// NewConfigService creates a new ConfigService.
func NewConfigService(db *pgxpool.Pool) *ConfigService {
	return &ConfigService{db: db}
}

// GetConfig retrieves the business config for an organization.
// Returns default config if none exists (lazy initialization).
func (s *ConfigService) GetConfig(ctx context.Context, orgID uuid.UUID) (*models.BusinessConfig, error) {
	query := fmt.Sprintf(`SELECT %s FROM business_config WHERE org_id = $1`, configReturnColumns)

	cfg, err := scanConfig(s.db.QueryRow(ctx, query, orgID))
	if err != nil {
		// L-6: Use errors.Is for robust wrapped-error comparison
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Info("config: no config found, returning defaults", "org_id", orgID)
			return models.DefaultBusinessConfig(orgID), nil
		}
		return nil, fmt.Errorf("get config: %w", err)
	}
	return &cfg, nil
}

// UpdateConfig upserts the business config for an organization.
// Uses INSERT ... ON CONFLICT to atomically create or update.
func (s *ConfigService) UpdateConfig(ctx context.Context, orgID uuid.UUID, speedMultiplier float64, workDays []int) (*models.BusinessConfig, error) {
	workDaysJSON, err := json.Marshal(workDays)
	if err != nil {
		return nil, fmt.Errorf("marshal work_days: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO business_config (org_id, speed_multiplier, work_days)
		VALUES ($1, $2, $3::jsonb)
		ON CONFLICT (org_id) DO UPDATE
		SET speed_multiplier = EXCLUDED.speed_multiplier,
		    work_days = EXCLUDED.work_days,
		    updated_at = NOW()
		RETURNING %s
	`, configReturnColumns)

	cfg, err := scanConfig(s.db.QueryRow(ctx, query, orgID, speedMultiplier, workDaysJSON))
	if err != nil {
		return nil, fmt.Errorf("upsert config: %w", err)
	}

	slog.Info("config: physics settings updated",
		"org_id", orgID,
		"speed_multiplier", speedMultiplier,
		"work_days", workDays,
	)
	return &cfg, nil
}
