# Build stage
FROM golang:1.23-alpine AS builder

# Install git for Go modules
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (this will be cached if go.mod/go.sum don't change)
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimized flags and build cache
# Remove -a flag to avoid rebuilding standard library
# Remove unnecessary static linking flags since CGO is disabled
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o passkey-auth .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS (no sqlite library needed)
RUN apk --no-cache add ca-certificates

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
