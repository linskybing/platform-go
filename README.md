# Platform API (Go)

RESTful API server for Kubernetes-based platform management. Built with Go, Gin framework, and PostgreSQL.

**Status**: Production Ready (core API); scheduler integration is active and still expanding

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [API Documentation](#api-documentation)
- [Documentation](#documentation)
- [Contributing](#contributing)

---

## Overview

Platform-go provides a RESTful API for managing Kubernetes resources, user authentication, project management, and persistent storage.

### Scheduler Integration (FlashJob)

The backend can submit jobs as FlashJob CRDs when `EXECUTOR_MODE=scheduler` and
`FLASH_SCHED_ENABLED=true` are set. The SchedulerExecutor creates FlashJob objects,
tracks status via a watch-based reconciler, and updates the job table accordingly.

Known gaps:
- Workflow submission and template listing are not fully implemented.
- Submit overrides (resources, env, mounts) require additional validation.
- Local executor still needs parity for Job/Workflow CRDs.

### Key Features

- User and group management with RBAC
- Kubernetes resource orchestration (namespaces, PVCs, pods)
- Project and configuration file management
- Image pull job tracking and monitoring
- Persistent storage lifecycle management
- Comprehensive audit logging
- Token-based authentication (JWT)
- PostgreSQL database with view-based queries

### Architecture

The platform follows a modular architecture:
- **API Server** - HTTP REST interface (port 8080)
- **Business Logic** - Application services and domain models
- **Data Layer** - Repository pattern with PostgreSQL
- **K8s Integration** - Client-go based Kubernetes operations

---

## Quick Start

### Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Kubernetes cluster (for full functionality)
- Optional: MinIO or S3-compatible storage

### Run Locally

```bash
# Clone repository
git clone https://github.com/linskybing/platform-go.git
cd platform-go

# Download dependencies
go mod download

# Run tests
go test ./...

# Build API
make build-api

# Run API server
./bin/api
```

---

## Installation

### Database Setup

1. Create PostgreSQL database:
```bash
createdb platform_db
```

2. Apply schema:
```bash
psql platform_db < infra/db/schema.sql
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env with database credentials
```

### Build from Source

```bash
# Build all components
make build

# Build API only
make build-api

# Build with optimizations
make build-production
```

### Docker Build

```bash
# Build API image
docker build -t platform-go-api:latest -f Dockerfile .

# Run container
docker run -p 8080:8080 platform-go-api:latest
```

---

## Configuration

### Environment Variables

Create `.env` file or set environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=platform_db

# Server Configuration
PORT=8080
GIN_MODE=release

# Kubernetes Configuration
KUBECONFIG=/path/to/kubeconfig
DEFAULT_STORAGE_CLASS=standard
USER_PV_SIZE=10Gi

# Security
JWT_SECRET=your_secret_key
BCRYPT_COST=12
```

### Kubernetes Secrets

For production deployment, use Kubernetes secrets:

```bash
kubectl create secret generic platform-secrets \
  --from-literal=db-password=your_password \
  --from-literal=jwt-secret=your_secret
```

---

## Project Structure

```
platform-go/
├── cmd/
│   └── api/              # API server entry point
├── internal/
│   ├── api/              # HTTP handlers and middleware
│   ├── application/      # Business logic services
│   │   ├── user/         # User management
│   │   ├── group/        # Group management
│   │   ├── project/      # Project management
│   │   ├── configfile/   # Config file management
│   │   ├── image/        # Image pull jobs
│   │   └── k8s/          # Kubernetes operations
│   ├── domain/           # Entity models and DTOs
│   ├── repository/       # Data access layer
│   ├── config/           # Configuration management
│   │   └── db/           # Database setup and migrations
│   ├── constants/        # Application constants
│   └── cron/             # Background jobs
├── pkg/
│   ├── k8s/              # Kubernetes client utilities
│   ├── utils/            # Helper functions
│   └── response/         # HTTP response utilities
├── k8s/                  # Kubernetes manifests
├── docs/                 # Documentation
├── Makefile              # Build automation
└── go.mod                # Go dependencies
```

---

## Development

### Code Quality

```bash
# Format code
make fmt

# Check formatting
make fmt-check

# Run linter
make vet

# Static analysis
go vet ./...
```

### Run Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Race detection
make test-race

# Verbose output
make test-verbose

# Specific package
go test ./internal/application/user -v
```

### Generate Coverage Report

```bash
# HTML report
make coverage-html

# Terminal report
go test -cover ./...
```

### Make Targets

Use `make help` to see all available commands. Key targets include:

```bash
# Skills-based commands (from .github/skills)
make skills-lint              # Production standards
make skills-compile           # Compilation check
make skills-migration         # Database migration validation
make docker-integration       # Docker integration tests
make ci-extended              # Extended CI pipeline
make production-check         # Full production validation
```

---

## Testing

### Test Statistics

- **Total Tests**: 100+ unit tests
- **Coverage**: 60%+ across core packages
- **Status**: All passing

### Test Coverage by Package

| Package | Tests | Coverage |
|---------|-------|----------|
| `internal/application/user` | 15+ | 70% |
| `internal/application/group` | 12+ | 65% |
| `internal/application/configfile` | 10+ | 60% |
| `internal/application/k8s` | 8+ | 55% |
| `pkg/utils` | 8+ | 75% |

### Run Specific Tests

```bash
# User service tests
go test ./internal/application/user -run TestRegisterUser -v

# Group service tests
go test ./internal/application/group -v

# Integration tests
make test-integration
```

---

## Deployment

### Kubernetes Deployment

#### Prerequisites

- Kubernetes cluster running
- `kubectl` configured
- Database accessible from cluster

#### Deploy Steps

```bash
# 1. Create namespace
kubectl create namespace platform

# 2. Apply secrets
kubectl apply -f k8s/secret.yaml

# 3. Deploy PostgreSQL (if needed)
kubectl apply -f k8s/postgres.yaml

# 4. Deploy API
kubectl apply -f k8s/go-api.yaml

# 5. Verify deployment
kubectl get pods -n platform
kubectl logs -f deployment/platform-api -n platform
```

#### Using Makefile

```bash
# Deploy all resources
make k8s-deploy

# Check status
make k8s-status

# View logs
make k8s-logs-api

# Delete resources
make k8s-delete
```

### Production Considerations

#### Security

- Use Kubernetes secrets for sensitive data
- Enable RBAC and network policies
- Use TLS for database connections
- Rotate JWT secrets regularly
- Implement rate limiting

#### Scaling

- Horizontal pod autoscaling based on CPU/memory
- Database connection pooling
- Read replicas for PostgreSQL
- Redis cache for frequently accessed data

#### Monitoring

- Prometheus metrics endpoint
- Structured logging with log levels
- Request tracing and correlation IDs
- Health check endpoints (`/health`, `/ready`)

---

## API Documentation

### Base URL

```
http://localhost:8080/api
```

### Authentication

Include JWT token in request header:

```
Authorization: Bearer <your_token>
```

### Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/login` | User login |
| POST | `/api/auth/register` | User registration |
| GET | `/api/users` | List users |
| POST | `/api/users` | Create user |
| GET | `/api/projects` | List projects |
| POST | `/api/projects` | Create project |
| GET | `/api/groups` | List groups |
| POST | `/api/storage/initialize` | Initialize user storage |

For complete API documentation, see [API Standards](docs/API_STANDARDS.md).

---

## Documentation

### Core Documentation

- **[API Standards](docs/API_STANDARDS.md)** - API design, response formats, error handling
- **[K8s Architecture Analysis](docs/K8S_ARCHITECTURE_ANALYSIS.md)** - Kubernetes integration and resource management

### Additional Resources

- `Makefile` - Build targets and automation
- `k8s/` - Kubernetes deployment manifests
- `.github/workflows/` - CI/CD pipeline configuration

---

## Contributing

### Workflow

1. Fork the repository
2. Create feature branch (`git checkout -b feature/new-feature`)
3. Make changes and add tests
4. Run tests and formatting (`make test && make fmt`)
5. Commit changes (`git commit -m 'Add new feature'`)
6. Push to branch (`git push origin feature/new-feature`)
7. Create Pull Request

### Code Standards

- Follow Go best practices and idioms
- Write unit tests for new features
- Maintain test coverage above 60%
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small

### Pull Request Guidelines

- Describe changes clearly in PR description
- Reference related issues
- Ensure all tests pass
- Update documentation if needed
- Keep commits atomic and well-described

---

**Last Updated**: 2026-02-02  
**Version**: 1.0.0  
**Status**: Production Ready

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
 - `internal/priority` - 3 tests 
 - `pkg/mps` - (deprecated/legacy) tests may remain
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

