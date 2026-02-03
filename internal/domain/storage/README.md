# Storage Domain

Manages all Kubernetes storage resources including Persistent Volume Claims and storage access configurations.

## Table of Contents

1. [Overview](#overview)
2. [Entities](#entities)
3. [DTOs](#dtos)
4. [Repository Interface](#repository-interface)
5. [Usage](#usage)

---

## Overview

The storage domain manages all Kubernetes storage resources including:

- Persistent Volume Claims (PVCs)
- Storage Hubs (pods that mount PVCs for access)
- Project storage infrastructure
- User storage infrastructure
- Volume mounts and access configurations

## Entities

### PersistentVolumeClaim

Database model for tracking Kubernetes PVCs.

- PVC identification and metadata
- Size and storage class information
- Lifecycle status tracking

### StorageHub

Database model for tracking storage access points (pods that mount volumes).

- Hub identification and configuration
- Volume mount mappings
- Access control information

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

1. **K8s application layer** - For storage orchestration
2. **API handlers** - For HTTP endpoints
3. **Services** - For business logic

It replaces the scattered storage types that were previously mixed into `job`, `resource`, and other domains.
