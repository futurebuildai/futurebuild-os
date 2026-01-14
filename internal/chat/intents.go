package chat

import (
	"strings"

	"github.com/colton/futurebuild/pkg/types"
)

// Keyword is a strictly typed string for matching rules.
// See CTO Audit Step 43.2
type Keyword string

// keywordRule defines a set of keywords that map to a specific Intent.
type keywordRule struct {
	Intent   types.Intent
	Keywords []Keyword
}

// ClassifyIntent maps a raw user message to a strict Intent type based on keyword matching.
// Architecture: Linear scan of an ordered slice to guarantee deterministic results.
// Logic: Specific-to-Generic priority (e.g., "Schedule" checks before "Update").
// See PRODUCTION_PLAN.md Step 43.2
func ClassifyIntent(message string) types.Intent {
	// 1. Normalize Input
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return types.IntentUnknown
	}

	// 2. Define Ordered Rules (Specific -> Generic)
	// This slice preserves order, unlike a map iteration.
	rules := []keywordRule{
		// High Priority: Distinct Nouns / Specific Objects
		{
			Intent:   types.IntentProcessInvoice,
			Keywords: []Keyword{"invoice", "bill", "receipt"},
		},
		{
			Intent:   types.IntentExplainDelay,
			Keywords: []Keyword{"delay", "late", "behind", "slip"},
		},
		{
			Intent:   types.IntentGetSchedule,
			Keywords: []Keyword{"schedule", "timeline", "gantt", "when"},
		},
		// Low Priority: Generic Verbs / Actions
		// Checked last so "Update Schedule" catches "Schedule" (above) first.
		{
			Intent:   types.IntentUpdateTaskStatus,
			Keywords: []Keyword{"status", "complete", "finish", "done", "update"},
		},
	}

	// 3. Execute Linear Scan
	for _, rule := range rules {
		for _, keyword := range rule.Keywords {
			if strings.Contains(normalized, string(keyword)) {
				return rule.Intent
			}
		}
	}

	// 4. Default Fallback
	return types.IntentUnknown
}

