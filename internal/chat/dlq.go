package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/worker"
	"github.com/hibiken/asynq"
)

// TypeMessageRetry is the Asynq task type for chat message retry.
const TypeMessageRetry = "chat:message_retry"

// DLQPersister defines the interface for dead letter queue operations.
// Enables async retry of failed chat message writes.
// See PRODUCTION_PLAN.md Critical Blocker C Remediation
type DLQPersister interface {
	EnqueueRetry(ctx context.Context, msg models.ChatMessage) error
}

// AsynqDLQ implements DLQPersister using Asynq for Redis-backed task queue.
type AsynqDLQ struct {
	client *asynq.Client
}

// NewAsynqDLQ creates a new Asynq-based Dead Letter Queue.
// redisAddr should match the worker's Redis connection (e.g., "localhost:6379").
func NewAsynqDLQ(redisAddr string) *AsynqDLQ {
	return &AsynqDLQ{
		client: asynq.NewClient(worker.ParseRedisOpt(redisAddr)),
	}
}

// EnqueueRetry adds a failed chat message to the retry queue.
// The message will be retried up to 3 times with exponential backoff.
func (d *AsynqDLQ) EnqueueRetry(ctx context.Context, msg models.ChatMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message for DLQ: %w", err)
	}

	task := asynq.NewTask(TypeMessageRetry, payload,
		asynq.MaxRetry(3),
		asynq.Queue("critical"),
	)

	_, err = d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue message to DLQ: %w", err)
	}

	return nil
}

// Close releases the underlying Asynq client resources.
func (d *AsynqDLQ) Close() error {
	return d.client.Close()
}
