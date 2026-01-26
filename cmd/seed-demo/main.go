// Package main provides a demo data seeder for FutureBuild.
//
// QUALITY GATE: L7/L8 Senior Engineer Standards
// - Idempotent execution (safe to run multiple times)
// - Transaction safety (all-or-nothing)
// - Environment protection (refuses to run in production)
// - Realistic construction industry data
// - Comprehensive error handling with context
//
// See LAUNCH_PLAN.md Task 4.1: Demo Seed Script
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

// DemoOrganization represents the demo organization configuration.
type DemoOrganization struct {
	ID           uuid.UUID
	Name         string
	Slug         string
	ProjectLimit int
}

// DemoUser represents a demo user configuration.
type DemoUser struct {
	ID    uuid.UUID
	Email string
	Name  string
	Role  string // Admin, Builder, Client, Subcontractor
}

// DemoProject represents a demo project configuration.
type DemoProject struct {
	ID                uuid.UUID
	Name              string
	Address           string
	Status            string // Preconstruction, Active, Paused, Completed
	GSF               float64
	PermitIssuedDate  time.Time
	CompletionPercent int // Used to determine task statuses
}

// DemoContact represents a subcontractor/client contact.
type DemoContact struct {
	ID                uuid.UUID
	Name              string
	Company           string
	Phone             string
	Email             string
	Role              string // Client, Subcontractor
	ContactPreference string // SMS, Email, Both
}

// DemoTask represents a project task with WBS code.
type DemoTask struct {
	WBSCode  string
	Name     string
	Duration float64
	Status   string // Pending, Ready, In_Progress, Completed, Blocked, Delayed
}

const (
	// DemoOrgSlug is the unique identifier for the demo organization.
	// Used for idempotency checks.
	DemoOrgSlug = "acme-builders-demo"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// SAFETY CHECK: Refuse to run in production
	env := os.Getenv("APP_ENV")
	if env == "production" {
		log.Fatal("SAFETY: Demo seeder cannot run in production environment")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	log.Println("=== FutureBuild Demo Seeder ===")
	log.Printf("Environment: %s", env)

	// Check if demo data already exists (idempotency)
	var existingOrgID uuid.UUID
	err = conn.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", DemoOrgSlug).Scan(&existingOrgID)
	if err == nil {
		log.Printf("Demo organization already exists (ID: %s). Skipping seed.", existingOrgID)
		log.Println("To re-seed, delete the organization first:")
		log.Printf("  DELETE FROM organizations WHERE slug = '%s';", DemoOrgSlug)
		return
	}
	if err != pgx.ErrNoRows {
		log.Fatalf("Error checking for existing demo data: %v", err)
	}

	// Begin transaction for atomic seeding
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("Unable to begin transaction: %v", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("Warning: rollback failed: %v", err)
		}
	}()

	// Seed demo data
	if err := seedDemoData(ctx, tx); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("=== Demo Seeding Complete ===")
	log.Println("")
	log.Println("Demo Credentials:")
	log.Println("  Admin:   admin@acme-builders.demo")
	log.Println("  Builder: mike.johnson@acme-builders.demo")
	log.Println("  Client:  sarah.chen@homeowner.demo")
	log.Println("")
	log.Println("Use magic link auth to log in.")
}

func seedDemoData(ctx context.Context, tx pgx.Tx) error {
	// =========================================================================
	// 1. Organization
	// =========================================================================
	org := DemoOrganization{
		ID:           uuid.New(),
		Name:         "Acme Builders",
		Slug:         DemoOrgSlug,
		ProjectLimit: 10,
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO organizations (id, name, slug, settings, project_limit)
		VALUES ($1, $2, $3, '{"demo": true}'::jsonb, $4)
	`, org.ID, org.Name, org.Slug, org.ProjectLimit)
	if err != nil {
		return fmt.Errorf("insert organization: %w", err)
	}
	log.Printf("Created organization: %s (%s)", org.Name, org.ID)

	// =========================================================================
	// 2. Users
	// =========================================================================
	users := []DemoUser{
		{ID: uuid.New(), Email: "admin@acme-builders.demo", Name: "Alex Thompson", Role: "Admin"},
		{ID: uuid.New(), Email: "mike.johnson@acme-builders.demo", Name: "Mike Johnson", Role: "Builder"},
		{ID: uuid.New(), Email: "sarah.chen@homeowner.demo", Name: "Sarah Chen", Role: "Client"},
	}

	for _, u := range users {
		_, err := tx.Exec(ctx, `
			INSERT INTO users (id, org_id, email, name, role)
			VALUES ($1, $2, $3, $4, $5)
		`, u.ID, org.ID, u.Email, u.Name, u.Role)
		if err != nil {
			return fmt.Errorf("insert user %s: %w", u.Email, err)
		}
		log.Printf("Created user: %s (%s)", u.Name, u.Role)
	}

	// =========================================================================
	// 3. Contacts (Subcontractors)
	// =========================================================================
	contacts := []DemoContact{
		{ID: uuid.New(), Name: "Rodriguez Electric", Company: "Rodriguez Electrical LLC", Phone: "+15551234001", Email: "dispatch@rodriguez-electric.demo", Role: "Subcontractor", ContactPreference: "SMS"},
		{ID: uuid.New(), Name: "Premium Plumbing", Company: "Premium Plumbing Co", Phone: "+15551234002", Email: "jobs@premium-plumbing.demo", Role: "Subcontractor", ContactPreference: "Both"},
		{ID: uuid.New(), Name: "HVAC Masters", Company: "HVAC Masters Inc", Phone: "+15551234003", Email: "service@hvac-masters.demo", Role: "Subcontractor", ContactPreference: "Email"},
		{ID: uuid.New(), Name: "Drywall Pro", Company: "Drywall Professionals", Phone: "+15551234004", Email: "quotes@drywallpro.demo", Role: "Subcontractor", ContactPreference: "SMS"},
		{ID: uuid.New(), Name: "Summit Roofing", Company: "Summit Roofing Group", Phone: "+15551234005", Email: "bids@summit-roofing.demo", Role: "Subcontractor", ContactPreference: "Both"},
	}

	for _, c := range contacts {
		_, err := tx.Exec(ctx, `
			INSERT INTO contacts (id, org_id, name, company, phone, email, global_role, contact_preference)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, c.ID, org.ID, c.Name, c.Company, c.Phone, c.Email, c.Role, c.ContactPreference)
		if err != nil {
			return fmt.Errorf("insert contact %s: %w", c.Name, err)
		}
		log.Printf("Created contact: %s (%s)", c.Name, c.Company)
	}

	// =========================================================================
	// 4. Projects (3 at different completion stages)
	// =========================================================================
	baseDate := time.Now().AddDate(0, -3, 0) // 3 months ago

	projects := []DemoProject{
		{
			ID:                uuid.New(),
			Name:              "Chen Residence - New Construction",
			Address:           "1847 Oakwood Drive, Austin, TX 78701",
			Status:            "Active",
			GSF:               2800,
			PermitIssuedDate:  baseDate,
			CompletionPercent: 75,
		},
		{
			ID:                uuid.New(),
			Name:              "Morrison Kitchen Remodel",
			Address:           "422 Elm Street, Round Rock, TX 78664",
			Status:            "Active",
			GSF:               450,
			PermitIssuedDate:  baseDate.AddDate(0, 1, 0),
			CompletionPercent: 45,
		},
		{
			ID:                uuid.New(),
			Name:              "Westside Commercial Build",
			Address:           "8900 Commerce Blvd, Cedar Park, TX 78613",
			Status:            "Preconstruction",
			GSF:               12000,
			PermitIssuedDate:  time.Now().AddDate(0, 0, 14), // 2 weeks from now
			CompletionPercent: 0,
		},
	}

	for _, p := range projects {
		targetEnd := p.PermitIssuedDate.AddDate(0, 6, 0) // 6 months from permit

		_, err := tx.Exec(ctx, `
			INSERT INTO projects (id, org_id, name, address, permit_issued_date, target_end_date, gsf, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, p.ID, org.ID, p.Name, p.Address, p.PermitIssuedDate, targetEnd, p.GSF, p.Status)
		if err != nil {
			return fmt.Errorf("insert project %s: %w", p.Name, err)
		}
		log.Printf("Created project: %s (%d%% complete)", p.Name, p.CompletionPercent)

		// Insert project context
		_, err = tx.Exec(ctx, `
			INSERT INTO project_context (id, project_id, supply_chain_volatility, rough_inspection_latency, final_inspection_latency, zip_code, climate_zone)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, uuid.New(), p.ID, 2, 2, 5, "78701", "2A-Hot-Humid")
		if err != nil {
			return fmt.Errorf("insert project_context for %s: %w", p.Name, err)
		}

		// Seed tasks for this project
		if err := seedProjectTasks(ctx, tx, p); err != nil {
			return fmt.Errorf("seed tasks for %s: %w", p.Name, err)
		}
	}

	// =========================================================================
	// 5. Project Assignments (link contacts to projects)
	// =========================================================================
	// Assign all subs to the main Chen Residence project
	chenProject := projects[0]
	phases := []string{"6.0", "7.0", "8.0", "9.0", "10.0"}
	for i, c := range contacts {
		phaseIdx := i % len(phases)
		_, err := tx.Exec(ctx, `
			INSERT INTO project_assignments (id, project_id, contact_id, wbs_phase_id)
			VALUES ($1, $2, $3, $4)
		`, uuid.New(), chenProject.ID, c.ID, phases[phaseIdx])
		if err != nil {
			return fmt.Errorf("insert assignment for %s: %w", c.Name, err)
		}
	}
	log.Printf("Created %d project assignments", len(contacts))

	// =========================================================================
	// 6. Auth Token for Demo Login (optional - for testing)
	// =========================================================================
	// Create a pre-generated magic link token for the admin user
	demoToken := "demo-magic-link-token-12345"
	tokenHash := hashToken(demoToken)
	adminUser := users[0]

	_, err = tx.Exec(ctx, `
		INSERT INTO auth_tokens (token_hash, user_id, expires_at, used)
		VALUES ($1, $2, $3, false)
	`, tokenHash, adminUser.ID, time.Now().Add(24*time.Hour))
	if err != nil {
		// Non-fatal - table might not exist or token might conflict
		log.Printf("Note: Could not create demo auth token: %v", err)
	}

	return nil
}

// seedProjectTasks creates realistic WBS tasks for a project based on completion percentage.
func seedProjectTasks(ctx context.Context, tx pgx.Tx, project DemoProject) error {
	// Realistic residential construction WBS tasks (CPM-res1.0 subset)
	allTasks := []DemoTask{
		// Phase 5: Site Work
		{WBSCode: "5.2", Name: "Permit Issued", Duration: 0, Status: "Completed"},
		{WBSCode: "5.3", Name: "Site Preparation", Duration: 3, Status: "Completed"},
		{WBSCode: "5.4", Name: "Excavation", Duration: 2, Status: "Completed"},

		// Phase 6: Foundation
		{WBSCode: "6.1", Name: "Footings", Duration: 3, Status: "Completed"},
		{WBSCode: "6.2", Name: "Foundation Walls", Duration: 4, Status: "Completed"},
		{WBSCode: "6.3", Name: "Foundation Waterproofing", Duration: 2, Status: "Completed"},
		{WBSCode: "6.4", Name: "Backfill", Duration: 1, Status: "Completed"},
		{WBSCode: "6.5", Name: "Slab on Grade", Duration: 3, Status: "Completed"},

		// Phase 7: Framing
		{WBSCode: "7.1", Name: "Floor Framing", Duration: 4, Status: "Completed"},
		{WBSCode: "7.2", Name: "Wall Framing", Duration: 6, Status: "Completed"},
		{WBSCode: "7.3", Name: "Roof Framing", Duration: 5, Status: "Completed"},
		{WBSCode: "7.4", Name: "Sheathing", Duration: 3, Status: "Completed"},
		{WBSCode: "7.5", Name: "Framing Inspection", Duration: 1, Status: "Completed"},

		// Phase 8: Exterior
		{WBSCode: "8.1", Name: "Windows & Doors", Duration: 3, Status: "In_Progress"},
		{WBSCode: "8.2", Name: "Roofing", Duration: 4, Status: "In_Progress"},
		{WBSCode: "8.3", Name: "Siding", Duration: 5, Status: "Pending"},
		{WBSCode: "8.4", Name: "Exterior Trim", Duration: 3, Status: "Pending"},

		// Phase 9: Rough-Ins
		{WBSCode: "9.1", Name: "Electrical Rough", Duration: 4, Status: "Pending"},
		{WBSCode: "9.2", Name: "Plumbing Rough", Duration: 4, Status: "Pending"},
		{WBSCode: "9.3", Name: "HVAC Rough", Duration: 3, Status: "Pending"},
		{WBSCode: "9.4", Name: "Rough Inspection", Duration: 1, Status: "Pending"},

		// Phase 10: Insulation & Drywall
		{WBSCode: "10.1", Name: "Insulation", Duration: 2, Status: "Pending"},
		{WBSCode: "10.2", Name: "Drywall Hang", Duration: 4, Status: "Pending"},
		{WBSCode: "10.3", Name: "Drywall Finish", Duration: 5, Status: "Pending"},

		// Phase 11: Interior Finishes
		{WBSCode: "11.1", Name: "Interior Paint", Duration: 5, Status: "Pending"},
		{WBSCode: "11.2", Name: "Cabinets", Duration: 3, Status: "Pending"},
		{WBSCode: "11.3", Name: "Countertops", Duration: 2, Status: "Pending"},
		{WBSCode: "11.4", Name: "Flooring", Duration: 4, Status: "Pending"},
		{WBSCode: "11.5", Name: "Interior Doors & Trim", Duration: 4, Status: "Pending"},

		// Phase 12: Final
		{WBSCode: "12.1", Name: "Electrical Finish", Duration: 2, Status: "Pending"},
		{WBSCode: "12.2", Name: "Plumbing Finish", Duration: 2, Status: "Pending"},
		{WBSCode: "12.3", Name: "HVAC Finish", Duration: 1, Status: "Pending"},
		{WBSCode: "12.4", Name: "Final Cleaning", Duration: 2, Status: "Pending"},
		{WBSCode: "12.5", Name: "Final Inspection", Duration: 1, Status: "Pending"},
		{WBSCode: "12.6", Name: "Certificate of Occupancy", Duration: 0, Status: "Pending"},
	}

	// Adjust statuses based on completion percentage
	completedCount := len(allTasks) * project.CompletionPercent / 100
	inProgressCount := 2 // Always have a couple in progress if not 0% or 100%

	if project.CompletionPercent == 0 {
		inProgressCount = 0
	}
	if project.CompletionPercent == 100 {
		completedCount = len(allTasks)
		inProgressCount = 0
	}

	startDate := project.PermitIssuedDate
	for i := range allTasks {
		task := &allTasks[i]

		// Override status based on completion
		if i < completedCount {
			task.Status = "Completed"
		} else if i < completedCount+inProgressCount {
			task.Status = "In_Progress"
		} else {
			task.Status = "Pending"
		}

		// Calculate dates
		earlyStart := startDate
		earlyFinish := earlyStart.AddDate(0, 0, int(task.Duration))
		if task.Duration > 0 {
			startDate = earlyFinish // Next task starts when this one finishes
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO project_tasks (id, project_id, wbs_code, name, early_start, early_finish, calculated_duration, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, uuid.New(), project.ID, task.WBSCode, task.Name, earlyStart, earlyFinish, task.Duration, task.Status)
		if err != nil {
			return fmt.Errorf("insert task %s: %w", task.WBSCode, err)
		}
	}

	log.Printf("  Created %d tasks (%d completed, %d in progress)", len(allTasks), completedCount, inProgressCount)
	return nil
}

// hashToken creates a SHA-256 hash of a token for secure storage.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
