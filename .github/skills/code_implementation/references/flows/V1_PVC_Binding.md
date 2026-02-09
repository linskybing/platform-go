---
title: V1. PVC Binding
description: Flow for creating project PVCs that bind to group storage, verification and persistence.
---

# V1. PVC Binding (Project ↔ Group Storage)

## Steps
1. Handler: `internal/api/handlers/pvc_binding_handler.go:CreateBinding` binds `storage.CreateProjectPVCBindingRequest`.
2. Manager: `internal/application/k8s/pvc_binding_manager.go:CreateProjectPVCBinding`:
   - Resolve source PV via `getPVNameFromPVC`.
   - Construct `PersistentVolumeClaim` with access modes and capacity.
   - Apply PVC to namespace using `pkg/k8s` API.
3. Persist binding metadata in DB via repository.

## Validation Points
- Confirm PV/PVC compatibility and accessModes.
- Ensure namespace and quota allow requested capacity.

## Tests
- Unit: test `createBindingPVC` with fake K8s client.
- Integration: K8s fixture tests under `test/integration/k8stest/`.

## References
- `internal/api/handlers/pvc_binding_handler.go`
- `internal/application/k8s/pvc_binding_manager.go`
- `internal/repository/storage_permission.go`
