---
title: I1. Image Pull & Pull Job
description: Flow for requesting image pulls, tracking pull jobs and exposing status endpoints.
---

# I1. Image Pull & Pull Job Flow

## Steps
1. Client requests image pull via handler `internal/api/handlers/image_handler.go`.
2. Handler calls `internal/application/image` service which creates a `pull job` record in DB and enqueues a worker.
3. Worker calls cluster agents or `pkg/k8s` to pull image(s) to cluster nodes or mirror registry.
4. Service tracks pull job state and exposes endpoints for active/failed pulls (`GetActivePullJobs`, `GetFailedPullJobs`).

## Validation Points
- Validate image names and tags; handle private registry auth.
- Ensure idempotency for repeated pull requests.

## Tests
- Unit: mock registry and repo; verify job lifecycle and state transitions.

## References
- `internal/api/handlers/image_handler.go`
- `internal/application/image`
- `internal/repository/job.go`
