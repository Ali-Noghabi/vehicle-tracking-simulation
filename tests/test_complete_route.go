package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"vehicle-tracking-simulation/internal/route-service/models"
)

func main() {
	// Configuration
	port := "8090"
	if len(os.Args) > 1 && os.Args[1] == "-port" && len(os.Args) > 2 {
		port = os.Args[2]
	}
	baseURL := fmt.Sprintf("http://localhost:%s", port)

	fmt.Println("=== Complete Route Structure Test ===")
	fmt.Printf("Testing service at: %s\n\n", baseURL)

	// Test data - London to London Bridge
	start := models.Coordinate{Latitude: 51.5074, Longitude: -0.1278}  // London
	end := models.Coordinate{Latitude: 51.5155, Longitude: -0.1419}    // London Bridge

	fmt.Println("1. Getting complete route structure...")
	if err := testCompleteRoute(baseURL, start, end); err != nil {
		log.Fatalf("Complete route test failed: %v", err)
	}

	fmt.Println("\n2. Getting route with step-by-step instructions...")
	if err := testRouteWithSteps(baseURL, start, end); err != nil {
		log.Fatalf("Route steps test failed: %v", err)
	}

	fmt.Println("\n=== Test completed ===")
}

func testCompleteRoute(baseURL string, start, end models.Coordinate) error {
	request := models.RouteRequest{
		StartCoordinate: start,
		EndCoordinate:   end,
		Profile:         "car",
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/route", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to post route request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("route endpoint returned error: %v", errorResp["error"])
		}
		return fmt.Errorf("route endpoint returned status %d", resp.StatusCode)
	}

	var routeResp models.RouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&routeResp); err != nil {
		return fmt.Errorf("failed to parse route response: %w", err)
	}

	if routeResp.Code != "Ok" {
		return fmt.Errorf("route response code not Ok: %s", routeResp.Code)
	}

	if len(routeResp.Routes) == 0 {
		return fmt.Errorf("no routes returned")
	}

	route := routeResp.Routes[0]
	
	// Display complete route structure
	fmt.Printf("\n=== ROUTE STRUCTURE ===\n")
	fmt.Printf("Response Code: %s\n", routeResp.Code)
	fmt.Printf("Number of Routes: %d\n", len(routeResp.Routes))
	
	fmt.Printf("\n--- MAIN ROUTE ---\n")
	fmt.Printf("Total Distance: %.2f meters (%.2f km)\n", route.Distance, route.Distance/1000)
	fmt.Printf("Total Duration: %.2f seconds (%.2f minutes)\n", route.Duration, route.Duration/60)
	fmt.Printf("Weight Name: %s\n", route.WeightName)
	fmt.Printf("Weight: %.2f\n", route.Weight)
	fmt.Printf("Summary: %s\n", route.Summary)
	fmt.Printf("Geometry (polyline): %s... (truncated)\n", route.Geometry[:min(50, len(route.Geometry))])
	
	fmt.Printf("\n--- LEGS (%d total) ---\n", len(route.Legs))
	for i, leg := range route.Legs {
		fmt.Printf("\nLeg %d:\n", i+1)
		fmt.Printf("  Distance: %.2f meters\n", leg.Distance)
		fmt.Printf("  Duration: %.2f seconds\n", leg.Duration)
		fmt.Printf("  Summary: %s\n", leg.Summary)
		fmt.Printf("  Number of Steps: %d\n", len(leg.Steps))
		
		if leg.Annotation != nil {
			fmt.Printf("  Annotation available: Yes\n")
		}
	}
	
	// Show waypoints if available
	if len(routeResp.Waypoints) > 0 {
		fmt.Printf("\n--- WAYPOINTS (%d total) ---\n", len(routeResp.Waypoints))
		for i, wp := range routeResp.Waypoints {
			fmt.Printf("\nWaypoint %d:\n", i+1)
			fmt.Printf("  Name: %s\n", wp.Name)
			fmt.Printf("  Location: [%.6f, %.6f]\n", wp.Location[0], wp.Location[1])
			fmt.Printf("  Distance from start: %.2f meters\n", wp.Distance)
			if wp.Hint != "" {
				fmt.Printf("  Hint: %s...\n", wp.Hint[:min(20, len(wp.Hint))])
			}
		}
	}
	
	// Calculate statistics
	fmt.Printf("\n--- STATISTICS ---\n")
	if route.Duration > 0 {
		distanceKm := route.Distance / 1000.0
		hours := route.Duration / 3600.0
		averageSpeed := distanceKm / hours
		fmt.Printf("Average Speed: %.2f km/h\n", averageSpeed)
		fmt.Printf("Pace: %.2f minutes per km\n", (route.Duration/60)/distanceKm)
	}
	
	return nil
}

func testRouteWithSteps(baseURL string, start, end models.Coordinate) error {
	request := models.RouteRequest{
		StartCoordinate: start,
		EndCoordinate:   end,
		Profile:         "car",
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/route", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to post route request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("route endpoint returned status %d", resp.StatusCode)
	}

	var routeResp models.RouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&routeResp); err != nil {
		return fmt.Errorf("failed to parse route response: %w", err)
	}

	if routeResp.Code != "Ok" || len(routeResp.Routes) == 0 {
		return fmt.Errorf("no valid route returned")
	}

	route := routeResp.Routes[0]
	
	fmt.Printf("\n=== STEP-BY-STEP INSTRUCTIONS ===\n")
	fmt.Printf("Total steps in route: %d\n\n", countTotalSteps(route))
	
	stepCounter := 1
	for legIndex, leg := range route.Legs {
		fmt.Printf("--- LEG %d (%.2f meters, %.2f seconds) ---\n", 
			legIndex+1, leg.Distance, leg.Duration)
		
		for stepIndex, step := range leg.Steps {
			fmt.Printf("\nStep %d.%d:\n", legIndex+1, stepIndex+1)
			fmt.Printf("  Distance: %.0f meters\n", step.Distance)
			fmt.Printf("  Duration: %.0f seconds\n", step.Duration)
			fmt.Printf("  Instruction: %s\n", step.Instruction)
			fmt.Printf("  Street: %s\n", step.Name)
			
			if step.Maneuver != nil {
				fmt.Printf("  Maneuver: %s", step.Maneuver.Type)
				if step.Maneuver.Modifier != "" {
					fmt.Printf(" (%s)", step.Maneuver.Modifier)
				}
				fmt.Printf("\n")
				
				if len(step.Maneuver.Location) == 2 {
					fmt.Printf("  Location: [%.6f, %.6f]\n", 
						step.Maneuver.Location[0], step.Maneuver.Location[1])
				}
			}
			
			stepCounter++
		}
	}
	
	return nil
}

func countTotalSteps(route models.Route) int {
	total := 0
	for _, leg := range route.Legs {
		total += len(leg.Steps)
	}
	return total
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}