---
title: C1. ConfigFile Deploy
description: Detailed flow for creating configfile instances, volume binding, K8s resource creation and tests.
---

# C1. ConfigFile Deploy (Create Instance)

## Steps
1. Handler: `internal/api/handlers/configfile_handler.go:CreateInstanceHandler` receives `id` and request.
2. Service: `internal/application/configfile/deploy.go:CreateInstance`:
   - Load configfile and project via `Repos.Project.GetProjectByID`.
   - Parse YAML documents; call `bindProjectAndUserVolumes` to map PVCs.
   - Build pod specs and call `pkg/k8s/client.go` functions to create resources.
   - Update instance status in DB and create audit log.

## Validation Points
- Ensure YAML has required documents and manifests.
- Validate target namespace and PVC availability.

## Tests
- Unit: mock K8s client and project repo; validate calls and error handling.
- Integration: see `internal/application/configfile_service_test.go` for fixtures and examples.

## References
- `internal/api/handlers/configfile_handler.go`
- `internal/application/configfile/deploy.go`
- `pkg/k8s/client.go`
- `internal/repository/audit.go`
