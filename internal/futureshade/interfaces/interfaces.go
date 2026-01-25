// Package interfaces defines service contracts for The Tribunal consensus system.
// See specs/TRIBUNAL_INTERFACES_specs.md for full specification.
package interfaces

import (
	"context"

	"github.com/colton/futurebuild/internal/futureshade/types"
)

// Juror represents a single LLM provider that can deliberate on a Case.
// Implementations must support context cancellation for timeout handling.
// Future implementations should include retry logic with exponential backoff
// for 5xx errors (do not retry 4xx errors).
type Juror interface {
	// Consult submits a Case to the Juror and returns their Verdict.
	// The context should be used for timeout and cancellation.
	Consult(ctx context.Context, c types.Case) (types.Verdict, error)

	// ID returns the ModelID of this Juror.
	ID() types.ModelID
}

// TheGavel decides the final outcome based on collected Verdicts.
// It implements the consensus logic (Unanimous, Majority, Supervisor).
type TheGavel interface {
	// Deliberate synthesizes multiple Verdicts into a single TribunalDecision.
	// The strategy parameter determines how consensus is evaluated.
	Deliberate(strategy types.ConsensusStrategy, verdicts []types.Verdict) (types.TribunalDecision, error)
}

// TribunalService is the public API for submitting Cases to The Tribunal.
// It orchestrates Juror consultations and Gavel deliberation.
//
// FAIL-SAFE REQUIREMENT: Implementations MUST fail closed (return error)
// if ShadowDB (persistence layer) is unavailable. No un-audited decisions
// may be made.
type TribunalService interface {
	// Adjudicate submits a Case for judgment and returns the final Decision.
	// The context should be used for overall timeout control.
	// Returns an error if persistence is unavailable (fail closed).
	Adjudicate(ctx context.Context, c types.Case) (types.TribunalDecision, error)
}
