package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCommunicationLogMarshaling(t *testing.T) {
	projectID := uuid.New()
	contactID := uuid.New()
	now := time.Now().UTC()

	log := CommunicationLog{
		ID:        uuid.New(),
		ProjectID: projectID,
		ContactID: &contactID,
		Direction: CommunicationDirectionOutbound,
		Content:   "Testing interaction history",
		Channel:   CommunicationChannelSMS,
		Timestamp: now,
	}

	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("Failed to marshal CommunicationLog: %v", err)
	}

	var decoded CommunicationLog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CommunicationLog: %v", err)
	}

	if decoded.ID != log.ID {
		t.Errorf("Expected ID %v, got %v", log.ID, decoded.ID)
	}
	if decoded.ProjectID != log.ProjectID {
		t.Errorf("Expected ProjectID %v, got %v", log.ProjectID, decoded.ProjectID)
	}
	if *decoded.ContactID != *log.ContactID {
		t.Errorf("Expected ContactID %v, got %v", *log.ContactID, *decoded.ContactID)
	}
	if decoded.Timestamp.Unix() != log.Timestamp.Unix() {
		t.Errorf("Expected Timestamp %v, got %v", log.Timestamp, decoded.Timestamp)
	}
}

func TestNotificationMarshaling(t *testing.T) {
	userID := uuid.New()

	note := Notification{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     NotificationTypeScheduleSlip,
		Priority: 1,
		Status:   NotificationStatusUnread,
	}

	data, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("Failed to marshal Notification: %v", err)
	}

	var decoded Notification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Notification: %v", err)
	}

	if decoded.ID != note.ID {
		t.Errorf("Expected ID %v, got %v", note.ID, decoded.ID)
	}
	if decoded.Type != note.Type {
		t.Errorf("Expected Type %v, got %v", note.Type, decoded.Type)
	}
	// Verify strict absence of fields not in spec (Marshaling works, but zero values would be present in struct if they existed)
}
