// Package types provides shared domain types for FutureBuild.
// WBSCode value object provides robust Work Breakdown Structure code parsing.
// See PRODUCTION_PLAN.md Technical Debt Remediation (P2) Section B.
package types

import (
	"errors"
	"regexp"
	"strings"
)

// --- Package-Level Compiled Regexes ---
// ENGINEERING STANDARD: Compile once at startup, not per-call.
var (
	// csiDotFormatRe matches CSI standard formats: "9", "9.1", "9.1.2", "14.3.1.5"
	csiDotFormatRe = regexp.MustCompile(`^(\d+)(?:\.\d+)*$`)

	// alphaDashFormatRe matches custom alpha formats: "A-100", "B-200-1"
	alphaDashFormatRe = regexp.MustCompile(`^([A-Za-z]+)[- ](\d+)`)

	// simpleAlphaRe matches single letters: "A", "B", "C"
	simpleAlphaRe = regexp.MustCompile(`^[A-Za-z]+$`)
)

// ErrInvalidWBSCode is returned when a WBS code cannot be parsed.
var ErrInvalidWBSCode = errors.New("invalid WBS code format")

// WBSCode represents a validated Work Breakdown Structure code.
// This is a value object that handles various WBS code formats:
//   - CSI standard: "9.1.2" → phase "9"
//   - CSI major only: "14.3" → phase "14"
//   - Custom alpha-numeric: "A-100" → phase "A"
//   - Simple numeric: "9" → phase "9"
type WBSCode struct {
	raw        string
	majorPhase string
}

// ParseWBSCode creates a WBSCode from a string, validating format.
// Returns ErrInvalidWBSCode if the format cannot be parsed.
func ParseWBSCode(s string) (WBSCode, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return WBSCode{}, ErrInvalidWBSCode
	}

	// Try CSI dot format first (most common): "9.1.2", "14.3", "9"
	if matches := csiDotFormatRe.FindStringSubmatch(s); matches != nil {
		return WBSCode{
			raw:        s,
			majorPhase: matches[1],
		}, nil
	}

	// Try alpha-dash format: "A-100", "B-200-1"
	if matches := alphaDashFormatRe.FindStringSubmatch(s); matches != nil {
		return WBSCode{
			raw:        s,
			majorPhase: strings.ToUpper(matches[1]),
		}, nil
	}

	// Try simple alpha format: "A", "B", "MEP"
	if simpleAlphaRe.MatchString(s) {
		return WBSCode{
			raw:        s,
			majorPhase: strings.ToUpper(s),
		}, nil
	}

	// Fallback: Use the first segment before any delimiter
	// This handles edge cases like "9-1-2" or "A.100"
	separators := []string{".", "-", "_", " "}
	for _, sep := range separators {
		if idx := strings.Index(s, sep); idx > 0 {
			phase := s[:idx]
			// Normalize alpha phases to uppercase
			if simpleAlphaRe.MatchString(phase) {
				phase = strings.ToUpper(phase)
			}
			return WBSCode{
				raw:        s,
				majorPhase: phase,
			}, nil
		}
	}

	// Final fallback: treat the entire input as the major phase
	return WBSCode{
		raw:        s,
		majorPhase: strings.ToUpper(s),
	}, nil
}

// MustParseWBSCode panics on invalid format.
// Use only for tests or known-good data from validated sources.
func MustParseWBSCode(s string) WBSCode {
	wbs, err := ParseWBSCode(s)
	if err != nil {
		panic("invalid WBS code: " + s)
	}
	return wbs
}

// GetMajorPhase returns the major phase component of the WBS code.
// Examples:
//   - "9.1.2" → "9"
//   - "14.3" → "14"
//   - "A-100" → "A"
//   - "9" → "9"
func (w WBSCode) GetMajorPhase() string {
	return w.majorPhase
}

// String returns the original WBS code string.
func (w WBSCode) String() string {
	return w.raw
}

// IsEmpty returns true if the WBSCode is the zero value.
func (w WBSCode) IsEmpty() bool {
	return w.raw == ""
}
