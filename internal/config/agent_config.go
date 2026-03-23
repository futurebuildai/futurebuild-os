package config

// AgentSettings holds the full agent configuration for an organization.
// Stored as JSONB in agent_configs table.
type AgentSettings struct {
	DailyFocus  DailyFocusConfig  `json:"daily_focus"`
	Procurement ProcurementConfig `json:"procurement"`
	SubLiaison  SubLiaisonConfig  `json:"sub_liaison"`
	Chat        ChatConfig        `json:"chat"`
}

// DailyFocusConfig configures the Daily Focus Agent.
type DailyFocusConfig struct {
	Enabled       bool   `json:"enabled"`
	RunTime       string `json:"run_time"`        // HH:MM format, e.g. "07:00"
	AIProvider    string `json:"ai_provider"`      // "claude" or "gemini"
	MaxFocusCards int    `json:"max_focus_cards"`  // 3-5
}

// SubLiaisonConfig configures the Sub Liaison Agent.
type SubLiaisonConfig struct {
	Enabled             bool   `json:"enabled"`
	ConfirmationWindow  string `json:"confirmation_window"` // e.g. "72h"
	AutoResendAfter     string `json:"auto_resend_after"`   // e.g. "24h"
}

// ChatConfig configures the Chat Intelligence layer.
type ChatConfig struct {
	AIProvider   string `json:"ai_provider"`    // "claude" or "gemini"
	MaxToolCalls int    `json:"max_tool_calls"` // Default: 10
}

// DefaultAgentSettings returns sensible defaults for a new org.
func DefaultAgentSettings() AgentSettings {
	return AgentSettings{
		DailyFocus: DailyFocusConfig{
			Enabled:       true,
			RunTime:       "07:00",
			AIProvider:    "claude",
			MaxFocusCards: 3,
		},
		Procurement: ProcurementConfig{
			LeadTimeWarningThreshold: 14,
			StagingBufferDays:        2,
			DefaultWeatherBufferDays: 3,
		},
		SubLiaison: SubLiaisonConfig{
			Enabled:            true,
			ConfirmationWindow: "72h",
			AutoResendAfter:    "24h",
		},
		Chat: ChatConfig{
			AIProvider:   "claude",
			MaxToolCalls: 10,
		},
	}
}
