package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// NotificationEnqueuer defines the interface for async notification delivery.
// P1 Performance Fix: Enables sidecar pattern for procurement notifications.
type NotificationEnqueuer interface {
	EnqueueNotification(ctx context.Context, itemID uuid.UUID, message string, ts time.Time) error
	Close() error
}

// AsynqNotificationEnqueuer implements NotificationEnqueuer using Asynq.
// Follows the pattern established by AsynqHydrationEnqueuer.
type AsynqNotificationEnqueuer struct {
	client *asynq.Client
}

// NewAsynqNotificationEnqueuer creates a new notification enqueuer.
func NewAsynqNotificationEnqueuer(redisAddr string) *AsynqNotificationEnqueuer {
	return &AsynqNotificationEnqueuer{
		client: asynq.NewClient(ParseRedisOpt(redisAddr)),
	}
}

// EnqueueNotification enqueues a procurement notification for async delivery.
// Implements NotificationEnqueuer.
func (e *AsynqNotificationEnqueuer) EnqueueNotification(ctx context.Context, itemID uuid.UUID, message string, ts time.Time) error {
	task, err := NewProcurementNotificationTask(itemID, message, ts)
	if err != nil {
		return fmt.Errorf("create notification task: %w", err)
	}

	_, err = e.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueue notification task: %w", err)
	}
	return nil
}

// Close releases the underlying Asynq client resources.
func (e *AsynqNotificationEnqueuer) Close() error {
	return e.client.Close()
}

// NoOpNotificationEnqueuer is a no-op implementation for testing or when async is disabled.
type NoOpNotificationEnqueuer struct{}

// EnqueueNotification is a no-op that always succeeds.
func (e *NoOpNotificationEnqueuer) EnqueueNotification(ctx context.Context, itemID uuid.UUID, message string, ts time.Time) error {
	return nil
}

// Close is a no-op.
func (e *NoOpNotificationEnqueuer) Close() error {
	return nil
}
