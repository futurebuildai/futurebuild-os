// Package db provides database infrastructure utilities.
package db

import (
	"context"

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
