package generator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	
	"vehicle-tracking-simulation/internal/route-generator/config"
	"vehicle-tracking-simulation/internal/route-service/models"
)

// RouteRequest represents a route request with start and end coordinates
type RouteRequest struct {
	ID        int
	Start     models.Coordinate
	End       models.Coordinate
	Profile   string
}

// RouteResult represents a route result with the route data
type RouteResult struct {
	ID     int
	Route  *models.Route
	Error  error
}

// Generator handles route generation
type Generator struct {
	config *config.Config
	rand   *rand.Rand
}

// NewGenerator creates a new route generator
func NewGenerator(cfg *config.Config) *Generator {
	source := rand.NewSource(cfg.RouteGenerator.RandomSeed)
	return &Generator{
		config: cfg,
		rand:   rand.New(source),
	}
}

// GenerateRouteRequests creates route requests based on the configured method
func (g *Generator) GenerateRouteRequests() ([]RouteRequest, error) {
	count := g.config.RouteGenerator.RouteCount
	
	switch g.config.RouteGenerator.Method {
	case "random":
		return g.generateRandomRequests(count)
	case "permutation":
		return g.generatePermutationRequests(count)
	default:
		return nil, fmt.Errorf("unknown generation method: %s", g.config.RouteGenerator.Method)
	}
}

// generateRandomRequests generates random coordinate pairs within the country bounds
func (g *Generator) generateRandomRequests(count int) ([]RouteRequest, error) {
	bounds, exists := g.config.RouteGenerator.CountryBounds[g.config.RouteGenerator.Country]
	if !exists {
		return nil, fmt.Errorf("country bounds not found for %s", g.config.RouteGenerator.Country)
	}
	
	requests := make([]RouteRequest, count)
	
	for i := 0; i < count; i++ {
		start := models.Coordinate{
			Latitude:  bounds.MinLat + g.rand.Float64()*(bounds.MaxLat-bounds.MinLat),
			Longitude: bounds.MinLng + g.rand.Float64()*(bounds.MaxLng-bounds.MinLng),
		}
		
		end := models.Coordinate{
			Latitude:  bounds.MinLat + g.rand.Float64()*(bounds.MaxLat-bounds.MinLat),
			Longitude: bounds.MinLng + g.rand.Float64()*(bounds.MaxLng-bounds.MinLng),
		}
		
		requests[i] = RouteRequest{
			ID:      i + 1,
			Start:   start,
			End:     end,
			Profile: "car", // Default profile
		}
	}
	
	return requests, nil
}

// generatePermutationRequests generates route requests by permuting location pairs
func (g *Generator) generatePermutationRequests(count int) ([]RouteRequest, error) {
	locations := g.config.RouteGenerator.LocationSet
	if len(locations) < 2 {
		return nil, fmt.Errorf("need at least 2 locations for permutation")
	}
	
	requests := make([]RouteRequest, count)
	
	// Generate all possible pairs
	allPairs := make([][2]int, 0)
	for i := 0; i < len(locations); i++ {
		for j := 0; j < len(locations); j++ {
			if i != j { // Don't create routes from a location to itself
				allPairs = append(allPairs, [2]int{i, j})
			}
		}
	}
	
	if len(allPairs) == 0 {
		return nil, fmt.Errorf("no valid location pairs found")
	}
	
	// Shuffle the pairs
	g.rand.Shuffle(len(allPairs), func(i, j int) {
		allPairs[i], allPairs[j] = allPairs[j], allPairs[i]
	})
	
	// Generate requests by cycling through shuffled pairs
	for i := 0; i < count; i++ {
		pair := allPairs[i%len(allPairs)]
		startLoc := locations[pair[0]]
		endLoc := locations[pair[1]]
		
		requests[i] = RouteRequest{
			ID:    i + 1,
			Start: models.Coordinate{Latitude: startLoc.Lat, Longitude: startLoc.Lng},
			End:   models.Coordinate{Latitude: endLoc.Lat, Longitude: endLoc.Lng},
			Profile: "car",
		}
	}
	
	return requests, nil
}

// ProcessRequests processes route requests in parallel and returns results
func (g *Generator) ProcessRequests(ctx context.Context, requests []RouteRequest, processor func(context.Context, RouteRequest) (*models.Route, error)) ([]RouteResult, error) {
	cfg := g.config.RouteGenerator.RouteService
	
	// Create channels for work and results
	workChan := make(chan RouteRequest, len(requests))
	resultChan := make(chan RouteResult, len(requests))
	
	// Send all requests to work channel
	for _, req := range requests {
		workChan <- req
	}
	close(workChan)
	
	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < cfg.MaxConcurrentRequests; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for req := range workChan {
				select {
				case <-ctx.Done():
					return
				default:
					route, err := processor(ctx, req)
					resultChan <- RouteResult{
						ID:    req.ID,
						Route: route,
						Error: err,
					}
				}
			}
		}(i)
	}
	
	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results
	results := make([]RouteResult, 0, len(requests))
	for result := range resultChan {
		results = append(results, result)
	}
	
	return results, nil
}

// GetRandomProfile returns a random routing profile
func (g *Generator) GetRandomProfile() string {
	profiles := []string{"car", "bike", "foot"}
	return profiles[g.rand.Intn(len(profiles))]
}