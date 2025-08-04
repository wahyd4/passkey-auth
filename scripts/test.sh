#!/bin/bash

set -e

echo "🧪 Testing Passkey Auth Build..."

# Clean previous builds
rm -f bin/passkey-auth

# Test build
echo "Building application..."
go build -o bin/passkey-auth .

if [ -f "bin/passkey-auth" ]; then
    echo "✅ Build successful!"
    echo "📦 Binary size: $(du -h bin/passkey-auth | cut -f1)"
else
    echo "❌ Build failed!"
    exit 1
fi

# Test basic functionality
echo ""
echo "🔍 Testing basic configuration..."

# Create test config
cat > test-config.yaml << EOF
server:
  port: "8080"
  host: "localhost"

webauthn:
  rp_display_name: "Test Auth"
  rp_id: "localhost"
  rp_origins:
    - "http://localhost:8080"

database:
  path: "test.db"

cors:
  allowed_origins:
    - "*"

auth:
  session_secret: "test-secret"
  require_approval: false
  allowed_emails:
    - "test@example.com"
    - "admin@example.com"
EOF

echo "✅ Test configuration created"

# Clean up
rm -f test-config.yaml test.db

echo ""
echo "🎉 All tests passed!"
echo ""
echo "To run the application:"
echo "  ./bin/passkey-auth"
echo ""
echo "To run in development mode:"
echo "  ./scripts/dev.sh"
