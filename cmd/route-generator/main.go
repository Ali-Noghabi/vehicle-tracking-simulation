package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	
	"vehicle-tracking-simulation/internal/route-generator/config"
	"vehicle-tracking-simulation/internal/route-generator/generator"
	"vehicle-tracking-simulation/internal/route-generator/processor"
	"vehicle-tracking-simulation/internal/route-generator/storage"
)

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()
	
	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	log.Printf("Starting route generator with configuration: %s", *configPath)
	log.Printf("Method: %s, Route count: %d", cfg.RouteGenerator.Method, cfg.RouteGenerator.RouteCount)
	
	// Create storage
	storage, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	
	log.Printf("Output directory: %s", storage.GetOutputDir())
	
	// Create generator
	gen := generator.NewGenerator(cfg)
	
	// Generate route requests
	requests, err := gen.GenerateRouteRequests()
	if err != nil {
		log.Fatalf("Failed to generate route requests: %v", err)
	}
	
	log.Printf("Generated %d route requests", len(requests))
	
	// Create processor
	routeProcessor := processor.NewRouteProcessor(cfg)
	
	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 
		time.Duration(cfg.RouteGenerator.RouteService.TimeoutSeconds+60)*time.Second)
	defer cancel()
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start processing in a goroutine
	resultsChan := make(chan []generator.RouteResult, 1)
	errorChan := make(chan error, 1)
	
	go func() {
		results, err := routeProcessor.ProcessRoutes(ctx, requests)
		if err != nil {
			errorChan <- err
			return
		}
		resultsChan <- results
	}()
	
	// Wait for results or signals
	var results []generator.RouteResult
	select {
	case <-sigChan:
		log.Println("Received shutdown signal, stopping...")
		cancel()
		// Wait a bit for cleanup
		time.Sleep(2 * time.Second)
		return
	case err := <-errorChan:
		log.Fatalf("Failed to process routes: %v", err)
	case results = <-resultsChan:
		log.Println("Route processing completed")
	}
	
	// Save results
	if err := saveResults(results, requests, storage, cfg); err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}
	
	log.Println("Route generation completed successfully")
}

func saveResults(results []generator.RouteResult, requests []generator.RouteRequest, 
	storage *storage.Storage, cfg *config.Config) error {
	
	log.Println("Saving route results...")
	
	// Create a map of requests by ID for easy lookup
	requestMap := make(map[int]generator.RouteRequest)
	for _, req := range requests {
		requestMap[req.ID] = req
	}
	
	// Count successful and failed routes
	successful := 0
	failed := 0
	
	// Save routes in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, len(results))
	
	for _, result := range results {
		wg.Add(1)
		go func(result generator.RouteResult) {
			defer wg.Done()
			
			req, exists := requestMap[result.ID]
			if !exists {
				errorChan <- fmt.Errorf("request not found for ID %d", result.ID)
				return
			}
			
			if err := storage.SaveRoute(result, req); err != nil {
				errorChan <- fmt.Errorf("failed to save route %d: %w", result.ID, err)
				return
			}
			
			if result.Error == nil {
				successful++
			} else {
				failed++
			}
		}(result)
	}
	
	// Wait for all saves to complete
	wg.Wait()
	close(errorChan)
	
	// Check for errors
	var saveErrors []error
	for err := range errorChan {
		saveErrors = append(saveErrors, err)
	}
	
	if len(saveErrors) > 0 {
		log.Printf("Encountered %d errors while saving routes", len(saveErrors))
		for _, err := range saveErrors {
			log.Printf("Save error: %v", err)
		}
	}
	
	// Save metadata
	if err := storage.SaveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}
	
	// Save summary
	total := len(results)
	duration := time.Duration(cfg.RouteGenerator.RouteService.TimeoutSeconds) * time.Second
	if err := storage.SaveSummary(total, successful, failed, duration); err != nil {
		return fmt.Errorf("failed to save summary: %w", err)
	}
	
	log.Printf("Saved %d routes (%d successful, %d failed)", total, successful, failed)
	log.Printf("Success rate: %.2f%%", float64(successful)/float64(total)*100)
	
	return nil
}