# Installation Guide

This guide covers multiple installation methods for the SPL Toolkit library.

## Prerequisites

- **Go**: 1.22+ (required for Go usage and building from source)
- **Python**: 3.8+ (required for Python bindings)
- **Make**: For building and development
- **Git**: For cloning the repository

## Installation Methods

### 1. From Source (Recommended)

Building from source gives you the latest features and allows customization:

```bash
# Clone the repository
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

### 2. Go Module

For Go projects, add the module dependency:

```bash
go get github.com/delgado-jacob/spl-toolkit
```

Then import in your Go code:

```go
import "github.com/delgado-jacob/spl-toolkit/pkg/mapper"
```

### 3. Python Package (Future)

> **Note**: PyPI package is planned for Phase 2. Currently use source installation.

```bash
# Future PyPI installation
pip install spl-toolkit
```

### 4. Docker

Use the containerized version for isolated execution:

```bash
# Build Docker image
make docker-build

# Run with Docker
docker run --rm -v $(PWD):/workspace -w /workspace spl-toolkit:1.0.0 --help

# Example usage
docker run --rm -v $(PWD):/workspace -w /workspace spl-toolkit:1.0.0 \
  map --config mappings.json --query "search src_ip=192.168.1.1"
```

### 5. Pre-built Binaries (Future)

> **Note**: Pre-built binaries will be available in GitHub releases for Phase 2.

```bash
# Future release download
curl -L https://github.com/delgado-jacob/spl-toolkit/releases/latest/download/spl-toolkit-linux-amd64 -o spl-toolkit
chmod +x spl-toolkit
```

## Verification

### Test Go Installation

```go
package main

import (
    "fmt"
    "github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func main() {
    m := mapper.New()
    fmt.Println("SPL Toolkit installed successfully!")
}
```

### Test Python Installation

```python
from spl_toolkit import SPLMapper

mapper = SPLMapper()
print("SPL Toolkit Python bindings installed successfully!")
```

### Test CLI

```bash
# Test CLI installation
./spl-toolkit version

# Test basic functionality
echo '{"mappings":[{"source":"src_ip","target":"source_ip"}]}' > test-config.json
./spl-toolkit map --config test-config.json --query "search src_ip=192.168.1.1"
```

## Development Setup

For contributors and advanced users:

```bash
# Clone and setup
git clone https://github.com/delgado-jacob/spl-toolkit.git
cd spl-toolkit

# Install development dependencies
make dev-setup

# Run full test suite
make dev-test

# Build all components
make build-all

# Format and lint code
make fmt lint
```

### Development Dependencies

The development setup installs:

- **Go dependencies**: From `go.mod`
- **Python dependencies**: From `python/requirements-dev.txt`
- **ANTLR4 runtime**: For grammar parsing
- **Testing tools**: For comprehensive testing
- **Linting tools**: For code quality

## Platform-Specific Notes

### macOS

```bash
# Install dependencies via Homebrew
brew install go python@3.11 make

# Continue with source installation
```

### Linux (Ubuntu/Debian)

```bash
# Install dependencies
sudo apt update
sudo apt install golang-go python3 python3-pip make build-essential

# Continue with source installation
```

### Windows

```powershell
# Install dependencies via Chocolatey
choco install golang python make

# Or use WSL for Linux-like environment
```

## Troubleshooting

### Common Issues

**Go Module Issues**
```bash
# Clear module cache
go clean -modcache
go mod download
```

**Python Binding Issues**
```bash
# Rebuild Python bindings
make python-clean
make python-build
```

**Permission Issues**
```bash
# Fix permissions on Unix systems
chmod +x spl-toolkit
```

**Missing Dependencies**
```bash
# Reinstall development dependencies
make dev-setup
```

### Performance Optimization

For production deployments:

```bash
# Build optimized binaries
make build-release

# Use static linking for deployment
CGO_ENABLED=1 go build -ldflags '-extldflags "-static"' ./cmd/main.go
```

## Next Steps

After installation:

1. **[Quick Start Guide](quickstart.md)** - Learn basic usage
2. **[Configuration](configuration.md)** - Set up field mappings
3. **[API Reference](api/)** - Explore the full API
4. **[Examples](examples/)** - See practical examples

## Support

If you encounter installation issues:

- Check the [Troubleshooting](troubleshooting.md) guide
- Search [GitHub Issues](https://github.com/delgado-jacob/spl-toolkit/issues)
- Ask for help in [GitHub Discussions](https://github.com/delgado-jacob/spl-toolkit/discussions)