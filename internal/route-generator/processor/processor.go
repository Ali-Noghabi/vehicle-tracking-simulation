package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"vehicle-tracking-simulation/internal/route-generator/config"
	"vehicle-tracking-simulation/internal/route-generator/generator"
	"vehicle-tracking-simulation/internal/route-service/models"
)

// RouteProcessor handles communication with the route service
type RouteProcessor struct {
	config *config.Config
	client *http.Client
}

// NewRouteProcessor creates a new route processor
func NewRouteProcessor(cfg *config.Config) *RouteProcessor {
	return &RouteProcessor{
		config: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.RouteGenerator.RouteService.TimeoutSeconds) * time.Second,
		},
	}
}

// ProcessRoute calls the route service to get route information with retry logic
func (p *RouteProcessor) ProcessRoute(ctx context.Context, req generator.RouteRequest) (*models.Route, error) {
	maxRetries := 3
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Prepare the request payload
		routeReq := models.RouteRequest{
			StartCoordinate: req.Start,
			EndCoordinate:   req.End,
			Profile:         req.Profile,
		}
		
		payload, err := json.Marshal(routeReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		
		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, "POST", 
			fmt.Sprintf("%s/api/v1/route", p.config.RouteGenerator.RouteService.BaseURL),
			bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		
		httpReq.Header.Set("Content-Type", "application/json")
		
		// Send request
		resp, err := p.client.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to send request: %w", attempt, err)
			if attempt < maxRetries {
				// Exponential backoff: 1s, 2s, 4s
				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
				fmt.Printf("Route %d attempt %d failed, retrying in %v: %v\n", req.ID, attempt, backoff, err)
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}
		
		// Read response
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to read response: %w", attempt, err)
			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
				fmt.Printf("Route %d attempt %d failed, retrying in %v: %v\n", req.ID, attempt, backoff, err)
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}
		
		// Check status code
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("attempt %d: route service returned error: %s - %s", attempt, resp.Status, string(body))
			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
				fmt.Printf("Route %d attempt %d failed with status %d, retrying in %v\n", req.ID, attempt, resp.StatusCode, backoff)
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}
		
		// Parse response
		var routeResp models.RouteResponse
		if err := json.Unmarshal(body, &routeResp); err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to parse response: %w", attempt, err)
			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
				fmt.Printf("Route %d attempt %d failed to parse, retrying in %v: %v\n", req.ID, attempt, backoff, err)
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}
		
		if routeResp.Code != "Ok" {
			// Don't retry on "NoRoute" errors - these are valid responses
			if routeResp.Code == "NoRoute" {
				fmt.Printf("Route %d: no route found (%.6f,%.6f -> %.6f,%.6f)\n", 
					req.ID, req.Start.Latitude, req.Start.Longitude, req.End.Latitude, req.End.Longitude)
				return nil, fmt.Errorf("no route found")
			}
			
			lastErr = fmt.Errorf("attempt %d: route service returned error: %s", attempt, routeResp.Message)
			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
				fmt.Printf("Route %d attempt %d returned error code %s, retrying in %v\n", req.ID, attempt, routeResp.Code, backoff)
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}
		
		if len(routeResp.Routes) == 0 {
			lastErr = fmt.Errorf("attempt %d: no route found", attempt)
			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
				fmt.Printf("Route %d attempt %d: no route found, retrying in %v\n", req.ID, attempt, backoff)
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}
		
		fmt.Printf("Route %d succeeded on attempt %d: %.6f,%.6f -> %.6f,%.6f\n", 
			req.ID, attempt, req.Start.Latitude, req.Start.Longitude, req.End.Latitude, req.End.Longitude)
		return &routeResp.Routes[0], nil
	}
	
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ProcessRoutes processes multiple routes in parallel
func (p *RouteProcessor) ProcessRoutes(ctx context.Context, requests []generator.RouteRequest) ([]generator.RouteResult, error) {
	gen := generator.NewGenerator(p.config)
	
	// Create processor function
	processorFunc := func(ctx context.Context, req generator.RouteRequest) (*models.Route, error) {
		return p.ProcessRoute(ctx, req)
	}
	
	// Process requests in parallel
	return gen.ProcessRequests(ctx, requests, processorFunc)
}