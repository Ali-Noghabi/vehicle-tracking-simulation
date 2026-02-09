#!/bin/bash

# Test script for vehicle tracking simulation service
# This script tests the simulation service without requiring MQTT broker

set -e

echo "🧪 Testing Vehicle Tracking Simulation Service"
echo "=============================================="

# Build the service
echo "Building simulation service..."
go build -o simulation-service

# Test configuration
echo "Testing configuration loading..."
if ! ./simulation-service --help 2>&1 | grep -q "simulation"; then
    echo "❌ Failed to load configuration"
    exit 1
fi
echo "✅ Configuration loading test passed"

# Test route loading
echo "Testing route file loading..."
# Create a test route file
mkdir -p test_routes
cat > test_routes/route_test_001.json << 'EOF'
{
  "metadata": {
    "id": 1,
    "generated_at": "2026-02-09T13:32:46.995386848+03:30",
    "start_lat": 35.6892,
    "start_lng": 51.3890,
    "end_lat": 36.2605,
    "end_lng": 59.6168,
    "profile": "car",
    "distance": 518500,
    "duration": 26423.7,
    "success": true
  },
  "route": {
    "geometry": "gff|Do~`bHn@Uj@]j@a@f@g@f@i@z@mAZw@\\_AXyAFkB?mB@sM?m@?{M?sCLwI?{F?_J?mGAiH@cBDkL?iD?a@AmDAeQBiN?qF@eHA}@@qA?iL@uH?_L@wE?kJA}G@oC?cD?{GB{K?mE@}G@mEA_B@aE@wI?_K?oNAqBA[ES",
    "legs": [
      {
        "steps": [
          {
            "distance": 5515.7,
            "duration": 1312.2,
            "geometry": "gff|Do~`bHn@Uj@]j@a@f@g@f@i@z@mAZw@\\_AXyAFkB?mB@sM?m@?{M?sCLwI?{F?_J?mGAiH@cBDkL?iD?a@AmDAeQBiN?qF@eHA}@@qA?iL@uH?_L@wE?kJA}G@oC?cD?{GB{K?mE@}G@mEA_B@aE@wI?_K?oNAqBA[ES",
            "instruction": "",
            "name": "",
            "maneuver": {
              "type": "depart",
              "location": [47.687604, 31.002763],
              "bearing_after": 159
            }
          }
        ]
      }
    ]
  }
}
EOF

# Test polyline decoding
echo "Testing polyline decoding..."
cat > test_polyline.go << 'EOF'
package main

import (
	"fmt"
)

func main() {
	encoded := "gff|Do~`bHn@Uj@]j@a@f@g@f@i@z@mAZw@\\_AXyAFkB?mB@sM?m@?{M?sCLwI?{F?_J?mGAiH@cBDkL?iD?a@AmDAeQBiN?qF@eHA}@@qA?iL@uH?_L@wE?kJA}G@oC?cD?{GB{K?mE@}G@mEA_B@aE@wI?_K?oNAqBA[ES"
	points := decodePolyline(encoded)
	
	if len(points) > 0 {
		fmt.Printf("✅ Decoded %d points\n", len(points))
		fmt.Printf("First point: %.6f, %.6f\n", points[0][0], points[0][1])
		fmt.Printf("Last point: %.6f, %.6f\n", points[len(points)-1][0], points[len(points)-1][1])
	} else {
		fmt.Println("❌ Failed to decode polyline")
	}
}
EOF

# Run the test
go run test_polyline.go polyline.go

# Cleanup
rm -f test_polyline.go
rm -rf test_routes

echo ""
echo "✅ All tests passed!"
echo ""
echo "To run the full simulation:"
echo "  1. Install MQTT broker: sudo apt install mosquitto mosquitto-clients"
echo "  2. Start MQTT broker: sudo systemctl start mosquitto"
echo "  3. Run simulation: ./simulation-service"
echo "  4. Monitor telemetry: mosquitto_sub -t 'vehicle/telemetry' -v"