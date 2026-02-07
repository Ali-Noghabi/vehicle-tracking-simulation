# Vehicle Tracking Simulation - Route Service

A microservice for finding routes between geographic coordinates, designed to be used in vehicle tracking and other location-based applications.

## Features

- **Multiple Routing Providers**: Support for OpenStreetMap (OSRM), with extensible architecture for Google Maps, Mapbox, etc.
- **Standardized Response Format**: Returns routes in OSRM-compatible format
- **Microservice Architecture**: Clean separation of concerns, easy to integrate with other services
- **RESTful API**: Simple HTTP endpoints for route finding
- **Configurable**: Command-line options for provider selection, timeouts, and custom URLs

## Project Structure

```
vehicle-tracking-simulation/
├── cmd/
│   └── route-service/
│       └── main.go           # Application entry point
├── internal/
│   └── route-service/
│       ├── api/
│       │   └── handler.go    # HTTP handlers
│       ├── models/
│       │   ├── route.go      # Route data models
│       │   └── coordinate.go # Coordinate utilities
│       ├── provider/
│       │   ├── provider.go   # Provider interface
│       │   ├── factory.go    # Provider factory
│       │   └── openstreetmap.go # OpenStreetMap implementation
│       └── service/
│           └── route_finder.go  # Business logic
├── scripts/
│   ├── run_service.sh        # Service management script
│   ├── test_route_service.sh # Bash test script
│   ├── test_route_details.sh # Detailed route test script
│   └── SERVICE_RUNNER.md     # Service runner documentation
├── tests/
│   ├── test_route_service.go # Go test client
│   └── test_complete_route.go # Complete route structure test
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## Installation

```bash
# Clone the repository
git clone <repo-url>
cd vehicle-tracking-simulation

# Download dependencies
go mod download

# Build the service
go build -o bin/route-service ./cmd/route-service
```

## Running the Service

```bash
# Start with default settings (OpenStreetMap)
./bin/route-service

# Start with custom port
./bin/route-service -port 8081

# Start with specific provider
./bin/route-service -provider openstreetmap -timeout 15

# Start with custom OSRM instance
./bin/route-service -provider openstreetmap -base-url https://your-osrm-instance.com
```

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-provider` | openstreetmap | Routing provider (openstreetmap, google, mapbox, here) |
| `-api-key` | "" | API key for paid providers |
| `-base-url` | "" | Custom base URL for the routing provider |
| `-port` | 8080 | Port to listen on |
| `-timeout` | 10 | Request timeout in seconds |

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

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
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

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
