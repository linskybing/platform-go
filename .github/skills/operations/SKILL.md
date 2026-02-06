---
name: operations
description: CI/CD automation, Kubernetes integration, caching strategies, production monitoring, and workflow scheduling for platform-go
license: Proprietary
metadata:
  author: platform-go
  version: "1.1"
  consolidated_from:
    - cicd-pipeline-optimization
    - kubernetes-integration
    - redis-caching
    - monitoring-observability
---

# Operations Excellence

Comprehensive guidelines for CI/CD automation, Kubernetes deployment, Redis caching, production monitoring, and workflow orchestration with Argo Workflow and Volcano Scheduler.

## CI/CD Pipeline

### GitHub Actions Workflow Structure
```yaml
name: CI/CD Pipeline
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.20"
      
      - name: Run Tests
        run: go test ./... -timeout 5m -race -coverprofile=coverage.out
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
  
  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build Docker Image
        run: docker build -t platform-go:${{ github.sha }} .
      
      - name: Push to Registry
        run: |
          docker tag platform-go:${{ github.sha }} docker.io/org/platform-go:latest
          docker push docker.io/org/platform-go:latest
  
  deploy:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to K8s
        run: kubectl apply -f k8s/
```

### Pipeline Optimization
- **Parallel jobs**: Test, lint, build simultaneously
- **Caching**: Go modules cache for faster builds
- **Artifact caching**: Docker layer caching
- **Early exit**: Fail fast on test failures

## Kubernetes Integration

### Deployment Pattern
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: platform-go
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: platform-go
  template:
    metadata:
      labels:
        app: platform-go
    spec:
      containers:
      - name: platform-go
        image: platform-go:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

### Client-go Best Practices
```go
// Initialize K8s client
kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
if err != nil {
    log.Fatal(err)
}

clientset, err := kubernetes.NewForConfig(kubeConfig)
if err != nil {
    log.Fatal(err)
}

// Use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// List resources with field selectors
pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
    FieldSelector: "status.phase=Running",
    LabelSelector: "app=platform-go",
})
```

### Pod Management
```go
// Create pod with proper initialization
pod := &corev1.Pod{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "user-storage-" + userID,
        Namespace: "default",
        Labels: map[string]string{
            "app":    "filebrowser",
            "userId": userID,
        },
    },
    Spec: corev1.PodSpec{
        Containers: []corev1.Container{{
            Name:  "filebrowser",
            Image: "filebrowser:latest",
            Ports: []corev1.ContainerPort{{
                ContainerPort: 80,
            }},
        }},
    },
}

created, err := clientset.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})

// Wait for pod ready
for i := 0; i < 30; i++ {
    pod, err := clientset.CoreV1().Pods("default").Get(ctx, "user-storage-"+userID, metav1.GetOptions{})
    if pod.Status.Phase == corev1.PodRunning {
        break
    }
    time.Sleep(1 * time.Second)
}
```

### Resource Cleanup
```go
// Delete pod with grace period
gracePeriod := int64(30)
propagation := metav1.DeletePropagationForeground

err := clientset.CoreV1().Pods("default").Delete(ctx, podName, metav1.DeleteOptions{
    GracePeriodSeconds: &gracePeriod,
    PropagationPolicy:  &propagation,
})

// Wait for actual deletion
time.Sleep(5 * time.Second) // Grace period + buffer
```

## Redis Caching Strategy

### Cache Architecture
```
┌──────────────────────────────────┐
│     Application Request          │
└────────────┬─────────────────────┘
             │
             ├─► Check Redis Cache
             │   ├─► HIT → Return cached data
             │   └─► MISS → Fetch from DB
             │          ├─► Update Redis
             │          └─► Return fresh data
             │
        Invalidation Events
             ├─► User modified → invalidate user cache
             ├─► Project updated → invalidate project cache
             └─► Cron job → refresh TTL
```

### Implementation Pattern
```go
// Generic cache operation
func GetOrFetch[T any](ctx context.Context, cacheKey string, 
    fetchFn func() (T, error)) (T, error) {
    
    // Try cache first
    if cached, err := cache.Get[T](ctx, cacheKey); err == nil {
        return cached, nil
    }
    
    // Fetch fresh data
    data, err := fetchFn()
    if err != nil {
        return data, err
    }
    
    // Store in cache with TTL
    _ = cache.Set(ctx, cacheKey, data, 1*time.Hour)
    return data, nil
}

// Usage
user, err := GetOrFetch(ctx, fmt.Sprintf("user:%d", userID),
    func() (*User, error) {
        return db.GetUser(ctx, userID)
    })
```

### Cache Invalidation
```go
// User-triggered invalidation
func UpdateUser(ctx context.Context, userID int, data *UpdateRequest) error {
    // Update database
    if err := db.UpdateUser(ctx, userID, data); err != nil {
        return err
    }
    
    // Invalidate related caches
    cache.Delete(ctx, fmt.Sprintf("user:%d", userID))
    cache.Delete(ctx, fmt.Sprintf("user:%d:projects", userID))
    cache.Delete(ctx, "user:list") // Invalidate list too
    
    return nil
}
```

## Production Monitoring

### Structured Logging
```go
log.WithFields(map[string]interface{}{
    "user_id": userID,
    "action":  "create_pod",
    "pod_name": podName,
    "duration_ms": elapsed,
}).Info("Pod created successfully")
```

### Metrics Collection
```go
// Request latency histogram
http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        metrics.RequestDuration.Observe(duration.Seconds())
    }()
    
    // Handler logic
})

// Error counting
if err != nil {
    metrics.ErrorCount.WithLabelValues(errorType).Inc()
    return err
}
```

### Health Check Endpoints
```go
// /health - Simple health check
func Health(c *gin.Context) {
    c.JSON(200, gin.H{"status": "healthy"})
}

// /ready - Readiness check (dependencies)
func Ready(c *gin.Context) {
    // Check database connection
    if err := db.Ping(c.Request.Context()); err != nil {
        c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
        return
    }
    
    // Check Redis connection
    if err := cache.Ping(c.Request.Context()); err != nil {
        c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"status": "ready"})
}
```

### Alert Thresholds
- **Error rate**: > 1% of requests
- **Latency**: p99 > 1 second
- **CPU**: > 80% sustained
- **Memory**: > 85% usage
- **Disk**: > 90% full

## Workflow & Job Scheduler

For complete guide on implementing Argo Workflow, Volcano Scheduler, and K8s native schedulers, including:
- System architecture and scheduler comparison
- Backend service design and API specifications
- Frontend integration patterns
- Deployment and configuration

**See:** [WORKFLOW_SCHEDULER.md](./WORKFLOW_SCHEDULER.md)

### Frontend Integration

Complete guide for frontend developers to integrate with workflow APIs including:
- TypeScript/JavaScript client library
- React components for workflow submission and monitoring
- Real-time progress tracking via WebSocket
- Queue management dashboard
- Complete code examples

**See:** [FRONTEND_INTEGRATION.md](./FRONTEND_INTEGRATION.md)

### Quick Overview
- **Volcano Scheduler**: Best for batch ML/HPC jobs with gang scheduling
- **K8s Native**: Standard pod priority and job queuing
- **Argo Workflows**: Complex multi-step workflow orchestration

## Tools & Scripts

### Deployment Scripts
```bash
# Deploy new version
bash .github/skills/operations/scripts/deploy.sh

# Check cluster health
bash .github/skills/operations/scripts/health-check.sh

# Cache monitoring
bash .github/skills/operations/scripts/monitor-cache.sh

# Kubernetes pod status
bash .github/skills/operations/scripts/k8s-status.sh

# Run integration tests
bash .github/skills/operations/scripts/run-integration-tests.sh

# K8s integration testing with Kind
bash .github/skills/operations/scripts/kind-setup.sh
bash .github/skills/operations/scripts/k8s-integration-test.sh
bash .github/skills/operations/scripts/kind-cleanup.sh
```

## References
- GitHub Actions: https://docs.github.com/en/actions
- Kubernetes API: https://kubernetes.io/docs/reference/
- Redis Documentation: https://redis.io/documentation
- Prometheus Metrics: https://prometheus.io/docs/
- Argo Workflows: https://argoproj.github.io/argo-workflows/
- Volcano Scheduler: https://volcano.sh/
