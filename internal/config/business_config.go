// Package config provides application configuration.
// See PRODUCTION_PLAN.md: Configuration Decoupling for business rules.
package config

import (
	"os"
	"strconv"
)

// PhysicsConfig holds tunable physics engine parameters.
// These control the DHSM (Duration & House Size Model) calculations.
// See CPM_RES_MODEL_SPEC.md Section 11.2.1
type PhysicsConfig struct {
	// StandardHouseSizeSF is the baseline GSF where SAF = 1.0.
	// Default: 2250.0 square feet.
	StandardHouseSizeSF float64

	// SizeAdjustmentExponent is the power curve for duration scaling.
	// SAF = (GSF / StandardHouseSizeSF) ^ SizeAdjustmentExponent
	// Default: 0.75
	SizeAdjustmentExponent float64

	// ConfigVersion for audit traceability.
	// Logged when schedules are calculated to track which config was used.
	ConfigVersion string
}

// ProcurementConfig holds tunable procurement agent parameters.
// These control order date calculations and alert timing.
// See PRODUCTION_PLAN.md Step 46
type ProcurementConfig struct {
	// StagingBufferDays is the time needed for jobsite staging before work begins.
	// Materials must arrive this many days before the task's EarlyStart.
	// Default: 2 days
	StagingBufferDays int

	// LeadTimeWarningThreshold is the number of days before the order deadline
	// that triggers a WARNING status instead of OK.
	// Default: 3 days
	LeadTimeWarningThreshold int

	// DefaultWeatherBufferDays is the conservative weather buffer when weather service
	// is unavailable or geocoding is not implemented.
	// P1 Fix: Default: 3 days (fail-safe, NEVER zero)
	// See PRODUCTION_PLAN.md Phase 49 Retrofit (Operation Ironclad Task 3)
	DefaultWeatherBufferDays int

	// ConfigVersion for audit traceability.
	ConfigVersion string
}

// DefaultPhysicsConfig returns a PhysicsConfig with safe production defaults.
// FAANG Threshold: Zero-value safety - if config is unset, use sensible defaults.
func DefaultPhysicsConfig() PhysicsConfig {
	return PhysicsConfig{
		StandardHouseSizeSF:    2250.0,
		SizeAdjustmentExponent: 0.75,
		ConfigVersion:          "default-v1",
	}
}

// DefaultProcurementConfig returns a ProcurementConfig with safe production defaults.
// FAANG Threshold: Zero-value safety - if config is unset, use sensible defaults.
func DefaultProcurementConfig() ProcurementConfig {
	return ProcurementConfig{
		StagingBufferDays:        2,
		LeadTimeWarningThreshold: 3,
		DefaultWeatherBufferDays: 3, // P1 Fix: Conservative default, NEVER zero
		ConfigVersion:            "default-v1",
	}
}

// WithDefaults returns a PhysicsConfig with zero values replaced by defaults.
func (c PhysicsConfig) WithDefaults() PhysicsConfig {
	defaults := DefaultPhysicsConfig()
	if c.StandardHouseSizeSF <= 0 {
		c.StandardHouseSizeSF = defaults.StandardHouseSizeSF
	}
	if c.SizeAdjustmentExponent <= 0 {
		c.SizeAdjustmentExponent = defaults.SizeAdjustmentExponent
	}
	if c.ConfigVersion == "" {
		c.ConfigVersion = defaults.ConfigVersion
	}
	return c
}

// WithDefaults returns a ProcurementConfig with zero values replaced by defaults.
func (c ProcurementConfig) WithDefaults() ProcurementConfig {
	defaults := DefaultProcurementConfig()
	if c.StagingBufferDays <= 0 {
		c.StagingBufferDays = defaults.StagingBufferDays
	}
	if c.LeadTimeWarningThreshold <= 0 {
		c.LeadTimeWarningThreshold = defaults.LeadTimeWarningThreshold
	}
	// P1 Fix: Ensure weather buffer is never zero (fail-safe)
	if c.DefaultWeatherBufferDays <= 0 {
		c.DefaultWeatherBufferDays = defaults.DefaultWeatherBufferDays
	}
	if c.ConfigVersion == "" {
		c.ConfigVersion = defaults.ConfigVersion
	}
	return c
}

// ConfigProvider defines the interface for retrieving business configuration.
// MVP: Returns static config loaded at startup.
// Preferred: Implementations can hot-reload from DB/Redis.
type ConfigProvider interface {
	GetPhysicsConfig() PhysicsConfig
	GetProcurementConfig() ProcurementConfig
}

// StaticConfigProvider is a simple implementation of ConfigProvider
// that returns fixed config values. Used for MVP.
type StaticConfigProvider struct {
	Physics     PhysicsConfig
	Procurement ProcurementConfig
}

// GetPhysicsConfig returns the physics configuration.
func (p *StaticConfigProvider) GetPhysicsConfig() PhysicsConfig {
	return p.Physics.WithDefaults()
}

// GetProcurementConfig returns the procurement configuration.
func (p *StaticConfigProvider) GetProcurementConfig() ProcurementConfig {
	return p.Procurement.WithDefaults()
}

// LoadPhysicsConfigFromEnv loads PhysicsConfig from environment variables.
// Falls back to defaults if env vars are not set.
func LoadPhysicsConfigFromEnv() PhysicsConfig {
	cfg := PhysicsConfig{}

	if v := os.Getenv("PHYSICS_STANDARD_HOUSE_SIZE_SF"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.StandardHouseSizeSF = f
		}
	}

	if v := os.Getenv("PHYSICS_SIZE_ADJUSTMENT_EXPONENT"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.SizeAdjustmentExponent = f
		}
	}

	cfg.ConfigVersion = os.Getenv("PHYSICS_CONFIG_VERSION")
	if cfg.ConfigVersion == "" {
		cfg.ConfigVersion = "env-v1"
	}

	return cfg.WithDefaults()
}

// LoadProcurementConfigFromEnv loads ProcurementConfig from environment variables.
// Falls back to defaults if env vars are not set.
func LoadProcurementConfigFromEnv() ProcurementConfig {
	cfg := ProcurementConfig{}

	if v := os.Getenv("PROCUREMENT_STAGING_BUFFER_DAYS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.StagingBufferDays = i
		}
	}

	if v := os.Getenv("PROCUREMENT_WARNING_THRESHOLD_DAYS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.LeadTimeWarningThreshold = i
		}
	}

	cfg.ConfigVersion = os.Getenv("PROCUREMENT_CONFIG_VERSION")
	if cfg.ConfigVersion == "" {
		cfg.ConfigVersion = "env-v1"
	}

	return cfg.WithDefaults()
}
