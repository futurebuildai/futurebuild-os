package chat

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
)

// =============================================================================
// PERSISTENCE STRATEGY PATTERN
// =============================================================================
// Implements the Strategy design pattern for different consistency guarantees.
// This decouples persistence logic from command execution (SRP compliance).
//
// See PRODUCTION_PLAN.md Orchestrator SRP Refactoring
// =============================================================================

// PersistenceStrategy defines the interface for different consistency guarantees.
// Each strategy encapsulates decision logic for database failures.
type PersistenceStrategy interface {
	// Persist attempts to save the message according to the strategy's consistency level.
	// Returns nil on success (including fallback success for BestEffort).
	// Returns error when the strategy's consistency requirements cannot be met.
	Persist(ctx context.Context, msg models.ChatMessage) error
}

// =============================================================================
// STRICT PERSISTENCE (Lane B)
// =============================================================================

// StrictPersistence fails if the DB write fails.
// Used for fast, internal operations where retry is safe and strict consistency is required.
// Maps to existing Lane B behavior.
type StrictPersistence struct {
	db MessagePersister
}

// NewStrictPersistence creates a new StrictPersistence strategy.
func NewStrictPersistence(db MessagePersister) *StrictPersistence {
	return &StrictPersistence{db: db}
}

// Persist saves the message or returns an error.
func (s *StrictPersistence) Persist(ctx context.Context, msg models.ChatMessage) error {
	if err := s.db.SaveMessage(ctx, msg); err != nil {
		slog.Error("chat: model persistence failed (strict mode)",
			"message_id", msg.ID,
			"project_id", msg.ProjectID,
			"error", err,
		)
		return fmt.Errorf("failed to persist model message: %w", err)
	}
	return nil
}

// =============================================================================
// BEST EFFORT PERSISTENCE (Lane A)
// =============================================================================

// BestEffortPersistence tries DB → DLQ → WAL fallback chain.
// Used for slow, external operations (AI) where the action already succeeded.
// Maps to existing Lane A behavior.
type BestEffortPersistence struct {
	db             MessagePersister
	dlq            DLQPersister
	wal            AuditWAL
	circuitBreaker AuditCircuitBreaker
}

// NewBestEffortPersistence creates a new BestEffortPersistence strategy.
func NewBestEffortPersistence(
	db MessagePersister,
	dlq DLQPersister,
	wal AuditWAL,
	circuitBreaker AuditCircuitBreaker,
) *BestEffortPersistence {
	return &BestEffortPersistence{
		db:             db,
		dlq:            dlq,
		wal:            wal,
		circuitBreaker: circuitBreaker,
	}
}

// Persist implements the DB → DLQ → WAL fallback chain.
// Returns nil if any fallback succeeds (best effort achieved).
// Returns error only when all fallbacks fail (audit system unavailable).
func (s *BestEffortPersistence) Persist(ctx context.Context, msg models.ChatMessage) error {
	// Happy path: DB write succeeds
	if err := s.db.SaveMessage(ctx, msg); err == nil {
		if s.circuitBreaker != nil {
			s.circuitBreaker.RecordSuccess()
		}
		return nil
	}

	// DB write failed - check circuit breaker first
	if s.circuitBreaker != nil && s.circuitBreaker.IsOpen() {
		slog.Error("AUDIT SYSTEM UNAVAILABLE: Degrading to read-only mode",
			"message_id", msg.ID,
			"project_id", msg.ProjectID,
		)
		return fmt.Errorf("system temporarily unavailable: audit system degraded")
	}

	// Log the original failure
	slog.Error("CRITICAL: Action succeeded but chat history save failed",
		"message_id", msg.ID,
		"project_id", msg.ProjectID,
	)

	// Fallback 1: Try DLQ (Redis-backed async retry)
	if dlqErr := s.dlq.EnqueueRetry(ctx, msg); dlqErr == nil {
		slog.Info("Message enqueued to DLQ for retry",
			"message_id", msg.ID,
		)
		if s.circuitBreaker != nil {
			s.circuitBreaker.RecordSuccess()
		}
		return nil // DLQ succeeded - best effort achieved
	} else {
		slog.Error("CRITICAL COMPLIANCE FAILURE: DLQ enqueue failed - trying WAL",
			"message_id", msg.ID,
			"error", dlqErr,
		)
	}

	// Fallback 2: Try local WAL (disk-based last resort)
	if s.wal != nil {
		if walErr := s.wal.AppendRecord(ctx, msg); walErr == nil {
			slog.Warn("Audit record written to local WAL for recovery",
				"message_id", msg.ID,
			)
			if s.circuitBreaker != nil {
				s.circuitBreaker.RecordSuccess()
			}
			return nil // WAL succeeded - best effort achieved
		} else {
			slog.Error("CATASTROPHIC: WAL write also failed",
				"message_id", msg.ID,
				"error", walErr,
			)
		}
	}

	// All audit systems failed - trip circuit breaker and return error
	if s.circuitBreaker != nil {
		s.circuitBreaker.RecordFailure()
	}
	return fmt.Errorf("audit system unavailable: please try again later")
}

// =============================================================================
// STRATEGY REGISTRY
// =============================================================================

// PersistenceStrategyRegistry maps consistency types to strategies.
type PersistenceStrategyRegistry struct {
	strategies map[types.ConsistencyType]PersistenceStrategy
}

// NewPersistenceStrategyRegistry creates a registry with default strategies.
func NewPersistenceStrategyRegistry(
	db MessagePersister,
	dlq DLQPersister,
	wal AuditWAL,
	circuitBreaker AuditCircuitBreaker,
) *PersistenceStrategyRegistry {
	return &PersistenceStrategyRegistry{
		strategies: map[types.ConsistencyType]PersistenceStrategy{
			types.ConsistencyStrict:     NewStrictPersistence(db),
			types.ConsistencyBestEffort: NewBestEffortPersistence(db, dlq, wal, circuitBreaker),
		},
	}
}

// Get returns the strategy for the given consistency type.
// Returns StrictPersistence as default if type is unknown.
func (r *PersistenceStrategyRegistry) Get(ct types.ConsistencyType) PersistenceStrategy {
	if strategy, ok := r.strategies[ct]; ok {
		return strategy
	}
	// Default to strict for unknown types (fail-safe)
	return r.strategies[types.ConsistencyStrict]
}

// Register adds a new strategy for a consistency type.
// This enables Open-Closed Principle: add new Lane C without modifying existing code.
func (r *PersistenceStrategyRegistry) Register(ct types.ConsistencyType, strategy PersistenceStrategy) {
	r.strategies[ct] = strategy
}
