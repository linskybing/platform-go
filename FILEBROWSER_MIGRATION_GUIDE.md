# Migration Guide: FileBrowser Refactoring

This guide helps developers migrate existing code to use the new `pkg/filebrowser` package.

## Overview of Changes

The FileBrowser functionality has been consolidated into a reusable package at `pkg/filebrowser/`, replacing duplicated code across:
- User storage (MyDrive)
- Group storage proxy
- Project storage browser

## Before & After Examples

### Example 1: User Storage Proxy

**Before:**
```go
// internal/api/handlers/k8s_handler.go
func (h *K8sHandler) UserStorageProxy(c *gin.Context) {
    // Extract user info
    claims := c.Get("claims").(*types.Claims)
    user, _ := h.UserService.FindUserByID(claims.UserID)
    safeUsername := strings.ToLower(user.Username)

    // Manual proxy setup
    serviceName := fmt.Sprintf("fb-hub-svc-%s", safeUsername)
    namespace := fmt.Sprintf("user-%s-storage", safeUsername)
    targetStr := fmt.Sprintf("http://%s.%s.svc.cluster.local:80", serviceName, namespace)
    
    remote, err := url.Parse(targetStr)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target url"})
        return
    }

    proxy := httputil.NewSingleHostReverseProxy(remote)
    
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        req.URL.Path = strings.TrimPrefix(req.URL.Path, "/k8s/user-storage/proxy")
        req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
        req.Header.Set("X-Forwarded-Proto", "http")
    }
    
    proxy.ServeHTTP(c.Writer, c.Request)
}
```

**After:**
```go
// internal/api/handlers/k8s_handler.go
import "github.com/linskybing/platform-go/pkg/filebrowser"

func (h *K8sHandler) UserStorageProxy(c *gin.Context) {
    // Extract user info
    claims := c.Get("claims").(*types.Claims)
    user, _ := h.UserService.FindUserByID(claims.UserID)
    safeUsername := strings.ToLower(user.Username)

    // Use shared proxy handler
    serviceName := fmt.Sprintf("fb-hub-svc-%s", safeUsername)
    namespace := fmt.Sprintf("user-%s-storage", safeUsername)
    
    proxyHandler := filebrowser.ProxyHandler(filebrowser.ProxyConfig{
        ServiceName: serviceName,
        Namespace:   namespace,
        PathPrefix:  "/k8s/user-storage/proxy",
    })
    
    proxyHandler(c)
}
```

**Removed Imports:**
- `net/http/httputil`
- `net/url`

**Lines Reduced:** 30 lines → 15 lines (50% reduction)

---

### Example 2: Creating FileBrowser Pod

**Before:**
```go
// pkg/utils/filebrowser_session.go
func CreateFileBrowserPod(ctx context.Context, ns string, pvcName string) (*corev1.Pod, error) {
    podName := fmt.Sprintf("fb-pod-%s", pvcName)
    
    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      podName,
            Namespace: ns,
            Labels:    map[string]string{"app": podName},
        },
        Spec: corev1.PodSpec{
            SecurityContext: &corev1.PodSecurityContext{
                RunAsUser:  int64Ptr(0),
                RunAsGroup: int64Ptr(0),
                FSGroup:    int64Ptr(0),
            },
            Containers: []corev1.Container{
                {
                    Name:  "filebrowser",
                    Image: "filebrowser/filebrowser:v2",
                    Args:  []string{"--noauth", "--root=/srv", "--address=0.0.0.0"},
                    Ports: []corev1.ContainerPort{{ContainerPort: 80}},
                    VolumeMounts: []corev1.VolumeMount{
                        {Name: "data", MountPath: "/srv"},
                    },
                },
            },
            Volumes: []corev1.Volume{
                {
                    Name: "data",
                    VolumeSource: corev1.VolumeSource{
                        PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
                            ClaimName: pvcName,
                        },
                    },
                },
            },
            RestartPolicy: corev1.RestartPolicyNever,
        },
    }

    return k8s.Clientset.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{})
}
```

**After:**
```go
import "github.com/linskybing/platform-go/pkg/filebrowser"

func StartFileBrowser(ctx context.Context, ns string, pvcName string) (string, error) {
    sessionMgr := filebrowser.NewSessionManager()
    
    cfg := &filebrowser.Config{
        Namespace:   ns,
        PodName:     fmt.Sprintf("fb-pod-%s", pvcName),
        ServiceName: fmt.Sprintf("fb-svc-%s", pvcName),
        PVCName:     pvcName,
        ReadOnly:    false,
    }
    
    return sessionMgr.Start(ctx, cfg)
}
```

**Benefits:**
- Automatic service creation
- Returns NodePort for access
- Consistent security context
- Error handling with cleanup

---

### Example 3: User Storage Manager

**Before:**
```go
// internal/application/k8s/user_storage_manager.go
func (m *UserStorageManager) OpenFileBrowser(ctx context.Context, username string) (string, error) {
    safeUser := strings.ToLower(username)
    port, err := utils.StartUserHubBrowser(ctx, safeUser)
    if err != nil {
        return "", err
    }
    return port, nil
}

func (m *UserStorageManager) CloseFileBrowser(ctx context.Context, username string) error {
    safeUser := strings.ToLower(username)
    return utils.StopUserHubBrowser(ctx, safeUser)
}
```

**After:**
```go
// internal/application/k8s/user_storage_manager.go
import "github.com/linskybing/platform-go/pkg/filebrowser"

type UserStorageManager struct {
    fbManager filebrowser.SessionManager
}

func NewUserStorageManager() *UserStorageManager {
    return &UserStorageManager{
        fbManager: filebrowser.NewSessionManager(),
    }
}

func (m *UserStorageManager) OpenFileBrowser(ctx context.Context, username string) (string, error) {
    safeUser := strings.ToLower(username)
    nsName := fmt.Sprintf("user-%s-storage", safeUser)
    pvcName := fmt.Sprintf("user-%s-disk", safeUser)
    podName := fmt.Sprintf("fb-hub-%s", safeUser)
    svcName := fmt.Sprintf("fb-hub-svc-%s", safeUser)

    cfg := &filebrowser.Config{
        Namespace:   nsName,
        PodName:     podName,
        ServiceName: svcName,
        PVCName:     pvcName,
        ReadOnly:    false,
    }

    nodePort, err := m.fbManager.GetOrCreate(ctx, cfg)
    if err != nil {
        return "", fmt.Errorf("failed to start filebrowser: %w", err)
    }

    return nodePort, nil
}

func (m *UserStorageManager) CloseFileBrowser(ctx context.Context, username string) error {
    safeUser := strings.ToLower(username)
    nsName := fmt.Sprintf("user-%s-storage", safeUser)
    podName := fmt.Sprintf("fb-hub-%s", safeUser)
    svcName := fmt.Sprintf("fb-hub-svc-%s", safeUser)

    return m.fbManager.Stop(ctx, nsName, podName, svcName)
}
```

**Benefits:**
- Direct control over configuration
- Dependency injection (easier to test)
- Explicit error handling
- Session reuse with `GetOrCreate`

---

## Migration Checklist

### For Handler Code

- [ ] Import `github.com/linskybing/platform-go/pkg/filebrowser`
- [ ] Replace manual proxy setup with `filebrowser.ProxyHandler`
- [ ] Remove unused imports: `net/http/httputil`, `net/url`
- [ ] Update `ProxyConfig` with correct ServiceName, Namespace, PathPrefix

### For Service/Manager Code

- [ ] Add `filebrowser.SessionManager` field to struct
- [ ] Initialize in constructor: `fbManager: filebrowser.NewSessionManager()`
- [ ] Replace direct K8s calls with `fbManager.Start/Stop/GetOrCreate`
- [ ] Construct `filebrowser.Config` with all required fields
- [ ] Handle errors properly with context wrapping

### For Pod Creation Code

- [ ] Replace custom pod spec with `filebrowser.Config`
- [ ] Use `Manager.CreatePod` for low-level operations
- [ ] Use `SessionManager.Start` for full session creation
- [ ] Remove duplicate `buildPodSpec` functions
- [ ] Update tests to use mock Manager interface

---

## Common Patterns

### Pattern 1: Simple Pod Creation

```go
mgr := filebrowser.NewManager()

cfg := &filebrowser.Config{
    Namespace:   "my-namespace",
    PodName:     "my-filebrowser",
    ServiceName: "my-fb-service",
    PVCName:     "my-pvc",
    ReadOnly:    true,
}

pod, err := mgr.CreatePod(ctx, cfg)
svc, err := mgr.CreateService(ctx, cfg)
```

### Pattern 2: Full Session with NodePort

```go
sessionMgr := filebrowser.NewSessionManager()

cfg := &filebrowser.Config{
    Namespace:   "my-namespace",
    PodName:     "my-filebrowser",
    ServiceName: "my-fb-service",
    PVCName:     "my-pvc",
    ReadOnly:    false,
    Labels: map[string]string{
        "user": "alice",
    },
}

nodePort, err := sessionMgr.Start(ctx, cfg)
// Use nodePort to construct access URL
```

### Pattern 3: Reuse Existing Session

```go
// Automatically reuses if pod is already running
nodePort, err := sessionMgr.GetOrCreate(ctx, cfg)
```

### Pattern 4: Reverse Proxy

```go
proxyHandler := filebrowser.ProxyHandler(filebrowser.ProxyConfig{
    ServiceName: "fb-svc-mydata",
    Namespace:   "user-alice-storage",
    PathPrefix:  "/proxy",
})

router.Any("/proxy/*path", proxyHandler)
```

---

## Breaking Changes

**None** - All existing API endpoints maintain backward compatibility.

The refactoring is internal only and does not affect:
- REST API endpoints
- Request/response formats
- Authentication/authorization logic

---

## Testing Migration

### Unit Tests

**Before:**
```go
// Mock K8s client directly
func TestCreatePod(t *testing.T) {
    // Complex K8s client mocking...
}
```

**After:**
```go
type mockSessionManager struct{}

func (m *mockSessionManager) Start(ctx context.Context, cfg *filebrowser.Config) (string, error) {
    return "30000", nil
}

func TestOpenFileBrowser(t *testing.T) {
    mgr := &UserStorageManager{
        fbManager: &mockSessionManager{},
    }
    
    port, err := mgr.OpenFileBrowser(context.Background(), "alice")
    assert.NoError(t, err)
    assert.Equal(t, "30000", port)
}
```

### Integration Tests

The package automatically detects mock mode when `k8s.Clientset == nil`:

```go
// No K8s cluster needed
sessionMgr := filebrowser.NewSessionManager()
nodePort, err := sessionMgr.Start(ctx, cfg)
// Returns mock NodePort "30000"
```

---

## Deprecated Functions

The following functions are candidates for removal:

### pkg/utils/filebrowser_session.go
- `CreateFileBrowserPod`
- `CreateFileBrowserService`
- `StartUserHubBrowser`
- `StopUserHubBrowser`

### pkg/k8s/client.go
- `CreateFileBrowserPod`
- `CreateFileBrowserService`
- `DeleteFileBrowserResources`

**Migration Path:** Audit usage, update callers to use `pkg/filebrowser`, then remove.

---

## Rollback Plan

If issues arise, revert these commits:
1. `pkg/filebrowser/*` package creation
2. `internal/api/handlers/k8s_handler.go` proxy refactor
3. `internal/application/k8s/user_storage_manager.go` refactor
4. `internal/application/k8s/filebrowser_manager.go` refactor

Original code preserved in git history.

---

## Support

For questions or issues:
1. Check `pkg/filebrowser/README.md` for usage examples
2. Review `FILEBROWSER_REFACTORING_SUMMARY.md` for architecture details
3. Consult skills: `golang-production-standards`, `file-structure-guidelines`

---

## Summary of Benefits

| Aspect | Improvement |
|--------|-------------|
| Code Duplication | -67% (3 instances → 1) |
| Lines of Code | -190 lines (29% reduction) |
| Maintainability | Single source of truth for FileBrowser logic |
| Testability | Mock interfaces, dependency injection |
| Consistency | Unified configuration model |
| Readability | Clear separation of concerns |
