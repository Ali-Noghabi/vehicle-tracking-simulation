#!/bin/bash

# Vehicle Tracking Route Service - Detailed Route Test
# Shows complete route structure with steps and instructions

set -e

# Configuration
PORT=${PORT:-8090}
BASE_URL="http://localhost:$PORT"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test data - London coordinates
START_LAT=51.5074
START_LON=-0.1278
END_LAT=51.5155
END_LON=-0.1419

# Helper functions
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check dependencies
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        echo "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_warning "jq is not installed, JSON output will not be formatted"
        HAS_JQ=false
    else
        HAS_JQ=true
    fi
}

# Test 1: Get complete route structure
test_complete_route() {
    print_info "Testing complete route structure..."
    
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
    
    local response=$(curl -s -X POST "$BASE_URL/api/v1/route" \
        -H "Content-Type: application/json" \
        -d "$request_body")
    
    if [ $? -ne 0 ]; then
        echo "Failed to reach route endpoint"
        return 1
    fi
    
    echo ""
    echo "=============================================="
    echo "COMPLETE ROUTE STRUCTURE"
    echo "=============================================="
    echo ""
    
    if [ "$HAS_JQ" = true ]; then
        # Show the complete JSON response
        echo "$response" | jq .
        
        # Extract and display key information
        echo ""
        echo "=============================================="
        echo "ROUTE SUMMARY"
        echo "=============================================="
        
        local code=$(echo "$response" | jq -r '.code')
        local distance=$(echo "$response" | jq -r '.routes[0].distance // "N/A"')
        local duration=$(echo "$response" | jq -r '.routes[0].duration // "N/A"')
        local summary=$(echo "$response" | jq -r '.routes[0].summary // "N/A"')
        local legs=$(echo "$response" | jq -r '.routes[0].legs | length // "0"')
        
        echo "Response Code: $code"
        echo "Total Distance: $distance meters"
        echo "Total Duration: $duration seconds"
        echo "Route Summary: $summary"
        echo "Number of Legs: $legs"
        
        # Show each leg
        if [ "$legs" -gt 0 ]; then
            echo ""
            echo "LEG DETAILS:"
            for ((i=0; i<legs; i++)); do
                local leg_distance=$(echo "$response" | jq -r ".routes[0].legs[$i].distance // \"N/A\"")
                local leg_duration=$(echo "$response" | jq -r ".routes[0].legs[$i].duration // \"N/A\"")
                local leg_summary=$(echo "$response" | jq -r ".routes[0].legs[$i].summary // \"N/A\"")
                local steps=$(echo "$response" | jq -r ".routes[0].legs[$i].steps | length // \"0\"")
                
                echo "  Leg $((i+1)):"
                echo "    Distance: $leg_distance meters"
                echo "    Duration: $leg_duration seconds"
                echo "    Summary: $leg_summary"
                echo "    Steps: $steps"
                
                # Show first few steps
                if [ "$steps" -gt 0 ] && [ "$steps" -le 5 ]; then
                    echo "    Step Instructions:"
                    for ((j=0; j<steps; j++)); do
                        local instruction=$(echo "$response" | jq -r ".routes[0].legs[$i].steps[$j].instruction // \"N/A\"")
                        local step_distance=$(echo "$response" | jq -r ".routes[0].legs[$i].steps[$j].distance // \"N/A\"")
                        echo "      $((j+1)). $instruction ($step_distance meters)"
                    done
                fi
            done
        fi
        
        # Show waypoints if available
        local waypoints=$(echo "$response" | jq -r '.waypoints | length // "0"')
        if [ "$waypoints" -gt 0 ]; then
            echo ""
            echo "WAYPOINTS:"
            for ((i=0; i<waypoints; i++)); do
                local name=$(echo "$response" | jq -r ".waypoints[$i].name // \"Unnamed\"")
                local location=$(echo "$response" | jq -r ".waypoints[$i].location // []")
                echo "  Waypoint $((i+1)): $name at $location"
            done
        fi
        
    else
        # Without jq, just show the raw response
        echo "$response"
    fi
    
    return 0
}

# Test 2: Get route with verbose output
test_verbose_route() {
    print_info "Testing route with verbose step-by-step instructions..."
    
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
    
    local response=$(curl -s -X POST "$BASE_URL/api/v1/route" \
        -H "Content-Type: application/json" \
        -d "$request_body")
    
    if [ $? -ne 0 ]; then
        echo "Failed to reach route endpoint"
        return 1
    fi
    
    echo ""
    echo "=============================================="
    echo "STEP-BY-STEP INSTRUCTIONS"
    echo "=============================================="
    echo ""
    
    if [ "$HAS_JQ" = true ]; then
        # Extract and display step-by-step instructions
        local legs=$(echo "$response" | jq -r '.routes[0].legs | length // "0"')
        
        for ((leg_index=0; leg_index<legs; leg_index++)); do
            echo "LEG $((leg_index+1)):"
            
            local steps=$(echo "$response" | jq -r ".routes[0].legs[$leg_index].steps | length // \"0\"")
            
            for ((step_index=0; step_index<steps; step_index++)); do
                local instruction=$(echo "$response" | jq -r ".routes[0].legs[$leg_index].steps[$step_index].instruction // \"Continue\"")
                local distance=$(echo "$response" | jq -r ".routes[0].legs[$leg_index].steps[$step_index].distance // \"0\"")
                local duration=$(echo "$response" | jq -r ".routes[0].legs[$leg_index].steps[$step_index].duration // \"0\"")
                local name=$(echo "$response" | jq -r ".routes[0].legs[$leg_index].steps[$step_index].name // \"Unnamed road\"")
                local maneuver_type=$(echo "$response" | jq -r ".routes[0].legs[$leg_index].steps[$step_index].maneuver.type // \"\"")
                local maneuver_modifier=$(echo "$response" | jq -r ".routes[0].legs[$leg_index].steps[$step_index].maneuver.modifier // \"\"")
                
                echo "  Step $((step_index+1)):"
                echo "    Instruction: $instruction"
                echo "    Distance: ${distance}m, Duration: ${duration}s"
                echo "    Road: $name"
                
                if [ -n "$maneuver_type" ]; then
                    echo "    Maneuver: $maneuver_type"
                    if [ -n "$maneuver_modifier" ]; then
                        echo "    Direction: $maneuver_modifier"
                    fi
                fi
                echo ""
            done
        done
    else
        echo "Install jq for formatted step-by-step instructions"
        echo "$response"
    fi
    
    return 0
}

# Test 3: Test multi-point route
test_multi_point_route() {
    print_info "Testing multi-point route..."
    
    local waypoint_lat=51.5088
    local waypoint_lon=-0.0977
    
    local request_body=$(cat <<EOF
{
    "waypoints": [
        {
            "latitude": $START_LAT,
            "longitude": $START_LON
        },
        {
            "latitude": $waypoint_lat,
            "longitude": $waypoint_lon
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
    
    local response=$(curl -s -X POST "$BASE_URL/api/v1/route/waypoints" \
        -H "Content-Type: application/json" \
        -d "$request_body")
    
    if [ $? -ne 0 ]; then
        echo "Failed to reach waypoints endpoint"
        return 1
    fi
    
    echo ""
    echo "=============================================="
    echo "MULTI-POINT ROUTE"
    echo "=============================================="
    echo ""
    
    if [ "$HAS_JQ" = true ]; then
        echo "$response" | jq .
        
        # Extract summary
        local code=$(echo "$response" | jq -r '.code')
        local distance=$(echo "$response" | jq -r '.routes[0].distance // "N/A"')
        local duration=$(echo "$response" | jq -r '.routes[0].duration // "N/A"')
        local legs=$(echo "$response" | jq -r '.routes[0].legs | length // "0"')
        
        echo ""
        echo "SUMMARY:"
        echo "  Response Code: $code"
        echo "  Total Distance: $distance meters"
        echo "  Total Duration: $duration seconds"
        echo "  Number of Legs: $legs"
        
        # Show each leg summary
        for ((i=0; i<legs; i++)); do
            local leg_distance=$(echo "$response" | jq -r ".routes[0].legs[$i].distance // \"N/A\"")
            local leg_duration=$(echo "$response" | jq -r ".routes[0].legs[$i].duration // \"N/A\"")
            echo "  Leg $((i+1)): ${leg_distance}m in ${leg_duration}s"
        done
    else
        echo "$response"
    fi
    
    return 0
}

# Main execution
main() {
    print_info "Starting Detailed Route Tests"
    echo "=============================================="
    
    check_dependencies
    
    # Check if service is running
    if ! curl -s "$BASE_URL/health" > /dev/null; then
        echo "Service is not running on port $PORT"
        echo "Start it with: ./run_service.sh start"
        exit 1
    fi
    
    # Run tests
    test_complete_route
    echo ""
    
    test_verbose_route
    echo ""
    
    test_multi_point_route
    
    echo ""
    print_success "All detailed route tests completed!"
}

# Handle command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -p, --port PORT    Port number (default: 8090)"
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