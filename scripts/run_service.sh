#!/bin/bash

# Vehicle Tracking Route Service Runner
# Script to start, stop, and manage the route service

set -e

# Configuration
SERVICE_NAME="vehicle-route-service"
SERVICE_BINARY="../bin/route-service"
BUILD_DIR="../bin"
LOG_FILE="../logs/route-service.log"
PID_FILE="../run/route-service.pid"
DEFAULT_PORT=8090
DEFAULT_TIMEOUT=30

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Create necessary directories
create_directories() {
    mkdir -p "$BUILD_DIR"
    mkdir -p "$(dirname "$LOG_FILE")"
    mkdir -p "$(dirname "$PID_FILE")"
}

# Build the service
build_service() {
    print_info "Building route service..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go to build the service."
        exit 1
    fi
    
    go build -o "$SERVICE_BINARY" ../cmd/route-service
    
    if [ $? -eq 0 ]; then
        print_success "Service built successfully: $SERVICE_BINARY"
    else
        print_error "Failed to build service"
        exit 1
    fi
}

# Start the service
start_service() {
    local port=${1:-$DEFAULT_PORT}
    local timeout=${2:-$DEFAULT_TIMEOUT}
    local provider=${3:-"openstreetmap"}
    local base_url=${4:-""}
    local api_key=${5:-""}
    
    print_info "Starting $SERVICE_NAME on port $port..."
    
    # Check if service is already running
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            print_warning "Service is already running with PID $pid"
            return 0
        else
            print_warning "Stale PID file found, removing..."
            rm -f "$PID_FILE"
        fi
    fi
    
    # Build if binary doesn't exist
    if [ ! -f "$SERVICE_BINARY" ]; then
        print_warning "Service binary not found, building..."
        build_service
    fi
    
    # Create directories
    create_directories
    
    # Build command
    local cmd="$SERVICE_BINARY -port $port -timeout $timeout -provider $provider"
    
    if [ -n "$base_url" ]; then
        cmd="$cmd -base-url $base_url"
    fi
    
    if [ -n "$api_key" ]; then
        cmd="$cmd -api-key $api_key"
    fi
    
    # Start service in background
    nohup $cmd >> "$LOG_FILE" 2>&1 &
    local pid=$!
    
    # Save PID
    echo $pid > "$PID_FILE"
    
    print_info "Service started with PID $pid"
    print_info "Logs: $LOG_FILE"
    
    # Wait a bit and check if it's running
    sleep 2
    if kill -0 "$pid" 2>/dev/null; then
        print_success "Service is running successfully"
        print_info "Available endpoints:"
        print_info "  Health:      http://localhost:$port/health"
        print_info "  Provider:    http://localhost:$port/api/v1/provider"
        print_info "  Find route:  http://localhost:$port/api/v1/route"
        print_info "  Waypoints:   http://localhost:$port/api/v1/route/waypoints"
    else
        print_error "Service failed to start. Check logs: $LOG_FILE"
        tail -20 "$LOG_FILE"
        return 1
    fi
}

# Stop the service
stop_service() {
    print_info "Stopping $SERVICE_NAME..."
    
    if [ ! -f "$PID_FILE" ]; then
        print_warning "No PID file found. Service may not be running."
        return 0
    fi
    
    local pid=$(cat "$PID_FILE")
    
    if kill -0 "$pid" 2>/dev/null; then
        kill "$pid"
        
        # Wait for process to terminate
        local timeout=10
        local count=0
        while kill -0 "$pid" 2>/dev/null && [ $count -lt $timeout ]; do
            sleep 1
            count=$((count + 1))
        done
        
        if kill -0 "$pid" 2>/dev/null; then
            print_warning "Service did not stop gracefully, forcing..."
            kill -9 "$pid"
        fi
        
        rm -f "$PID_FILE"
        print_success "Service stopped"
    else
        print_warning "Service with PID $pid is not running"
        rm -f "$PID_FILE"
    fi
}

# Restart the service
restart_service() {
    stop_service
    sleep 2
    start_service "$@"
}

# Check service status
status_service() {
    print_info "Checking $SERVICE_NAME status..."
    
    if [ ! -f "$PID_FILE" ]; then
        print_error "Service is not running (no PID file)"
        return 1
    fi
    
    local pid=$(cat "$PID_FILE")
    
    if kill -0 "$pid" 2>/dev/null; then
        print_success "Service is running with PID $pid"
        
        # Try to get health status
        local port=$(ps -p "$pid" -o args= | grep -o '\-port [0-9]*' | awk '{print $2}')
        if [ -z "$port" ]; then
            port=$DEFAULT_PORT
        fi
        
        print_info "Checking health endpoint..."
        if command -v curl &> /dev/null; then
            if curl -s "http://localhost:$port/health" > /dev/null; then
                print_success "Health endpoint is responding"
            else
                print_warning "Health endpoint is not responding"
            fi
        fi
        
        # Show recent logs
        print_info "Recent logs:"
        tail -5 "$LOG_FILE" 2>/dev/null || print_warning "No log file found"
    else
        print_error "Service is not running (stale PID: $pid)"
        rm -f "$PID_FILE"
        return 1
    fi
}

# Show logs
show_logs() {
    if [ ! -f "$LOG_FILE" ]; then
        print_error "Log file not found: $LOG_FILE"
        return 1
    fi
    
    local lines=${1:-20}
    
    print_info "Showing last $lines lines of logs:"
    echo "=============================================="
    tail -n "$lines" "$LOG_FILE"
    echo "=============================================="
    print_info "Full log file: $LOG_FILE"
}

# Clean up (remove build artifacts, logs, etc.)
clean_service() {
    print_info "Cleaning up..."
    
    stop_service
    
    # Remove build artifacts
    if [ -f "$SERVICE_BINARY" ]; then
        rm -f "$SERVICE_BINARY"
        print_success "Removed binary: $SERVICE_BINARY"
    fi
    
    # Remove logs
    if [ -f "$LOG_FILE" ]; then
        rm -f "$LOG_FILE"
        print_success "Removed log file: $LOG_FILE"
    fi
    
    # Remove PID file
    if [ -f "$PID_FILE" ]; then
        rm -f "$PID_FILE"
        print_success "Removed PID file: $PID_FILE"
    fi
    
    # Remove directories if empty
    rmdir "$(dirname "$LOG_FILE")" 2>/dev/null || true
    rmdir "$(dirname "$PID_FILE")" 2>/dev/null || true
    rmdir "$BUILD_DIR" 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Show help
show_help() {
    echo "Vehicle Tracking Route Service Manager"
    echo "======================================"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  start [port] [timeout] [provider] [base-url] [api-key]  Start the service"
    echo "  stop                                                    Stop the service"
    echo "  restart [port] [timeout] [provider] [base-url] [api-key] Restart the service"
    echo "  status                                                  Check service status"
    echo "  logs [lines]                                            Show logs (default: 20 lines)"
    echo "  build                                                   Build the service"
    echo "  clean                                                   Clean up artifacts and stop service"
    echo "  help                                                    Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 start                     # Start on default port 8090"
    echo "  $0 start 8080 10             # Start on port 8080 with 10s timeout"
    echo "  $0 start 8081 30 openstreetmap https://your-osrm-instance.com"
    echo "  $0 stop                      # Stop the service"
    echo "  $0 status                    # Check if service is running"
    echo "  $0 logs 50                   # Show last 50 lines of logs"
    echo ""
    echo "Default values:"
    echo "  Port: $DEFAULT_PORT"
    echo "  Timeout: ${DEFAULT_TIMEOUT}s"
    echo "  Provider: openstreetmap"
    echo ""
}

# Main execution
main() {
    local command=${1:-"help"}
    
    case $command in
        start)
            shift
            start_service "$@"
            ;;
        stop)
            stop_service
            ;;
        restart)
            shift
            restart_service "$@"
            ;;
        status)
            status_service
            ;;
        logs)
            shift
            show_logs "$@"
            ;;
        build)
            build_service
            ;;
        clean)
            clean_service
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
