// Package tribunal contains types and logic for The Tribunal consensus system.
// Step 64: Stub package - implementation in Step 65.
package tribunal

// DecisionStatus represents the outcome of a Tribunal decision.
type DecisionStatus string

const (
	DecisionPending  DecisionStatus = "pending"
	DecisionApproved DecisionStatus = "approved"
	DecisionRejected DecisionStatus = "rejected"
	DecisionConflict DecisionStatus = "conflict"
)
