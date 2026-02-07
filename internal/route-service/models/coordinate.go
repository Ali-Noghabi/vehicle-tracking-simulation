package models

import (
	"fmt"
	"math"
)

// Validate validates the coordinate values
func (c *Coordinate) Validate() error {
	if c.Latitude < -90 || c.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", c.Latitude)
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", c.Longitude)
	}
	return nil
}

// IsZero returns true if the coordinate is at zero/zero position
func (c *Coordinate) IsZero() bool {
	return c.Latitude == 0 && c.Longitude == 0
}

// DistanceTo calculates the distance to another coordinate in meters
// using the Haversine formula
func (c *Coordinate) DistanceTo(other Coordinate) float64 {
	const earthRadius = 6371000 // meters

	lat1 := c.Latitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	deltaLat := (other.Latitude - c.Latitude) * math.Pi / 180
	deltaLon := (other.Longitude - c.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	cVal := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * cVal
}

// BearingTo calculates the initial bearing to another coordinate in degrees
func (c *Coordinate) BearingTo(other Coordinate) float64 {
	lat1 := c.Latitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	deltaLon := (other.Longitude - c.Longitude) * math.Pi / 180

	x := math.Sin(deltaLon) * math.Cos(lat2)
	y := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLon)

	bearing := math.Atan2(x, y) * 180 / math.Pi
	bearing = math.Mod(bearing+360, 360)

	return bearing
}
