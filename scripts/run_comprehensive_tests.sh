#!/bin/bash

# Comprehensive Test Runner for Vehicle Tracking Route Generator
# Tests all combinations: online/local OSRM × random/permutation methods

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "\n${BLUE}══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check dependencies
check_dependencies() {
    if ! command -v go &> /dev/null; then
        print_error "Go is required but not installed"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi
}

# Build the project
build_project() {
    print_info "Building project..."
    if make build-service build-generator; then
        print_success "Project built successfully"
    else
        print_error "Failed to build project"
        exit 1
    fi
}

# Start route service with specified provider
start_service() {
    local provider=$1
    local port=$2
    
    print_info "Starting route service with $provider provider on port $port..."
    
    # Kill any existing service on this port
    pkill -f "route-service.*-port $port" 2>/dev/null || true
    
    # Start service in background
    ./bin/route-service -port "$port" -provider "$provider" > "service_${provider}_${port}.log" 2>&1 &
    SERVICE_PID=$!
    
    # Wait for service to start
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "http://localhost:$port/health" > /dev/null; then
            print_success "Service started with $provider provider"
            return 0
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            print_error "Service did not start within $max_attempts seconds"
            return 1
        fi
        
        attempt=$((attempt + 1))
        sleep 1
    done
}

# Stop route service
stop_service() {
    if [ -n "$SERVICE_PID" ]; then
        print_info "Stopping route service..."
        kill $SERVICE_PID 2>/dev/null || true
        wait $SERVICE_PID 2>/dev/null || true
        SERVICE_PID=""
    fi
}

# Run route generator test
run_generator_test() {
    local config_file=$1
    local test_name=$2
    
    print_info "Running $test_name..."
    
    # Clean previous results
    rm -rf "./test_results" 2>/dev/null || true
    
    # Run generator
    local start_time=$(date +%s)
    if ./bin/route-generator -config "$config_file"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        # Check results
        local output_dir=$(grep -A2 "output:" "$config_file" | grep "directory:" | awk '{print $2}' | tr -d '"')
        if [ -d "$output_dir" ]; then
            local total_files=$(find "$output_dir" -name "*.json" | wc -l)
            local success_files=$(find "$output_dir" -name "*.json" -exec grep -l '"code":"Ok"' {} \; | wc -l)
            local success_rate=$((success_files * 100 / total_files))
            
            print_success "$test_name completed in ${duration}s"
            print_success "Results: $success_files/$total_files successful ($success_rate%)"
            
            # Clean up test results
            rm -rf "$output_dir"
            
            return 0
        else
            print_error "Output directory not found: $output_dir"
            return 1
        fi
    else
        print_error "$test_name failed"
        return 1
    fi
}

# Test route service endpoints
test_service_endpoints() {
    local port=$1
    local provider=$2
    
    print_info "Testing route service endpoints with $provider provider..."
    
    # Test health endpoint
    if curl -s "http://localhost:$port/health" | grep -q '"status":"healthy"'; then
        print_success "Health endpoint working"
    else
        print_error "Health endpoint failed"
        return 1
    fi
    
    # Test provider endpoint
    if curl -s "http://localhost:$port/api/v1/provider" | grep -q "\"provider\":\"$provider\""; then
        print_success "Provider endpoint working"
    else
        print_error "Provider endpoint failed"
        return 1
    fi
    
    # Test route endpoint
    local route_request='{
        "start": {"latitude": 35.6892, "longitude": 51.3890},
        "end": {"latitude": 36.2605, "longitude": 59.6168},
        "profile": "car"
    }'
    
    local response=$(curl -s -X POST "http://localhost:$port/api/v1/route" \
        -H "Content-Type: application/json" \
        -d "$route_request")
    
    if echo "$response" | grep -q '"code":"Ok"'; then
        print_success "Route endpoint working"
    else
        print_error "Route endpoint failed"
        return 1
    fi
    
    return 0
}

# Main test execution
main() {
    print_header "VEHICLE TRACKING ROUTE GENERATOR - COMPREHENSIVE TESTS"
    
    check_dependencies
    build_project
    
    # Test matrix
    local tests=(
        "online_random:scripts/test_config_online_random.yaml"
        "online_permutation:scripts/test_config_online_permutation.yaml"
        "local_random:scripts/test_config_local_random.yaml"
        "local_permutation:scripts/test_config_local_permutation.yaml"
    )
    
    local passed_tests=0
    local failed_tests=0
    
    for test_spec in "${tests[@]}"; do
        IFS=':' read -r test_name config_file <<< "$test_spec"
        
        print_header "TEST: $test_name"
        
        # Determine provider and port
        if [[ "$test_name" == online* ]]; then
            provider="openstreetmap"
            port=8091
        else
            provider="local-osrm"
            port=8092
        fi
        
        # Start service
        if start_service "$provider" "$port"; then
            # Update config with correct port
            sed -i.bak "s|localhost:8090|localhost:$port|g" "$config_file"
            
            # Test service endpoints
            if test_service_endpoints "$port" "$provider"; then
                # Run generator test
                if run_generator_test "$config_file" "$test_name"; then
                    ((passed_tests++))
                else
                    ((failed_tests++))
                fi
            else
                ((failed_tests++))
            fi
            
            # Stop service
            stop_service
            
            # Restore original config
            mv "${config_file}.bak" "$config_file"
        else
            ((failed_tests++))
        fi
        
        # Small delay between tests
        sleep 2
    done
    
    # Print summary
    print_header "TEST SUMMARY"
    echo -e "Total tests: ${#tests[@]}"
    echo -e "Passed: ${GREEN}$passed_tests${NC}"
    echo -e "Failed: ${RED}$failed_tests${NC}"
    
    if [ $failed_tests -eq 0 ]; then
        print_success "All tests passed! The route generator is working correctly."
        echo -e "\n${GREEN}✅ Ready for production use!${NC}"
    else
        print_error "Some tests failed. Check the logs for details."
        exit 1
    fi
    
    # Clean up
    rm -f service_*.log 2>/dev/null || true
}

# Handle cleanup on exit
cleanup() {
    stop_service
    rm -f service_*.log 2>/dev/null || true
}

trap cleanup EXIT

# Run main function
main "$@"