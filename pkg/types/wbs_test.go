package types

import (
	"testing"
)

func TestParseWBSCode(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantMajorPhase string
		wantErr        bool
	}{
		// CSI standard dot formats
		{name: "CSI simple", input: "9", wantMajorPhase: "9"},
		{name: "CSI two level", input: "9.1", wantMajorPhase: "9"},
		{name: "CSI three level", input: "9.1.2", wantMajorPhase: "9"},
		{name: "CSI double digit", input: "14.3", wantMajorPhase: "14"},
		{name: "CSI deep nesting", input: "14.3.1.5", wantMajorPhase: "14"},

		// Alpha-dash formats (common in construction)
		{name: "alpha dash", input: "A-100", wantMajorPhase: "A"},
		{name: "alpha dash deep", input: "B-200-1", wantMajorPhase: "B"},
		{name: "alpha space number", input: "MEP 100", wantMajorPhase: "MEP"},

		// Simple alpha formats
		{name: "single letter", input: "A", wantMajorPhase: "A"},
		{name: "lowercase letter", input: "b", wantMajorPhase: "B"},
		{name: "multi letter", input: "MEP", wantMajorPhase: "MEP"},

		// Edge cases with fallback handling
		{name: "dash format", input: "9-1-2", wantMajorPhase: "9"},
		{name: "underscore format", input: "9_1_2", wantMajorPhase: "9"},
		{name: "mixed format", input: "A.100", wantMajorPhase: "A"},

		// Whitespace handling
		{name: "leading space", input: " 9.1", wantMajorPhase: "9"},
		{name: "trailing space", input: "9.1 ", wantMajorPhase: "9"},

		// Error cases
		{name: "empty string", input: "", wantErr: true},
		{name: "whitespace only", input: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wbs, err := ParseWBSCode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseWBSCode(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseWBSCode(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got := wbs.GetMajorPhase(); got != tt.wantMajorPhase {
				t.Errorf("ParseWBSCode(%q).GetMajorPhase() = %q, want %q", tt.input, got, tt.wantMajorPhase)
			}
		})
	}
}

func TestWBSCode_String(t *testing.T) {
	wbs, _ := ParseWBSCode("9.1.2")
	if got := wbs.String(); got != "9.1.2" {
		t.Errorf("WBSCode.String() = %q, want %q", got, "9.1.2")
	}
}

func TestMustParseWBSCode(t *testing.T) {
	// Valid code should not panic
	wbs := MustParseWBSCode("9.1.2")
	if wbs.GetMajorPhase() != "9" {
		t.Errorf("MustParseWBSCode returned wrong phase")
	}

	// Invalid code should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseWBSCode with empty string should have panicked")
		}
	}()
	_ = MustParseWBSCode("")
}

func TestWBSCode_IsEmpty(t *testing.T) {
	var zero WBSCode
	if !zero.IsEmpty() {
		t.Error("Zero WBSCode should be empty")
	}

	wbs, _ := ParseWBSCode("9.1")
	if wbs.IsEmpty() {
		t.Error("Parsed WBSCode should not be empty")
	}
}
