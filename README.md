# Vehicle Tracking Simulation - Route Service & Generator

A complete vehicle tracking simulation system with route microservice and high-performance route generator for creating large datasets of realistic routes.

## Features

### Route Service
- **Multiple Routing Providers**: Support for OpenStreetMap (OSRM), Local OSRM instance, with extensible architecture for Google Maps, Mapbox, etc.
- **Standardized Response Format**: Returns routes in OSRM-compatible format
- **Microservice Architecture**: Clean separation of concerns, easy to integrate with other services
- **RESTful API**: Simple HTTP endpoints for route finding
- **Configurable**: Command-line options for provider selection, timeouts, and custom URLs
- **Health Monitoring**: Built-in health checks and provider information

### Route Generator
- **Two Generation Methods**: Random coordinates within country bounds or permutations from location sets
- **Parallel Processing**: Configurable concurrent requests for high-throughput generation
- **Scalable**: Designed to generate 17,000+ routes efficiently
- **Storage Format**: Saves routes in simulation-ready format for future vehicle tracking
- **Reproducible**: Configurable random seeds for consistent results
- **Retry Logic**: Automatic retry with exponential backoff for failed requests
- **Performance Optimized**: 50+ concurrent requests with local OSRM, 3+ with public OSRM

## Project Structure

```
vehicle-tracking-simulation/
├── cmd/
│   ├── route-service/
│   │   └── main.go           # Route service entry point
│   └── route-generator/
│       └── main.go           # Route generator entry point
├── internal/
│   ├── route-service/
│   │   ├── api/
│   │   │   └── handler.go    # HTTP handlers
│   │   ├── models/
│   │   │   ├── route.go      # Route data models
│   │   │   └── coordinate.go # Coordinate utilities
│   │   ├── provider/
│   │   │   ├── provider.go   # Provider interface
│   │   │   ├── factory.go    # Provider factory
│   │   │   ├── openstreetmap.go # OpenStreetMap implementation
│   │   │   └── localosrm.go  # Local OSRM implementation
│   │   └── service/
│   │       └── route_finder.go  # Business logic
│   └── route-generator/
│       ├── config/
│       │   └── config.go     # Configuration management
│       ├── generator/
│       │   └── generator.go  # Route generation logic
│       ├── processor/
│       │   └── processor.go  # Route service communication
│       └── storage/
│           └── storage.go    # Route data storage
├── scripts/
│   ├── run_service.sh                # Service management script
│   ├── run_comprehensive_tests.sh    # Complete test suite runner
│   ├── test_route_service.sh         # Route service API test script
│   ├── test_route_details.sh         # Detailed route test script
│   ├── test_route_generator.sh       # Route generator test script
│   ├── test_config.yaml              # Main test configuration
│   ├── test_config_online_random.yaml    # Online OSRM random test
│   ├── test_config_online_permutation.yaml # Online OSRM permutation test
│   ├── test_config_local_random.yaml     # Local OSRM random test
│   ├── test_config_local_permutation.yaml # Local OSRM permutation test
│   └── SERVICE_RUNNER.md             # Service runner documentation
├── tests/
│   ├── test_route_service.go # Go test client
│   └── test_complete_route.go # Complete route structure test
├── config.yaml               # Route generator configuration
├── Makefile                  # Build automation
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Git
- curl (for testing)
- jq (optional, for JSON formatting in tests)

### Installation

```bash
# Clone the repository
git clone <repo-url>
cd vehicle-tracking-simulation

# Download dependencies
go mod download

# Build everything
make build

# Or build individually
make build-service    # Build route service
make build-generator  # Build route generator
```

## Running the Service

### Route Service

```bash
# Build and run the route service with default settings
make run

# Or run with custom port
make run-port PORT=8080

# Run with different providers
make run-online-osrm      # Public OSRM (rate limited)
make run-local-osrm       # Local OSRM instance (no rate limits)

# Build only
make build

# Clean build artifacts
make clean
```

### Using the Service Runner Script

```bash
# Start service
./scripts/run_service.sh start 8090

# Stop service
./scripts/run_service.sh stop

# Check status
./scripts/run_service.sh status

# View logs
./scripts/run_service.sh logs
```

### Route Generator

The route generator creates large datasets of routes for vehicle tracking simulations.

### Testing

```bash
# Quick test - verify everything works
make test-service

# Generate test routes with different scenarios
make run-test-local-random     # Local OSRM + Random coordinates
make run-test-local-permutation # Local OSRM + Location permutations

# For online testing (requires internet)
make run-test-random           # Online OSRM + Random
make run-test-permutation      # Online OSRM + Permutation
```

#### Configuration

Create a configuration file (e.g., `config.yaml`):

```yaml
route_generator:
  route_count: 1000
  method: "random"  # or "permutation"
  country: "Iran"
  
  country_bounds:
    Iran:
      min_lat: 25.0
      max_lat: 40.0
      min_lng: 44.0
      max_lng: 63.0
  
  location_set:  # For permutation method
    - name: "Tehran"
      lat: 35.6892
      lng: 51.3890
    - name: "Mashhad"
      lat: 36.2605
      lng: 59.6168
  
  route_service:
    base_url: "http://localhost:8090"
    timeout_seconds: 30
    max_concurrent_requests: 20  # 50 for local OSRM, 3 for public OSRM
  
  output:
    directory: "./generated_routes"
    format: "json"
    compress: false
    
  random_seed: 42
```

#### Running the Generator

```bash
# Build and run the generator
make run-generator

# Or run manually
./bin/route-generator -config config.yaml

# Test with pre-configured test scenarios
./scripts/run_comprehensive_tests.sh
```

#### Output Format

Routes are saved in the configured output directory with:
- Individual route files (`route_000001.json`, etc.)
- Metadata index (`metadata.json`)
- Generation summary (`summary.json`)

Each route file contains complete route data for simulation, including:
- Start/end coordinates and metadata
- Complete route geometry (encoded polyline)
- Step-by-step navigation instructions
- Distance and duration information
- Speed annotations for realistic tracking
- Waypoint information and street names

#### Performance Characteristics

| Provider | Max Concurrent | Success Rate | 1000 Routes | 17,000 Routes |
|----------|---------------|--------------|-------------|---------------|
| Public OSRM | 3 requests | ~90% | ~1.5 hours | ~94 hours (4 days) |
| Local OSRM | 50 requests | ~85% | ~2 minutes | ~35 minutes |

**Recommendation**: Use local OSRM for large-scale route generation (>1000 routes).

### Route Service Command Line Options

```bash
./bin/route-service -help
```

| Option | Default | Description |
|--------|---------|-------------|
| `-provider` | openstreetmap | Routing provider (`openstreetmap`, `local-osrm`) |
| `-api-key` | "" | API key for paid providers (future) |
| `-base-url` | "" | Custom base URL for the routing provider |
| `-port` | 8080 | Port to listen on |
| `-timeout` | 10 | Request timeout in seconds |
| `-help` | false | Show help message |

### Route Generator Command Line Options

```bash
./bin/route-generator -help
```

| Option | Default | Description |
|--------|---------|-------------|
| `-config` | config.yaml | Configuration file path |
| `-help` | false | Show help message |

### Using Local OSRM Instance

For generating large numbers of routes (17,000+), use a local OSRM instance:

```bash
# Start route service with local OSRM provider
./bin/route-service -provider local-osrm -port 8090 -timeout 30

# Update config.yaml to use higher concurrency for local instance
# route_service:
#   max_concurrent_requests: 50  # Much higher for local OSRM
```

**Benefits of Local OSRM:**
- No rate limiting
- Faster response times
- Can handle 50+ concurrent requests
- Complete control over routing data

**Setting up Local OSRM:**

1. **Install OSRM**:
   ```bash
   # Using Docker (recommended)
   docker run -t -v $(pwd):/data osrm/osrm-backend osrm-extract -p /opt/car.lua /data/iran-latest.osm.pbf
   docker run -t -v $(pwd):/data osrm/osrm-backend osrm-partition /data/iran-latest.osrm
   docker run -t -v $(pwd):/data osrm/osrm-backend osrm-customize /data/iran-latest.osrm
   
   # Run OSRM backend
   docker run -t -i -p 5000:5000 -v $(pwd):/data osrm/osrm-backend osrm-routed --algorithm mld /data/iran-latest.osrm
   ```

2. **Download OSM data** for your region:
   ```bash
   # Iran data from Geofabrik
   wget https://download.geofabrik.de/asia/iran-latest.osm.pbf
   ```

3. **Verify OSRM is running**:
   ```bash
   curl "http://localhost:5000/route/v1/driving/51.3890,35.6892;59.6168,36.2605?steps=true"
   ```

4. **Use with route service**:
   ```bash
   ./bin/route-service -provider local-osrm -port 8090 -timeout 30
   ```

## Testing

### Quick Test Commands

```bash
# Test route service API endpoints
make test-service

# Test route generator
make test-generator

# Run comprehensive tests (all scenarios)
make test-comprehensive

# Test specific scenarios:
make run-test-random          # Online OSRM + Random method
make run-test-permutation     # Online OSRM + Permutation method  
make run-test-local-random    # Local OSRM + Random method
make run-test-local-permutation # Local OSRM + Permutation method
```

### Manual Testing

```bash
# 1. Start the service with local OSRM
make run-local-osrm

# 2. In another terminal, test the generator
make run-test-local-random

# 3. Check service health
curl http://localhost:8090/health

# 4. Test a single route
curl -X POST http://localhost:8090/api/v1/route \
  -H "Content-Type: application/json" \
  -d '{
    "start": {"latitude": 35.6892, "longitude": 51.3890},
    "end": {"latitude": 36.2605, "longitude": 59.6168},
    "profile": "car"
  }'
```

### Test Scripts

The `scripts/` directory contains several test scripts:

```bash
# Test route service endpoints
./scripts/test_route_service.sh

# Test route generator
./scripts/test_route_generator.sh

# Run comprehensive tests
./scripts/run_comprehensive_tests.sh

# Test with custom port
PORT=8080 ./scripts/test_route_service.sh
```

## API Endpoints

### 1. Health Check

Check if the service is running.

```http
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "route-service",
  "provider": "openstreetmap"
}
```

### 2. Find Route

Find a route between two coordinates.

```http
POST /api/v1/route
Content-Type: application/json
```

Request Body:
```json
{
  "start": {
    "latitude": 51.5074,
    "longitude": -0.1278
  },
  "end": {
    "latitude": 51.5155,
    "longitude": -0.1419
  },
  "profile": "car"
}
```

Response:
```json
{
  "code": "Ok",
  "routes": [
    {
      "geometry": "w|v~Fq~w~L~M~R...",
      "legs": [...],
      "distance": 2500.5,
      "duration": 400.0,
      "weight_name": "routability",
      "weight": 250.05,
      "summary": "Strand, Aldwych"
    }
  ],
  "waypoints": [...]
}
```

### 3. Find Route with Waypoints

Find a route through multiple intermediate points.

```http
POST /api/v1/route/waypoints
Content-Type: application/json
```

Request Body:
```json
{
  "waypoints": [
    {"latitude": 51.5074, "longitude": -0.1278},
    {"latitude": 51.5155, "longitude": -0.1419},
    {"latitude": 51.5088, "longitude": -0.0977}
  ],
  "profile": "car"
}
```

### 4. Get Provider Info

Get information about the current routing provider.

```http
GET /api/v1/provider
```

Response:
```json
{
  "provider": "openstreetmap"
}
```

## Route Response Format

The service returns routes in **OSRM-compatible format**, which is a standard for routing services:

- **Coordinates**: In `[longitude, latitude]` order (GeoJSON standard)
- **Polyline Encoding**: Geometry uses encoded polyline format for efficient compression
- **Turn-by-turn Instructions**: Includes detailed steps with maneuvers
- **Route Statistics**: Distance (meters), duration (seconds), average speed

## Architecture

### Provider Pattern

The service uses a provider interface to support multiple routing backends:

```go
type Provider interface {
    FindRoute(start Coordinate, end Coordinate, profile string) (*RouteResponse, error)
    FindRouteWithWaypoints(waypoints []Coordinate, profile string) (*RouteResponse, error)
    ProviderName() string
}
```

### Separation of Concerns

- **API layer** (`api/`): HTTP handling and routing
- **Service layer** (`service/`): Business logic and validation
- **Provider layer** (`provider/`): External API integration
- **Models layer** (`models/`): Data structures and domain models

## Why This Structure?

- **Microservice-ready**: Clean boundaries make it easy to extract or combine services
- **Testable**: Interfaces enable easy mocking for unit tests
- **Extensible**: Adding new providers doesn't require modifying existing code
- **Standard format**: OSRM format ensures compatibility with mapping libraries

## Testing

### Comprehensive Test Suite

```bash
# Run the complete test suite (all providers × all methods)
./scripts/run_comprehensive_tests.sh

# Individual test scripts
./scripts/test_route_service.sh          # Test route service API
./scripts/test_route_details.sh          # Test detailed route information
./scripts/test_route_generator.sh        # Test route generator

# Go tests
go test ./...                            # Run all Go tests
go test -cover ./...                     # Run with coverage
go test -v ./...                         # Run with verbose output
```

### Test Scenarios

The comprehensive test suite validates:

1. **Route Service API**: Health checks, provider info, route finding
2. **Online OSRM with Random Method**: 50 random routes with public OSRM
3. **Online OSRM with Permutation Method**: 50 routes between predefined locations
4. **Local OSRM with Random Method**: 50 random routes with local OSRM
5. **Local OSRM with Permutation Method**: 50 routes between predefined locations

Each test measures:
- Success rate (routes found / routes attempted)
- Processing time
- Concurrency performance
- Error handling and retry logic

## Deployment

### Production Recommendations

1. **For Route Service**:
   - Use `local-osrm` provider for high-volume applications
   - Set appropriate timeouts (30+ seconds for complex routes)
   - Monitor health endpoints
   - Consider load balancing for multiple instances

2. **For Route Generation**:
   - Use local OSRM for batches > 1000 routes
   - Adjust `max_concurrent_requests` based on OSRM capacity
   - Monitor success rates and adjust bounding boxes
   - Use reproducible random seeds for testing

3. **Performance Tuning**:
   ```yaml
   # config.yaml for production
   route_service:
     timeout_seconds: 60
     max_concurrent_requests: 50  # Local OSRM can handle this
   
   # For public OSRM
   route_service:
     timeout_seconds: 30
     max_concurrent_requests: 3   # Respect rate limits
   ```

### Monitoring

- Health endpoint: `GET /health`
- Provider info: `GET /api/v1/provider`
- Log success/failure rates from route generator
- Monitor processing times for large batches

## Vehicle Tracking Simulation Service

Now that you have generated routes, you can simulate vehicle movement and send telemetry data via MQTT.

### Features

- **Realistic Vehicle Simulation**: Vehicles move along generated routes with realistic speeds
- **MQTT Telemetry**: Sends real-time telemetry data in batch format
- **Polyline Decoding**: Accurately follows route geometry from generated routes
- **Configurable Parameters**: Speed variation, update intervals, telemetry ranges
- **Batch Processing**: Efficient batch telemetry sending for multiple vehicles

### Project Structure Update

```
vehicle-tracking-simulation/
├── cmd/
│   ├── route-service/
│   │   └── main.go           # Route service entry point
│   ├── route-generator/
│   │   └── main.go           # Route generator entry point
│   └── simulation-service/   # NEW: Vehicle tracking simulation
│       ├── main.go           # Simulation service entry point
│       ├── polyline.go       # Polyline decoding utilities
│       ├── route_iterator.go # Route position calculation
│       ├── batch_telemetry.go # MQTT batch telemetry
│       ├── config.yaml       # Simulation configuration
│       └── go.mod           # Go module
├── test_results/
│   └── local_random/         # Generated routes for simulation
│       ├── route_000001.json
│       ├── route_000002.json
│       └── ...
└── ...
```

### Quick Start

```bash
# 1. Install MQTT broker (if not already installed)
sudo apt install mosquitto mosquitto-clients
sudo systemctl start mosquitto

# 2. Build simulation service
make build-simulation

# 3. Run simulation (uses routes from test_results/local_random/)
make run-simulation

# 4. Monitor telemetry (in another terminal)
mosquitto_sub -t 'vehicle/telemetry' -v
mosquitto_sub -t 'vehicle/telemetry_batch' -v
```

### Configuration

Edit `cmd/simulation-service/config.yaml`:

```yaml
mqtt:
  broker: "tcp://localhost:1883"
  topic: "vehicle/telemetry"
  client_id: "vehicle_simulator"
  qos: 0
  retain: false

simulation:
  update_interval: "5s"      # Time between telemetry updates
  simulation_speed: 1.0       # 1.0 = real-time, 2.0 = 2x speed
  routes_path: "../../test_results/local_random"
  speed_variation: 0.2        # ±20% variation from average speed
  
  # Telemetry parameters
  altitude_range: [100, 150]  # meters
  accuracy_range: [5, 15]     # meters
  battery_range: [80, 100]    # percentage
  signal_range: [70, 100]     # percentage
```

### Telemetry Format

Individual telemetry (sent to `vehicle/telemetry`):
```json
{
  "vehicle_id": "vehicle_001",
  "timestamp": 1739116800,
  "lat": 35.6892,
  "lon": 51.3890,
  "spd": 45.2,
  "hdg": 123.5,
  "alt": 120.5,
  "acc": 8.2,
  "battery": 92.5,
  "signal": 85.0
}
```

Batch telemetry (sent to `vehicle/telemetry_batch`):
```json
{
  "batch_id": "batch_1739116800123456789",
  "timestamp": 1739116800,
  "vehicles": [
    { ...telemetry 1... },
    { ...telemetry 2... },
    { ...telemetry 3... }
  ],
  "batch_size": 3
}
```

### How It Works

1. **Loads generated routes** from `test_results/local_random/`
2. **Creates vehicle simulators** for each successful route
3. **Calculates speed range** based on route distance/duration
4. **Updates positions** along route geometry using polyline decoding
5. **Sends telemetry** via MQTT at configured intervals
6. **Supports batch telemetry** for efficient bulk sending

### Speed Calculation

The simulation uses realistic speed calculations:
- **Average speed** = route distance / route duration
- **Speed variation** = ±20% (configurable) from average
- **Random speed** within range for each update
- **Position calculation** based on elapsed time × current speed

### Testing the Simulation

```bash
# Build and run simulation
make run-simulation

# Monitor output
# You should see logs like:
# Loaded 44 routes from test_results/local_random
# Created simulator for vehicle_001 (distance: 518500m, duration: 26424s, avg speed: 19.6 m/s)
# Sent 44 telemetry updates at 14:30:05

# Test MQTT connectivity
mosquitto_pub -t 'test' -m 'hello'
mosquitto_sub -t 'test' -v
```

### Integration with Route Generation

The simulation service works seamlessly with the route generator:

```bash
# 1. Generate routes
make run-test-local-random

# 2. Start simulation
make run-simulation

# 3. Monitor telemetry
mosquitto_sub -t 'vehicle/telemetry' -v | jq .
```

## Future Enhancements

- [ ] Add Google Maps provider
- [ ] Add Mapbox provider
- [ ] Add HERE Technologies provider
- [ ] Implement caching for frequent routes
- [ ] Add support for real-time traffic data
- [ ] Implement route optimization for multiple stops
- [ ] Add support for avoiding tolls and highways
- [ ] Add Prometheus metrics
- [ ] Implement rate limiting
- [ ] Add database storage for generated routes
- [ ] Implement batch processing with resume capability
- [ ] **Vehicle Tracking Simulation** ✓ COMPLETED
- [ ] Add web dashboard for monitoring simulation
- [ ] Implement historical playback of routes
- [ ] Add geofencing and alerting
- [ ] Support for multiple vehicle types (trucks, buses, etc.)
- [ ] Integration with mapping visualization tools

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
