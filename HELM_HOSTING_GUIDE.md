# Hosting Passkey Auth Helm Chart on GitHub

This guide explains how to host your Helm chart on GitHub Pages and make it available via a public Helm repository.

## Prerequisites

- GitHub repository with your Helm chart
- GitHub Actions enabled
- GitHub Pages enabled

## Setup Steps

### 1. Repository Structure

Ensure your repository has this structure:
```
passkey-auth/
├── .github/
│   ├── workflows/
│   │   └── helm-release.yml
│   └── cr.yaml
├── helm/
│   └── passkey-auth/
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── README.md
│       ├── templates/
│       └── examples/
└── README.md
```

### 2. Enable GitHub Pages

1. Go to your GitHub repository
2. Navigate to **Settings** > **Pages**
3. Under **Source**, select **GitHub Actions**
4. Save the configuration

### 3. Configure Repository Settings

1. **Enable GitHub Actions**:
   - Go to **Settings** > **Actions** > **General**
   - Enable "Allow all actions and reusable workflows"

2. **Set up GitHub Pages permissions**:
   - Go to **Settings** > **Actions** > **General**
   - Under "Workflow permissions", select "Read and write permissions"
   - Check "Allow GitHub Actions to create and approve pull requests"

### 4. Update Chart Configuration

Edit `.github/cr.yaml` to match your repository:

```yaml
owner: YOUR_GITHUB_USERNAME  # Change this
git-repo: passkey-auth       # Change if different
charts-repo: https://YOUR_GITHUB_USERNAME.github.io/passkey-auth
target-branch: gh-pages
package-path: .cr-release-packages
index-path: .cr-index
skip-existing: true
```

### 5. Create Your First Release

1. **Tag your first release**:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. **Or push changes to trigger workflow**:
   ```bash
   git add .
   git commit -m "Add Helm chart"
   git push origin main
   ```

The GitHub Action will automatically:
- Lint and test your chart
- Package the chart
- Create a GitHub release
- Update the Helm repository index
- Deploy to GitHub Pages

### 6. Verify the Setup

1. **Check GitHub Actions**:
   - Go to **Actions** tab in your repository
   - Verify the "Release Helm Chart" workflow completes successfully

2. **Check GitHub Pages**:
   - Go to **Settings** > **Pages**
   - You should see "Your site is published at https://username.github.io/passkey-auth"

3. **Test the Helm repository**:
   ```bash
   helm repo add passkey-auth https://YOUR_USERNAME.github.io/passkey-auth
   helm repo update
   helm search repo passkey-auth
   ```

## Using Your Hosted Chart

### Add the Repository

```bash
helm repo add passkey-auth https://YOUR_USERNAME.github.io/passkey-auth
helm repo update
```

### Install the Chart

```bash
# Basic installation
helm install my-passkey-auth passkey-auth/passkey-auth

# With custom values
helm install my-passkey-auth passkey-auth/passkey-auth \
  --set config.webauthn.rpId=auth.example.com \
  --set secrets.sessionSecret="your-secure-secret"

# With values file
helm install my-passkey-auth passkey-auth/passkey-auth \
  -f values-production.yaml
```

### Search Available Versions

```bash
helm search repo passkey-auth --versions
```

## Maintenance and Updates

### Releasing New Versions

1. **Update Chart.yaml**:
   ```yaml
   version: 0.2.0  # Increment version
   appVersion: "v1.1.0"  # Update app version if needed
   ```

2. **Commit and push**:
   ```bash
   git add helm/passkey-auth/Chart.yaml
   git commit -m "Bump chart version to 0.2.0"
   git push origin main
   ```

3. **The workflow will automatically**:
   - Package the new version
   - Create a GitHub release
   - Update the Helm repository

### Testing Charts Locally

```bash
# Lint the chart
helm lint helm/passkey-auth/

# Template the chart (dry run)
helm template my-passkey-auth helm/passkey-auth/ \
  --set config.webauthn.rpId=test.local

# Install locally for testing
helm install test-release helm/passkey-auth/ \
  --dry-run --debug
```

## Advanced Configuration

### Custom Domain for Helm Repository

If you want to use a custom domain instead of `username.github.io`:

1. **Set up custom domain in GitHub Pages**:
   - Go to **Settings** > **Pages**
   - Add your custom domain (e.g., `charts.example.com`)

2. **Update `.github/cr.yaml`**:
   ```yaml
   charts-repo: https://charts.example.com
   ```

3. **Configure DNS**:
   - Add CNAME record pointing to `username.github.io`

### Multiple Charts in One Repository

If you have multiple charts:

```
helm/
├── passkey-auth/
│   ├── Chart.yaml
│   └── ...
├── another-chart/
│   ├── Chart.yaml
│   └── ...
```

The workflow will automatically detect and release all charts.

### Private Repositories

For private repositories, users will need:

1. **GitHub Personal Access Token**:
   ```bash
   helm repo add passkey-auth https://username:TOKEN@username.github.io/passkey-auth
   ```

2. **Or configure helm with auth**:
   ```bash
   helm repo add passkey-auth https://username.github.io/passkey-auth \
     --username YOUR_USERNAME \
     --password YOUR_TOKEN
   ```

## Troubleshooting

### Common Issues

1. **GitHub Actions Fails**:
   - Check workflow permissions in repository settings
   - Verify GitHub Pages is enabled
   - Check if there are syntax errors in the chart

2. **Chart Not Found**:
   - Verify the repository URL is correct
   - Check if GitHub Pages deployment completed
   - Ensure chart name matches directory name

3. **Permission Denied**:
   - Verify GitHub Actions has write permissions
   - Check if GitHub Pages is enabled for the repository

### Debug Commands

```bash
# Check repository status
helm repo list

# Update repositories
helm repo update

# Debug template rendering
helm template my-release helm/passkey-auth/ --debug

# Validate chart
helm lint helm/passkey-auth/

# Check chart dependencies
helm dependency list helm/passkey-auth/
```

## Security Considerations

### Repository Security

1. **Secrets Management**:
   - Never commit sensitive values to the repository
   - Use GitHub Secrets for sensitive configuration
   - Document security requirements in README

2. **Chart Signing** (Optional):
   ```bash
   # Generate GPG key for chart signing
   gpg --gen-key

   # Export public key
   gpg --armor --export your-email@example.com > public.key

   # Add to chart-releaser config
   echo "sign: true" >> .github/cr.yaml
   ```

3. **Dependency Security**:
   - Regularly update chart dependencies
   - Use dependency vulnerability scanning
   - Pin specific versions in production

### Best Practices

1. **Version Management**:
   - Follow semantic versioning
   - Update `appVersion` when application changes
   - Update `version` when chart changes

2. **Documentation**:
   - Keep README.md updated
   - Document breaking changes
   - Provide migration guides

3. **Testing**:
   - Test charts before releasing
   - Use CI/CD for automated testing
   - Validate on different Kubernetes versions

## Example Repository

You can see a complete example at: `https://github.com/YOUR_USERNAME/passkey-auth`

The hosted Helm repository will be available at: `https://YOUR_USERNAME.github.io/passkey-auth`

## Support

If you encounter issues:
1. Check the GitHub Actions logs
2. Verify chart syntax with `helm lint`
3. Review GitHub Pages deployment status
4. Open an issue in the repository for help
