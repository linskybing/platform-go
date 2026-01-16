# Platform API (Go)

RESTful API server written in Go. A Scheduler component is planned but not yet implemented. Uses Gin (HTTP) and PostgreSQL. The backend includes an object-storage interface compatible with MinIO/S3, but MinIO is not deployed by default.

**Status**: API Ready; Scheduler planned

## Documentation

- **[Project Structure Guide](docs/PROJECT_STRUCTURE.md)** - Detailed architecture overview
- **[Testing Report](docs/TESTING_REPORT.md)** - Complete test results and analysis
- **[Quick Reference](docs/QUICK_REFERENCE.md)** - Common commands and tips

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 12+
- (Optional) MinIO for object storage

### Local Development

```bash
# Download dependencies
go mod download

# Run all tests
go test ./...

# Build binaries
make build

# Build API only
make build-api

# Build Scheduler only
make build-scheduler
```

## Notes

- Read each helper script before running; some modify system state
- Ensure PostgreSQL is running before deployment
 - MinIO is not deployed by default. The backend can connect to an external S3/MinIO-compatible object storage; configure endpoint and credentials via Kubernetes Secret or environment variables.

### Deployment & development cautions (important)

- Initial database and `.env`: The project depends on `backend/infra/db/schema.sql` for database schema and seed data (for example, creating the initial `admin` user). Before starting or deploying the backend, provide a `.env` file (or Kubernetes Secret) with the correct database connection string, credentials, and required settings. Without proper DB credentials or permissions, initialization will fail.

- `backend/scripts/build_image.sh` (update registry): This script helps build and push the backend image to a registry. Before deploying to your cluster, update the script's `HARBOR`/`REGISTRY`/`IMAGE`/`TAG` values or modify the script to accept environment variables so the image is pushed to a registry accessible by your cluster.

- Development manifests using `hostPath`: Development manifests may mount local host files (for example certificates or configs) into Pods using `hostPath`. This is convenient for single-node testing but is insecure and non-portable for multi-node or production clusters. Replace `hostPath` with `Secret`, `ConfigMap`, or `PVC` before deploying to shared clusters.

- Suggested deploy order: Build and push images, update manifest image strings, then apply manifests. Suggested order:

  1. `kubectl apply -f ca.yaml` (if TLS/CA manifests are required)
  2. `kubectl apply -f k8s/go-api.yaml` (or `go-api.yaml`)
  3. `kubectl apply -f k8s/postgres.yaml`

  The goal is to ensure Deployments reference images that exist in the registry and that any DB PV/PVC is ready.

### Kubernetes Deployment

```bash
# Deploy all resources
make k8s-deploy

# View status
make k8s-status

# View logs
make k8s-logs-api
make k8s-logs-scheduler

# Cleanup
make k8s-delete
```

## Project Structure

```
platform-go/
 cmd/ # Application entry points
 api/ # REST API server
 scheduler/ # Background job scheduler
 internal/
 api/ # HTTP layer
 application/ # Business logic
 domain/ # Entity models
 repository/ # Data access
 scheduler/ # Scheduling infrastructure
 priority/ # Resource priority management
 pkg/ # Reusable packages
 k8s/ # Kubernetes manifests
 infra/ # Infrastructure setup
 docs/ # Documentation
```

## Common Commands

### Testing
```bash
make test # Run all tests
make test-coverage # Generate coverage report
make test-race # Check for race conditions
```

### Code Quality
```bash
make fmt # Format code
make fmt-check # Check format
make vet # Static analysis
```

### Build & Deploy
```bash
make build # Build API + Scheduler
make clean # Clean artifacts
make ci # Full CI pipeline
make all # Build + Deploy
```

See [Quick Reference](docs/QUICK_REFERENCE.md) for more commands.

## Test Results

- **Unit Tests**: 100+ tests, all passing
- **Code Format**: All files properly formatted
- **Static Analysis**: No issues found
- **Build**: Both API and Scheduler binaries successful

Test coverage:
- `internal/application` - 50+ tests 
- `internal/scheduler` - 10+ tests 
- `internal/priority` - 3 tests 
- `pkg/mps` - 4 tests 
- `pkg/utils` - 8 tests 

See [Testing Report](docs/TESTING_REPORT.md) for details.

## Docker & Kubernetes

### Build Docker Images
Build the API image. The Scheduler image target is present in the repo but the Scheduler component is not implemented yet.
```bash
docker build -t platform-go-api:latest -f deploy/api.Dockerfile .
# docker build -t platform-go-scheduler:latest -f deploy/scheduler.Dockerfile .  # scheduler not implemented
```

### Kubernetes Deployment
Apply API manifests and related resources. Scheduler manifests are provided as placeholders and should not be applied until the Scheduler is implemented.
```bash
# Apply all manifests (API + resources)
kubectl apply -f k8s/

# Or apply individually
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/go-api.yaml
# kubectl apply -f k8s/go-scheduler.yaml  # scheduler not implemented
```

## Configuration

Environment variables (via K8s secrets):
```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=<from-secret>
DB_PASSWORD=<from-secret>
```

See `k8s/secret.yaml` for sensitive data management.

## Helper Scripts

- `scripts/dev.sh` - Local development setup
- `scripts/build_images.sh` - Docker build automation
- `scripts/setup_fakek8s.sh` - Local K8s cluster setup
- `scripts/create_gpu_pod.py` - GPU pod creation helper

## API Documentation

Swagger documentation available at `/swagger.html` when server is running.

See `internal/api/handlers/` for endpoint implementations.

## Development Tips

### Run specific test
```bash
go test ./internal/application -run TestCreateUserGroup -v
```

### Generate coverage HTML
```bash
make coverage-html # Opens in browser
```

### Check code format issues
```bash
make fmt-check
```

### View detailed test output
```bash
make test-verbose
```

## Architecture

1. **API Server** (`cmd/api/`) - HTTP REST interface
 - Port 8080
 - Request/response handling
 - Business logic orchestration

2. **Scheduler** - planned background job processor (not implemented)

Shared code areas:
- Domain models (`internal/domain/`)
- Business logic (`internal/application/`)
- Data access layer (`internal/repository/`)

## Key Files

| File | Purpose |
|---|---|
| `Makefile` | Build and development tasks |
| `.github/workflows/integration-test.yml` | CI/CD pipeline |
| `go.mod` / `go.sum` | Go dependencies |
| `k8s/go-api.yaml` | API deployment config |
| `k8s/go-scheduler.yaml` | Scheduler deployment config |

## Features

- Modular architecture with clear separation of concerns
- Comprehensive unit tests (100+ tests)
- Kubernetes-ready with full YAML configs
- Independent API and Scheduler services
- Priority-based job scheduling
- GPU resource management
- Full audit logging
- Project and configuration file management

## Contributing

1. Make changes to code
2. Run tests: `make test`
3. Format code: `make fmt`
4. Build: `make build`
5. Test in Kubernetes: `make k8s-deploy`

---

**Last Updated**: 2026-01-15 
**Status**: Production Ready

