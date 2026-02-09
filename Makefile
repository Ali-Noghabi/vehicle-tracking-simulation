# Makefile for Vehicle Tracking Route Service

.PHONY: all build build-service build-generator clean test test-service test-generator \
        run run-service run-generator run-local-osrm run-online-osrm \
        run-test-random run-test-permutation run-test-local-random run-test-local-permutation \
        deps fmt vet help

# Variables
BINARY_NAME=route-service
GENERATOR_NAME=route-generator
SIMULATION_NAME=simulation-service
BUILD_DIR=bin
SOURCE_DIR=cmd/route-service
GENERATOR_SOURCE_DIR=cmd/route-generator
SIMULATION_SOURCE_DIR=cmd/simulation-service
DEFAULT_PORT=8090
LOCAL_OSRM_PORT=5000

# Default target
all: build

help:
	@echo "🚗 Vehicle Tracking Route Service - Makefile Commands"
	@echo "======================================================"
	@echo ""
	@echo "📦 BUILD COMMANDS:"
	@echo "  all              - Build everything (service + generator)"
	@echo "  build            - Build everything (service + generator)"
	@echo "  build-service    - Build the route service binary"
	@echo "  build-generator  - Build the route generator binary"
	@echo "  build-simulation - Build vehicle tracking simulation service"
	@echo ""
	@echo "🧹 CLEANUP:"
	@echo "  clean            - Remove all build artifacts and generated files"
	@echo ""
	@echo "🧪 TESTING:"
	@echo "  test             - Run all tests"
	@echo "  test-service     - Test route service API endpoints"
	@echo "  test-generator   - Test route generator with sample config"
	@echo "  test-comprehensive - Run all tests (service + generator)"
	@echo "  test-local-random - Test local OSRM with random method"
	@echo ""
	@echo "▶️  RUN SERVICE (different providers):"
	@echo "  run              - Run service with default provider (openstreetmap) on port 8090"
	@echo "  run-service      - Run service with default provider on port 8090"
	@echo "  run-local-osrm   - Run service with local OSRM provider (http://localhost:5000)"
	@echo "  run-online-osrm  - Run service with online OSRM provider"
	@echo "  run-port         - Run service on custom port (PORT=8080)"
	@echo "  run-simulation   - Run vehicle tracking simulation (requires MQTT broker)"
	@echo ""
	@echo "🎲 RUN GENERATOR (different test scenarios):"
	@echo "  run-generator           - Run generator with main config.yaml"
	@echo "  run-test-random         - Test: Online OSRM + Random method"
	@echo "  run-test-permutation    - Test: Online OSRM + Permutation method"
	@echo "  run-test-local-random   - Test: Local OSRM + Random method"
	@echo "  run-test-local-permutation - Test: Local OSRM + Permutation method"
	@echo ""
	@echo "🔧 DEVELOPMENT:"
	@echo "  deps             - Download Go dependencies"
	@echo "  fmt              - Format Go code"
	@echo "  vet              - Vet Go code"
	@echo ""
	@echo "📚 EXAMPLES:"
	@echo "  make run-local-osrm"
	@echo "  make run-test-local-random"
	@echo "  make test-service"
	@echo "  make run-simulation"
	@echo "  make simulation-help"
	@echo ""

# Build everything
build: build-service build-generator build-simulation

# Build the route service
build-service:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(SOURCE_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build the route generator
build-generator:
	@echo "Building $(GENERATOR_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(GENERATOR_NAME) ./$(GENERATOR_SOURCE_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(GENERATOR_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf logs
	@rm -rf run
	@rm -rf generated_routes
	@echo "Clean complete"

# Run all tests
test: test-service test-generator

# Test route service API endpoints
test-service: build-service
	@echo "Testing route service API endpoints..."
	@chmod +x scripts/test_route_service.sh 2>/dev/null || true
	@./scripts/test_route_service.sh

# Test route generator
test-generator: build-generator
	@echo "Testing route generator..."
	@chmod +x scripts/test_route_generator.sh 2>/dev/null || true
	@./scripts/test_route_generator.sh

# Run comprehensive tests
test-comprehensive: build
	@echo "Running comprehensive tests..."
	@chmod +x scripts/run_comprehensive_tests.sh 2>/dev/null || true
	@./scripts/run_comprehensive_tests.sh

# Run the route service (default provider)
run: run-service

# Run the route service (default provider)
run-service: build-service
	@echo "Starting service with default provider (openstreetmap) on port $(DEFAULT_PORT)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -port $(DEFAULT_PORT)

# Run with custom port
run-port: build-service
	@echo "Starting service on port $(PORT)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -port $(PORT)

# Run the route service with local OSRM provider
run-local-osrm: build-service
	@echo "Starting service with local OSRM provider on port $(DEFAULT_PORT)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -port $(DEFAULT_PORT) -provider local-osrm -base-url http://localhost:$(LOCAL_OSRM_PORT)

# Run the route service with online OSRM provider
run-online-osrm: build-service
	@echo "Starting service with online OSRM provider on port $(DEFAULT_PORT)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -port $(DEFAULT_PORT) -provider openstreetmap

# Run the route generator
run-generator: build-generator
	@echo "Starting route generator with main config..."
	@./$(BUILD_DIR)/$(GENERATOR_NAME) -config config.yaml

# Test: Online OSRM + Random method
run-test-random: build-generator
	@echo "Testing: Online OSRM + Random method..."
	@./$(BUILD_DIR)/$(GENERATOR_NAME) -config scripts/test_config_online_random.yaml

# Test: Online OSRM + Permutation method
run-test-permutation: build-generator
	@echo "Testing: Online OSRM + Permutation method..."
	@./$(BUILD_DIR)/$(GENERATOR_NAME) -config scripts/test_config_online_permutation.yaml

# Test: Local OSRM + Random method
run-test-local-random: build-generator
	@echo "Testing: Local OSRM + Random method..."
	@./$(BUILD_DIR)/$(GENERATOR_NAME) -config scripts/test_config_local_random.yaml

# Test: Local OSRM + Permutation method
run-test-local-permutation: build-generator
	@echo "Testing: Local OSRM + Permutation method..."
	@./$(BUILD_DIR)/$(GENERATOR_NAME) -config scripts/test_config_local_permutation.yaml

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
vet:
	@echo "Checking code with go vet..."
	@go vet ./...

## Simulation
build-simulation: ## Build vehicle tracking simulation service
	@echo "Building $(SIMULATION_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(SIMULATION_NAME) ./$(SIMULATION_SOURCE_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(SIMULATION_NAME)"

run-simulation: build-simulation ## Run vehicle tracking simulation (requires MQTT broker)
	@echo "Starting vehicle tracking simulation..."
	@echo "Note: Requires MQTT broker running (e.g., mosquitto)"
	@./$(BUILD_DIR)/$(SIMULATION_NAME) -config cmd/simulation-service/config.yaml

simulation-help: ## Show simulation service help
	@echo "Vehicle Tracking Simulation Service"
	@echo ""
	@echo "To run simulation:"
	@echo "  1. Install MQTT broker: sudo apt install mosquitto mosquitto-clients"
	@echo "  2. Start MQTT broker: sudo systemctl start mosquitto"
	@echo "  3. Run simulation: make run-simulation"
	@echo ""
	@echo "To monitor telemetry:"
	@echo "  mosquitto_sub -t 'vehicle/telemetry' -v"
	@echo "  mosquitto_sub -t 'vehicle/telemetry_batch' -v"