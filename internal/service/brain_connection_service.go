package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BrainConnection represents the FB-Brain connection for an org.
type BrainConnection struct {
	ID             uuid.UUID        `json:"id"`
	OrgID          uuid.UUID        `json:"org_id"`
	BrainURL       string           `json:"brain_url"`
	IntegrationKey string           `json:"integration_key"`
	Status         string           `json:"status"`
	LastSyncAt     *time.Time       `json:"last_sync_at"`
	Platforms      []BrainPlatform  `json:"platforms"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// BrainPlatform represents a connected platform in the ecosystem.
type BrainPlatform struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// BrainConnectionService manages FB-Brain connection settings.
type BrainConnectionService struct {
	db *pgxpool.Pool
}

// NewBrainConnectionService creates a BrainConnectionService.
func NewBrainConnectionService(db *pgxpool.Pool) *BrainConnectionService {
	return &BrainConnectionService{db: db}
}

// GetConnection returns the brain connection for an org, or defaults if none exist.
func (s *BrainConnectionService) GetConnection(ctx context.Context, orgID uuid.UUID) (*BrainConnection, error) {
	query := `SELECT id, org_id, brain_url, integration_key, status, last_sync_at, platforms, updated_at
	          FROM brain_connections WHERE org_id = $1`

	var conn BrainConnection
	var platformsRaw json.RawMessage
	err := s.db.QueryRow(ctx, query, orgID).Scan(
		&conn.ID, &conn.OrgID, &conn.BrainURL, &conn.IntegrationKey,
		&conn.Status, &conn.LastSyncAt, &platformsRaw, &conn.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return &BrainConnection{
			OrgID:     orgID,
			Status:    "disconnected",
			Platforms: []BrainPlatform{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get brain connection: %w", err)
	}

	if err := json.Unmarshal(platformsRaw, &conn.Platforms); err != nil {
		conn.Platforms = []BrainPlatform{}
	}

	// Mask integration key for display
	if len(conn.IntegrationKey) > 8 {
		conn.IntegrationKey = conn.IntegrationKey[:4] + "..." + conn.IntegrationKey[len(conn.IntegrationKey)-4:]
	}

	return &conn, nil
}

// UpdateConnection upserts the brain connection settings for an org.
func (s *BrainConnectionService) UpdateConnection(ctx context.Context, orgID uuid.UUID, brainURL string) error {
	query := `
		INSERT INTO brain_connections (org_id, brain_url, status, updated_at)
		VALUES ($1, $2, 'connecting', now())
		ON CONFLICT (org_id) DO UPDATE SET
			brain_url = EXCLUDED.brain_url,
			status = 'connecting',
			updated_at = now()
	`
	_, err := s.db.Exec(ctx, query, orgID, brainURL)
	if err != nil {
		return fmt.Errorf("upsert brain connection: %w", err)
	}
	return nil
}

// RegenerateKey generates a new integration key for the brain connection.
func (s *BrainConnectionService) RegenerateKey(ctx context.Context, orgID uuid.UUID) (string, error) {
	key, err := generateIntegrationKey()
	if err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}

	query := `
		INSERT INTO brain_connections (org_id, integration_key, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (org_id) DO UPDATE SET
			integration_key = EXCLUDED.integration_key,
			updated_at = now()
	`
	_, err = s.db.Exec(ctx, query, orgID, key)
	if err != nil {
		return "", fmt.Errorf("upsert integration key: %w", err)
	}
	return key, nil
}

func generateIntegrationKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "fbk_" + hex.EncodeToString(b), nil
}
