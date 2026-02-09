---
title: Architecture Implementation Flows
description: High-level deployment, CI/CD, and workflow scheduler flows mapped to repository files.
---

# Architecture Flows

## A1. CI/CD Build & Deploy Flow

1. Push/PR triggers GitHub Actions workflow (see `.github/workflows/` or `.github/scripts/`).
2. Test job: `go test ./... -race -coverprofile=coverage.out` (scripts in `.github/scripts`), uploads coverage.
3. Build job: Docker build using `Dockerfile`/`Dockerfile.integration` and tag with commit SHA.
4. Push job: pushes to container registry; artifacts available for deployment job.
5. Deploy job: applies `k8s/` manifests (e.g., `k8s/go-api.yaml`) via `kubectl apply -f k8s/`.
6. Post-deploy: readiness/liveness checks use `/ready` and `/health` handlers defined in application code.

Reference files:
- `.github/workflows` and `.github/scripts` (CI steps)
- `Dockerfile`, `Dockerfile.integration`
- `k8s/` (manifest templates)
- health endpoints implemented in handler packages

## A2. Workflow Submission & Scheduling Flow

1. Client submits a workflow via POST `/api/v1/workflows` (route in `internal/api/routes`).
2. Handler validates payload and enqueues workflow in internal queue or persists in DB via `internal/repository/job.go`.
3. Scheduler picks job and uses `pkg/k8s` or integrates with Argo/Volcano to create job spec (`pkg/k8s/client.go`/`internal/application/k8s`).
4. Progress is streamed via WebSocket endpoints (see `pkg/k8s/websocket.go` and `internal/api/handlers` for watch endpoints).
5. On completion, update job state in DB and emit audit logs via `internal/repository/audit.go`.

Reference files:
- `internal/repository/job.go`, `pkg/k8s/client.go`, `internal/application/k8s`
