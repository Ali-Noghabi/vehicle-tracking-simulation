package models

import "time"

// Coordinate represents a geographic point
type Coordinate struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// RouteRequest represents the input for route finding
type RouteRequest struct {
	StartCoordinate Coordinate `json:"start" validate:"required"`
	EndCoordinate   Coordinate `json:"end" validate:"required"`
	Profile         string     `json:"profile,omitempty"` // car, bike, foot, etc.
}

// Leg represents a segment of a route between two waypoints
type Leg struct {
	Steps       []Step  `json:"steps"`
	Distance    float64 `json:"distance"`    // in meters
	Duration    float64 `json:"duration"`    // in seconds
	Summary     string  `json:"summary"`
	Annotation  *Annotation `json:"annotation,omitempty"`
}

// Step represents a single maneuver in the route
type Step struct {
	Distance    float64   `json:"distance"`    // in meters
	Duration    float64   `json:"duration"`    // in seconds
	Geometry    string    `json:"geometry"`    // encoded polyline
	Instruction string    `json:"instruction"`
	Name        string    `json:"name"`        // street name
	Maneuver    *Maneuver `json:"maneuver,omitempty"`
}

// Maneuver describes the action to take at a step
type Maneuver struct {
	Type        string    `json:"type"`        // turn, merge, etc.
	Modifier    string    `json:"modifier,omitempty"` // left, right, straight, etc.
	Location    []float64 `json:"location"`    // [longitude, latitude]
	BearingBefore int     `json:"bearing_before,omitempty"`
	BearingAfter  int     `json:"bearing_after,omitempty"`
}

// Annotation contains additional information about the route
type Annotation struct {
	Duration []float64 `json:"duration,omitempty"`
	Distance []float64 `json:"distance,omitempty"`
	Speed    []float64 `json:"speed,omitempty"`
}

// RouteResponse is the standard response format (OSRM-compatible)
type RouteResponse struct {
	Code    string `json:"code"`    // "Ok" on success
	Message string `json:"message,omitempty"` // error message if code is not "Ok"

	Routes []Route `json:"routes"`

	// Optional: waypoints for multi-point routes
	Waypoints []Waypoint `json:"waypoints,omitempty"`
}

// Route represents a complete route from start to end
type Route struct {
	Geometry   string  `json:"geometry"`     // encoded polyline
	Legs       []Leg   `json:"legs"`
	Distance   float64 `json:"distance"`     // total distance in meters
	Duration   float64 `json:"duration"`     // total duration in seconds
	WeightName string  `json:"weight_name"`  // "routability" or "duration"
	Weight     float64 `json:"weight"`       // calculated weight
	Summary    string  `json:"summary"`      // text summary of route
}

// Waypoint represents intermediate points in the route
type Waypoint struct {
	Name      string     `json:"name"`       // location name
	Location  []float64  `json:"location"`   // [longitude, latitude]
	Distance  float64    `json:"distance"`   // distance from start
	Hint      string     `json:"hint"`       // internal hint for OSRM
	SnappedDistance float64 `json:"snapped_distance,omitempty"`
}

// RouteStats contains statistics about the route
type RouteStats struct {
	TotalDistance float64       `json:"total_distance"` // meters
	TotalDuration time.Duration `json:"total_duration"` // time.Duration
	AverageSpeed float64       `json:"average_speed"` // km/h
}
