package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AgentPendingActionStatus represents the lifecycle state of a pending action.
type AgentPendingActionStatus string

const (
	PendingActionStatusPending  AgentPendingActionStatus = "pending"
	PendingActionStatusApproved AgentPendingActionStatus = "approved"
	PendingActionStatusRejected AgentPendingActionStatus = "rejected"
	PendingActionStatusExpired  AgentPendingActionStatus = "expired"
)

// AgentPendingAction represents an agent-recommended action awaiting human approval.
// Created when an agent calls create_approval_card. Linked 1:1 with a feed card.
type AgentPendingAction struct {
	ID              uuid.UUID                `json:"id" db:"id"`
	OrgID           uuid.UUID                `json:"org_id" db:"org_id"`
	ProjectID       uuid.UUID                `json:"project_id" db:"project_id"`
	FeedCardID      uuid.UUID                `json:"feed_card_id" db:"feed_card_id"`
	AgentSource     string                   `json:"agent_source" db:"agent_source"`
	ActionType      string                   `json:"action_type" db:"action_type"`
	ActionPayload   json.RawMessage          `json:"action_payload" db:"action_payload"`
	Status          AgentPendingActionStatus `json:"status" db:"status"`
	ApprovedBy      *uuid.UUID               `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt      *time.Time               `json:"approved_at,omitempty" db:"approved_at"`
	RejectionReason *string                  `json:"rejection_reason,omitempty" db:"rejection_reason"`
	ExpiresAt       *time.Time               `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt       time.Time                `json:"created_at" db:"created_at"`

	// Denormalized from feed_card JOIN (for API response)
	Headline    string `json:"headline,omitempty" db:"headline"`
	ProjectName string `json:"project_name,omitempty" db:"project_name"`
}
