# Storage Domain

The storage domain manages all Kubernetes storage resources including:
- Persistent Volume Claims (PVCs)
- Storage Hubs (pods that mount PVCs for access)
- Project storage infrastructure
- User storage infrastructure
- Volume mounts and access configurations

## Entities

### PersistentVolumeClaim
Database model for tracking Kubernetes PVCs.

### StorageHub
Database model for tracking storage access points (pods that mount volumes).

## DTOs

### VolumeSpec
Generic volume specification used for volume mounting and PVC operations.

### CreatePVCRequest / ExpandPVCRequest
Request DTOs for PVC lifecycle operations.

### CreateProjectStorageRequest / ProjectStorageInfo
Request/response for project-scoped storage.

### ExpandStorageRequest / UserStorageInfo
Request/response for user-scoped storage.

### VolumeMount
Configuration for mounting volumes in containers.

### FileBrowserConfig / FileBrowserInfo
Configuration and information for FileBrowser instances.

## Repository Interface

The `Repository` interface defines CRUD operations for:
- PVC management (create, read, list, update, delete)
- StorageHub management

## Usage

This domain should be imported and used by:
1. **k8s application layer** - For storage orchestration
2. **API handlers** - For HTTP endpoints
3. **Services** - For business logic

It replaces the scattered storage types that were previously mixed into `job`, `resource`, and other domains.
