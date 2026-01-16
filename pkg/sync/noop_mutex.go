// Package sync provides synchronization primitives for distributed systems.
package sync

import (
	"context"
	"time"
)

// NoOpMutex always succeeds when acquiring locks.
// Use for single-instance deployments, testing, or when distributed locking is not required.
//
// L7 Standard: This is a "zero-value safe" implementation - inject this when
// you need the interface but don't require actual distributed locking.
type NoOpMutex struct{}

// TryLock always succeeds and returns a no-op unlock function.
func (n *NoOpMutex) TryLock(ctx context.Context, key string, ttl time.Duration) (UnlockFunc, error) {
	return func() error { return nil }, nil
}

// Ensure NoOpMutex implements DistributedMutex at compile time.
var _ DistributedMutex = (*NoOpMutex)(nil)
