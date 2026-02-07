package service

import (
	"fmt"
	"time"

	"vehicle-tracking-simulation/internal/route-service/models"
	"vehicle-tracking-simulation/internal/route-service/provider"
)

// RouteFinder is the service for finding routes between coordinates
// It uses a provider interface to allow different routing backends
type RouteFinder struct {
	provider provider.Provider
}

// NewRouteFinder creates a new RouteFinder service with the specified provider
func NewRouteFinder(p provider.Provider) *RouteFinder {
	return &RouteFinder{
		provider: p,
	}
}

// FindRoute finds a route between two coordinates
func (rf *RouteFinder) FindRoute(req models.RouteRequest) (*models.RouteResponse, error) {
	// Validate request
	if err := rf.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Use provider to find route
	routeResp, err := rf.provider.FindRoute(req.StartCoordinate, req.EndCoordinate, req.Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to find route: %w", err)
	}

	// Additional validation of response
	if routeResp == nil || len(routeResp.Routes) == 0 {
		return nil, fmt.Errorf("no route found between the specified coordinates")
	}

	return routeResp, nil
}

// FindRouteWithWaypoints finds a route through multiple waypoints
func (rf *RouteFinder) FindRouteWithWaypoints(waypoints []models.Coordinate, profile string) (*models.RouteResponse, error) {
	if len(waypoints) < 2 {
		return nil, fmt.Errorf("at least 2 waypoints required")
	}

	routeResp, err := rf.provider.FindRouteWithWaypoints(waypoints, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to find route: %w", err)
	}

	if routeResp == nil || len(routeResp.Routes) == 0 {
		return nil, fmt.Errorf("no route found")
	}

	return routeResp, nil
}

// GetRouteStats calculates statistics for a route
func (rf *RouteFinder) GetRouteStats(route *models.Route) *models.RouteStats {
	stats := &models.RouteStats{
		TotalDistance: route.Distance,
		TotalDuration: time.Duration(route.Duration) * time.Second,
	}

	// Calculate average speed: distance (km) / time (hours)
	if stats.TotalDuration > 0 {
		distanceKm := route.Distance / 1000.0
		hours := stats.TotalDuration.Hours()
		if hours > 0 {
			stats.AverageSpeed = distanceKm / hours
		}
	}

	return stats
}

// GetProvider returns the current routing provider
func (rf *RouteFinder) GetProvider() provider.Provider {
	return rf.provider
}

// validateRequest validates the route request
func (rf *RouteFinder) validateRequest(req models.RouteRequest) error {
	// Check coordinates are within valid ranges
	if req.StartCoordinate.Latitude < -90 || req.StartCoordinate.Latitude > 90 {
		return fmt.Errorf("invalid start latitude: %f", req.StartCoordinate.Latitude)
	}
	if req.StartCoordinate.Longitude < -180 || req.StartCoordinate.Longitude > 180 {
		return fmt.Errorf("invalid start longitude: %f", req.StartCoordinate.Longitude)
	}
	if req.EndCoordinate.Latitude < -90 || req.EndCoordinate.Latitude > 90 {
		return fmt.Errorf("invalid end latitude: %f", req.EndCoordinate.Latitude)
	}
	if req.EndCoordinate.Longitude < -180 || req.EndCoordinate.Longitude > 180 {
		return fmt.Errorf("invalid end longitude: %f", req.EndCoordinate.Longitude)
	}

	return nil
}