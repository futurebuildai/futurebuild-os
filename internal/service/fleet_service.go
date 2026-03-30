package service

import (
	"context"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FleetService implements FleetServicer.
// See BACKEND_SCOPE.md Section 20.3
type FleetService struct {
	db *pgxpool.Pool
}

func NewFleetService(db *pgxpool.Pool) *FleetService {
	return &FleetService{db: db}
}

func (s *FleetService) CreateFleetAsset(ctx context.Context, orgID uuid.UUID, asset *models.FleetAsset) error {
	asset.OrgID = orgID
	query := `
		INSERT INTO fleet_assets (org_id, asset_number, asset_type, make, model, year, vin, license_plate,
			purchase_date, purchase_cost_cents, current_value_cents, status, location, notes, visible_to_roles)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at`

	return s.db.QueryRow(ctx, query,
		asset.OrgID, asset.AssetNumber, asset.AssetType, asset.Make, asset.Model,
		asset.Year, asset.VIN, asset.LicensePlate, asset.PurchaseDate,
		asset.PurchaseCostCents, asset.CurrentValueCents, asset.Status,
		asset.Location, asset.Notes, asset.VisibleToRoles,
	).Scan(&asset.ID, &asset.CreatedAt, &asset.UpdatedAt)
}

func (s *FleetService) GetFleetAsset(ctx context.Context, assetID, orgID uuid.UUID) (*models.FleetAsset, error) {
	query := `
		SELECT id, org_id, asset_number, asset_type, make, model, year, vin, license_plate,
			purchase_date, purchase_cost_cents, current_value_cents, status, location, notes,
			visible_to_roles, created_at, updated_at
		FROM fleet_assets
		WHERE id = $1 AND org_id = $2`

	var asset models.FleetAsset
	err := s.db.QueryRow(ctx, query, assetID, orgID).Scan(
		&asset.ID, &asset.OrgID, &asset.AssetNumber, &asset.AssetType,
		&asset.Make, &asset.Model, &asset.Year, &asset.VIN, &asset.LicensePlate,
		&asset.PurchaseDate, &asset.PurchaseCostCents, &asset.CurrentValueCents,
		&asset.Status, &asset.Location, &asset.Notes,
		&asset.VisibleToRoles, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *FleetService) ListFleetAssets(ctx context.Context, orgID uuid.UUID, status, assetType, callerRole string) ([]models.FleetAsset, error) {
	query := `
		SELECT id, org_id, asset_number, asset_type, make, model, year, vin, license_plate,
			purchase_date, purchase_cost_cents, current_value_cents, status, location, notes,
			visible_to_roles, created_at, updated_at
		FROM fleet_assets
		WHERE org_id = $1`
	args := []interface{}{orgID}
	argIdx := 2

	// Phase 20: Role-based visibility — Admins and PMs see everything; others see only
	// assets where visible_to_roles is NULL/empty OR their role is in the array.
	if callerRole != "" && callerRole != "Admin" && callerRole != "PM" {
		query += fmt.Sprintf(" AND (visible_to_roles IS NULL OR visible_to_roles = '{}' OR $%d = ANY(visible_to_roles))", argIdx)
		args = append(args, callerRole)
		argIdx++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if assetType != "" {
		query += fmt.Sprintf(" AND asset_type = $%d", argIdx)
		args = append(args, assetType)
	}
	query += " ORDER BY asset_number"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []models.FleetAsset
	for rows.Next() {
		var a models.FleetAsset
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.AssetNumber, &a.AssetType,
			&a.Make, &a.Model, &a.Year, &a.VIN, &a.LicensePlate,
			&a.PurchaseDate, &a.PurchaseCostCents, &a.CurrentValueCents,
			&a.Status, &a.Location, &a.Notes,
			&a.VisibleToRoles, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}

	return assets, nil
}

func (s *FleetService) UpdateFleetAsset(ctx context.Context, assetID, orgID uuid.UUID, asset *models.FleetAsset) (*models.FleetAsset, error) {
	query := `
		UPDATE fleet_assets SET
			asset_number = $3, asset_type = $4, make = $5, model = $6, year = $7,
			vin = $8, license_plate = $9, purchase_date = $10, purchase_cost_cents = $11,
			current_value_cents = $12, status = $13, location = $14, notes = $15, visible_to_roles = $16
		WHERE id = $1 AND org_id = $2
		RETURNING id, org_id, asset_number, asset_type, make, model, year, vin, license_plate,
			purchase_date, purchase_cost_cents, current_value_cents, status, location, notes,
			visible_to_roles, created_at, updated_at`

	var updated models.FleetAsset
	err := s.db.QueryRow(ctx, query,
		assetID, orgID, asset.AssetNumber, asset.AssetType, asset.Make, asset.Model,
		asset.Year, asset.VIN, asset.LicensePlate, asset.PurchaseDate,
		asset.PurchaseCostCents, asset.CurrentValueCents, asset.Status,
		asset.Location, asset.Notes, asset.VisibleToRoles,
	).Scan(
		&updated.ID, &updated.OrgID, &updated.AssetNumber, &updated.AssetType,
		&updated.Make, &updated.Model, &updated.Year, &updated.VIN, &updated.LicensePlate,
		&updated.PurchaseDate, &updated.PurchaseCostCents, &updated.CurrentValueCents,
		&updated.Status, &updated.Location, &updated.Notes,
		&updated.VisibleToRoles, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// AllocateEquipment assigns an asset to a project with conflict detection.
// The DB GIST constraint is the final safety net, but we check availability first
// for a friendlier error message.
func (s *FleetService) AllocateEquipment(ctx context.Context, alloc *models.EquipmentAllocation) error {
	// Pre-check availability for better error messages
	available, err := s.CheckEquipmentAvailability(ctx, alloc.AssetID, alloc.AllocatedFrom, alloc.AllocatedTo)
	if err != nil {
		return err
	}
	if !available {
		return fmt.Errorf("equipment %s is not available for the requested dates %s to %s",
			alloc.AssetID, alloc.AllocatedFrom.Format("2006-01-02"), alloc.AllocatedTo.Format("2006-01-02"))
	}

	query := `
		INSERT INTO equipment_allocations (asset_id, project_id, task_id, allocated_from, allocated_to, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return s.db.QueryRow(ctx, query,
		alloc.AssetID, alloc.ProjectID, alloc.TaskID,
		alloc.AllocatedFrom, alloc.AllocatedTo, alloc.Status, alloc.Notes,
	).Scan(&alloc.ID, &alloc.CreatedAt, &alloc.UpdatedAt)
}

// CheckEquipmentAvailability returns true if the asset has no overlapping allocations.
// Deterministic date range overlap check.
func (s *FleetService) CheckEquipmentAvailability(ctx context.Context, assetID uuid.UUID, from, to time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM equipment_allocations
		WHERE asset_id = $1
			AND status IN ('planned', 'active')
			AND daterange($2::date, $3::date, '[]') && daterange(allocated_from, allocated_to, '[]')`

	var count int
	err := s.db.QueryRow(ctx, query, assetID, from, to).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (s *FleetService) GetProjectEquipment(ctx context.Context, projectID, orgID uuid.UUID) ([]models.EquipmentAllocation, error) {
	query := `
		SELECT ea.id, ea.asset_id, ea.project_id, ea.task_id, ea.allocated_from, ea.allocated_to,
			ea.status, ea.notes, ea.created_at, ea.updated_at
		FROM equipment_allocations ea
		INNER JOIN fleet_assets fa ON ea.asset_id = fa.id
		WHERE ea.project_id = $1 AND fa.org_id = $2
		ORDER BY ea.allocated_from`

	rows, err := s.db.Query(ctx, query, projectID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocs []models.EquipmentAllocation
	for rows.Next() {
		var a models.EquipmentAllocation
		if err := rows.Scan(
			&a.ID, &a.AssetID, &a.ProjectID, &a.TaskID,
			&a.AllocatedFrom, &a.AllocatedTo, &a.Status, &a.Notes,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		allocs = append(allocs, a)
	}

	return allocs, nil
}

func (s *FleetService) LogMaintenance(ctx context.Context, log *models.MaintenanceLog) error {
	query := `
		INSERT INTO maintenance_logs (asset_id, maintenance_type, description, scheduled_date, completed_date, cost_cents, vendor_name, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return s.db.QueryRow(ctx, query,
		log.AssetID, log.MaintenanceType, log.Description, log.ScheduledDate,
		log.CompletedDate, log.CostCents, log.VendorName, log.Notes,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
}

func (s *FleetService) GetMaintenanceHistory(ctx context.Context, assetID uuid.UUID) ([]models.MaintenanceLog, error) {
	query := `
		SELECT id, asset_id, maintenance_type, description, scheduled_date, completed_date,
			cost_cents, vendor_name, notes, created_at, updated_at
		FROM maintenance_logs
		WHERE asset_id = $1
		ORDER BY COALESCE(completed_date, scheduled_date) DESC`

	rows, err := s.db.Query(ctx, query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.MaintenanceLog
	for rows.Next() {
		var l models.MaintenanceLog
		if err := rows.Scan(
			&l.ID, &l.AssetID, &l.MaintenanceType, &l.Description,
			&l.ScheduledDate, &l.CompletedDate, &l.CostCents, &l.VendorName,
			&l.Notes, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

// GetUpcomingMaintenance finds scheduled maintenance within the given days across all org assets.
func (s *FleetService) GetUpcomingMaintenance(ctx context.Context, orgID uuid.UUID, withinDays int) ([]models.MaintenanceLog, error) {
	query := `
		SELECT ml.id, ml.asset_id, ml.maintenance_type, ml.description, ml.scheduled_date, ml.completed_date,
			ml.cost_cents, ml.vendor_name, ml.notes, ml.created_at, ml.updated_at
		FROM maintenance_logs ml
		INNER JOIN fleet_assets fa ON ml.asset_id = fa.id
		WHERE fa.org_id = $1
			AND ml.completed_date IS NULL
			AND ml.scheduled_date IS NOT NULL
			AND ml.scheduled_date <= NOW() + ($2 || ' days')::interval
		ORDER BY ml.scheduled_date ASC`

	rows, err := s.db.Query(ctx, query, orgID, fmt.Sprintf("%d", withinDays))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.MaintenanceLog
	for rows.Next() {
		var l models.MaintenanceLog
		if err := rows.Scan(
			&l.ID, &l.AssetID, &l.MaintenanceType, &l.Description,
			&l.ScheduledDate, &l.CompletedDate, &l.CostCents, &l.VendorName,
			&l.Notes, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}
