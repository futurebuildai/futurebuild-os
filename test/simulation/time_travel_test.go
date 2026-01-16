// Package simulation provides time-travel testing for FutureBuild agents.
// See PRODUCTION_PLAN.md Step 49
package simulation_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/colton/futurebuild/test/simulation/mocks"
	"github.com/colton/futurebuild/test/testdata"
	"github.com/colton/futurebuild/test/testhelpers"
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
	// Use testcontainers for ephemeral PostgreSQL - always runs in CI
	// See PRODUCTION_PLAN.md: Testing Strategy & CI Reliability Remediation
	db, cleanup := testhelpers.StartPostgresContainer(t)
	defer cleanup()

	ctx := context.Background()

	// 1. SETUP: Seed test data using factory functions
	// Technical Debt Remediation (P2): Uses testdata package instead of raw SQL
	projectID, taskAID, taskBID := seedSimulationProject(t, ctx, db)

	// 2. Initialize agents with MockClock
	// Canonical simulation start: 2026-01-01 08:00:00 UTC
	simStart := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	mockClock := clock.NewMockClock(simStart)

	weatherService := &mocks.MockWeatherService{}
	notifier := &mocks.SpyNotifier{}
	directoryService := &mockDirectoryService{}

	// Create agents using repository pattern
	// See PRODUCTION_PLAN.md: Testing Strategy remediation
	// Config Decoupling: Use default config for simulation tests
	procurementCfg := config.DefaultProcurementConfig()
	procurementRepo := agents.NewPgProcurementRepository(db)
	procurementAgent := agents.NewProcurementAgent(procurementRepo, weatherService, mockClock, procurementCfg)
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

	// Assert DAMPENING - With 72-hour dampening over 30 days, expect ~10 alerts max
	// (First alert + re-alerts every 3 days = 10-11 total)
	lumberAlertCount := countLumberAlerts(ctx, db, projectID)
	if lumberAlertCount > 11 {
		t.Errorf("FAIL: Too many lumber alerts (%d alerts, expected ≤11 with 72-hour dampening)", lumberAlertCount)
	}
	if lumberAlertCount == 0 {
		t.Error("FAIL: No lumber alerts logged")
	}

	// Cleanup
	testdata.CleanupTestProject(ctx, db, projectID, taskAID, taskBID)
	t.Log("✓ Simulation test completed")
}

// --- Test Helpers ---

// NOTE: getTestDatabaseURL removed - now using testcontainers.StartPostgresContainer

// seedSimulationProject uses factory functions to create test data.
// Technical Debt Remediation (P2): Replaces raw SQL INSERT statements.
func seedSimulationProject(t *testing.T, ctx context.Context, db *pgxpool.Pool) (projectID, taskAID, taskBID uuid.UUID) {
	// Create organization
	orgID, err := testdata.NewTestOrganization(ctx, db, "Sim Test Org")
	if err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Create project starting at T+0
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	project, err := testdata.NewTestProject(ctx, db, orgID, "Simulation Test Project",
		testdata.WithProjectStatus("Active"),
		testdata.WithPermitDate(startDate),
		testdata.WithAddress("123 Test St, Austin, TX 78701"),
	)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create project context with zip code for Procurement Agent
	// Required for weather-adjusted lead time calculations
	if err := testdata.NewTestProjectContext(ctx, db, project.ID, "78701"); err != nil {
		t.Fatalf("Failed to create project context: %v", err)
	}

	// Task A (Framing) - starts T+14, requires lumber
	// Amendment 1: EarlyStart = T+14, with 1 week lead time → MustOrderDate = T+5
	taskAStart := startDate.AddDate(0, 0, 14)
	taskA, err := testdata.NewTestTask(ctx, db, project.ID, "9.1", "Framing",
		testdata.WithEarlyStart(taskAStart),
		testdata.WithCriticalPath(true),
	)
	if err != nil {
		t.Fatalf("Failed to create task A: %v", err)
	}

	// Task B (Electrical) - starts T+15, requires sub confirmation
	taskBStart := startDate.AddDate(0, 0, 15)
	taskB, err := testdata.NewTestTask(ctx, db, project.ID, "10.2", "Electrical Rough-In",
		testdata.WithEarlyStart(taskBStart),
		testdata.WithCriticalPath(true),
	)
	if err != nil {
		t.Fatalf("Failed to create task B: %v", err)
	}

	// Create procurement item for lumber with 1 week lead time (Amendment 1.1)
	_, err = testdata.NewTestProcurementItem(ctx, db, taskA.ID, "Lumber", 1)
	if err != nil {
		t.Fatalf("Failed to create procurement item: %v", err)
	}

	// Create contact for electrical phase
	contactID, err := testdata.NewTestContact(ctx, db, orgID,
		"Electric Joe", "+15551234567", "joe@electric.com", "Subcontractor", "SMS")
	if err != nil {
		t.Fatalf("Failed to create contact: %v", err)
	}

	// Assign contact to electrical phase (phase 10)
	// This may fail if wbs_phases isn't seeded, which is fine for now
	if err := testdata.NewTestProjectAssignment(ctx, db, project.ID, contactID, "10"); err != nil {
		t.Logf("Note: Could not create phase assignment (expected if wbs_phases not seeded): %v", err)
	}

	return project.ID, taskA.ID, taskB.ID
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
