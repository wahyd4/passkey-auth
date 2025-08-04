#!/bin/bash

# GitHub Repository Setup Script
# Run this script after creating the repository on GitHub

set -e

echo "🚀 Passkey Auth - GitHub Repository Setup"
echo "=========================================="

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ Error: Not in a git repository. Run this from the project root."
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "⚠️  Warning: You have uncommitted changes."
    echo "Please commit or stash them before proceeding."
    exit 1
fi

# Prompt for GitHub username and repository name
read -p "Enter your GitHub username: " GITHUB_USERNAME
read -p "Enter the repository name (default: passkey-auth): " REPO_NAME
REPO_NAME=${REPO_NAME:-passkey-auth}

# Set the repository URL
REPO_URL="https://github.com/${GITHUB_USERNAME}/${REPO_NAME}.git"

echo ""
echo "📋 Repository Details:"
echo "   Username: ${GITHUB_USERNAME}"
echo "   Repository: ${REPO_NAME}"
echo "   URL: ${REPO_URL}"
echo ""

# Confirm before proceeding
read -p "Do you want to proceed? (y/N): " CONFIRM
if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo "⚙️  Setting up remote repository..."

# Add remote origin
if git remote get-url origin >/dev/null 2>&1; then
    echo "📝 Updating existing origin remote..."
    git remote set-url origin "${REPO_URL}"
else
    echo "📝 Adding origin remote..."
    git remote add origin "${REPO_URL}"
fi

# Set upstream branch and push
echo "📤 Pushing to GitHub..."
git branch -M main
git push -u origin main

echo ""
echo "✅ Success! Your repository has been pushed to GitHub."
echo ""
echo "🔗 Repository URL: https://github.com/${GITHUB_USERNAME}/${REPO_NAME}"
echo ""
echo "📋 Next Steps:"
echo "   1. Visit your repository on GitHub"
echo "   2. Add repository description and topics"
echo "   3. Configure branch protection rules (optional)"
echo "   4. Set up GitHub Pages for documentation (optional)"
echo "   5. Configure secrets for GitHub Actions:"
echo "      - DOCKER_USERNAME (for Docker Hub publishing)"
echo "      - DOCKER_PASSWORD (for Docker Hub publishing)"
echo ""
echo "🎉 Your open source project is now live!"
