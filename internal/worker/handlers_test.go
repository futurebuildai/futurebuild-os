package worker

import (
	"testing"

	"github.com/colton/futurebuild/internal/futureshade/tribunal"
	"github.com/stretchr/testify/assert"
)

func TestWorkerHandler_FormatPRComment(t *testing.T) {
	resp := &tribunal.TribunalResponse{
		Status:         tribunal.DecisionApproved,
		ConsensusScore: 0.95,
		Summary:        "Code looks excellent.",
		Plan:           "No remediation needed.",
	}

	comment := formatPRComment(resp)

	assert.Contains(t, comment, "## FutureBuild AI Review :white_check_mark:")
	assert.Contains(t, comment, "**Status**: APPROVED")
	assert.Contains(t, comment, "**Consensus Score**: 0.95")
	assert.Contains(t, comment, "Code looks excellent.")
}

func TestSanitizePRDiff(t *testing.T) {
	diff := "line1\n---\nline2\n===\nline3"
	sanitized := sanitizePRDiff(diff)

	assert.Contains(t, sanitized, "[separator]")
	assert.NotContains(t, sanitized, "\n---\n")
	assert.NotContains(t, sanitized, "\n===\n")
}
