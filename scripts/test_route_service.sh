#!/bin/bash

# Vehicle Tracking Route Service Test Script
# This script tests the route service API endpoints

set -e

# Configuration
PORT=${PORT:-8090}
BASE_URL="http://localhost:$PORT"
HEALTH_ENDPOINT="$BASE_URL/health"
ROUTE_ENDPOINT="$BASE_URL/api/v1/route"
WAYPOINTS_ENDPOINT="$BASE_URL/api/v1/route/waypoints"
PROVIDER_ENDPOINT="$BASE_URL/api/v1/provider"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test data - London coordinates
START_LAT=51.5074
START_LON=-0.1278
END_LAT=51.5155
END_LON=-0.1419
WAYPOINT_LAT=51.5088
WAYPOINT_LON=-0.0977

# Helper functions
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if curl is available
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_info "jq is not installed, JSON output will not be formatted"
        HAS_JQ=false
    else
        HAS_JQ=true
    fi
}

# Wait for service to be ready
wait_for_service() {
    local max_attempts=30
    local attempt=1
    
    print_info "Waiting for service to start on port $PORT..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$HEALTH_ENDPOINT" > /dev/null; then
            print_success "Service is ready!"
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

# Test health endpoint
test_health() {
    print_info "Testing health endpoint..."
    
    local response=$(curl -s "$HEALTH_ENDPOINT")
    
    if [ $? -ne 0 ]; then
        print_error "Failed to reach health endpoint"
        return 1
    fi
    
    if echo "$response" | grep -q '"status":"healthy"'; then
        print_success "Health check passed"
        
        if [ "$HAS_JQ" = true ]; then
            echo "$response" | jq .
        else
            echo "$response"
        fi
        return 0
    else
        print_error "Health check failed"
        echo "$response"
        return 1
    fi
}

# Test provider endpoint
test_provider() {
    print_info "Testing provider endpoint..."
    
    local response=$(curl -s "$PROVIDER_ENDPOINT")
    
    if [ $? -ne 0 ]; then
        print_error "Failed to reach provider endpoint"
        return 1
    fi
    
    if echo "$response" | grep -q '"provider"'; then
        print_success "Provider endpoint working"
        
        if [ "$HAS_JQ" = true ]; then
            echo "$response" | jq .
        else
            echo "$response"
        fi
        return 0
    else
        print_error "Provider endpoint failed"
        echo "$response"
        return 1
    fi
}

# Test route endpoint
test_route() {
    print_info "Testing route endpoint..."
    
    local request_body=$(cat <<EOF
{
    "start": {
        "latitude": $START_LAT,
        "longitude": $START_LON
    },
    "end": {
        "latitude": $END_LAT,
        "longitude": $END_LON
    },
    "profile": "car"
}
EOF
    )
    
    local response=$(curl -s -X POST "$ROUTE_ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "$request_body")
    
    if [ $? -ne 0 ]; then
        print_error "Failed to reach route endpoint"
        return 1
    fi
    
    if echo "$response" | grep -q '"code":"Ok"'; then
        print_success "Route endpoint working"
        
        # Extract basic route info
        if [ "$HAS_JQ" = true ]; then
            local distance=$(echo "$response" | jq -r '.routes[0].distance // "N/A"')
            local duration=$(echo "$response" | jq -r '.routes[0].duration // "N/A"')
            local summary=$(echo "$response" | jq -r '.routes[0].summary // "N/A"')
            
            echo "Route found:"
            echo "  Distance: $distance meters"
            echo "  Duration: $duration seconds"
            echo "  Summary: $summary"
            
            # Show full response if requested
            if [ "${VERBOSE:-false}" = true ]; then
                echo "$response" | jq .
            fi
        else
            echo "$response"
        fi
        return 0
    elif echo "$response" | grep -q '"error"'; then
        print_error "Route endpoint returned error"
        echo "$response"
        return 1
    else
        print_error "Unexpected response from route endpoint"
        echo "$response"
        return 1
    fi
}

# Test waypoints endpoint
test_waypoints() {
    print_info "Testing waypoints endpoint..."
    
    local request_body=$(cat <<EOF
{
    "waypoints": [
        {
            "latitude": $START_LAT,
            "longitude": $START_LON
        },
        {
            "latitude": $WAYPOINT_LAT,
            "longitude": $WAYPOINT_LON
        },
        {
            "latitude": $END_LAT,
            "longitude": $END_LON
        }
    ],
    "profile": "car"
}
EOF
    )
    
    local response=$(curl -s -X POST "$WAYPOINTS_ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "$request_body")
    
    if [ $? -ne 0 ]; then
        print_error "Failed to reach waypoints endpoint"
        return 1
    fi
    
    if echo "$response" | grep -q '"code":"Ok"'; then
        print_success "Waypoints endpoint working"
        
        if [ "$HAS_JQ" = true ]; then
            local distance=$(echo "$response" | jq -r '.routes[0].distance // "N/A"')
            local duration=$(echo "$response" | jq -r '.routes[0].duration // "N/A"')
            
            echo "Multi-point route found:"
            echo "  Total distance: $distance meters"
            echo "  Total duration: $duration seconds"
            
            if [ "${VERBOSE:-false}" = true ]; then
                echo "$response" | jq .
            fi
        else
            echo "$response"
        fi
        return 0
    elif echo "$response" | grep -q '"error"'; then
        print_error "Waypoints endpoint returned error"
        echo "$response"
        return 1
    else
        print_error "Unexpected response from waypoints endpoint"
        echo "$response"
        return 1
    fi
}

# Test error cases
test_errors() {
    print_info "Testing error cases..."
    
    # Test invalid coordinates
    local invalid_request=$(cat <<EOF
{
    "start": {
        "latitude": 100,  # Invalid latitude (> 90)
        "longitude": -0.1278
    },
    "end": {
        "latitude": 51.5155,
        "longitude": -0.1419
    }
}
EOF
    )
    
    local response=$(curl -s -X POST "$ROUTE_ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "$invalid_request")
    
    if echo "$response" | grep -q '"error"'; then
        print_success "Error handling working (invalid coordinates)"
    else
        print_error "Error handling failed for invalid coordinates"
    fi
    
    # Test missing required fields
    local missing_fields=$(cat <<EOF
{
    "start": {
        "latitude": 51.5074,
        "longitude": -0.1278
    }
    # Missing "end" field
}
EOF
    )
    
    response=$(curl -s -X POST "$ROUTE_ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "$missing_fields")
    
    if echo "$response" | grep -q '"error"'; then
        print_success "Error handling working (missing fields)"
    else
        print_error "Error handling failed for missing fields"
    fi
}

# Test different profiles
test_profiles() {
    print_info "Testing different routing profiles..."
    
    local profiles=("car" "bike" "foot")
    
    for profile in "${profiles[@]}"; do
        print_info "Testing $profile profile..."
        
        local request_body=$(cat <<EOF
{
    "start": {
        "latitude": $START_LAT,
        "longitude": $START_LON
    },
    "end": {
        "latitude": $END_LAT,
        "longitude": $END_LON
    },
    "profile": "$profile"
}
EOF
        )
        
        local response=$(curl -s -X POST "$ROUTE_ENDPOINT" \
            -H "Content-Type: application/json" \
            -d "$request_body")
        
        if echo "$response" | grep -q '"code":"Ok"'; then
            print_success "$profile profile working"
        else
            print_error "$profile profile failed"
        fi
    done
}

# Main test execution
main() {
    print_info "Starting Vehicle Tracking Route Service Tests"
    echo "=============================================="
    
    check_dependencies
    
    # Check if service is already running
    if ! curl -s "$HEALTH_ENDPOINT" > /dev/null; then
        print_info "Service not running. Please start it first:"
        echo "  ./bin/route-service -port $PORT"
        echo ""
        echo "Or use the service runner:"
        echo "  ./scripts/run_service.sh start $PORT"
        echo ""
        read -p "Press Enter when service is running, or Ctrl+C to cancel..."
    fi
    
    wait_for_service
    
    # Run tests
    local tests_passed=0
    local tests_failed=0
    
    if test_health; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    if test_provider; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    if test_route; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    if test_waypoints; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    test_errors
    test_profiles
    
    echo ""
    echo "=============================================="
    print_info "Test Summary:"
    echo "  Tests passed: $tests_passed"
    echo "  Tests failed: $tests_failed"
    
    if [ $tests_failed -eq 0 ]; then
        print_success "All tests passed! Route service is working correctly."
    else
        print_error "Some tests failed. Check the output above for details."
        exit 1
    fi
}

# Handle command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -p, --port PORT    Port number (default: 8080)"
            echo "  -v, --verbose      Show verbose output"
            echo "  -h, --help         Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h for help"
            exit 1
            ;;
    esac
done

# Run main function
main
