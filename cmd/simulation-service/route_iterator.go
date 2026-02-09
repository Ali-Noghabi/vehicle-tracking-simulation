package main

import (
	"math"
	"math/rand"
	"time"
)

// RouteIterator provides position calculation along a route
type RouteIterator struct {
	Route         *Route
	Points        [][2]float64
	SegmentLengths []float64
	TotalLength   float64
	CurrentIndex  int
	CurrentPos    float64 // position along current segment (0-1)
}

// NewRouteIterator creates a new iterator for a route
func NewRouteIterator(route *Route) *RouteIterator {
	// Decode the polyline geometry
	points := decodePolyline(route.Route.Geometry)
	
	// Calculate segment lengths
	segmentLengths := make([]float64, len(points)-1)
	totalLength := 0.0
	
	for i := 0; i < len(points)-1; i++ {
		dist := calculateDistance(
			points[i][0], points[i][1],
			points[i+1][0], points[i+1][1],
		)
		segmentLengths[i] = dist
		totalLength += dist
	}
	
	return &RouteIterator{
		Route:         route,
		Points:        points,
		SegmentLengths: segmentLengths,
		TotalLength:   totalLength,
		CurrentIndex:  0,
		CurrentPos:    0,
	}
}

// CalculatePosition calculates position along route based on distance traveled
func (ri *RouteIterator) CalculatePosition(distanceTraveled float64) (lat, lng, heading float64) {
	if distanceTraveled >= ri.TotalLength {
		// At or beyond end of route
		lastPoint := ri.Points[len(ri.Points)-1]
		secondLast := ri.Points[len(ri.Points)-2]
		return lastPoint[0], lastPoint[1], calculateHeading(
			secondLast[0], secondLast[1],
			lastPoint[0], lastPoint[1],
		)
	}
	
	// Find which segment we're in
	accumulated := 0.0
	for i, segmentLength := range ri.SegmentLengths {
		if distanceTraveled <= accumulated+segmentLength {
			// Found the segment
			segmentProgress := (distanceTraveled - accumulated) / segmentLength
			
			p1 := ri.Points[i]
			p2 := ri.Points[i+1]
			point := interpolatePoint(p1, p2, segmentProgress)
			
			// Calculate heading based on segment direction
			heading = calculateHeading(p1[0], p1[1], p2[0], p2[1])
			
			return point[0], point[1], heading
		}
		accumulated += segmentLength
	}
	
	// Should not reach here
	lastPoint := ri.Points[len(ri.Points)-1]
	return lastPoint[0], lastPoint[1], 0
}

// UpdateVehicleSimulator updates the vehicle simulator with proper route iteration
func (v *VehicleSimulator) UpdateWithRouteIterator(currentTime time.Time) *Telemetry {
	elapsed := currentTime.Sub(v.StartTime).Seconds()
	
	// Use random speed within range for realism
	speed := v.SpeedRange[0] + rand.Float64()*(v.SpeedRange[1]-v.SpeedRange[0])
	distanceTraveled := speed * elapsed
	
	// Create iterator if not exists
	if v.RouteIterator == nil {
		v.RouteIterator = NewRouteIterator(v.Route)
	}
	
	// Calculate position along route
	lat, lng, heading := v.RouteIterator.CalculatePosition(distanceTraveled)
	
	// Generate random values with validation
	altitude := 100 + rand.Float64()*50
	accuracy := 5 + rand.Float64()*10
	battery := 80 + rand.Float64()*20
	signal := 70 + rand.Float64()*30
	
	// Validate all values to ensure they're valid numbers
	if math.IsNaN(speed) || math.IsInf(speed, 0) {
		speed = 0.0
	}
	if math.IsNaN(heading) || math.IsInf(heading, 0) {
		heading = 0.0
	}
	if math.IsNaN(altitude) || math.IsInf(altitude, 0) {
		altitude = 100.0
	}
	if math.IsNaN(accuracy) || math.IsInf(accuracy, 0) {
		accuracy = 10.0
	}
	if math.IsNaN(battery) || math.IsInf(battery, 0) {
		battery = 90.0
	}
	if math.IsNaN(signal) || math.IsInf(signal, 0) {
		signal = 85.0
	}
	
	telemetry := &Telemetry{
		VehicleID: v.VehicleID,
		Timestamp: currentTime.Unix(),
		Lat:       lat,
		Lon:       lng,
		Speed:     speed * 3.6, // Convert m/s to km/h
		Heading:   heading,
		Altitude:  altitude,
		Accuracy:  accuracy,
		Battery:   battery,
		Signal:    signal,
	}
	
	return telemetry
}