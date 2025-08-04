# GitHub Repository Setup Guide

This guide will help you create a GitHub repository and push your Passkey Auth project to GitHub.

## Prerequisites

- Git installed and configured
- GitHub account
- SSH keys set up with GitHub (recommended) or HTTPS access

## Step-by-Step Instructions

### 1. Create GitHub Repository

1. Go to [GitHub](https://github.com) and sign in
2. Click the "+" icon in the top right corner
3. Select "New repository"
4. Fill in the repository details:
   - **Repository name**: `passkey-auth` (or your preferred name)
   - **Description**: "WebAuthn (FIDO2) passkey authentication service for Kubernetes nginx ingress"
   - **Visibility**: Public (for open source) or Private
   - **DO NOT** initialize with README, .gitignore, or license (we already have these)
5. Click "Create repository"

### 2. Push Your Code

#### Option A: Use the Setup Script (Recommended)
```bash
# Run the automated setup script
./scripts/github-setup.sh
```

#### Option B: Manual Setup
```bash
# Set your GitHub username and repository name
GITHUB_USERNAME="your-username"
REPO_NAME="passkey-auth"

# Add remote origin
git remote add origin https://github.com/${GITHUB_USERNAME}/${REPO_NAME}.git

# Push to GitHub
git branch -M main
git push -u origin main
```

### 3. Configure Repository Settings

After pushing, configure your repository:

#### Basic Settings
1. Go to your repository on GitHub
2. Click "Settings" tab
3. Add a description: "WebAuthn (FIDO2) passkey authentication service for Kubernetes nginx ingress"
4. Add topics: `webauthn`, `fido2`, `passkey`, `kubernetes`, `authentication`, `golang`, `nginx`, `ingress`
5. Add website URL if you have a demo

#### Branch Protection (Recommended)
1. Go to Settings → Branches
2. Click "Add rule"
3. Branch name pattern: `main`
4. Enable:
   - ✅ Require a pull request before merging
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - ✅ Include administrators

#### GitHub Actions Secrets (For CI/CD)
If you want to publish Docker images:

1. Go to Settings → Secrets and variables → Actions
2. Add repository secrets:
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Your Docker Hub password or access token

### 4. Optional Enhancements

#### GitHub Pages (Documentation)
1. Go to Settings → Pages
2. Source: Deploy from a branch
3. Branch: `main`
4. Folder: `/docs` (create docs folder if needed)

#### Issue Templates
The repository already includes:
- Bug report template
- Feature request template
- Pull request template

#### Discussions
1. Go to Settings → General
2. Scroll to "Features"
3. Enable "Discussions" for community Q&A

#### Security
1. Go to Settings → Security
2. Enable "Dependency graph"
3. Enable "Dependabot alerts"
4. Enable "Dependabot security updates"

## Repository Structure

Your repository now includes:

```
passkey-auth/
├── .github/                 # GitHub templates and workflows
│   ├── ISSUE_TEMPLATE/     # Bug and feature request templates
│   ├── workflows/          # GitHub Actions CI/CD
│   └── pull_request_template.md
├── internal/               # Go application code
├── k8s/                   # Kubernetes manifests
├── scripts/               # Build and deployment scripts
├── web/                   # Web interface
├── CHANGELOG.md           # Version history
├── CONTRIBUTING.md        # Contribution guidelines
├── LICENSE               # MIT License
├── README.md             # Main documentation
├── Dockerfile            # Docker build configuration
└── Makefile              # Development commands
```

## Next Steps

1. **Star your repository** to bookmark it
2. **Watch releases** to get notified of updates
3. **Create your first release** when ready
4. **Share with the community** on relevant platforms
5. **Write blog posts** about your project
6. **Submit to awesome lists** related to authentication or Kubernetes

## Promoting Your Open Source Project

- **Reddit**: Post in r/golang, r/kubernetes, r/selfhosted
- **Hacker News**: Share when you have significant updates
- **Dev.to**: Write technical blog posts
- **Twitter/X**: Share updates and engage with the community
- **Kubernetes Slack**: Share in relevant channels
- **Awesome Lists**: Submit to awesome-go, awesome-kubernetes, etc.

## Support

If you need help with GitHub setup:
- [GitHub Docs](https://docs.github.com)
- [Git Handbook](https://guides.github.com/introduction/git-handbook/)
- [GitHub Community](https://github.community)

Happy coding! 🚀
