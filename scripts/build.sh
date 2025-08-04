#!/bin/bash

set -e

echo "🚀 Building Passkey Auth..."

# Build Docker image
echo "Building Docker image..."
docker build -t passkey-auth:latest .

echo "✅ Build complete!"
echo ""
echo "Next steps:"
echo "1. Update k8s/deployment.yaml with your domain and session secret"
echo "2. Run ./scripts/deploy.sh to deploy to Kubernetes"
echo "3. Configure your nginx ingress to use passkey auth"
