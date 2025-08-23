# SPL Toolkit REST API Server

The SPL Toolkit includes a REST API server that provides programmatic access to all core functionality through HTTP endpoints.

## Quick Start

### Building and Running

```bash
# Build the server
make build-server

# Run the server
make run-server

# Or run directly
./build/spl-toolkit-server
```

The server will start on port 8080 by default. You can change the port by setting the `PORT` environment variable.

```bash
PORT=3000 ./build/spl-toolkit-server
```

### Using Docker

```bash
# Build Docker image
make docker-build-server

# Run in Docker
make docker-run-server

# Or run directly
docker run -p 8080:8080 spl-toolkit-server:0.1.1
```

## API Documentation

Once the server is running, you can access:

- **Interactive API Documentation**: http://localhost:8080/api/v1/docs
- **OpenAPI Specification**: http://localhost:8080/api/v1/openapi.json
- **Health Check**: http://localhost:8080/api/v1/health

## API Endpoints

### Health Check
```
GET /api/v1/health
```

Returns the service health status.

### Query Mapping
```
POST /api/v1/query/map
```

Apply field mappings to transform a SPL query.

**Request Body:**
```json
{
  "query": "search src_ip=192.168.1.1 | stats count by dest_port",
  "mappings": [
    {"source": "src_ip", "target": "source_ip"},
    {"source": "dest_port", "target": "destination_port"}
  ],
  "context": {
    "sourcetype": "access_combined"
  }
}
```

**Response:**
```json
{
  "original_query": "search src_ip=192.168.1.1 | stats count by dest_port",
  "mapped_query": "search source_ip=192.168.1.1 | stats count by destination_port",
  "success": true
}
```

### Query Discovery
```
POST /api/v1/query/discover
```

Analyze a SPL query to discover datamodels, lookups, fields, and other components.

**Request Body:**
```json
{
  "query": "search sourcetype=access_combined | stats count by src_ip"
}
```

**Response:**
```json
{
  "query": "search sourcetype=access_combined | stats count by src_ip",
  "query_info": {
    "datamodels": [],
    "datasets": [],
    "lookups": [],
    "macros": [],
    "sources": [],
    "sourcetypes": ["access_combined"],
    "input_fields": ["src_ip"]
  },
  "success": true
}
```

### Query Validation
```
POST /api/v1/query/validate
```

Validate SPL query syntax.

**Request Body:**
```json
{
  "query": "search index=web | stats count by status"
}
```

**Response:**
```json
{
  "query": "search index=web | stats count by status",
  "valid": true,
  "success": true
}
```

### Load Mappings
```
POST /api/v1/mappings
```

Load field mapping configuration for subsequent operations.

**Request Body (Simple Mappings):**
```json
{
  "mappings": [
    {"source": "src_ip", "target": "source_ip"},
    {"source": "dst_ip", "target": "destination_ip"}
  ]
}
```

**Request Body (Configuration with Conditional Rules):**
```json
{
  "config": {
    "version": "1.0",
    "name": "Web Access Logs Mapping",
    "mappings": [
      {"source": "ip", "target": "client_ip"}
    ],
    "rules": [
      {
        "id": "apache_logs",
        "name": "Apache Access Log Fields",
        "conditions": [
          {
            "type": "sourcetype",
            "operator": "equals",
            "value": "access_combined"
          }
        ],
        "mappings": [
          {"source": "clientip", "target": "source_address"},
          {"source": "status", "target": "http_status_code"}
        ],
        "priority": 1,
        "enabled": true
      }
    ]
  }
}
```

## Configuration

The server can be configured using environment variables:

- `PORT`: Server port (default: 8080)

## Testing

```bash
# Run API tests
make test-server

# Run all tests
make test
```

## Architecture

The REST API server is built using Go's standard library `http.ServeMux` for routing and includes:

- **Request/Response Validation**: All payloads are validated with detailed error messages
- **OpenAPI Documentation**: Comprehensive API documentation with Swagger UI
- **Middleware**: Logging, CORS, and content-type handling
- **Error Handling**: Consistent error response format
- **Health Checks**: Built-in health check endpoint
- **Graceful Shutdown**: Proper server shutdown handling

The server leverages the same core `pkg/mapper` library used by the CLI tool, ensuring consistency between interfaces.

## Library vs Microservice Usage

The project is organized to support both use cases:

### As a Library
```go
import "github.com/delgado-jacob/spl-toolkit/pkg/mapper"

m := mapper.New()
result, err := m.MapQuery("search src_ip=192.168.1.1")
```

### As a Microservice
Deploy the server binary (`spl-toolkit-server`) to provide HTTP API access to the same functionality.

This dual approach allows for maximum flexibility in how the SPL Toolkit is integrated into different environments and workflows.