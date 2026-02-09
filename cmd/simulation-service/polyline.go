package main

import (
	"math"
)

// decodePolyline decodes a Google Maps encoded polyline
// Returns slice of [lat, lng] pairs
func decodePolyline(encoded string) [][2]float64 {
	var points [][2]float64
	var index, lat, lng int32

	for index < int32(len(encoded)) {
		var b int32
		var shift uint
		var result int32
		
		for {
			b = int32(encoded[index]) - 63
			index++
			result |= (b & 0x1F) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		
		if (result & 1) != 0 {
			result = ^(result >> 1)
		} else {
			result = result >> 1
		}
		
		lat += result

		shift = 0
		result = 0
		
		for {
			b = int32(encoded[index]) - 63
			index++
			result |= (b & 0x1F) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		
		if (result & 1) != 0 {
			result = ^(result >> 1)
		} else {
			result = result >> 1
		}
		
		lng += result
		
		points = append(points, [2]float64{
			float64(lat) / 1e5,
			float64(lng) / 1e5,
		})
	}
	
	return points
}

// calculateDistance calculates distance between two points using Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180
	
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return R * c
}

// interpolatePoint interpolates a point along a polyline segment
func interpolatePoint(p1, p2 [2]float64, fraction float64) [2]float64 {
	return [2]float64{
		p1[0] + fraction*(p2[0]-p1[0]),
		p1[1] + fraction*(p2[1]-p1[1]),
	}
}

// calculateHeading calculates heading from point1 to point2 in degrees
func calculateHeading(lat1, lon1, lat2, lon2 float64) float64 {
	// Check if points are identical (or very close)
	if math.Abs(lat1-lat2) < 1e-9 && math.Abs(lon1-lon2) < 1e-9 {
		return 0.0 // Default heading when points are identical
	}
	
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180
	
	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) -
		math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	
	// Check for zero values that could cause NaN
	if math.Abs(x) < 1e-9 && math.Abs(y) < 1e-9 {
		return 0.0
	}
	
	θ := math.Atan2(y, x)
	heading := θ * 180 / math.Pi
	
	// Normalize to 0-360
	if heading < 0 {
		heading += 360
	}
	
	// Ensure heading is a valid number
	if math.IsNaN(heading) || math.IsInf(heading, 0) {
		return 0.0
	}
	
	return heading
}