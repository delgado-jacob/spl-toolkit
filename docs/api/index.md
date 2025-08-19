---
title: "API Reference"
layout: page
---

# API Reference

Complete reference documentation for the SPL Toolkit library APIs.

## Available APIs

The SPL Toolkit provides multiple interfaces for different use cases:

### [Go API](go.md)
Native Go library for direct integration into Go applications.

- **Core Package**: `github.com/delgado-jacob/spl-toolkit/pkg/mapper`
- **Performance**: Zero-overhead native Go performance
- **Use Case**: Go applications, microservices, CLI tools

### [Python API](python.md)  
Python bindings with C shared library integration.

- **Package**: `spl_toolkit`
- **Performance**: Near-native performance via CGO bindings
- **Use Case**: Python applications, data analysis, Jupyter notebooks

### [CLI Reference](cli.md)
Command-line interface for scripting and automation.

- **Binary**: `spl-toolkit`
- **Performance**: Suitable for scripting and batch processing
- **Use Case**: Shell scripts, CI/CD pipelines, interactive usage

## Common Operations

All APIs support these core operations:

### Field Mapping
Transform SPL queries by applying field name mappings based on configuration rules.

```bash
# CLI
spl-toolkit map --config config.json --query "search src_ip=192.168.1.1"
```

```go
// Go
mapper := mapper.New()
mapper.LoadMappings(configBytes)
result, err := mapper.MapQuery(query)
```

```python
# Python
mapper = SPLMapper(config=config)
result = mapper.map_query(query)
```

### Query Discovery
Extract components and metadata from SPL queries.

```bash
# CLI
spl-toolkit discover --query "search sourcetype=apache | stats count by clientip"
```

```go
// Go
info, err := mapper.DiscoverQuery(query)
```

```python
# Python
info = mapper.discover_query(query)
```

### Configuration Management
Load, validate, and manage mapping configurations.

```bash
# CLI
spl-toolkit validate --config config.json
```

```go
// Go
err := mapper.LoadMappings(configBytes)
```

```python
# Python
mapper = SPLMapper(config=config_dict)
```

## Response Formats

### Mapping Results

All APIs return consistent mapping results:

```json
{
  "original_query": "search src_ip=192.168.1.1",
  "mapped_query": "search source_ip=192.168.1.1",
  "applied_rules": ["network_standardization"],
  "field_mappings": {
    "src_ip": "source_ip"
  }
}
```

### Discovery Results

Query discovery returns structured information:

```json
{
  "input_fields": ["src_ip", "dst_port"],
  "derived_fields": ["calculated_field"],
  "sources": ["/var/log/apache/access.log"],
  "source_types": ["access_combined"],
  "lookups": ["geo_lookup"],
  "macros": ["security_filter"],
  "datamodels": ["Network_Traffic"],
  "commands": ["search", "stats", "eval"]
}
```

## Error Handling

### Common Error Types

All APIs use consistent error handling:

```json
{
  "error": "ParseError",
  "message": "Invalid SPL syntax at line 1, position 15",
  "code": "E001",
  "details": {
    "line": 1,
    "position": 15,
    "context": "search src_ip=192.168.1.1 |"
  }
}
```

### Error Codes

| Code | Type | Description |
|------|------|-------------|
| E001 | ParseError | SPL syntax error |
| E002 | ConfigError | Invalid configuration |
| E003 | MappingError | Field mapping failure |
| E004 | ValidationError | Input validation failure |
| E005 | InternalError | Unexpected internal error |

## Performance Considerations

### Go API
- **Fastest**: Direct native execution
- **Memory**: ~2MB base + ~10KB per rule
- **Throughput**: ~10,000 queries/second

### Python API
- **Performance**: 95% of native Go performance
- **Memory**: ~3MB base + ~15KB per rule  
- **Throughput**: ~8,000 queries/second

### CLI
- **Performance**: Includes process startup overhead
- **Memory**: Same as Go API
- **Throughput**: ~100 queries/second (due to process overhead)

## Threading and Concurrency

### Go API
```go
// Thread-safe after initialization
mapper := mapper.New()
mapper.LoadMappings(config)

// Safe to use from multiple goroutines
go func() {
    result, _ := mapper.MapQuery(query1)
}()
go func() {
    result, _ := mapper.MapQuery(query2)  
}()
```

### Python API
```python
# Thread-safe after initialization
mapper = SPLMapper(config=config)

# Safe to use from multiple threads
import threading

def worker(query):
    result = mapper.map_query(query)

threading.Thread(target=worker, args=(query1,)).start()
threading.Thread(target=worker, args=(query2,)).start()
```

## Configuration Loading

### File-based Configuration
```bash
# CLI
spl-toolkit map --config /path/to/config.json

# Go
configData, _ := os.ReadFile("/path/to/config.json")
mapper.LoadMappings(configData)

# Python
with open("/path/to/config.json") as f:
    config = json.load(f)
mapper = SPLMapper(config=config)
```

### Programmatic Configuration
```go
// Go
config := mapper.Config{
    Version: "1.0",
    Mappings: []mapper.FieldMapping{
        {Source: "src_ip", Target: "source_ip"},
    },
}
mapper.LoadConfig(config)
```

```python
# Python
config = {
    "version": "1.0",
    "mappings": [
        {"source": "src_ip", "target": "source_ip"}
    ]
}
mapper = SPLMapper(config=config)
```

## Advanced Features

### Context-Aware Mapping
Apply different mappings based on query context:

```python
# Python with context
context = {"sourcetype": "apache_access"}
result = mapper.map_query_with_context(query, context)
```

### Batch Processing
Process multiple queries efficiently:

```go
// Go batch processing
queries := []string{query1, query2, query3}
results, err := mapper.MapQueries(queries)
```

### Custom Grammar Extensions
Extend the SPL grammar for custom syntax:

```go
// Go grammar extension
mapper.RegisterCustomCommand("mycommand", customHandler)
```

## Integration Examples

### Web Service Integration
```go
// Go HTTP handler
func mapQueryHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    result, err := mapper.MapQuery(query)
    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }
    json.NewEncoder(w).Encode(result)
}
```

### Jupyter Notebook Integration
```python
# Python Jupyter integration
import spl_toolkit

def map_spl(query, config_file="config.json"):
    with open(config_file) as f:
        config = json.load(f)
    
    mapper = spl_toolkit.SPLMapper(config=config)
    return mapper.map_query(query)

# Use in notebook cell
result = map_spl("search src_ip=192.168.1.1")
print(result)
```

## Next Steps

Choose your preferred API for detailed documentation:

- **[Go API Documentation](go.md)** - Complete Go API reference
- **[Python API Documentation](python.md)** - Complete Python API reference  
- **[CLI Documentation](cli.md)** - Command-line interface reference

Or explore related topics:

- **[Configuration Guide](../configuration.md)** - Learn about configuration options
- **[Examples](../examples/)** - See practical usage examples
- **[Architecture](../architecture.md)** - Understand the technical architecture