# Multi-stage build for Moon - Dynamic Headless Engine
# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Install build tools: gcc and musl-dev are required for CGO (go-sqlite3)
RUN apk add --no-cache ca-certificates gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Create the default data directory so it exists in the final image
RUN mkdir -p /opt/moon

# Build a fully static binary using musl's static libc so the scratch image works.
# CGO must be enabled for go-sqlite3.
RUN CGO_ENABLED=1 GOOS=linux go build -a \
    -ldflags="-w -s -extldflags '-static'" \
    -o moon ./cmd

# Stage 2: Runtime - using scratch for minimal image
FROM scratch

# Copy CA certificates from builder (needed for HTTPS if any)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the pre-created data directory so SQLite can create the database file
COPY --from=builder /opt/moon /opt/moon

# Copy binary from builder
COPY --from=builder /build/moon /usr/local/bin/moon

# Expose default port
EXPOSE 6006

# Run moon in foreground
CMD ["/usr/local/bin/moon"]
