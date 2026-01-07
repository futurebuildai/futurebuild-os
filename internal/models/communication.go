package models

import (
	"time"

	"github.com/google/uuid"
)

// See DATA_SPINE_SPEC.md Section 5.1
type CommunicationDirection string

const (
	CommunicationDirectionInbound  CommunicationDirection = "Inbound"
	CommunicationDirectionOutbound CommunicationDirection = "Outbound"
)

type CommunicationChannel string

const (
	CommunicationChannelSMS   CommunicationChannel = "SMS"
	CommunicationChannelChat  CommunicationChannel = "Chat"
	CommunicationChannelEmail CommunicationChannel = "Email"
)

// CommunicationLog represents the history of interaction.
// See DATA_SPINE_SPEC.md Section 5.1 (Revised with Logical fix for User logging)
type CommunicationLog struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	ProjectID uuid.UUID              `json:"project_id" db:"project_id" validate:"required"`
	ContactID *uuid.UUID             `json:"contact_id,omitempty" db:"contact_id"`
	Direction CommunicationDirection `json:"direction" db:"direction" validate:"required"`
	Content   string                 `json:"content" db:"content" validate:"required"`
	Channel   CommunicationChannel   `json:"channel" db:"channel" validate:"required"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// See DATA_SPINE_SPEC.md Section 5.2
type NotificationType string

const (
	NotificationTypeScheduleSlip  NotificationType = "Schedule_Slip"
	NotificationTypeInvoiceReady  NotificationType = "Invoice_Ready"
	NotificationTypeAssignmentNew NotificationType = "Assignment_New"
	NotificationTypeDailyBriefing NotificationType = "Daily_Briefing"
)

type NotificationStatus string

const (
	NotificationStatusUnread    NotificationStatus = "Unread"
	NotificationStatusRead      NotificationStatus = "Read"
	NotificationStatusDismissed NotificationStatus = "Dismissed"
)

// Notification represents a system alert.
// See DATA_SPINE_SPEC.md Section 5.2 (Strict Parity: No timestamps)
type Notification struct {
	ID       uuid.UUID          `json:"id" db:"id"`
	UserID   uuid.UUID          `json:"user_id" db:"user_id" validate:"required"`
	Type     NotificationType   `json:"type" db:"type" validate:"required"`
	Priority int                `json:"priority" db:"priority"`
	Status   NotificationStatus `json:"status" db:"status"`
}
