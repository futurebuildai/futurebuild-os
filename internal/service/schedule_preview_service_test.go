package service

import (
	"testing"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/models"
)

func newTestPreviewService() *SchedulePreviewService {
	return NewSchedulePreviewService(nil, config.PhysicsConfig{})
}

func TestSchedulePreview_Greenfield_SlabOneStory(t *testing.T) {
	svc := newTestPreviewService()
	req := SchedulePreviewRequest{
		SquareFootage:  2000,
		FoundationType: "slab",
		StartDate:      "2025-04-01",
		Stories:        1,
	}

	resp, err := svc.GeneratePreview(req)
	if err != nil {
		t.Fatalf("GeneratePreview failed: %v", err)
	}

	if resp.ProjectedEnd == "" {
		t.Error("expected projected end date")
	}
	if resp.TotalWorkingDays <= 0 {
		t.Error("expected positive working days")
	}
	if len(resp.CriticalPath) == 0 {
		t.Error("expected non-empty critical path")
	}
	if len(resp.PhaseTimeline) == 0 {
		t.Error("expected non-empty phase timeline")
	}
	if resp.GanttPreview == nil {
		t.Error("expected gantt preview")
	}
	if resp.CompletionPercent != 0 {
		t.Errorf("expected 0%% completion for greenfield, got %.2f%%", resp.CompletionPercent)
	}

	// Slab should have scope changes
	hasSlabChange := false
	for _, sc := range resp.ScopeChanges {
		if len(sc.TasksRemoved) > 0 {
			hasSlabChange = true
		}
	}
	if !hasSlabChange {
		t.Error("expected slab foundation scope change")
	}
}

func TestSchedulePreview_Greenfield_BasementTwoStory(t *testing.T) {
	svc := newTestPreviewService()
	req := SchedulePreviewRequest{
		SquareFootage:  4500,
		FoundationType: "basement",
		StartDate:      "2025-04-01",
		Stories:        2,
	}

	resp, err := svc.GeneratePreview(req)
	if err != nil {
		t.Fatalf("GeneratePreview failed: %v", err)
	}

	if resp.ProjectedEnd == "" {
		t.Error("expected projected end date")
	}

	// Basement should have more tasks → longer timeline
	if resp.TotalWorkingDays <= 0 {
		t.Error("expected positive working days for basement build")
	}
}

func TestSchedulePreview_InProgress(t *testing.T) {
	svc := newTestPreviewService()
	req := SchedulePreviewRequest{
		SquareFootage:  3000,
		FoundationType: "slab",
		StartDate:      "2025-01-15",
		Stories:        2,
		IsInProgress:   true,
		CurrentDate:    "2025-04-01",
		CompletedPhases: []CompletedPhaseInput{
			{WBSCode: "7.x", ActualEnd: "2025-02-15", Status: "completed"},
			{WBSCode: "8.x", ActualEnd: "2025-03-20", Status: "completed"},
		},
	}

	resp, err := svc.GeneratePreview(req)
	if err != nil {
		t.Fatalf("GeneratePreview failed: %v", err)
	}

	if resp.CompletionPercent <= 0 {
		t.Error("expected positive completion percent for in-progress project")
	}
	if resp.RemainingDays <= 0 {
		t.Error("expected positive remaining days")
	}

	// Completed phases should show as completed in timeline
	for _, phase := range resp.PhaseTimeline {
		if phase.WBSCode == "7.x" || phase.WBSCode == "8.x" {
			if phase.Status != "completed" {
				t.Errorf("expected phase %s to be completed, got %s", phase.WBSCode, phase.Status)
			}
		}
	}

	// Gantt should show completed tasks
	if resp.GanttPreview != nil {
		hasCompleted := false
		for _, task := range resp.GanttPreview.Tasks {
			if task.Status == "Completed" {
				hasCompleted = true
				break
			}
		}
		if !hasCompleted {
			t.Error("expected completed tasks in gantt preview")
		}
	}
}

func TestSchedulePreview_WithLongLeadItems(t *testing.T) {
	svc := newTestPreviewService()
	req := SchedulePreviewRequest{
		SquareFootage:  3200,
		FoundationType: "crawlspace",
		StartDate:      "2025-04-01",
		Stories:        2,
		LongLeadItems: []models.LongLeadItem{
			{
				Name:               "Marvin Ultimate Windows",
				Brand:              "marvin ultimate",
				Category:           "windows",
				EstimatedLeadWeeks: 14,
				WBSCode:            "9.5",
			},
			{
				Name:               "Sub-Zero Refrigerator",
				Brand:              "sub-zero",
				Category:           "appliances",
				EstimatedLeadWeeks: 10,
			},
		},
	}

	resp, err := svc.GeneratePreview(req)
	if err != nil {
		t.Fatalf("GeneratePreview failed: %v", err)
	}

	if len(resp.ProcurementDates) != 2 {
		t.Errorf("expected 2 procurement dates, got %d", len(resp.ProcurementDates))
	}

	for _, pd := range resp.ProcurementDates {
		if pd.OrderByDate == "" {
			t.Errorf("expected order by date for %s", pd.ItemName)
		}
		if pd.InstallDate == "" {
			t.Errorf("expected install date for %s", pd.ItemName)
		}
		if pd.Status == "" {
			t.Errorf("expected status for %s", pd.ItemName)
		}
	}
}

func TestSchedulePreview_WithoutWeather(t *testing.T) {
	svc := newTestPreviewService()
	req := SchedulePreviewRequest{
		SquareFootage:  2000,
		FoundationType: "slab",
		StartDate:      "2025-04-01",
		Stories:        1,
	}

	resp, err := svc.GeneratePreview(req)
	if err != nil {
		t.Fatalf("GeneratePreview failed: %v", err)
	}

	// No weather service → no weather impact
	if resp.WeatherImpact != nil {
		t.Error("expected no weather impact without weather service")
	}
}

func TestCompareScenarios_DifferentStartDates(t *testing.T) {
	svc := newTestPreviewService()
	req := ScenarioComparisonRequest{
		Base: SchedulePreviewRequest{
			SquareFootage:  2500,
			FoundationType: "slab",
			StartDate:      "2025-03-01",
			Stories:        1,
		},
		Alternatives: []SchedulePreviewRequest{
			{
				SquareFootage:  2500,
				FoundationType: "slab",
				StartDate:      "2025-04-01",
				Stories:        1,
			},
		},
	}

	resp, err := svc.CompareScenarios(req)
	if err != nil {
		t.Fatalf("CompareScenarios failed: %v", err)
	}

	if len(resp.Scenarios) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(resp.Scenarios))
	}

	// Alternative starts later, so should end later (positive delta)
	if resp.Scenarios[1].DeltaDays <= 0 {
		t.Errorf("expected positive delta for later start, got %d", resp.Scenarios[1].DeltaDays)
	}
}

func TestCompareScenarios_MaxAlternatives(t *testing.T) {
	svc := newTestPreviewService()
	req := ScenarioComparisonRequest{
		Base: SchedulePreviewRequest{
			SquareFootage:  2500,
			FoundationType: "slab",
			StartDate:      "2025-03-01",
			Stories:        1,
		},
		Alternatives: []SchedulePreviewRequest{
			{SquareFootage: 2500, FoundationType: "slab", StartDate: "2025-04-01", Stories: 1},
			{SquareFootage: 2500, FoundationType: "slab", StartDate: "2025-05-01", Stories: 1},
			{SquareFootage: 2500, FoundationType: "slab", StartDate: "2025-06-01", Stories: 1},
			{SquareFootage: 2500, FoundationType: "slab", StartDate: "2025-07-01", Stories: 1}, // 4th — over limit
		},
	}

	_, err := svc.CompareScenarios(req)
	if err == nil {
		t.Error("expected error for >3 alternatives")
	}
}
