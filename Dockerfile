# Build stage
FROM golang:1.23-alpine AS builder

# Install git for Go modules (no need for gcc or sqlite-dev anymore)
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with pure Go (no CGO needed)
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o passkey-auth .

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
