# SPL Toolkit - Build and Release Automation

.PHONY: build build-server build-shared build-all test test-coverage clean install lint fmt deps deps-update python-deps python-build python-test python-install python-wheel python-sdist python-dist release help generate-docs

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
SERVER_BINARY_NAME=spl-toolkit-server
SHARED_LIB_NAME=libspl_toolkit

# Build directories
BUILD_DIR=build
DIST_DIR=dist

# Python variables
PYTHON=python3
PIP=pip3

# Release version is owned by the repository VERSION file.
VERSION := $(shell cat VERSION)
VERSION_LDFLAGS=-X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version=$(VERSION)

# Operating system detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    SHARED_EXT=.so
endif
ifeq ($(UNAME_S),Darwin)
    SHARED_EXT=.dylib
    NATIVE_BUILD_ENV=MACOSX_DEPLOYMENT_TARGET=15.0 CGO_CFLAGS="$(CGO_CFLAGS) -mmacosx-version-min=15.0" CGO_LDFLAGS="$(CGO_LDFLAGS) -mmacosx-version-min=15.0"
endif
ifeq ($(OS),Windows_NT)
    SHARED_EXT=.dll
endif

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	$(GOMOD) download

deps-update: ## Intentionally update Go module metadata
	$(GOMOD) tidy

fmt: ## Format Go code
	$(GOCMD) fmt ./...

lint: ## Check handwritten Go formatting, vet, and tests
	$(PYTHON) tools/check_go.py

test: ## Run Go tests
	$(GOTEST) -mod=readonly -race ./...

test-coverage: ## Run tests and show coverage
	$(GOTEST) -mod=readonly -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build: ## Build the main binary
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -mod=readonly -trimpath -ldflags "$(VERSION_LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

build-server: ## Build the REST API server binary
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -mod=readonly -trimpath -ldflags "$(VERSION_LDFLAGS)" -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) ./cmd/server

build-shared: ## Build shared library for Python bindings
	mkdir -p $(BUILD_DIR)
	$(NATIVE_BUILD_ENV) $(GOBUILD) -mod=readonly -trimpath -buildmode=c-shared -ldflags "$(VERSION_LDFLAGS)" -o $(BUILD_DIR)/$(SHARED_LIB_NAME)$(SHARED_EXT) ./pkg/bindings

build-all: build build-server build-shared ## Build CLI, server binary, and shared library

python-deps: ## Install pinned Python development dependencies
	$(PIP) install -r python/requirements-dev.txt

python-build: python-deps ## Build self-contained Python wheel and sdist
	mkdir -p $(DIST_DIR)
	rm -f $(DIST_DIR)/spl_toolkit-*.whl $(DIST_DIR)/spl_toolkit-*.tar.gz
	$(PYTHON) -m build --no-isolation --sdist --wheel --outdir $(DIST_DIR) python

python-test: python-build ## Test installed wheels outside the checkout
	$(PYTHON) tools/check_package.py --sdist $(DIST_DIR)/spl_toolkit-$(VERSION).tar.gz --wheel-dir $(DIST_DIR)

python-install: python-wheel ## Install the built native wheel; rebuild after source changes
	$(PIP) install --force-reinstall --no-deps $(DIST_DIR)/spl_toolkit-$(VERSION)-*.whl

python-wheel: python-deps ## Build a native Python wheel
	mkdir -p $(DIST_DIR)
	rm -f $(DIST_DIR)/spl_toolkit-*.whl
	$(PYTHON) -m build --no-isolation --wheel --outdir $(DIST_DIR) python

python-sdist: python-deps ## Build a self-contained Python source distribution
	mkdir -p $(DIST_DIR)
	rm -f $(DIST_DIR)/spl_toolkit-*.tar.gz
	$(PYTHON) -m build --no-isolation --sdist --outdir $(DIST_DIR) python

python-dist: python-build ## Build Python distribution packages

install: build ## Install CLI binary to /usr/local/bin
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

install-server: build-server ## Install server binary to /usr/local/bin
	sudo cp $(BUILD_DIR)/$(SERVER_BINARY_NAME) /usr/local/bin/

install-all: install install-server ## Install both CLI and server binaries

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
	# Copy Go binaries
	cp $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/
	cp $(BUILD_DIR)/$(SERVER_BINARY_NAME) $(DIST_DIR)/
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

# Server targets
run-server: build-server ## Run the API server
	$(BUILD_DIR)/$(SERVER_BINARY_NAME)

test-server: build-server ## Test server endpoints
	$(GOTEST) -v ./pkg/api/...

# Docker targets (optional)
docker-build: ## Build Docker image
	docker build -t spl-toolkit:$(VERSION) .

docker-build-server: ## Build Docker image for server
	docker build -f Dockerfile.server -t spl-toolkit-server:$(VERSION) .

docker-test: ## Test in Docker container
	docker run --rm -v $(PWD):/workspace -w /workspace spl-toolkit:$(VERSION) make test

docker-run-server: docker-build-server ## Run server in Docker container
	docker run -p 8080:8080 spl-toolkit-server:$(VERSION)

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

# OpenAPI generation
generate-docs: ## Generate OpenAPI documentation
	@echo "Generating OpenAPI documentation..."
	@go run github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc4 init --v3.1 -g cmd/server/main.go -o docs

# Tools installation
install-tools: ## Install development tools
	$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GOCMD) install golang.org/x/tools/cmd/godoc@latest
	$(GOCMD) install github.com/swaggo/swag/v2/cmd/swag@latest
