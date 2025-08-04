# Build stage
FROM golang:1.23-bullseye AS builder

# Install build dependencies including sqlite
RUN apt-get update && apt-get install -y gcc libsqlite3-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (remove static linking for SQLite compatibility)
RUN CGO_ENABLED=1 GOOS=linux go build -a -o passkey-auth .

# Final stage
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates libsqlite3-0 && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/passkey-auth .

# Copy web files
COPY --from=builder /app/web ./web/

# Copy default config
COPY --from=builder /app/config.yaml .

# Create directory for database
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./passkey-auth"]
