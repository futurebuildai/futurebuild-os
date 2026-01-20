// Package sync provides synchronization primitives for distributed systems.
// See L7 Technical Debt Remediation: Task 2 (Distributed Locking)
package sync

import (
	"context"
	"errors"
	"time"
)

// ErrLockHeld is returned when another process holds the lock.
// This is a sentinel error that callers should check with errors.Is().
var ErrLockHeld = errors.New("lock held by another process")

// UnlockFunc releases a distributed lock.
// Must be called exactly once after acquiring a lock.
// Calling unlock multiple times or after TTL expiry is a no-op.
type UnlockFunc func() error

// DistributedMutex provides cross-process locking for distributed systems.
// Implementations may use Redis, PostgreSQL advisory locks, etcd, etc.
//
// L7 Standard: This interface enables:
// - Preventing duplicate execution in Blue/Green deployments
// - Protecting critical sections across application replicas
// - TTL-based auto-expiry to prevent deadlocks from crashed processes
//
// Usage:
//
//	unlock, err := mutex.TryLock(ctx, "my:lock:key", 5*time.Minute)
//	if errors.Is(err, sync.ErrLockHeld) {
//	    // Another instance is running, skip gracefully
//	    return nil
//	}
//	if err != nil {
//	    return fmt.Errorf("acquiring lock: %w", err)
//	}
//	defer unlock()  // SAFETY: Always release lock
type DistributedMutex interface {
	// TryLock attempts to acquire a lock with the given key and TTL.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - key: Unique identifier for the lock (e.g., "futurebuild:agent:procurement:lock")
	//   - ttl: Time-to-live for the lock (must auto-expire to prevent deadlocks)
	//
	// Returns:
	//   - UnlockFunc: Function to release the lock (must be called exactly once)
	//   - error: ErrLockHeld if lock is held by another process, or other errors
	//
	// The lock MUST auto-expire after TTL to prevent deadlocks from crashed processes.
	TryLock(ctx context.Context, key string, ttl time.Duration) (UnlockFunc, error)

	// ExtendLock extends the TTL of an existing lock.
	// Must be called by the process holding the lock (heartbeat).
	ExtendLock(ctx context.Context, key string, ttl time.Duration) error
}
