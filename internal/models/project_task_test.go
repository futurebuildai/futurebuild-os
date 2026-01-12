package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

func TestProjectTaskMarshaling(t *testing.T) {
	projectID := uuid.New()
	now := time.Now().Truncate(time.Second).UTC()
	override := 2.5

	task := ProjectTask{
		ID:                      uuid.New(),
		ProjectID:               projectID,
		WBSCode:                 "9.3",
		Name:                    "Rough Plumbing",
		EarlyStart:              &now,
		EarlyFinish:             &now,
		CalculatedDuration:      5.0,
		WeatherAdjustedDuration: 6.5,
		ManualOverrideDays:      &override,
		Status:                  types.TaskStatusPending,
		VerifiedByVision:        true,
		VerificationConfidence:  0.95,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal ProjectTask: %v", err)
	}

	var decoded ProjectTask
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ProjectTask: %v", err)
	}

	if decoded.ID != task.ID {
		t.Errorf("Expected ID %v, got %v", task.ID, decoded.ID)
	}
}

func TestProjectAssignmentMarshaling(t *testing.T) {
	projectID := uuid.New()
	contactID := uuid.New()
	phaseID := "9.x"

	assign := ProjectAssignment{
		ID:         uuid.New(),
		ProjectID:  projectID,
		ContactID:  contactID,
		WBSPhaseID: phaseID,
	}

	data, err := json.Marshal(assign)
	if err != nil {
		t.Fatalf("Failed to marshal ProjectAssignment: %v", err)
	}

	var decoded ProjectAssignment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ProjectAssignment: %v", err)
	}

	if decoded.WBSPhaseID != assign.WBSPhaseID {
		t.Errorf("Expected WBSPhaseID %v, got %v", assign.WBSPhaseID, decoded.WBSPhaseID)
	}
}
