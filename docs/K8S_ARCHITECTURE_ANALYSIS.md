# Kubernetes Architecture Analysis

Analysis of Kubernetes integration and resource management in the platform-go project.

## Table of Contents

- [Overview](#overview)
- [Architecture Components](#architecture-components)
- [Resource Management](#resource-management)
- [Storage Architecture](#storage-architecture)
- [Networking](#networking)
- [Security](#security)
- [Best Practices](#best-practices)

## Overview

Platform-go integrates with Kubernetes to manage containerized workloads, storage, and networking.

### Key Features

- Dynamic namespace management
- Persistent volume claim (PVC) lifecycle
- Pod and deployment orchestration
- ConfigMap and Secret management
- Service exposure and networking

## Architecture Components

### API Server Integration

The platform communicates with Kubernetes API server for all cluster operations.

**Client Configuration**:
- In-cluster config for production
- Kubeconfig for development
- Timeout: 10-30 seconds per operation

### Namespace Organization

Namespaces are used to isolate resources per user and group.

**Naming Patterns**:
- User storage: `user-{username}-storage`
- Group storage: `group-{groupid}-storage`
- System: `default`, `kube-system`

### Resource Types

| Resource | Purpose | Lifecycle |
|----------|---------|-----------|
| Namespace | Isolation boundary | Created with user/group |
| PVC | Persistent storage | Managed independently |
| Pod | Workload execution | Ephemeral |
| Deployment | Pod management | Long-running services |
| Service | Network exposure | Tied to deployments |
| ConfigMap | Configuration | Application-specific |
| Secret | Sensitive data | Encrypted at rest |

## Resource Management

### Storage Management

**User Storage**:
- Each user gets dedicated namespace and PVC
- Default size: 10Gi (configurable)
- Storage class: Configurable via environment
- Expansion: Supported via API

**Group Storage**:
- Shared storage across group members
- Naming: `group-{id}-{uuid8}`
- Lifecycle: Independent of user storage

### Pod Lifecycle

**Creation Flow**:
1. Validate user permissions
2. Apply ConfigMaps and Secrets
3. Create deployment manifest
4. Submit to Kubernetes API
5. Monitor status and events

**Deletion Flow**:
1. Check for dependent resources
2. Delete deployment
3. Clean up ConfigMaps/Secrets
4. Remove namespace (if empty)

## Storage Architecture

### Persistent Volume Claims

**PVC Configuration**:
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: user-john-disk
  namespace: user-john-storage
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
```

### Storage Classes

- `standard` - Default HDD storage
- `fast` - SSD storage (optional)
- `shared` - NFS/shared storage (optional)

### Volume Mounts

User workloads mount PVCs at standard paths:
- User storage: `/data`
- Group storage: `/shared`
- Config files: `/config`

## Networking

### Service Types

**ClusterIP** (default):
- Internal cluster communication
- Not accessible externally

**NodePort**:
- Exposes service on node IP
- Port range: 30000-32767

**LoadBalancer**:
- External cloud load balancer
- Production deployments

### File Browser Access

Platform provides web-based file browser for storage access.

**Implementation**:
- Port forwarding for development
- Ingress for production
- Authentication via platform tokens

## Security

### RBAC Integration

**Service Account**:
- Platform uses dedicated service account
- Permissions scoped to required operations
- No cluster-admin privileges

**Required Permissions**:
- Create/delete namespaces
- Manage PVCs
- Deploy pods and services
- Read/write ConfigMaps and Secrets

### Network Policies

Network policies restrict inter-namespace communication.

**Default Rules**:
- Deny all ingress by default
- Allow egress to DNS
- Explicit allow rules per application

### Secret Management

**Storage**:
- Kubernetes Secrets for sensitive data
- Base64 encoded (not encrypted by default)
- Encryption at rest recommended

**Access Control**:
- RBAC limits secret access
- Secrets mounted as volumes or env vars
- Rotation policy recommended

## Best Practices

### Resource Quotas

Set quotas to prevent resource exhaustion:
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: user-quota
spec:
  hard:
    pods: "10"
    persistentvolumeclaims: "5"
    requests.storage: "100Gi"
```

### Health Checks

Implement liveness and readiness probes:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

### Image Pull Policies

- Use `IfNotPresent` for stable images
- Use `Always` for development
- Configure image pull secrets for private registries

### Cleanup Strategies

**Automated Cleanup**:
- CronJob for unused resources
- TTL controllers for finished jobs
- Namespace finalizers for cleanup hooks

**Manual Cleanup**:
- API endpoints for resource deletion
- Cascade delete for dependent resources
- Validation before deletion

### Monitoring

**Metrics Collection**:
- Pod resource usage (CPU, memory)
- PVC capacity and usage
- API request latency
- Error rates

**Logging**:
- Centralized log aggregation
- Structured logging format
- Log retention policies

## Related Documentation

- [API Standards](API_STANDARDS.md)
- [Main README](../README.md)
