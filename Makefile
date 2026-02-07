# Makefile for Vehicle Tracking Route Service

.PHONY: build clean test run help

# Variables
BINARY_NAME=route-service
BUILD_DIR=bin
SOURCE_DIR=cmd/route-service

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the service binary"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  run      - Run the service on default port (8090)"
	@echo "  help     - Show this help message"

# Build the service
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(SOURCE_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf logs
	@rm -rf run
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@cd tests && go run test_route_service.go -port 8090

# Run the service
run: build
	@echo "Starting service on port 8090..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -port 8090

# Run with custom port
run-port: build
	@echo "Starting service on port $(PORT)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -port $(PORT)

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