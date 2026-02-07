package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"vehicle-tracking-simulation/internal/route-service/models"
)

func main() {
	// Configuration
	port := "8090"
	if len(os.Args) > 1 && os.Args[1] == "-port" && len(os.Args) > 2 {
		port = os.Args[2]
	}
	baseURL := fmt.Sprintf("http://localhost:%s", port)

	// Test data - London coordinates
	start := models.Coordinate{Latitude: 51.5074, Longitude: -0.1278}  // London
	end := models.Coordinate{Latitude: 51.5155, Longitude: -0.1419}    // London Bridge
	waypoint := models.Coordinate{Latitude: 51.5088, Longitude: -0.0977} // Tower of London

	fmt.Println("=== Vehicle Tracking Route Service Test ===")
	fmt.Printf("Testing service at: %s\n\n", baseURL)

	// Test 1: Health check
	fmt.Println("1. Testing health endpoint...")
	if err := testHealth(baseURL); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	// Test 2: Provider info
	fmt.Println("\n2. Testing provider endpoint...")
	if err := testProvider(baseURL); err != nil {
		log.Fatalf("Provider test failed: %v", err)
	}

	// Test 3: Simple route
	fmt.Println("\n3. Testing simple route...")
	if err := testSimpleRoute(baseURL, start, end); err != nil {
		log.Fatalf("Simple route test failed: %v", err)
	}

	// Test 4: Route with waypoints
	fmt.Println("\n4. Testing route with waypoints...")
	if err := testRouteWithWaypoints(baseURL, start, waypoint, end); err != nil {
		log.Fatalf("Waypoints test failed: %v", err)
	}

	// Test 5: Different profiles
	fmt.Println("\n5. Testing different routing profiles...")
	profiles := []string{"car", "bike", "foot"}
	for _, profile := range profiles {
		if err := testProfile(baseURL, start, end, profile); err != nil {
			log.Printf("Profile %s test failed: %v", profile, err)
		}
	}

	// Test 6: Error cases
	fmt.Println("\n6. Testing error cases...")
	if err := testErrorCases(baseURL); err != nil {
		log.Printf("Error cases test failed: %v", err)
	}

	fmt.Println("\n=== All tests completed ===")
}

func testHealth(baseURL string) error {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		return fmt.Errorf("failed to reach health endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse health response: %w", err)
	}

	fmt.Printf("  Status: %s\n", result["status"])
	fmt.Printf("  Service: %s\n", result["service"])
	fmt.Printf("  Provider: %s\n", result["provider"])
	return nil
}

func testProvider(baseURL string) error {
	resp, err := http.Get(baseURL + "/api/v1/provider")
	if err != nil {
		return fmt.Errorf("failed to reach provider endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("provider endpoint returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse provider response: %w", err)
	}

	fmt.Printf("  Current provider: %s\n", result["provider"])
	return nil
}

func testSimpleRoute(baseURL string, start, end models.Coordinate) error {
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
	fmt.Printf("  Route found:\n")
	fmt.Printf("    Distance: %.2f meters\n", route.Distance)
	fmt.Printf("    Duration: %.2f seconds\n", route.Duration)
	fmt.Printf("    Summary: %s\n", route.Summary)

	// Calculate average speed
	if route.Duration > 0 {
		distanceKm := route.Distance / 1000.0
		hours := route.Duration / 3600.0
		averageSpeed := distanceKm / hours
		fmt.Printf("    Average speed: %.2f km/h\n", averageSpeed)
	}

	return nil
}

func testRouteWithWaypoints(baseURL string, start, waypoint, end models.Coordinate) error {
	waypoints := []models.Coordinate{start, waypoint, end}

	request := map[string]interface{}{
		"waypoints": waypoints,
		"profile":   "car",
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/route/waypoints", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to post waypoints request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("waypoints endpoint returned error: %v", errorResp["error"])
		}
		return fmt.Errorf("waypoints endpoint returned status %d", resp.StatusCode)
	}

	var routeResp models.RouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&routeResp); err != nil {
		return fmt.Errorf("failed to parse waypoints response: %w", err)
	}

	if routeResp.Code != "Ok" {
		return fmt.Errorf("waypoints response code not Ok: %s", routeResp.Code)
	}

	if len(routeResp.Routes) == 0 {
		return fmt.Errorf("no routes returned")
	}

	route := routeResp.Routes[0]
	fmt.Printf("  Multi-point route found:\n")
	fmt.Printf("    Total distance: %.2f meters\n", route.Distance)
	fmt.Printf("    Total duration: %.2f seconds\n", route.Duration)
	fmt.Printf("    Number of legs: %d\n", len(route.Legs))

	return nil
}

func testProfile(baseURL string, start, end models.Coordinate, profile string) error {
	request := models.RouteRequest{
		StartCoordinate: start,
		EndCoordinate:   end,
		Profile:         profile,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/route", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to post %s profile request: %w", profile, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s profile returned status %d", profile, resp.StatusCode)
	}

	var routeResp models.RouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&routeResp); err != nil {
		return fmt.Errorf("failed to parse %s profile response: %w", profile, err)
	}

	if routeResp.Code == "Ok" && len(routeResp.Routes) > 0 {
		route := routeResp.Routes[0]
		fmt.Printf("  %s profile: %.2f meters, %.2f seconds\n", profile, route.Distance, route.Duration)
		return nil
	}

	return fmt.Errorf("%s profile failed: %s", profile, routeResp.Code)
}

func testErrorCases(baseURL string) error {
	// Test 1: Invalid coordinates (latitude > 90)
	invalidRequest := map[string]interface{}{
		"start": map[string]float64{
			"latitude":  100.0, // Invalid
			"longitude": -0.1278,
		},
		"end": map[string]float64{
			"latitude":  51.5155,
			"longitude": -0.1419,
		},
	}

	body, err := json.Marshal(invalidRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal invalid request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/route", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to post invalid request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("  ✓ Invalid coordinates handled correctly")
	} else {
		fmt.Println("  ✗ Invalid coordinates not handled correctly")
	}

	// Test 2: Missing required field
	missingFieldRequest := map[string]interface{}{
		"start": map[string]float64{
			"latitude":  51.5074,
			"longitude": -0.1278,
		},
		// Missing "end" field
	}

	body, err = json.Marshal(missingFieldRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal missing field request: %w", err)
	}

	resp, err = http.Post(baseURL+"/api/v1/route", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to post missing field request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		fmt.Println("  ✓ Missing field handled correctly")
	} else {
		fmt.Println("  ✗ Missing field not handled correctly")
	}

	return nil
}

// Helper function to create HTTP client with timeout
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}
