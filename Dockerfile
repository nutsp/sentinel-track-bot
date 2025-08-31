# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fix-track-bot .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S fixtrack && \
    adduser -u 1001 -S fixtrack -G fixtrack

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/fix-track-bot .

# Copy configuration files
COPY --from=builder /build/config.example.yaml ./config.yaml

# Change ownership to non-root user
RUN chown -R fixtrack:fixtrack /app

# Switch to non-root user
USER fixtrack

# Expose port (if needed for health checks)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD pgrep fix-track-bot || exit 1

# Run the binary
ENTRYPOINT ["./fix-track-bot"]
