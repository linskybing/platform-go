---
title: P1. Project Creation
description: Detailed flow for creating a project, validations, post-create tasks and tests.
---

# P1. Project Creation

## Steps
1. Request: `POST /api/v1/projects` → `internal/api/routes/project.go`.
2. Handler: `internal/api/handlers/project_handler.go:CreateProject` binds `project.CreateProjectDTO` and extracts `user_id` from context.
3. Service: `internal/application/project/service.go:CreateProject` performs:
   - Validate GID & owner membership via `internal/repository/group.go` or `internal/application/group` services.
   - Create `project.Project` model; call `Repos.Project.CreateProject(p)`.
   - Optionally call `AllocateProjectResources` (see service comments/tests for behavior).
4. Repo: `internal/repository/project.go:CreateProject` inserts and returns `PID`.
5. Post-create async tasks: enqueue namespace/PVC creation if required using `internal/application/k8s` managers; use background goroutine or job queue.

## Validation Points
- DTO validation (required fields, length).
- User membership in group (GID) before allowing creation.
- Proper error wrapping and status codes: 400 validation, 409 duplicate, 500 server errors.

## Caching & Invalidation
- If project lists are cached, invalidate keys such as `project:list` and `user:{id}:projects` after create.

## Tests
- Unit: mock `Repos.Project` and test `CreateProject` success and failure paths (table-driven).
- Integration: `test/integration/project_handler_test.go` provides end-to-end examples.

## References
- `internal/api/routes/project.go`
- `internal/api/handlers/project_handler.go`
- `internal/application/project/service.go`
- `internal/repository/project.go`
