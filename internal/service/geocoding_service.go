package service

// GeocodingService provides address-to-coordinate resolution.
// See PRODUCTION_PLAN.md Critical Blocker A Remediation
//
// MVP Implementation: Returns hardcoded Austin, TX coordinates.
// Post-launch: Wire to Google Geocoding API or similar.
type GeocodingService struct{}

// NewGeocodingService creates a new GeocodingService instance.
func NewGeocodingService() *GeocodingService {
	return &GeocodingService{}
}

// Geocode returns lat/long for a given address string.
// MVP: Returns Austin, TX coordinates. Wire to real geocoding API post-launch.
func (s *GeocodingService) Geocode(address string) (lat, lng float64, err error) {
	// TODO: Integrate with Google Geocoding API
	// For now, return Austin, TX as default (maintains backward compatibility)
	return 30.2672, -97.7431, nil
}
