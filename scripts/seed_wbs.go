package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WBSTemplate struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	IsDefault     bool   `json:"is_default"`
	EntryPointWBS string `json:"entry_point_wbs"`
}

type WBSTask struct {
	Code             string   `json:"code"`
	Name             string   `json:"name"`
	BaseDurationDays float64  `json:"base_duration_days"`
	ResponsibleParty string   `json:"responsible_party"`
	Deliverable      string   `json:"deliverable"`
	Notes            string   `json:"notes"`
	IsInspection     bool     `json:"is_inspection"`
	IsMilestone      bool     `json:"is_milestone"`
	IsLongLead       bool     `json:"is_long_lead"`
	LeadTimeWeeksMin int      `json:"lead_time_weeks_min"`
	LeadTimeWeeksMax int      `json:"lead_time_weeks_max"`
	PredecessorCodes []string `json:"predecessor_codes"`
}

type WBSPhase struct {
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	IsWeatherSensitive bool      `json:"is_weather_sensitive"`
	SortOrder          int       `json:"sort_order"`
	Tasks              []WBSTask `json:"tasks"`
}

type WBSMasterData struct {
	Template WBSTemplate `json:"template"`
	Phases   []WBSPhase  `json:"phases"`
}

func main() {
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

	// Read JSON data
	data, err := os.ReadFile("internal/data/wbs_master.json")
	if err != nil {
		log.Fatalf("Unable to read WBS master data: %v", err)
	}

	var wbsData WBSMasterData
	if err := json.Unmarshal(data, &wbsData); err != nil {
		log.Fatalf("Unable to unmarshal WBS master data: %v", err)
	}

	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("Unable to begin transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Insert Template
	templateID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO wbs_templates (id, name, version, is_default, entry_point_wbs, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		templateID, wbsData.Template.Name, wbsData.Template.Version, wbsData.Template.IsDefault, wbsData.Template.EntryPointWBS, time.Now())
	if err != nil {
		log.Fatalf("Failed to insert template: %v", err)
	}

	fmt.Printf("Seeding template: %s\n", wbsData.Template.Name)

	// 2. Insert Phases and Tasks
	for _, phase := range wbsData.Phases {
		phaseID := uuid.New()
		_, err = tx.Exec(ctx, `
			INSERT INTO wbs_phases (id, template_id, code, name, is_weather_sensitive, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			phaseID, templateID, phase.Code, phase.Name, phase.IsWeatherSensitive, phase.SortOrder)
		if err != nil {
			log.Fatalf("Failed to insert phase %s: %v", phase.Code, err)
		}

		fmt.Printf("  Seeding phase: %s (%s)\n", phase.Name, phase.Code)

		for _, task := range phase.Tasks {
			taskID := uuid.New()
			_, err = tx.Exec(ctx, `
				INSERT INTO wbs_tasks (id, phase_id, code, name, base_duration_days, responsible_party, deliverable, notes, is_inspection, is_milestone, is_long_lead, lead_time_weeks_min, lead_time_weeks_max, predecessor_codes, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
				taskID, phaseID, task.Code, task.Name, task.BaseDurationDays, task.ResponsibleParty, task.Deliverable, task.Notes, task.IsInspection, task.IsMilestone, task.IsLongLead, task.LeadTimeWeeksMin, task.LeadTimeWeeksMax, task.PredecessorCodes, time.Now())
			if err != nil {
				log.Fatalf("Failed to insert task %s: %v", task.Code, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("WBS Seeding complete!")
}
