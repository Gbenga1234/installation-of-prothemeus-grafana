<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

## Go Microservices IT Consulting Platform

Project: Complete Go microservices architecture with Kubernetes and Argo CD deployment.

### Architecture Overview
- **Services**: Gateway (3000), Auth (3001), Consulting (3002), Booking (3003)
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Deployment**: Kubernetes with Argo CD
- **Load Balancer**: External access via LoadBalancer service

### Key Technologies
- Go 1.21 with Gin framework
- PostgreSQL for persistent data
- Redis for caching
- Docker containerization
- Kubernetes orchestration
- Argo CD for GitOps deployment

### Development Workflow

1. **Local Testing**: `docker-compose up -d` to run all services
2. **Building Images**: Use provided Dockerfiles for each service
3. **Kubernetes Deployment**: Apply manifests in k8s/manifests/
4. **Argo CD**: Use k8s/argocd/ for production deployments

### File Structure
- `services/` - Microservice implementations
- `k8s/manifests/` - Kubernetes deployment manifests
- `k8s/argocd/` - Argo CD Application definitions
- `Dockerfile.*` - Service-specific container builds
- `docker-compose.yml` - Local development environment

### Important Notes
- Change JWT_SECRET in k8s/manifests/01-namespace-config.yaml before production
- Update container registry references (your-registry) in deployment manifests
- Database credentials should be replaced with secure values for production
- Services use ClusterIP by default; Gateway uses LoadBalancer for external access
