#!/bin/bash

set -e

echo "🚀 Deploying Passkey Auth to Kubernetes..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Create namespace
echo "Creating namespace..."
kubectl apply -f k8s/namespace.yaml

# Deploy the application
echo "Deploying application..."
kubectl apply -f k8s/deployment.yaml

# Wait for deployment to be ready
echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/passkey-auth -n passkey-auth

echo "✅ Deployment complete!"
echo ""
echo "To check the status:"
echo "  kubectl get pods -n passkey-auth"
echo ""
echo "To view logs:"
echo "  kubectl logs -f deployment/passkey-auth -n passkey-auth"
echo ""
echo "To configure nginx ingress, see k8s/ingress-example.yaml"
