package worker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// AsynqHydrationEnqueuer implements service.HydrationEnqueuer using Asynq.
// P1 Performance Fix: Enables event-driven hydration on project creation.
type AsynqHydrationEnqueuer struct {
	client *asynq.Client
}

// NewAsynqHydrationEnqueuer creates a new hydration enqueuer.
func NewAsynqHydrationEnqueuer(redisAddr string) *AsynqHydrationEnqueuer {
	return &AsynqHydrationEnqueuer{
		client: asynq.NewClient(ParseRedisOpt(redisAddr)),
	}
}

// EnqueueHydration enqueues a project hydration task.
// Implements service.HydrationEnqueuer.
func (e *AsynqHydrationEnqueuer) EnqueueHydration(ctx context.Context, projectID uuid.UUID) error {
	task, err := NewHydrateProjectTask(projectID)
	if err != nil {
		return fmt.Errorf("create hydration task: %w", err)
	}

	_, err = e.client.EnqueueContext(ctx, task, asynq.Queue("default"))
	if err != nil {
		return fmt.Errorf("enqueue hydration task: %w", err)
	}
	return nil
}

// Close releases the underlying Asynq client resources.
func (e *AsynqHydrationEnqueuer) Close() error {
	return e.client.Close()
}
