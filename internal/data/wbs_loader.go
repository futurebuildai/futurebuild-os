// Package data provides master data loading for the FutureBuild scheduling engine.
package data

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/colton/futurebuild/pkg/types"
)

//go:embed wbs_master.json
var wbsMasterFS embed.FS

// WBSMasterData is the top-level JSON structure for the WBS template file.
type WBSMasterData struct {
	Template WBSTemplate `json:"template"`
	Phases   []WBSPhase  `json:"phases"`
}

// WBSTemplate holds metadata about the WBS template.
type WBSTemplate struct {
	EntryPointWBS string `json:"entry_point_wbs"`
	IsDefault     bool   `json:"is_default"`
	Name          string `json:"name"`
	Version       string `json:"version"`
}

// WBSPhase represents a phase grouping in the WBS template.
type WBSPhase struct {
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	IsWeatherSensitive bool      `json:"is_weather_sensitive"`
	SortOrder          int       `json:"sort_order"`
	Tasks              []WBSTask `json:"tasks"`
}

// WBSTask is a single task from the WBS master template (JSON representation).
// This is the loader-specific struct — it lacks DB fields like ID and PhaseID
// which are only present in the models.WBSTask (database model).
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

// WBSDependency represents a dependency edge parsed from the WBS template.
type WBSDependency struct {
	PredecessorCode string
	SuccessorCode   string
	Type            types.DependencyType
	LagDays         int
}

var (
	masterData     *WBSMasterData
	masterDataOnce sync.Once
	masterDataErr  error
)

// LoadMasterWBS parses wbs_master.json into typed structs for reuse by
// the scoping engine, preview service, and project hydration.
// The result is cached after the first load (singleton pattern).
func LoadMasterWBS() ([]WBSTask, []WBSDependency, error) {
	masterDataOnce.Do(func() {
		data, err := wbsMasterFS.ReadFile("wbs_master.json")
		if err != nil {
			masterDataErr = fmt.Errorf("failed to read wbs_master.json: %w", err)
			return
		}
		var parsed WBSMasterData
		if err := json.Unmarshal(data, &parsed); err != nil {
			masterDataErr = fmt.Errorf("failed to parse wbs_master.json: %w", err)
			return
		}
		masterData = &parsed
	})

	if masterDataErr != nil {
		return nil, nil, masterDataErr
	}

	tasks, deps := flattenWBS(masterData)
	return tasks, deps, nil
}

// LoadMasterWBSPhases returns the raw phase-grouped structure.
func LoadMasterWBSPhases() ([]WBSPhase, error) {
	// Trigger the master data load
	_, _, err := LoadMasterWBS()
	if err != nil {
		return nil, err
	}
	// Return a copy of phases
	result := make([]WBSPhase, len(masterData.Phases))
	copy(result, masterData.Phases)
	return result, nil
}

// flattenWBS extracts all tasks and builds dependency edges from predecessor_codes.
func flattenWBS(data *WBSMasterData) ([]WBSTask, []WBSDependency) {
	var tasks []WBSTask
	var deps []WBSDependency

	for _, phase := range data.Phases {
		for _, task := range phase.Tasks {
			tasks = append(tasks, task)

			for _, predCode := range task.PredecessorCodes {
				deps = append(deps, WBSDependency{
					PredecessorCode: predCode,
					SuccessorCode:   task.Code,
					Type:            types.DependencyTypeFS, // All WBS template deps are FS
					LagDays:         0,
				})
			}
		}
	}

	return tasks, deps
}

// TasksByCode returns a map from WBS code to WBSTask for O(1) lookups.
func TasksByCode(tasks []WBSTask) map[string]WBSTask {
	m := make(map[string]WBSTask, len(tasks))
	for _, t := range tasks {
		m[t.Code] = t
	}
	return m
}
