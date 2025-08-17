# Quick Start Guide

Get up and running with SPL Toolkit in minutes. This guide covers the most common use cases with practical examples.

## Installation

First, install the library following the [Installation Guide](installation.md):

```bash
git clone https://github.com/delgado-jacob/spl-toolkit.git
cd spl-toolkit
make dev-setup && make build && make python-build
```

## Basic Field Mapping

### Go Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func main() {
    // Create a new mapper
    m := mapper.New()
    
    // Define field mappings
    mappingsJSON := `[
        {"source": "src_ip", "target": "source_ip"},
        {"source": "dst_ip", "target": "destination_ip"},
        {"source": "src_port", "target": "source_port"},
        {"source": "dst_port", "target": "destination_port"}
    ]`
    
    // Load mappings
    err := m.LoadMappings([]byte(mappingsJSON))
    if err != nil {
        log.Fatal(err)
    }
    
    // Original query
    query := "search src_ip=192.168.1.1 dst_port=80 | stats count by src_ip, dst_ip"
    
    // Apply field mappings
    mappedQuery, err := m.MapQuery(query)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Original: %s\n", query)
    fmt.Printf("Mapped:   %s\n", mappedQuery)
    // Output: search source_ip=192.168.1.1 destination_port=80 | stats count by source_ip, destination_ip
}
```

### Python Example

```python
from spl_toolkit import SPLMapper

# Create mapper with basic configuration
config = {
    "version": "0.1.0",
    "mappings": [
        {"source": "src_ip", "target": "source_ip"},
        {"source": "dst_ip", "target": "destination_ip"},
        {"source": "src_port", "target": "source_port"},
        {"source": "dst_port", "target": "destination_port"}
    ]
}

mapper = SPLMapper(config=config)

# Original query
query = "search src_ip=192.168.1.1 dst_port=80 | stats count by src_ip, dst_ip"

# Apply mappings
mapped_query = mapper.map_query(query)

print(f"Original: {query}")
print(f"Mapped:   {mapped_query}")
# Output: search source_ip=192.168.1.1 destination_port=80 | stats count by source_ip, destination_ip
```

## Query Discovery

Discover what components and fields a query uses:

### Go Discovery

```go
// Discover query information
info, err := m.DiscoverQuery(query)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Input Fields: %v\n", info.InputFields)
fmt.Printf("Sources: %v\n", info.Sources)
fmt.Printf("Sourcetypes: %v\n", info.Sourcetypes)
fmt.Printf("Lookups: %v\n", info.Lookups)
fmt.Printf("Macros: %v\n", info.Macros)
```

### Python Discovery

```python
# Discover query components
info = mapper.discover_query(query)

print(f"Input Fields: {info.input_fields}")
print(f"Sources: {info.sources}")
print(f"Sourcetypes: {info.source_types}")
print(f"Lookups: {info.lookups}")
print(f"Macros: {info.macros}")
```

## Conditional Mapping Rules

Apply different mappings based on conditions:

```python
config = {
    "version": "0.1.0",
    "mappings": [
        # Default mappings
        {"source": "ip", "target": "client_ip"}
    ],
    "rules": [
        {
            "id": "apache_logs",
            "name": "Apache Log Format",
            "conditions": [
                {
                    "type": "sourcetype",
                    "operator": "equals",
                    "value": "access_combined"
                }
            ],
            "mappings": [
                {"source": "clientip", "target": "source_address"},
                {"source": "status", "target": "http_status_code"},
                {"source": "bytes", "target": "response_size"}
            ],
            "priority": 1,
            "enabled": True
        },
        {
            "id": "nginx_logs",
            "name": "Nginx Access Logs",
            "conditions": [
                {
                    "type": "combination",
                    "operator": "and",
                    "children": [
                        {"type": "sourcetype", "operator": "equals", "value": "nginx_access"},
                        {"type": "field_exists", "field": "remote_addr", "operator": "exists"}
                    ]
                }
            ],
            "mappings": [
                {"source": "remote_addr", "target": "source_address"},
                {"source": "request_status", "target": "http_status_code"}
            ],
            "priority": 2,
            "enabled": True
        }
    ]
}

mapper = SPLMapper(config=config)

# Map query with context
query = "search sourcetype=access_combined clientip=192.168.1.1 status=200"
context = {"sourcetype": "access_combined"}
mapped = mapper.map_query_with_context(query, context)

print(f"Mapped: {mapped}")
# Output: search sourcetype=access_combined source_address=192.168.1.1 http_status_code=200
```

## CLI Usage

Use the command-line interface for scripting and automation:

```bash
# Create a mapping configuration file
cat > mappings.json << EOF
{
  "version": "0.1.0",
  "mappings": [
    {"source": "src_ip", "target": "source_ip"},
    {"source": "dst_ip", "target": "destination_ip"}
  ]
}
EOF

# Map a query using CLI
./spl-toolkit map \
  --config mappings.json \
  --query "search src_ip=192.168.1.1 dst_ip=10.0.0.1"

# Discover query information
./spl-toolkit discover \
  --query "search sourcetype=apache clientip=192.168.1.1 | stats count by clientip"

# Output format options
./spl-toolkit discover \
  --query "search src_ip=192.168.1.1" \
  --format json \
  --output discovery.json
```

## Common Patterns

### Web Server Log Mapping

```python
# Common web server field mappings
web_config = {
    "version": "0.1.0",
    "rules": [
        {
            "id": "web_logs",
            "conditions": [
                {"type": "sourcetype", "operator": "contains", "value": "access"}
            ],
            "mappings": [
                {"source": "clientip", "target": "src_ip"},
                {"source": "remote_addr", "target": "src_ip"},
                {"source": "status", "target": "http_status"},
                {"source": "request_status", "target": "http_status"},
                {"source": "bytes", "target": "bytes_out"},
                {"source": "response_size", "target": "bytes_out"},
                {"source": "useragent", "target": "http_user_agent"},
                {"source": "user_agent", "target": "http_user_agent"}
            ]
        }
    ]
}
```

### Network Traffic Mapping

```python
# Network traffic field standardization
network_config = {
    "version": "0.1.0",
    "mappings": [
        {"source": "src_ip", "target": "source_ip"},
        {"source": "srcip", "target": "source_ip"},
        {"source": "source", "target": "source_ip"},
        {"source": "dst_ip", "target": "dest_ip"},
        {"source": "dstip", "target": "dest_ip"},
        {"source": "destination", "target": "dest_ip"},
        {"source": "src_port", "target": "source_port"},
        {"source": "srcport", "target": "source_port"},
        {"source": "dst_port", "target": "dest_port"},
        {"source": "dstport", "target": "dest_port"}
    ]
}
```

### Security Event Mapping

```python
# Security event normalization
security_config = {
    "version": "0.1.0",
    "rules": [
        {
            "id": "windows_security",
            "conditions": [
                {"type": "sourcetype", "operator": "equals", "value": "WinEventLog:Security"}
            ],
            "mappings": [
                {"source": "Account_Name", "target": "user"},
                {"source": "EventCode", "target": "event_id"},
                {"source": "Computer", "target": "host"},
                {"source": "LogonType", "target": "logon_type"}
            ]
        },
        {
            "id": "linux_auth",
            "conditions": [
                {"type": "source", "operator": "contains", "value": "/var/log/auth.log"}
            ],
            "mappings": [
                {"source": "user", "target": "username"},
                {"source": "pid", "target": "process_id"},
                {"source": "host", "target": "hostname"}
            ]
        }
    ]
}
```

## Next Steps

Now that you've seen the basics:

1. **[Configuration Guide](configuration.md)** - Learn advanced configuration options
2. **[Field Mapping](features/mapping.md)** - Deep dive into mapping capabilities
3. **[Discovery Engine](features/discovery.md)** - Explore discovery features
4. **[API Reference](api/)** - Complete API documentation
5. **[Examples](examples/)** - More complex examples and patterns

## Tips for Success

- **Start Simple**: Begin with basic field mappings before adding complex rules
- **Test Incrementally**: Validate mappings with sample queries
- **Use Discovery**: Understand your queries before mapping them
- **Leverage Context**: Use conditional rules for different data sources
- **Monitor Performance**: Profile large-scale mapping operations

## Getting Help

- 📖 **Documentation**: Browse the complete [documentation](index.md)
- 🐛 **Issues**: Report bugs on [GitHub Issues](https://github.com/delgado-jacob/spl-toolkit/issues)
- 💬 **Community**: Join [GitHub Discussions](https://github.com/delgado-jacob/spl-toolkit/discussions)
- 🔧 **Contributing**: See the [Contributing Guide](contributing.md)