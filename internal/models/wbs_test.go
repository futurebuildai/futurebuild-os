package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWBSTemplate_JSON(t *testing.T) {
	tmpl := WBSTemplate{
		ID:            uuid.New(),
		Name:          "Default Template",
		Version:       "1.2.0",
		IsDefault:     true,
		EntryPointWBS: "5.2",
		CreatedAt:     time.Now(),
	}

	data, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatalf("Failed to marshal WBSTemplate: %v", err)
	}

	var decoded WBSTemplate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal WBSTemplate: %v", err)
	}

	if decoded.ID != tmpl.ID || decoded.Name != tmpl.Name {
		t.Errorf("Decoded template does not match original")
	}
}

func TestWBSTask_JSON(t *testing.T) {
	task := WBSTask{
		ID:               uuid.New(),
		PhaseID:          uuid.New(),
		Code:             "9.1",
		Name:             "Framing",
		BaseDurationDays: 10.5,
		PredecessorCodes: []string{"8.1", "8.2"},
		CreatedAt:        time.Now(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal WBSTask: %v", err)
	}

	var decoded WBSTask
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal WBSTask: %v", err)
	}

	if decoded.Code != task.Code || len(decoded.PredecessorCodes) != 2 {
		t.Errorf("Decoded task does not match original")
	}
}
