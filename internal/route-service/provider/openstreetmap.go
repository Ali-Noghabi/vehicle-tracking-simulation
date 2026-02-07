package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"vehicle-tracking-simulation/internal/route-service/models"
)

// OpenStreetMapProvider implements Provider interface using OSRM API
type OpenStreetMapProvider struct {
	BaseURL string
	Client  *http.Client
}

// NewOpenStreetMapProvider creates a new OpenStreetMap routing provider
// Using public OSRM demo server by default
// For production, you should run your own OSRM instance or use a commercial alternative
func NewOpenStreetMapProvider(config RouteFinderConfig) *OpenStreetMapProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		// Public OSRM demo server (not recommended for production)
		baseURL = "https://router.project-osrm.org"
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if config.Timeout == 0 {
		timeout = 10 * time.Second
	}

	return &OpenStreetMapProvider{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// FindRoute finds a route between start and end coordinates using OSRM
// OSRM API documentation: https://project-osrm.org/docs/v5.24.0/api/
func (p *OpenStreetMapProvider) FindRoute(start models.Coordinate, end models.Coordinate, profile string) (*models.RouteResponse, error) {
	// Default to car profile if not specified
	if profile == "" {
		profile = "driving"
	}

	// OSRM profiles: car, bike, foot
	osrmProfile := p.mapProfile(profile)

	// Build URL: /route/v1/{profile}/{lon},{lat};{lon},{lat}
	// Note: OSRM uses [lon, lat] order
	coordinates := fmt.Sprintf("%f,%f;%f,%f", start.Longitude, start.Latitude, end.Longitude, end.Latitude)

	apiURL := fmt.Sprintf("%s/route/v1/%s/%s", p.BaseURL, osrmProfile, coordinates)

	// Build query parameters
	params := url.Values{}
	params.Add("overview", "full")        // Return full geometry
	params.Add("geometries", "polyline")  // Use polyline encoding
	params.Add("steps", "true")           // Include turn-by-turn instructions
	params.Add("annotations", "true")     // Include speed, duration, distance data

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// Make HTTP request
	resp, err := p.Client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call OSRM API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSRM API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse OSRM response (it's already in our standard format)
	var routeResp models.RouteResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse OSRM response: %w", err)
	}

	return &routeResp, nil
}

// FindRouteWithWaypoints finds a route through multiple intermediate points
func (p *OpenStreetMapProvider) FindRouteWithWaypoints(waypoints []models.Coordinate, profile string) (*models.RouteResponse, error) {
	if len(waypoints) < 2 {
		return nil, fmt.Errorf("at least 2 waypoints required")
	}

	if profile == "" {
		profile = "driving"
	}

	osrmProfile := p.mapProfile(profile)

	// Build coordinate string with semicolon separators
	coords := ""
	for i, wp := range waypoints {
		if i > 0 {
			coords += ";"
		}
		coords += fmt.Sprintf("%f,%f", wp.Longitude, wp.Latitude)
	}

	apiURL := fmt.Sprintf("%s/route/v1/%s/%s", p.BaseURL, osrmProfile, coords)

	params := url.Values{}
	params.Add("overview", "full")
	params.Add("geometries", "polyline")
	params.Add("steps", "true")
	params.Add("annotations", "true")

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := p.Client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call OSRM API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSRM API returned status %d: %s", resp.StatusCode, string(body))
	}

	var routeResp models.RouteResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse OSRM response: %w", err)
	}

	return &routeResp, nil
}

// ProviderName returns the name of this provider
func (p *OpenStreetMapProvider) ProviderName() string {
	return "openstreetmap"
}

// mapProfile maps generic profiles to OSRM-specific profiles
func (p *OpenStreetMapProvider) mapProfile(profile string) string {
	switch profile {
	case "car", "driving", "vehicle":
		return "car"
	case "bike", "bicycle", "cycling":
		return "bike"
	case "foot", "walk", "walking":
		return "foot"
	default:
		return "car" // default to car
	}
}
