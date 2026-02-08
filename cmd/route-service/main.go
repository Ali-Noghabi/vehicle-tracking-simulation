package main

import (
	"flag"
	"log"
	"net/http"

	"vehicle-tracking-simulation/internal/route-service/api"
	"vehicle-tracking-simulation/internal/route-service/provider"
	"vehicle-tracking-simulation/internal/route-service/service"
)

func main() {
	// Command line flags
	providerType := flag.String("provider", "openstreetmap", "Routing provider: openstreetmap, local-osrm, google, mapbox, here")
	apiKey := flag.String("api-key", "", "API key for the routing provider (if required)")
	baseURL := flag.String("base-url", "", "Custom base URL for the routing provider")
	port := flag.String("port", "8080", "Port to listen on")
	timeout := flag.Int("timeout", 10, "Request timeout in seconds")

	flag.Parse()

	// Configure the routing provider
	config := provider.RouteFinderConfig{
		ProviderType: *providerType,
		APIKey:       *apiKey,
		BaseURL:      *baseURL,
		Timeout:      *timeout,
	}

	// Create the provider
	routingProvider, err := provider.NewProvider(config)
	if err != nil {
		log.Fatalf("Failed to create routing provider: %v", err)
	}

	log.Printf("Using routing provider: %s", routingProvider.ProviderName())

	// Create the route finder service
	routeFinder := service.NewRouteFinder(routingProvider)

	// Create HTTP handler
	handler := api.NewHandler(routeFinder)

	// Start server
	addr := ":" + *port
	log.Printf("Starting route service on %s", addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health")
	log.Printf("  GET  /api/v1/provider")
	log.Printf("  POST /api/v1/route")
	log.Printf("  POST /api/v1/route/waypoints")

	if err := http.ListenAndServe(addr, handler.GetRouter()); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
