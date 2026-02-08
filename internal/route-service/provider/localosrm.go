package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"vehicle-tracking-simulation/internal/route-service/models"
)

// LocalOSRMProvider implements Provider interface using a local OSRM instance
// This is optimized for high-volume route generation with local data
type LocalOSRMProvider struct {
	BaseURL string
	Client  *http.Client
}

// NewLocalOSRMProvider creates a new local OSRM routing provider
// Defaults to localhost:5000 which is the standard OSRM port
func NewLocalOSRMProvider(config RouteFinderConfig) *LocalOSRMProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		// Default local OSRM instance
		baseURL = "http://localhost:5000"
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if config.Timeout == 0 {
		// Longer timeout for local instance (can handle more data)
		timeout = 30 * time.Second
	}

	return &LocalOSRMProvider{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// FindRoute finds a route between start and end coordinates using local OSRM
func (p *LocalOSRMProvider) FindRoute(start models.Coordinate, end models.Coordinate, profile string) (*models.RouteResponse, error) {
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
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters for detailed response
	q := req.URL.Query()
	q.Add("overview", "full")
	q.Add("steps", "true")
	q.Add("annotations", "true")
	q.Add("geometries", "polyline")  // Use polyline encoding (default)
	req.URL.RawQuery = q.Encode()

	// Send request
	resp, err := p.Client.Do(req)
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
func (p *LocalOSRMProvider) FindRouteWithWaypoints(waypoints []models.Coordinate, profile string) (*models.RouteResponse, error) {
	if len(waypoints) < 2 {
		return nil, fmt.Errorf("at least 2 waypoints required")
	}

	if profile == "" {
		profile = "driving"
	}

	osrmProfile := p.mapProfile(profile)

	// Build coordinates string: lon1,lat1;lon2,lat2;lon3,lat3...
	var coordsBuilder string
	for i, wp := range waypoints {
		if i > 0 {
			coordsBuilder += ";"
		}
		coordsBuilder += fmt.Sprintf("%f,%f", wp.Longitude, wp.Latitude)
	}

	apiURL := fmt.Sprintf("%s/route/v1/%s/%s", p.BaseURL, osrmProfile, coordsBuilder)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("overview", "full")
	q.Add("steps", "true")
	q.Add("annotations", "true")
	q.Add("geometries", "polyline")  // Use polyline encoding (default)
	req.URL.RawQuery = q.Encode()

	resp, err := p.Client.Do(req)
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
func (p *LocalOSRMProvider) ProviderName() string {
	return "local-osrm"
}

// mapProfile maps generic profiles to OSRM-specific profiles
func (p *LocalOSRMProvider) mapProfile(profile string) string {
	switch profile {
	case "driving", "car":
		return "car"
	case "biking", "bike", "bicycle":
		return "bike"
	case "walking", "foot", "pedestrian":
		return "foot"
	default:
		return "car"
	}
}