---
title: "Configuration Guide"
layout: page
---

# Configuration Guide

SPL Toolkit uses JSON-based configuration files to define field mappings and conditional rules. This guide covers all configuration options and patterns.

## Configuration Structure

### Basic Structure

```json
{
  "version": "0.1.0",
  "name": "Configuration Name",
  "description": "Optional description of this configuration",
  "mappings": [...],
  "rules": [...]
}
```

### Required Fields

- `version`: Configuration format version (currently "1.0")
- `mappings`: Array of basic field mappings

### Optional Fields

- `name`: Human-readable configuration name
- `description`: Configuration description
- `rules`: Array of conditional mapping rules

## Basic Field Mappings

### Simple Mappings

```json
{
  "version": "0.1.0",
  "mappings": [
    {"source": "src_ip", "target": "source_ip"},
    {"source": "dst_ip", "target": "destination_ip"},
    {"source": "src_port", "target": "source_port"},
    {"source": "dst_port", "target": "destination_port"}
  ]
}
```

### Mapping with Comments

```json
{
  "version": "0.1.0",
  "name": "Network Traffic Standardization",
  "description": "Standardizes network field names across different log sources",
  "mappings": [
    {
      "source": "src_ip",
      "target": "source_ip",
      "comment": "Source IP address standardization"
    },
    {
      "source": "dst_ip", 
      "target": "destination_ip",
      "comment": "Destination IP address standardization"
    }
  ]
}
```

## Conditional Rules

Rules allow you to apply different mappings based on query context, such as sourcetype, field presence, or field values.

### Rule Structure

```json
{
  "id": "unique_rule_id",
  "name": "Human readable name",
  "description": "Optional rule description",
  "conditions": [...],
  "mappings": [...],
  "priority": 1,
  "enabled": true
}
```

### Rule Fields

- `id`: Unique identifier for the rule
- `name`: Human-readable rule name
- `conditions`: Array of conditions that must be met
- `mappings`: Array of field mappings to apply when conditions match
- `priority`: Rule priority (lower numbers = higher priority)
- `enabled`: Whether the rule is active

## Condition Types

### Sourcetype Conditions

```json
{
  "type": "sourcetype",
  "operator": "equals",
  "value": "access_combined"
}
```

**Operators**: `equals`, `contains`, `starts_with`, `ends_with`, `regex`

### Source Conditions

```json
{
  "type": "source",
  "operator": "contains",
  "value": "/var/log/apache"
}
```

### Field Existence Conditions

```json
{
  "type": "field_exists",
  "field": "clientip",
  "operator": "exists"
}
```

**Operators**: `exists`, `not_exists`

### Field Value Conditions

```json
{
  "type": "field_value",
  "field": "status",
  "operator": "equals",
  "value": "200"
}
```

**Operators**: `equals`, `not_equals`, `contains`, `greater_than`, `less_than`, `regex`

### Combination Conditions

```json
{
  "type": "combination",
  "operator": "and",
  "children": [
    {"type": "sourcetype", "operator": "equals", "value": "nginx_access"},
    {"type": "field_exists", "field": "remote_addr", "operator": "exists"}
  ]
}
```

**Operators**: `and`, `or`, `not`

## Complete Example

### Web Server Logs Configuration

```json
{
  "version": "0.1.0",
  "name": "Web Server Log Standardization",
  "description": "Normalizes field names across Apache, Nginx, and IIS logs",
  "mappings": [
    {"source": "ip", "target": "client_ip"},
    {"source": "time", "target": "timestamp"},
    {"source": "method", "target": "http_method"}
  ],
  "rules": [
    {
      "id": "apache_combined",
      "name": "Apache Combined Log Format",
      "description": "Field mappings for Apache access_combined logs",
      "conditions": [
        {
          "type": "sourcetype",
          "operator": "equals",
          "value": "access_combined"
        }
      ],
      "mappings": [
        {"source": "clientip", "target": "source_address"},
        {"source": "ident", "target": "user_identity"},
        {"source": "user", "target": "authenticated_user"},
        {"source": "timestamp", "target": "request_time"},
        {"source": "method", "target": "http_method"},
        {"source": "uri", "target": "http_uri"},
        {"source": "protocol", "target": "http_version"},
        {"source": "status", "target": "http_status_code"},
        {"source": "bytes", "target": "response_size"},
        {"source": "referer", "target": "http_referer"},
        {"source": "useragent", "target": "http_user_agent"}
      ],
      "priority": 1,
      "enabled": true
    },
    {
      "id": "nginx_access",
      "name": "Nginx Access Logs",
      "description": "Field mappings for Nginx access logs",
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
        {"source": "remote_user", "target": "authenticated_user"},
        {"source": "time_local", "target": "request_time"},
        {"source": "request", "target": "http_request"},
        {"source": "request_method", "target": "http_method"},
        {"source": "request_uri", "target": "http_uri"},
        {"source": "server_protocol", "target": "http_version"},
        {"source": "status", "target": "http_status_code"},
        {"source": "body_bytes_sent", "target": "response_size"},
        {"source": "http_referer", "target": "http_referer"},
        {"source": "http_user_agent", "target": "http_user_agent"}
      ],
      "priority": 2,
      "enabled": true
    },
    {
      "id": "iis_w3c",
      "name": "IIS W3C Extended Log Format",
      "description": "Field mappings for IIS W3C logs",
      "conditions": [
        {
          "type": "combination",
          "operator": "or",
          "children": [
            {"type": "sourcetype", "operator": "equals", "value": "iis"},
            {"type": "source", "operator": "contains", "value": "u_ex"}
          ]
        }
      ],
      "mappings": [
        {"source": "c_ip", "target": "source_address"},
        {"source": "cs_username", "target": "authenticated_user"},
        {"source": "date", "target": "request_date"},
        {"source": "time", "target": "request_time"},
        {"source": "cs_method", "target": "http_method"},
        {"source": "cs_uri_stem", "target": "http_uri"},
        {"source": "cs_version", "target": "http_version"},
        {"source": "sc_status", "target": "http_status_code"},
        {"source": "sc_bytes", "target": "response_size"},
        {"source": "cs_referer", "target": "http_referer"},
        {"source": "cs_user_agent", "target": "http_user_agent"}
      ],
      "priority": 3,
      "enabled": true
    }
  ]
}
```

## Advanced Patterns

### Priority and Precedence

Rules are evaluated in priority order (lower numbers first). The first matching rule's mappings are applied along with base mappings.

```json
{
  "rules": [
    {
      "id": "high_priority",
      "priority": 1,
      "conditions": [...],
      "mappings": [...]
    },
    {
      "id": "low_priority", 
      "priority": 10,
      "conditions": [...],
      "mappings": [...]
    }
  ]
}
```

### Conditional Field Values

Map based on specific field values:

```json
{
  "id": "security_events",
  "conditions": [
    {
      "type": "field_value",
      "field": "EventCode",
      "operator": "equals",
      "value": "4624"
    }
  ],
  "mappings": [
    {"source": "Account_Name", "target": "logon_user"},
    {"source": "Logon_Type", "target": "logon_method"}
  ]
}
```

### Regex Conditions

Use regular expressions for complex matching:

```json
{
  "conditions": [
    {
      "type": "sourcetype",
      "operator": "regex",
      "value": "^(apache|nginx|httpd)_.*"
    }
  ]
}
```

### Multiple Source Types

Handle multiple related source types:

```json
{
  "id": "firewall_logs",
  "conditions": [
    {
      "type": "sourcetype",
      "operator": "regex", 
      "value": "^(cisco_asa|palo_alto|fortinet)$"
    }
  ],
  "mappings": [
    {"source": "src", "target": "source_ip"},
    {"source": "dst", "target": "dest_ip"},
    {"source": "action", "target": "firewall_action"}
  ]
}
```

## Configuration Validation

### Required Validation

The library validates:
- JSON syntax
- Required fields presence
- Version compatibility
- Rule ID uniqueness
- Circular mapping dependencies

### Best Practices

1. **Use Descriptive IDs**: Make rule IDs meaningful
2. **Set Priorities**: Order rules by specificity
3. **Test Conditions**: Verify rule conditions match expected data
4. **Document Rules**: Add names and descriptions
5. **Enable Incrementally**: Start with basic mappings, add rules gradually

### Validation Example

```bash
# Validate configuration file
./spl-toolkit validate --config config.json

# Test configuration against sample queries
./spl-toolkit test --config config.json --queries test_queries.txt
```

## Configuration Templates

### Network Security

```json
{
  "version": "0.1.0",
  "name": "Network Security Events",
  "mappings": [
    {"source": "src_ip", "target": "source_ip"},
    {"source": "dst_ip", "target": "dest_ip"},
    {"source": "protocol", "target": "network_protocol"}
  ],
  "rules": [
    {
      "id": "firewall",
      "conditions": [{"type": "sourcetype", "operator": "contains", "value": "firewall"}],
      "mappings": [
        {"source": "action", "target": "firewall_action"},
        {"source": "rule", "target": "firewall_rule"}
      ]
    }
  ]
}
```

### Application Logs

```json
{
  "version": "0.1.0", 
  "name": "Application Logging",
  "mappings": [
    {"source": "level", "target": "log_level"},
    {"source": "msg", "target": "message"}
  ],
  "rules": [
    {
      "id": "java_app",
      "conditions": [{"type": "source", "operator": "contains", "value": ".log"}],
      "mappings": [
        {"source": "class", "target": "java_class"},
        {"source": "thread", "target": "thread_name"}
      ]
    }
  ]
}
```

## Environment Variables

Configure defaults via environment variables:

```bash
export SPL_MAPPER_CONFIG="/path/to/default/config.json"
export SPL_MAPPER_STRICT_MODE="true"
export SPL_MAPPER_DEBUG="false"
```

## Next Steps

- **[Field Mapping](features/mapping.md)** - Deep dive into mapping features
- **[Discovery](features/discovery.md)** - Learn about query discovery
- **[API Reference](api/)** - Complete API documentation
- **[Examples](examples/)** - More configuration examples