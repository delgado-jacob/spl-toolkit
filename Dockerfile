# Multi-stage Dockerfile for SPL Toolkit

# Build stage
FROM golang:1.22-bullseye AS builder

# Set working directory
WORKDIR /build

# Copy Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Ensure dependencies and go.sum are up-to-date (resolves missing checksum issues)
RUN go mod tidy

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o spl-toolkit ./cmd

# Build the shared library
RUN CGO_ENABLED=1 GOOS=linux go build -buildmode=c-shared -o libspl_toolkit.so ./pkg/bindings

# Final stage
FROM python:3.11-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libc6 \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -u 1000 spluser

# Set working directory
WORKDIR /app

# Copy binaries from build stage
COPY --from=builder /build/spl-toolkit /usr/local/bin/
COPY --from=builder /build/libspl_toolkit.so /usr/local/lib/

# Copy Python package
COPY python/ ./python/
COPY --from=builder /build/libspl_toolkit.so ./python/spl_toolkit/

# Install Python package
RUN cd python && pip install --no-cache-dir .

# Set environment variables
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
ENV PATH=/usr/local/bin:$PATH

# Switch to non-root user
USER spluser

# Set entrypoint
ENTRYPOINT ["spl-toolkit"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD spl-toolkit --version || exit 1

# Labels
LABEL org.opencontainers.image.title="SPL Toolkit"
LABEL org.opencontainers.image.description="Programmatic analysis and manipulation of Splunk SPL queries"
LABEL org.opencontainers.image.vendor="SPL Toolkit Team"
LABEL org.opencontainers.image.source="https://github.com/delgado-jacob/spl-toolkit"
LABEL org.opencontainers.image.documentation="https://github.com/delgado-jacob/spl-toolkit/docs"