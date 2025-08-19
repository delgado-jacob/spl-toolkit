# SPL Toolkit - Build and Release Automation

.PHONY: build test clean install lint fmt deps python-build python-test python-install release help # vet

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
# GOVET=$(GOCMD) vet  # vet not compatible with ANTLR4 generated code

# Binary names
BINARY_NAME=spl-toolkit
SHARED_LIB_NAME=libspl_toolkit

# Build directories
BUILD_DIR=build
DIST_DIR=dist

# Python variables
PYTHON=python3
PIP=pip3

# Version (can be overridden)
VERSION?=0.1.0

# Operating system detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    SHARED_EXT=.so
endif
ifeq ($(UNAME_S),Darwin)
    SHARED_EXT=.dylib
endif
ifeq ($(UNAME_S),Windows)
    SHARED_EXT=.dll
endif

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## Format Go code
	$(GOFMT) -s -w .

#vet: ## Run go vet
#	$(GOVET) ./...

lint: fmt # vet ## Run linting tools

test: deps ## Run Go tests
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build: deps lint ## Build the main binary
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

build-shared: deps lint ## Build shared library for Python bindings
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -buildmode=c-shared -o $(BUILD_DIR)/$(SHARED_LIB_NAME)$(SHARED_EXT) ./pkg/bindings

build-all: build build-shared ## Build both binary and shared library

python-deps: ## Install Python development dependencies
	$(PIP) install -r python/requirements-dev.txt

python-build: build-shared ## Build Python package
	cd python && $(PYTHON) setup.py build_ext --inplace
	cp $(BUILD_DIR)/$(SHARED_LIB_NAME)$(SHARED_EXT) python/spl_toolkit/

python-test: python-build ## Run Python tests
	cd python && $(PYTHON) -m pytest tests/ -v

python-install: python-build ## Install Python package locally
	cd python && $(PIP) install -e .

python-wheel: python-build ## Build Python wheel
	cd python && $(PYTHON) setup.py bdist_wheel

python-sdist: ## Build Python source distribution
	cd python && $(PYTHON) setup.py sdist

python-dist: python-wheel python-sdist ## Build Python distribution packages

install: build ## Install binary to /usr/local/bin
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf python/build
	rm -rf python/dist
	rm -rf python/*.egg-info
	rm -f python/spl_toolkit/*.so
	rm -f python/spl_toolkit/*.dylib
	rm -f python/spl_toolkit/*.dll
	rm -f coverage.out coverage.html

# Release targets
VERSION_FILE=VERSION
tag: ## Create and push a new version tag
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required"; exit 1; fi
	@echo $(VERSION) > $(VERSION_FILE)
	git add $(VERSION_FILE)
	git commit -m "Release version $(VERSION)"
	git tag -a v$(VERSION) -m "Release version $(VERSION)"
	git push origin v$(VERSION)
	git push origin main

release-prep: clean deps test python-test ## Prepare for release
	@echo "Release preparation complete"

release-build: release-prep build-all python-dist ## Build release artifacts
	mkdir -p $(DIST_DIR)
	# Copy Go binary
	cp $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/
	cp $(BUILD_DIR)/$(SHARED_LIB_NAME)$(SHARED_EXT) $(DIST_DIR)/
	# Copy Python distributions
	cp python/dist/* $(DIST_DIR)/ 2>/dev/null || true

release: release-build ## Create a full release
	@echo "Release $(VERSION) built successfully"
	@echo "Artifacts available in $(DIST_DIR)/"

# Development targets
dev-setup: deps python-deps ## Set up development environment
	@echo "Development environment ready"

dev-test: test python-test ## Run all tests

dev-watch: ## Watch for changes and run tests (requires entr)
	find . -name "*.go" | entr -c make test

# Docker targets (optional)
docker-build: ## Build Docker image
	docker build -t spl-toolkit:$(VERSION) .

docker-test: ## Test in Docker container
	docker run --rm -v $(PWD):/workspace -w /workspace spl-toolkit:$(VERSION) make test

# Documentation targets
docs: ## Generate documentation
	$(GOCMD) doc -all ./pkg/mapper > docs/API.md

docs-serve: ## Serve documentation locally (requires godoc)
	godoc -http=:6060

# Benchmarking
bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

# Security scanning
security: ## Run security analysis
	gosec ./...

# Tools installation
install-tools: ## Install development tools
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOGET) golang.org/x/tools/cmd/godoc@latest