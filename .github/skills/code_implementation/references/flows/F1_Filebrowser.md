---
title: F1. Filebrowser Session & Creation
description: Flow for creating filebrowser pods/services and returning session access.
---

# F1. Filebrowser Session & Creation Flow

## Steps
1. Handler: `internal/api/handlers/filebrowser_handler.go:GetAccess` validates request and calls `FileBrowserManager.GetFileBrowserAccess`.
2. Manager: `internal/application/k8s/filebrowser_manager.go:GetFileBrowserAccess`:
   - Verify user PVCs and permissions via `permission_manager`.
   - Call `pkg/k8s/CreateFileBrowserPod` and `pkg/k8s/CreateFileBrowserService`.
   - Create session via `pkg/filebrowser/session.go` to return access token/URL.

## Validation Points
- Ensure PVCs exist and are mounted read/write as required.
- Validate access policies and permission checks.

## Tests
- Unit: mock `pkg/k8s` and permission manager to validate flow logic.

## References
- `internal/api/handlers/filebrowser_handler.go`
- `internal/application/k8s/filebrowser_manager.go`
- `pkg/k8s/client.go`
- `pkg/filebrowser/session.go`
