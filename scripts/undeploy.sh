#!/bin/bash

set -e

echo "🧹 Undeploying Passkey Auth from Kubernetes..."

# Delete the application
echo "Deleting application..."
kubectl delete -f k8s/deployment.yaml --ignore-not-found=true

# Delete namespace (this will also delete PVC - data will be lost!)
read -p "⚠️  This will delete all data. Are you sure? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    kubectl delete -f k8s/namespace.yaml --ignore-not-found=true
    echo "✅ Undeployment complete!"
else
    echo "❌ Cancelled"
fi
