# FileBrowser Package

Unified FileBrowser pod and service management for platform-go.

## Overview

This package provides reusable components for creating, managing, and accessing FileBrowser instances in Kubernetes. It eliminates code duplication between user storage (MyDrive) and group storage proxy implementations.

## Components

### Manager
Low-level Kubernetes resource management for FileBrowser pods and services.

```go
type Manager interface {
    CreatePod(ctx context.Context, cfg *Config) (*corev1.Pod, error)
    CreateService(ctx context.Context, cfg *Config) (*corev1.Service, error)
    DeletePod(ctx context.Context, namespace, podName string) error
    DeleteService(ctx context.Context, namespace, serviceName string) error
    WaitForReady(ctx context.Context, namespace, podName string) error
    GetPod(ctx context.Context, namespace, podName string) (*corev1.Pod, error)
    GetService(ctx context.Context, namespace, serviceName string) (*corev1.Service, error)
    DeleteResources(ctx context.Context, namespace, podName, serviceName string) error
}
```

### SessionManager
High-level session management with automatic pod and service lifecycle.

```go
type SessionManager interface {
    Start(ctx context.Context, cfg *Config) (nodePort string, err error)
    Stop(ctx context.Context, namespace, podName, serviceName string) error
    GetOrCreate(ctx context.Context, cfg *Config) (nodePort string, err error)
}
```

### ProxyHandler
Generic reverse proxy handler for routing HTTP traffic to FileBrowser instances.

```go
func ProxyHandler(cfg ProxyConfig) gin.HandlerFunc
```

## Usage Examples

### Creating a FileBrowser Session

```go
import "github.com/linskybing/platform-go/pkg/filebrowser"

// Create session manager
sessionMgr := filebrowser.NewSessionManager()

// Configure FileBrowser
cfg := &filebrowser.Config{
    Namespace:   "user-alice-storage",
    PodName:     "fb-hub-alice",
    ServiceName: "fb-hub-svc-alice",
    PVCName:     "user-alice-disk",
    ReadOnly:    false,
    Labels: map[string]string{
        "user": "alice",
    },
}

// Start session (creates pod + service)
nodePort, err := sessionMgr.Start(ctx, cfg)
if err != nil {
    return fmt.Errorf("failed to start filebrowser: %w", err)
}

fmt.Printf("FileBrowser available on port %s\n", nodePort)
```

### Reusing Existing Sessions

```go
// GetOrCreate reuses existing pod if running
nodePort, err := sessionMgr.GetOrCreate(ctx, cfg)
// Returns immediately if pod already exists and is running
```

### Stopping a Session

```go
err := sessionMgr.Stop(ctx, "user-alice-storage", "fb-hub-alice", "fb-hub-svc-alice")
// Deletes both pod and service
```

### Setting Up Reverse Proxy

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/linskybing/platform-go/pkg/filebrowser"
)

// Configure proxy
proxyHandler := filebrowser.ProxyHandler(filebrowser.ProxyConfig{
    ServiceName: "fb-hub-svc-alice",
    Namespace:   "user-alice-storage",
    PathPrefix:  "/k8s/user-storage/proxy",
})

// Register route
router.Any("/k8s/user-storage/proxy/*path", proxyHandler)
```

## Configuration

### Config Struct

```go
type Config struct {
    Namespace   string            // K8s namespace
    PodName     string            // FileBrowser pod name
    ServiceName string            // K8s service name
    PVCName     string            // PVC to mount
    BaseURL     string            // FileBrowser base URL (optional)
    ReadOnly    bool              // Mount PVC as read-only
    Labels      map[string]string // Additional pod/service labels
}
```

### Default Values

| Constant | Value | Description |
|----------|-------|-------------|
| DefaultImage | `filebrowser/filebrowser:v2` | FileBrowser container image |
| DefaultPort | `80` | Container port |
| DefaultRunAsUser | `1000` | Security context UID |
| DefaultRunAsGroup | `1000` | Security context GID |
| DefaultFSGroup | `1000` | Filesystem group ID |

## Security

### Pod Security Context

All FileBrowser pods are created with:
- **RunAsUser**: 1000 (non-root)
- **RunAsGroup**: 1000
- **FSGroup**: 1000 (ensures consistent file ownership)

### PVC Mount Options

- **ReadOnly**: Configurable per session
  - `true`: Mount PVC as read-only (safe for viewers)
  - `false`: Mount PVC as read-write (for editors)

## Architecture

### Separation of Concerns

```
Manager (manager.go)
├── Low-level K8s operations
├── Pod/Service CRUD
└── Mock detection for testing

SessionManager (session.go)
├── High-level orchestration
├── Pod + Service lifecycle
└── Session reuse logic

ProxyHandler (proxy.go)
└── HTTP reverse proxy to FileBrowser
```

### Mock Support

For testing environments where Kubernetes is not available:

```go
if k8s.Clientset == nil {
    // Returns mock pod/service
    // Allows handlers to function without K8s cluster
}
```

## Integration Points

### User Storage (MyDrive)

- **Handler**: `internal/api/handlers/k8s_handler.go`
  - `OpenMyDrive`: Creates user FileBrowser
  - `StopMyDrive`: Stops user FileBrowser
  - `UserStorageProxy`: Proxies to user FileBrowser

- **Manager**: `internal/application/k8s/user_storage_manager.go`
  - Uses `SessionManager` for pod/service lifecycle

### Group Storage

- **Manager**: `internal/application/k8s/filebrowser_manager.go`
  - Uses `SessionManager` for permission-based FileBrowser access
  - Supports read-only and read-write pods

## Best Practices

### 1. Always Use Context

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

nodePort, err := sessionMgr.Start(ctx, cfg)
```

### 2. Cleanup on Error

```go
nodePort, err := sessionMgr.Start(ctx, cfg)
if err != nil {
    // SessionManager automatically cleans up partial resources
    return err
}
```

### 3. Reuse Sessions

```go
// Prefer GetOrCreate over Start to avoid unnecessary pod creation
nodePort, err := sessionMgr.GetOrCreate(ctx, cfg)
```

### 4. Label Your Resources

```go
cfg := &filebrowser.Config{
    // ... other fields ...
    Labels: map[string]string{
        "user":         username,
        "group":        groupID,
        "access-type":  "read-only",
    },
}
```

## Testing

### Unit Tests

Mock the Manager interface:

```go
type mockManager struct{}

func (m *mockManager) CreatePod(ctx context.Context, cfg *Config) (*corev1.Pod, error) {
    return &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{Name: cfg.PodName},
        Status:     corev1.PodStatus{Phase: corev1.PodRunning},
    }, nil
}
```

### Integration Tests

The package detects when `k8s.Clientset` is nil and returns mock responses, allowing integration tests to run without a Kubernetes cluster.

## Troubleshooting

### Pod Not Ready

```go
err := mgr.WaitForReady(ctx, namespace, podName)
if err != nil {
    // Pod failed to reach ready state within timeout
    // Check pod logs for FileBrowser startup errors
}
```

### Service Not Accessible

Verify service NodePort:

```bash
kubectl -n user-alice-storage get svc fb-hub-svc-alice
```

### PVC Mount Issues

Check PVC status and access modes:

```bash
kubectl -n user-alice-storage get pvc user-alice-disk
```

## Metrics

Recommended metrics to track:

- FileBrowser pod creation time
- Session reuse rate
- Active FileBrowser instances
- Pod failures and restarts

## Future Enhancements

- [ ] Automatic cleanup of idle pods (TTL)
- [ ] Session caching to reduce K8s API calls
- [ ] Health check endpoints
- [ ] Custom FileBrowser configurations
- [ ] Resource limits (CPU/Memory)

## References

- [FileBrowser Documentation](https://filebrowser.org/)
- [Kubernetes Pod API](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#pod-v1-core)
- [Kubernetes Service API](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#service-v1-core)
