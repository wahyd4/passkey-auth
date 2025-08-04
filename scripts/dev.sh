#!/bin/bash

set -e

echo "🔧 Starting Passkey Auth in development mode..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
go mod download

# Set development environment variables
export PORT=8080
export WEBAUTHN_RP_ID=localhost
export DATABASE_PATH=./dev-passkey-auth.db
export SESSION_SECRET=dev-secret-not-for-production

echo "🚀 Starting server..."
echo "Access the admin interface at: http://localhost:8080"
echo "Press Ctrl+C to stop"

# Run the application
go run main.go
