---
title: Implementation Flows Index
description: Consolidated implementation flows for platform-go (Project, ConfigFile, PVC, ImagePull, Filebrowser, Auth). Centralized reference to code paths, tests, and quality guidance.
---

# Implementation Flows Index

This file consolidates all implementation flows and points to dedicated subfiles for each flow. Use the links to jump directly to detailed, focused flow documents.

## Quick Index

- [P1. Project Creation](flows/P1_Project_Creation.md#p1-project-creation)
- [C1. ConfigFile Deploy (Create Instance)](flows/C1_ConfigFile_Deploy.md#c1-configfile-deploy)
- [V1. PVC Binding (Project ↔ Group Storage)](flows/V1_PVC_Binding.md#v1-pvc-binding)
- [I1. Image Pull & Pull Job](flows/I1_Image_Pull.md#i1-image-pull--pull-job-flow)
- [F1. Filebrowser Session & Creation](flows/F1_Filebrowser.md#f1-filebrowser-session--creation-flow)
- [Auth. User Signup/Login/Access](flows/Auth_User_Flow.md#user-flow-signup--login--access-project)

---

Open any of the links above to view the detailed step-by-step flow, validation points, test recommendations, and repository file references.
---

## P1. Project Creation (expanded)

Steps:
1. Request: `POST /api/v1/projects` → `internal/api/routes/project.go`.
2. Handler: `internal/api/handlers/project_handler.go:CreateProject` binds `project.CreateProjectDTO` and extracts `user_id` from context.
3. Service: `internal/application/project/service.go:CreateProject` performs:
   - Validate GID & owner membership via `internal/repository/group.go` or `internal/application/group` services.
   - Create `project.Project` model; call `Repos.Project.CreateProject(p)`.
   - Optionally call `AllocateProjectResources` (see service comments/tests for behavior).
4. Repo: `internal/repository/project.go:CreateProject` inserts and returns `PID`.
5. Post-create async tasks: enqueue namespace/PVC creation if required using `internal/application/k8s` managers; use background goroutine or job queue.

Tests:
- Unit: mock `Repos.Project` and test `CreateProject` success and failure paths.
- Integration: `test/integration/project_handler_test.go` shows end-to-end behavior.

Reference files:
- `internal/api/routes/project.go`
- `internal/api/handlers/project_handler.go`
- `internal/application/project/service.go`
- `internal/repository/project.go`

---

## C1. ConfigFile Deploy (expanded)

Steps:
1. Handler: `internal/api/handlers/configfile_handler.go:CreateInstanceHandler` receives `id` and `CreateInstance` request.
2. Service: `internal/application/configfile/deploy.go:CreateInstance`:
   - Load configfile and project via `Repos.Project.GetProjectByID`.
   - Parse YAML documents; call `bindProjectAndUserVolumes` to map PVCs.
   - Build pod specs and call `pkg/k8s/client.go` functions to create resources.
   - Update instance status in DB and create audit log.

Tests:
- Unit: mock K8s client and project repo; validate that `CreateInstance` attempts correct calls.
- Integration: `internal/application/configfile_service_test.go` contains examples.

Reference files:
- `internal/api/handlers/configfile_handler.go`
- `internal/application/configfile/deploy.go`
- `pkg/k8s/client.go`
- `internal/repository/audit.go`

---

## V1. PVC Binding (expanded)

Steps:
1. Handler: `internal/api/handlers/pvc_binding_handler.go:CreateBinding` binds `storage.CreateProjectPVCBindingRequest`.
2. Manager: `internal/application/k8s/pvc_binding_manager.go:CreateProjectPVCBinding`:
   - Resolve source PV via `getPVNameFromPVC`.
   - Construct `PersistentVolumeClaim` with access modes and capacity.
   - Apply PVC to namespace using `pkg/k8s` API.
3. Persist binding metadata in DB via repository.

Tests:
- Unit: test `createBindingPVC` with fake K8s client.
- Integration: K8s fixture tests under `test/integration/k8stest/`.

Reference files:
- `internal/api/handlers/pvc_binding_handler.go`
- `internal/application/k8s/pvc_binding_manager.go`
- `internal/repository/storage_permission.go`

---

## I1. Image Pull & Pull Job Flow

Steps:
1. Client requests image pull (via image handler `internal/api/handlers/image_handler.go`).
2. Handler calls `internal/application/image` service which creates a `pull job` record in DB and enqueues a worker.
3. Worker calls cluster agents or `pkg/k8s` to pull images to cluster nodes or to registry mirror.
4. Service tracks pull job state and exposes endpoints for active/failed pulls (`GetActivePullJobs`, `GetFailedPullJobs`).

Tests:
- Unit: mock registry and repo; verify job lifecycle updates.

Reference files:
- `internal/api/handlers/image_handler.go`
- `internal/application/image`
- `internal/repository/job.go`

---

## F1. Filebrowser Session & Creation Flow

Steps:
1. Handler: `internal/api/handlers/filebrowser_handler.go:GetAccess` validates request and calls `FileBrowserManager.GetFileBrowserAccess`.
2. Manager: `internal/application/k8s/filebrowser_manager.go:GetFileBrowserAccess`:
   - Verify user PVCs and permissions via `permission_manager`.
   - Call `pkg/k8s/CreateFileBrowserPod` and `pkg/k8s/CreateFileBrowserService`.
   - Create session via `pkg/filebrowser/session.go` to return access token/URL.

Tests:
- Unit: mock `pkg/k8s` and permission manager to validate flow.

Reference files:
- `internal/api/handlers/filebrowser_handler.go`
- `internal/application/k8s/filebrowser_manager.go`
- `pkg/k8s/client.go`
- `pkg/filebrowser/session.go`

---

## Auth & User Flow (Signup → Login → Access Project)

This consolidates the detailed user_flow content and maps to authentication, session, and authorization code.

Steps (summary):
1. Signup: `internal/api/handlers/auth_handler.go:Signup` → `internal/application/user/service.go:RegisterUser` → `internal/repository/user.go:CreateUser`.
2. Login: `internal/api/handlers/auth_handler.go:Login` → `internal/application/user/service.go:Authenticate` → issue JWT (env secret) and optional refresh token.
3. Middleware: `internal/api/middleware/*` extracts and validates JWT or `X-API-Key`, attaches `user_id` to context.
4. Project access: `internal/api/handlers/project_handler.go:GetProjectByID` → `internal/application/project/service.go:GetProject` performs authorization checks.

Tests & Quality:
- Unit: table-driven tests for `Authenticate`, mock `UserRepo` with gomock.
- Integration: `test/integration/auth_test.go` and project handler integration tests.
- Follow `code-quality` SKILL.md: context-first, error wrapping, <=200 lines/file where practical.

Reference files:
- `internal/api/handlers/auth_handler.go`
- `internal/application/user/service.go`
- `internal/repository/user.go`
- `internal/api/middleware/extractors.go`

---

## How to use this index

- Read relevant flow section and follow referenced files as anchors for implementation.
- When modifying behavior, update this index to keep team documentation accurate.
- Prefer small, testable changes; add unit tests and integration tests as noted per flow.
