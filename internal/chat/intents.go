package chat

import (
	"regexp"
	"sort"
	"strings"

	"github.com/colton/futurebuild/pkg/types"
)

// IntentClassifier defines the interface for intent classification.
// Implementations must provide a Classify method that maps a raw message to an Intent.
// See PRODUCTION_PLAN.md Step 43.2 (Weighted Regex Router Refactor)
type IntentClassifier interface {
	Classify(message string) types.Intent
}

// IntentRule defines a weighted regex pattern for intent classification.
type IntentRule struct {
	Intent  types.Intent
	Pattern *regexp.Regexp
	Weight  int // Higher weight = more specific match
}

// RegexClassifier implements IntentClassifier using weighted regex patterns.
// It accumulates scores for each intent based on matching patterns and
// returns the intent with the highest score.
type RegexClassifier struct {
	rules []IntentRule
}

// NewRegexClassifier creates a RegexClassifier with custom rules.
func NewRegexClassifier(rules []IntentRule) *RegexClassifier {
	return &RegexClassifier{rules: rules}
}

// =============================================================================
// RULE ORDERING AND SHADOWING PREVENTION
// =============================================================================
//
// CRITICAL: Intent classification uses a WEIGHTED SCORING system, not first-match.
// However, the rule definitions below follow a deliberate ordering principle to
// prevent "shadowing" bugs where generic patterns could outweigh specific ones.
//
// ORDERING PRINCIPLE: Specific-to-Generic
//
// 1. ACTION INTENTS (Weight: 10+) - Patterns that MODIFY state
//    These require an action verb + domain noun combination.
//    Examples: "update schedule", "process invoice", "mark complete"
//
// 2. QUERY INTENTS (Weight: 5) - Patterns that READ state
//    These match on domain nouns without requiring action verbs.
//    Examples: "show me the schedule", "what is the delay"
//
// WHY THIS MATTERS:
// Input "Update Schedule" contains both:
//   - "update" (action verb) → should trigger state modification
//   - "schedule" (domain noun) → could trigger schedule query
//
// By giving ACTION patterns a higher weight (10) than QUERY patterns (5),
// the action intent wins when both patterns match.
//
// ADDING NEW RULES:
// - If your rule MODIFIES state: Use Weight >= 10
// - If your rule READS state: Use Weight <= 5
// - Always add tests for phrases that could match multiple patterns
//
// =============================================================================

// NewDefaultRegexClassifier creates a RegexClassifier with the standard production rules.
// Rules are designed to prevent false positives and implement Specific-to-Generic matching.
// See PRODUCTION_PLAN.md Step 43.2
func NewDefaultRegexClassifier() *RegexClassifier {
	return &RegexClassifier{
		rules: []IntentRule{
			// =================================================================
			// ACTION INTENTS (Weight: 10) - State-modifying operations
			// These patterns require both an ACTION VERB and a DOMAIN NOUN.
			// =================================================================

			// ProcessInvoice: Requires action verb + invoice noun (Weight 10)
			// Matches: "process invoice", "upload bill", "scan receipt"
			// Does NOT match: "invoice for delay" (no action verb before noun)
			{
				Intent:  types.IntentProcessInvoice,
				Pattern: regexp.MustCompile(`(?i)\b(upload|process|scan|check|submit).*(invoice|bill|receipt)`),
				Weight:  10,
			},

			// UpdateTaskStatus: Requires action verb + status/task noun (Weight 10)
			// Matches: "update schedule", "mark task complete", "set as done", "change status"
			// The pattern now includes "schedule" and "task" to capture modification intent.
			{
				Intent:  types.IntentUpdateTaskStatus,
				Pattern: regexp.MustCompile(`(?i)\b(mark|set|update|change)\b.*(status|complete|done|schedule|task|progress)`),
				Weight:  10,
			},

			// =================================================================
			// QUERY INTENTS (Weight: 5) - Read-only operations
			// These patterns match on DOMAIN NOUNS without requiring action verbs.
			// Lower weight ensures they don't shadow action intents.
			// =================================================================

			// ExplainDelay: Generic delay-related keywords (Weight 5)
			// Matches: "why is it delayed", "running late", "behind schedule", "slipped"
			// Uses word-start boundary only to match stems with suffixes (delayed, slipping, waiting)
			{
				Intent:  types.IntentExplainDelay,
				Pattern: regexp.MustCompile(`(?i)\b(delay|late|slip|behind|wait)`),
				Weight:  5,
			},

			// GetSchedule: Generic schedule-related keywords (Weight 5)
			// Matches: "show schedule", "project timeline", "when is..."
			// NOTE: "schedule" alone triggers this, but "update schedule" triggers
			// UpdateTaskStatus due to higher weight from action verb match.
			{
				Intent:  types.IntentGetSchedule,
				Pattern: regexp.MustCompile(`(?i)\b(schedule|timeline|gantt|when|date)\b`),
				Weight:  5,
			},

			// =================================================================
			// IMPLICIT ACTION INTENTS (Weight: 3) - Noun-only shorthand
			// Construction workers are curt. "invoice" implies "process this invoice".
			// Weight 3 ensures explicit action verbs (Weight 10) still take priority.
			// See Step 2: Regression Verification (Invoice Shorthand)
			// =================================================================

			// ProcessInvoice (Shorthand): Noun-only invoice/bill/receipt (Weight 3)
			// Matches: "invoice", "bill", "receipt" without action verb
			// Lower weight than action patterns but triggers invoice processing.
			{
				Intent:  types.IntentProcessInvoice,
				Pattern: regexp.MustCompile(`(?i)\b(invoice|bill|receipt)\b`),
				Weight:  3,
			},

			// ContactSubcontractor: Communicate with trade partners (Weight 10)
			// Matches: "ask the plumber", "contact electrician", "message the sub"
			// See PRODUCTION_PLAN.md Step 47 (Sub Liaison Agent)
			{
				Intent:  types.IntentContactSubcontractor,
				Pattern: regexp.MustCompile(`(?i)\b(ask|contact|check with|message|reach out to|call).*(plumber|electrician|sub|trade|contractor|hvac|drywall|painter|roofer|framer)`),
				Weight:  10,
			},
		},
	}
}

// Classify implements IntentClassifier by scoring all matching patterns
// and returning the intent with the highest accumulated score.
//
// Normalization: The input is trimmed of leading/trailing whitespace
// before processing. Case-insensitive matching is handled by regex flags.
func (c *RegexClassifier) Classify(message string) types.Intent {
	// Normalize input: trim whitespace exactly once
	normalized := strings.TrimSpace(message)
	if normalized == "" {
		return types.IntentUnknown
	}

	// Score accumulator: intent -> total weight
	scores := make(map[types.Intent]int)

	for _, rule := range c.rules {
		if rule.Pattern.MatchString(normalized) {
			scores[rule.Intent] += rule.Weight
		}
	}

	// Find intent with highest score using deterministic tie-breaking.
	// ENGINEERING STANDARD: Map iteration order is non-deterministic in Go.
	// We sort intents by score (descending), then alphabetically by intent string
	// to ensure reproducible results when multiple intents have equal scores.
	if len(scores) == 0 {
		return types.IntentUnknown
	}

	// Collect intents into a slice for deterministic sorting
	type intentScore struct {
		intent types.Intent
		score  int
	}
	candidates := make([]intentScore, 0, len(scores))
	for intent, score := range scores {
		candidates = append(candidates, intentScore{intent: intent, score: score})
	}

	// Sort: highest score first, then alphabetically by intent for tie-breaking
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score // Higher score wins
		}
		return string(candidates[i].intent) < string(candidates[j].intent) // Alphabetical tie-breaker
	})

	return candidates[0].intent
}
