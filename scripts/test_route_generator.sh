#!/bin/bash

# Test script for route generator
# This script tests the route generator with a small number of routes

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=== Route Generator Test ==="
echo "Project root: $PROJECT_ROOT"
echo ""

# Check if route service is running
echo "Checking if route service is running..."
if ! curl -s http://localhost:8090/health > /dev/null; then
    echo "Route service is not running on port 8090"
    echo "Please start it first: make run"
    exit 1
fi

echo "Route service is running"
echo ""

# Create test configuration
echo "Creating test configuration..."
cd "$PROJECT_ROOT"
cat > test_config.yaml << 'EOF'
# Test Route Generator Configuration
route_generator:
  # Small number for testing
  route_count: 10
  
  # Test both methods
  method: "random"
  
  # Country for random generation
  country: "Iran"
  
  # Country bounding boxes (latitude, longitude)
  country_bounds:
    Iran:
      min_lat: 25.0
      max_lat: 40.0
      min_lng: 44.0
      max_lng: 63.0
  
  # Location set for permutation method
  location_set:
    - name: "Tehran"
      lat: 35.6892
      lng: 51.3890
    - name: "Mashhad"
      lat: 36.2605
      lng: 59.6168
    - name: "Isfahan"
      lat: 32.6546
      lng: 51.6680
  
  # Route service configuration
  route_service:
    base_url: "http://localhost:8090"
    timeout_seconds: 10
    max_concurrent_requests: 5
  
  # Output configuration
  output:
    directory: "./test_generated_routes"
    format: "json"
    compress: false
    
  # Random seed for reproducibility
  random_seed: 42
EOF

echo "Test configuration created: test_config.yaml"
echo ""

# Build the generator
echo "Building route generator..."
make build-generator
echo ""

# Test random method
echo "=== Testing Random Method ==="
./bin/route-generator -config test_config.yaml
echo ""

# Check output
echo "Checking output..."
if [ -d "./test_generated_routes" ]; then
    echo "Output directory created successfully"
    echo ""
    
    echo "Generated files:"
    ls -la ./test_generated_routes/
    echo ""
    
    echo "Metadata content:"
    if command -v jq >/dev/null 2>&1; then
        cat ./test_generated_routes/metadata.json | jq '. | length'
        echo " routes generated"
        echo ""
        
        echo "Summary:"
        cat ./test_generated_routes/summary.json | jq '.'
        echo ""
    else
        echo "jq not installed, showing raw JSON"
        echo "Number of routes: $(grep -c '"id"' ./test_generated_routes/metadata.json || echo "unknown")"
        echo ""
        cat ./test_generated_routes/summary.json
        echo ""
    fi
else
    echo "ERROR: Output directory not created"
    exit 1
fi

# Test permutation method
echo "=== Testing Permutation Method ==="
# Update config to use permutation
sed -i 's/method: "random"/method: "permutation"/' test_config.yaml
sed -i 's/directory: ".\/test_generated_routes"/directory: ".\/test_generated_routes_perm"/' test_config.yaml

./bin/route-generator -config test_config.yaml
echo ""

# Check permutation output
echo "Checking permutation output..."
if [ -d "./test_generated_routes_perm" ]; then
    echo "Permutation output directory created successfully"
    echo ""
    
    echo "Permutation summary:"
    if command -v jq >/dev/null 2>&1; then
        cat ./test_generated_routes_perm/summary.json | jq '.'
        echo ""
    else
        cat ./test_generated_routes_perm/summary.json
        echo ""
    fi
fi

# Cleanup
echo "=== Cleaning up ==="
rm -f test_config.yaml
rm -rf test_generated_routes
rm -rf test_generated_routes_perm

echo ""
echo "=== Test completed successfully ==="
echo "The route generator is working correctly with both methods."
echo "To generate 17000 routes, update config.yaml with:"
echo "  route_count: 17000"
echo "  max_concurrent_requests: 50 (or higher)"
echo "  timeout_seconds: 30 (or higher)"
echo ""
echo "Then run: make run-generator"