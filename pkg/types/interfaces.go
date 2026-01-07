package types

// WeatherService defines the integration for the SWIM Model.
// See API_AND_TYPES_SPEC.md Section 2.1
type WeatherService interface {
	GetForecast(lat, long float64) (Forecast, error)
}

// VisionService defines the Validation Protocol service.
// See API_AND_TYPES_SPEC.md Section 2.2
type VisionService interface {
	// VerifyTask returns (is_verified, confidence_score, error)
	VerifyTask(imageURL string, taskDescription string) (bool, float64, error)
}

// NotificationService defines the outbound communication service.
// See API_AND_TYPES_SPEC.md Section 2.3
type NotificationService interface {
	SendSMS(contactID string, message string) error
}

// DirectoryService defines contact and assignment lookups.
// See API_AND_TYPES_SPEC.md Section 2.4
type DirectoryService interface {
	GetContactForPhase(phaseID string) (Contact, error)
}
