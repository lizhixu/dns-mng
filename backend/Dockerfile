# syntax=docker/dockerfile:1.6

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# git is required by `go mod download` for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build a static, fully-stripped binary.
# - CGO_ENABLED=0: pure Go, no C toolchain needed (the project uses modernc.org/sqlite).
# - go build cache + --mount=type=cache reuse modules and build cache across builds.
# - -trimpath -ldflags="-s -w": strip paths and debug info, smaller binary.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o dns-mng .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# ca-certificates: TLS for HTTPS calls to DNS providers.
# wget: used by the healthcheck defined in docker-compose.yaml.
RUN apk --no-cache add ca-certificates wget

# Copy binary from builder
COPY --from=builder /app/dns-mng .

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./dns-mng"]
