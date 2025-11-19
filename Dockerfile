# syntax=docker/dockerfile:1

# ----- Builder stage: build all cmd/... tools -----
FROM golang:1.24-alpine AS builder

# Install necessary build dependencies for fetching modules
RUN apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /src

# Cache modules
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

# Build and install all commands under ./cmd/... into /out
ENV CGO_ENABLED=0
ENV GOBIN=/out
RUN --mount=type=cache,target=/go/pkg/mod \
    go install ./cmd/...

# ----- Final runtime image: minimal alpine with tools installed -----
FROM alpine:3.20

# Ensure common utilities are present
RUN apk add --no-cache bash curl ca-certificates && update-ca-certificates

# Copy binaries from builder
COPY --from=builder /out/ /usr/local/bin/

# Working directory for mounting the host CWD
WORKDIR /work

# Default command shows available tools if run without args
CMD ["/bin/sh", "-lc", "echo 'Tools installed:' && ls -1 /usr/local/bin && echo '\nMount your project with: docker run --rm -it -v $PWD:/work IMAGE <tool> ...'"]
