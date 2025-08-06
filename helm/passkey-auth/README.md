# Passkey Auth Helm Chart

A Helm chart for deploying Passkey Auth, a WebAuthn-based passkey authentication provider that integrates with Kubernetes Nginx Ingress controller.

## Overview

This chart deploys a secure, passwordless authentication service using WebAuthn/FIDO2 passkeys.

## TL;DR

```bash
helm repo add passkey-auth https://your-github-username.github.io/passkey-auth-helm
helm repo update
helm install my-passkey-auth passkey-auth/passkey-auth \
  --set config.webauthn.rpId=auth.example.com \
  --set config.webauthn.rpOrigins="{https://auth.example.com}" \
  --set secrets.sessionSecret="your-very-long-random-secret-key-here"
```

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Nginx Ingress Controller
- Cert-Manager (for TLS certificates)
- StorageClass for persistent volumes

## Installation

### Add Helm Repository

```bash
helm repo add passkey-auth https://your-github-username.github.io/passkey-auth-helm
helm repo update
```


### Install from Local Chart

```bash
git clone https://github.com/wahyd4/passkey-auth.git
cd passkey-auth
helm install my-passkey-auth ./helm/passkey-auth \
  --values ./helm/passkey-auth/values.yaml
```

The command deploys Passkey Auth on the Kubernetes cluster with the default configuration. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Configuration

### Core Configuration

The chart can be configured using the `values.yaml` file or by passing values via `--set` flags.

#### Required Configuration

```yaml
config:
  webauthn:
    rpId: "auth.example.com"                    # Your authentication domain
    rpOrigins:
      - "https://auth.example.com"              # Allowed origins for WebAuthn

  cors:
    allowedOrigins:
      - "https://*.example.com"                 # CORS allowed origins

  auth:
    cookieDomain: ".example.com"                # Cookie domain for SSO
    allowedEmails:
      - "admin@example.com"                     # Allowed user emails

secrets:
  sessionSecret: "your-secure-random-secret"    # Session signing secret

ingress:
  hosts:
    - host: auth.example.com
      paths:
        - path: /
          pathType: Prefix
```


## Setup Authentication for Your Services

Add these annotations to your ingress resources to protect them with passkey authentication:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-protected-app
  annotations:
    nginx.ingress.kubernetes.io/auth-url: "https://auth.example.com/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://auth.example.com/login?rd=$scheme://$http_host$request_uri"
    nginx.ingress.kubernetes.io/auth-response-headers: "X-Auth-User,X-Auth-Email"
spec:
  # ... your ingress spec
```

## Advanced Configuration

### Custom Storage

```yaml
persistence:
  enabled: true
  existingClaim: "my-existing-pvc"
  storageClass: "ssd-encrypted"
  size: 10Gi
```

### External Secrets

```yaml
envFrom:
  - secretRef:
      name: external-secrets

secrets: {}  # Don't create internal secret
```


## Parameters

### Common parameters

| Name                | Description                                                                              | Value |
| ------------------- | ---------------------------------------------------------------------------------------- | ----- |
| `nameOverride`      | String to partially override passkey-auth.fullname template                             | `""`  |
| `fullnameOverride`  | String to fully override passkey-auth.fullname template                                 | `""`  |
| `commonLabels`      | Add labels to all the deployed resources                                                 | `{}`  |
| `commonAnnotations` | Add annotations to all the deployed resources                                            | `{}`  |

### Passkey Auth parameters

| Name                          | Description                                                     | Value                              |
| ----------------------------- | --------------------------------------------------------------- | ---------------------------------- |
| `replicaCount`                | Number of Passkey Auth replicas to deploy                      | `1`                                |
| `image.repository`            | Passkey Auth image repository                                   | `ghcr.io/wahyd4/passkey-auth`     |
| `image.tag`                   | Passkey Auth image tag (immutable tags are recommended)        | `main`                             |
| `image.pullPolicy`            | Passkey Auth image pull policy                                 | `Always`                           |
| `image.pullSecrets`           | Passkey Auth image pull secrets                                | `[]`                               |

### WebAuthn configuration

| Name                              | Description                                                     | Value                          |
| --------------------------------- | --------------------------------------------------------------- | ------------------------------ |
| `config.webauthn.rpDisplayName`  | WebAuthn Relying Party display name                            | `Passkey Auth`                 |
| `config.webauthn.rpId`           | WebAuthn Relying Party ID (must match your domain)            | `pass.example.com`             |
| `config.webauthn.rpOrigins`      | Allowed origins for WebAuthn (array)                           | `["https://pass.example.com"]` |

### CORS configuration

| Name                              | Description                                                     | Value                            |
| --------------------------------- | --------------------------------------------------------------- | -------------------------------- |
| `config.cors.allowedOrigins`     | CORS allowed origins (array)                                   | `["https://*.example.com"]`     |

### Authentication configuration

| Name                              | Description                                                     | Value                          |
| --------------------------------- | --------------------------------------------------------------- | ------------------------------ |
| `config.auth.requireApproval`    | Require admin approval for new user registrations              | `true`                         |
| `config.auth.cookieDomain`       | Cookie domain for SSO (e.g., .example.com)                     | `.example.com`                 |
| `config.auth.allowedEmails`      | List of allowed email addresses (array)                        | `["admin@example.com"]`        |
| `config.auth.allowedDomains`     | List of allowed email domains (array)                          | `[]`                           |

### Service configuration

| Name           | Description                         | Value       |
| -------------- | ----------------------------------- | ----------- |
| `service.type` | Kubernetes service type             | `ClusterIP` |
| `service.port` | Kubernetes service port             | `80`        |

### Ingress configuration

| Name                       | Description                                                     | Value                          |
| -------------------------- | --------------------------------------------------------------- | ------------------------------ |
| `ingress.enabled`          | Enable ingress controller resource                              | `true`                         |
| `ingress.className`        | IngressClass that will be used to implement the Ingress        | `nginx`                        |
| `ingress.annotations`      | Additional annotations for the Ingress resource                 | `{}`                           |
| `ingress.hosts[0].host`    | Hostname for the ingress                                        | `pass.example.com`             |
| `ingress.hosts[0].paths`   | Paths for the ingress                                           | `[{path: "/", pathType: "Prefix"}]` |
| `ingress.tls`              | TLS configuration for ingress                                   | `[]`                           |

### Persistence configuration

| Name                          | Description                                                     | Value              |
| ----------------------------- | --------------------------------------------------------------- | ------------------ |
| `persistence.enabled`         | Enable persistent volume for data storage                      | `true`             |
| `persistence.storageClass`    | Persistent Volume storage class                                 | `""`               |
| `persistence.accessMode`      | Persistent Volume access mode                                   | `ReadWriteOnce`    |
| `persistence.size`            | Persistent Volume size                                          | `2Gi`              |
| `persistence.existingClaim`   | Use existing persistent volume claim                            | `""`               |

### Security configuration

| Name                                              | Description                                                     | Value                                          |
| ------------------------------------------------- | --------------------------------------------------------------- | ---------------------------------------------- |
| `secrets.sessionSecret`                          | Session secret for signing cookies (change in production!)     | `change-me-in-production-use-long-random-string` |
| `podSecurityContext.fsGroup`                     | Group ID for the pods                                          | `1000`                                         |
| `securityContext.allowPrivilegeEscalation`      | Allow privilege escalation for containers                      | `false`                                        |
| `securityContext.runAsNonRoot`                  | Run containers as non-root user                                | `true`                                         |
| `securityContext.runAsUser`                     | User ID for the containers                                     | `1000`                                         |
| `securityContext.capabilities.drop`             | Dropped capabilities                                           | `["ALL"]`                                      |

### Resource management

| Name                          | Description                                                     | Value    |
| ----------------------------- | --------------------------------------------------------------- | -------- |
| `resources.limits.cpu`        | CPU resource limits                                             | `400m`   |
| `resources.limits.memory`     | Memory resource limits                                          | `512Mi`  |
| `resources.requests.cpu`      | CPU resource requests                                           | `100m`   |
| `resources.requests.memory`   | Memory resource requests                                        | `128Mi`  |

### Autoscaling configuration

| Name                                               | Description                                                     | Value   |
| -------------------------------------------------- | --------------------------------------------------------------- | ------- |
| `autoscaling.enabled`                             | Enable Horizontal Pod Autoscaler                               | `false` |
| `autoscaling.minReplicas`                         | Minimum number of replicas                                     | `1`     |
| `autoscaling.maxReplicas`                         | Maximum number of replicas                                     | `3`     |
| `autoscaling.targetCPUUtilizationPercentage`     | Target CPU utilization percentage                              | `80`    |
| `autoscaling.targetMemoryUtilizationPercentage`  | Target memory utilization percentage                           | `""`    |

### Environment variables

| Name             | Description                                                     | Value   |
| ---------------- | --------------------------------------------------------------- | ------- |
| `env.CONFIG_PATH` | Path to the configuration file                                 | `/app/config.yaml` |
| `env.ADMIN_EMAIL` | Admin email for auto-approval (optional)                      | `""`    |
| `env.DEFAULT_EMAIL` | Default email for initial setup (optional)                  | `""`    |

### Other parameters

| Name          | Description                                                     | Value   |
| ------------- | --------------------------------------------------------------- | ------- |
| `nodeSelector` | Node labels for pod assignment                                 | `{}`    |
| `tolerations`  | Tolerations for pod assignment                                 | `[]`    |
| `affinity`     | Affinity for pod assignment                                    | `{}`    |
