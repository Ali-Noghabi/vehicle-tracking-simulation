package storage

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"vehicle-tracking-simulation/internal/route-generator/config"
	"vehicle-tracking-simulation/internal/route-generator/generator"
	"vehicle-tracking-simulation/internal/route-service/models"
)

// RouteMetadata contains metadata about a generated route
type RouteMetadata struct {
	ID           int       `json:"id"`
	GeneratedAt  time.Time `json:"generated_at"`
	StartLat     float64   `json:"start_lat"`
	StartLng     float64   `json:"start_lng"`
	EndLat       float64   `json:"end_lat"`
	EndLng       float64   `json:"end_lng"`
	Profile      string    `json:"profile"`
	Distance     float64   `json:"distance"`
	Duration     float64   `json:"duration"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// RouteData contains the complete route data for simulation
type RouteData struct {
	Metadata RouteMetadata `json:"metadata"`
	Route    *models.Route `json:"route,omitempty"`
}

// Storage handles saving route data for future simulation
type Storage struct {
	config     *config.Config
	outputDir  string
	fileMutex  sync.Mutex
	metadata   []RouteMetadata
}

// NewStorage creates a new storage instance
func NewStorage(cfg *config.Config) (*Storage, error) {
	outputDir := cfg.RouteGenerator.Output.Directory
	
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	
	return &Storage{
		config:    cfg,
		outputDir: outputDir,
		metadata:  make([]RouteMetadata, 0),
	}, nil
}

// SaveRoute saves a single route result
func (s *Storage) SaveRoute(result generator.RouteResult, request generator.RouteRequest) error {
	s.fileMutex.Lock()
	defer s.fileMutex.Unlock()
	
	// Create metadata
	metadata := RouteMetadata{
		ID:          request.ID,
		GeneratedAt: time.Now(),
		StartLat:    request.Start.Latitude,
		StartLng:    request.Start.Longitude,
		EndLat:      request.End.Latitude,
		EndLng:      request.End.Longitude,
		Profile:     request.Profile,
		Success:     result.Error == nil,
	}
	
	if result.Error != nil {
		metadata.ErrorMessage = result.Error.Error()
	} else if result.Route != nil {
		metadata.Distance = result.Route.Distance
		metadata.Duration = result.Route.Duration
	}
	
	// Create route data
	routeData := RouteData{
		Metadata: metadata,
		Route:    result.Route,
	}
	
	// Save individual route file
	if err := s.saveIndividualRoute(routeData); err != nil {
		return fmt.Errorf("failed to save individual route: %w", err)
	}
	
	// Add to metadata collection
	s.metadata = append(s.metadata, metadata)
	
	return nil
}

// saveIndividualRoute saves a single route to its own file
func (s *Storage) saveIndividualRoute(routeData RouteData) error {
	// Create filename
	filename := fmt.Sprintf("route_%06d.json", routeData.Metadata.ID)
	if s.config.RouteGenerator.Output.Compress {
		filename += ".gz"
	}
	
	filePath := filepath.Join(s.outputDir, filename)
	
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	// Create writer (with compression if enabled)
	var writer interface {
		Write([]byte) (int, error)
		Close() error
	}
	
	if s.config.RouteGenerator.Output.Compress {
		gzipWriter := gzip.NewWriter(file)
		writer = gzipWriter
		defer gzipWriter.Close()
	} else {
		writer = file
	}
	
	// Encode and write JSON
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(routeData); err != nil {
		return fmt.Errorf("failed to encode route data: %w", err)
	}
	
	return nil
}

// SaveMetadata saves the metadata index file
func (s *Storage) SaveMetadata() error {
	s.fileMutex.Lock()
	defer s.fileMutex.Unlock()
	
	// Create metadata file path
	filename := "metadata.json"
	if s.config.RouteGenerator.Output.Compress {
		filename += ".gz"
	}
	
	filePath := filepath.Join(s.outputDir, filename)
	
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()
	
	// Create writer (with compression if enabled)
	var writer interface {
		Write([]byte) (int, error)
		Close() error
	}
	
	if s.config.RouteGenerator.Output.Compress {
		gzipWriter := gzip.NewWriter(file)
		writer = gzipWriter
		defer gzipWriter.Close()
	} else {
		writer = file
	}
	
	// Encode and write metadata
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(s.metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}
	
	return nil
}

// SaveSummary saves a summary of the generation process
func (s *Storage) SaveSummary(totalRoutes int, successfulRoutes int, failedRoutes int, duration time.Duration) error {
	summary := struct {
		TotalRoutes      int           `json:"total_routes"`
		SuccessfulRoutes int           `json:"successful_routes"`
		FailedRoutes     int           `json:"failed_routes"`
		SuccessRate      float64       `json:"success_rate"`
		Duration         time.Duration `json:"duration"`
		GeneratedAt      time.Time     `json:"generated_at"`
		Method           string        `json:"method"`
		Country          string        `json:"country,omitempty"`
		LocationCount    int           `json:"location_count,omitempty"`
	}{
		TotalRoutes:      totalRoutes,
		SuccessfulRoutes: successfulRoutes,
		FailedRoutes:     failedRoutes,
		SuccessRate:      float64(successfulRoutes) / float64(totalRoutes) * 100,
		Duration:         duration,
		GeneratedAt:      time.Now(),
		Method:           s.config.RouteGenerator.Method,
	}
	
	if s.config.RouteGenerator.Method == "random" {
		summary.Country = s.config.RouteGenerator.Country
	} else {
		summary.LocationCount = len(s.config.RouteGenerator.LocationSet)
	}
	
	// Create summary file path
	filename := "summary.json"
	if s.config.RouteGenerator.Output.Compress {
		filename += ".gz"
	}
	
	filePath := filepath.Join(s.outputDir, filename)
	
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %w", err)
	}
	defer file.Close()
	
	// Create writer (with compression if enabled)
	var writer interface {
		Write([]byte) (int, error)
		Close() error
	}
	
	if s.config.RouteGenerator.Output.Compress {
		gzipWriter := gzip.NewWriter(file)
		writer = gzipWriter
		defer gzipWriter.Close()
	} else {
		writer = file
	}
	
	// Encode and write summary
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(summary); err != nil {
		return fmt.Errorf("failed to encode summary: %w", err)
	}
	
	return nil
}

// GetOutputDir returns the output directory path
func (s *Storage) GetOutputDir() string {
	return s.outputDir
}