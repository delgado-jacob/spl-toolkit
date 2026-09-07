FROM golang:1.26.8-bookworm@sha256:9fdc884aacc3bec89b20ffc69f4bb369c78210e3e4f600387b5128b12c199f81 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY VERSION LICENSE README.md ./
COPY cmd ./cmd
COPY internal ./internal
COPY parser ./parser
COPY pkg ./pkg
COPY python ./python

RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 python3-pip python3-venv \
    && rm -rf /var/lib/apt/lists/*
RUN python3 -m pip install --break-system-packages --no-cache-dir -r python/requirements-build.txt
RUN VERSION_VALUE="$(cat VERSION)" \
    && CGO_ENABLED=0 GOTOOLCHAIN=local go build -mod=readonly -trimpath -buildvcs=false \
       -ldflags="-buildid= -X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version=${VERSION_VALUE}" \
       -o /out/spl-toolkit ./cmd
RUN mkdir -p /tmp/spl-native /out/wheel \
    && GOTMPDIR=/tmp/spl-native SOURCE_DATE_EPOCH=0 GOTOOLCHAIN=local \
       python3 -m build --no-isolation --wheel --outdir /out/wheel python

FROM python:3.11.9-slim-bookworm@sha256:8fb099199b9f2d70342674bd9dbccd3ed03a258f26bbd1d556822c6dfc60c317

RUN useradd --create-home --uid 1000 spluser
COPY --from=builder /out/spl-toolkit /usr/local/bin/spl-toolkit
COPY --from=builder /out/wheel/*.whl /tmp/spl-toolkit-wheel/
RUN python -m pip install --no-cache-dir --no-deps /tmp/spl-toolkit-wheel/*.whl \
    && rm -rf /tmp/spl-toolkit-wheel
WORKDIR /app
USER spluser
ENTRYPOINT ["spl-toolkit"]
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 CMD ["spl-toolkit", "version"]
LABEL org.opencontainers.image.title="SPL Toolkit" \
      org.opencontainers.image.description="Offline SPL field mapping and discovery" \
      org.opencontainers.image.source="https://github.com/delgado-jacob/spl-toolkit"
