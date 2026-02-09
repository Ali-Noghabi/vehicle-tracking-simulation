package main

import (
	"encoding/json"
	"flag"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/yaml.v3"
)

// Route represents a generated route file structure
type Route struct {
	Metadata struct {
		ID          int     `json:"id"`
		GeneratedAt string  `json:"generated_at"`
		StartLat    float64 `json:"start_lat"`
		StartLng    float64 `json:"start_lng"`
		EndLat      float64 `json:"end_lat"`
		EndLng      float64 `json:"end_lng"`
		Profile     string  `json:"profile"`
		Distance    float64 `json:"distance"` // meters
		Duration    float64 `json:"duration"` // seconds
		Success     bool    `json:"success"`
	} `json:"metadata"`
	Route struct {
		Geometry string `json:"geometry"`
		Legs     []struct {
			Steps []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Geometry string  `json:"geometry"`
			} `json:"steps"`
		} `json:"legs"`
	} `json:"route"`
}

// Telemetry represents MQTT telemetry data
type Telemetry struct {
	VehicleID int     `json:"vehicle_id"`
	Timestamp int64   `json:"timestamp"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Speed     float64 `json:"spd"`
	Heading   float64 `json:"hdg"`
	Altitude  float64 `json:"alt"`
	Accuracy  float64 `json:"acc"`
	Battery   float64 `json:"battery"`
	Signal    float64 `json:"signal"`
}

// validate ensures all telemetry values are valid numbers
func (t *Telemetry) validate() {
	if math.IsNaN(t.Lat) || math.IsInf(t.Lat, 0) {
		t.Lat = 0.0
	}
	if math.IsNaN(t.Lon) || math.IsInf(t.Lon, 0) {
		t.Lon = 0.0
	}
	if math.IsNaN(t.Speed) || math.IsInf(t.Speed, 0) {
		t.Speed = 0.0
	}
	if math.IsNaN(t.Heading) || math.IsInf(t.Heading, 0) {
		t.Heading = 0.0
	}
	if math.IsNaN(t.Altitude) || math.IsInf(t.Altitude, 0) {
		t.Altitude = 100.0
	}
	if math.IsNaN(t.Accuracy) || math.IsInf(t.Accuracy, 0) {
		t.Accuracy = 10.0
	}
	if math.IsNaN(t.Battery) || math.IsInf(t.Battery, 0) {
		t.Battery = 90.0
	}
	if math.IsNaN(t.Signal) || math.IsInf(t.Signal, 0) {
		t.Signal = 85.0
	}
}

// VehicleSimulator simulates a vehicle moving along a route
type VehicleSimulator struct {
	VehicleID     int
	Route         *Route
	RouteIterator *RouteIterator
	StartTime     time.Time
	SpeedRange    [2]float64 // min and max speed in m/s
}

// Config holds simulation configuration
type Config struct {
	MQTT struct {
		Broker   string `yaml:"broker"`
		Topic    string `yaml:"topic"`
		ClientID string `yaml:"client_id"`
		QoS      int    `yaml:"qos"`
		Retain   bool   `yaml:"retain"`
	} `yaml:"mqtt"`

	Simulation struct {
		UpdateInterval  string  `yaml:"update_interval"`
		SimulationSpeed float64 `yaml:"simulation_speed"`
		RoutesPath      string  `yaml:"routes_path"`
		SpeedVariation  float64 `yaml:"speed_variation"`

		AltitudeRange [2]float64 `yaml:"altitude_range"`
		AccuracyRange [2]float64 `yaml:"accuracy_range"`
		BatteryRange  [2]float64 `yaml:"battery_range"`
		SignalRange   [2]float64 `yaml:"signal_range"`
	} `yaml:"simulation"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
}

// parseDuration parses duration string with default fallback
func parseDuration(durStr string, defaultDur time.Duration) time.Duration {
	if dur, err := time.ParseDuration(durStr); err == nil {
		return dur
	}
	return defaultDur
}

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load routes
	routes, err := loadRoutes(config.Simulation.RoutesPath)
	if err != nil {
		log.Fatalf("Failed to load routes: %v", err)
	}
	log.Printf("Loaded %d routes from %s", len(routes), config.Simulation.RoutesPath)

	// Connect to MQTT broker
	client := connectMQTT(config.MQTT.Broker, config.MQTT.ClientID)
	defer client.Disconnect(250)

	// Create vehicle simulators
	simulators := make([]*VehicleSimulator, 0, len(routes))
	for _, route := range routes {
		if !route.Metadata.Success {
			continue
		}

		simulator := &VehicleSimulator{
			VehicleID: route.Metadata.ID,
			Route:     route,
			StartTime: time.Now(),
		}

		// Calculate speed range based on route distance and duration
		avgSpeed := 0.0
		if route.Metadata.Duration > 0 {
			avgSpeed = route.Metadata.Distance / route.Metadata.Duration // m/s
		} else {
			avgSpeed = 20.0 // Default average speed if duration is 0
		}

		// Ensure avgSpeed is a valid number
		if math.IsNaN(avgSpeed) || math.IsInf(avgSpeed, 0) || avgSpeed <= 0 {
			avgSpeed = 20.0 // Default average speed
		}

		variation := config.Simulation.SpeedVariation
		simulator.SpeedRange = [2]float64{
			avgSpeed * (1 - variation), // min speed
			avgSpeed * (1 + variation), // max speed
		}

		simulators = append(simulators, simulator)
		log.Printf("Created simulator for vehicle %d (distance: %.0fm, duration: %.0fs, avg speed: %.1f m/s, range: %.1f-%.1f m/s)",
			simulator.VehicleID, route.Metadata.Distance, route.Metadata.Duration, avgSpeed,
			simulator.SpeedRange[0], simulator.SpeedRange[1])
	}

	// Create batch sender
	batchSender := NewTelemetryBatchSender(10, 30*time.Second)

	// Start simulation
	log.Printf("Starting simulation of %d vehicles", len(simulators))
	updateInterval := parseDuration(config.Simulation.UpdateInterval, 5*time.Second)
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for range ticker.C {
		simulationTime := time.Now()
		var telemetries []Telemetry

		for _, simulator := range simulators {
			telemetry := simulator.UpdateWithRouteIterator(simulationTime)
			if telemetry != nil {
				// Apply configuration ranges
				telemetry.Altitude = config.Simulation.AltitudeRange[0] +
					rand.Float64()*(config.Simulation.AltitudeRange[1]-config.Simulation.AltitudeRange[0])
				telemetry.Accuracy = config.Simulation.AccuracyRange[0] +
					rand.Float64()*(config.Simulation.AccuracyRange[1]-config.Simulation.AccuracyRange[0])
				telemetry.Battery = config.Simulation.BatteryRange[0] +
					rand.Float64()*(config.Simulation.BatteryRange[1]-config.Simulation.BatteryRange[0])
				telemetry.Signal = config.Simulation.SignalRange[0] +
					rand.Float64()*(config.Simulation.SignalRange[1]-config.Simulation.SignalRange[0])

				// Validate all values are valid numbers
				telemetry.validate()

				telemetries = append(telemetries, *telemetry)
			}
		}

		// Send individual telemetry
		for _, telemetry := range telemetries {
			sendTelemetry(client, config.MQTT.Topic, &telemetry)

			// Also add to batch
			if ready, batch := batchSender.AddTelemetry(config.MQTT.Topic+"_batch", telemetry); ready {
				SendBatchTelemetry(client, config.MQTT.Topic+"_batch", batch)
			}
		}

		log.Printf("Sent %d telemetry updates at %s", len(telemetries), simulationTime.Format("15:04:05"))
	}
}

func loadRoutes(path string) ([]*Route, error) {
	files, err := filepath.Glob(filepath.Join(path, "route_*.json"))
	if err != nil {
		return nil, err
	}

	routes := make([]*Route, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", file, err)
			continue
		}

		var route Route
		if err := json.Unmarshal(data, &route); err != nil {
			log.Printf("Warning: Failed to parse %s: %v", file, err)
			continue
		}

		routes = append(routes, &route)
	}

	return routes, nil
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func connectMQTT(broker, clientID string) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}
	log.Printf("Connected to MQTT broker at %s", broker)
	return client
}

func (v *VehicleSimulator) update(currentTime time.Time) *Telemetry {
	return v.UpdateWithRouteIterator(currentTime)
}

func sendTelemetry(client mqtt.Client, topic string, telemetry *Telemetry) {
	data, err := json.Marshal(telemetry)
	if err != nil {
		log.Printf("Failed to marshal telemetry: %v", err)
		return
	}

	token := client.Publish(topic, 0, false, data)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to publish telemetry: %v", token.Error())
	}
}
