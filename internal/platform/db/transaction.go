// Package db provides database infrastructure utilities.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// txKey is the context key for distributed transaction propagation.
// Uses empty struct for zero-allocation key type.
type txKey struct{}

// InjectTx adds a pgx.Tx to the context for distributed transaction propagation.
// The caller is responsible for Commit/Rollback lifecycle management.
// See PRODUCTION_PLAN.md Step 45 (Zombie Write Fix)
func InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx retrieves a pgx.Tx from the context if present.
// Returns (tx, true) if a transaction was injected, (nil, false) otherwise.
// See PRODUCTION_PLAN.md Step 45 (Zombie Write Fix)
func ExtractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

// Transactor defines the interface for beginning transactions.
// *pgxpool.Pool satisfies this interface.
type Transactor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// RunInTx executes fn within a transaction.
// Standardizes Begin, Defer Rollback, Commit pattern to prevent copy-paste errors.
// L7 Code Review: Transaction Ergonomics fix.
func RunInTx(ctx context.Context, pool Transactor, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer rollback - no-op if already committed
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
