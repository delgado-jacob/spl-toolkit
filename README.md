---
title: "SPL Toolkit"
layout: page
description: "A robust, language-aware library for programmatic analysis and manipulation of Splunk SPL queries"
---

# SPL Toolkit

A robust, language-aware library for programmatic analysis and manipulation of Splunk SPL queries, written in Go with Python bindings.

[![CI/CD Pipeline](https://github.com/delgado-jacob/spl-toolkit/actions/workflows/ci.yml/badge.svg)](https://github.com/delgado-jacob/spl-toolkit/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/delgado-jacob/spl-toolkit)](https://goreportcard.com/report/github.com/delgado-jacob/spl-toolkit)
[![GoDoc](https://godoc.org/github.com/delgado-jacob/spl-toolkit?status.svg)](https://godoc.org/github.com/delgado-jacob/spl-toolkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

### Phase 1 ✅ (Current Implementation)
- **Field Mapping**: Dynamic mapping of query fields from one schema to another using JSON configuration
- **SPL Parsing**: Robust SPL query parsing using ANTLR4 grammar with AST-based processing
- **Discovery**: Extract datamodels, datasets, lookups, sources, sourcetypes, and input fields from queries
- **Token Stream Rewriting**: Context-aware field mapping that preserves SPL syntax and semantics
- **Python Bindings**: Full Python API with C shared library integration
- **Conditional Mapping**: Basic rule-based field mappings with conditions

### Phase 2 🚧 (Partially Implemented)
- **Advanced Conditional Rules**: Enhanced rule-based field mappings with complex conditions
- **DataModel Mapping**: Map between different datamodel structures (basic support available)

### Phase 3 🔮 (Future)
- **Query Translation**: Convert between raw searches and datamodel/tstats queries
- **Index ↔ DataModel**: Translate queries between index-based and datamodel-based approaches

### Phase 4 🔮 (Future)
- **Auto-mapping**: Generate mapping tables from two log representations of the same data

### Phase 5 🔮 (Future)
- **Template-based**: Auto-generate mappings from Splunk event templates

## Quick Start

### Go Usage

```go
package main

import (
    "fmt"
    "github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func main() {
    // Create a new mapper
    m := mapper.New()
    
    // Load basic field mappings
    mappingsJSON := `[
        {"source": "src_ip", "target": "source_ip"},
        {"source": "dst_ip", "target": "destination_ip"}
    ]`
    m.LoadMappings([]byte(mappingsJSON))
    
    // Map a query
    query := "search src_ip=192.168.1.1 dst_port=80"
    mappedQuery, err := m.MapQuery(query)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Original: %s\n", query)
    fmt.Printf("Mapped: %s\n", mappedQuery)
    
    // Discover query information
    info, err := m.DiscoverQuery(query)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Input fields: %v\n", info.InputFields)
}
```

### Python Usage

```python
from spl_toolkit import SPLMapper

# Create mapper with configuration
config = {
    "version": "1.0",
    "mappings": [
        {"source": "src_ip", "target": "source_ip"},
        {"source": "dst_ip", "target": "destination_ip"}
    ],
    "rules": [
        {
            "id": "apache_logs",
            "conditions": [
                {"type": "sourcetype", "operator": "equals", "value": "access_combined"}
            ],
            "mappings": [
                {"source": "clientip", "target": "source_address"}
            ],
            "enabled": True
        }
    ]
}

mapper = SPLMapper(config=config)

# Map a query with context
query = "search sourcetype=access_combined clientip=192.168.1.1"
context = {"sourcetype": "access_combined"}
mapped = mapper.map_query_with_context(query, context)

print(f"Mapped: {mapped}")

# Discover query information
info = mapper.discover_query(query)
print(f"Source types: {info.source_types}")
print(f"Input fields: {info.input_fields}")
```

## Installation

### From Source (Recommended)

```bash
# Clone repository
git clone https://github.com/delgado-jacob/spl-toolkit.git
cd spl-toolkit

# Setup development environment
make dev-setup

# Build Go library
make build

# Build Python bindings
make python-build

# Install Python package locally
make python-install

# Run tests to verify installation
make dev-test
```

### Go Module

```bash
go get github.com/delgado-jacob/spl-toolkit
```

### Docker

```bash
# Build Docker image
make docker-build

# Run with Docker
docker run --rm -v $(PWD):/workspace -w /workspace spl-toolkit:0.1.1 --help
```

### Requirements

- Go 1.22+
- Python 3.8+ (for Python bindings)
- Make
- Git

## Configuration Format

### Basic Mapping

```json
{
  "version": "0.1.1",
  "name": "Basic Field Mappings",
  "mappings": [
    {"source": "src_ip", "target": "source_ip"},
    {"source": "dst_ip", "target": "destination_ip"},
    {"source": "src_port", "target": "source_port"}
  ]
}
```

### Conditional Rules

```json
{
  "version": "0.1.1",
  "name": "Web Server Logs",
  "mappings": [
    {"source": "ip", "target": "client_ip"}
  ],
  "rules": [
    {
      "id": "apache_combined",
      "name": "Apache Combined Log Format",
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
      "enabled": true
    },
    {
      "id": "nginx_access",
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
      "enabled": true
    }
  ]
}
```

## Discovery Capabilities

The library can automatically discover and extract:

- **DataModels**: `| datamodel Network_Traffic` → `["Network_Traffic"]`
- **Lookups**: `| inputlookup ip_geo.csv` → `["ip_geo"]`
- **Macros**: `\`get_indexes(security)\`` → `["get_indexes"]`
- **Sources**: `source="/var/log/apache2/access.log"` → `["/var/log/apache2/access.log"]`
- **Sourcetypes**: `sourcetype=access_combined` → `["access_combined"]`
- **Input Fields**: All field references required for the query to function

## Development

### Prerequisites

- Go 1.22+ (current project uses Go 1.22)
- Python 3.8+
- Make
- Git

### Setup Development Environment

```bash
# Clone and setup
git clone https://github.com/delgado-jacob/spl-toolkit.git
cd spl-toolkit

# Install dependencies
make dev-setup

# Run tests
make dev-test

# Build everything
make build-all
```

### Running Tests

```bash
# Go tests
make test

# Python tests  
make python-test

# All tests
make dev-test

# With coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Lint code
make lint

# Security scan
make security
```

## Architecture

The SPL Toolkit uses a **Grammar-First Architecture** built on ANTLR4 for robust SPL parsing and analysis:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Python API    │    │     Go Library   │    │  ANTLR4 Parser  │
│                 │    │                  │    │                 │
│  ┌───────────┐  │    │  ┌─────────────┐ │    │  ┌───────────┐  │
│  │   Mapper  │  │◄──►│  │   Mapper    │ │◄──►│  │ SPL Grammar │  │
│  └───────────┘  │    │  └─────────────┘ │    │  └───────────┘  │
│                 │    │                  │    │                 │
│  ┌───────────┐  │    │  ┌─────────────┐ │    │  ┌───────────┐  │
│  │QueryInfo  │  │    │  │ Discovery   │ │    │  │  AST Tree   │
│  └───────────┘  │    │  └─────────────┘ │    │  └───────────┘  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  C Bindings     │    │Token Stream      │    │ AST Listeners   │
│  (cgo)          │    │Rewriting Engine  │    │ & Visitors      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Key Components

- **ANTLR4 Grammar**: Complete SPL language definition for accurate parsing
- **AST-Based Processing**: Uses listener patterns for robust language-aware analysis
- **Token Stream Rewriting**: Preserves query structure while applying field mappings
- **Context-Aware Discovery**: Distinguishes input fields from derived fields with hierarchical context tracking
- **Python/Go Interop**: C shared library bindings for cross-language functionality

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run tests: `make dev-test`
5. Commit your changes: `git commit -am 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Commit Message Format

```
type(scope): subject

body

footer
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

## Roadmap

- [x] **Phase 1**: Basic field mapping and discovery ✅
- [x] **Phase 2**: Conditional rules and datamodel mapping 🚧 (Partially Complete)
- [ ] **Phase 3**: Query translation (raw ↔ datamodel/tstats)
- [ ] **Phase 4**: Auto-mapping from dual log representations
- [ ] **Phase 5**: Template-based auto-mapping

## Performance

Current implementation benchmarks (Go 1.22 on modern hardware):

- **Parse Query**: ~100μs for typical queries using ANTLR4
- **Apply Mappings**: ~50μs for 100 field mappings with token stream rewriting
- **Discovery**: ~200μs for complex queries with full AST traversal
- **Memory Usage**: ~2MB base + ~10KB per mapping rule
- **Test Coverage**: 64.1% with comprehensive test suite

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- ANTLR4 for the parsing framework
- Clemens Sageder for the original SPL grammar
- The Splunk community for inspiration and requirements

## Support

- 📖 [Documentation](https://delgado-jacob.github.io/spl-toolkit/)
- 🐛 [Issue Tracker](https://github.com/delgado-jacob/spl-toolkit/issues)
- 💬 [Discussions](https://github.com/delgado-jacob/spl-toolkit/discussions)

---

**Note**: This is a defensive security tool designed for legitimate SPL query analysis and manipulation. It should not be used for malicious purposes.