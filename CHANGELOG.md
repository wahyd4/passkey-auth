# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Passkey Auth
- WebAuthn (FIDO2) authentication support
- SQLite database with user management
- Email-based user identification
- Admin approval workflow for new users
- Email allowlist support
- Session management with secure cookies
- Kubernetes deployment configuration
- nginx ingress auth backend support
- Docker containerization with Debian base
- Web-based admin interface
- User registration and login flows
- Real-time WebAuthn credential management

### Features
- **Authentication**: WebAuthn/FIDO2 passkey authentication
- **Database**: SQLite with user and credential storage
- **Admin Interface**: Web UI for user management and approval
- **Email Allowlist**: Restrict registration to specific email addresses
- **Session Management**: Secure session handling with configurable secrets
- **Docker Support**: Multi-stage builds with Debian base for SQLite compatibility
- **Kubernetes Ready**: Complete deployment manifests and ingress configuration
- **nginx Integration**: Auth backend for protecting upstream services

### Security
- WebAuthn challenge/response authentication
- Secure session cookie handling
- Admin approval workflow
- Email-based access control
- HTTPS-ready configuration

### Documentation
- Comprehensive README with setup instructions
- API endpoint documentation
- Kubernetes deployment guide
- Docker usage examples
- Troubleshooting guide
- Contributing guidelines

### Technical Details
- Go 1.21+ with CGO for SQLite
- Base64URL encoding/decoding for WebAuthn data
- CORS support for cross-origin requests
- Environment variable configuration
- Health check endpoints
- Structured logging

## [0.1.0] - 2025-08-04

### Added
- Initial project structure
- Core WebAuthn implementation
- SQLite database integration
- Basic web interface
- Docker containerization
- Kubernetes manifests

---

**Note**: This project follows semantic versioning. For upgrade instructions and breaking changes, see the README.md file.
