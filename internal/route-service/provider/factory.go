package provider

import (
	"fmt"
)

// NewProvider creates a routing provider based on the specified type
// This factory pattern allows easy switching between different routing services
func NewProvider(config RouteFinderConfig) (Provider, error) {
	switch config.ProviderType {
	case "", "openstreetmap", "osrm":
		return NewOpenStreetMapProvider(config), nil

	case "local-osrm", "localosrm":
		return NewLocalOSRMProvider(config), nil

	case "google", "googlemaps":
		// TODO: Implement Google Maps provider
		return nil, fmt.Errorf("Google Maps provider not yet implemented")

	case "mapbox":
		// TODO: Implement Mapbox provider
		return nil, fmt.Errorf("Mapbox provider not yet implemented")

	case "here":
		// TODO: Implement HERE Technologies provider
		return nil, fmt.Errorf("HERE provider not yet implemented")

	case "graphhopper":
		// TODO: Implement GraphHopper provider
		return nil, fmt.Errorf("GraphHopper provider not yet implemented")

	default:
		return nil, fmt.Errorf("unknown provider type: %s", config.ProviderType)
	}
}
