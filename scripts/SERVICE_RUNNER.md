# Quick Start Guide for Vehicle Tracking Route Service

## Using the Service Runner Script

The `run_service.sh` script provides easy management of your route service:

### Basic Commands:

```bash
# Show help
./run_service.sh help

# Build the service
./run_service.sh build

# Start the service (default port 8090)
./run_service.sh start

# Start on specific port
./run_service.sh start 8080

# Start with custom configuration
./run_service.sh start 8081 30 openstreetmap https://your-osrm-instance.com

# Check status
./run_service.sh status

# Show logs
./run_service.sh logs
./run_service.sh logs 50  # Show 50 lines

# Stop the service
./run_service.sh stop

# Restart the service
./run_service.sh restart

# Clean up everything
./run_service.sh clean
```

### Example Workflow:

1. **First time setup:**
```bash
./run_service.sh build    # Build the service
./run_service.sh start    # Start on default port 8090
```

2. **Check if it's working:**
```bash
./run_service.sh status   # Check status
./run_service.sh logs     # View logs
```

3. **Test the API:**
```bash
# In another terminal, run the test script
./test_route_service.sh -p 8090
```

4. **Stop when done:**
```bash
./run_service.sh stop
```

### Advanced Usage:

```bash
# Start with Google Maps provider (when implemented)
./run_service.sh start 8080 30 google "" "YOUR_GOOGLE_API_KEY"

# Start with custom OSRM instance
./run_service.sh start 8080 30 openstreetmap "http://localhost:5000"

# Start with Mapbox provider (when implemented)
./run_service.sh start 8080 30 mapbox "" "YOUR_MAPBOX_TOKEN"
```

### Service Management:

- **PID File**: `./run/route-service.pid` - Stores the process ID
- **Logs**: `./logs/route-service.log` - All service output
- **Binary**: `./bin/route-service` - Compiled service binary

### Quick Test:

```bash
# Start the service
./run_service.sh start 8090

# In another terminal, test it
curl http://localhost:8090/health
curl http://localhost:8090/api/v1/provider

# Or use the test script
./test_route_service.sh -p 8090

# When done
./run_service.sh stop
```

The runner script handles all the complexity: building, starting, stopping, logging, and cleanup.
