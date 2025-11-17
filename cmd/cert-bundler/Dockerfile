# Build and runtime image for cert-bundler
# Usage (from repo root or cmd/cert-bundler directory):
#   docker build -t cert-bundler:latest -f cmd/cert-bundler/Dockerfile .
#   docker run --rm -v "$PWD":/work cert-bundler:latest
# This expects a /work/bundle.yaml file in the mounted directory and
# will write generated bundles to /work/bundle.

# Build stage
FROM golang:1.24.3-alpine AS build
WORKDIR /src

# Copy go module files and download dependencies first for better caching
RUN go install git.wntrmute.dev/kyle/goutils/cmd/cert-bundler@v1.13.2 && \
    mv /go/bin/cert-bundler /usr/local/bin/cert-bundler

# Runtime stage (kept as golang:alpine per requirement)
FROM golang:1.24.3-alpine

# Create a work directory that users will typically mount into
WORKDIR /work
VOLUME ["/work"]

# Copy the built binary from the builder stage
COPY --from=build /usr/local/bin/cert-bundler /usr/local/bin/cert-bundler

# Default command: read bundle.yaml from current directory and output to ./bundle
ENTRYPOINT ["/usr/local/bin/cert-bundler"]
CMD ["-c", "/work/bundle.yaml", "-o", "/work/bundle"]
