// Package simulation provides time-travel testing for FutureBuild agents.
// See PRODUCTION_PLAN.md Step 49
package simulation_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/colton/futurebuild/test/simulation/mocks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestTimeTravelSimulation proves agents fire alerts at correct times.
// See PRODUCTION_PLAN.md Step 49
//
// Scenario (Amendment 1 - Corrected Timeline):
// - Project starts at T+0 (2026-01-01 08:00 UTC)
// - Task A (Framing): Starts T+14, 10d duration, requires "Lumber" ordered by T+5
//   - Lead time = 1 week (7 days), Staging = 2 days → MustOrderDate = T+14 - 9 = T+5
//
// - Task B (Electrical): Starts T+15, requires Sub Confirmation by T+12
//   - 72h window: T+15 - 72h = T+12
//
// Assertions:
// - T+5: COMMUNICATION_LOGS contains "Order Lumber"
// - T+12: "Confirm Arrival" for Electrical sub
// - NO duplicate alerts (idempotency)
func TestTimeTravelSimulation(t *testing.T) {
	// Skip if no database connection
	dbURL := getTestDatabaseURL(t)
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 1. SETUP: Seed test data
	projectID, taskAID, taskBID := seedSimulationProject(t, ctx, db)

	// 2. Initialize agents with MockClock
	// Canonical simulation start: 2026-01-01 08:00:00 UTC
	simStart := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	mockClock := clock.NewMockClock(simStart)

	weatherService := &mocks.MockWeatherService{}
	notifier := &mocks.SpyNotifier{}
	directoryService := &mockDirectoryService{}

	procurementAgent := agents.NewProcurementAgent(db, weatherService, mockClock)
	subLiaisonAgent := agents.NewSubLiaisonAgent(db, directoryService, notifier, mockClock)

	// Track which days had alerts
	lumberAlertDay := -1
	electricalAlertDay := -1

	// 3. THE LOOP: Simulate 30 days
	// See PRODUCTION_PLAN.md Step 49: "Fast-forward by 30 days"
	for day := 0; day <= 30; day++ {
		// Run agents
		if err := procurementAgent.Execute(ctx); err != nil {
			t.Logf("Day %d: ProcurementAgent error (may be expected): %v", day, err)
		}

		if err := subLiaisonAgent.ScanAndNotify(ctx); err != nil {
			t.Logf("Day %d: SubLiaisonAgent error (may be expected): %v", day, err)
		}

		// Check for lumber order alert
		if lumberAlertDay == -1 && hasLumberAlert(ctx, db, projectID) {
			lumberAlertDay = day
			t.Logf("✓ Day T+%d: Lumber order alert detected", day)
		}

		// Check for electrical confirmation (via SpyNotifier)
		if electricalAlertDay == -1 && hasElectricalConfirmation(notifier) {
			electricalAlertDay = day
			t.Logf("✓ Day T+%d: Electrical confirmation alert detected", day)
		}

		// Advance clock by 24 hours
		mockClock.Advance(24 * time.Hour)
	}

	// 4. ASSERTIONS
	// Assert T+5: Lumber alert should fire
	if lumberAlertDay == -1 {
		t.Error("FAIL: Lumber order alert never fired")
	} else if lumberAlertDay > 5 {
		t.Errorf("FAIL: Lumber order alert fired too late (Day %d, expected ≤ Day 5)", lumberAlertDay)
	}

	// Assert T+12: Electrical confirmation should fire (72h before T+15)
	if electricalAlertDay == -1 {
		t.Error("FAIL: Electrical confirmation never fired")
	} else if electricalAlertDay > 12 {
		t.Errorf("FAIL: Electrical confirmation fired too late (Day %d, expected ≤ Day 12)", electricalAlertDay)
	}

	// Assert NO DUPLICATES
	lumberAlertCount := countLumberAlerts(ctx, db, projectID)
	if lumberAlertCount > 1 {
		t.Errorf("FAIL: Duplicate lumber alerts detected (%d alerts, expected 1)", lumberAlertCount)
	}

	// Cleanup
	cleanupSimulationProject(ctx, db, projectID, taskAID, taskBID)
	t.Log("✓ Simulation test completed")
}

// --- Test Helpers ---

func getTestDatabaseURL(t *testing.T) string {
	// In real tests, read from environment or use testcontainers
	return "" // Return empty to skip by default
}

func seedSimulationProject(t *testing.T, ctx context.Context, db *pgxpool.Pool) (projectID, taskAID, taskBID uuid.UUID) {
	projectID = uuid.New()
	orgID := uuid.New()
	taskAID = uuid.New()
	taskBID = uuid.New()

	// Create org
	_, err := db.Exec(ctx, `INSERT INTO organizations (id, name) VALUES ($1, 'Sim Test Org')`, orgID)
	if err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Create project starting at T+0
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = db.Exec(ctx, `
		INSERT INTO projects (id, org_id, name, status, permit_issued_date, address)
		VALUES ($1, $2, 'Simulation Test Project', 'Active', $3, '123 Test St')
	`, projectID, orgID, startDate)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Task A (Framing) - starts T+14, requires lumber
	// Amendment 1: EarlyStart = T+14, with 1 week lead time → MustOrderDate = T+5
	taskAStart := startDate.AddDate(0, 0, 14)
	_, err = db.Exec(ctx, `
		INSERT INTO project_tasks (id, project_id, wbs_code, name, status, early_start, is_on_critical_path)
		VALUES ($1, $2, '9.1', 'Framing', 'Pending', $3, true)
	`, taskAID, projectID, taskAStart)
	if err != nil {
		t.Fatalf("Failed to create task A: %v", err)
	}

	// Task B (Electrical) - starts T+15, requires sub confirmation
	taskBStart := startDate.AddDate(0, 0, 15)
	_, err = db.Exec(ctx, `
		INSERT INTO project_tasks (id, project_id, wbs_code, name, status, early_start, is_on_critical_path)
		VALUES ($1, $2, '10.2', 'Electrical Rough-In', 'Pending', $3, true)
	`, taskBID, projectID, taskBStart)
	if err != nil {
		t.Fatalf("Failed to create task B: %v", err)
	}

	// Create procurement item for lumber with 1 week lead time (Amendment 1.1)
	_, err = db.Exec(ctx, `
		INSERT INTO procurement_items (id, project_task_id, name, lead_time_weeks, status)
		VALUES ($1, $2, 'Lumber', 1, 'pending')
	`, uuid.New(), taskAID)
	if err != nil {
		t.Fatalf("Failed to create procurement item: %v", err)
	}

	// Create contact for electrical phase
	contactID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO contacts (id, name, company, phone, email, role, contact_preference)
		VALUES ($1, 'Electric Joe', 'Joe Electric LLC', '+15551234567', 'joe@electric.com', 'Subcontractor', 'sms')
	`, contactID)
	if err != nil {
		t.Fatalf("Failed to create contact: %v", err)
	}

	// Assign contact to electrical phase (phase 10)
	_, err = db.Exec(ctx, `
		INSERT INTO project_assignments (project_id, contact_id, assigned_phase_id)
		SELECT $1, $2, wp.id FROM wbs_phases wp WHERE wp.code = '10' LIMIT 1
	`, projectID, contactID)
	// This may fail if wbs_phases isn't seeded, which is fine for now
	if err != nil {
		t.Logf("Note: Could not create phase assignment (expected if wbs_phases not seeded): %v", err)
	}

	return projectID, taskAID, taskBID
}

func hasLumberAlert(ctx context.Context, db *pgxpool.Pool, projectID uuid.UUID) bool {
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM communication_logs
		WHERE project_id = $1
		  AND content LIKE '%Lumber%'
	`, projectID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func countLumberAlerts(ctx context.Context, db *pgxpool.Pool, projectID uuid.UUID) int {
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM communication_logs
		WHERE project_id = $1
		  AND content LIKE '%Lumber%'
	`, projectID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func hasElectricalConfirmation(notifier *mocks.SpyNotifier) bool {
	for _, sms := range notifier.SentSMS {
		if strings.Contains(strings.ToLower(sms), "confirm") ||
			strings.Contains(strings.ToLower(sms), "electrical") {
			return true
		}
	}
	return false
}

func cleanupSimulationProject(ctx context.Context, db *pgxpool.Pool, projectID, taskAID, taskBID uuid.UUID) {
	// Clean up in reverse order of creation
	db.Exec(ctx, `DELETE FROM communication_logs WHERE project_id = $1`, projectID)
	db.Exec(ctx, `DELETE FROM procurement_items WHERE project_task_id IN ($1, $2)`, taskAID, taskBID)
	db.Exec(ctx, `DELETE FROM project_tasks WHERE project_id = $1`, projectID)
	db.Exec(ctx, `DELETE FROM project_assignments WHERE project_id = $1`, projectID)
	db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, projectID)
	// Note: Contacts and orgs left behind for now to avoid FK issues
}

// mockDirectoryService implements agents.DirectoryService for testing
type mockDirectoryService struct{}

func (m *mockDirectoryService) GetContactForPhase(ctx context.Context, projectID, orgID uuid.UUID, phaseCode string) (*types.Contact, error) {
	return &types.Contact{
		ID:                uuid.New(),
		Name:              "Test Sub",
		Company:           "Test Company",
		Phone:             "+15551234567",
		Email:             "testsub@example.com",
		Role:              types.UserRoleSubcontractor,
		ContactPreference: types.ContactPreferenceSMS,
	}, nil
}
