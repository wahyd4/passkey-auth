# 🔐 Passkey Auth for Kubernetes Nginx Ingress

A WebAuthn-based passkey authentication provider that integrates seamlessly with Kubernetes nginx ingress controllers. This service provides secure, passwordless authentication using passkeys (FIDO2/WebAuthn) and acts as an auth backend for nginx ingress.

## ✨ Features

- **Pa2. **"User not found" during login**:
   - Ensure user is registered and approved (if required)
   - Check that email address matches exactly

3. **WebAuthn errors**:dless Authentication**: Uses WebAuthn/FIDO2 passkeys for secure authentication
- **Email-Based Access Control**: Users are identified by email addresses with configurable allowlists
- **Nginx Ingress Integration**: Works as an auth backend using nginx `auth_request` directive
- **User Management**: Admin interface for managing users and their approval status
- **Access Control**: Configure allowed email addresses and approval requirements
- **Kubernetes Native**: Designed specifically for Kubernetes deployment
- **Persistent Storage**: Uses SQLite with persistent volumes for data storage
- **Modern UI**: Clean, responsive web interface for user registration and management

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   User Browser  │    │  Nginx Ingress   │    │  Your App       │
│                 │    │                  │    │                 │
│  1. Request     │───▶│  2. Auth Check   │───▶│  4. Serve App   │
│  4. Redirect    │◀───│  3. 401/302      │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              │ auth_request
                              ▼
                       ┌──────────────────┐
                       │  Passkey Auth    │
                       │                  │
                       │  - WebAuthn      │
                       │  - User Mgmt     │
                       │  - Session Mgmt  │
                       └──────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster with nginx ingress controller
- Docker
- kubectl configured to access your cluster

### 1. Clone and Build

```bash
git clone <repository-url>
cd passkey-auth

# Build the Docker image
./scripts/build.sh
```

### 2. Configure

Edit `k8s/deployment.yaml` to update:

```yaml
# Update these values in the ConfigMap
webauthn:
  rp_id: "your-domain.com"           # Your domain
  rp_origins:
    - "https://your-domain.com"       # Your domain with protocol

cors:
  allowed_origins:
    - "https://your-domain.com"       # Your domain with protocol

auth:
  session_secret: "your-secure-secret-key"  # Generate a secure random string
```

### 3. Deploy to Kubernetes

```bash
# Deploy the passkey auth service
./scripts/deploy.sh

# Verify deployment
kubectl get pods -n passkey-auth
kubectl logs -f deployment/passkey-auth -n passkey-auth
```

### 4. Configure Your App's Ingress

Update your application's ingress to use passkey auth:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: your-app-ingress
  annotations:
    # Auth backend - points to passkey auth service
    nginx.ingress.kubernetes.io/auth-url: "http://passkey-auth-service.passkey-auth.svc.cluster.local/auth"

    # Redirect unauthorized users to login page
    nginx.ingress.kubernetes.io/auth-signin: "https://your-domain.com/auth"

    # Pass user info to your app
    nginx.ingress.kubernetes.io/auth-response-headers: "X-Auth-User,X-Auth-User-ID"
spec:
  rules:
  - host: your-app.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: your-app-service
            port:
              number: 80
```

### 5. Create Ingress for Passkey Auth

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: passkey-auth-ingress
  namespace: passkey-auth
spec:
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /auth
        pathType: Prefix
        backend:
          service:
            name: passkey-auth-service
            port:
              number: 80
```

Apply the ingress:

```bash
kubectl apply -f k8s/ingress-example.yaml
```

## 👥 User Management

### Access the Admin Interface

1. Navigate to `https://your-domain.com/auth` in your browser
2. You'll see the admin dashboard with three tabs:
   - **Register User**: Register new users with passkeys
   - **Test Login**: Test authentication
   - **Manage Users**: View and manage all users

### User Registration Flow

1. **Register**: Admin enters email address and display name
2. **Email Validation**: System checks if email is in the allowlist (if configured)
3. **Passkey Creation**: Browser prompts for passkey creation (TouchID, Windows Hello, etc.)
4. **Approval**: If `require_approval` is enabled, admin must approve users manually
5. **Authentication**: Users can now authenticate with their passkeys

### User Approval Process

When `require_approval` is enabled in the configuration:

1. **New users register** but cannot authenticate until approved
2. **Admin reviews pending users** in the "Manage Users" tab
3. **Admin clicks "Approve"** next to pending users
4. **Users can now authenticate** with their passkeys

**To approve a user:**
1. Navigate to the admin interface at `https://your-domain.com/auth`
2. Click the "Manage Users" tab
3. Find users with "Pending" status
4. Click the "Approve" button next to the user
5. Confirm the approval in the dialog

**User Status Indicators:**
- 🟢 **Approved**: User can authenticate
- 🟡 **Pending**: User registered but needs approval

### Configuration Options

In `config.yaml` or environment variables:

```yaml
auth:
  require_approval: true    # Require admin approval for new users
  session_secret: "secret"  # Session encryption key
  allowed_emails:           # Email allowlist (empty = allow all)
    - "admin@company.com"
    - "user@company.com"
```

Environment variable overrides:
- `WEBAUTHN_RP_ID`: WebAuthn Relying Party ID (your domain)
- `SESSION_SECRET`: Session encryption secret
- `DATABASE_PATH`: SQLite database file path
- `PORT`: Server port (default: 8080)
- `ALLOWED_EMAILS`: Comma-separated list of allowed emails

## 🔧 Development

### Local Development

1. **Install Go dependencies**:
```bash
go mod download
```

2. **Run locally**:
```bash
# Update config.yaml for local development
go run main.go
```

3. **Access the interface**:
```
http://localhost:8080
```

### Project Structure

```
passkey-auth/
├── main.go                 # Application entry point
├── internal/
│   ├── auth/               # WebAuthn implementation
│   ├── config/             # Configuration management
│   ├── database/           # SQLite database layer
│   └── handlers/           # HTTP handlers
├── web/                    # Static web files
├── k8s/                    # Kubernetes manifests
├── scripts/                # Deployment scripts
└── config.yaml            # Configuration file
```

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/register/begin` | POST | Start passkey registration |
| `/api/register/finish` | POST | Complete passkey registration |
| `/api/login/begin` | POST | Start passkey authentication |
| `/api/login/finish` | POST | Complete passkey authentication |
| `/api/logout` | POST | Logout user |
| `/auth` | GET | Nginx auth check endpoint |
| `/api/users` | GET | List all users (admin) |
| `/api/users` | POST | Create new user (admin) |
| `/api/users/{id}` | PUT | Update user (approve/admin) |
| `/api/users/{id}` | DELETE | Delete user (admin) |
| `/health` | GET | Health check |

## 🔒 Security Considerations

### Production Deployment

1. **HTTPS Only**: Always use HTTPS in production
2. **Secure Session Secret**: Use a strong, random session secret
3. **Domain Configuration**: Ensure `rp_id` matches your domain exactly
4. **Network Security**: Use Kubernetes network policies to restrict access
5. **Regular Backups**: Backup the SQLite database regularly

### Session Configuration

```yaml
# Secure session configuration for production
auth:
  session_secret: "your-256-bit-random-key"  # Use a proper secret manager
```

### Database Security

The SQLite database contains:
- User information (email addresses, display names)
- WebAuthn credentials (public keys, metadata)
- No passwords or private keys are stored

## 🔐 Email Access Control

### Allowlist Configuration

You can control who can register by configuring an email allowlist:

```yaml
# config.yaml
auth:
  allowed_emails:
    - "admin@yourcompany.com"
    - "engineering@yourcompany.com"
    - "support@yourcompany.com"
```

Or via environment variable:
```bash
export ALLOWED_EMAILS="admin@company.com,user1@company.com,user2@company.com"
```

### Access Control Options

1. **Open Registration** (allowed_emails is empty or not set):
   - Any email address can register
   - Suitable for internal/trusted environments

2. **Allowlist Mode** (allowed_emails configured):
   - Only specified email addresses can register
   - Recommended for production environments

3. **Combined with Approval**:
   - Users must be in allowlist AND get admin approval
   - Maximum security for sensitive applications

## 📊 Monitoring

### Health Checks

The service exposes a health endpoint at `/health`:

```bash
curl http://passkey-auth-service.passkey-auth.svc.cluster.local/health
```

### Logs

View application logs:

```bash
kubectl logs -f deployment/passkey-auth -n passkey-auth
```

### Metrics

For production, consider adding metrics collection:
- Authentication success/failure rates
- User registration rates
- Session duration statistics

## 🐛 Troubleshooting

### Common Issues

1. **Docker build failures with SQLite CGO errors**:
   - The Dockerfile uses Debian-based images (golang:1.21-bullseye) instead of Alpine
   - This resolves musl vs glibc compatibility issues with go-sqlite3
   - If you encounter `pread64` or `pwrite64` errors, ensure you're using a glibc-based image

2. **WebAuthn encoding errors** (challenge not ArrayBuffer):
   - The web interface includes base64url conversion functions
   - Ensure you're using the included HTML file, not a custom frontend
   - Binary WebAuthn data must be converted between base64url and ArrayBuffer

3. **"User not found" during login**:
   - Ensure user is registered and approved (if required)
   - Check that username matches exactly

4. **WebAuthn errors**:
   - Verify `rp_id` matches your domain
   - Ensure HTTPS is used (required for WebAuthn)
   - Check browser support for WebAuthn

5. **Auth backend not working**:
   - Verify ingress annotations are correct
   - Check that the auth service is accessible from nginx
   - Review nginx ingress controller logs

6. **Session issues**:
   - Ensure session secret is consistent
   - Check cookie settings (secure flag for HTTPS)
   - Verify session storage is persistent

### Debug Commands

```bash
# Check pod status
kubectl get pods -n passkey-auth

# View logs
kubectl logs deployment/passkey-auth -n passkey-auth

# Check service
kubectl get svc -n passkey-auth

# Test auth endpoint
kubectl exec -it deployment/passkey-auth -n passkey-auth -- wget -O- http://localhost:8080/health

# Port forward for local testing
kubectl port-forward svc/passkey-auth-service 8080:80 -n passkey-auth
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## 🙏 Acknowledgments

- [go-webauthn](https://github.com/go-webauthn/webauthn) - WebAuthn library for Go
- [Gorilla](https://github.com/gorilla) - HTTP utilities for Go
- WebAuthn/FIDO2 specifications
- Kubernetes and nginx ingress controller teams

---

For support or questions, please create an issue in the repository.
