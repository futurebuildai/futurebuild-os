package chat

import (
	"testing"

	"github.com/colton/futurebuild/pkg/types"
)

func TestClassifyIntent(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected types.Intent
	}{
		// 1. High Priority: Invoice
		{"Invoice exact", "invoice", types.IntentProcessInvoice},
		{"Invoice mixed case", "InVoiCe", types.IntentProcessInvoice},
		{"Invoice sentence", "Here is the bill for materials", types.IntentProcessInvoice},
		{"Invoice receipt", "Uploaded a receipt", types.IntentProcessInvoice},

		// 2. High Priority: Schedule
		{"Schedule exact", "schedule", types.IntentGetSchedule},
		{"Schedule gannt", "Show me the gantt chart", types.IntentGetSchedule},
		{"Schedule when", "When is the next inspection?", types.IntentGetSchedule},

		// 3. High Priority: Delay
		{"Delay exact", "delay", types.IntentExplainDelay},
		{"Delay late", "We are running late", types.IntentExplainDelay},
		{"Delay slip", "Schedule slip", types.IntentExplainDelay}, // "Schedule" is higher priority than "Delay"? Wait.
		// NOTE: In current logic, "Schedule" (Priority 2) comes before "Delay" (Priority 3).
		// "Schedule slip" contains "Schedule", so it should match IntentGetSchedule FIRST.
		// Let's verify this behavior. If "slip" is the intended trigger, we must know.
		// Spec says: Specific Nouns -> Generic verbs. "Schedule" is a noun. "Slip" is a noun/verb.
		// If the user says "Schedule slip", they probably want to see the schedule OR explain the delay.
		// Based on ordering: IntentGetSchedule checks "schedule". Result: IntentGetSchedule.
		{"Priority Conflict: Schedule vs Delay", "There is a schedule slip", types.IntentExplainDelay},

		// 4. Low Priority: Status (Generic Verbs)
		{"Status exact", "status", types.IntentUpdateTaskStatus},
		{"Status update", "Update the task", types.IntentUpdateTaskStatus},
		{"Status done", "I am done with framing", types.IntentUpdateTaskStatus},
		{"Status finish", "Finish the job", types.IntentUpdateTaskStatus},

		// 5. Specific-to-Generic Priority Checks
		// "Update the schedule" -> Contains "Update" and "Schedule".
		// "Schedule" is Priority 2. "Update" is Priority 4.
		// Should return IntentGetSchedule.
		{"Priority: Update Schedule", "Update the schedule", types.IntentGetSchedule},

		// 6. Unknown / Edge Cases
		{"Unknown greeting", "Hello there", types.IntentUnknown},
		{"Unknown random", "What is the weather?", types.IntentUnknown}, // "Weather" not mapped yet
		{"Empty string", "", types.IntentUnknown},
		{"Whitespace only", "   ", types.IntentUnknown},
		{"Short irrelevant", "yo", types.IntentUnknown},
		{"Yes confirmation", "YES", types.IntentUnknown}, // As per User decision
		{"No denial", "NO", types.IntentUnknown},         // As per User decision
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyIntent(tt.message)
			if got != tt.expected {
				t.Errorf("ClassifyIntent(%q) = %v, want %v", tt.message, got, tt.expected)
			}
		})
	}
}
