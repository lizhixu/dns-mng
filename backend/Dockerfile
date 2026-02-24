# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o dns-mng .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite wget

# Copy binary from builder
COPY --from=builder /app/dns-mng .

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./dns-mng"]
