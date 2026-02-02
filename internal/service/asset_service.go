package service

import (
	"context"
	"fmt"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Centralized column list for project_assets queries.
// M1 Fix: Single source of truth prevents column mismatch drift.
const assetReturnColumns = `pa.id, pa.project_id, pa.task_id, pa.uploaded_by,
		       pa.file_name, pa.file_url, pa.mime_type, pa.file_size_bytes,
		       pa.analysis_status, pa.analysis_result,
		       pa.created_at, pa.updated_at`

// MaxAssetsPerProject is the hard limit on assets returned in a list query.
// H1 Fix: Prevents unbounded result sets (DoS protection).
const MaxAssetsPerProject = 200

// scanAsset scans a row into a ProjectAsset struct using the centralized column order.
func scanAsset(scanner interface{ Scan(...any) error }) (models.ProjectAsset, error) {
	var a models.ProjectAsset
	err := scanner.Scan(
		&a.ID, &a.ProjectID, &a.TaskID, &a.UploadedBy,
		&a.FileName, &a.FileURL, &a.MimeType, &a.FileSizeBytes,
		&a.AnalysisStatus, &a.AnalysisResult,
		&a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

// AssetService manages project assets (uploaded photos).
// See STEP_84_FIELD_FEEDBACK.md Section 2
type AssetService struct {
	db *pgxpool.Pool
}

// NewAssetService creates a new AssetService.
func NewAssetService(db *pgxpool.Pool) *AssetService {
	return &AssetService{db: db}
}

// GetAssetStatus retrieves the analysis status of a project asset with multi-tenancy guard.
// See STEP_84_FIELD_FEEDBACK.md Section 2.1
func (s *AssetService) GetAssetStatus(ctx context.Context, assetID uuid.UUID, orgID uuid.UUID) (*models.ProjectAsset, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM project_assets pa
		JOIN projects p ON pa.project_id = p.id
		WHERE pa.id = $1 AND p.org_id = $2
	`, assetReturnColumns)

	asset, err := scanAsset(s.db.QueryRow(ctx, query, assetID, orgID))
	if err != nil {
		return nil, fmt.Errorf("get asset status: %w", err)
	}

	return &asset, nil
}

// ListProjectAssets retrieves assets for a project with multi-tenancy guard.
// H1 Fix: Hard LIMIT prevents unbounded result sets.
// See STEP_85_VISION_BADGES.md Section 2
func (s *AssetService) ListProjectAssets(ctx context.Context, projectID uuid.UUID, orgID uuid.UUID) ([]models.ProjectAsset, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM project_assets pa
		JOIN projects p ON pa.project_id = p.id
		WHERE pa.project_id = $1 AND p.org_id = $2
		ORDER BY pa.created_at DESC
		LIMIT %d
	`, assetReturnColumns, MaxAssetsPerProject)

	rows, err := s.db.Query(ctx, query, projectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list project assets: %w", err)
	}
	defer rows.Close()

	var assets []models.ProjectAsset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, fmt.Errorf("scan project asset: %w", err)
		}
		assets = append(assets, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project assets: %w", err)
	}

	return assets, nil
}
