# Production Deployment Guide

This guide covers deploying Passkey Auth in a production Kubernetes environment.

## Prerequisites

- Kubernetes cluster with nginx ingress controller
- Domain name with SSL certificate
- kubectl access to the cluster
- Docker registry access (optional, for custom builds)

## Step 1: Prepare Configuration

1. **Generate Session Secret**:
```bash
openssl rand -base64 32
```

2. **Update Kubernetes Configuration**:

Edit `k8s/deployment.yaml` and update the ConfigMap:

```yaml
data:
  config.yaml: |
    server:
      port: "8080"
      host: "0.0.0.0"

    webauthn:
      rp_display_name: "Your Company Auth"
      rp_id: "auth.yourcompany.com"        # Your auth domain
      rp_origins:
        - "https://auth.yourcompany.com"    # Your auth URL
        - "https://app.yourcompany.com"     # Your app URLs

    database:
      path: "/data/passkey-auth.db"

    cors:
      allowed_origins:
        - "https://auth.yourcompany.com"
        - "https://app.yourcompany.com"

    auth:
      session_secret: "YOUR_GENERATED_SECRET_HERE"
      require_approval: true
```

## Step 2: Deploy to Kubernetes

1. **Create Namespace**:
```bash
kubectl apply -f k8s/namespace.yaml
```

2. **Deploy Application**:
```bash
kubectl apply -f k8s/deployment.yaml
```

3. **Verify Deployment**:
```bash
kubectl get pods -n passkey-auth
kubectl logs -f deployment/passkey-auth -n passkey-auth
```

## Step 3: Configure SSL/TLS

Create an ingress with SSL termination:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: passkey-auth-ingress
  namespace: passkey-auth
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"  # If using cert-manager
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - auth.yourcompany.com
    secretName: passkey-auth-tls
  rules:
  - host: auth.yourcompany.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: passkey-auth-service
            port:
              number: 80
```

## Step 4: Configure Your Application Ingress

Update your application's ingress to use passkey auth:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: your-app-ingress
  annotations:
    # Auth configuration
    nginx.ingress.kubernetes.io/auth-url: "http://passkey-auth-service.passkey-auth.svc.cluster.local/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://auth.yourcompany.com"
    nginx.ingress.kubernetes.io/auth-response-headers: "X-Auth-User,X-Auth-User-ID"

    # SSL configuration
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - app.yourcompany.com
    secretName: your-app-tls
  rules:
  - host: app.yourcompany.com
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

## Step 5: Set Up Monitoring (Optional)

### Health Checks

The service includes health checks at `/health`. Configure monitoring:

```yaml
# Add to your monitoring stack
- job_name: 'passkey-auth'
  static_configs:
  - targets: ['passkey-auth-service.passkey-auth.svc.cluster.local:80']
  metrics_path: '/health'
```

### Log Aggregation

Configure log forwarding to your logging system:

```bash
kubectl logs -f deployment/passkey-auth -n passkey-auth | your-log-forwarder
```

## Step 6: Backup Strategy

### Database Backup

Set up regular backups of the SQLite database:

```bash
# Create a backup cronjob
apiVersion: batch/v1
kind: CronJob
metadata:
  name: passkey-auth-backup
  namespace: passkey-auth
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: alpine
            command:
            - /bin/sh
            - -c
            - |
              cp /data/passkey-auth.db /backup/passkey-auth-$(date +%Y%m%d).db
              # Upload to your backup storage
            volumeMounts:
            - name: data
              mountPath: /data
            - name: backup
              mountPath: /backup
          restartPolicy: OnFailure
          volumes:
          - name: data
            persistentVolumeClaim:
              claimName: passkey-auth-storage
          - name: backup
            # Configure your backup storage volume
```

## Security Checklist

- [ ] HTTPS enforced for all endpoints
- [ ] Session secret is randomly generated and secure
- [ ] CORS origins are specifically configured (not "*")
- [ ] WebAuthn RP ID matches your domain exactly
- [ ] Network policies restrict pod-to-pod communication
- [ ] Regular security updates applied
- [ ] Database backups are encrypted and tested
- [ ] Access logs are monitored
- [ ] Kubernetes RBAC is properly configured

## Troubleshooting

### Common Production Issues

1. **WebAuthn Registration Fails**:
   - Verify RP ID matches domain exactly
   - Ensure HTTPS is properly configured
   - Check browser developer console for errors

2. **Auth Backend Returns 502**:
   - Verify service is running: `kubectl get pods -n passkey-auth`
   - Check service connectivity: `kubectl get svc -n passkey-auth`
   - Review nginx ingress logs

3. **Session Issues**:
   - Verify session secret is consistent across restarts
   - Check cookie domain settings
   - Ensure persistent storage is working

### Debug Commands

```bash
# Check service status
kubectl get all -n passkey-auth

# View logs
kubectl logs -f deployment/passkey-auth -n passkey-auth

# Test auth endpoint
kubectl exec -it deployment/passkey-auth -n passkey-auth -- wget -O- http://localhost:8080/health

# Check ingress
kubectl describe ingress -n passkey-auth

# Port forward for testing
kubectl port-forward svc/passkey-auth-service 8080:80 -n passkey-auth
```

## Scaling Considerations

### High Availability

For high availability, consider:

1. **Multiple Replicas**: Increase replica count in deployment
2. **Session Storage**: Use Redis for shared session storage
3. **Database**: Consider PostgreSQL for better concurrent access
4. **Load Balancing**: Ensure proper session affinity

### Performance Tuning

1. **Resource Limits**: Adjust based on usage patterns
2. **Database Optimization**: Regular VACUUM for SQLite
3. **Caching**: Add caching layer for frequently accessed data

## Updates and Maintenance

### Rolling Updates

```bash
# Update the image
kubectl set image deployment/passkey-auth passkey-auth=passkey-auth:v1.1.0 -n passkey-auth

# Monitor rollout
kubectl rollout status deployment/passkey-auth -n passkey-auth

# Rollback if needed
kubectl rollout undo deployment/passkey-auth -n passkey-auth
```

### Database Migrations

For schema changes, implement migration scripts and run them during maintenance windows.

---

This production guide ensures a secure, reliable deployment of Passkey Auth in your Kubernetes environment.
