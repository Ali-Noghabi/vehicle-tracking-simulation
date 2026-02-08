package config

import (
	"fmt"
	"os"
	
	"gopkg.in/yaml.v3"
)

// Location represents a geographic coordinate
type Location struct {
	Name string  `yaml:"name"`
	Lat  float64 `yaml:"lat"`
	Lng  float64 `yaml:"lng"`
}

// CountryBounds defines geographic boundaries
type CountryBounds struct {
	MinLat float64 `yaml:"min_lat"`
	MaxLat float64 `yaml:"max_lat"`
	MinLng float64 `yaml:"min_lng"`
	MaxLng float64 `yaml:"max_lng"`
}

// RouteServiceConfig defines route service connection settings
type RouteServiceConfig struct {
	BaseURL             string `yaml:"base_url"`
	TimeoutSeconds      int    `yaml:"timeout_seconds"`
	MaxConcurrentRequests int  `yaml:"max_concurrent_requests"`
}

// OutputConfig defines output file settings
type OutputConfig struct {
	Directory string `yaml:"directory"`
	Format    string `yaml:"format"`  // "json" or "binary"
	Compress  bool   `yaml:"compress"`
}

// Config is the main configuration structure
type Config struct {
	RouteGenerator struct {
		RouteCount   int                     `yaml:"route_count"`
		Method       string                  `yaml:"method"`  // "random" or "permutation"
		Country      string                  `yaml:"country"`
		CountryBounds map[string]CountryBounds `yaml:"country_bounds"`
		LocationSet  []Location              `yaml:"location_set"`
		RouteService RouteServiceConfig      `yaml:"route_service"`
		Output       OutputConfig            `yaml:"output"`
		RandomSeed   int64                   `yaml:"random_seed"`
	} `yaml:"route_generator"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.RouteGenerator.RouteCount <= 0 {
		return fmt.Errorf("route_count must be positive")
	}
	
	if c.RouteGenerator.Method != "random" && c.RouteGenerator.Method != "permutation" {
		return fmt.Errorf("method must be 'random' or 'permutation'")
	}
	
	if c.RouteGenerator.Method == "random" {
		if _, exists := c.RouteGenerator.CountryBounds[c.RouteGenerator.Country]; !exists {
			return fmt.Errorf("country bounds not defined for %s", c.RouteGenerator.Country)
		}
	}
	
	if c.RouteGenerator.Method == "permutation" {
		if len(c.RouteGenerator.LocationSet) < 2 {
			return fmt.Errorf("location_set must contain at least 2 locations for permutation method")
		}
	}
	
	if c.RouteGenerator.RouteService.MaxConcurrentRequests <= 0 {
		return fmt.Errorf("max_concurrent_requests must be positive")
	}
	
	if c.RouteGenerator.RouteService.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be positive")
	}
	
	return nil
}