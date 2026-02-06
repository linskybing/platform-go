---
name: kubernetes-integration
description: Best practices for Kubernetes API integration, resource management, and client-go usage in the platform-go project
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
---


# Kubernetes Integration Skill

This skill provides guidelines and patterns for working with Kubernetes resources in the platform-go project.

## When to Use

Apply this skill when:
- Creating, reading, updating, or deleting Kubernetes resources (Pods, PVCs, Services)
- Managing namespaces and resource organization
- Handling GPU resource allocation and MPS configuration
- Implementing retry logic for Kubernetes API calls
- Writing tests that interact with Kubernetes
- Querying Kubernetes resources with label selectors
- Managing storage (PVC) lifecycle operations
- Monitoring Kubernetes resources or resource cleanup

## Core Principles

### 1. Client Initialization & Safety

```go
// Always check if client is initialized (handles test environments)
func (s *Service) CreatePod(ctx context.Context, spec PodSpec) error {
    if k8s.Clientset == nil {
        log.Println("K8s client not initialized, skipping pod creation")
        return nil // Or return mock response for tests
    }
    
    // Proceed with K8s operations
    pod, err := k8s.Clientset.CoreV1().Pods(spec.Namespace).Create(ctx, &spec.Pod, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("failed to create pod: %w", err)
    }
    return nil
}
```

### 2. Namespace Management

```go
// Generate safe namespace names
func GenerateSafeResourceName(prefix, name string, id uint) string {
    safe := strings.ToLower(name)
    safe = regexp.MustCompile("[^a-z0-9-]").ReplaceAllString(safe, "-")
    safe = strings.Trim(safe, "-")
    return fmt.Sprintf("%s-%d-%s", prefix, id, safe)
}

// Ensure namespace exists with labels
func EnsureNamespace(ctx context.Context, name string, labels map[string]string) error {
    if k8s.Clientset == nil {
        return nil
    }
    
    _, err := k8s.Clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
    if err == nil {
        return nil // Already exists
    }
    
    ns := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            Name:   name,
            Labels: labels,
        },
    }
    
    _, err = k8s.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
    return err
}
```

### 3. Label Selectors for Efficient Queries

```go
// Use label selectors instead of fetching all resources
func ListGroupPVCs(ctx context.Context, groupID uint, namespace string) ([]corev1.PersistentVolumeClaim, error) {
    listOpts := metav1.ListOptions{
        LabelSelector: fmt.Sprintf("storage-type=group,group-id=%d", groupID),
    }
    
    pvcs, err := k8s.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, listOpts)
    if err != nil {
        return nil, fmt.Errorf("failed to list PVCs: %w", err)
    }
    
    return pvcs.Items, nil
}

// Standard labels to use
const (
    LabelManagedBy   = "app.kubernetes.io/managed-by"
    LabelName        = "app.kubernetes.io/name"
    LabelComponent   = "app.kubernetes.io/component"
    LabelStorageType = "storage-type"
    LabelGroupID     = "group-id"
    LabelProjectID   = "project-id"
)
```

### 4. Resource Specifications

```go
// Always set resource requests and limits
func CreateJobPod(spec JobSpec) *corev1.Pod {
    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      spec.Name,
            Namespace: spec.Namespace,
            Labels: map[string]string{
                "app":       "platform-job",
                "job-id":    spec.ID,
                "priority":  spec.Priority,
            },
        },
        Spec: corev1.PodSpec{
            RestartPolicy: corev1.RestartPolicyNever,
            PriorityClassName: spec.PriorityClassName,
            Containers: []corev1.Container{
                {
                    Name:    "job",
                    Image:   spec.Image,
                    Command: spec.Command,
                    Resources: corev1.ResourceRequirements{
                        Requests: corev1.ResourceList{
                            corev1.ResourceCPU:    resource.MustParse("100m"),
                            corev1.ResourceMemory: resource.MustParse("128Mi"),
                        },
                        Limits: corev1.ResourceList{
                            corev1.ResourceCPU:    resource.MustParse("1000m"),
                            corev1.ResourceMemory: resource.MustParse("1Gi"),
                        },
                    },
                },
            },
        },
    }
    
    // Add GPU resources if requested
    if spec.GPUCount > 0 {
        addGPUResources(&pod.Spec.Containers[0], spec)
    }
    
    return pod
}
```

### 5. GPU Resource Management

```go
// Handle both dedicated and shared GPU types
func addGPUResources(container *corev1.Container, spec JobSpec) {
    switch spec.GPUType {
    case "dedicated":
        container.Resources.Limits["nvidia.com/gpu"] = resource.MustParse(fmt.Sprintf("%d", spec.GPUCount))
        
    case "shared":
        container.Resources.Limits["nvidia.com/gpu.shared"] = resource.MustParse(fmt.Sprintf("%d", spec.GPUCount))
        
        // Add MPS configuration via annotations
        if spec.Annotations == nil {
            spec.Annotations = make(map[string]string)
        }
        spec.Annotations["mps.nvidia.com/threads"] = "50"
        spec.Annotations["mps.nvidia.com/vram"] = "8000M"
    }
}

// Calculate GPU usage across namespaces
func CountGPUUsage(ctx context.Context, namespacePrefix string) (int, error) {
    namespaces, err := GetFilteredNamespaces(namespacePrefix)
    if err != nil {
        return 0, err
    }
    
    totalGPU := 0
    for _, ns := range namespaces {
        pods, err := k8s.Clientset.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
        if err != nil {
            continue // Log but don't fail entire count
        }
        
        for _, pod := range pods.Items {
            if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
                for _, container := range pod.Spec.Containers {
                    if qty, ok := container.Resources.Requests["nvidia.com/gpu"]; ok {
                        totalGPU += int(qty.Value())
                    }
                    if qty, ok := container.Resources.Requests["nvidia.com/gpu.shared"]; ok {
                        totalGPU += int(qty.Value())
                    }
                }
            }
        }
    }
    
    return totalGPU, nil
}
```

### 6. Storage (PVC) Management

```go
// Create PVC with proper labels and storage class
func CreatePVC(ctx context.Context, namespace, name string, capacityGi int, storageClass string) error {
    qty, err := resource.ParseQuantity(fmt.Sprintf("%dGi", capacityGi))
    if err != nil {
        return fmt.Errorf("invalid capacity: %w", err)
    }
    
    pvc := &corev1.PersistentVolumeClaim{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
            Labels: map[string]string{
                LabelManagedBy: "platform",
                LabelName:      "storage",
            },
        },
        Spec: corev1.PersistentVolumeClaimSpec{
            AccessModes: []corev1.PersistentVolumeAccessMode{
                corev1.ReadWriteMany,
            },
            Resources: corev1.VolumeResourceRequirements{
                Requests: corev1.ResourceList{
                    corev1.ResourceStorage: qty,
                },
            },
            StorageClassName: &storageClass,
        },
    }
    
    _, err = k8s.Clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("failed to create PVC: %w", err)
    }
    
    return nil
}

// Expand existing PVC
func ExpandPVC(ctx context.Context, namespace, name, newSize string) error {
    pvc, err := k8s.Clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        return fmt.Errorf("failed to get PVC: %w", err)
    }
    
    newQty, err := resource.ParseQuantity(newSize)
    if err != nil {
        return fmt.Errorf("invalid size: %w", err)
    }
    
    pvc.Spec.Resources.Requests[corev1.ResourceStorage] = newQty
    
    _, err = k8s.Clientset.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvc, metav1.UpdateOptions{})
    if err != nil {
        return fmt.Errorf("failed to expand PVC: %w", err)
    }
    
    return nil
}
```

### 7. Service & Networking

```go
// Create NodePort service for external access
func CreateService(ctx context.Context, namespace, name string, port int32) (string, error) {
    svc := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
        },
        Spec: corev1.ServiceSpec{
            Type: corev1.ServiceTypeNodePort,
            Selector: map[string]string{
                "app": name,
            },
            Ports: []corev1.ServicePort{
                {
                    Port:       port,
                    TargetPort: intstr.FromInt(int(port)),
                    Protocol:   corev1.ProtocolTCP,
                },
            },
        },
    }
    
    result, err := k8s.Clientset.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
    if err != nil {
        return "", fmt.Errorf("failed to create service: %w", err)
    }
    
    // Return NodePort for external access
    if len(result.Spec.Ports) > 0 {
        return fmt.Sprintf("%d", result.Spec.Ports[0].NodePort), nil
    }
    
    return "", errors.New("no ports assigned")
}
```

### 8. Error Handling & Retries

```go
// Handle "already exists" errors gracefully
func CreateResource(ctx context.Context, obj interface{}) error {
    err := create(ctx, obj)
    if err != nil {
        if apierrors.IsAlreadyExists(err) {
            log.Printf("Resource already exists, skipping creation")
            return nil
        }
        return fmt.Errorf("failed to create resource: %w", err)
    }
    return nil
}

// Use retry with exponential backoff for transient errors
func CreateWithRetry(ctx context.Context, obj interface{}) error {
    backoff := wait.Backoff{
        Steps:    5,
        Duration: 500 * time.Millisecond,
        Factor:   2.0,
        Jitter:   0.1,
    }
    
    return retry.OnError(backoff, func(err error) bool {
        // Retry on transient errors
        return apierrors.IsServerTimeout(err) || 
               apierrors.IsServiceUnavailable(err) ||
               apierrors.IsTooManyRequests(err)
    }, func() error {
        return create(ctx, obj)
    })
}
```

### 9. Cleanup & Resource Deletion

```go
// Delete resources with proper cleanup
func DeleteNamespace(ctx context.Context, name string) error {
    if k8s.Clientset == nil {
        return nil
    }
    
    // Set deletion timeout
    propagation := metav1.DeletePropagationForeground
    deleteOpts := metav1.DeleteOptions{
        PropagationPolicy: &propagation,
    }
    
    err := k8s.Clientset.CoreV1().Namespaces().Delete(ctx, name, deleteOpts)
    if err != nil {
        if apierrors.IsNotFound(err) {
            return nil // Already deleted
        }
        return fmt.Errorf("failed to delete namespace %s: %w", name, err)
    }
    
    // Wait for namespace to be fully deleted (optional, with timeout)
    return waitForNamespaceDeletion(ctx, name, 60*time.Second)
}
```

### 10. Monitoring & Observability

```go
// Add metrics and logging
func MonitorPodStatus(ctx context.Context, namespace, name string) (*PodStatus, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        log.Printf("[METRICS] GetPod duration=%v namespace=%s name=%s", duration, namespace, name)
    }()
    
    pod, err := k8s.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get pod: %w", err)
    }
    
    return &PodStatus{
        Phase:     string(pod.Status.Phase),
        Message:   pod.Status.Message,
        StartTime: pod.Status.StartTime,
    }, nil
}
```

## Testing with Fake Kubernetes Client

```go
// Use fake clientset for unit tests
import "k8s.io/client-go/kubernetes/fake"

func TestCreatePod(t *testing.T) {
    fakeClient := fake.NewSimpleClientset()
    k8s.Clientset = fakeClient
    
    err := CreatePod(context.Background(), podSpec)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    pods, _ := fakeClient.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
    if len(pods.Items) != 1 {
        t.Errorf("expected 1 pod, got %d", len(pods.Items))
    }
}
```

## Common Patterns

### Pattern: Get or Create
```go
func GetOrCreateNamespace(ctx context.Context, name string) error {
    _, err := k8s.Clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
    if err == nil {
        return nil // Already exists
    }
    
    if !apierrors.IsNotFound(err) {
        return fmt.Errorf("failed to check namespace: %w", err)
    }
    
    // Create if not found
    ns := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{Name: name},
    }
    _, err = k8s.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
    return err
}
```

### Pattern: List with Pagination
```go
func ListAllPodsInNamespace(ctx context.Context, namespace string) ([]corev1.Pod, error) {
    var allPods []corev1.Pod
    continueToken := ""
    
    for {
        list, err := k8s.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
            Limit:    100,
            Continue: continueToken,
        })
        if err != nil {
            return nil, err
        }
        
        allPods = append(allPods, list.Items...)
        
        continueToken = list.Continue
        if continueToken == "" {
            break
        }
    }
    
    return allPods, nil
}
```

## Performance Considerations

- Use label selectors to minimize data transfer
- Cache frequently accessed resources (with TTL)
- Batch operations when possible
- Set appropriate timeouts for all K8s API calls
- Use watch API for real-time updates instead of polling

## Security Best Practices

- Use RBAC to limit service account permissions
- Validate all namespace and resource names
- Never expose K8s API credentials in logs
- Use network policies to isolate namespaces
- Scan container images for vulnerabilities
