# Contributing to Passkey Auth

Thank you for your interest in contributing to Passkey Auth! This document provides guidelines for contributing to the project.

## 🚀 Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/passkey-auth.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Test your changes
6. Commit and push to your fork
7. Create a Pull Request

## 🛠️ Development Setup

### Prerequisites
- Go 1.21 or later
- Docker (for containerized testing)
- Make (optional, for build scripts)

### Local Development
```bash
# Clone the repository
git clone https://github.com/yourusername/passkey-auth.git
cd passkey-auth

# Install dependencies
go mod download

# Run the application
go run .

# Run tests
go test ./...
```

### Docker Development
```bash
# Build Docker image
./scripts/build.sh

# Run with Docker
docker run -p 8080:8080 passkey-auth
```

## 📋 Guidelines

### Code Style
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and small

### Commit Messages
- Use clear, descriptive commit messages
- Start with a verb in present tense ("Add", "Fix", "Update")
- Include context for why the change was made

Example:
```
Add email allowlist support for user registration

- Allows administrators to restrict registration to specific email domains
- Configurable via config.yaml or environment variables
- Backwards compatible (empty list allows all emails)
```

### Testing
- Write tests for new features
- Ensure existing tests pass
- Test both success and error cases
- Include integration tests for new endpoints

### Documentation
- Update README.md for new features
- Add inline code comments
- Update API documentation if endpoints change
- Include examples in documentation

## 🐛 Bug Reports

When reporting bugs, please include:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs or error messages

## 💡 Feature Requests

When requesting features:
- Describe the use case
- Explain why it would be valuable
- Consider backwards compatibility
- Provide examples if possible

## 🔐 Security

For security vulnerabilities:
- Do not open public issues
- Contact maintainers directly
- Provide detailed reproduction steps
- Allow time for fixes before disclosure

## 📜 Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Respect different viewpoints and experiences

## 🏷️ Release Process

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create release PR
4. Tag release after merge
5. Build and publish Docker images
6. Create GitHub release with notes

## 📞 Getting Help

- Check existing issues and documentation first
- Open an issue for bugs or feature requests
- Join discussions in existing issues
- Ask questions in issues (we're happy to help!)

Thank you for contributing! 🎉
