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

## What is SPL Toolkit?

SPL Toolkit is a powerful library that enables programmatic analysis and manipulation of Splunk Search Processing Language (SPL) queries. Built with a **Grammar-First Architecture** using ANTLR4, it provides robust, language-aware processing that avoids fragile regex-based approaches.

## Core Capabilities

### 🔄 Field Mapping
- **Dynamic Schema Translation**: Map query fields from one schema to another using JSON configuration
- **Context-Aware Processing**: Respects derived field contexts and handles renamed fields properly
- **Token Stream Rewriting**: Preserves SPL syntax and semantics during transformations

### 🔍 Discovery Engine
- **Grammar-Aware Analysis**: Uses AST traversal to extract components from SPL queries
- **Resource Detection**: Identifies datamodels, lookups, macros, sources, and sourcetypes
- **Field Classification**: Distinguishes between input fields and derived fields with context sensitivity

### ⚙️ Advanced Features
- **Conditional Mapping Rules**: Apply mappings based on field values, sourcetypes, and complex conditions
- **DataModel Support**: Map between different datamodel structures
- **Python & Go APIs**: Full language bindings for cross-platform integration

## Quick Example

```python
from spl_toolkit import SPLMapper

# Create mapper with field mappings
config = {
    "mappings": [
        {"source": "src_ip", "target": "source_ip"},
        {"source": "dst_ip", "target": "destination_ip"}
    ]
}

mapper = SPLMapper(config=config)

# Transform a query
query = "search src_ip=192.168.1.1 dst_port=80"
mapped = mapper.map_query(query)
# Result: "search source_ip=192.168.1.1 dst_port=80"

# Discover query components
info = mapper.discover_query(query)
print(f"Input fields: {info.input_fields}")
```

## Get Started

Choose your preferred approach:

- **[Installation Guide](installation.md)** - Get up and running quickly
- **[Quick Start](quickstart.md)** - Basic usage examples
- **[API Reference](api/)** - Detailed API documentation
- **[Configuration](configuration.md)** - Advanced configuration options

## Documentation Sections

### Getting Started
- [Installation](installation.md)
- [Quick Start](quickstart.md)
- [Basic Examples](examples/basic.md)

### Core Features
- [Field Mapping](features/mapping.md)
- [Discovery Engine](features/discovery.md)
- [Configuration System](configuration.md)

### API Reference
- [Go API](api/go.md)
- [Python API](api/python.md)
- [CLI Reference](api/cli.md)

### Advanced Topics
- [Architecture](architecture.md)
- [Grammar & AST](grammar.md)
- [Performance](performance.md)
- [Contributing](contributing.md)

### Examples & Tutorials
- [Basic Usage](examples/basic.md)
- [Advanced Mapping](examples/advanced-mapping.md)
- [Discovery Examples](examples/discovery.md)
- [Integration Patterns](examples/integration.md)

## Architecture Highlights

The SPL Toolkit uses a **Grammar-First Architecture** that ensures robust and accurate SPL processing:

```
ANTLR4 Grammar → AST Generation → Listener-Based Analysis → Token Stream Rewriting
```

This approach provides:
- **Language Accuracy**: Full SPL grammar compliance
- **Robustness**: No fragile regex patterns
- **Extensibility**: Easy to add new SPL features
- **Performance**: Efficient AST-based processing

## Why Choose SPL Toolkit?

- ✅ **Grammar-Based**: Uses official SPL grammar for accurate parsing
- ✅ **Context-Aware**: Understands field derivation and scoping
- ✅ **Performance**: Optimized for production workloads
- ✅ **Cross-Language**: Go library with Python bindings
- ✅ **Well-Tested**: Comprehensive test coverage
- ✅ **Open Source**: MIT licensed with active development

## Project Status

| Phase | Status | Description |
|-------|--------|-------------|
| **Phase 1** | ✅ Complete | Basic field mapping and discovery |
| **Phase 2** | 🚧 Partial | Conditional rules and datamodel mapping |
| **Phase 3** | 🔮 Planned | Query translation (raw ↔ datamodel/tstats) |
| **Phase 4** | 🔮 Planned | Auto-mapping from dual log representations |
| **Phase 5** | 🔮 Planned | Template-based auto-mapping |

## Support & Community

- 📖 **Documentation**: You're reading it!
- 🐛 **Issues**: [GitHub Issues](https://github.com/delgado-jacob/spl-toolkit/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/delgado-jacob/spl-toolkit/discussions)
- 🔧 **Contributing**: See our [Contributing Guide](contributing.md)

---

**Note**: This is a defensive security tool designed for legitimate SPL query analysis and manipulation. It should not be used for malicious purposes.