# Platform API (Go)

RESTful API server for Kubernetes-based platform management. Built with Go, Gin framework, and PostgreSQL.

**Status**: Production Ready (Modern Cloud-Native Architecture)

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)

---

## Overview

Platform-go is a high-performance, multi-tenant backend for managing Kubernetes resources, user authentication, and project hierarchies. It utilizes advanced PostgreSQL patterns (ltree, SKIP LOCKED) to ensure high concurrency and strict data integrity.

### Modernization Status (2026)

This project has been re-architected to support cloud-native scale:
- **Global UUIDs**: All entities use UUID v4.
- **Hierarchical Projects**: Optimized tree traversal using PostgreSQL `ltree`.
- **Concurrent Job Queue**: Lock-free job fetching using `FOR UPDATE SKIP LOCKED`.
- **Strict Integrity**: Polymorphic relationships enforced via Class Table Inheritance (CTI).

---

## Key Features

- **Identity & Access**: JWT-based auth with RBAC (User, Manager, Admin).
- **Hierarchical Management**: Groups and Projects organized in a deeply nested tree structure.
- **Git-like Configuration**: Content-Addressable Storage (CAS) for immutable config versioning.
- **Job Orchestration**: High-throughput job queue with priority scheduling and preemption support.
- **Resource Planning**: Time-window based quotas enforced by database exclusion constraints.
- **Storage Management**: Persistent storage lifecycle with permissions and K8s integration.
- **Audit Logging**: Comprehensive activity tracking for compliance.

---

## Architecture

The platform follows a modular, domain-driven architecture:

- **API Layer**: HTTP REST interface (Gin) with standardized responses.
- **Application Layer**: Business logic services orchestrating domain operations.
- **Domain Layer**: Core entities and business rules (DDD).
- **Repository Layer**: Data access using GORM with PostgreSQL-specific optimizations.
- **Infrastructure**:
    - **PostgreSQL**: Primary data store (with `ltree`, `btree_gist`, `uuid-ossp`).
    - **Redis**: Caching layer for high-read endpoints.
    - **Kubernetes**: Target runtime for workloads.

---

## Quick Start

### Prerequisites

- Go 1.21 or later
- PostgreSQL 13 or later (requires extensions)
- Kubernetes cluster (optional for local dev)

### Run Locally

```bash
# Clone repository
git clone https://github.com/linskybing/platform-go.git
cd platform-go

# Download dependencies
go mod download

# Run all tests (including integration)
bash scripts/run-integration-tests.sh

# Build API
make build-api

# Run API server
./bin/api
```

---

## Installation

### Database Setup

The system relies on specific PostgreSQL extensions.

1. Create PostgreSQL database:
```bash
createdb platform_db
```

2. Apply schema (auto-migrated by app on startup, but initial setup recommended):
```bash
# Ensure extensions are enabled
psql platform_db -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
psql platform_db -c "CREATE EXTENSION IF NOT EXISTS ltree;"
psql platform_db -c "CREATE EXTENSION IF NOT EXISTS btree_gist;"
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env with database credentials
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

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database Host | `localhost` |
| `DB_PORT` | Database Port | `5432` |
| `DB_USER` | Database User | `postgres` |
| `DB_PASSWORD` | Database Password | - |
| `DB_NAME` | Database Name | `platform_db` |
| `PORT` | API Server Port | `8080` |
| `JWT_SECRET` | Secret for token signing | - |
| `KUBECONFIG` | Path to kubeconfig | - |

---

## Development

### Code Quality

```bash
# Format code
make fmt

# Static analysis
go vet ./...
```

### Run Tests

We prioritize integration tests to ensure database contract validity.

```bash
# Run full integration suite (Docker-based)
bash scripts/run-integration-tests.sh

# Run unit tests only
go test ./internal/...
```

### Make Targets

```bash
make build            # Build all binaries
make test             # Run unit tests
make coverage-html    # Generate test coverage report
make k8s-deploy       # Deploy to current K8s context
```

---

## Deployment

### Kubernetes Deployment

1. **Secrets**: Create `platform-secrets` with `db-password` and `jwt-secret`.
2. **Database**: Deploy PostgreSQL with persistent storage.
3. **API**: Apply manifests in `k8s/`.

```bash
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/go-api.yaml
```

---

## API Documentation

The API follows a standardized response format:

```json
{
  "code": 200,
  "message": "Success",
  "data": { ... }
}
```

### Core Resources

- **Auth**: `/api/auth/login`, `/api/auth/register`
- **Users**: `/api/users`
- **Groups**: `/api/groups`
- **Projects**: `/api/projects` (Supports `g_id` for group association)
- **Config Files**: `/api/configfiles` (Versioning)
- **Jobs**: `/api/jobs` (Submission & Tracking)

See [Frontend Integration Guide](docs/en/frontend_integration.md) for detailed wiring instructions.

---

## Documentation

- **[Architecture Guide](docs/en/architecture.md)**: Deep dive into CTI, ltree, and queue design.
- **[Database Schema](docs/en/database.md)**: Schema details and constraints.
- **[Frontend Integration](docs/en/frontend_integration.md)**: Guide for UI developers.

---

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/new-feature`)
3. Make changes and **add tests**
4. Verify with `bash scripts/run-integration-tests.sh`
5. Create Pull Request

---

**Last Updated**: 2026-02-14
**Status**: Production Ready
