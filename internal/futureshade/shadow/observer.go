package shadow

import (
	"context"
)

// EventType categorizes system events for observation.
type EventType string

const (
	EventPRCreated      EventType = "pr_created"
	EventPRUpdated      EventType = "pr_updated"
	EventTaskCompleted  EventType = "task_completed"
	EventDocumentAdded  EventType = "document_added"
	EventScheduleChange EventType = "schedule_change"
)

// Event represents a system event observed by the Shadow layer.
type Event struct {
	Type     EventType              `json:"type"`
	SourceID string                 `json:"source_id"`
	Payload  map[string]interface{} `json:"payload"`
}

// Observer defines the interface for the Shadow observation system.
// The Shadow layer observes system events without modifying them,
// forwarding relevant events to the Tribunal for analysis when needed.
type Observer interface {
	// Observe processes a system event.
	// Returns nil if the event was processed successfully.
	Observe(ctx context.Context, event Event) error

	// Start begins the observer's background processes.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the observer.
	Stop(ctx context.Context) error
}

// observer implements the Observer interface.
type observer struct {
	enabled bool
}

// NewObserver creates a new Shadow observer.
func NewObserver(enabled bool) Observer {
	return &observer{enabled: enabled}
}

// Observe processes a system event.
// This is a stub implementation for Step 64.
func (o *observer) Observe(ctx context.Context, event Event) error {
	if !o.enabled {
		return nil
	}
	// TODO: Step 65+ will implement actual observation logic
	return nil
}

// Start begins the observer's background processes.
// This is a stub implementation for Step 64.
func (o *observer) Start(ctx context.Context) error {
	// TODO: Step 65+ will implement event subscription
	return nil
}

// Stop gracefully shuts down the observer.
func (o *observer) Stop(ctx context.Context) error {
	return nil
}
