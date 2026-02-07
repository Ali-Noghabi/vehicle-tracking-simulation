package provider

import "vehicle-tracking-simulation/internal/route-service/models"

// Provider defines the interface for routing service providers
// This allows easy switching between different routing APIs (OpenStreetMap, Google Maps, Mapbox, etc.)
type Provider interface {
	// FindRoute finds a route between two coordinates
	// Returns a RouteResponse which follows OSRM-compatible format
	FindRoute(start models.Coordinate, end models.Coordinate, profile string) (*models.RouteResponse, error)

	// FindRouteWithWaypoints finds a route through multiple waypoints
	FindRouteWithWaypoints(waypoints []models.Coordinate, profile string) (*models.RouteResponse, error)

	// ProviderName returns the name of the routing provider
	ProviderName() string
}

// RouteFinderConfig contains configuration for routing providers
type RouteFinderConfig struct {
	ProviderType string                 // "openstreetmap", "google", "mapbox", etc.
	APIKey       string                 // API key for paid services
	BaseURL      string                 // Base URL for the routing service
	Timeout      int                    // Timeout in seconds
	ExtraParams  map[string]interface{} // Additional parameters for the provider
}
