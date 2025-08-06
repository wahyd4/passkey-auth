#!/bin/bash

# Helm Chart Validation Script
# This script validates the Helm chart before deployment

set -e

CHART_DIR="helm/passkey-auth"
NAMESPACE="passkey-auth-test"

echo "🔍 Validating Passkey Auth Helm Chart..."

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo "❌ Helm is not installed. Please install Helm first."
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install kubectl first."
    exit 1
fi

echo "✅ Prerequisites check passed"

# Lint the chart
echo "🔧 Linting Helm chart..."
if helm lint $CHART_DIR; then
    echo "✅ Chart linting passed"
else
    echo "❌ Chart linting failed"
    exit 1
fi

# Template the chart with test values
echo "📝 Templating chart with test values..."
helm template test-release $CHART_DIR \
    --set config.webauthn.rpId=test.example.com \
    --set config.webauthn.rpOrigins="{https://test.example.com}" \
    --set config.cors.allowedOrigins="{https://test.example.com}" \
    --set config.auth.cookieDomain=".example.com" \
    --set config.auth.allowedEmails="{admin@example.com}" \
    --set secrets.sessionSecret="test-secret-for-validation-only" \
    --set ingress.hosts[0].host=test.example.com \
    --set ingress.tls[0].hosts="{test.example.com}" \
    > /tmp/passkey-auth-template.yaml

if [ $? -eq 0 ]; then
    echo "✅ Chart templating passed"
else
    echo "❌ Chart templating failed"
    exit 1
fi

# Validate Kubernetes manifests
echo "🔍 Validating Kubernetes manifests..."
if kubectl apply --dry-run=client -f /tmp/passkey-auth-template.yaml; then
    echo "✅ Kubernetes manifest validation passed"
else
    echo "❌ Kubernetes manifest validation failed"
    exit 1
fi

# Check for required values
echo "🔧 Checking for required configuration..."

REQUIRED_VALUES=(
    "config.webauthn.rpId"
    "config.webauthn.rpOrigins"
    "secrets.sessionSecret"
)

for value in "${REQUIRED_VALUES[@]}"; do
    if helm template test-release $CHART_DIR --show-only templates/configmap.yaml | grep -q "REQUIRED"; then
        echo "❌ Required value not set: $value"
        exit 1
    fi
done

echo "✅ Required values check passed"

# Test with production values
if [ -f "$CHART_DIR/examples/values-production.yaml" ]; then
    echo "🏭 Testing with production values..."
    helm template test-release $CHART_DIR \
        -f $CHART_DIR/examples/values-production.yaml \
        --set secrets.sessionSecret="test-secret" \
        > /tmp/passkey-auth-production.yaml

    if kubectl apply --dry-run=client -f /tmp/passkey-auth-production.yaml; then
        echo "✅ Production values validation passed"
    else
        echo "❌ Production values validation failed"
        exit 1
    fi
fi

# Test with development values
if [ -f "$CHART_DIR/examples/values-development.yaml" ]; then
    echo "🚀 Testing with development values..."
    helm template test-release $CHART_DIR \
        -f $CHART_DIR/examples/values-development.yaml \
        > /tmp/passkey-auth-development.yaml

    if kubectl apply --dry-run=client -f /tmp/passkey-auth-development.yaml; then
        echo "✅ Development values validation passed"
    else
        echo "❌ Development values validation failed"
        exit 1
    fi
fi

# Cleanup
rm -f /tmp/passkey-auth-*.yaml

echo "🎉 All validations passed! The Helm chart is ready for deployment."
echo ""
echo "Next steps:"
echo "1. Update values in values.yaml or use --set flags"
echo "2. Install the chart: helm install my-passkey-auth $CHART_DIR"
echo "3. Configure your ingress controllers to use the auth backend"
echo ""
echo "For production deployment, make sure to:"
echo "• Set a secure session secret"
echo "• Configure proper domains and origins"
echo "• Set up TLS certificates"
echo "• Configure allowed email addresses"
