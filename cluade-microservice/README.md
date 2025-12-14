# IT Consulting Platform - Go Microservices

A complete, production-ready IT consulting platform built with Go microservices, containerized with Docker, and deployed to Kubernetes using Argo CD.

## Architecture

The platform consists of 4 core microservices:

- **Gateway Service** (Port 3000) — API gateway/reverse proxy routing requests to backend services
- **Auth Service** (Port 3001) — User authentication, JWT token generation and validation
- **Consulting Service** (Port 3002) — Manages consultant profiles, specialties, and rates
- **Booking Service** (Port 3003) — Handles appointment bookings and scheduling

**Supporting Infrastructure:**
- PostgreSQL 15 — Persistent data storage with automatic initialization
- Redis 7 — Caching and session management
- Kubernetes Load Balancer — External access to the gateway service

## Project Structure

```
.
├── services/
│   ├── gateway/          # API Gateway service
│   ├── auth/             # Authentication service
│   ├── consulting/       # Consultant management service
│   └── booking/          # Booking/scheduling service
├── shared/               # Shared utilities (middleware, etc.)
├── k8s/
│   ├── manifests/        # Kubernetes deployment manifests
│   ├── argocd/           # Argo CD Application definitions
│   └── database/         # Database initialization scripts
├── Dockerfile.*          # Individual service Dockerfiles
├── docker-compose.yml    # Local development compose file
├── go.mod               # Go module definition
└── README.md            # This file
```

## Prerequisites

### For Local Development
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL client tools (optional, for manual DB access)

### For Kubernetes Deployment
- Kubernetes cluster (1.24+) with sufficient resources
- kubectl configured to access the cluster
- Argo CD installed and running in the `argocd` namespace
- Docker registry accessible to your cluster
- LoadBalancer support (cloud provider or metallb for on-prem)

## Local Development with Docker Compose

Start all services locally:

```bash
docker-compose up -d
```

Services will be available at:
- Gateway: http://localhost:3000
- Auth: http://localhost:3001
- Consulting: http://localhost:3002
- Booking: http://localhost:3003
- PostgreSQL: localhost:5432
- Redis: localhost:6379

Check service health:

```bash
curl http://localhost:3000/health
curl http://localhost:3001/health
curl http://localhost:3002/health
curl http://localhost:3003/health
```

Stop all services:

```bash
docker-compose down
```

## Building Container Images

Build all images for a container registry:

```bash
docker build -f Dockerfile.gateway -t your-registry/consulting-gateway:latest .
docker build -f Dockerfile.auth -t your-registry/consulting-auth:latest .
docker build -f Dockerfile.consulting -t your-registry/consulting-service:latest .
docker build -f Dockerfile.booking -t your-registry/consulting-booking:latest .
```

Push to your registry:

```bash
docker push your-registry/consulting-gateway:latest
docker push your-registry/consulting-auth:latest
docker push your-registry/consulting-service:latest
docker push your-registry/consulting-booking:latest
```

## Kubernetes Deployment

### Manual Deployment

1. Update image references in Kubernetes manifests:

```bash
# Edit k8s/manifests/04-gateway.yaml, 05-auth.yaml, 06-consulting.yaml, 07-booking.yaml
# Replace 'your-registry' with your actual container registry
```

2. Apply manifests:

```bash
kubectl apply -f k8s/manifests/01-namespace-config.yaml
kubectl apply -f k8s/manifests/02-postgres.yaml
kubectl apply -f k8s/manifests/03-redis.yaml
kubectl apply -f k8s/manifests/04-gateway.yaml
kubectl apply -f k8s/manifests/05-auth.yaml
kubectl apply -f k8s/manifests/06-consulting.yaml
kubectl apply -f k8s/manifests/07-booking.yaml
```

3. Verify deployment:

```bash
kubectl -n consulting get pods
kubectl -n consulting get svc
```

4. Access the gateway via LoadBalancer:

```bash
kubectl -n consulting get svc gateway-service
# Copy the EXTERNAL-IP and access http://EXTERNAL-IP
```

If no EXTERNAL-IP is assigned (local cluster), use port-forward:

```bash
kubectl -n consulting port-forward svc/gateway-service 8000:80
# Access http://localhost:8000
```

### Argo CD Deployment

1. Ensure Argo CD is installed and running:

```bash
kubectl -n argocd get pods
```

2. Update the Argo CD Application manifest with your Git repository:

```bash
# Edit k8s/argocd/01-consulting-application.yaml
# Update the repoURL to your Git repository
```

3. Apply the Argo CD Application:

```bash
kubectl apply -f k8s/argocd/02-project.yaml
kubectl apply -f k8s/argocd/01-consulting-application.yaml
```

4. Monitor deployment via Argo CD CLI or UI:

```bash
argocd app list
argocd app get consulting-platform
argocd app wait consulting-platform
```

Or access the Argo CD UI (port-forward if needed):

```bash
kubectl -n argocd port-forward svc/argocd-server 8080:443
# Access https://localhost:8080 (accept self-signed cert)
```

## API Endpoints

### Auth Service

```bash
# Register a new user
curl -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Consulting Service

```bash
# List all consultants
curl http://localhost:3000/consultants

# Create a consultant
curl -X POST http://localhost:3000/consultants \
  -H "Content-Type: application/json" \
  -d '{
    "name":"John Smith",
    "speciality":"Cloud Architecture",
    "hourly_rate":150.00,
    "description":"10+ years experience"
  }'

# Get consultant details
curl http://localhost:3000/consultants/consultant-xyz
```

### Booking Service

```bash
# List all bookings
curl http://localhost:3000/bookings

# Create a booking
curl -X POST http://localhost:3000/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "consultant_id":"consultant-xyz",
    "user_id":"user-abc",
    "scheduled_at":"2024-01-15T14:00:00Z",
    "duration":60
  }'

# Get booking details
curl http://localhost:3000/bookings/booking-123

# Update booking status
curl -X PUT http://localhost:3000/bookings/booking-123 \
  -H "Content-Type: application/json" \
  -d '{"status":"completed"}'
```

## Configuration

### Environment Variables

All services respect these environment variables:

```bash
# Services
GATEWAY_PORT=3000
AUTH_PORT=3001
CONSULTING_PORT=3002
BOOKING_PORT=3003

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/consulting_db?sslmode=disable

# Redis
REDIS_URL=localhost:6379

# Security (must be changed in production)
JWT_SECRET=your-secret-key-change-in-production
```

### Kubernetes ConfigMaps and Secrets

ConfigMaps and Secrets are defined in `k8s/manifests/01-namespace-config.yaml`. For production deployments:

1. **Change the JWT Secret**: Replace `your-secret-key-change-in-production` with a strong, randomly generated key
2. **Change Database Credentials**: Update PostgreSQL username and password
3. **Use External Secret Management**: Consider using Sealed Secrets, External Secrets Operator, or HashiCorp Vault

## Scaling and Performance

### Horizontal Scaling

Update replica counts in deployment manifests:

```yaml
# In k8s/manifests/04-gateway.yaml, 05-auth.yaml, 06-consulting.yaml, 07-booking.yaml
spec:
  replicas: 3  # increase from 2
```

### Resource Limits

Adjust resource requests/limits in deployment manifests for your cluster capacity:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "512Mi"
    cpu: "1000m"
```

## Monitoring (Optional)

Integrate with Prometheus and Grafana (same cluster or external):

```bash
# Services expose /health endpoint for liveness/readiness probes
# Add Prometheus scrape config if Prometheus is available

# Or use the provided Prometheus/Grafana setup:
cd ../installation\ of\ prothemeus-grafana
kubectl apply -f prometheus.yml
kubectl apply -f grafana.yml
```

## Troubleshooting

### Check Pod Status

```bash
kubectl -n consulting describe pod <pod-name>
kubectl -n consulting logs <pod-name>
```

### Database Connection Issues

```bash
# Test PostgreSQL connectivity from a pod
kubectl -n consulting exec -it <pod-name> -- sh
# Inside pod: psql -h postgres-service -U postgres -d consulting_db
```

### Service Discovery

```bash
# Verify services are discoverable within the cluster
kubectl -n consulting get svc
kubectl -n consulting get endpoints
```

## Cleanup

### Remove Local Docker Containers

```bash
docker-compose down -v
# Removes containers and volumes
```

### Remove Kubernetes Resources

```bash
# Via kubectl
kubectl delete -f k8s/manifests/
kubectl delete ns consulting

# Via Argo CD
kubectl -n argocd delete application consulting-platform
kubectl -n argocd delete appproject consulting
```

## Security Considerations

### Current State (Development)
- JWT secret is hardcoded (insecure)
- Database passwords in plaintext in manifests (insecure)
- No HTTPS/TLS enforcement

### Production Recommendations

1. **Secrets Management**
   - Use Kubernetes Sealed Secrets or External Secrets Operator
   - Store secrets in HashiCorp Vault, AWS Secrets Manager, or Azure Key Vault
   - Never commit secrets to Git

2. **Network Security**
   - Enable Network Policies to restrict inter-service communication
   - Use HTTPS/TLS for all external communications
   - Implement an Ingress Controller with TLS

3. **Authentication & Authorization**
   - Implement OAuth2/OIDC with external identity provider
   - Add role-based access control (RBAC)
   - Implement API rate limiting and throttling

4. **Database Security**
   - Use managed database services (AWS RDS, Azure Database)
   - Enable encryption at rest and in transit
   - Implement database backup and recovery procedures
   - Use separate credentials per environment

5. **Container Security**
   - Scan images for vulnerabilities (Trivy, Snyk)
   - Use minimal base images (alpine, distroless)
   - Implement Pod Security Standards/Policies

6. **Monitoring & Logging**
   - Centralize logs (ELK, Splunk, CloudWatch)
   - Monitor for security events
   - Set up alerts for anomalies

## Next Steps

1. Set up a Git repository for your deployment manifests
2. Configure Argo CD to track your repository
3. Implement CI/CD pipeline to build and push container images
4. Add comprehensive unit and integration tests
5. Set up development, staging, and production environments
6. Implement API versioning and backward compatibility strategy
7. Add comprehensive API documentation (OpenAPI/Swagger)
8. Implement distributed tracing (Jaeger, DataDog)

## License

MIT License - See LICENSE file for details

## Support

For issues, questions, or contributions, please refer to the project's GitHub repository.

---

**Happy consulting!** 🚀
